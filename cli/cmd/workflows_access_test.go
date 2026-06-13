package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWorkflowsAccessListHitsMembershipsEndpoint pins path/method + workflow-id
// query param.
func TestWorkflowsAccessListHitsMembershipsEndpoint(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod, seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod, seenQuery = r.URL.Path, r.Method, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"wmem_1","workflow_id":"wf_1","subject_type":"user","subject_id":"user_1","role":"workflow-viewer","is_active":true}],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsAccessListCmd.Flags().Set("workflow-id", "wf_1"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsAccessListCmd.Flags().Set("workflow-id", "") })

	stdout, err := captureStdAndRun(t, func() error {
		return workflowsAccessListCmd.RunE(workflowsAccessListCmd, nil)
	})
	if err != nil {
		t.Fatalf("workflows access list: %v", err)
	}
	if seenMethod != http.MethodGet || seenPath != "/v1/workflow-memberships" {
		t.Fatalf("got %s %s, want GET /v1/workflow-memberships", seenMethod, seenPath)
	}
	if !strings.Contains(seenQuery, "workflow_id=wf_1") {
		t.Fatalf("query %q missing workflow_id=wf_1", seenQuery)
	}
	if !strings.Contains(stdout, "wmem_1") {
		t.Fatalf("list output missing grant, got:\n%s", stdout)
	}
}

// TestWorkflowsAccessUpdatePatchesRole pins method, path, body.
func TestWorkflowsAccessUpdatePatchesRole(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod string
	var seenBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod = r.URL.Path, r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"wmem_1","workflow_id":"wf_1","subject_type":"user","subject_id":"user_1","role":"workflow-viewer","is_active":true}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsAccessUpdateCmd.Flags().Set("role", "workflow-viewer"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsAccessUpdateCmd.Flags().Set("role", "") })

	if _, err := captureStdAndRun(t, func() error {
		return workflowsAccessUpdateCmd.RunE(workflowsAccessUpdateCmd, []string{"wmem_1"})
	}); err != nil {
		t.Fatalf("workflows access update: %v", err)
	}
	if seenMethod != http.MethodPatch || seenPath != "/v1/workflow-memberships/wmem_1" {
		t.Fatalf("got %s %s, want PATCH .../workflow-memberships/wmem_1", seenMethod, seenPath)
	}
	if seenBody["role"] != "workflow-viewer" {
		t.Fatalf("body role = %v, want workflow-viewer", seenBody["role"])
	}
}

// TestWorkflowsAccessUpdateRejectsBadRole pins that a project-* role is rejected
// for a workflow grant.
func TestWorkflowsAccessUpdateRejectsBadRole(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("RETAB_API_BASE_URL", "http://127.0.0.1:0")

	if err := workflowsAccessUpdateCmd.Flags().Set("role", "project-viewer"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsAccessUpdateCmd.Flags().Set("role", "") })

	_, err := captureStdAndRun(t, func() error {
		return workflowsAccessUpdateCmd.RunE(workflowsAccessUpdateCmd, []string{"wmem_1"})
	})
	if err == nil || !strings.Contains(err.Error(), "invalid --role") {
		t.Fatalf("expected invalid-role error, got %v", err)
	}
}

// TestWorkflowsAccessRevokeDeletes pins method and path (with --yes).
func TestWorkflowsAccessRevokeDeletes(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"wmem_1","workflow_id":"wf_1","subject_type":"user","subject_id":"user_1","role":"workflow-viewer","is_active":false}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsAccessRevokeCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsAccessRevokeCmd.Flags().Set("yes", "false") })

	if _, err := captureStdAndRun(t, func() error {
		return workflowsAccessRevokeCmd.RunE(workflowsAccessRevokeCmd, []string{"wmem_1"})
	}); err != nil {
		t.Fatalf("workflows access revoke: %v", err)
	}
	if seenMethod != http.MethodDelete || seenPath != "/v1/workflow-memberships/wmem_1" {
		t.Fatalf("got %s %s, want DELETE .../workflow-memberships/wmem_1", seenMethod, seenPath)
	}
}
