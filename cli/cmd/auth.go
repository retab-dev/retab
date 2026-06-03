package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Retab authentication",
	Long: `Manage credentials the CLI uses to talk to the Retab API.

Covers interactive OAuth login, headless API-key login, inspecting the
currently active credential, and clearing local state. Credentials are
resolved with --api-key flag > RETAB_API_KEY env > ~/.retab/config.json
(mode 0600, parent dir 0700).`,
	Example: `  # Interactive login (browser OAuth by default)
  retab auth login

  # Headless / CI login with a long-lived key
  retab auth login --api-key=sk_live_abc123

  # Show which credential the CLI would use right now
  retab auth status

  # Forget the local credential (does not revoke the key server-side)
  retab auth logout`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Retab",
	Long: `Log in to Retab and persist the credential locally.

By default, opens a browser for an OAuth login flow via WorkOS. The
resulting tokens are saved to ~/.retab/config.json (mode 0600) and
refreshed transparently when needed.

For headless setups (CI, servers, scripts) skip the browser flow and
store a long-lived API key instead with --api-key. Re-running login is
always safe: it overwrites the saved credential in place, which is also
how key rotation works. RETAB_API_KEY remains honored as a process-wide
override and takes precedence over anything written to disk.`,
	Example: `  # Interactive OAuth flow (opens a browser)
  retab auth login

  # Headless / CI: pass the key inline
  retab auth login --api-key=sk_live_abc123

  # Headless without echoing the key to history
  retab auth login --api-key="$RETAB_API_KEY"

  # Prompt for an API key without opening a browser
  retab auth login --browser=false`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")
		baseURL, _ := cmd.Flags().GetString("base-url")
		useBrowser, _ := cmd.Flags().GetBool("browser")
		live, _ := cmd.Root().PersistentFlags().GetBool("live")
		setDefault, _ := cmd.Flags().GetBool("default")

		// Explicit API-key path. Direct, no browser. Backward compatible
		// with the original login UX.
		if apiKey != "" || (!useBrowser && os.Getenv("RETAB_API_KEY") != "") {
			return runAPIKeyLogin(apiKey, baseURL, live, setDefault)
		}

		// If `--no-browser` was requested and no key was given, prompt.
		if !useBrowser {
			prompted, err := promptSecret("API key (leave blank to use browser flow): ")
			if err != nil {
				return err
			}
			if strings.TrimSpace(prompted) != "" {
				return runAPIKeyLogin(prompted, baseURL, live, setDefault)
			}
			// fall through to browser flow
		}

		ctx, cancel := ctxFor(cmd)
		defer cancel()

		// Discovery + browser OAuth flow.
		effectiveBaseURL := baseURL
		if effectiveBaseURL == "" {
			effectiveBaseURL = os.Getenv("RETAB_BASE_URL")
		}
		disc, err := fetchOAuthDiscovery(ctx, effectiveBaseURL)
		if err != nil {
			return fmt.Errorf("OAuth discovery failed: %w", err)
		}
		tokens, err := runLoginFlow(ctx, disc, openBrowser)
		if err != nil {
			return err
		}

		cfg, _ := loadConfig()
		cfg.OAuth = tokens
		// Switching to OAuth wipes the legacy API key; users who want
		// both can re-run `retab auth login --api-key …` afterward.
		cfg.APIKey = ""
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		path, _ := configPath()
		fmt.Fprintf(os.Stderr, "Logged in. Saved OAuth tokens to %s\n", path)
		return nil
	}),
}

// runAPIKeyLogin persists an API key. It writes BOTH the legacy top-level
// `api_key` field (so older CLI versions and unchanged code paths keep
// working) AND a named environment profile, picking the profile slug from
// the --live flag or the key prefix. The profile becomes the local default
// when none exists yet or --default was passed.
func runAPIKeyLogin(apiKey, baseURL string, live, setDefault bool) error {
	if apiKey == "" {
		apiKey = os.Getenv("RETAB_API_KEY")
	}
	if apiKey == "" {
		prompted, err := promptSecret("API key: ")
		if err != nil {
			return err
		}
		apiKey = strings.TrimSpace(prompted)
	}
	if apiKey == "" {
		return fmt.Errorf("api key is required")
	}

	// Decide the profile slug. --live always means production; otherwise
	// infer from the key prefix; legacy/unknown prefixes default to test
	// (the safer day-to-day environment) unless --live was given.
	slug := slugTest
	if live {
		slug = slugProduction
	} else if env := environmentFromKeyPrefix(apiKey); env != "" {
		slug = env
	}

	cfg, _ := loadConfig()
	// Wipe stale OAuth state — explicit key login is the user's intent.
	cfg.OAuth = nil
	// Keep writing the legacy field for backward compatibility.
	cfg.APIKey = apiKey
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if cfg.Environments == nil {
		cfg.Environments = map[string]*environmentProfile{}
	}
	created := time.Now().UTC().Format(time.RFC3339)
	if existing := cfg.Environments[slug]; existing != nil {
		created = existing.CreatedAt
	}
	cfg.Environments[slug] = &environmentProfile{
		Name:          defaultDisplayName(slug),
		APIKey:        apiKey,
		APIKeyPreview: redactKey(apiKey),
		BaseURL:       baseURL,
		CreatedAt:     created,
	}
	if cfg.DefaultEnvironment == "" || setDefault {
		cfg.DefaultEnvironment = slug
	}
	if err := saveConfig(cfg); err != nil {
		return err
	}
	path, _ := configPath()
	fmt.Fprintf(os.Stderr, "Saved API key to %s (environment %q).\n", path, slug)
	if cfg.DefaultEnvironment == slug {
		fmt.Fprintf(os.Stderr, "Default CLI environment is now %q.\n", slug)
	}
	return nil
}

// authUseCmd sets the local default environment profile. `live` is folded
// onto `production`.
var authUseCmd = &cobra.Command{
	Use:   "use <test|live|slug>",
	Short: "Set the default local environment profile",
	Long: `Set which stored environment profile the CLI uses by default.

This is equivalent to ` + "`retab env switch`" + ` and only changes local state —
it never mutates server state or changes API-key authorization.`,
	Example: `  retab auth use test
  retab auth use live`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		slug, err := validateSlug(args[0])
		if err != nil {
			return err
		}
		cfg, _ := loadConfig()
		if cfg.Environments[slug] == nil {
			return fmt.Errorf("no environment profile %q. Run `retab auth login --api-key <key>` or `retab env add %s --api-key <key>` first", slug, slug)
		}
		cfg.DefaultEnvironment = slug
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Default CLI environment is now %q.\n", slug)
		return nil
	}),
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove ~/.retab/config.json",
	Long: `Delete the local credential file at ~/.retab/config.json.

Only clears LOCAL state — the API key (or OAuth refresh token) is not
revoked on the server. To revoke a key for real, rotate it in the Retab
dashboard. After logout, commands that need authentication will fail
until ` + "`retab auth login`" + ` runs again or RETAB_API_KEY is set in
the environment.`,
	Example: `  # Forget the credential on this machine
  retab auth logout

  # Switch accounts: forget, then log in again
  retab auth logout && retab auth login`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := deleteConfig(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Cleared local credentials.")
		return nil
	}),
}

// authStatusResponse mirrors the server's GET /v1/auth/status payload.
type authStatusResponse struct {
	Authenticated   bool   `json:"authenticated"`
	OrganizationID  string `json:"organization_id"`
	EnvironmentID   string `json:"environment_id"`
	EnvironmentSlug string `json:"environment_slug"`
	AuthMethod      string `json:"auth_method"`
	Region          string `json:"region"`
	KeyPrefix       string `json:"key_prefix"`
	KeyName         string `json:"key_name"`
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show which credential the CLI is using and its resolved environment",
	Long: `Report the credential the CLI would use for the next request and verify
it against the server's ` + "`GET /v1/auth/status`" + ` endpoint.

The server resolves the real customer environment from the API key, so
` + "`auth status`" + ` shows the authoritative ` + "`resolved_environment`" + ` rather than
guessing from the key prefix. It also lists every configured environment
profile with redacted key previews — full keys are never printed.

Apply ` + "`--live`" + ` or ` + "`--env <slug>`" + ` to inspect what a one-off override
would resolve to.`,
	Example: `  # Quick check
  retab auth status

  # Script — assert authenticated
  retab auth status | jq -e '.authenticated == true'

  # What would --live resolve to?
  retab auth status --live`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		jsonOnly, _ := cmd.Flags().GetBool("json")
		cfg, _ := loadConfig()

		out := map[string]any{}
		if cfg.DefaultEnvironment != "" {
			out["default_environment"] = cfg.DefaultEnvironment
		}

		// Configured profiles, redacted. Always shown regardless of probe.
		if len(cfg.Environments) > 0 {
			profiles := map[string]any{}
			for slug, p := range cfg.Environments {
				entry := map[string]any{
					"configured":      true,
					"api_key_preview": p.APIKeyPreview,
				}
				if entry["api_key_preview"] == "" {
					entry["api_key_preview"] = redactKey(p.APIKey)
				}
				if p.ServerEnvironmentSlug != "" {
					entry["server_environment_slug"] = p.ServerEnvironmentSlug
				}
				profiles[slug] = entry
			}
			out["profiles"] = profiles
		}

		// Resolve the credential the next command would use.
		cred, err := resolveCredential(cmd)
		if err != nil {
			out["authenticated"] = false
			out["error"] = err.Error()
			out["hint"] = "run `retab auth login` to authenticate"
			return writeAuthStatus(cmd.OutOrStdout(), out, jsonOnly)
		}
		out["credential_source"] = string(cred.Source)
		if preview := cred.KeyPreview(); preview != "" {
			out["active_credential"] = preview
		}
		if cred.ProfileSlug != "" {
			out["active_profile"] = cred.ProfileSlug
		}
		if cred.BaseURL != "" {
			out["base_url"] = cred.BaseURL
		}

		// Verify against GET /v1/auth/status — authoritative environment.
		resp, probeErr := cliJSONRequest(cmd, http.MethodGet, "v1/auth/status", nil, nil)
		if probeErr != nil {
			out["authenticated"] = false
			out["error"] = probeErr.Error()
			if cred.ExpectedEnvironment != "" {
				out["expected_environment"] = cred.ExpectedEnvironment
			}
			return writeAuthStatus(cmd.OutOrStdout(), out, jsonOnly)
		}

		status := decodeAuthStatus(resp)
		out["authenticated"] = status.Authenticated
		if status.OrganizationID != "" {
			out["organization_id"] = status.OrganizationID
		}
		if status.EnvironmentSlug != "" {
			out["resolved_environment"] = status.EnvironmentSlug
		}
		if status.EnvironmentID != "" {
			out["environment_id"] = status.EnvironmentID
		}
		if status.AuthMethod != "" {
			out["auth_method"] = status.AuthMethod
		}
		if status.Region != "" {
			out["region"] = status.Region
		}
		if status.KeyName != "" {
			out["key_name"] = status.KeyName
		}
		if status.KeyPrefix != "" {
			out["key_prefix"] = status.KeyPrefix
		}
		// Legacy keys resolve to production without a modern prefix —
		// surface that explicitly so users know it's a compatibility path.
		if status.KeyPrefix == "sk_retab" && status.EnvironmentSlug == slugProduction {
			out["legacy_key"] = true
		}
		// Flag a prefix/server mismatch — the server is authoritative.
		if cred.ExpectedEnvironment != "" && status.EnvironmentSlug != "" &&
			cred.ExpectedEnvironment != status.EnvironmentSlug {
			out["environment_mismatch"] = fmt.Sprintf(
				"local guess %q but server resolved %q", cred.ExpectedEnvironment, status.EnvironmentSlug)
		}
		return writeAuthStatus(cmd.OutOrStdout(), out, jsonOnly)
	}),
}

// decodeAuthStatus converts the loosely-typed cliJSONRequest result into a
// typed authStatusResponse.
func decodeAuthStatus(raw any) authStatusResponse {
	var status authStatusResponse
	obj, ok := raw.(map[string]any)
	if !ok {
		return status
	}
	if v, ok := obj["authenticated"].(bool); ok {
		status.Authenticated = v
	}
	getStr := func(k string) string {
		s, _ := obj[k].(string)
		return s
	}
	status.OrganizationID = getStr("organization_id")
	status.EnvironmentID = getStr("environment_id")
	status.EnvironmentSlug = getStr("environment_slug")
	status.AuthMethod = getStr("auth_method")
	status.Region = getStr("region")
	status.KeyPrefix = getStr("key_prefix")
	status.KeyName = getStr("key_name")
	return status
}

// writeAuthStatus decides whether `auth status` renders JSON or a human
// status block, and writes the result to w.
//
// Output mode:
//   - --json flag set         → JSON
//   - w is not a TTY          → JSON (so pipes / redirects stay parseable)
//   - otherwise               → human block
//
// JSON output is byte-equivalent to the legacy `printJSON(out)` form
// (encoding/json's encoder with 2-space indent, trailing newline) so any
// script consuming `auth status | jq …` keeps working.
func writeAuthStatus(w io.Writer, out map[string]any, jsonOnly bool) error {
	if jsonOnly || !isTerminalWriter(w) {
		return writeAuthStatusJSON(w, out)
	}
	return writeAuthStatusHuman(w, out)
}

// writeAuthStatusJSON renders the legacy JSON shape. Encoder (not
// MarshalIndent) keeps the trailing-newline behaviour of printJSON.
func writeAuthStatusJSON(w io.Writer, out map[string]any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

// writeAuthStatusHuman renders the status block. Colour is applied only
// when the destination is a real TTY and NO_COLOR is unset, reusing
// paletteFor's discipline so the rules match `retab --help`.
//
// Layout (fields are omitted when absent):
//
//	Logged in as <preview>
//	Source:       <credential_source>
//	Environment:  <resolved_environment>
//	Status:       <valid|invalid|not authenticated>
//
//	Configured environments:
//	  <slug>  <preview>
func writeAuthStatusHuman(w io.Writer, out map[string]any) error {
	s := paletteFor(w)

	// Accept both the v2 keys and the legacy keys (api_key_preview/source/
	// valid) so existing callers and tests keep rendering correctly.
	preview, _ := out["active_credential"].(string)
	if preview == "" {
		preview, _ = out["api_key_preview"].(string)
	}
	source, _ := out["credential_source"].(string)
	if source == "" {
		source, _ = out["source"].(string)
	}

	if preview != "" {
		fmt.Fprintf(w, "Logged in as %s%s%s\n", s.brand, preview, s.reset)
	} else {
		fmt.Fprintln(w, "Not logged in")
	}

	if source == "" {
		source = "none"
	}
	fmt.Fprintf(w, "%sSource:%s       %s\n", s.dim, s.reset, source)

	if env, ok := out["resolved_environment"].(string); ok && env != "" {
		fmt.Fprintf(w, "%sEnvironment:%s  %s\n", s.dim, s.reset, env)
	}

	// Status line. Prefer the v2 `authenticated` boolean; fall back to the
	// legacy `valid` key. Absent means we never probed.
	var status string
	switch {
	case boolKey(out, "authenticated"):
		status = "valid"
	case keyPresent(out, "authenticated") || keyPresent(out, "valid"):
		if boolKey(out, "valid") {
			status = "valid"
		} else {
			status = "invalid"
		}
	default:
		status = "not authenticated"
	}
	fmt.Fprintf(w, "%sStatus:%s       %s\n", s.dim, s.reset, status)

	// Configured environments block, if any.
	if profiles, ok := out["profiles"].(map[string]any); ok && len(profiles) > 0 {
		slugs := make([]string, 0, len(profiles))
		for slug := range profiles {
			slugs = append(slugs, slug)
		}
		sort.Strings(slugs)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Configured environments:")
		for _, slug := range slugs {
			entry, _ := profiles[slug].(map[string]any)
			p, _ := entry["api_key_preview"].(string)
			fmt.Fprintf(w, "  %s%s%s  %s\n", s.dim, slug, s.reset, p)
		}
	}
	return nil
}

// boolKey returns out[key] as a bool (false if absent or not a bool).
func boolKey(out map[string]any, key string) bool {
	b, _ := out[key].(bool)
	return b
}

// keyPresent reports whether key exists in out.
func keyPresent(out map[string]any, key string) bool {
	_, ok := out[key]
	return ok
}

// isTerminalWriter mirrors paletteFor's TTY check — true only when w is
// an *os.File pointing at a terminal. bytes.Buffer, pipes, and redirected
// files all return false, which is what we want for "render JSON for
// machines, human block for humans".
func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// redactKey masks a credential for display: the first 4 and last 4
// characters are kept, the middle replaced by a FIXED-WIDTH asterisk run.
//
// The width is capped at redactMaskWidth rather than len(key)-8. An OAuth
// access token is a ~900-1200 character JWT; the old len-8 mask dumped a
// full screen of asterisks into `--debug` output and `auth status`, and
// reproduced the exact length of the secret. A fixed mask is both readable
// and leaks nothing about the credential's size.
func redactKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	const redactMaskWidth = 8
	mask := len(key) - 8
	if mask > redactMaskWidth {
		mask = redactMaskWidth
	}
	return key[:4] + strings.Repeat("*", mask) + key[len(key)-4:]
}

func promptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func init() {
	authLoginCmd.Flags().String("api-key", "", "skip the browser flow and store this API key (also reads RETAB_API_KEY)")
	authLoginCmd.Flags().String("base-url", "", "override the default API base URL")
	authLoginCmd.Flags().Bool("browser", true, "open a browser for OAuth login (set --browser=false to prompt for an API key)")
	authLoginCmd.Flags().Bool("default", false, "set the stored profile as the local default environment")

	// `--json` forces JSON output even on a TTY. Without it, status auto-
	// detects: TTY → human block, pipe/redirect → JSON (same shape as today).
	authStatusCmd.Flags().Bool("json", false, "emit JSON even when stdout is a TTY (default: human-readable on TTY, JSON when piped)")

	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd, authUseCmd)
	rootCmd.AddCommand(authCmd)
}
