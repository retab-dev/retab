//go:build !retab_oagen_cli_workflows_runs

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestWorkflowsRunsCreateSendsMetadataPayload exercises the whole --metadata
// path end to end through the create command's RunE: repeatable key=value flags
// are parsed, attached to the request, and posted as a JSON `metadata` object.
// A value containing '=' must survive (strings.Cut splits on the first '=' only).
func TestWorkflowsRunsCreateSendsMetadataPayload(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var postedMetadata map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/blocks" && r.URL.Query().Get("workflow_id") == "wf_123" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []map[string]any{{"id": "block_start", "type": "start_document", "label": "Document"}},
				"list_metadata": map[string]any{},
			})
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		postedMetadata, _ = body["metadata"].(map[string]any)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "run_123", "status": "running"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "eval-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")
	cmd.Flags().StringArray("metadata", nil, "")

	if err := cmd.Flags().Set("document-url", "block_start=https://storage.retab.com/org_x/invoice.pdf"); err != nil {
		t.Fatal(err)
	}
	for _, pair := range []string{"customer=acme", "note=key=with=equals"} {
		if err := cmd.Flags().Set("metadata", pair); err != nil {
			t.Fatal(err)
		}
	}

	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_123") {
		t.Fatalf("expected run response on stdout, got:\n%s", stdout)
	}
	if postedMetadata["customer"] != "acme" {
		t.Fatalf("posted metadata customer = %#v, want acme (full=%#v)", postedMetadata["customer"], postedMetadata)
	}
	if postedMetadata["note"] != "key=with=equals" {
		t.Fatalf("posted metadata note = %#v, want key=with=equals", postedMetadata["note"])
	}
}

// TestWorkflowsRunsCreateRejectsMalformedMetadata pins the error path: a
// --metadata value with no '=' is rejected before any network call, with the
// flag name in the message (RunE wraps parseKVStringList's error as "--metadata:
// ...").
func TestWorkflowsRunsCreateRejectsMalformedMetadata(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "eval-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().String("json-inputs-file", "", "")
	cmd.Flags().StringArray("metadata", nil, "")

	if err := cmd.Flags().Set("metadata", "no-equals-sign"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected an error for malformed --metadata, got nil")
	}
	if !strings.Contains(err.Error(), "--metadata") {
		t.Fatalf("error %q should mention --metadata", err.Error())
	}
	if !strings.Contains(err.Error(), "no-equals-sign") {
		t.Fatalf("error %q should include the offending pair", err.Error())
	}
}
