package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWorkflowsArtifactsGet exercises the restored `workflows artifacts get`
// command. The HTTP route is `GET /workflows/artifacts/{artifact_id}` — the
// server infers the operation from the id prefix, so the CLI only needs the id.
func TestWorkflowsArtifactsGet(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/workflows/artifacts/extr_lz1_abc" {
			t.Fatalf("path = %s, want /workflows/artifacts/extr_lz1_abc", r.URL.Path)
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

	if len(requests) != 1 || requests[0] != "GET /workflows/artifacts/extr_lz1_abc" {
		t.Fatalf("requests = %v, want one GET /workflows/artifacts/extr_lz1_abc", requests)
	}
	if !strings.Contains(stdout, `"id": "extr_lz1_abc"`) {
		t.Fatalf("expected stdout to contain the artifact id, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"operation": "extraction"`) {
		t.Fatalf("expected stdout to contain the operation, got:\n%s", stdout)
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
