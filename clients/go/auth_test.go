package retab

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const legacyAPIKeyHeader = "Api" + "-Key"

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

// TestApiKeyAuthSendsAuthorizationHeader is the baseline: API keys use the
// same Authorization Bearer transport as access tokens.
func TestApiKeyAuthSendsAuthorizationHeader(t *testing.T) {
	var gotApiKey, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotApiKey = r.Header.Get(legacyAPIKeyHeader)
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	client, err := NewClient("sk_eval_abc", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Files.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if gotAuth != "Bearer sk_eval_abc" {
		t.Errorf("Authorization header: got %q, want %q", gotAuth, "Bearer sk_eval_abc")
	}
	if gotApiKey != "" {
		t.Errorf("legacy credential header should be empty for API-key auth, got %q", gotApiKey)
	}
}

// TestBearerTokenAuthSendsAuthorizationHeader exercises the new OAuth path.
// Critically, the legacy API-key header must NOT be sent.
func TestBearerTokenAuthSendsAuthorizationHeader(t *testing.T) {
	var gotApiKey, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotApiKey = r.Header.Get(legacyAPIKeyHeader)
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
		t.Errorf("legacy credential header should be empty for Bearer auth, got %q", gotApiKey)
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
