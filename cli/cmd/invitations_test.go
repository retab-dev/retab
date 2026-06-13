package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestInvitationsListHitsInternalEndpoint pins the list path/method and that
// the bare array renders.
func TestInvitationsListHitsInternalEndpoint(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath, seenMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"invitation_1","email":"new@acme.com","state":"pending","role":null,"created_at":null,"expires_at":null}]`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, err := captureStdAndRun(t, func() error {
		return invitationsListCmd.RunE(invitationsListCmd, nil)
	})
	if err != nil {
		t.Fatalf("invitations list: %v", err)
	}
	if seenMethod != http.MethodGet || seenPath != "/internal/workos/organizations/invitations" {
		t.Fatalf("got %s %s, want GET /internal/workos/organizations/invitations", seenMethod, seenPath)
	}
	if !strings.Contains(stdout, "invitation_1") || !strings.Contains(stdout, "new@acme.com") {
		t.Fatalf("invitations list output missing invitation, got:\n%s", stdout)
	}
}

// TestInvitationsSendPostsEmailAndRole pins method, path, and body for send.
func TestInvitationsSendPostsEmailAndRole(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"id":"invitation_1","email":"new@acme.com","state":"pending","role":"admin","created_at":null,"expires_at":null}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := invitationsCreateCmd.Flags().Set("email", "new@acme.com"); err != nil {
		t.Fatal(err)
	}
	if err := invitationsCreateCmd.Flags().Set("role", "admin"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = invitationsCreateCmd.Flags().Set("email", "")
		_ = invitationsCreateCmd.Flags().Set("role", "")
	})

	if _, err := captureStdAndRun(t, func() error {
		return invitationsCreateCmd.RunE(invitationsCreateCmd, nil)
	}); err != nil {
		t.Fatalf("invitations send: %v", err)
	}
	if seenMethod != http.MethodPost || seenPath != "/internal/workos/organizations/invitations" {
		t.Fatalf("got %s %s, want POST /internal/workos/organizations/invitations", seenMethod, seenPath)
	}
	if seenBody["email"] != "new@acme.com" || seenBody["role"] != "admin" {
		t.Fatalf("body = %v, want email+role set", seenBody)
	}
}

// TestInvitationsSendOmitsRoleWhenAbsent pins that no --role means the body
// carries no role key (server defaults to member).
func TestInvitationsSendOmitsRoleWhenAbsent(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"invitation_1","email":"new@acme.com","state":"pending","role":null,"created_at":null,"expires_at":null}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := invitationsCreateCmd.Flags().Set("email", "new@acme.com"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = invitationsCreateCmd.Flags().Set("email", "") })

	if _, err := captureStdAndRun(t, func() error {
		return invitationsCreateCmd.RunE(invitationsCreateCmd, nil)
	}); err != nil {
		t.Fatalf("invitations send: %v", err)
	}
	if _, ok := seenBody["role"]; ok {
		t.Fatalf("body should omit role when --role absent, got %v", seenBody)
	}
}

// TestInvitationsRevokeDeletes pins method and path for revoke (with --yes).
func TestInvitationsRevokeDeletes(t *testing.T) {
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

	if err := invitationsDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = invitationsDeleteCmd.Flags().Set("yes", "false") })

	if _, err := captureStdAndRun(t, func() error {
		return invitationsDeleteCmd.RunE(invitationsDeleteCmd, []string{"invitation_1"})
	}); err != nil {
		t.Fatalf("invitations revoke: %v", err)
	}
	if seenMethod != http.MethodDelete || seenPath != "/internal/workos/organizations/invitations/invitation_1" {
		t.Fatalf("got %s %s, want DELETE .../invitations/invitation_1", seenMethod, seenPath)
	}
}
