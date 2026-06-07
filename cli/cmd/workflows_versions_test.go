//go:build !retab_oagen_cli_workflows

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkflowsVersionCommandsUseGeneratedRoutes(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/versions":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "list_metadata": map[string]any{}})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/versions/diff":
			_ = json.NewEncoder(w).Encode(map[string]any{"workflow_id": "wf_123", "changes": []any{}})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/versions/wfv_1":
			_ = json.NewEncoder(w).Encode(map[string]any{"workflow_id": "wf_123", "workflow_version_id": "wfv_1"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/versions/wfv_1/restore":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "wf_123", "name": "Restored"})
		default:
			t.Fatalf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsVersionsCmd.RunE(workflowsVersionsCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("versions: %v", err)
	}
	if err := workflowsDiffCmd.Flags().Set("from-version-id", "wfv_0"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsDiffCmd.Flags().Set("to-version-id", "wfv_1"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsDiffCmd.Flags().Set("from-version-id", "")
		_ = workflowsDiffCmd.Flags().Set("to-version-id", "")
	})
	if err := workflowsDiffCmd.RunE(workflowsDiffCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("diff: %v", err)
	}
	if err := workflowsVersionCmd.RunE(workflowsVersionCmd, []string{"wf_123", "wfv_1"}); err != nil {
		t.Fatalf("version: %v", err)
	}
	if err := workflowsVersionRestoreCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsVersionRestoreCmd.Flags().Set("yes", "false") })
	if err := workflowsVersionRestoreCmd.RunE(workflowsVersionRestoreCmd, []string{"wf_123", "wfv_1"}); err != nil {
		t.Fatalf("version-restore: %v", err)
	}

	got := strings.Join(seen, "\n")
	for _, want := range []string{
		"GET /v1/workflows/versions?workflow_id=wf_123",
		"GET /v1/workflows/versions/diff?from_workflow_version_id=wfv_0&to_workflow_version_id=wfv_1&workflow_id=wf_123",
		"GET /v1/workflows/versions/wfv_1?workflow_id=wf_123",
		"POST /v1/workflows/versions/wfv_1/restore?workflow_id=wf_123",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing request %q in:\n%s", want, got)
		}
	}
}

func TestWorkflowsBlockAndEdgeVersionCommandsUseGeneratedRoutes(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/versions"):
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "list_metadata": map[string]any{}})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/versions/diff"):
			_ = json.NewEncoder(w).Encode(map[string]any{"changes": []any{}})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/blocks/versions/bv_1":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "bv_1", "workflow_id": "wf_123"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/blocks/versions/bv_1/restore":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "block_1", "type": "extract"})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/edges/versions/ev_1":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "ev_1", "workflow_id": "wf_123"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/edges/versions/ev_1/restore":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "edge_1", "source_block": "a", "target_block": "b"})
		default:
			t.Fatalf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksVersionsCmd.RunE(workflowsBlocksVersionsCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("blocks versions: %v", err)
	}
	if err := workflowsBlocksDiffCmd.RunE(workflowsBlocksDiffCmd, []string{"bv_0", "bv_1"}); err != nil {
		t.Fatalf("blocks diff: %v", err)
	}
	if err := workflowsBlocksVersionCmd.RunE(workflowsBlocksVersionCmd, []string{"bv_1"}); err != nil {
		t.Fatalf("blocks version: %v", err)
	}
	if err := workflowsBlocksVersionRestoreCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksVersionRestoreCmd.Flags().Set("yes", "false") })
	if err := workflowsBlocksVersionRestoreCmd.RunE(workflowsBlocksVersionRestoreCmd, []string{"bv_1"}); err != nil {
		t.Fatalf("blocks version-restore: %v", err)
	}

	if err := workflowsEdgesVersionsCmd.RunE(workflowsEdgesVersionsCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("edges versions: %v", err)
	}
	if err := workflowsEdgesDiffCmd.RunE(workflowsEdgesDiffCmd, []string{"ev_0", "ev_1"}); err != nil {
		t.Fatalf("edges diff: %v", err)
	}
	if err := workflowsEdgesVersionCmd.RunE(workflowsEdgesVersionCmd, []string{"ev_1"}); err != nil {
		t.Fatalf("edges version: %v", err)
	}
	if err := workflowsEdgesVersionRestoreCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsEdgesVersionRestoreCmd.Flags().Set("yes", "false") })
	if err := workflowsEdgesVersionRestoreCmd.RunE(workflowsEdgesVersionRestoreCmd, []string{"ev_1"}); err != nil {
		t.Fatalf("edges version-restore: %v", err)
	}

	got := strings.Join(seen, "\n")
	for _, want := range []string{
		"GET /v1/workflows/blocks/versions?workflow_id=wf_123",
		"GET /v1/workflows/blocks/versions/diff?from_block_version_id=bv_0&to_block_version_id=bv_1",
		"GET /v1/workflows/blocks/versions/bv_1",
		"POST /v1/workflows/blocks/versions/bv_1/restore",
		"GET /v1/workflows/edges/versions?workflow_id=wf_123",
		"GET /v1/workflows/edges/versions/diff?from_edge_version_id=ev_0&to_edge_version_id=ev_1",
		"GET /v1/workflows/edges/versions/ev_1",
		"POST /v1/workflows/edges/versions/ev_1/restore",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing request %q in:\n%s", want, got)
		}
	}
}
