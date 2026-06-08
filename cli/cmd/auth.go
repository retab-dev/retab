package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
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

		// Explicit API-key path. Direct, no browser. Backward compatible
		// with the original login UX.
		if apiKey != "" || (!useBrowser && os.Getenv("RETAB_API_KEY") != "") {
			return runAPIKeyLogin(apiKey, baseURL)
		}

		// If `--no-browser` was requested and no key was given, prompt.
		if !useBrowser {
			prompted, err := promptSecret("API key (leave blank to use browser flow): ")
			if err != nil {
				return err
			}
			if strings.TrimSpace(prompted) != "" {
				return runAPIKeyLogin(prompted, baseURL)
			}
			// fall through to browser flow
		}

		ctx, cancel := ctxFor(cmd)
		defer cancel()

		// Discovery + browser OAuth flow. A fresh login without an
		// explicit override should reset the CLI back to production,
		// not inherit a stale local/staging base_url from a previous
		// account or dev session.
		loginBaseURL := configuredLoginBaseURL(baseURL)
		disc, err := fetchOAuthDiscovery(ctx, loginBaseURL)
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
		cfg.BaseURL = stripLegacyV1Suffix(loginBaseURL)
		environment, envErr := selectOAuthLoginEnvironment(ctx, loginBaseURL, tokens, cfg.EnvironmentID)
		if environment != nil {
			cfg.EnvironmentID = environment.ID
			// Persist the type so the offline production-confirmation gate
			// can tell whether this OAuth session targets production.
			cfg.EnvironmentType = string(environment.Type)
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		path, _ := configPath()
		fmt.Fprintf(os.Stderr, "Logged in. Saved OAuth tokens to %s\n", path)
		if envErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not resolve environments after login: %v\n", envErr)
			fmt.Fprintln(os.Stderr, "  Run `retab env list` and `retab env switch <environment-id-or-name>` before environment-scoped commands.")
			return nil
		}
		if environment != nil {
			fmt.Fprintf(os.Stderr, "Environment: %s (%s)\n", environment.Name, environment.ID)
		}
		return nil
	}),
}

func configuredLoginBaseURL(flagBaseURL string) string {
	if flagBaseURL != "" {
		return flagBaseURL
	}
	if envBaseURL := os.Getenv("RETAB_API_BASE_URL"); envBaseURL != "" {
		return envBaseURL
	}
	if envBaseURL := os.Getenv("RETAB_BASE_URL"); envBaseURL != "" {
		return envBaseURL
	}
	return defaultAPIBaseURL
}

func selectOAuthLoginEnvironment(
	ctx context.Context,
	baseURL string,
	tokens *oauthTokens,
	currentEnvironmentID string,
) (*cliEnvironment, error) {
	if tokens == nil || strings.TrimSpace(tokens.AccessToken) == "" {
		return nil, fmt.Errorf("OAuth access token is empty")
	}
	var environments cliPaginatedList[cliEnvironment]
	err := doCLIJSONRequest(
		ctx,
		http.DefaultClient,
		canonicalAPIBaseURL(baseURL),
		http.MethodGet,
		"/v1/environments",
		nil,
		nil,
		"",
		tokens.AccessToken,
		&environments,
	)
	if err != nil {
		return nil, err
	}
	environment := chooseLoginEnvironment(currentEnvironmentID, &environments)
	if environment == nil {
		return nil, fmt.Errorf("no environments are available for this organization")
	}
	return environment, nil
}

func chooseLoginEnvironment(currentEnvironmentID string, list *cliPaginatedList[cliEnvironment]) *cliEnvironment {
	if list == nil {
		return nil
	}
	if strings.TrimSpace(currentEnvironmentID) != "" {
		for i := range list.Data {
			if list.Data[i].ID == currentEnvironmentID {
				return &list.Data[i]
			}
		}
	}
	for i := range list.Data {
		if list.Data[i].IsDefault != nil && *list.Data[i].IsDefault {
			return &list.Data[i]
		}
	}
	for i := range list.Data {
		if list.Data[i].Type == cliEnvironmentTypeProduction {
			return &list.Data[i]
		}
	}
	if len(list.Data) > 0 {
		return &list.Data[0]
	}
	return nil
}

// runAPIKeyLogin persists an API key — the legacy auth path. Kept first-
// class so anyone scripting `retab auth login --api-key …` keeps working.
func runAPIKeyLogin(apiKey, baseURL string) error {
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
	cfg, _ := loadConfig()
	cfg.APIKey = apiKey
	// Wipe stale OAuth state — explicit key login is the user's intent.
	cfg.OAuth = nil
	cfg.BaseURL = stripLegacyV1Suffix(configuredLoginBaseURL(baseURL))
	if err := saveConfig(cfg); err != nil {
		return err
	}
	path, _ := configPath()
	fmt.Fprintf(os.Stderr, "Saved API key to %s\n", path)
	return nil
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

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show which credentials the CLI is using and verify them",
	Long: `Report the credential the CLI would use for the next request, and
verify it works by making a best-effort API call (a tiny ` + "`workflows list`" + `
probe). Output is JSON with the credential ` + "`source`" + ` (flag, env, or
config), a redacted ` + "`api_key_preview`" + `, the effective ` + "`base_url`" + `, and a
` + "`valid`" + ` boolean reflecting the probe.

Useful for debugging "why is the wrong account being used?" — the
` + "`source`" + ` field disambiguates --api-key vs RETAB_API_KEY vs the config
file. Resolution order: --api-key > RETAB_API_KEY > ~/.retab/config.json.

Output formatting follows the global --output flag: ` + "`--output table`" + `
renders a KEY/VALUE table view, ` + "`--output json`" + ` forces JSON even on a
TTY, and the default auto-detects (JSON for pipes / redirects, a
compact human block for interactive terminals).`,
	Example: `  # Quick check
  retab auth status

  # In a script — assert authenticated and key is valid
  retab auth status | jq -e '.valid == true'

  # Verify a specific key without persisting it
  retab --api-key=sk_test_xyz auth status`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		flagKey, _ := cmd.Root().PersistentFlags().GetString("api-key")
		envKey := os.Getenv("RETAB_API_KEY")
		cfg, _ := loadConfig()

		var source string
		switch {
		case flagKey != "":
			source = "--api-key flag"
		case envKey != "":
			source = "RETAB_API_KEY env"
		case cfg.OAuth != nil && cfg.OAuth.AccessToken != "":
			source = "~/.retab/config.json (oauth)"
		case cfg.APIKey != "":
			source = "~/.retab/config.json (api_key)"
		}
		baseURL, err := resolvedAuthStatusBaseURL(cmd, cfg)
		if err != nil {
			return err
		}

		out := map[string]any{
			"authenticated": source != "",
			"base_url":      baseURL,
			"source":        source,
		}
		if cfg.OAuth != nil && cfg.OAuth.AccessToken != "" {
			out["oauth"] = map[string]any{
				"authkit_domain": cfg.OAuth.AuthKitDomain,
				"expires_at":     cfg.OAuth.ExpiresAt,
				"has_refresh":    cfg.OAuth.RefreshToken != "",
			}
		}
		if source == "~/.retab/config.json (oauth)" {
			addSelectedEnvironmentStatus(cmd, cfg, baseURL, out)
		}
		if flagKey != "" || envKey != "" || cfg.APIKey != "" {
			key := flagKey
			if key == "" {
				key = envKey
			}
			if key == "" {
				key = cfg.APIKey
			}
			out["api_key_preview"] = redactKey(key)
		}

		jsonOnly, _ := cmd.Flags().GetBool("json")
		// `--output` is the global formatting flag every command honours.
		// When explicitly set to "json" or "table" it wins over the
		// command-local --json flag and the TTY auto-detect — anything
		// else (empty / "auto") falls through to the existing
		// jsonOnly + TTY behaviour so existing scripts and the human
		// 3-line block keep their current rules.
		outputFormat, err := resolveAuthOutputFormat(cmd)
		if err != nil {
			return err
		}

		if source == "" {
			out["hint"] = "run `retab auth login` to authenticate"
			return writeAuthStatusWithFormat(cmd.OutOrStdout(), out, jsonOnly, outputFormat)
		}

		// Best-effort verification. Use the auth status endpoint instead
		// of a workflow route so fresh OAuth logins do not need an
		// environment selection just to prove the credential works.
		if err := probeAuthStatus(cmd); err != nil {
			out["valid"] = false
			out["error"] = err.Error()
		} else {
			out["valid"] = true
		}
		addAuthOrganizationStatus(cmd, out)
		return writeAuthStatusWithFormat(cmd.OutOrStdout(), out, jsonOnly, outputFormat)
	}),
}

func resolvedAuthStatusBaseURL(cmd *cobra.Command, cfg retabConfig) (string, error) {
	flagBaseURL, _ := cmd.Root().PersistentFlags().GetString("base-url")
	baseURL := flagBaseURL
	if baseURL == "" {
		baseURL = os.Getenv("RETAB_API_BASE_URL")
	}
	if baseURL == "" {
		baseURL = os.Getenv("RETAB_BASE_URL")
	}
	if err := validateBaseURL(baseURL); err != nil {
		return "", err
	}
	if baseURL == "" {
		baseURL = cfg.BaseURL
	}
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	return stripLegacyV1Suffix(baseURL), nil
}

func addSelectedEnvironmentStatus(cmd *cobra.Command, cfg retabConfig, baseURL string, out map[string]any) {
	environmentID, source := selectedEnvironmentIDWithSource(cmd, cfg)
	if environmentID == "" {
		out["environment"] = nil
		return
	}

	environmentOut := map[string]any{
		"id":     environmentID,
		"source": source,
	}
	environment, err := getSelectedEnvironmentForAuthStatus(cmd, cfg, baseURL, environmentID)
	if err != nil {
		environmentOut["error"] = err.Error()
		out["environment"] = environmentOut
		return
	}
	environmentOut["name"] = environment.Name
	environmentOut["type"] = string(environment.Type)
	if environment.IsDefault != nil {
		environmentOut["is_default"] = *environment.IsDefault
	}
	out["environment"] = environmentOut
}

func getSelectedEnvironmentForAuthStatus(
	cmd *cobra.Command,
	cfg retabConfig,
	baseURL string,
	environmentID string,
) (*cliEnvironment, error) {
	if cfg.OAuth == nil || strings.TrimSpace(cfg.OAuth.AccessToken) == "" {
		return nil, fmt.Errorf("OAuth access token is empty")
	}
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	token, err := makeOAuthTokenProvider(cfg.OAuth)(ctx)
	if err != nil {
		return nil, err
	}
	var environment cliEnvironment
	err = doCLIJSONRequest(
		ctx,
		http.DefaultClient,
		canonicalAPIBaseURL(baseURL),
		http.MethodGet,
		"/v1/environments/"+url.PathEscape(environmentID),
		nil,
		nil,
		"",
		token,
		&environment,
	)
	if err != nil {
		return nil, err
	}
	return &environment, nil
}

func probeAuthStatus(cmd *cobra.Command) error {
	return cliJSONRequestInto(cmd, http.MethodGet, "/v1/auth/status", nil, nil, nil)
}

// cliAuthOrganization mirrors the /v1/auth/organization response: the org id
// is always present, the WorkOS-resolved name is best-effort (absent when the
// lookup degraded server-side).
type cliAuthOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// addAuthOrganizationStatus resolves the organization bound to the active
// credential and records it on the status payload. Best-effort: any failure
// (network, server) is swallowed so `auth status` still renders the rest of
// the block — the org line is informational, not a verification signal.
func addAuthOrganizationStatus(cmd *cobra.Command, out map[string]any) {
	var organization cliAuthOrganization
	if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/auth/organization", nil, nil, &organization); err != nil {
		return
	}
	if organization.ID == "" {
		return
	}
	organizationOut := map[string]any{"id": organization.ID}
	if organization.Name != "" {
		organizationOut["name"] = organization.Name
	}
	out["organization"] = organizationOut
}

// resolveAuthOutputFormat reads the global --output persistent flag.
// Returns OutputAuto for empty / "auto" so the caller falls back to the
// existing jsonOnly + TTY logic; explicit "json" / "table" produce the
// matching OutputFormat. Unknown values are already rejected at parse
// time by outputFlagValue in root.go — defensive-only here.
func resolveAuthOutputFormat(cmd *cobra.Command) (OutputFormat, error) {
	var raw string
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	switch raw {
	case "", "auto":
		return OutputAuto, nil
	case string(OutputJSON):
		return OutputJSON, nil
	case string(OutputTable):
		return OutputTable, nil
	default:
		return "", fmt.Errorf("invalid --output value %q (want: json | table | auto)", raw)
	}
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
//
// This shim preserves the pre-existing call signature for tests that
// don't exercise the global --output flag; the runtime path goes
// through writeAuthStatusWithFormat which adds OutputTable routing.
func writeAuthStatus(w io.Writer, out map[string]any, jsonOnly bool) error {
	return writeAuthStatusWithFormat(w, out, jsonOnly, OutputAuto)
}

// writeAuthStatusWithFormat routes the status payload to JSON, table,
// or human-block rendering. The global --output flag (OutputJSON /
// OutputTable) wins when set; OutputAuto falls back to the existing
// rule set (jsonOnly flag OR non-TTY → JSON, TTY → human block).
//
// The table renderer is the new path added for issue #7 — `retab
// --output table auth status` was previously silently emitting JSON
// because nothing consulted the global flag here.
func writeAuthStatusWithFormat(w io.Writer, out map[string]any, jsonOnly bool, format OutputFormat) error {
	switch format {
	case OutputJSON:
		return writeAuthStatusJSON(w, out)
	case OutputTable:
		return writeAuthStatusTable(w, out)
	}
	// OutputAuto (or unset) — preserve the historical behaviour.
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

// writeAuthStatusHuman renders the human status block. Colour is
// applied only when the destination is a real TTY and NO_COLOR is unset,
// reusing paletteFor's discipline so the rules match `retab --help`.
//
// Layout:
//
//	Logged in as <preview>
//	Source:  <source>
//	Base URL:  <api base url>
//	Organization:  <name> (<id>)  (when resolvable)
//	Environment:  <selected environment>  (OAuth, when available)
//	Status:  <valid|invalid|not authenticated>
//
// "Source" / "Base URL" / "Environment" / "Status" labels are dim; the
// api-key preview is bold magenta (matching the Retab wordmark elsewhere in
// the CLI).
func writeAuthStatusHuman(w io.Writer, out map[string]any) error {
	s := paletteFor(w)

	preview, _ := out["api_key_preview"].(string)
	source, _ := out["source"].(string)
	authenticated, _ := out["authenticated"].(bool)

	// First line — API keys have a redacted preview; OAuth credentials do
	// not, so fall back to the explicit authenticated/source fields.
	if preview != "" {
		if _, err := fmt.Fprintf(w, "Logged in as %s%s%s\n", s.brand, preview, s.reset); err != nil {
			return err
		}
	} else if authenticated && source != "" {
		if _, err := fmt.Fprintln(w, "Logged in with OAuth"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(w, "Not logged in"); err != nil {
			return err
		}
	}

	// Second line — credential source.
	if source == "" {
		source = "none"
	}
	if _, err := fmt.Fprintf(w, "%sSource:%s  %s\n", s.dim, s.reset, source); err != nil {
		return err
	}

	if baseURL, _ := out["base_url"].(string); baseURL != "" {
		if _, err := fmt.Fprintf(w, "%sBase URL:%s  %s\n", s.dim, s.reset, baseURL); err != nil {
			return err
		}
	}

	if organizationLine := authStatusOrganizationDisplay(out); organizationLine != "" {
		if _, err := fmt.Fprintf(w, "%sOrganization:%s  %s\n", s.dim, s.reset, organizationLine); err != nil {
			return err
		}
	}

	if environmentLine := authStatusEnvironmentDisplay(out); environmentLine != "" {
		if _, err := fmt.Fprintf(w, "%sEnvironment:%s  %s\n", s.dim, s.reset, environmentLine); err != nil {
			return err
		}
	}

	// Final line — verification result. The `valid` key is absent when
	// we never got far enough to probe (no creds, or hint path); treat
	// that as "not authenticated" so the user gets a definite answer.
	var status string
	if v, ok := out["valid"]; ok {
		if b, _ := v.(bool); b {
			status = "valid"
		} else {
			// Just "invalid" — full error detail (often multi-line, with
			// URL/code/body) stays in the JSON's `error` field. Splatting
			// it onto the Status: line would shred the calm 3-line block
			// the human view exists to provide.
			status = "invalid"
		}
	} else {
		status = "not authenticated"
	}
	_, err := fmt.Fprintf(w, "%sStatus:%s  %s\n", s.dim, s.reset, status)
	return err
}

func authStatusOrganizationDisplay(out map[string]any) string {
	organizationAny, ok := out["organization"]
	if !ok || organizationAny == nil {
		return ""
	}
	organization, ok := organizationAny.(map[string]any)
	if !ok {
		return ""
	}
	id, _ := organization["id"].(string)
	name, _ := organization["name"].(string)
	if name != "" && id != "" {
		return fmt.Sprintf("%s (%s)", name, id)
	}
	return id
}

func authStatusEnvironmentDisplay(out map[string]any) string {
	environmentAny, ok := out["environment"]
	if !ok {
		return ""
	}
	if environmentAny == nil {
		return "none selected"
	}
	environment, ok := environmentAny.(map[string]any)
	if !ok {
		return ""
	}
	id, _ := environment["id"].(string)
	name, _ := environment["name"].(string)
	if name != "" && id != "" {
		return fmt.Sprintf("%s (%s)", name, id)
	}
	if id != "" {
		if _, hasError := environment["error"].(string); hasError {
			return fmt.Sprintf("%s (unverified)", id)
		}
		return id
	}
	return "none selected"
}

// writeAuthStatusTable renders the auth payload as a KEY  VALUE
// two-column table. The shape is small and flat (6 well-known rows at
// most), so the generic list-table renderer in output.go is the wrong
// fit — it expects a `data: [...]` list of records. Instead we emit
// each row directly through text/tabwriter, matching the alignment and
// padding rules the generic renderer uses so the look is consistent
// with the rest of the CLI's table output.
//
// Row order is fixed (not map iteration order) so the output is
// deterministic across runs and easy to scan visually. Optional rows
// (BASE_URL, OAUTH_DOMAIN, EXPIRES_AT, HAS_REFRESH, API_KEY_PREVIEW,
// ORGANIZATION, ENVIRONMENT, ENVIRONMENT_SOURCE, ENVIRONMENT_ERROR, VALID,
// ERROR, HINT) are only emitted when present in the payload — rendering an
// empty value would just be noise.
func writeAuthStatusTable(w io.Writer, out map[string]any) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	var writeErr error
	row := func(key string, value any) {
		if writeErr != nil {
			return
		}
		_, writeErr = fmt.Fprintf(tw, "%s\t%v\n", key, value)
	}

	// AUTHENTICATED is always present — the rest are optional and
	// emitted only when set.
	if v, ok := out["authenticated"]; ok {
		row("AUTHENTICATED", v)
	}
	if v, ok := out["source"]; ok {
		if s, _ := v.(string); s != "" {
			row("SOURCE", s)
		}
	}
	if v, ok := out["base_url"]; ok {
		if s, _ := v.(string); s != "" {
			row("BASE_URL", s)
		}
	}
	if v, ok := out["api_key_preview"]; ok {
		if s, _ := v.(string); s != "" {
			row("API_KEY_PREVIEW", s)
		}
	}
	if oauthAny, ok := out["oauth"]; ok {
		if oauth, ok := oauthAny.(map[string]any); ok {
			if d, _ := oauth["authkit_domain"].(string); d != "" {
				row("OAUTH_DOMAIN", d)
			}
			// ExpiresAt arrives here as a time.Time (the in-memory
			// config shape) rather than a JSON string. Stringify both
			// representations so the row prints either way.
			if e := stringifyExpiresAt(oauth["expires_at"]); e != "" {
				row("EXPIRES_AT", e)
			}
			if hr, ok := oauth["has_refresh"]; ok {
				row("HAS_REFRESH", hr)
			}
		}
	}
	if organizationLine := authStatusOrganizationDisplay(out); organizationLine != "" {
		row("ORGANIZATION", organizationLine)
	}
	if environmentLine := authStatusEnvironmentDisplay(out); environmentLine != "" {
		row("ENVIRONMENT", environmentLine)
	}
	if environmentAny, ok := out["environment"]; ok && environmentAny != nil {
		if environment, ok := environmentAny.(map[string]any); ok {
			if s, _ := environment["source"].(string); s != "" {
				row("ENVIRONMENT_SOURCE", s)
			}
			if s, _ := environment["error"].(string); s != "" {
				row("ENVIRONMENT_ERROR", s)
			}
		}
	}
	if v, ok := out["valid"]; ok {
		row("VALID", v)
	}
	if v, ok := out["error"]; ok {
		if s, _ := v.(string); s != "" {
			row("ERROR", s)
		}
	}
	if v, ok := out["hint"]; ok {
		if s, _ := v.(string); s != "" {
			row("HINT", s)
		}
	}

	if writeErr != nil {
		return writeErr
	}
	return tw.Flush()
}

// stringifyExpiresAt renders an OAuth expiry to RFC3339. The in-memory
// AuthConfig.OAuth.ExpiresAt is a time.Time, so the auth status payload
// carries the native value rather than the JSON-marshaled string — a
// plain `value.(string)` assertion would silently miss it. Accept both
// shapes so the table row renders identically whether the payload
// originated from an in-memory config or a JSON round-trip.
func stringifyExpiresAt(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case time.Time:
		if t.IsZero() {
			return ""
		}
		return t.UTC().Format(time.RFC3339)
	case *time.Time:
		if t == nil || t.IsZero() {
			return ""
		}
		return t.UTC().Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
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

	// `--json` forces JSON output even on a TTY. Without it, status auto-
	// detects: TTY → human block, pipe/redirect → JSON (same shape as today).
	authStatusCmd.Flags().Bool("json", false, "emit JSON even when stdout is a TTY (default: human-readable on TTY, JSON when piped)")

	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
