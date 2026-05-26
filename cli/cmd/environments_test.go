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

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

func TestResolveEnvironmentSelectionUsesIDBeforeName(t *testing.T) {
	list := &retab.PaginatedList[retab.Environment]{Data: []retab.Environment{
		{ID: "env_prod", Name: "Production", Type: retab.EnvironmentType("production")},
		{ID: "Production", Name: "Shadow", Type: retab.EnvironmentType("non_production")},
	}}

	got, err := resolveEnvironmentSelection("Production", list)
	if err != nil {
		t.Fatalf("resolveEnvironmentSelection: %v", err)
	}
	if got.ID != "Production" {
		t.Fatalf("selected id = %q, want direct id match", got.ID)
	}
}

func TestResolveEnvironmentSelectionRejectsAmbiguousName(t *testing.T) {
	list := &retab.PaginatedList[retab.Environment]{Data: []retab.Environment{
		{ID: "env_1", Name: "Staging", Type: retab.EnvironmentType("non_production")},
		{ID: "env_2", Name: "Staging", Type: retab.EnvironmentType("non_production")},
	}}

	_, err := resolveEnvironmentSelection("Staging", list)
	if err == nil {
		t.Fatal("expected ambiguous name error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("error = %q, want ambiguous", err.Error())
	}
}

func TestEnvSwitchPersistsIDAndDoesNotSendStaleEnvironmentHeader(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	if err := saveConfig(retabConfig{EnvironmentID: "env_stale"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	var seenEnvironmentHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenEnvironmentHeader = r.Header.Get(legacyEnvironmentHeaderNameForTest())
		if r.URL.Path != "/v1/environments" {
			t.Fatalf("path = %q, want /v1/environments", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"env_staging","name":"Staging","type":"non_production"}],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := envSwitchCmd.RunE(envSwitchCmd, []string{"Staging"}); err != nil {
		t.Fatalf("env switch: %v", err)
	}
	if seenEnvironmentHeader != "" {
		t.Fatalf("env switch sent stale environment header %q", seenEnvironmentHeader)
	}
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.EnvironmentID != "env_staging" {
		t.Fatalf("EnvironmentID = %q, want env_staging", cfg.EnvironmentID)
	}
}

func TestNewClientDoesNotUseSelectedEnvironmentForAPIKeyAuth(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	if err := saveConfig(retabConfig{EnvironmentID: "env_staging"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	var seenEnvironmentHeader string
	var seenAPIKey string
	var seenAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/auth/dashboard-context" {
			t.Fatalf("API-key auth must not mint dashboard context tokens")
		}
		seenEnvironmentHeader = r.Header.Get(legacyEnvironmentHeaderNameForTest())
		seenAPIKey = r.Header.Get("Api-Key")
		seenAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/workflows" {
			t.Fatalf("path = %q, want /v1/workflows", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	client, err := newClient(rootCmd)
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	if _, err := client.Workflows.List(context.Background(), &retab.WorkflowsListParams{}); err != nil {
		t.Fatalf("workflows list: %v", err)
	}
	if seenEnvironmentHeader != "" {
		t.Fatalf("environment header = %q, want empty", seenEnvironmentHeader)
	}
	if seenAPIKey != "test-key" {
		t.Fatalf("Api-Key = %q, want test-key", seenAPIKey)
	}
	if seenAuth != "" {
		t.Fatalf("Authorization = %q, want empty", seenAuth)
	}
}

func TestNewClientUsesDashboardContextTokenForSelectedOAuthEnvironment(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	const environmentID = "env_staging"
	var contextCalls int
	var workflowCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(legacyEnvironmentHeaderNameForTest()) != "" {
			t.Fatalf("request to %s sent forbidden environment header %q", r.URL.Path, r.Header.Get(legacyEnvironmentHeaderNameForTest()))
		}
		switch r.URL.Path {
		case "/v1/auth/dashboard-context":
			contextCalls++
			if r.Header.Get("Authorization") != "Bearer at_cli" {
				t.Fatalf("dashboard context auth = %q, want raw OAuth bearer", r.Header.Get("Authorization"))
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode dashboard context body: %v", err)
			}
			if body["environment_id"] != environmentID {
				t.Fatalf("environment_id = %q, want %s", body["environment_id"], environmentID)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"token": "ctx_cli",
				"expires_at": "2035-01-01T00:00:00Z",
				"token_type": "Bearer",
				"environment": {"id": "env_staging", "name": "Staging", "type": "non_production"},
				"region": "eu",
				"ws_path": ""
			}`))
		case "/v1/workflows":
			workflowCalls++
			if r.Header.Get("Authorization") != "Bearer ctx_cli" {
				t.Fatalf("workflow auth = %q, want dashboard context bearer", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Api-Key") != "" {
				t.Fatalf("Api-Key should be empty for dashboard context auth, got %q", r.Header.Get("Api-Key"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[]}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	if err := saveConfig(retabConfig{
		BaseURL:       server.URL,
		EnvironmentID: environmentID,
		OAuth: &oauthTokens{
			AccessToken:   "at_cli",
			RefreshToken:  "rt_cli",
			TokenType:     "Bearer",
			ExpiresAt:     time.Now().Add(time.Hour),
			AuthKitDomain: "auth.example.com",
			ClientID:      "client_123",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	client, err := newClient(rootCmd)
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	if _, err := client.Workflows.List(context.Background(), &retab.WorkflowsListParams{}); err != nil {
		t.Fatalf("workflows list: %v", err)
	}
	if contextCalls != 1 {
		t.Fatalf("dashboard context calls = %d, want 1", contextCalls)
	}
	if workflowCalls != 1 {
		t.Fatalf("workflow calls = %d, want 1", workflowCalls)
	}
}

func TestCLIJSONRequestUsesDashboardContextTokenForSelectedOAuthEnvironment(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	const environmentID = "env_staging"
	var contextCalls int
	var probeCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(legacyEnvironmentHeaderNameForTest()) != "" {
			t.Fatalf("request to %s sent forbidden environment header %q", r.URL.Path, r.Header.Get(legacyEnvironmentHeaderNameForTest()))
		}
		switch r.URL.Path {
		case "/v1/auth/dashboard-context":
			contextCalls++
			if r.Header.Get("Authorization") != "Bearer at_raw_json" {
				t.Fatalf("dashboard context auth = %q, want raw OAuth bearer", r.Header.Get("Authorization"))
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode dashboard context body: %v", err)
			}
			if body["environment_id"] != environmentID {
				t.Fatalf("environment_id = %q, want %s", body["environment_id"], environmentID)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"token": "ctx_raw_json",
				"expires_at": "2035-01-01T00:00:00Z",
				"token_type": "Bearer",
				"environment": {"id": "env_staging", "name": "Staging", "type": "non_production"},
				"region": "eu",
				"ws_path": ""
			}`))
		case "/v1/raw-probe":
			probeCalls++
			if r.Header.Get("Authorization") != "Bearer ctx_raw_json" {
				t.Fatalf("probe auth = %q, want dashboard context bearer", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Api-Key") != "" {
				t.Fatalf("Api-Key should be empty for dashboard context auth, got %q", r.Header.Get("Api-Key"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	if err := saveConfig(retabConfig{
		BaseURL:       server.URL,
		EnvironmentID: environmentID,
		OAuth: &oauthTokens{
			AccessToken:   "at_raw_json",
			RefreshToken:  "rt_raw_json",
			TokenType:     "Bearer",
			ExpiresAt:     time.Now().Add(time.Hour),
			AuthKitDomain: "auth.example.com",
			ClientID:      "client_123",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	result, err := cliJSONRequest(rootCmd, http.MethodGet, "/v1/raw-probe", nil, nil)
	if err != nil {
		t.Fatalf("cliJSONRequest: %v", err)
	}
	if result.(map[string]any)["ok"] != true {
		t.Fatalf("result = %#v, want ok=true", result)
	}
	if contextCalls != 1 {
		t.Fatalf("dashboard context calls = %d, want 1", contextCalls)
	}
	if probeCalls != 1 {
		t.Fatalf("probe calls = %d, want 1", probeCalls)
	}
}

func TestEnvSwitchUsesRawOAuthAndDoesNotMintDashboardContext(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/auth/dashboard-context" {
			t.Fatalf("env management commands must not mint dashboard context tokens")
		}
		if r.Header.Get(legacyEnvironmentHeaderNameForTest()) != "" {
			t.Fatalf("env switch sent forbidden environment header %q", r.Header.Get(legacyEnvironmentHeaderNameForTest()))
		}
		if r.Header.Get("Authorization") != "Bearer at_env" {
			t.Fatalf("env switch auth = %q, want raw OAuth bearer", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/v1/environments" {
			t.Fatalf("path = %q, want /v1/environments", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"env_staging","name":"Staging","type":"non_production"}],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()

	if err := saveConfig(retabConfig{
		BaseURL:       server.URL,
		EnvironmentID: "env_stale",
		OAuth: &oauthTokens{
			AccessToken:   "at_env",
			RefreshToken:  "rt_env",
			TokenType:     "Bearer",
			ExpiresAt:     time.Now().Add(time.Hour),
			AuthKitDomain: "auth.example.com",
			ClientID:      "client_123",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	if err := envSwitchCmd.RunE(envSwitchCmd, []string{"Staging"}); err != nil {
		t.Fatalf("env switch: %v", err)
	}
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.EnvironmentID != "env_staging" {
		t.Fatalf("EnvironmentID = %q, want env_staging", cfg.EnvironmentID)
	}
}

func TestEnvWhichShowsSelectedEnvironment(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("RETAB_BASE_URL", "")

	isDefault := true
	var seenEnvironmentHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenEnvironmentHeader = r.Header.Get(legacyEnvironmentHeaderNameForTest())
		if r.URL.Path != "/v1/environments/env_prod" {
			t.Fatalf("path = %q, want /v1/environments/env_prod", r.URL.Path)
		}
		if r.Header.Get("Api-Key") != "test-key" {
			t.Fatalf("Api-Key = %q, want test-key", r.Header.Get("Api-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(retab.Environment{
			ID:        "env_prod",
			Name:      "Production",
			Type:      retab.AuthStatusEnvironmentTypeProduction,
			IsDefault: &isDefault,
		})
	}))
	defer server.Close()

	if err := saveConfig(retabConfig{BaseURL: server.URL, EnvironmentID: "env_prod"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	var out bytes.Buffer
	envWhichCmd.SetOut(&out)
	t.Cleanup(func() { envWhichCmd.SetOut(nil) })
	if err := envWhichCmd.RunE(envWhichCmd, nil); err != nil {
		t.Fatalf("env which: %v", err)
	}
	if seenEnvironmentHeader != "" {
		t.Fatalf("env which sent forbidden environment header %q", seenEnvironmentHeader)
	}
	var decoded selectedEnvironment
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("env which output should be JSON for non-TTY writer: %v\n%s", err, out.String())
	}
	if decoded.ID != "env_prod" || decoded.Name != "Production" || decoded.Type != "production" || !decoded.IsDefault {
		t.Fatalf("selected environment = %#v", decoded)
	}
	if decoded.Source != "~/.retab/config.json" {
		t.Fatalf("source = %q, want config", decoded.Source)
	}
}

func TestEnvAddValidatesType(t *testing.T) {
	cmd := &cobra.Command{Use: "test-env-add", RunE: envAddCmd.RunE}
	cmd.Flags().String("name", "Preview", "")
	cmd.Flags().String("type", "local", "")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected invalid type error")
	}
	if !strings.Contains(err.Error(), "invalid --type") {
		t.Fatalf("error = %q, want invalid --type", err.Error())
	}
}

func TestEnvironmentListJSONShape(t *testing.T) {
	result := &retab.PaginatedList[retab.Environment]{Data: []retab.Environment{
		{ID: "env_prod", Name: "Production", Type: retab.EnvironmentType("production")},
	}}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"data"`) {
		t.Fatalf("environment list should use pagination envelope, got %s", raw)
	}
}

func TestEnvironmentTableDefaultCellHidesFalseForTypedRows(t *testing.T) {
	isDefault := true
	notDefault := false

	if got := environmentCell(retab.Environment{IsDefault: &isDefault}, "is_default"); got != "true" {
		t.Fatalf("default cell = %q, want true", got)
	}
	if got := environmentCell(retab.Environment{IsDefault: &notDefault}, "is_default"); got != "" {
		t.Fatalf("non-default cell = %q, want blank", got)
	}
	if got := environmentCell(map[string]any{"is_default": false}, "is_default"); got != "" {
		t.Fatalf("map non-default cell = %q, want blank", got)
	}
}

func resetEnvironmentCommandPersistentFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"api-key", "base-url", "environment-id", "output"} {
		if err := rootCmd.PersistentFlags().Set(name, ""); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
	}
	t.Cleanup(func() {
		for _, name := range []string{"api-key", "base-url", "environment-id", "output"} {
			_ = rootCmd.PersistentFlags().Set(name, "")
		}
	})
}

func legacyEnvironmentHeaderNameForTest() string {
	return strings.Join([]string{"X", "Retab", "Environment", "Id"}, "-")
}
