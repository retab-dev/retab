package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMembersListHitsInternalEndpoint pins that `retab members list` fetches
// the internal org-members surface and renders the bare array.
func TestMembersListHitsInternalEndpoint(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"user_1","email":"a@acme.com","role":"admin","first_name":"Ada","last_name":"Lovelace"}]`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, err := captureStdAndRun(t, func() error {
		return membersListCmd.RunE(membersListCmd, nil)
	})
	if err != nil {
		t.Fatalf("members list: %v", err)
	}
	if seenMethod != http.MethodGet || seenPath != "/internal/workos/organizations/members" {
		t.Fatalf("got %s %s, want GET /internal/workos/organizations/members", seenMethod, seenPath)
	}
	if !strings.Contains(stdout, "user_1") || !strings.Contains(stdout, "a@acme.com") {
		t.Fatalf("members list output missing member, got:\n%s", stdout)
	}
}

// TestMembersSetRolePatchesRole pins method, path, and body for set-role.
func TestMembersSetRolePatchesRole(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"id":"user_1","email":"a@acme.com","role":"admin"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := membersUpdateCmd.Flags().Set("role", "admin"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = membersUpdateCmd.Flags().Set("role", "") })

	if _, err := captureStdAndRun(t, func() error {
		return membersUpdateCmd.RunE(membersUpdateCmd, []string{"user_1"})
	}); err != nil {
		t.Fatalf("members set-role: %v", err)
	}
	if seenMethod != http.MethodPatch || seenPath != "/internal/workos/organizations/members/user_1/role" {
		t.Fatalf("got %s %s, want PATCH .../members/user_1/role", seenMethod, seenPath)
	}
	if seenBody["role"] != "admin" {
		t.Fatalf("body role = %v, want admin", seenBody["role"])
	}
}

// TestMembersUpdateResolvesEmail pins that an email positional arg is resolved to
// a user id via the member list before the role PATCH.
func TestMembersUpdateResolvesEmail(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var patchedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/internal/workos/organizations/members" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`[{"id":"user_alice","email":"alice@acme.com","role":"member"}]`))
		case r.Method == http.MethodPatch:
			patchedPath = r.URL.Path
			_, _ = w.Write([]byte(`{"id":"user_alice","email":"alice@acme.com","role":"admin"}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := membersUpdateCmd.Flags().Set("role", "admin"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = membersUpdateCmd.Flags().Set("role", "") })

	if _, err := captureStdAndRun(t, func() error {
		return membersUpdateCmd.RunE(membersUpdateCmd, []string{"alice@acme.com"})
	}); err != nil {
		t.Fatalf("members update <email>: %v", err)
	}
	if patchedPath != "/internal/workos/organizations/members/user_alice/role" {
		t.Fatalf("patched %q, want .../members/user_alice/role", patchedPath)
	}
}

// TestMembersSetRoleRejectsUnknownRole pins client-side role validation.
func TestMembersSetRoleRejectsUnknownRole(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("RETAB_API_BASE_URL", "http://127.0.0.1:0")

	if err := membersUpdateCmd.Flags().Set("role", "owner"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = membersUpdateCmd.Flags().Set("role", "") })

	_, err := captureStdAndRun(t, func() error {
		return membersUpdateCmd.RunE(membersUpdateCmd, []string{"user_1"})
	})
	if err == nil || !strings.Contains(err.Error(), "invalid --role") {
		t.Fatalf("expected invalid-role error, got %v", err)
	}
}

// TestMembersRemoveDeletes pins method and path for remove (with --yes).
func TestMembersRemoveDeletes(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := membersRemoveCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = membersRemoveCmd.Flags().Set("yes", "false") })

	if _, err := captureStdAndRun(t, func() error {
		return membersRemoveCmd.RunE(membersRemoveCmd, []string{"user_1"})
	}); err != nil {
		t.Fatalf("members remove: %v", err)
	}
	if seenMethod != http.MethodDelete || seenPath != "/internal/workos/organizations/members/user_1" {
		t.Fatalf("got %s %s, want DELETE .../members/user_1", seenMethod, seenPath)
	}
}

// TestMembersPermissionsHitsEndpoint pins method and path for permissions.
func TestMembersPermissionsHitsEndpoint(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"projects":[{"id":"proj_1","name":"Invoices","roles":["editor"]}],"workflows":[]}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, err := captureStdAndRun(t, func() error {
		return membersPermissionsCmd.RunE(membersPermissionsCmd, []string{"user_1"})
	})
	if err != nil {
		t.Fatalf("members permissions: %v", err)
	}
	if seenMethod != http.MethodGet || seenPath != "/internal/workos/organizations/members/user_1/permissions" {
		t.Fatalf("got %s %s, want GET .../members/user_1/permissions", seenMethod, seenPath)
	}
	if !strings.Contains(stdout, "proj_1") {
		t.Fatalf("permissions output missing grant, got:\n%s", stdout)
	}
}
