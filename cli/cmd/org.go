package cmd

// `retab org list` / `retab org switch` — list and switch the WorkOS
// organization the CLI's OAuth session is scoped to. Mirrors the `env`
// resource: an organization is a scope you select, exactly like an environment
// (which itself lives inside an organization).
//
// Background: a CLI OAuth login (see oauth.go) authenticates as a PUBLIC OAuth
// app (PKCE, no client secret) and lands you in whichever organization WorkOS
// resolved for the session. Unlike the dashboard — a confidential client that
// can call WorkOS `refreshSession({ organizationId })` server-side — a public
// client's refresh token CANNOT be exchanged for a different org (WorkOS
// rejects it with invalid_client / unauthorized_client). So org switching has
// no browserless path: `retab org switch` re-runs the browser auth flow with
// `organization_id` + `provider=authkit`, which makes AuthKit auto-select the
// target org and mint a fresh token already scoped to it. The user must be a
// member of the target org (run `retab org list` to confirm).

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

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

// orgListRow is one rendered row of `retab org list`. The `current` flag
// replaces the "(current)" table marker for json/csv consumers.
type orgListRow struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Current bool   `json:"current"`
}

func orgRowCell(row any, key string) string {
	o, ok := row.(orgListRow)
	if !ok {
		return ""
	}
	switch key {
	case "id":
		return o.ID
	case "name":
		if o.Name == "" {
			return "-"
		}
		return o.Name
	case "current":
		if o.Current {
			return "(current)"
		}
		return ""
	default:
		return ""
	}
}

var orgListColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return orgRowCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return orgRowCell(row, "name") }},
	{Header: "CURRENT", Extract: func(row any) string { return orgRowCell(row, "current") }},
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

func fetchCurrentOrganizationWithToken(ctx context.Context, baseURL string, oauth *oauthTokens) (cliAuthOrganization, error) {
	token, err := makeOAuthTokenProvider(oauth)(ctx)
	if err != nil {
		return cliAuthOrganization{}, err
	}
	var organization cliAuthOrganization
	if err := doCLIJSONRequest(
		ctx,
		http.DefaultClient,
		canonicalAPIBaseURL(baseURL),
		http.MethodGet,
		"/v1/auth/organization",
		nil,
		nil,
		"",
		token,
		&organization,
	); err != nil {
		return cliAuthOrganization{}, err
	}
	return organization, nil
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
moves the session to a different organization you belong to by re-authenticating
in the browser scoped to that organization.`,
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

		// Mirror every other list command: honor --output (and the
		// auto→JSON-when-piped convention used by `env list` / `projects
		// list`). A bare TTY invocation still renders the human table.
		format, err := ResolveOutputFormat(cmd, os.Stdout)
		if err != nil {
			return err
		}

		rows := make([]any, 0, len(orgs))
		for _, o := range orgs {
			rows = append(rows, orgListRow{
				ID:      o.ID,
				Name:    o.Name,
				Current: current.ID != "" && o.ID == current.ID,
			})
		}

		if format == OutputJSON {
			return printJSON(map[string]any{"data": rows})
		}
		if format == OutputCSV {
			return renderAutoCSV(os.Stdout, rows, orgListColumns)
		}
		return renderAutoTable(os.Stdout, rows, orgListColumns)
	}),
}

var orgSwitchCmd = &cobra.Command{
	Use:   "switch <organization-id-or-name>",
	Short: "Switch the active organization (re-authenticates in the browser)",
	Long: `Switch the organization your CLI session is scoped to.

Opens a browser to re-authenticate scoped to the target organization, then
persists the fresh token pair. A CLI logs in as a public OAuth client, whose
refresh token WorkOS will not exchange for a different organization — so unlike
the dashboard there is no token-only switch, and the browser step is required.
You must already be a member of the target organization. After switching, the
selected environment is reset to the new organization's default (environments
are per-organization).

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
		current, currentErr := fetchCurrentOrganizationWithToken(ctx, baseURL, cfg.OAuth)
		if currentErr == nil && current.ID == target {
			display := organizationDisplay(current.ID, current.Name)
			fmt.Fprintf(os.Stderr, "Already using organization %s\n", display)
			return nil
		}

		// Switch by RE-AUTHENTICATING scoped to the target org. A CLI logs in as
		// a public OAuth app (PKCE, no secret); WorkOS will not exchange that
		// kind of refresh token for a different organization, so there is no
		// browserless path. Instead we run the same browser flow as `auth login`
		// with organization_id + provider=authkit, which makes AuthKit
		// auto-select the target org and mint a token already scoped to it.
		// Discovery uses the SESSION's base URL (not the login default) so we
		// stay on whatever deployment this session targets.
		disc, err := fetchOAuthDiscovery(ctx, baseURL)
		if err != nil {
			return fmt.Errorf("OAuth discovery failed: %w", err)
		}
		tokens, err := runLoginFlow(ctx, disc, browserOpener, target)
		if err != nil {
			return err
		}
		if tokens == nil || strings.TrimSpace(tokens.AccessToken) == "" {
			return fmt.Errorf("organization switch returned no access token")
		}

		// Confirm the session actually landed in the requested org. AuthKit
		// auto-selects it when the user is a member (and `org list` already
		// proved membership), but if it landed elsewhere, say so plainly rather
		// than reporting a switch that did not happen.
		if landed := accessTokenOrgID(tokens.AccessToken); landed != "" && landed != target {
			fmt.Fprintf(os.Stderr,
				"warning: requested organization %s but the new session is scoped to %s\n",
				target, landed)
			target = landed
		}

		// Persist the freshly-minted tokens. runLoginFlow already stamped
		// AuthKitDomain/ClientID/ExpiresAt from discovery, so the transparent
		// refresh path keeps working.
		cfg.OAuth = tokens

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
