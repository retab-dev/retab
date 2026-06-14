package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWorkflowsExperimentsListForwardsPagination pins that the pagination flags
// added to `workflows experiments list` (--limit / --order / --before / --after)
// reach the request query. The backend already accepted these; the CLI just
// hadn't exposed them.
func TestWorkflowsExperimentsListForwardsPagination(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, val := range map[string]string{"limit": "5", "order": "asc", "after": "exp_1"} {
		if err := workflowsExperimentsListCmd.Flags().Set(flag, val); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = workflowsExperimentsListCmd.Flags().Set("limit", "0")
		_ = workflowsExperimentsListCmd.Flags().Set("order", "")
		_ = workflowsExperimentsListCmd.Flags().Set("after", "")
	})

	if _, err := captureStdAndRun(t, func() error {
		return workflowsExperimentsListCmd.RunE(workflowsExperimentsListCmd, []string{"wrk_1"})
	}); err != nil {
		t.Fatalf("experiments list: %v", err)
	}

	for _, want := range []string{"workflow_id=wrk_1", "limit=5", "order=asc", "after=exp_1"} {
		if !strings.Contains(seenQuery, want) {
			t.Fatalf("query %q missing %q", seenQuery, want)
		}
	}
}
