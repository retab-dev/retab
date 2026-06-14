package cmd

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestResolveTargetOrganization(t *testing.T) {
	orgs := []cliOrganization{
		{ID: "org_acme", Name: "Acme Inc"},
		{ID: "org_beta", Name: "Beta"},
		{ID: "org_dup1", Name: "Same Name"},
		{ID: "org_dup2", Name: "Same Name"},
	}

	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "exact id", input: "org_beta", want: "org_beta"},
		{name: "name case-insensitive", input: "acme inc", want: "org_acme"},
		{name: "name exact", input: "Beta", want: "org_beta"},
		{name: "unknown org id passthrough", input: "org_unlisted", want: "org_unlisted"},
		{name: "ambiguous name", input: "Same Name", wantErr: true},
		{name: "unknown name", input: "Nope", wantErr: true},
		{name: "empty", input: "   ", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveTargetOrganization(tc.input, orgs)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveTargetOrganization(%q) = %q, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveTargetOrganization(%q): %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("resolveTargetOrganization(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// orgScopedAccessToken builds an unsigned JWT-shaped access token carrying an
// `org_id` claim, so accessTokenOrgID can confirm which org the switch landed
// in. Only the payload segment is read; header/sig are placeholders.
func orgScopedAccessToken(t *testing.T, orgID string) string {
	t.Helper()
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"org_id":"` + orgID + `"}`))
	return "h." + payload + ".s"
}

func TestOrgSwitchEndToEnd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	isDefault := true
	betaToken := orgScopedAccessToken(t, "org_beta")

	// The backend's confidential switch endpoint exchanges the stored refresh
	// token (+ target org) for a fresh pair already scoped to org_beta. Capture
	// the request body to assert the CLI sent the right refresh_token + org.
	var switchReq cliSwitchOrganizationRequest
	expiresIn := 3600
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/cli/organizations":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliOrganizationsResponse{
				Data: []cliOrganization{
					{ID: "org_acme", Name: "Acme Inc"},
					{ID: "org_beta", Name: "Beta"},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/organization":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliAuthOrganization{ID: "org_acme", Name: "Acme Inc"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/auth/cli/switch-organization":
			_ = json.NewDecoder(r.Body).Decode(&switchReq)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliSwitchOrganizationResponse{
				AccessToken:    betaToken,
				RefreshToken:   "rt_new",
				TokenType:      "Bearer",
				ExpiresIn:      &expiresIn,
				OrganizationID: "org_beta",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/environments":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliPaginatedList[cliEnvironment]{
				Data: []cliEnvironment{
					{ID: "env_beta_prod", Name: "Production", Type: cliEnvironmentTypeProduction, IsDefault: &isDefault},
				},
			})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Start logged into Acme, with a stale environment from that org.
	if err := saveConfig(retabConfig{
		BaseURL:       server.URL,
		EnvironmentID: "env_acme_old",
		OAuth: &oauthTokens{
			AccessToken:      orgScopedAccessToken(t, "org_acme"),
			RefreshToken:     "rt_old",
			ExpiresAt:        time.Now().Add(time.Hour),
			WorkosAPIBaseURL: "https://api.workos.com",
			ClientID:         "client_cli",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("api-key", "", "")
	cmd.PersistentFlags().String("base-url", "", "")
	cmd.PersistentFlags().Bool("debug", false, "")

	if err := orgSwitchCmd.RunE(cmd, []string{"Beta"}); err != nil {
		t.Fatalf("org switch: %v", err)
	}

	// The CLI exchanged the stored refresh token for the resolved target org.
	if switchReq.RefreshToken != "rt_old" {
		t.Fatalf("switch refresh_token = %q, want rt_old", switchReq.RefreshToken)
	}
	if switchReq.OrganizationID != "org_beta" {
		t.Fatalf("switch organization_id = %q, want org_beta", switchReq.OrganizationID)
	}

	// The new tokens and the new org's default environment were persisted.
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.OAuth == nil {
		t.Fatal("OAuth config was cleared")
	}
	if cfg.OAuth.AccessToken != betaToken {
		t.Fatalf("AccessToken = %q, want the org_beta-scoped token", cfg.OAuth.AccessToken)
	}
	if cfg.OAuth.RefreshToken != "rt_new" {
		t.Fatalf("RefreshToken = %q, want rt_new", cfg.OAuth.RefreshToken)
	}
	// Discovery fields survive so transparent refresh keeps working.
	if cfg.OAuth.WorkosAPIBaseURL != "https://api.workos.com" || cfg.OAuth.ClientID != "client_cli" {
		t.Fatalf("discovery fields lost: base=%q client=%q", cfg.OAuth.WorkosAPIBaseURL, cfg.OAuth.ClientID)
	}
	if cfg.OAuth.OrganizationID != "org_beta" {
		t.Fatalf("OrganizationID = %q, want org_beta", cfg.OAuth.OrganizationID)
	}
	if cfg.EnvironmentID != "env_beta_prod" {
		t.Fatalf("EnvironmentID = %q, want env_beta_prod (re-selected for new org)", cfg.EnvironmentID)
	}
	if cfg.EnvironmentType != "production" {
		t.Fatalf("EnvironmentType = %q, want production", cfg.EnvironmentType)
	}
}

func TestOrgSwitchCurrentOrganizationIsNoop(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	// No discovery / authorize endpoints are stubbed: a no-op switch must
	// short-circuit BEFORE any browser re-auth. Hitting /v1/auth/cli/config or
	// the WorkOS flow would fail the unexpected-request guard below.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/cli/organizations":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliOrganizationsResponse{
				Data: []cliOrganization{
					{ID: "org_acme", Name: "Acme Inc"},
					{ID: "org_beta", Name: "Beta"},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/organization":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliAuthOrganization{ID: "org_beta", Name: "Beta"})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	if err := saveConfig(retabConfig{
		BaseURL:       server.URL,
		EnvironmentID: "env_beta_prod",
		OAuth: &oauthTokens{
			AccessToken:      "at_current",
			RefreshToken:     "rt_current",
			ExpiresAt:        time.Now().Add(time.Hour),
			WorkosAPIBaseURL: "https://api.workos.com",
			ClientID:         "client_cli",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("api-key", "", "")
	cmd.PersistentFlags().String("base-url", "", "")
	cmd.PersistentFlags().Bool("debug", false, "")

	if err := orgSwitchCmd.RunE(cmd, []string{"Beta"}); err != nil {
		t.Fatalf("switch current org should be a no-op: %v", err)
	}
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.EnvironmentID != "env_beta_prod" {
		t.Fatalf("EnvironmentID = %q, want unchanged env_beta_prod", cfg.EnvironmentID)
	}
	if cfg.OAuth == nil || cfg.OAuth.AccessToken != "at_current" || cfg.OAuth.RefreshToken != "rt_current" {
		t.Fatalf("OAuth tokens changed on no-op switch: %#v", cfg.OAuth)
	}
}

// TestOrgListHonorsOutputFormat pins the regression that `org list` routes
// through --output like every other list command: JSON carries the
// current-org marker as a `current` flag, and the table/csv renderers expose
// the same data. Before the fix, org list always hand-rolled a tabwriter and
// silently ignored --output (so `--output json` still printed a table).
func TestOrgListHonorsOutputFormat(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/cli/organizations":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliOrganizationsResponse{
				Data: []cliOrganization{
					{ID: "org_acme", Name: "Acme Inc"},
					{ID: "org_beta", Name: "Beta"},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/organization":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliAuthOrganization{ID: "org_beta", Name: "Beta"})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// OAuth login with NO selected environment, so cliJSONRequestInto uses the
	// raw token directly (no dashboard-context exchange hop to mock).
	if err := saveConfig(retabConfig{
		BaseURL: server.URL,
		OAuth: &oauthTokens{
			AccessToken:      "at_live",
			RefreshToken:     "rt_live",
			ExpiresAt:        time.Now().Add(time.Hour),
			WorkosAPIBaseURL: "https://api.workos.com",
			ClientID:         "client_cli",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	run := func(t *testing.T, output string) string {
		t.Helper()
		cmd := &cobra.Command{}
		cmd.PersistentFlags().String("api-key", "", "")
		cmd.PersistentFlags().String("base-url", "", "")
		cmd.PersistentFlags().Bool("debug", false, "")
		cmd.PersistentFlags().String("output", output, "")

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe: %v", err)
		}
		orig := os.Stdout
		os.Stdout = w
		t.Cleanup(func() { os.Stdout = orig })

		out := make(chan string, 1)
		go func() {
			b, _ := io.ReadAll(r)
			out <- string(b)
		}()

		if err := orgListCmd.RunE(cmd, nil); err != nil {
			_ = w.Close()
			os.Stdout = orig
			t.Fatalf("org list (--output %s): %v", output, err)
		}
		_ = w.Close()
		os.Stdout = orig
		return <-out
	}

	t.Run("json", func(t *testing.T) {
		got := run(t, "json")
		var parsed struct {
			Data []orgListRow `json:"data"`
		}
		if err := json.Unmarshal([]byte(got), &parsed); err != nil {
			t.Fatalf("output is not JSON: %v\n%s", err, got)
		}
		if len(parsed.Data) != 2 {
			t.Fatalf("want 2 orgs, got %d: %s", len(parsed.Data), got)
		}
		var current []string
		for _, o := range parsed.Data {
			if o.Current {
				current = append(current, o.ID)
			}
		}
		if len(current) != 1 || current[0] != "org_beta" {
			t.Fatalf("want only org_beta current, got %v\n%s", current, got)
		}
	})

	t.Run("table", func(t *testing.T) {
		got := run(t, "table")
		if !strings.Contains(got, "ID") || !strings.Contains(got, "CURRENT") {
			t.Fatalf("table missing header: %s", got)
		}
		// The current org gets the "(current)" marker; the other does not.
		for _, line := range strings.Split(got, "\n") {
			if strings.HasPrefix(line, "org_beta") && !strings.Contains(line, "(current)") {
				t.Fatalf("org_beta row missing (current) marker: %q", line)
			}
			if strings.HasPrefix(line, "org_acme") && strings.Contains(line, "(current)") {
				t.Fatalf("org_acme row wrongly marked current: %q", line)
			}
		}
	})

	t.Run("csv", func(t *testing.T) {
		got := run(t, "csv")
		if !strings.HasPrefix(got, "ID,NAME,CURRENT") {
			t.Fatalf("csv header wrong: %s", got)
		}
		if !strings.Contains(got, "org_beta,Beta,(current)") {
			t.Fatalf("csv missing current marker row: %s", got)
		}
	})
}

func TestOrgSwitchRequiresOAuth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")

	// API-key login only — no OAuth refresh token to exchange.
	if err := saveConfig(retabConfig{APIKey: "sk_live_abc"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("api-key", "", "")
	cmd.PersistentFlags().String("base-url", "", "")
	cmd.PersistentFlags().Bool("debug", false, "")

	err := orgSwitchCmd.RunE(cmd, []string{"org_beta"})
	if err == nil {
		t.Fatal("org switch with API-key session should error")
	}
}
