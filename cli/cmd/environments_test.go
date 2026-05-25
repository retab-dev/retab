package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		seenEnvironmentHeader = r.Header.Get("X-Retab-Environment-Id")
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

func TestNewClientSendsSelectedEnvironmentHeader(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	if err := saveConfig(retabConfig{EnvironmentID: "env_staging"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	var seenEnvironmentHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenEnvironmentHeader = r.Header.Get("X-Retab-Environment-Id")
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
	if seenEnvironmentHeader != "env_staging" {
		t.Fatalf("environment header = %q, want env_staging", seenEnvironmentHeader)
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

func resetEnvironmentCommandPersistentFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"api-key", "base-url", "environment-id"} {
		if err := rootCmd.PersistentFlags().Set(name, ""); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
	}
	t.Cleanup(func() {
		for _, name := range []string{"api-key", "base-url", "environment-id"} {
			_ = rootCmd.PersistentFlags().Set(name, "")
		}
	})
}
