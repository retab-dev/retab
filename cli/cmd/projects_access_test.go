package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestProjectsAccessListHitsMembershipsEndpoint pins path/method + that the
// project-id and filter flags become query params.
func TestProjectsAccessListHitsMembershipsEndpoint(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod, seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod, seenQuery = r.URL.Path, r.Method, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"pmem_1","project_id":"proj_1","subject_type":"user","subject_id":"user_1","role":"project-editor","is_active":true}],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := projectsAccessListCmd.Flags().Set("project-id", "proj_1"); err != nil {
		t.Fatal(err)
	}
	if err := projectsAccessListCmd.Flags().Set("role", "project-editor"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = projectsAccessListCmd.Flags().Set("project-id", "")
		_ = projectsAccessListCmd.Flags().Set("role", "")
	})

	stdout, err := captureStdAndRun(t, func() error {
		return projectsAccessListCmd.RunE(projectsAccessListCmd, nil)
	})
	if err != nil {
		t.Fatalf("projects access list: %v", err)
	}
	if seenMethod != http.MethodGet || seenPath != "/v1/project-memberships" {
		t.Fatalf("got %s %s, want GET /v1/project-memberships", seenMethod, seenPath)
	}
	for _, want := range []string{"project_id=proj_1", "role=project-editor"} {
		if !strings.Contains(seenQuery, want) {
			t.Fatalf("query %q missing %q", seenQuery, want)
		}
	}
	if !strings.Contains(stdout, "pmem_1") {
		t.Fatalf("list output missing grant, got:\n%s", stdout)
	}
}

// TestProjectsAccessListAcceptsPositionalID pins that the project id can be
// passed positionally (the preferred form) and reaches the query.
func TestProjectsAccessListAcceptsPositionalID(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if _, err := captureStdAndRun(t, func() error {
		return projectsAccessListCmd.RunE(projectsAccessListCmd, []string{"proj_pos"})
	}); err != nil {
		t.Fatalf("projects access list <positional>: %v", err)
	}
	if !strings.Contains(seenQuery, "project_id=proj_pos") {
		t.Fatalf("query %q missing project_id=proj_pos", seenQuery)
	}
}

// TestProjectsAccessListPositionalFlagConflict pins that a positional id that
// disagrees with --project-id is a clean error.
func TestProjectsAccessListPositionalFlagConflict(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("RETAB_API_BASE_URL", "http://127.0.0.1:0")

	if err := projectsAccessListCmd.Flags().Set("project-id", "proj_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = projectsAccessListCmd.Flags().Set("project-id", "") })

	_, err := captureStdAndRun(t, func() error {
		return projectsAccessListCmd.RunE(projectsAccessListCmd, []string{"proj_pos"})
	})
	if err == nil || !strings.Contains(err.Error(), "conflicting project-id") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

// TestProjectsAccessGrantPostsBody pins method, path, and the create body.
func TestProjectsAccessGrantPostsBody(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"id":"pmem_1","project_id":"proj_1","subject_type":"user","subject_id":"user_1","role":"project-editor","is_active":true}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, val := range map[string]string{"project-id": "proj_1", "subject-id": "user_1", "role": "project-editor"} {
		if err := projectsAccessGrantCmd.Flags().Set(flag, val); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = projectsAccessGrantCmd.Flags().Set("project-id", "")
		_ = projectsAccessGrantCmd.Flags().Set("subject-id", "")
		_ = projectsAccessGrantCmd.Flags().Set("role", "")
		_ = projectsAccessGrantCmd.Flags().Set("subject-type", "user")
	})

	if _, err := captureStdAndRun(t, func() error {
		return projectsAccessGrantCmd.RunE(projectsAccessGrantCmd, nil)
	}); err != nil {
		t.Fatalf("projects access grant: %v", err)
	}
	if seenMethod != http.MethodPost || seenPath != "/v1/project-memberships" {
		t.Fatalf("got %s %s, want POST /v1/project-memberships", seenMethod, seenPath)
	}
	if seenBody["project_id"] != "proj_1" || seenBody["subject_id"] != "user_1" ||
		seenBody["role"] != "project-editor" || seenBody["subject_type"] != "user" {
		t.Fatalf("unexpected body: %v", seenBody)
	}
}

// TestProjectsAccessGrantResolvesEmail pins that --email is resolved to a user id
// via the member list and sent as subject_id.
func TestProjectsAccessGrantResolvesEmail(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var membersListed bool
	var seenPostBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/internal/workos/organizations/members" && r.Method == http.MethodGet:
			membersListed = true
			_, _ = w.Write([]byte(`[{"id":"user_alice","email":"alice@acme.com","role":"member"}]`))
		case r.URL.Path == "/v1/project-memberships" && r.Method == http.MethodPost:
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &seenPostBody)
			_, _ = w.Write([]byte(`{"id":"pmem_1","project_id":"proj_1","subject_type":"user","subject_id":"user_alice","role":"project-editor","is_active":true}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, val := range map[string]string{"project-id": "proj_1", "email": "alice@acme.com", "role": "project-editor"} {
		if err := projectsAccessGrantCmd.Flags().Set(flag, val); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = projectsAccessGrantCmd.Flags().Set("project-id", "")
		_ = projectsAccessGrantCmd.Flags().Set("email", "")
		_ = projectsAccessGrantCmd.Flags().Set("role", "")
	})

	if _, err := captureStdAndRun(t, func() error {
		return projectsAccessGrantCmd.RunE(projectsAccessGrantCmd, nil)
	}); err != nil {
		t.Fatalf("projects access grant --email: %v", err)
	}
	if !membersListed {
		t.Fatalf("expected member list lookup for email resolution")
	}
	if seenPostBody["subject_id"] != "user_alice" || seenPostBody["subject_type"] != "user" {
		t.Fatalf("resolved body = %v, want subject_id=user_alice subject_type=user", seenPostBody)
	}
}

// TestProjectsAccessGrantEmailSubjectIDMutuallyExclusive pins the XOR guard.
func TestProjectsAccessGrantEmailSubjectIDMutuallyExclusive(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("RETAB_API_BASE_URL", "http://127.0.0.1:0")

	for flag, val := range map[string]string{"project-id": "proj_1", "email": "alice@acme.com", "subject-id": "user_1", "role": "project-editor"} {
		if err := projectsAccessGrantCmd.Flags().Set(flag, val); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = projectsAccessGrantCmd.Flags().Set("project-id", "")
		_ = projectsAccessGrantCmd.Flags().Set("email", "")
		_ = projectsAccessGrantCmd.Flags().Set("subject-id", "")
		_ = projectsAccessGrantCmd.Flags().Set("role", "")
	})

	_, err := captureStdAndRun(t, func() error {
		return projectsAccessGrantCmd.RunE(projectsAccessGrantCmd, nil)
	})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutually-exclusive error, got %v", err)
	}
}

// TestProjectsAccessGrantRejectsBadRole pins client-side role validation.
func TestProjectsAccessGrantRejectsBadRole(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("RETAB_API_BASE_URL", "http://127.0.0.1:0")

	for flag, val := range map[string]string{"project-id": "proj_1", "subject-id": "user_1", "role": "admin"} {
		if err := projectsAccessGrantCmd.Flags().Set(flag, val); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = projectsAccessGrantCmd.Flags().Set("project-id", "")
		_ = projectsAccessGrantCmd.Flags().Set("subject-id", "")
		_ = projectsAccessGrantCmd.Flags().Set("role", "")
	})

	_, err := captureStdAndRun(t, func() error {
		return projectsAccessGrantCmd.RunE(projectsAccessGrantCmd, nil)
	})
	if err == nil || !strings.Contains(err.Error(), "invalid --role") {
		t.Fatalf("expected invalid-role error, got %v", err)
	}
}

// TestProjectsAccessUpdatePatchesRole pins method, path, and body.
func TestProjectsAccessUpdatePatchesRole(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"id":"pmem_1","project_id":"proj_1","subject_type":"user","subject_id":"user_1","role":"project-viewer","is_active":true}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := projectsAccessUpdateCmd.Flags().Set("role", "project-viewer"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = projectsAccessUpdateCmd.Flags().Set("role", "") })

	if _, err := captureStdAndRun(t, func() error {
		return projectsAccessUpdateCmd.RunE(projectsAccessUpdateCmd, []string{"pmem_1"})
	}); err != nil {
		t.Fatalf("projects access update: %v", err)
	}
	if seenMethod != http.MethodPatch || seenPath != "/v1/project-memberships/pmem_1" {
		t.Fatalf("got %s %s, want PATCH .../project-memberships/pmem_1", seenMethod, seenPath)
	}
	if seenBody["role"] != "project-viewer" {
		t.Fatalf("body role = %v, want project-viewer", seenBody["role"])
	}
}

// TestProjectsAccessRevokeDeletes pins method and path (with --yes).
func TestProjectsAccessRevokeDeletes(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"pmem_1","project_id":"proj_1","subject_type":"user","subject_id":"user_1","role":"project-editor","is_active":false}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := projectsAccessRevokeCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = projectsAccessRevokeCmd.Flags().Set("yes", "false") })

	if _, err := captureStdAndRun(t, func() error {
		return projectsAccessRevokeCmd.RunE(projectsAccessRevokeCmd, []string{"pmem_1"})
	}); err != nil {
		t.Fatalf("projects access revoke: %v", err)
	}
	if seenMethod != http.MethodDelete || seenPath != "/v1/project-memberships/pmem_1" {
		t.Fatalf("got %s %s, want DELETE .../project-memberships/pmem_1", seenMethod, seenPath)
	}
}
