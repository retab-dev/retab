//go:build !retab_oagen_cli_workflows_artifacts

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWorkflowsArtifactsGet exercises the restored `workflows artifacts get`
// command. The HTTP route is `GET /v1/workflows/artifacts/{artifact_id}` — the
// server infers the operation from the id prefix, so the CLI only needs the id.
func TestWorkflowsArtifactsGet(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/workflows/artifacts/extr_lz1_abc" {
			t.Fatalf("path = %s, want /v1/workflows/artifacts/extr_lz1_abc", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "extr_lz1_abc",
			"operation": "extraction",
			"payload":   map[string]any{"hello": "world"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := workflowsArtifactsGetCmd.RunE(workflowsArtifactsGetCmd, []string{"extr_lz1_abc"}); err != nil {
			t.Fatalf("artifacts get: %v", err)
		}
	})

	if len(requests) != 1 || requests[0] != "GET /v1/workflows/artifacts/extr_lz1_abc" {
		t.Fatalf("requests = %v, want one GET /v1/workflows/artifacts/extr_lz1_abc", requests)
	}
	if !strings.Contains(stdout, `"id": "extr_lz1_abc"`) {
		t.Fatalf("expected stdout to contain the artifact id, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"operation": "extraction"`) {
		t.Fatalf("expected stdout to contain the operation, got:\n%s", stdout)
	}
}

func TestWorkflowsArtifactsListPreservesDereferencedFields(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	if err := rootCmd.PersistentFlags().Set("output", "json"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/workflows/artifacts" {
			t.Fatalf("path = %s, want /v1/workflows/artifacts", r.URL.Path)
		}
		if r.URL.Query().Get("run_id") != "run_xyz" {
			t.Fatalf("run_id = %q, want run_xyz", r.URL.Query().Get("run_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"id":        "fninv_abc",
				"operation": "function_invocation",
				"run_id":    "run_xyz",
				"step_id":   "step_function",
				"block_id":  "block_function",
				"output":    map[string]any{"ok": true},
			}},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := workflowsArtifactsListCmd.RunE(workflowsArtifactsListCmd, []string{"run_xyz"}); err != nil {
			t.Fatalf("artifacts list: %v", err)
		}
	})

	if !strings.Contains(stdout, `"step_id": "step_function"`) {
		t.Fatalf("expected stdout to contain step_id, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"block_id": "block_function"`) {
		t.Fatalf("expected stdout to contain block_id, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"output": {`) || !strings.Contains(stdout, `"ok": true`) {
		t.Fatalf("expected stdout to contain artifact output, got:\n%s", stdout)
	}
}

// TestWorkflowsArtifactsCmdLongHasNoLeadingTabs guards Bug B: the second
// paragraph of `workflows artifacts --help` was hard-tab-indented in the Go
// string literal, producing a visibly broken help block. Every line of the
// Long block must align with the first paragraph (no leading tab).
func TestWorkflowsArtifactsCmdLongHasNoLeadingTabs(t *testing.T) {
	for i, line := range strings.Split(workflowsArtifactsCmd.Long, "\n") {
		if strings.HasPrefix(line, "\t") {
			t.Fatalf("workflowsArtifactsCmd.Long line %d starts with a tab: %q", i+1, line)
		}
	}
}
