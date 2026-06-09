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
	// diff via the flag form.
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
		t.Fatalf("diff (flags): %v", err)
	}
	// diff via the positional form (flags reset to empty first).
	_ = workflowsDiffCmd.Flags().Set("from-version-id", "")
	_ = workflowsDiffCmd.Flags().Set("to-version-id", "")
	if err := workflowsDiffCmd.RunE(workflowsDiffCmd, []string{"wf_123", "wfv_0", "wfv_1"}); err != nil {
		t.Fatalf("diff (positional): %v", err)
	}
	if err := workflowsVersionCmd.RunE(workflowsVersionCmd, []string{"wf_123", "wfv_1"}); err != nil {
		t.Fatalf("version: %v", err)
	}
	if err := workflowsVersionRestoreCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsVersionRestoreCmd.Flags().Set("yes", "false") })
	// version-restore via the positional form.
	if err := workflowsVersionRestoreCmd.RunE(workflowsVersionRestoreCmd, []string{"wf_123", "wfv_1"}); err != nil {
		t.Fatalf("version-restore (positional): %v", err)
	}
	// version-restore via the --version-id flag form.
	if err := workflowsVersionRestoreCmd.Flags().Set("version-id", "wfv_1"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsVersionRestoreCmd.Flags().Set("version-id", "") })
	if err := workflowsVersionRestoreCmd.RunE(workflowsVersionRestoreCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("version-restore (flag): %v", err)
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

func TestResolveDiffVersions(t *testing.T) {
	t.Run("positional supplies both", func(t *testing.T) {
		from, to, err := resolveDiffVersions([]string{"wf_1", "wfv_a", "wfv_b"}, "", "")
		if err != nil {
			t.Fatal(err)
		}
		if from != "wfv_a" || to != "wfv_b" {
			t.Fatalf("got (%q,%q)", from, to)
		}
	})

	t.Run("flags supply both", func(t *testing.T) {
		from, to, err := resolveDiffVersions([]string{"wf_1"}, "wfv_a", "wfv_b")
		if err != nil {
			t.Fatal(err)
		}
		if from != "wfv_a" || to != "wfv_b" {
			t.Fatalf("got (%q,%q)", from, to)
		}
	})

	t.Run("positional and flag resolve identically", func(t *testing.T) {
		fromPos, toPos, err := resolveDiffVersions([]string{"wf_1", "wfv_a", "wfv_b"}, "", "")
		if err != nil {
			t.Fatal(err)
		}
		fromFlag, toFlag, err := resolveDiffVersions([]string{"wf_1"}, "wfv_a", "wfv_b")
		if err != nil {
			t.Fatal(err)
		}
		if fromPos != fromFlag || toPos != toFlag {
			t.Fatalf("positional (%q,%q) != flag (%q,%q)", fromPos, toPos, fromFlag, toFlag)
		}
	})

	t.Run("matching positional and flag allowed", func(t *testing.T) {
		from, to, err := resolveDiffVersions([]string{"wf_1", "wfv_a", "wfv_b"}, "wfv_a", "wfv_b")
		if err != nil {
			t.Fatal(err)
		}
		if from != "wfv_a" || to != "wfv_b" {
			t.Fatalf("got (%q,%q)", from, to)
		}
	})

	t.Run("conflicting from errors", func(t *testing.T) {
		_, _, err := resolveDiffVersions([]string{"wf_1", "wfv_a", "wfv_b"}, "wfv_x", "")
		if err == nil {
			t.Fatal("expected conflict error")
		}
		if !strings.Contains(err.Error(), "--from-version-id") || !strings.Contains(err.Error(), "differ") {
			t.Fatalf("error should name flag and conflict: %v", err)
		}
	})

	t.Run("conflicting to errors", func(t *testing.T) {
		_, _, err := resolveDiffVersions([]string{"wf_1", "wfv_a", "wfv_b"}, "", "wfv_y")
		if err == nil {
			t.Fatal("expected conflict error")
		}
		if !strings.Contains(err.Error(), "--to-version-id") {
			t.Fatalf("error should name --to-version-id: %v", err)
		}
	})

	t.Run("missing from mentions both forms", func(t *testing.T) {
		_, _, err := resolveDiffVersions([]string{"wf_1"}, "", "wfv_b")
		if err == nil {
			t.Fatal("expected missing-from error")
		}
		if !strings.Contains(err.Error(), "positionally") || !strings.Contains(err.Error(), "--from-version-id") {
			t.Fatalf("error should mention both forms: %v", err)
		}
	})

	t.Run("missing to mentions both forms", func(t *testing.T) {
		_, _, err := resolveDiffVersions([]string{"wf_1"}, "wfv_a", "")
		if err == nil {
			t.Fatal("expected missing-to error")
		}
		if !strings.Contains(err.Error(), "positionally") || !strings.Contains(err.Error(), "--to-version-id") {
			t.Fatalf("error should mention both forms: %v", err)
		}
	})
}

func TestWorkflowsDiffArgs(t *testing.T) {
	t.Run("one positional ok", func(t *testing.T) {
		if err := workflowsDiffCmd.Args(workflowsDiffCmd, []string{"wf_1"}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("three positionals ok", func(t *testing.T) {
		if err := workflowsDiffCmd.Args(workflowsDiffCmd, []string{"wf_1", "wfv_a", "wfv_b"}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("two positionals rejected", func(t *testing.T) {
		err := workflowsDiffCmd.Args(workflowsDiffCmd, []string{"wf_1", "wfv_a"})
		if err == nil {
			t.Fatal("expected error for exactly two positionals")
		}
		if !strings.Contains(err.Error(), "both") {
			t.Fatalf("error should explain both ids are needed: %v", err)
		}
	})
	t.Run("four positionals rejected", func(t *testing.T) {
		if err := workflowsDiffCmd.Args(workflowsDiffCmd, []string{"wf_1", "a", "b", "c"}); err == nil {
			t.Fatal("expected error for four positionals")
		}
	})
}

func TestResolveRestoreVersionID(t *testing.T) {
	t.Run("positional", func(t *testing.T) {
		got, err := resolveRestoreVersionID([]string{"wf_1", "wfv_a"}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "wfv_a" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("flag", func(t *testing.T) {
		got, err := resolveRestoreVersionID([]string{"wf_1"}, "wfv_a")
		if err != nil {
			t.Fatal(err)
		}
		if got != "wfv_a" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("positional and flag resolve identically", func(t *testing.T) {
		pos, err := resolveRestoreVersionID([]string{"wf_1", "wfv_a"}, "")
		if err != nil {
			t.Fatal(err)
		}
		flag, err := resolveRestoreVersionID([]string{"wf_1"}, "wfv_a")
		if err != nil {
			t.Fatal(err)
		}
		if pos != flag {
			t.Fatalf("positional %q != flag %q", pos, flag)
		}
	})

	t.Run("matching positional and flag allowed", func(t *testing.T) {
		got, err := resolveRestoreVersionID([]string{"wf_1", "wfv_a"}, "wfv_a")
		if err != nil {
			t.Fatal(err)
		}
		if got != "wfv_a" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("conflict errors", func(t *testing.T) {
		_, err := resolveRestoreVersionID([]string{"wf_1", "wfv_a"}, "wfv_b")
		if err == nil {
			t.Fatal("expected conflict error")
		}
		if !strings.Contains(err.Error(), "--version-id") || !strings.Contains(err.Error(), "differ") {
			t.Fatalf("error should name flag and conflict: %v", err)
		}
	})

	t.Run("missing mentions both forms", func(t *testing.T) {
		_, err := resolveRestoreVersionID([]string{"wf_1"}, "")
		if err == nil {
			t.Fatal("expected missing error")
		}
		if !strings.Contains(err.Error(), "positionally") || !strings.Contains(err.Error(), "--version-id") {
			t.Fatalf("error should mention both forms: %v", err)
		}
	})
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
