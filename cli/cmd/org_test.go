package cmd

import (
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

func TestOrgSwitchEndToEnd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	var switchBody map[string]string
	var switchAuth string
	isDefault := true

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
		case r.Method == http.MethodPost && r.URL.Path == "/v1/auth/cli/switch-organization":
			switchAuth = r.Header.Get("Authorization")
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &switchBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliSwitchOrganizationResponse{
				AccessToken:    "at_new",
				RefreshToken:   "rt_new",
				TokenType:      "Bearer",
				ExpiresIn:      600,
				OrganizationID: "org_beta",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/organization":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cliAuthOrganization{ID: "org_acme", Name: "Acme Inc"})
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
			AccessToken:   "at_old",
			RefreshToken:  "rt_old",
			ExpiresAt:     time.Now().Add(time.Hour),
			AuthKitDomain: "auth.example.com",
			ClientID:      "client_cli",
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

	// The switch request carried the stored refresh token + resolved org id and
	// no bearer (the refresh token is the credential).
	if switchBody["refresh_token"] != "rt_old" {
		t.Fatalf("switch refresh_token = %q, want rt_old", switchBody["refresh_token"])
	}
	if switchBody["organization_id"] != "org_beta" {
		t.Fatalf("switch organization_id = %q, want org_beta", switchBody["organization_id"])
	}
	if switchAuth != "" {
		t.Fatalf("switch request carried Authorization %q, want none", switchAuth)
	}

	// The new tokens and the new org's default environment were persisted.
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.OAuth == nil {
		t.Fatal("OAuth config was cleared")
	}
	if cfg.OAuth.AccessToken != "at_new" {
		t.Fatalf("AccessToken = %q, want at_new", cfg.OAuth.AccessToken)
	}
	if cfg.OAuth.RefreshToken != "rt_new" {
		t.Fatalf("RefreshToken = %q, want rt_new", cfg.OAuth.RefreshToken)
	}
	// Discovery fields survive so transparent refresh keeps working.
	if cfg.OAuth.AuthKitDomain != "auth.example.com" || cfg.OAuth.ClientID != "client_cli" {
		t.Fatalf("discovery fields lost: domain=%q client=%q", cfg.OAuth.AuthKitDomain, cfg.OAuth.ClientID)
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

	var switchCalled bool
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
		case r.Method == http.MethodPost && r.URL.Path == "/v1/auth/cli/switch-organization":
			switchCalled = true
			w.WriteHeader(http.StatusForbidden)
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
			AccessToken:   "at_current",
			RefreshToken:  "rt_current",
			ExpiresAt:     time.Now().Add(time.Hour),
			AuthKitDomain: "auth.example.com",
			ClientID:      "client_cli",
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
	if switchCalled {
		t.Fatal("switch endpoint was called for the already-current organization")
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
			AccessToken:   "at_live",
			RefreshToken:  "rt_live",
			ExpiresAt:     time.Now().Add(time.Hour),
			AuthKitDomain: "auth.example.com",
			ClientID:      "client_cli",
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
