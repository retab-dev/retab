package cmd

// OAuth 2.0 Authorization Code + PKCE flow with loopback redirect.
//
// Used by `retab auth login` to obtain a WorkOS-issued access token without
// requiring the user to paste anything. Reference: RFC 8252 §4 "Authorization
// Flow for Native Apps" and RFC 7636 "PKCE".
//
// Architecture:
//   1. CLI hits GET /v1/auth/cli/config on the Retab API to discover the
//      WorkOS authkit_domain + the CLI's client_id. No secret involved.
//   2. CLI generates a fresh PKCE code_verifier (43+ chars high-entropy
//      random) and its S256-derived code_challenge.
//   3. CLI binds a TCP listener on 127.0.0.1, taking the first free port
//      from a fixed set (cliRedirectPorts) — WorkOS matches redirect_uri
//      against an exact pre-registered list — and constructs
//      https://{authkit_domain}/oauth2/authorize?... with
//      redirect_uri=http://127.0.0.1:<port>/callback.
//   4. CLI opens the user's default browser to that URL. User logs in
//      through WorkOS AuthKit (which transparently dispatches to SSO,
//      Google, email/password, magic link, etc.).
//   5. WorkOS redirects back to the loopback with ?code=...&state=... .
//      The callback handler verifies state, captures code, returns a
//      friendly success page, and signals the main goroutine.
//   6. CLI POSTs code + code_verifier to https://{authkit_domain}/oauth2/token
//      and persists the resulting access_token + refresh_token + expiry.
//
// Why the WorkOS endpoint instead of a Retab proxy:
//   - One fewer endpoint to operate.
//   - The backend's identity resolver already accepts WorkOS JWTs natively
//     (see services/internal/workos/functions.py::resolve_identity).
//
// If we ever want to wrap tokens in longer-lived Retab session JWTs we can
// introduce a `/v1/auth/cli/token` proxy and switch this code over without
// breaking the config schema — `AuthKitDomain` is recorded in oauthTokens
// so existing installs keep refreshing against WorkOS.

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// cliConfigPath is the well-known Retab discovery path. Kept as a constant
// so tests can override it via a stub server.
const cliConfigPath = "/v1/auth/cli/config"

// loginTimeout caps how long we wait for the user to complete the browser
// flow. Five minutes is generous for SSO with MFA but bounded enough that
// a forgotten terminal doesn't hold a listener forever.
const loginTimeout = 5 * time.Minute

// refreshLeeway is how far before `expires_at` we proactively refresh. The
// CLI builds a fresh client per command, so the only reason to give
// headroom is in-flight latency between refresh and use.
const refreshLeeway = 30 * time.Second

// cliRedirectPorts is the fixed set of loopback ports the CLI tries for
// the OAuth callback. WorkOS validates redirect_uri against an exact
// pre-registered list, and its OAuth-application editor rejects the
// RFC 8252 port wildcard, so the CLI cannot use a kernel-picked port.
// Every port here MUST be registered on the WorkOS CLI application as
// `http://127.0.0.1:<port>/callback`. The CLI binds the first free one;
// the rest are fallbacks for when an earlier login left a listener around
// or another process holds the port.
var cliRedirectPorts = []int{42817, 42818, 42819, 42820, 42821, 42822, 42823}

// bindLoopbackListener binds 127.0.0.1 on the first available port from
// cliRedirectPorts and returns the listener plus the chosen port. It
// errors only when every registered port is occupied.
func bindLoopbackListener() (net.Listener, int, error) {
	for _, port := range cliRedirectPorts {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return listener, port, nil
		}
	}
	return nil, 0, fmt.Errorf(
		"all loopback callback ports are in use (%v); close any other `retab auth login` still running and retry",
		cliRedirectPorts,
	)
}

// cliOAuthDiscovery is the shape returned by GET /v1/auth/cli/config.
type cliOAuthDiscovery struct {
	AuthKitDomain string   `json:"authkit_domain"`
	ClientID      string   `json:"client_id"`
	Scopes        []string `json:"scopes"`
}

// tokenResponse mirrors the WorkOS /oauth2/token response.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	Scope        string `json:"scope,omitempty"`

	// Error fields, present on 400/401.
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// fetchOAuthDiscovery hits the Retab API for the CLI OAuth config. Public
// endpoint — no auth required.
func fetchOAuthDiscovery(ctx context.Context, baseURL string) (*cliOAuthDiscovery, error) {
	if baseURL == "" {
		baseURL = "https://api.retab.com/v1"
	}
	// The discovery endpoint is exposed at /v1/auth/cli/config but baseURL
	// already typically ends in /v1. Be defensive about both.
	endpoint := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(endpoint, "/v1") {
		endpoint += "/auth/cli/config"
	} else {
		endpoint += cliConfigPath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build discovery request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery returned %d from %s — is the deployment configured for CLI OAuth?", resp.StatusCode, endpoint)
	}
	var out cliOAuthDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode discovery: %w", err)
	}
	if out.AuthKitDomain == "" || out.ClientID == "" {
		return nil, fmt.Errorf("discovery response missing authkit_domain or client_id")
	}
	return &out, nil
}

// generatePKCE produces a (verifier, challenge) pair per RFC 7636. The
// verifier is 32 raw random bytes encoded with base64url (43 chars). The
// challenge is the base64url-encoded SHA-256 of the verifier.
func generatePKCE() (verifier, challenge string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("read random: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// generateState returns a 32-byte base64url-encoded random string for CSRF
// protection on the authorize -> callback round trip.
func generateState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// callbackResult is the outcome of the loopback redirect: either a code
// (success) or an OAuth error (user denied, WorkOS misconfig, etc.).
type callbackResult struct {
	code  string
	state string
	err   error
}

// runLoginFlow drives the full PKCE flow end-to-end and returns persisted
// tokens. The browser is opened with `opener`; tests pass a stub that hits
// the callback URL itself instead of shelling out.
func runLoginFlow(ctx context.Context, disc *cliOAuthDiscovery, opener func(url string) error) (*oauthTokens, error) {
	verifier, challenge, err := generatePKCE()
	if err != nil {
		return nil, err
	}
	state, err := generateState()
	if err != nil {
		return nil, err
	}

	// Bind 127.0.0.1 on one of the pre-registered callback ports. RFC 8252
	// §7.3 blesses loopback for native apps, but WorkOS matches redirect_uri
	// against an exact pre-registered list — its OAuth-application editor
	// rejects the port wildcard — so we cannot let the kernel pick a port.
	listener, port, err := bindLoopbackListener()
	if err != nil {
		return nil, err
	}
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	resultCh := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		gotState := q.Get("state")
		if errCode := q.Get("error"); errCode != "" {
			errDesc := q.Get("error_description")
			renderHTML(w, false, errDesc)
			select {
			case resultCh <- callbackResult{err: fmt.Errorf("WorkOS returned %s: %s", errCode, errDesc)}:
			default:
			}
			return
		}
		if gotState != state {
			renderHTML(w, false, "Login flow could not be verified. Please try again.")
			select {
			case resultCh <- callbackResult{err: fmt.Errorf("state mismatch: expected %q, got %q (possible CSRF)", state, gotState)}:
			default:
			}
			return
		}
		code := q.Get("code")
		if code == "" {
			renderHTML(w, false, "Missing authorization code.")
			select {
			case resultCh <- callbackResult{err: fmt.Errorf("callback missing 'code' query parameter")}:
			default:
			}
			return
		}
		renderHTML(w, true, "")
		select {
		case resultCh <- callbackResult{code: code, state: gotState}:
		default:
		}
	})

	// Catch-all so users who hit the loopback root get a friendly page
	// instead of 404.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/callback?error=manual_visit&error_description=Open+the+URL+from+your+terminal+to+complete+login.", http.StatusFound)
	})

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	authorizeURL := buildAuthorizeURL(disc, redirectURI, challenge, state)
	if err := opener(authorizeURL); err != nil {
		// Don't fail — print the URL so the user can open it manually.
		fmt.Fprintf(os.Stderr, "Could not open browser automatically.\nVisit this URL to log in:\n\n  %s\n\n", authorizeURL)
	} else {
		fmt.Fprintf(os.Stderr, "Opened browser to complete login.\nIf nothing happened, visit:\n\n  %s\n\n", authorizeURL)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, loginTimeout)
	defer cancel()

	var res callbackResult
	select {
	case res = <-resultCh:
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("login timed out after %s — re-run `retab auth login`", loginTimeout)
	}
	if res.err != nil {
		return nil, res.err
	}

	tokens, err := exchangeAuthorizationCode(timeoutCtx, disc, res.code, verifier, redirectURI)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// buildAuthorizeURL constructs the WorkOS authorize URL. Pure function so
// it's trivially testable.
func buildAuthorizeURL(disc *cliOAuthDiscovery, redirectURI, codeChallenge, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", disc.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	scopes := disc.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email", "offline_access"}
	}
	// Belt-and-suspenders: `offline_access` MUST be in the scope list or
	// WorkOS won't issue a refresh_token, and the CLI silently degrades to
	// ~10 min sessions. If the deployment's discovery response forgets to
	// include it, we add it here. Idempotent in the common case.
	scopes = ensureOfflineAccess(scopes)
	q.Set("scope", strings.Join(scopes, " "))
	return fmt.Sprintf("https://%s/oauth2/authorize?%s", disc.AuthKitDomain, q.Encode())
}

// ensureOfflineAccess appends `offline_access` to scopes if absent. The
// scope is what tells WorkOS to mint a refresh_token alongside the access
// token; without it the CLI cannot keep the session alive past ~10 min.
func ensureOfflineAccess(scopes []string) []string {
	for _, s := range scopes {
		if s == "offline_access" {
			return scopes
		}
	}
	return append(scopes, "offline_access")
}

// exchangeAuthorizationCode redeems the auth code for tokens. The body is
// form-urlencoded per RFC 6749 §4.1.3.
func exchangeAuthorizationCode(ctx context.Context, disc *cliOAuthDiscovery, code, verifier, redirectURI string) (*oauthTokens, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", disc.ClientID)
	form.Set("code_verifier", verifier)
	form.Set("redirect_uri", redirectURI)

	tokenURL := fmt.Sprintf("https://%s/oauth2/token", disc.AuthKitDomain)
	return postTokenEndpoint(ctx, tokenURL, form, disc)
}

// refreshAccessToken trades a refresh_token for a new access_token. On
// success the returned struct carries the same AuthKitDomain/ClientID so
// the caller can save it back to disk without re-doing discovery.
//
// WorkOS AuthKit rotates refresh tokens by default (each call returns a
// new refresh_token and invalidates the one we just used), so the result
// MUST be persisted promptly — see makeOAuthTokenProvider.
//
// Defensive note: if the server happens to omit `refresh_token` from the
// rotation response (other OAuth providers signal "keep using the same
// one" this way), we preserve the caller's existing refresh_token rather
// than wiping it. WorkOS in practice always returns a new one, but the
// fallback costs nothing and protects against future regressions.
func refreshAccessToken(ctx context.Context, tok *oauthTokens) (*oauthTokens, error) {
	if tok == nil || tok.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh_token on file; re-run `retab auth login`")
	}
	if tok.AuthKitDomain == "" || tok.ClientID == "" {
		return nil, fmt.Errorf("config is missing authkit_domain/client_id; re-run `retab auth login`")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tok.RefreshToken)
	form.Set("client_id", tok.ClientID)
	tokenURL := fmt.Sprintf("https://%s/oauth2/token", tok.AuthKitDomain)
	disc := &cliOAuthDiscovery{AuthKitDomain: tok.AuthKitDomain, ClientID: tok.ClientID}
	refreshed, err := postTokenEndpoint(ctx, tokenURL, form, disc)
	if err != nil {
		return nil, err
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = tok.RefreshToken
	}
	return refreshed, nil
}

// tokenHTTPClient is the HTTP client used for OAuth token endpoint POSTs.
// Tests override it to trust the self-signed cert served by
// httptest.NewTLSServer; production never assigns to it.
var tokenHTTPClient = &http.Client{Timeout: 30 * time.Second}

// postTokenEndpoint is the shared POST-form helper. It maps server-side
// OAuth errors to typed Go errors so the caller can decide whether to
// prompt re-login (`invalid_grant`) or surface the message as-is.
func postTokenEndpoint(ctx context.Context, tokenURL string, form url.Values, disc *cliOAuthDiscovery) (*oauthTokens, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := tokenHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", tokenURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode token response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		if tr.Error == "invalid_grant" {
			return nil, fmt.Errorf("refresh failed: %s — run `retab auth login` to re-authenticate", tr.ErrorDescription)
		}
		if tr.Error != "" {
			return nil, fmt.Errorf("token endpoint %s: %s", tr.Error, tr.ErrorDescription)
		}
		return nil, fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned no access_token")
	}
	ttl := time.Duration(tr.ExpiresIn) * time.Second
	if ttl <= 0 {
		// WorkOS access tokens default to 10 minutes; if the server didn't
		// report a TTL, assume short.
		ttl = 10 * time.Minute
	}
	return &oauthTokens{
		AccessToken:   tr.AccessToken,
		RefreshToken:  tr.RefreshToken,
		TokenType:     tr.TokenType,
		ExpiresAt:     time.Now().Add(ttl),
		Scope:         tr.Scope,
		AuthKitDomain: disc.AuthKitDomain,
		ClientID:      disc.ClientID,
	}, nil
}

// openBrowser launches the user's default browser. Best-effort: on
// platforms where we can't determine the right command, callers should
// fall back to printing the URL.
func openBrowser(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "linux":
		cmd = exec.Command("xdg-open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		return fmt.Errorf("don't know how to open browser on %s", runtime.GOOS)
	}
	return cmd.Start()
}

// renderHTML writes the page the user sees after the OAuth redirect. We
// keep it intentionally minimal — most users will close the tab seconds
// after it appears.
func renderHTML(w http.ResponseWriter, success bool, detail string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if success {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successHTML))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		// Don't reflect arbitrary server text into HTML without escaping.
		// `detail` originates from query params, so go through the
		// template's escaping by way of a simple substitution.
		page := strings.ReplaceAll(failureHTMLTemplate, "{{DETAIL}}", htmlEscape(detail))
		_, _ = w.Write([]byte(page))
	}
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return r.Replace(s)
}

const successHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>Retab CLI — Logged in</title>
<style>
body { font: 16px/1.5 -apple-system, system-ui, sans-serif; padding: 4rem 2rem; text-align: center; color: #222; }
h1 { color: #D02572; margin-bottom: 0.5rem; }
.box { max-width: 28rem; margin: 0 auto; padding: 1.5rem; border: 1px solid #eee; border-radius: 8px; }
.tick { font-size: 2.5rem; }
</style></head>
<body><div class="box"><div class="tick">✓</div>
<h1>You're logged in</h1>
<p>You can close this tab and return to the terminal.</p>
</div></body></html>`

const failureHTMLTemplate = `<!doctype html>
<html><head><meta charset="utf-8"><title>Retab CLI — Login failed</title>
<style>
body { font: 16px/1.5 -apple-system, system-ui, sans-serif; padding: 4rem 2rem; text-align: center; color: #222; }
h1 { color: #c0392b; margin-bottom: 0.5rem; }
.box { max-width: 28rem; margin: 0 auto; padding: 1.5rem; border: 1px solid #f3d4d4; border-radius: 8px; background: #fdf7f7; }
.detail { color: #666; font-size: 0.9rem; margin-top: 0.75rem; }
</style></head>
<body><div class="box">
<h1>Login failed</h1>
<p>Return to your terminal — it will show the error.</p>
<p class="detail">{{DETAIL}}</p>
</div></body></html>`
