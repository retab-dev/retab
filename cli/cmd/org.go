package cmd

// `retab org list` / `retab org switch` — list and switch the WorkOS
// organization the CLI's OAuth session is scoped to. Mirrors the `env`
// resource: an organization is a scope you select, exactly like an environment
// (which itself lives inside an organization).
//
// Background: a CLI OAuth login (see oauth.go) uses the WorkOS User Management
// device flow and lands you in whichever organization WorkOS resolved for the
// session. Device-flow tokens ARE User Management session tokens, so switching
// is browserless: `retab org switch` POSTs the stored refresh_token + target
// organization_id to the backend's `/v1/auth/cli/switch-organization` endpoint,
// which performs the confidential `refresh + organization_id` exchange
// server-side and returns a fresh token pair already scoped to the target org.
// The user must be a member of the target org (run `retab org list` to confirm).

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// cliSwitchOrganizationRequest is the POST /v1/auth/cli/switch-organization
// body. The refresh token IS the credential, so the route is public.
type cliSwitchOrganizationRequest struct {
	RefreshToken   string `json:"refresh_token"`
	OrganizationID string `json:"organization_id"`
}

// cliSwitchOrganizationResponse mirrors the subset of the WorkOS token response
// the backend returns to the CLI.
type cliSwitchOrganizationResponse struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	TokenType      string `json:"token_type"`
	ExpiresIn      *int   `json:"expires_in,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
}

// switchOrganizationViaBackend POSTs the stored refresh token + target org to
// the backend's confidential switch endpoint and returns the fresh token pair.
// It uses the RAW request (not the env-scoped dashboard-context wrapper):
// switching organizations is a user-level operation and the route is public
// (the refresh token authenticates it), so no bearer token is attached.
func switchOrganizationViaBackend(ctx context.Context, baseURL, refreshToken, organizationID string) (cliSwitchOrganizationResponse, error) {
	body := cliSwitchOrganizationRequest{RefreshToken: refreshToken, OrganizationID: organizationID}
	var resp cliSwitchOrganizationResponse
	if err := doCLIJSONRequest(
		ctx,
		http.DefaultClient,
		canonicalAPIBaseURL(baseURL),
		http.MethodPost,
		"/v1/auth/cli/switch-organization",
		nil,
		body,
		"",
		"",
		&resp,
	); err != nil {
		return cliSwitchOrganizationResponse{}, err
	}
	return resp, nil
}

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
moves the session to a different organization you belong to via a browserless
token exchange through the Retab backend.`,
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
	Short: "Switch the active organization (browserless token exchange)",
	Long: `Switch the organization your CLI session is scoped to.

Exchanges your stored OAuth refresh token for a token pair scoped to the target
organization, via the Retab backend's confidential WorkOS exchange — no browser
re-authentication required. You must already be a member of the target
organization. After switching, the selected environment is reset to the new
organization's default (environments are per-organization).

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

		baseURL, err := resolvedAuthStatusBaseURL(cmd, cfg, nil)
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

		// listOrganizationsWithToken / fetchCurrentOrganizationWithToken may
		// have transparently refreshed and rotated the stored refresh token, so
		// reload to read the freshest one before handing it to the backend.
		cfg, _ = loadConfig()
		if cfg.OAuth == nil || strings.TrimSpace(cfg.OAuth.RefreshToken) == "" {
			return fmt.Errorf("OAuth session was cleared; run `retab auth login`")
		}

		// Switch by exchanging the refresh token through the backend's
		// confidential WorkOS endpoint. Device-flow refresh tokens ARE User
		// Management session tokens, so this returns a fresh pair already scoped
		// to the target org — no browser re-auth. The endpoint uses the
		// SESSION's base URL so we stay on whatever deployment this session
		// targets.
		switched, err := switchOrganizationViaBackend(ctx, baseURL, cfg.OAuth.RefreshToken, target)
		if err != nil {
			return err
		}
		if strings.TrimSpace(switched.AccessToken) == "" {
			return fmt.Errorf("organization switch returned no access token")
		}

		// Confirm the new token actually landed in the requested org. The
		// backend scopes the exchange to it (and `org list` already proved
		// membership), but if it landed elsewhere, say so plainly rather than
		// reporting a switch that did not happen.
		if landed := accessTokenOrgID(switched.AccessToken); landed != "" && landed != target {
			fmt.Fprintf(os.Stderr,
				"warning: requested organization %s but the new session is scoped to %s\n",
				target, landed)
			target = landed
		}

		// Persist the freshly-minted tokens, preserving ClientID +
		// WorkosAPIBaseURL from the prior session so transparent refresh keeps
		// working (the backend response carries only the token pair).
		ttl := 10 * time.Minute
		if switched.ExpiresIn != nil && *switched.ExpiresIn > 0 {
			ttl = time.Duration(*switched.ExpiresIn) * time.Second
		}
		refreshToken := switched.RefreshToken
		if refreshToken == "" {
			refreshToken = cfg.OAuth.RefreshToken
		}
		cfg.OAuth = &oauthTokens{
			AccessToken:      switched.AccessToken,
			RefreshToken:     refreshToken,
			TokenType:        switched.TokenType,
			ExpiresAt:        time.Now().Add(ttl),
			ClientID:         cfg.OAuth.ClientID,
			WorkosAPIBaseURL: cfg.OAuth.WorkosAPIBaseURL,
			OrganizationID:   target,
		}

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
