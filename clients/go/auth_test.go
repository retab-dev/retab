package retab

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewClientStillRequiresCredentials guards the no-creds regression: when
// neither an API key nor a bearer token is supplied the constructor must fail
// fast with a clear message rather than silently authenticate as nobody.
func TestNewClientStillRequiresCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")

	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error when no credentials provided")
	}
	if !strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("expected helpful 'no credentials' message, got %q", err.Error())
	}
}

// TestApiKeyAuthSendsApiKeyHeader is the baseline: the legacy API-key path
// must continue to send `Api-Key` and must NOT send `Authorization`.
func TestApiKeyAuthSendsApiKeyHeader(t *testing.T) {
	var gotApiKey, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotApiKey = r.Header.Get("Api-Key")
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	client, err := NewClient("sk_test_abc", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Files.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if gotApiKey != "sk_test_abc" {
		t.Errorf("Api-Key header: got %q, want %q", gotApiKey, "sk_test_abc")
	}
	if gotAuth != "" {
		t.Errorf("Authorization header should be empty for API-key auth, got %q", gotAuth)
	}
}

// TestBearerTokenAuthSendsAuthorizationHeader exercises the new OAuth path.
// Critically, the Api-Key header must NOT be sent — otherwise the backend's
// auth cascade gets two credentials and the unused one shows up in audit
// logs as a 'failed' attempt.
func TestBearerTokenAuthSendsAuthorizationHeader(t *testing.T) {
	var gotApiKey, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotApiKey = r.Header.Get("Api-Key")
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	client, err := NewClient("",
		WithBaseURL(srv.URL),
		WithBearerToken("workos_at_xyz"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Files.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if gotAuth != "Bearer workos_at_xyz" {
		t.Errorf("Authorization header: got %q, want %q", gotAuth, "Bearer workos_at_xyz")
	}
	if gotApiKey != "" {
		t.Errorf("Api-Key header should be empty for Bearer auth, got %q", gotApiKey)
	}
}

// TestBearerTokenProviderIsCalledPerRequest verifies the refresh hook: the
// provider must be invoked on every request, not cached at construction.
// This is what enables transparent token refresh in the CLI.
func TestBearerTokenProviderIsCalledPerRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	calls := 0
	provider := func(context.Context) (string, error) {
		calls++
		return fmt.Sprintf("token_%d", calls), nil
	}

	client, err := NewClient("",
		WithBaseURL(srv.URL),
		WithBearerTokenProvider(provider),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Three independent requests.
	for i := 0; i < 3; i++ {
		if _, err := client.Files.List(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	}

	if calls != 3 {
		t.Errorf("provider invocations: got %d, want 3", calls)
	}
}

// TestBearerTokenProviderErrorAborts ensures that a failing provider stops
// the request before any network IO — otherwise the user would see a
// confusing 401 from the server instead of the actual error.
func TestBearerTokenProviderErrorAborts(t *testing.T) {
	reached := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
	}))
	defer srv.Close()

	client, err := NewClient("",
		WithBaseURL(srv.URL),
		WithBearerTokenProvider(func(context.Context) (string, error) {
			return "", fmt.Errorf("refresh token expired")
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Files.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from failing provider")
	}
	if !strings.Contains(err.Error(), "refresh token expired") {
		t.Errorf("expected wrapped provider error, got %q", err.Error())
	}
	if reached {
		t.Error("server was reached even though provider failed; should have aborted before network IO")
	}
}

// TestEmptyBearerTokenIsRejected guards against silently sending
// `Authorization: Bearer ` (with no token) — that's worse than sending
// nothing because the server might treat the empty string differently.
func TestEmptyBearerTokenIsRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be reached")
	}))
	defer srv.Close()

	client, err := NewClient("",
		WithBaseURL(srv.URL),
		WithBearerTokenProvider(func(context.Context) (string, error) { return "", nil }),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Files.List(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "empty token") {
		t.Errorf("expected 'empty token' error, got %v", err)
	}
}
