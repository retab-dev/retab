package cmd

// `retab org list` / `retab org switch` — list and switch the WorkOS
// organization the CLI's OAuth session is scoped to. Mirrors the `env`
// resource: an organization is a scope you select, exactly like an environment
// (which itself lives inside an organization).
//
// Background: a CLI OAuth login (see oauth.go) lands you in whichever
// organization WorkOS resolved for the browser session, and re-running
// `retab auth login` cannot change that — WorkOS silently reuses the existing
// browser session and its last-selected org. Switching organizations is a
// token operation, not a browser one: the backend exchanges the stored refresh
// token for a fresh token pair scoped to the target org (the dashboard does the
// same via WorkOS `refreshSession({ organizationId })`). No browser, no
// re-login. The user must already be a member of the target org; the backend
// surfaces a clear error otherwise.

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// cliOrganization mirrors one entry of GET /v1/auth/cli/organizations.
type cliOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cliOrganizationsResponse struct {
	Data []cliOrganization `json:"data"`
}

// cliSwitchOrganizationResponse mirrors POST /v1/auth/cli/switch-organization —
// the subset of the WorkOS token response the CLI persists.
type cliSwitchOrganizationResponse struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	TokenType      string `json:"token_type"`
	ExpiresIn      int    `json:"expires_in"`
	OrganizationID string `json:"organization_id"`
}

// fetchCLIOrganizations lists the organizations the active credential's user
// belongs to, using whatever credential `cliJSONRequestInto` resolves (API key
// or OAuth). Used by `retab org list`.
func fetchCLIOrganizations(cmd *cobra.Command) ([]cliOrganization, error) {
	var resp cliOrganizationsResponse
	if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/auth/cli/organizations", nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// listOrganizationsWithToken lists the user's organizations using the RAW OAuth
// access token, bypassing the environment-scoped dashboard-context wrapper that
// `cliJSONRequestInto` applies. Listing organizations is a user-level operation
// — it is not scoped to an environment — and the org being switched FROM may
// have a stale/invalid selected environment, so minting an env-scoped token
// here would be both wrong and fragile. Mirrors `selectOAuthLoginEnvironment`.
//
// The token provider transparently refreshes (and persists the rotated refresh
// token) when the access token is near expiry, so callers should reload config
// afterwards before using the refresh token elsewhere.
func listOrganizationsWithToken(ctx context.Context, baseURL string, oauth *oauthTokens) ([]cliOrganization, error) {
	token, err := makeOAuthTokenProvider(oauth)(ctx)
	if err != nil {
		return nil, err
	}
	var resp cliOrganizationsResponse
	if err := doCLIJSONRequest(
		ctx,
		http.DefaultClient,
		canonicalAPIBaseURL(baseURL),
		http.MethodGet,
		"/v1/auth/cli/organizations",
		nil,
		nil,
		"",
		token,
		&resp,
	); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// organizationDisplay renders "<name> (<id>)" when a name is known, else the id.
func organizationDisplay(id, name string) string {
	if strings.TrimSpace(name) != "" {
		return fmt.Sprintf("%s (%s)", name, id)
	}
	return id
}

// resolveTargetOrganization maps the user's argument (an org id or a name) to a
// concrete organization id. Exact id match wins; otherwise a case-insensitive
// name match is required to be unique. An `org_`-prefixed argument that is not
// in the list is allowed through unchanged — the listing may be incomplete and
// the backend still enforces membership — so the only hard failures are an
// ambiguous name or a non-id string that matches nothing.
func resolveTargetOrganization(input string, orgs []cliOrganization) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("an organization id or name is required")
	}
	for _, o := range orgs {
		if o.ID == trimmed {
			return o.ID, nil
		}
	}
	var matches []cliOrganization
	for _, o := range orgs {
		if strings.EqualFold(strings.TrimSpace(o.Name), trimmed) {
			matches = append(matches, o)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0].ID, nil
	case 0:
		if strings.HasPrefix(trimmed, "org_") {
			return trimmed, nil
		}
		return "", fmt.Errorf(
			"no organization matches %q. Run `retab org list` to see the organizations you belong to",
			trimmed,
		)
	default:
		return "", fmt.Errorf(
			"organization name %q is ambiguous across %d organizations; pass the organization id instead",
			trimmed, len(matches),
		)
	}
}

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "List and switch the active WorkOS organization",
	Long: `Inspect and switch the organization your CLI session is scoped to.

Organizations are the top-level scope above environments. A login lands you in
whichever organization WorkOS resolved for your session; ` + "`retab org switch`" + `
moves the session to a different organization you belong to without a browser
re-login.`,
	Example: `  retab org list
  retab org switch "Acme Inc"`,
}

var orgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the organizations you can switch into",
	Long: `List every WorkOS organization the current login is a member of.

The organization the active session is currently scoped to is marked
"(current)". Pass an id or name to ` + "`retab org switch`" + ` to move into it.`,
	Example: `  retab org list`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		orgs, err := fetchCLIOrganizations(cmd)
		if err != nil {
			return err
		}
		if len(orgs) == 0 {
			fmt.Fprintln(os.Stderr, "You do not belong to any organizations.")
			return nil
		}

		// Best-effort current-org marker; never fails the listing.
		var current cliAuthOrganization
		_ = cliJSONRequestInto(cmd, http.MethodGet, "/v1/auth/organization", nil, nil, &current)

		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		for _, o := range orgs {
			marker := ""
			if current.ID != "" && o.ID == current.ID {
				marker = "(current)"
			}
			name := o.Name
			if name == "" {
				name = "-"
			}
			if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", o.ID, name, marker); err != nil {
				return err
			}
		}
		return tw.Flush()
	}),
}

var orgSwitchCmd = &cobra.Command{
	Use:   "switch <organization-id-or-name>",
	Short: "Switch the active organization without re-logging in",
	Long: `Switch the organization your CLI session is scoped to.

Exchanges the stored OAuth refresh token for a fresh token pair scoped to the
target organization — no browser and no re-login. You must already be a member
of the target organization. After switching, the selected environment is reset
to the new organization's default (environments are per-organization).

Only works for OAuth logins. API-key sessions are already bound to a single
organization; to use a different org's key, run ` + "`retab auth login --api-key <key>`" + `.

The argument accepts either an organization id (org_...) or a name; run
` + "`retab org list`" + ` to see what you can switch into.`,
	Example: `  # By name
  retab org switch "Acme Inc"

  # By id
  retab org switch org_01HXYZ...`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		cfg, _ := loadConfig()
		if cfg.OAuth == nil || strings.TrimSpace(cfg.OAuth.RefreshToken) == "" {
			return fmt.Errorf(
				"organization switching requires an OAuth login. Run `retab auth login` first " +
					"(API-key sessions are already scoped to one organization)",
			)
		}

		ctx, cancel := ctxFor(cmd)
		defer cancel()

		baseURL, err := resolvedAuthStatusBaseURL(cmd, cfg)
		if err != nil {
			return err
		}
		canonical := canonicalAPIBaseURL(baseURL)

		// Resolve the target org. A bare org id is used as-is (the backend
		// enforces membership); a name must be resolved against the user's org
		// list. Listing uses the raw OAuth token and may rotate+persist the
		// refresh token, so reload config afterwards before reading it.
		target := strings.TrimSpace(args[0])
		var orgs []cliOrganization
		if !strings.HasPrefix(target, "org_") {
			orgs, err = listOrganizationsWithToken(ctx, baseURL, cfg.OAuth)
			if err != nil {
				return fmt.Errorf("could not list organizations to resolve %q: %w", target, err)
			}
			cfg, _ = loadConfig()
			if cfg.OAuth == nil || strings.TrimSpace(cfg.OAuth.RefreshToken) == "" {
				return fmt.Errorf("OAuth session was cleared; run `retab auth login`")
			}
			resolved, resolveErr := resolveTargetOrganization(target, orgs)
			if resolveErr != nil {
				return resolveErr
			}
			target = resolved
		}

		// Exchange the refresh token for a token pair scoped to the new org. The
		// endpoint is public — the refresh token IS the credential — so no
		// bearer/api-key is attached.
		var switched cliSwitchOrganizationResponse
		body := map[string]string{
			"refresh_token":   cfg.OAuth.RefreshToken,
			"organization_id": target,
		}
		if err := doCLIJSONRequest(
			ctx,
			http.DefaultClient,
			canonical,
			http.MethodPost,
			"/v1/auth/cli/switch-organization",
			nil,
			body,
			"",
			"",
			&switched,
		); err != nil {
			return err
		}
		if switched.AccessToken == "" {
			return fmt.Errorf("organization switch returned no access token")
		}

		// Persist the new tokens in place, preserving discovery fields so the
		// transparent refresh path keeps working.
		cfg.OAuth.AccessToken = switched.AccessToken
		if switched.RefreshToken != "" {
			cfg.OAuth.RefreshToken = switched.RefreshToken
		}
		if switched.TokenType != "" {
			cfg.OAuth.TokenType = switched.TokenType
		} else {
			cfg.OAuth.TokenType = "Bearer"
		}
		ttl := time.Duration(switched.ExpiresIn) * time.Second
		if ttl <= 0 {
			ttl = 10 * time.Minute
		}
		cfg.OAuth.ExpiresAt = time.Now().Add(ttl)

		// Environments are per-organization, so the previously-selected
		// environment id no longer belongs to this session. Re-select the new
		// org's default before saving. A failure here is non-fatal: the tokens
		// are already valid, the user just needs to pick an environment.
		cfg.EnvironmentID = ""
		cfg.EnvironmentType = ""
		environment, envErr := selectOAuthLoginEnvironment(ctx, baseURL, cfg.OAuth, "")
		if environment != nil {
			cfg.EnvironmentID = environment.ID
			cfg.EnvironmentType = string(environment.Type)
		}

		if err := saveConfig(cfg); err != nil {
			return err
		}

		display := target
		for _, o := range orgs {
			if o.ID == target {
				display = organizationDisplay(o.ID, o.Name)
				break
			}
		}
		fmt.Fprintf(os.Stderr, "Switched to organization %s\n", display)
		if envErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not resolve environments for the new organization: %v\n", envErr)
			fmt.Fprintln(os.Stderr, "  Run `retab env list` and `retab env switch <environment-id-or-name>` before environment-scoped commands.")
			return nil
		}
		if environment != nil {
			fmt.Fprintf(os.Stderr, "Environment: %s (%s)\n", environment.Name, environment.ID)
		}
		return nil
	}),
}

func init() {
	orgCmd.AddCommand(orgListCmd, orgSwitchCmd)
	rootCmd.AddCommand(orgCmd)
}
