package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWorkflowsListFieldsTrimmingSurvivesToOutput pins that `workflows list
// --fields` trimming reaches the printed JSON. The server honors the
// `fields` query param and returns only the requested keys; the CLI must
// not re-inflate the typed Workflow struct's zero-valued fields (name,
// description, email_trigger) back into the output — that would defeat the
// whole point of the flag ("trim large list payloads").
func TestWorkflowsListFieldsTrimmingSurvivesToOutput(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("fields"); got != "id" {
			t.Fatalf("fields query = %q, want id", got)
		}
		w.Header().Set("Content-Type", "application/json")
		// Server honors `fields=id` and trims each row to just that key.
		_, _ = w.Write([]byte(`{"data":[{"id":"wf_trim_1"}],"list_metadata":{"after":"wf_trim_1"}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := workflowsListCmd.Flags().Set("fields", "id"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsListCmd.Flags().Set("fields", "") })

	workflowsListCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsListCmd.SetContext(nil) })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsListCmd.RunE(workflowsListCmd, nil); err != nil {
			t.Fatalf("workflows list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "wf_trim_1") {
		t.Fatalf("expected workflow id in output:\n%s", stdout)
	}
	for _, leaked := range []string{`"name"`, `"description"`, `"email_trigger"`} {
		if strings.Contains(stdout, leaked) {
			t.Fatalf("%s leaked into --fields id output (typed-struct re-inflation):\n%s", leaked, stdout)
		}
	}
}
