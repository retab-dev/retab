package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// specApplyTestServer returns a server that answers plan (no destroy) + apply
// success, for exercising the apply command's client-side behavior.
func specApplyTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		base := map[string]any{
			"workflow_id": "wrk_created_new",
			"action":      "create",
			"created":     true,
			"diagnostics": map[string]any{"is_valid": true},
			"summary": map[string]any{
				"add": 1, "change": 0, "destroy": 0, "replace": 0, "noop": 0,
				"total": 1, "has_changes": true,
			},
		}
		_ = json.NewEncoder(w).Encode(base)
	}))
}

const specWithMetadataID = `apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  id: wrk_source_123
  name: Source
spec:
  blocks:
    start:
      type: start_json
`

const specWithoutMetadataID = `apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  name: Fresh
spec:
  blocks:
    start:
      type: start_json
`

// TestWorkflowsSpecApplyWarnsOnMetadataIDWithProjectID pins that applying a spec
// that carries metadata.id with --project-id (create-new) warns about the
// silent-duplicate footgun and points at --to.
func TestWorkflowsSpecApplyWarnsOnMetadataIDWithProjectID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	server := specApplyTestServer(t)
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "wf.yaml")
	if err := os.WriteFile(path, []byte(specWithMetadataID), 0o644); err != nil {
		t.Fatal(err)
	}

	withSpecApplyProjectID(t)
	_, stderr := captureStd(t, func() {
		if err := workflowsSpecApplyCmd.RunE(workflowsSpecApplyCmd, []string{path}); err != nil {
			t.Fatalf("spec apply: %v", err)
		}
	})
	for _, want := range []string{"wrk_source_123", "NEW workflow", "--to wrk_source_123"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("warning missing %q; stderr:\n%s", want, stderr)
		}
	}
}

// TestWorkflowsSpecApplyNoWarnWithoutMetadataID pins that a spec with no
// metadata.id applies cleanly with no footgun warning.
func TestWorkflowsSpecApplyNoWarnWithoutMetadataID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	server := specApplyTestServer(t)
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "wf.yaml")
	if err := os.WriteFile(path, []byte(specWithoutMetadataID), 0o644); err != nil {
		t.Fatal(err)
	}

	withSpecApplyProjectID(t)
	_, stderr := captureStd(t, func() {
		if err := workflowsSpecApplyCmd.RunE(workflowsSpecApplyCmd, []string{path}); err != nil {
			t.Fatalf("spec apply: %v", err)
		}
	})
	if strings.Contains(stderr, "metadata.id") || strings.Contains(stderr, "NEW workflow") {
		t.Fatalf("unexpected footgun warning for a spec without metadata.id; stderr:\n%s", stderr)
	}
}
