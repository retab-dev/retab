package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// Reserved environment slugs. `test` and `production` are fixed canonical
// environments; `live` is accepted as user input but always normalized to
// `production` so the stored slug stays stable.
const (
	slugTest       = "test"
	slugProduction = "production"
	slugLiveAlias  = "live"
)

// customSlugPattern matches user-defined environment slugs. Conservative on
// purpose — slugs are command-line identifiers, not free text.
var customSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// normalizeSlug lowercases the input and folds the `live` alias onto
// `production`. It does not validate — call validateSlug for that.
func normalizeSlug(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == slugLiveAlias {
		return slugProduction
	}
	return s
}

// validateSlug returns the canonical slug for raw, or an error. `test`,
// `production`, and the `live` alias are always accepted; any other slug
// must match customSlugPattern.
func validateSlug(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("environment slug must not be empty")
	}
	s := normalizeSlug(raw)
	if s == slugTest || s == slugProduction {
		return s, nil
	}
	if !customSlugPattern.MatchString(s) {
		return "", fmt.Errorf("invalid environment slug %q: must match ^[a-z0-9][a-z0-9_-]{0,62}$", raw)
	}
	return s, nil
}

// credentialSource enumerates where a resolved credential came from. It is
// surfaced in `auth status` so users can debug "why is the wrong account
// being used?".
type credentialSource string

const (
	sourceFlagKey     credentialSource = "--api-key flag"
	sourceEnvKey      credentialSource = "RETAB_API_KEY env"
	sourceEnvFlag     credentialSource = "--env profile"
	sourceLiveFlag    credentialSource = "--live profile"
	sourceDefaultEnv  credentialSource = "default environment profile"
	sourceOAuth       credentialSource = "stored OAuth session"
	sourceAccessToken credentialSource = "stored access_token"
	sourceLegacyKey   credentialSource = "legacy stored api_key"
)

// resolvedCredential is the structured output of the shared credential
// resolver. Exactly one of APIKey / OAuth / AccessToken is populated.
type resolvedCredential struct {
	// Source records which precedence branch won.
	Source credentialSource

	// Override is true when the credential came from an explicit
	// per-command selector (--api-key, RETAB_API_KEY, --env, --live)
	// rather than the stored default.
	Override bool

	// APIKey is set for API-key auth modes (everything except OAuth).
	APIKey string

	// OAuth is set only for the stored-OAuth-session branch.
	OAuth *oauthTokens

	// AccessToken is set only for stored scoped access-token auth.
	AccessToken string

	// BaseURL is the resolved Retab deployment URL ("" means SDK default).
	BaseURL string

	// ProfileSlug is the stored profile slug that was selected, if any.
	ProfileSlug string

	// ExpectedEnvironment is the CLI's local guess at the customer
	// environment, derived from the key prefix or the profile slug. The
	// server is authoritative — this is a UX hint only.
	ExpectedEnvironment string
}

// KeyPreview returns a redacted preview of the resolved static credential, or
// "" for OAuth credentials. Never returns the full key/token.
func (r resolvedCredential) KeyPreview() string {
	switch {
	case r.APIKey != "":
		return redactKey(r.APIKey)
	case r.AccessToken != "":
		return redactKey(r.AccessToken)
	default:
		return ""
	}
}

// environmentFromKeyPrefix maps an API key prefix to the customer
// environment the CLI expects. Legacy `sk_retab_` keys resolve to
// production. Unknown prefixes return "".
func environmentFromKeyPrefix(key string) string {
	switch {
	case strings.HasPrefix(key, "rt_test_"):
		return slugTest
	case strings.HasPrefix(key, "rt_live_"):
		return slugProduction
	case strings.HasPrefix(key, "sk_retab_test_"):
		return slugTest
	case strings.HasPrefix(key, "sk_retab_"):
		return slugProduction
	default:
		return ""
	}
}

// oauthExpectedEnvironment maps an OAuth session's persisted environment
// type onto the production-confirmation gate's environment slug. OAuth
// sessions carry no environment-scoped API key prefix, so the gate relies
// on the type captured at selection time (env switch / claim / auth login,
// stored in cfg.EnvironmentType).
//
//   - production      -> slugProduction (high-risk commands are gated)
//   - non_production  -> ""             (never gated)
//   - empty / unknown -> slugProduction (fail SAFE)
//
// The empty case covers configs written before environment-type
// persistence (legacy/pre-rollout) and any session that never resolved an
// environment. It fails safe to production — the same conservative stance
// as legacy API keys — so a credential whose environment cannot be locally
// proven non-production is gated rather than silently allowed. The server
// remains authoritative; this only drives the local confirmation prompt.
func oauthExpectedEnvironment(cfg retabConfig) string {
	if cliEnvironmentType(strings.TrimSpace(cfg.EnvironmentType)) == cliEnvironmentTypeNonProduction {
		return ""
	}
	return slugProduction
}

// resolveBaseURL applies the deployment-selection precedence shared by
// every command: --base-url flag > RETAB_BASE_URL env > profile base_url >
// stored base_url. An empty result means "use the SDK default".
//
// RETAB_BASE_URL selects a Retab deployment only; it is never the customer
// environment selector.
func resolveBaseURL(cmd *cobra.Command, cfg retabConfig, profile *environmentProfile) string {
	if v, _ := cmd.Root().PersistentFlags().GetString("base-url"); v != "" {
		return v
	}
	if v := os.Getenv("RETAB_BASE_URL"); v != "" {
		return v
	}
	if profile != nil && profile.BaseURL != "" {
		return profile.BaseURL
	}
	return cfg.BaseURL
}

// resolveCredential is the single shared credential resolver used by
// newClient, cliJSONRequest, and `auth status`. It implements the
// blueprint's "Final Credential Resolution" precedence:
//
//  1. --api-key flag
//  2. RETAB_API_KEY environment variable
//  3. --env <slug> stored environment profile
//  4. --live stored production profile
//  5. stored default_environment profile
//  6. stored access_token
//  7. stored OAuth session
//  8. legacy stored api_key
//  9. unauthenticated error
//
// Conflicting selectors are rejected up front (see the blueprint's
// "Conflict rules"): an explicit key already selects the environment, so
// combining it with --live or --env is an error rather than a silent
// override.
func resolveCredential(cmd *cobra.Command) (resolvedCredential, error) {
	flagKey, _ := cmd.Root().PersistentFlags().GetString("api-key")
	envKey := os.Getenv("RETAB_API_KEY")
	live, _ := cmd.Root().PersistentFlags().GetBool("live")
	envFlag, _ := cmd.Root().PersistentFlags().GetString("env")

	cfg, _ := loadConfig()

	// --- Conflict detection -------------------------------------------------
	// An explicit key (flag or env) already encodes the environment.
	// Combining it with a CLI environment selector creates two sources of
	// truth, so fail loudly per the blueprint.
	if flagKey != "" && envFlag != "" {
		return resolvedCredential{}, fmt.Errorf("--api-key cannot be combined with --env; the key already selects the environment")
	}
	if flagKey != "" && live {
		return resolvedCredential{}, fmt.Errorf("--live cannot be combined with --api-key; the key already selects the environment")
	}
	if envKey != "" && envFlag != "" && flagKey == "" {
		return resolvedCredential{}, fmt.Errorf("--env cannot be combined with RETAB_API_KEY; the environment variable key already selects the environment")
	}
	if envKey != "" && live && flagKey == "" {
		return resolvedCredential{}, fmt.Errorf("--live cannot be combined with RETAB_API_KEY; the environment variable key already selects the environment")
	}
	// --live is an alias for --env production; --env test contradicts it.
	if live && envFlag != "" {
		normalized, err := validateSlug(envFlag)
		if err != nil {
			return resolvedCredential{}, err
		}
		if normalized != slugProduction {
			return resolvedCredential{}, fmt.Errorf("--live conflicts with --env %s; --live is an alias for --env production", envFlag)
		}
	}

	// --- 1. --api-key flag --------------------------------------------------
	if flagKey != "" {
		return resolvedCredential{
			Source:              sourceFlagKey,
			Override:            true,
			APIKey:              flagKey,
			BaseURL:             resolveBaseURL(cmd, cfg, nil),
			ExpectedEnvironment: environmentFromKeyPrefix(flagKey),
		}, nil
	}

	// --- 2. RETAB_API_KEY env ----------------------------------------------
	if envKey != "" {
		return resolvedCredential{
			Source:              sourceEnvKey,
			Override:            true,
			APIKey:              envKey,
			BaseURL:             resolveBaseURL(cmd, cfg, nil),
			ExpectedEnvironment: environmentFromKeyPrefix(envKey),
		}, nil
	}

	// --- 3. --env <slug> stored profile ------------------------------------
	if envFlag != "" {
		slug, err := validateSlug(envFlag)
		if err != nil {
			return resolvedCredential{}, err
		}
		profile := cfg.Environments[slug]
		if profile == nil || profile.APIKey == "" {
			return resolvedCredential{}, fmt.Errorf("no credential configured for environment %q. Run `retab env add %s --api-key <key>`", slug, slug)
		}
		return profileCredential(cmd, cfg, slug, profile, sourceEnvFlag, true), nil
	}

	// --- 4. --live stored production profile -------------------------------
	if live {
		profile := cfg.Environments[slugProduction]
		if profile == nil || profile.APIKey == "" {
			return resolvedCredential{}, fmt.Errorf("no live credential configured. Run `retab auth login --live --api-key rt_live_...` or pass `--api-key rt_live_...` for this command")
		}
		return profileCredential(cmd, cfg, slugProduction, profile, sourceLiveFlag, true), nil
	}

	// --- 5. stored default_environment profile -----------------------------
	if cfg.DefaultEnvironment != "" {
		profile := cfg.Environments[cfg.DefaultEnvironment]
		if profile != nil && profile.APIKey != "" {
			return profileCredential(cmd, cfg, cfg.DefaultEnvironment, profile, sourceDefaultEnv, false), nil
		}
	}

	// --- 6. stored access_token --------------------------------------------
	if cfg.AccessToken != "" {
		return resolvedCredential{
			Source:              sourceAccessToken,
			AccessToken:         cfg.AccessToken,
			BaseURL:             resolveBaseURL(cmd, cfg, nil),
			ExpectedEnvironment: slugProduction,
		}, nil
	}

	// --- 7. stored OAuth session -------------------------------------------
	if cfg.OAuth != nil && cfg.OAuth.AccessToken != "" {
		return resolvedCredential{
			Source:              sourceOAuth,
			OAuth:               cfg.OAuth,
			BaseURL:             resolveBaseURL(cmd, cfg, nil),
			ExpectedEnvironment: oauthExpectedEnvironment(cfg),
		}, nil
	}

	// --- 8. legacy stored api_key ------------------------------------------
	if cfg.APIKey != "" {
		// Classify by prefix exactly as branches 1 and 2 do for the same key
		// supplied via --api-key / RETAB_API_KEY. `auth login --api-key` stores
		// ANY non-acctk_ key here, including an rt_test_ one, so hardcoding
		// production meant a test key demanded the production confirmation (and
		// hard-failed CI with "production write requires --confirm") purely
		// because it had been saved rather than passed. Unknown and legacy
		// prefixes still resolve to production — the safe default for a key the
		// CLI cannot place.
		expected := environmentFromKeyPrefix(cfg.APIKey)
		if expected == "" {
			expected = slugProduction
		}
		return resolvedCredential{
			Source:              sourceLegacyKey,
			APIKey:              cfg.APIKey,
			BaseURL:             resolveBaseURL(cmd, cfg, nil),
			ExpectedEnvironment: expected,
		}, nil
	}

	// --- 9. unauthenticated -------------------------------------------------
	return resolvedCredential{}, fmt.Errorf("no credentials configured. Run `retab auth login` or set RETAB_API_KEY")
}

// profileCredential builds a resolvedCredential from a stored profile.
func profileCredential(cmd *cobra.Command, cfg retabConfig, slug string, profile *environmentProfile, source credentialSource, override bool) resolvedCredential {
	expected := profile.ServerEnvironmentSlug
	if expected == "" {
		expected = environmentFromKeyPrefix(profile.APIKey)
	}
	if expected == "" {
		expected = slug
	}
	return resolvedCredential{
		Source:              source,
		Override:            override,
		APIKey:              profile.APIKey,
		BaseURL:             resolveBaseURL(cmd, cfg, profile),
		ProfileSlug:         slug,
		ExpectedEnvironment: expected,
	}
}
