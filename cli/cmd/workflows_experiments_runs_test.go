package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkflowsExperimentsRunsCreateUsesFlatMetaPattern(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/experiments/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "exprun_123",
			"workflow_id":   "wf_123",
			"experiment_id": "exp_123",
			"status":        "running",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsRunsCreateCmd.RunE(workflowsExperimentsRunsCreateCmd, []string{"wf_123", "exp_123"}); err != nil {
			t.Fatalf("experiment runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "exprun_123") {
		t.Fatalf("expected experiment run response on stdout, got:\n%s", stdout)
	}
	if body["workflow_id"] != "wf_123" || body["experiment_id"] != "exp_123" {
		t.Fatalf("experiment run create body = %#v", body)
	}
}
