package cmd

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// withTrustingTokenClient swaps the package-level tokenHTTPClient for one
// that trusts the test server's self-signed cert, and restores it after
// the test. Without this, the WorkOS endpoint round-trip would fail with
// x509 "unknown authority".
func withTrustingTokenClient(t *testing.T, c *http.Client) {
	t.Helper()
	prev := tokenHTTPClient
	tokenHTTPClient = c
	t.Cleanup(func() { tokenHTTPClient = prev })
}

// TestAccessTokenOrgID pins the org_id claim decode used to confirm an org
// switch landed where it was asked to.
func TestAccessTokenOrgID(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"org_id":"org_xyz","sub":"user_1"}`))
	if got := accessTokenOrgID("h." + payload + ".s"); got != "org_xyz" {
		t.Errorf("org_id = %q, want org_xyz", got)
	}
	// Malformed / non-JWT inputs degrade to "" rather than panicking.
	for _, bad := range []string{"", "not-a-jwt", "a.b", "h..s"} {
		if got := accessTokenOrgID(bad); got != "" {
			t.Errorf("accessTokenOrgID(%q) = %q, want empty", bad, got)
		}
	}
}

// TestFetchOAuthDiscoveryParsesResponse hits a stub server that mimics
// the real /v1/auth/cli/config payload for the device flow.
func TestFetchOAuthDiscoveryParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/cli/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("discovery must not require auth")
		}
		_, _ = w.Write([]byte(`{"client_id":"client_42","workos_api_base_url":"https://api.workos.com","authkit_domain":"auth.stage.retab.com","scopes":["openid","email"]}`))
	}))
	defer srv.Close()

	disc, err := fetchOAuthDiscovery(context.Background(), srv.URL+"/v1")
	if err != nil {
		t.Fatal(err)
	}
	if disc.ClientID != "client_42" {
		t.Errorf("client_id = %q, want client_42", disc.ClientID)
	}
	if disc.WorkosAPIBaseURL != "https://api.workos.com" {
		t.Errorf("workos_api_base_url = %q, want https://api.workos.com", disc.WorkosAPIBaseURL)
	}
}

// TestFetchOAuthDiscoveryDefaultsWorkosBaseURL — discovery is allowed to omit
// workos_api_base_url; the CLI must default it to https://api.workos.com so the
// device flow has somewhere to run.
func TestFetchOAuthDiscoveryDefaultsWorkosBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"client_id":"client_42"}`))
	}))
	defer srv.Close()

	disc, err := fetchOAuthDiscovery(context.Background(), srv.URL+"/v1")
	if err != nil {
		t.Fatal(err)
	}
	if disc.WorkosAPIBaseURL != defaultWorkosAPIBaseURL {
		t.Errorf("workos_api_base_url = %q, want default %q", disc.WorkosAPIBaseURL, defaultWorkosAPIBaseURL)
	}
}

// TestFetchOAuthDiscoveryAppendsPathWhenBaseURLLacksV1 — `baseURL` may be
// either `https://api.retab.com` or `https://api.retab.com/v1`. The
// function must DTRT for both shapes.
func TestFetchOAuthDiscoveryAppendsPathWhenBaseURLLacksV1(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"client_id":"y","workos_api_base_url":"https://api.workos.com"}`))
	}))
	defer srv.Close()

	_, err := fetchOAuthDiscovery(context.Background(), srv.URL) // no /v1 suffix
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/auth/cli/config" {
		t.Errorf("path: got %q, want /v1/auth/cli/config", gotPath)
	}
}

// newDeviceFlowWorkOS stands up a fake WorkOS User Management server that
// implements the device authorization + authenticate endpoints. The
// authenticate endpoint returns authorization_pending exactly once (so the test
// exercises the poll loop) and then the token bag.
func newDeviceFlowWorkOS(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var authenticateCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/user_management/authorize/device", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("client_id") != "client_test" {
			t.Errorf("device authorize client_id = %q, want client_test", r.PostForm.Get("client_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"device_code": "dc_abc",
			"user_code": "WXYZ-1234",
			"verification_uri": "https://workos.com/device",
			"verification_uri_complete": "https://workos.com/device?code=WXYZ-1234",
			"expires_in": 300,
			"interval": 1
		}`))
	})
	mux.HandleFunc("/user_management/authenticate", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if gt := r.PostForm.Get("grant_type"); gt != deviceCodeGrantType {
			t.Errorf("authenticate grant_type = %q, want %q", gt, deviceCodeGrantType)
		}
		if r.PostForm.Get("device_code") != "dc_abc" {
			t.Errorf("authenticate device_code = %q, want dc_abc", r.PostForm.Get("device_code"))
		}
		w.Header().Set("Content-Type", "application/json")
		if authenticateCalls.Add(1) == 1 {
			// First poll: still waiting for the user.
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"authorization_pending","error_description":"waiting"}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"access_token": "at_xyz",
			"refresh_token": "rt_xyz",
			"token_type": "Bearer",
			"expires_in": 3600,
			"organization_id": "org_login"
		}`))
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	return srv, &authenticateCalls
}

// TestRunLoginFlowDeviceHappyPath exercises the device flow end-to-end:
//   - device authorization returns user/device codes
//   - the poll loop tolerates a single authorization_pending
//   - the token bag is parsed, with org_id + sensible expiry recorded
//   - discovery fields (client_id / workos_api_base_url) are echoed onto tokens
//
// This is the load-bearing test for the entire login UX.
func TestRunLoginFlowDeviceHappyPath(t *testing.T) {
	workosSrv, calls := newDeviceFlowWorkOS(t)
	withTrustingTokenClient(t, workosSrv.Client())

	disc := &cliOAuthDiscovery{
		ClientID:         "client_test",
		WorkosAPIBaseURL: workosSrv.URL,
	}

	var openedURL string
	opener := func(target string) error {
		openedURL = target
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tokens, err := runLoginFlow(ctx, disc, opener)
	if err != nil {
		t.Fatalf("login flow failed: %v", err)
	}

	if tokens.AccessToken != "at_xyz" || tokens.RefreshToken != "rt_xyz" {
		t.Errorf("unexpected tokens: %+v", tokens)
	}
	if tokens.ClientID != disc.ClientID {
		t.Errorf("client_id not echoed onto tokens: got %q", tokens.ClientID)
	}
	if tokens.WorkosAPIBaseURL != disc.WorkosAPIBaseURL {
		t.Errorf("workos_api_base_url not echoed onto tokens: got %q", tokens.WorkosAPIBaseURL)
	}
	if tokens.OrganizationID != "org_login" {
		t.Errorf("organization_id = %q, want org_login", tokens.OrganizationID)
	}
	if time.Until(tokens.ExpiresAt) < 50*time.Minute {
		t.Errorf("expiry too short; expected ~1h, got %v", time.Until(tokens.ExpiresAt))
	}
	// The poll loop must have hit authenticate twice: pending, then success.
	if n := calls.Load(); n != 2 {
		t.Errorf("authenticate calls = %d, want 2 (one pending + one success)", n)
	}
	// The browser was opened to the completed verification URL (embeds the code).
	if !strings.Contains(openedURL, "WXYZ-1234") {
		t.Errorf("opener URL %q should contain the user_code", openedURL)
	}
}

// TestRunLoginFlowSurfacesAccessDenied — when the user denies the request,
// the device poll must fail clearly instead of looping forever.
func TestRunLoginFlowSurfacesAccessDenied(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user_management/authorize/device", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"UC","verification_uri":"https://x","verification_uri_complete":"https://x?c=UC","expires_in":300,"interval":1}`))
	})
	mux.HandleFunc("/user_management/authenticate", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"access_denied","error_description":"user denied"}`))
	})
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	disc := &cliOAuthDiscovery{ClientID: "c", WorkosAPIBaseURL: srv.URL}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := runLoginFlow(ctx, disc, func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error from access_denied response")
	}
	if !strings.Contains(err.Error(), "denied") {
		t.Errorf("error should explain the denial; got %v", err)
	}
}

// TestRunLoginFlowSurfacesExpiredToken — when the device code expires before
// approval, the poll must fail with a clear re-run instruction.
func TestRunLoginFlowSurfacesExpiredToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user_management/authorize/device", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"UC","verification_uri":"https://x","verification_uri_complete":"https://x?c=UC","expires_in":300,"interval":1}`))
	})
	mux.HandleFunc("/user_management/authenticate", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"expired_token","error_description":"code expired"}`))
	})
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	disc := &cliOAuthDiscovery{ClientID: "c", WorkosAPIBaseURL: srv.URL}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := runLoginFlow(ctx, disc, func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error from expired_token response")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("error should explain the expiry; got %v", err)
	}
}

// TestRefreshAccessTokenSwapsTokens — the refresh path posts a refresh_token
// grant to {workos_api_base}/user_management/authenticate (no client_secret, no
// organization_id) and produces a token bag carrying the new access token while
// preserving the recorded client_id/workos_api_base_url for future refreshes.
func TestRefreshAccessTokenSwapsTokens(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user_management/authenticate" {
			t.Errorf("refresh hit %q, want /user_management/authenticate", r.URL.Path)
		}
		_ = r.ParseForm()
		if r.PostForm.Get("grant_type") != "refresh_token" {
			t.Errorf("wrong grant_type: %q", r.PostForm.Get("grant_type"))
		}
		if r.PostForm.Get("refresh_token") != "old_refresh" {
			t.Errorf("wrong refresh_token: %q", r.PostForm.Get("refresh_token"))
		}
		if r.PostForm.Get("client_id") != "c" {
			t.Errorf("wrong client_id: %q", r.PostForm.Get("client_id"))
		}
		// A public client must NOT send a secret or organization_id on refresh.
		if r.PostForm.Has("client_secret") {
			t.Error("refresh must not send client_secret")
		}
		if r.PostForm.Has("organization_id") {
			t.Error("refresh must not send organization_id")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "new_access",
			"refresh_token": "new_refresh",
			"token_type": "Bearer",
			"expires_in": 600
		}`))
	}))
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	tok := &oauthTokens{
		AccessToken:      "old_access",
		RefreshToken:     "old_refresh",
		ClientID:         "c",
		WorkosAPIBaseURL: srv.URL,
		ExpiresAt:        time.Now(),
	}
	got, err := refreshAccessToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got.AccessToken != "new_access" || got.RefreshToken != "new_refresh" {
		t.Errorf("unexpected token bag: %+v", got)
	}
	// Discovery fields survive so subsequent refreshes find their way home.
	if got.ClientID != "c" || got.WorkosAPIBaseURL != srv.URL {
		t.Errorf("discovery fields lost: client=%q base=%q", got.ClientID, got.WorkosAPIBaseURL)
	}
}
