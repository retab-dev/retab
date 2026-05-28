package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// `auth status` is the single most common script-and-eyeball straddle in
// the CLI: humans run it at the prompt to verify they're logged in, and
// scripts pipe it into `jq` to gate other commands. The tests below pin
// the two contracts that, if broken, would silently degrade either side:
//
//  1. Non-TTY writer → JSON. Anything that looks like a pipe or redirect
//     must keep emitting the legacy JSON shape so `auth status | jq` keeps
//     working.
//  2. --json flag → JSON. Even on a TTY, scripts that want JSON should
//     get JSON by setting the flag (and never have to spoof a non-TTY).
//  3. Human output → compact block with the expected labels.
//  4. ANSI escapes NEVER leak into a non-TTY destination. `bytes.Buffer`
//     is the prototypical non-TTY case — a stray `\x1b[` there means
//     `auth status > status.txt` would corrupt the file with control
//     codes.

// Sample payload mirrors the shape the real handler builds when the user
// is authenticated via the env var and the workflows probe succeeded.
func sampleAuthStatus() map[string]any {
	return map[string]any{
		"authenticated":   true,
		"source":          "RETAB_API_KEY env",
		"valid":           true,
		"api_key_preview": "sk_r********************************************xUxQ",
	}
}

// Non-TTY destination (bytes.Buffer) must produce JSON, not the human
// block — so `auth status > out.txt` and `auth status | jq` keep working.
func TestWriteAuthStatus_NonTTYWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatus(&buf, sampleAuthStatus(), false); err != nil {
		t.Fatalf("writeAuthStatus: %v", err)
	}

	// Must parse as JSON.
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("non-TTY output should be JSON; unmarshal failed: %v\n%s", err, buf.String())
	}
	// And carry every key we put in.
	for _, k := range []string{"authenticated", "source", "valid", "api_key_preview"} {
		if _, ok := decoded[k]; !ok {
			t.Errorf("non-TTY JSON output missing key %q:\n%s", k, buf.String())
		}
	}
	// "Logged in as" is the human-mode header — must NOT appear when we
	// emit JSON, or scripts would choke parsing it.
	if strings.Contains(buf.String(), "Logged in as") {
		t.Errorf("non-TTY output should not contain human-mode prose:\n%s", buf.String())
	}
}

// --json forces JSON even when the destination would otherwise render
// human (e.g. a TTY). bytes.Buffer is already non-TTY so we get JSON
// regardless — what we're really pinning here is that --json doesn't
// flip us into human mode by accident.
func TestWriteAuthStatus_JSONFlagForcesJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatus(&buf, sampleAuthStatus(), true); err != nil {
		t.Fatalf("writeAuthStatus: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("--json forced JSON should parse; unmarshal failed: %v\n%s", err, buf.String())
	}
	if strings.Contains(buf.String(), "Logged in as") {
		t.Errorf("--json forced output should not contain human-mode prose:\n%s", buf.String())
	}
}

// JSON output must be byte-equivalent to the legacy `printJSON(out)` form
// — same field set, same alphabetical key order, same 2-space indent,
// same trailing newline. Anything else breaks existing `jq` callsites.
func TestWriteAuthStatusJSON_ShapeIsByteEquivalent(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatusJSON(&buf, sampleAuthStatus()); err != nil {
		t.Fatalf("writeAuthStatusJSON: %v", err)
	}
	// Reproduce the exact legacy form: encoding/json.Encoder with 2-space
	// indent, no HTML escape, trailing newline.
	var want bytes.Buffer
	enc := json.NewEncoder(&want)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(sampleAuthStatus()); err != nil {
		t.Fatalf("encode legacy form: %v", err)
	}
	if buf.String() != want.String() {
		t.Errorf("JSON byte-equivalence regressed.\n got: %q\nwant: %q", buf.String(), want.String())
	}
}

// Human mode (forced by calling writeAuthStatusHuman directly) must
// render the expected lines. We don't pin exact wording beyond these
// labels so the prose can be tweaked without breaking the test.
func TestWriteAuthStatusHuman_Has3LineBlock(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatusHuman(&buf, sampleAuthStatus()); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Logged in as", "Source:", "Status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("human output missing %q:\n%s", want, out)
		}
	}
	// Status line should report "valid" when the probe succeeded.
	if !strings.Contains(out, "valid") {
		t.Errorf("human output should report 'valid' when out[\"valid\"]==true:\n%s", out)
	}
}

// Even though writeAuthStatusHuman renders the human block, the writer
// here is a bytes.Buffer (non-TTY) — paletteFor returns the zero palette
// for non-files, so NO ANSI escapes should appear. Catches the bug where
// someone hardcodes "\x1b[…" instead of going through the palette.
func TestWriteAuthStatusHuman_NoANSILeakForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatusHuman(&buf, sampleAuthStatus()); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("ANSI escape leaked into non-TTY human output:\n%s", buf.String())
	}
}

// writeAuthStatus (the dispatch entry point) routes a non-TTY writer to
// JSON regardless of the --json flag. Pair-check with the human test:
// the same payload that produces a JSON shape here produces a human block
// only when forced through writeAuthStatusHuman.
func TestWriteAuthStatus_NoANSILeakForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatus(&buf, sampleAuthStatus(), false); err != nil {
		t.Fatalf("writeAuthStatus: %v", err)
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("ANSI escape leaked into non-TTY dispatched output:\n%s", buf.String())
	}
}

// "Not logged in" path — when no credential resolved, api_key_preview is
// absent. Human output should fall back gracefully (no blank "Logged in
// as " line, no panic on a missing key).
func TestWriteAuthStatusHuman_NotLoggedIn(t *testing.T) {
	var buf bytes.Buffer
	payload := map[string]any{
		"authenticated": false,
		"source":        "",
		"hint":          "run `retab auth login` to authenticate",
	}
	if err := writeAuthStatusHuman(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Not logged in") {
		t.Errorf("expected 'Not logged in' on empty-source path:\n%s", out)
	}
	// No half-rendered "Logged in as " line with a missing preview.
	if strings.Contains(out, "Logged in as \n") || strings.Contains(out, "Logged in as <") {
		t.Errorf("rendered a malformed 'Logged in as' line for empty preview:\n%s", out)
	}
}

// OAuth credentials do not have an api_key_preview. The human renderer
// must use the explicit authenticated/source fields rather than treating
// the missing preview as "not logged in".
func TestWriteAuthStatusHuman_AuthenticatedOAuth(t *testing.T) {
	var buf bytes.Buffer
	payload := map[string]any{
		"authenticated": true,
		"source":        "~/.retab/config.json (oauth)",
		"valid":         true,
		"oauth": map[string]any{
			"authkit_domain": "meaningful-awakening-88-staging.authkit.app",
			"expires_at":     "2026-05-21T23:45:46Z",
			"has_refresh":    true,
		},
		"environment": map[string]any{
			"id":     "env_prod",
			"name":   "Production",
			"type":   "production",
			"source": "~/.retab/config.json",
		},
	}
	if err := writeAuthStatusHuman(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "Not logged in") {
		t.Fatalf("OAuth status should not render as not logged in:\n%s", out)
	}
	if !strings.Contains(out, "Logged in with OAuth") {
		t.Fatalf("OAuth status should render an OAuth login header:\n%s", out)
	}
	if !strings.Contains(out, "Environment:") || !strings.Contains(out, "Production (env_prod)") {
		t.Fatalf("OAuth status should render the selected environment:\n%s", out)
	}
	if !strings.Contains(out, "valid") {
		t.Fatalf("OAuth status should report valid probe result:\n%s", out)
	}
}

func TestWriteAuthStatusHuman_AuthenticatedOAuthNoEnvironment(t *testing.T) {
	var buf bytes.Buffer
	payload := map[string]any{
		"authenticated": true,
		"source":        "~/.retab/config.json (oauth)",
		"valid":         true,
		"environment":   nil,
	}
	if err := writeAuthStatusHuman(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Environment:") || !strings.Contains(out, "none selected") {
		t.Fatalf("OAuth status should render missing environment selection:\n%s", out)
	}
}

func TestProbeAuthStatus_UsesAuthStatusEndpointForOAuth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	var seenPath string
	var seenAuth string
	var seenAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenAuth = r.Header.Get("Authorization")
		seenAPIKey = r.Header.Get("Api-Key")
		if r.URL.Path != "/v1/auth/status" {
			t.Errorf("probe path = %q, want /v1/auth/status", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"authenticated": true,
			"auth_method": "bearer_token",
			"organization_id": "org_123",
			"environment": null,
			"key": null
		}`))
	}))
	defer server.Close()

	if err := saveConfig(retabConfig{
		BaseURL: server.URL,
		OAuth: &oauthTokens{
			AccessToken:   "at_probe",
			RefreshToken:  "rt_probe",
			ExpiresAt:     time.Now().Add(time.Hour),
			AuthKitDomain: "auth.example.com",
			ClientID:      "client_123",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("api-key", "", "")
	cmd.PersistentFlags().String("base-url", "", "")
	cmd.PersistentFlags().Bool("debug", false, "")

	if err := probeAuthStatus(cmd); err != nil {
		t.Fatalf("probeAuthStatus: %v", err)
	}
	if seenPath != "/v1/auth/status" {
		t.Fatalf("probe path = %q, want /v1/auth/status", seenPath)
	}
	if seenAuth != "Bearer at_probe" {
		t.Fatalf("Authorization header = %q, want Bearer at_probe", seenAuth)
	}
	if seenAPIKey != "" {
		t.Fatalf("Api-Key header should be empty for OAuth probe, got %q", seenAPIKey)
	}
}

func TestAddSelectedEnvironmentStatusIncludesSelectedEnvironment(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	isDefault := true
	var seenAuth string
	var seenAPIKey string
	var seenEnvironmentHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		seenAPIKey = r.Header.Get("Api-Key")
		seenEnvironmentHeader = r.Header.Get(legacyEnvironmentHeaderNameForTest())
		if r.URL.Path != "/v1/environments/env_prod" {
			t.Fatalf("path = %q, want /v1/environments/env_prod", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cliEnvironment{
			ID:        "env_prod",
			Name:      "Production",
			Type:      cliEnvironmentTypeProduction,
			IsDefault: &isDefault,
		})
	}))
	defer server.Close()

	out := map[string]any{}
	addSelectedEnvironmentStatus(
		rootCmd,
		retabConfig{
			EnvironmentID: "env_prod",
			OAuth: &oauthTokens{
				AccessToken:   "at_status",
				RefreshToken:  "rt_status",
				ExpiresAt:     time.Now().Add(time.Hour),
				AuthKitDomain: "auth.example.com",
				ClientID:      "client_123",
			},
		},
		server.URL,
		out,
	)
	if seenAuth != "Bearer at_status" {
		t.Fatalf("Authorization = %q, want raw OAuth bearer", seenAuth)
	}
	if seenAPIKey != "" {
		t.Fatalf("Api-Key should be empty for OAuth environment lookup, got %q", seenAPIKey)
	}
	if seenEnvironmentHeader != "" {
		t.Fatalf("environment lookup sent forbidden environment header %q", seenEnvironmentHeader)
	}
	environment, ok := out["environment"].(map[string]any)
	if !ok {
		t.Fatalf("environment = %#v, want map", out["environment"])
	}
	if environment["id"] != "env_prod" || environment["name"] != "Production" || environment["type"] != "production" {
		t.Fatalf("environment = %#v", environment)
	}
	if environment["source"] != "~/.retab/config.json" {
		t.Fatalf("source = %q, want config", environment["source"])
	}
	if environment["is_default"] != true {
		t.Fatalf("is_default = %#v, want true", environment["is_default"])
	}
}

func TestSelectOAuthLoginEnvironmentUsesRawOAuthAndPicksDefault(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")

	isDefault := true
	var seenAuth string
	var seenEnvironmentHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		seenEnvironmentHeader = r.Header.Get(legacyEnvironmentHeaderNameForTest())
		if r.URL.Path != "/v1/environments" {
			t.Fatalf("path = %q, want /v1/environments", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cliPaginatedList[cliEnvironment]{
			Data: []cliEnvironment{
				{ID: "env_staging", Name: "Staging", Type: cliEnvironmentTypeNonProduction},
				{ID: "env_prod", Name: "Production", Type: cliEnvironmentTypeProduction, IsDefault: &isDefault},
			},
		})
	}))
	defer server.Close()

	environment, err := selectOAuthLoginEnvironment(
		context.Background(),
		server.URL,
		&oauthTokens{AccessToken: "at_login"},
		"",
	)
	if err != nil {
		t.Fatalf("selectOAuthLoginEnvironment: %v", err)
	}
	if seenAuth != "Bearer at_login" {
		t.Fatalf("Authorization = %q, want raw OAuth bearer", seenAuth)
	}
	if seenEnvironmentHeader != "" {
		t.Fatalf("login env resolution sent forbidden environment header %q", seenEnvironmentHeader)
	}
	if environment.ID != "env_prod" {
		t.Fatalf("selected environment = %s, want default production", environment.ID)
	}
}

func TestChooseLoginEnvironmentPreservesExistingSelection(t *testing.T) {
	isDefault := true
	list := &cliPaginatedList[cliEnvironment]{
		Data: []cliEnvironment{
			{ID: "env_staging", Name: "Staging", Type: cliEnvironmentTypeNonProduction},
			{ID: "env_prod", Name: "Production", Type: cliEnvironmentTypeProduction, IsDefault: &isDefault},
		},
	}

	environment := chooseLoginEnvironment("env_staging", list)
	if environment == nil {
		t.Fatal("expected selected environment")
	}
	if environment.ID != "env_staging" {
		t.Fatalf("selected environment = %s, want existing selection", environment.ID)
	}
}

func TestConfiguredLoginBaseURLDefaultsToProductionInsteadOfStoredLocalhost(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	if err := saveConfig(retabConfig{BaseURL: "http://localhost:4000"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	got := configuredLoginBaseURL("")
	if got != defaultAPIBaseURL {
		t.Fatalf("login base URL = %q, want %q", got, defaultAPIBaseURL)
	}
}

func TestAPIKeyLoginDefaultsToProductionInsteadOfStoredLocalhost(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")

	if err := saveConfig(retabConfig{BaseURL: "http://localhost:4000"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	if err := runAPIKeyLogin("sk_live_test", ""); err != nil {
		t.Fatalf("runAPIKeyLogin: %v", err)
	}
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.BaseURL != defaultAPIBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, defaultAPIBaseURL)
	}
	if cfg.APIKey != "sk_live_test" {
		t.Fatalf("APIKey = %q, want sk_live_test", cfg.APIKey)
	}
}

// `auth status --output table` (the global formatting flag every other
// command honours) must render a key/value table, NOT silently fall
// through to JSON. The bug was that writeAuthStatus only knew about the
// command-local --json flag and TTY auto-detect; the global --output
// table flag was advertised but ignored, so scripts setting it on the
// root saw raw JSON instead of structured key/value rows.
func TestWriteAuthStatusTable_AuthenticatedOAuth(t *testing.T) {
	var buf bytes.Buffer
	payload := map[string]any{
		"authenticated": true,
		"source":        "~/.retab/config.json (oauth)",
		"base_url":      "http://localhost:4000/v1",
		"valid":         true,
		"oauth": map[string]any{
			"authkit_domain": "meaningful-awakening-88-staging.authkit.app",
			"expires_at":     "2026-05-21T23:45:46Z",
			"has_refresh":    true,
		},
		"environment": map[string]any{
			"id":     "env_prod",
			"name":   "Production",
			"type":   "production",
			"source": "~/.retab/config.json",
		},
	}
	if err := writeAuthStatusTable(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusTable: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"AUTHENTICATED",
		"true",
		"BASE_URL",
		"http://localhost:4000/v1",
		"SOURCE",
		"~/.retab/config.json (oauth)",
		"OAUTH_DOMAIN",
		"meaningful-awakening-88-staging.authkit.app",
		"EXPIRES_AT",
		"2026-05-21T23:45:46Z",
		"HAS_REFRESH",
		"ENVIRONMENT",
		"Production (env_prod)",
		"ENVIRONMENT_SOURCE",
		"~/.retab/config.json",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
	// No JSON braces — we're not falling back to JSON.
	if strings.Contains(out, "{") || strings.Contains(out, "\"authenticated\"") {
		t.Errorf("table output should not be JSON:\n%s", out)
	}
	// No ANSI escapes leaking through.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("ANSI escape leaked into table output:\n%s", out)
	}
}

// The in-memory AuthConfig.OAuth.ExpiresAt is a time.Time, so the
// payload built by authStatusCmd.RunE carries the native value (not a
// JSON string). The table renderer must handle both shapes — the
// initial implementation silently dropped EXPIRES_AT for the time.Time
// case, leaving real users without the expiry row even though it's the
// single most actionable field for "did my refresh fail?" diagnostics.
func TestWriteAuthStatusTable_ExpiresAtAsTimeValue(t *testing.T) {
	expires := time.Date(2026, 5, 21, 23, 45, 46, 0, time.UTC)
	var buf bytes.Buffer
	payload := map[string]any{
		"authenticated": true,
		"source":        "~/.retab/config.json (oauth)",
		"oauth": map[string]any{
			"authkit_domain": "x.authkit.app",
			"expires_at":     expires,
			"has_refresh":    true,
		},
	}
	if err := writeAuthStatusTable(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusTable: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "EXPIRES_AT") {
		t.Errorf("table output missing EXPIRES_AT row (time.Time case):\n%s", out)
	}
	if !strings.Contains(out, "2026-05-21T23:45:46Z") {
		t.Errorf("table output missing RFC3339-formatted expiry:\n%s", out)
	}
}

func TestWriteAuthStatusTable_NotLoggedIn(t *testing.T) {
	var buf bytes.Buffer
	payload := map[string]any{
		"authenticated": false,
		"source":        "",
		"hint":          "run `retab auth login` to authenticate",
	}
	if err := writeAuthStatusTable(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusTable: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"AUTHENTICATED", "false", "HINT", "run `retab auth login`"} {
		if !strings.Contains(out, want) {
			t.Errorf("not-logged-in table output missing %q:\n%s", want, out)
		}
	}
}

// writeAuthStatus must dispatch to the table renderer when outputFormat
// is OutputTable, regardless of --json or TTY state. This is the
// integration glue that was missing — the renderers existed, the auto-
// detect existed, but nothing routed --output table to a table view.
func TestWriteAuthStatus_OutputTableRoutesToTable(t *testing.T) {
	var buf bytes.Buffer
	payload := sampleAuthStatus()
	if err := writeAuthStatusWithFormat(&buf, payload, false, OutputTable); err != nil {
		t.Fatalf("writeAuthStatusWithFormat: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "AUTHENTICATED") {
		t.Errorf("--output table should emit AUTHENTICATED row, got:\n%s", out)
	}
	if strings.Contains(out, "{") {
		t.Errorf("--output table should not emit JSON braces, got:\n%s", out)
	}
}

// --output json overrides TTY auto-detect — even on a TTY writer, json
// wins. Non-TTY here (bytes.Buffer) plus OutputJSON is the trivial case
// but confirms the routing.
func TestWriteAuthStatus_OutputJSONRoutesToJSON(t *testing.T) {
	var buf bytes.Buffer
	payload := sampleAuthStatus()
	if err := writeAuthStatusWithFormat(&buf, payload, false, OutputJSON); err != nil {
		t.Fatalf("writeAuthStatusWithFormat: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("--output json should produce JSON, got %v:\n%s", err, buf.String())
	}
	if _, ok := decoded["authenticated"]; !ok {
		t.Errorf("--output json missing 'authenticated' key:\n%s", buf.String())
	}
}
