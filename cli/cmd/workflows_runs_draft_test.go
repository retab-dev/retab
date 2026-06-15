package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newRunsCreateTestCmd builds a bare command wired to the real runs-create RunE
// with the flags the handler reads, matching the other runs-create tests.
func newRunsCreateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().Bool("draft", false, "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")
	return cmd
}

// TestWorkflowsRunsCreateDraftFlagSendsDraftVersion pins that --draft maps to
// version "draft" in the request body (alias for --version draft).
func TestWorkflowsRunsCreateDraftFlagSendsDraftVersion(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sentVersion any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		sentVersion = body["version"]
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "run_draft", "status": "running"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newRunsCreateTestCmd()
	if err := cmd.Flags().Set("draft", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("document-url", "block_start=https://storage.retab.com/org_x/invoice.pdf"); err != nil {
		t.Fatal(err)
	}
	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create --draft: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_draft") {
		t.Fatalf("expected run response, got:\n%s", stdout)
	}
	if sentVersion != "draft" {
		t.Fatalf("body version = %#v, want \"draft\"", sentVersion)
	}
}

// TestWorkflowsRunsCreateDraftConflictsWithVersion pins that --draft together
// with an explicit non-draft --version is rejected before any request.
func TestWorkflowsRunsCreateDraftConflictsWithVersion(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("request must not be sent on a flag conflict: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newRunsCreateTestCmd()
	_ = cmd.Flags().Set("draft", "true")
	_ = cmd.Flags().Set("version", "production")
	_ = cmd.Flags().Set("document-url", "block_start=https://storage.retab.com/org_x/invoice.pdf")

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil || !strings.Contains(err.Error(), "--draft conflicts with --version") {
		t.Fatalf("expected --draft/--version conflict error, got %v", err)
	}
}
