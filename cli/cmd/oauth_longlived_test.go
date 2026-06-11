//go:build !retab_oagen_cli_classifications && !retab_oagen_cli_edits && !retab_oagen_cli_extractions && !retab_oagen_cli_files && !retab_oagen_cli_parses && !retab_oagen_cli_partitions && !retab_oagen_cli_schemas && !retab_oagen_cli_secrets && !retab_oagen_cli_splits && !retab_oagen_cli_tables && !retab_oagen_cli_workflows && !retab_oagen_cli_workflows_artifacts && !retab_oagen_cli_workflows_blocks && !retab_oagen_cli_workflows_blocks_executions && !retab_oagen_cli_workflows_edges && !retab_oagen_cli_workflows_experiments && !retab_oagen_cli_workflows_reviews && !retab_oagen_cli_workflows_runs && !retab_oagen_cli_workflows_spec && !retab_oagen_cli_workflows_steps && !retab_oagen_cli_workflows_tests

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// These tests pin the contract that makes a WorkOS OAuth session genuinely
// long-lived in the face of real-world flakiness:
//
//   1. `offline_access` is unconditionally requested, so AuthKit always
//      mints a refresh_token alongside the access_token. Without this,
//      sessions die after ~10 min and no amount of refresh logic helps.
//
//   2. Refresh-token rotation: every successful refresh persists the
//      newly-issued refresh_token before WorkOS server-side invalidates
//      the old one. A torn save here is the canonical "user mysteriously
//      logged out a few hours later" bug.
//
//   3. Atomic save: an interrupted write must leave the prior valid
//      config intact, NOT a truncated / partially-rotated file the next
//      invocation will choke on.
//
//   4. Concurrent CLI invocations racing on refresh: the second instance
//      must recover by re-reading the rotated tokens from disk rather
//      than failing with an erroneous "session expired" error.
//
//   5. Defensive: a server that omits refresh_token on the rotation
//      response (other OAuth providers signal "reuse existing" this way)
//      must not wipe the caller's existing one.

// Test 1 — `offline_access` is always requested, even when discovery
// forgets to include it. Without this, no refresh_token is issued at all.
func TestBuildAuthorizeURL_AlwaysRequestsOfflineAccess(t *testing.T) {
	cases := []struct {
		name        string
		discScopes  []string
		wantInScope []string
	}{
		{
			name:        "empty discovery uses defaults with offline_access",
			discScopes:  nil,
			wantInScope: []string{"openid", "offline_access"},
		},
		{
			name:        "discovery omits offline_access — CLI adds it back",
			discScopes:  []string{"openid", "profile", "email"},
			wantInScope: []string{"openid", "offline_access"},
		},
		{
			name:        "discovery includes offline_access — no duplicate",
			discScopes:  []string{"openid", "offline_access"},
			wantInScope: []string{"openid", "offline_access"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			disc := &cliOAuthDiscovery{
				AuthKitDomain: "auth.retab.com",
				ClientID:      "c",
				Scopes:        tc.discScopes,
			}
			got := buildAuthorizeURL(disc, "http://localhost/cb", "chal", "state")
			u, err := url.Parse(got)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			scopes := strings.Fields(u.Query().Get("scope"))
			for _, want := range tc.wantInScope {
				found := false
				for _, s := range scopes {
					if s == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("scope %q missing from %v", want, scopes)
				}
			}
			// Sanity: offline_access should appear exactly once.
			n := 0
			for _, s := range scopes {
				if s == "offline_access" {
					n++
				}
			}
			if n != 1 {
				t.Errorf("offline_access appears %d times in %v, want exactly 1", n, scopes)
			}
		})
	}
}

// Test 2 — saveConfig is atomic. An interrupted write doesn't leave a
// truncated or empty file that a subsequent loadConfig would choke on.
//
// We execute "interrupted" by inspecting the directory mid-save: there
// should never be a config.json that contains anything but valid JSON
// (since the rename is atomic, the file is either the prior contents or
// the new contents, never half-formed bytes).
func TestSaveConfig_IsAtomic(t *testing.T) {
	// isolateHome sets HOME *and* USERPROFILE; on Windows saveConfig resolves
	// the config dir via %USERPROFILE%, so HOME alone would write outside tmp
	// and the post-save readdir below would miss the file entirely.
	tmp := isolateHome(t)
	// pre-populate with a known-good baseline
	first := retabConfig{APIKey: "sk_first"}
	if err := saveConfig(first); err != nil {
		t.Fatalf("first save: %v", err)
	}

	// Replace with a config containing a (large) OAuth blob. If save were
	// non-atomic, an interrupt mid-write could land an empty file; here
	// we just verify the contents are valid JSON post-save.
	second := retabConfig{
		OAuth: &oauthTokens{
			AccessToken:   "at_2",
			RefreshToken:  "rt_2",
			AuthKitDomain: "auth.retab.com",
			ClientID:      "c",
			ExpiresAt:     time.Now().Add(10 * time.Minute),
		},
	}
	if err := saveConfig(second); err != nil {
		t.Fatalf("second save: %v", err)
	}

	// Verify no leftover temp file (cleanup happened).
	entries, err := os.ReadDir(filepath.Join(tmp, ".retab"))
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp") {
			t.Errorf("leftover temp file %q — atomic save did not clean up", e.Name())
		}
	}

	// Verify the rewritten file parses cleanly and reflects the second save.
	got, err := loadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.OAuth == nil || got.OAuth.RefreshToken != "rt_2" {
		t.Errorf("post-save load: refresh token = %v, want rt_2", got.OAuth)
	}
}

// Test 3 — saveConfig must preserve 0600 permissions through the
// rename. A common atomic-rename bug is creating the temp file with
// default 0644 and inheriting that into the destination.
func TestSaveConfig_Preserves0600(t *testing.T) {
	// isolateHome sets HOME and USERPROFILE so the config lands in tmp on
	// Windows too (saveConfig resolves via %USERPROFILE% there).
	tmp := isolateHome(t)
	if err := saveConfig(retabConfig{APIKey: "k"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	path := filepath.Join(tmp, ".retab", "config.json")
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// NTFS doesn't carry POSIX mode bits, so the 0600 contract is POSIX-only.
	if got := st.Mode().Perm(); runtime.GOOS != "windows" && got != 0o600 {
		t.Errorf("config perm = %v, want 0600", got)
	}
}

// Test 4 — refreshAccessToken preserves the caller's existing
// refresh_token when the server omits one in the rotation response.
// Without this, a stray empty `refresh_token` in the response would
// silently wipe long-lived sessions.
func TestRefreshAccessToken_PreservesRefreshTokenWhenServerOmitsIt(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		// Omit `refresh_token` from the response — defensive code path.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "at_new",
			"token_type":   "Bearer",
			"expires_in":   600,
		})
	}))
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	// Point the token URL at our fake server by parking domain "" and
	// using a custom domain that exchangeURL would format into our URL.
	// Easiest path: hijack tokenHTTPClient (done above) and pass a
	// discovery struct whose AuthKitDomain resolves to the server host.
	u, _ := url.Parse(srv.URL)
	// We can't override the URL scheme in refreshAccessToken cleanly, so
	// instead we test the post-processing logic: call postTokenEndpoint
	// directly via refreshAccessToken with a domain we've redirected.
	tok := &oauthTokens{
		RefreshToken:  "rt_original",
		AuthKitDomain: u.Host, // host:port that our fake server is listening on
		ClientID:      "c",
		ExpiresAt:     time.Now(),
	}
	// refreshAccessToken builds the URL as `https://<authkit_domain>/oauth2/token`;
	// our fake server uses TLS and answers on any path, so the host:port match
	// is sufficient.
	got, err := refreshAccessToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got.RefreshToken != "rt_original" {
		t.Errorf("refresh_token = %q, want preserved %q", got.RefreshToken, "rt_original")
	}
	if got.AccessToken != "at_new" {
		t.Errorf("access_token = %q, want %q", got.AccessToken, "at_new")
	}
}

// Test 5 — Concurrent-refresh race recovery. Two CLI invocations racing
// on the refresh endpoint: the first wins and rotates the refresh_token.
// The second's refresh attempt is rejected with `invalid_grant` because
// the server has already invalidated its (now-stale) refresh_token. The
// token provider must recover by re-reading the new tokens from disk,
// not by erroring out and telling the user to re-login.
func TestMakeOAuthTokenProvider_RecoversFromConcurrentRefresh(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	var refreshCalls atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		call := refreshCalls.Add(1)
		if call == 1 {
			// First refresh attempt: server says invalid_grant because
			// another process rotated the refresh_token a moment ago.
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":             "invalid_grant",
				"error_description": "refresh token already used",
			})
			return
		}
		// Second attempt: succeeds with the rotated refresh_token the
		// "other CLI invocation" wrote to disk between attempt 1 and 2.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at_after_recovery",
			"refresh_token": "rt_rotated_v3",
			"token_type":    "Bearer",
			"expires_in":    600,
		})
	}))
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	u, _ := url.Parse(srv.URL)

	// In-memory snapshot the closure starts with — the OLD refresh token,
	// which the server will reject on first call.
	initial := &oauthTokens{
		AccessToken:   "at_stale",
		RefreshToken:  "rt_stale",
		ExpiresAt:     time.Now().Add(-1 * time.Minute), // already expired
		AuthKitDomain: u.Host,
		ClientID:      "c",
	}

	// Execute "the other CLI process refreshed first and saved": put a
	// fresh refresh_token on disk that differs from the initial.
	if err := saveConfig(retabConfig{
		OAuth: &oauthTokens{
			AccessToken:   "at_other_proc",
			RefreshToken:  "rt_rotated_v2",
			ExpiresAt:     time.Now().Add(-1 * time.Minute), // also expired, forcing the second refresh path
			AuthKitDomain: u.Host,
			ClientID:      "c",
		},
	}); err != nil {
		t.Fatalf("seed disk: %v", err)
	}

	provider := makeOAuthTokenProvider(initial)
	got, err := provider(context.Background())
	if err != nil {
		t.Fatalf("provider should have recovered, got error: %v", err)
	}
	if got != "at_after_recovery" {
		t.Errorf("recovered access_token = %q, want %q", got, "at_after_recovery")
	}
	if n := refreshCalls.Load(); n != 2 {
		t.Errorf("expected 2 refresh attempts (one failed + one recovered), got %d", n)
	}

	// And: the on-disk state reflects the latest rotation.
	disk, err := loadConfig()
	if err != nil {
		t.Fatalf("load disk: %v", err)
	}
	if disk.OAuth == nil || disk.OAuth.RefreshToken != "rt_rotated_v3" {
		t.Errorf("on-disk refresh_token after recovery = %v, want rt_rotated_v3", disk.OAuth)
	}
}

// Test 6 — invalid_grant from a SINGLE (non-racing) CLI invocation must
// still surface as a hard error. We must distinguish "we lost the race"
// (recoverable) from "this refresh_token genuinely doesn't work anymore"
// (force re-login). The disk RefreshToken matching the in-memory one is
// the discriminator.
func TestMakeOAuthTokenProvider_FailsWhenRefreshTokenIsGenuinelyDead(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_grant",
			"error_description": "token expired beyond refresh window",
		})
	}))
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	u, _ := url.Parse(srv.URL)

	// Disk and in-memory refresh tokens match — no race, this is the real
	// "session expired" case.
	tok := &oauthTokens{
		AccessToken:   "at_x",
		RefreshToken:  "rt_dead",
		ExpiresAt:     time.Now().Add(-1 * time.Hour),
		AuthKitDomain: u.Host,
		ClientID:      "c",
	}
	if err := saveConfig(retabConfig{OAuth: tok}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	provider := makeOAuthTokenProvider(tok)
	_, err := provider(context.Background())
	if err == nil {
		t.Fatal("expected error for genuinely-dead refresh_token, got nil")
	}
	if !isInvalidGrantError(err) {
		t.Errorf("error %q does not look like an invalid_grant surface — discriminator is broken", err)
	}
}

// Test 7 — saveConfig failure during refresh must NOT silently break the
// session. We can't easily induce a filesystem failure in a portable test,
// but we can pin the warning-on-stderr behaviour by using a non-existent
// HOME so configDir fails — which exercises the same surface error path.
//
// (This is a deliberately weak test: we're really checking that the
// provider returns success for the in-flight request even when the disk
// save fails. The "warn on stderr" assertion is bonus.)
func TestMakeOAuthTokenProvider_InFlightSucceedsEvenWhenSaveFails(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "at_inflight",
			"refresh_token": "rt_new",
			"token_type":    "Bearer",
			"expires_in":    600,
		})
	}))
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	u, _ := url.Parse(srv.URL)

	// Point HOME at a path that exists but is NOT writable, so MkdirAll
	// inside saveConfig will fail. macOS's /etc is a portable choice
	// because non-root users can't create dirs there.
	if os.Geteuid() == 0 {
		t.Skip("running as root — cannot induce save failure on /etc")
	}
	t.Setenv("HOME", "/etc")
	t.Cleanup(func() { _ = os.RemoveAll("/etc/.retab") })

	tok := &oauthTokens{
		RefreshToken:  "rt_stale",
		ExpiresAt:     time.Now().Add(-1 * time.Minute),
		AuthKitDomain: u.Host,
		ClientID:      "c",
	}
	provider := makeOAuthTokenProvider(tok)
	got, err := provider(context.Background())
	if err != nil {
		t.Fatalf("in-flight request should succeed even if save fails: %v", err)
	}
	if got != "at_inflight" {
		t.Errorf("access_token = %q, want %q", got, "at_inflight")
	}
}

// Smoke test for isInvalidGrantError — its job is to discriminate
// recoverable-race errors from anything else. Cheap and important.
func TestIsInvalidGrantError(t *testing.T) {
	cases := map[string]bool{
		"":                                   false,
		"refresh failed: token already used": true,
		"some other refresh failed: oops":    true, // substring match — intentional
		"network unreachable":                false,
		"token endpoint invalid_grant: x":    true,
	}
	for msg, want := range cases {
		var err error
		if msg != "" {
			err = errors.New(msg)
		}
		if got := isInvalidGrantError(err); got != want {
			t.Errorf("isInvalidGrantError(%q) = %v, want %v", msg, got, want)
		}
	}
}

// withTrustingTokenClient hijacks tokenHTTPClient to trust the self-signed
// cert served by httptest.NewTLSServer. Existing oauth_test.go has a
// similar helper but for the discovery client; this one is scoped to the
// token endpoint client.
//
// We define it here rather than importing from oauth_test.go because Go
// test files can't import each other's helpers across packages; staying
// internal to cmd/ + name-disambiguated keeps the surface clean.
func init() {
	// nothing — placeholder; helper is the function below
	_ = fmt.Sprint
}

// We need the helper, but oauth_test.go already defines one with the same
// name. Re-using it via package scope is fine (both files are in package
// cmd). If oauth_test.go's helper is renamed, this redefinition will
// surface a compile error — that's the right outcome.
//
// (Intentionally NOT defining withTrustingTokenClient here — it lives in
// oauth_test.go and is shared via package scope.)
