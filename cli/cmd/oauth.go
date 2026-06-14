package cmd

// WorkOS User Management device authorization flow.
//
// Used by `retab auth login` to obtain a WorkOS-issued session token without a
// browser redirect/loopback. Reference: RFC 8628 "OAuth 2.0 Device
// Authorization Grant" and WorkOS User Management device-auth endpoints.
//
// Why device auth (hard cutover from the old loopback + PKCE flow):
//   - Device-flow tokens ARE User Management session tokens. The backend's
//     confidential `/v1/auth/cli/switch-organization` endpoint can exchange the
//     resulting refresh_token (+ organization_id) for a token scoped to a
//     different org, so `retab org switch` is browserless again. The old public
//     OAuth-app tokens could not be exchanged for a different org at all.
//   - No loopback HTTP server, no pre-registered redirect ports, no PKCE
//     challenge round-trip. The CLI prints a short user_code and a URL; the
//     user approves in any browser (works headless / over SSH).
//
// Architecture:
//   1. CLI hits GET /v1/auth/cli/config on the Retab API to discover the
//      WorkOS User Management client_id and workos_api_base_url. No secret.
//   2. CLI POSTs client_id to {base}/user_management/authorize/device and gets
//      back a device_code + user_code + verification URLs + poll interval.
//   3. CLI prints the user_code, opens verification_uri_complete in a browser
//      (best-effort), and also prints the URL for manual / headless use.
//   4. CLI polls {base}/user_management/authenticate with
//      grant_type=urn:ietf:params:oauth:grant-type:device_code until the user
//      approves (or the code expires). On success it gets access_token,
//      refresh_token, expires_in, organization_id.
//   5. CLI persists the tokens (see oauthTokens in config.go).
//
// Keep-alive: refresh runs against the SAME User Management authenticate
// endpoint with grant_type=refresh_token + client_id (no client_secret, no
// organization_id). Public User Management clients refresh without a secret.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// defaultWorkosAPIBaseURL is used when discovery does not pin a
// workos_api_base_url. The WorkOS User Management endpoints all live under it.
const defaultWorkosAPIBaseURL = "https://api.workos.com"

// deviceCodeGrantType is the RFC 8628 grant type for polling the token
// endpoint with a device_code.
const deviceCodeGrantType = "urn:ietf:params:oauth:grant-type:device_code"

// loginTimeout caps how long we wait for the user to approve the device code
// in the browser. The device endpoint also returns its own `expires_in`; we
// stop at whichever fires first. Five minutes is generous for SSO with MFA but
// bounded enough that a forgotten terminal doesn't poll forever.
const loginTimeout = 5 * time.Minute

// refreshLeeway is how far before `expires_at` we proactively refresh. The
// CLI builds a fresh client per command, so the only reason to give
// headroom is in-flight latency between refresh and use.
const refreshLeeway = 30 * time.Second

// defaultDevicePollInterval is the fallback poll cadence when the device
// authorization response omits `interval`.
const defaultDevicePollInterval = 5 * time.Second

// cliOAuthDiscovery is the shape returned by GET /v1/auth/cli/config. The
// device flow only needs ClientID + WorkosAPIBaseURL; AuthKitDomain/Scopes are
// accepted (and ignored) so the discovery payload stays forward-compatible.
type cliOAuthDiscovery struct {
	ClientID         string   `json:"client_id"`
	WorkosAPIBaseURL string   `json:"workos_api_base_url"`
	AuthKitDomain    string   `json:"authkit_domain,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
}

// deviceAuthorizationResponse mirrors the WorkOS
// /user_management/authorize/device response (RFC 8628 §3.2).
type deviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"` // seconds
	Interval                int    `json:"interval"`   // seconds

	// Error fields, present on a non-200.
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// tokenResponse mirrors the WorkOS /user_management/authenticate response for
// both the device-code poll and the refresh_token grant.
type tokenResponse struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token,omitempty"`
	TokenType      string `json:"token_type"`
	ExpiresIn      int    `json:"expires_in"` // seconds
	OrganizationID string `json:"organization_id,omitempty"`

	// Error fields, present on 400/401 (and on device-poll pending states).
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
	if out.ClientID == "" {
		return nil, fmt.Errorf("discovery response missing client_id")
	}
	if out.WorkosAPIBaseURL == "" {
		out.WorkosAPIBaseURL = defaultWorkosAPIBaseURL
	}
	return &out, nil
}

// runLoginFlow drives the full device authorization flow end-to-end and
// returns persisted tokens. The browser is opened with `opener`; tests pass a
// stub that does nothing (the device flow needs no loopback — the test server
// approves the code on its own once the poll arrives).
func runLoginFlow(ctx context.Context, disc *cliOAuthDiscovery, opener func(url string) error) (*oauthTokens, error) {
	base := workosBaseURL(disc.WorkosAPIBaseURL)

	device, err := requestDeviceAuthorization(ctx, base, disc.ClientID)
	if err != nil {
		return nil, err
	}

	// Tell the user what to do. The user_code is the load-bearing artifact —
	// print it prominently. verification_uri_complete embeds the code so the
	// browser pre-fills it; print the bare verification_uri + code too so a
	// headless / SSH user can approve from another machine.
	verifyURL := device.VerificationURIComplete
	if verifyURL == "" {
		verifyURL = device.VerificationURI
	}
	fmt.Fprintf(os.Stderr, "\nTo log in, visit:\n\n  %s\n\nand enter the code:\n\n  %s\n\n", verifyURL, device.UserCode)
	if opener != nil && verifyURL != "" {
		if err := opener(verifyURL); err == nil {
			fmt.Fprintln(os.Stderr, "Opened your browser to complete login. Waiting for approval...")
		} else {
			fmt.Fprintln(os.Stderr, "Open the URL above to complete login. Waiting for approval...")
		}
	} else {
		fmt.Fprintln(os.Stderr, "Waiting for approval...")
	}

	return pollDeviceToken(ctx, base, disc, device)
}

// requestDeviceAuthorization starts the device flow (RFC 8628 §3.1).
func requestDeviceAuthorization(ctx context.Context, base, clientID string) (*deviceAuthorizationResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)

	endpoint := base + "/user_management/authorize/device"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build device authorization request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := tokenHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var dr deviceAuthorizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return nil, fmt.Errorf("decode device authorization response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		if dr.Error != "" {
			return nil, fmt.Errorf("device authorization %s: %s", dr.Error, dr.ErrorDescription)
		}
		return nil, fmt.Errorf("device authorization endpoint returned %d", resp.StatusCode)
	}
	if dr.DeviceCode == "" || dr.UserCode == "" {
		return nil, fmt.Errorf("device authorization response missing device_code/user_code")
	}
	return &dr, nil
}

// pollDeviceToken polls the authenticate endpoint until the user approves the
// device code, the code expires, or the login times out.
func pollDeviceToken(ctx context.Context, base string, disc *cliOAuthDiscovery, device *deviceAuthorizationResponse) (*oauthTokens, error) {
	interval := time.Duration(device.Interval) * time.Second
	if interval <= 0 {
		interval = defaultDevicePollInterval
	}

	// Bound by whichever fires first: the device code's own expiry or our
	// login timeout.
	deadline := time.Now().Add(loginTimeout)
	if device.ExpiresIn > 0 {
		if codeDeadline := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second); codeDeadline.Before(deadline) {
			deadline = codeDeadline
		}
	}
	timeoutCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	form := url.Values{}
	form.Set("grant_type", deviceCodeGrantType)
	form.Set("device_code", device.DeviceCode)
	form.Set("client_id", disc.ClientID)

	endpoint := base + "/user_management/authenticate"
	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("login timed out waiting for approval — re-run `retab auth login`")
		case <-time.After(interval):
		}

		tokens, retryable, slowDown, err := postDeviceAuthenticate(timeoutCtx, endpoint, form, disc)
		if retryable {
			if slowDown {
				// RFC 8628 §3.5: on slow_down, increase the interval by 5s.
				interval += 5 * time.Second
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		return tokens, nil
	}
}

// postDeviceAuthenticate performs one device-code poll. It returns:
//   - tokens (non-nil) on success;
//   - retryable=true when the server says authorization_pending / slow_down
//     (keep polling), with slowDown set for the latter;
//   - a hard error otherwise (access_denied, expired_token, transport, ...).
func postDeviceAuthenticate(ctx context.Context, endpoint string, form url.Values, disc *cliOAuthDiscovery) (tokens *oauthTokens, retryable, slowDown bool, err error) {
	tr, status, perr := postForm(ctx, endpoint, form)
	if perr != nil {
		return nil, false, false, perr
	}
	if status == http.StatusOK && tr.Error == "" {
		toks, terr := tokensFromResponse(tr, disc)
		return toks, false, false, terr
	}
	switch tr.Error {
	case "authorization_pending":
		return nil, true, false, nil
	case "slow_down":
		return nil, true, true, nil
	case "access_denied":
		return nil, false, false, fmt.Errorf("login was denied — re-run `retab auth login` and approve the request")
	case "expired_token":
		return nil, false, false, fmt.Errorf("the login code expired before it was approved — re-run `retab auth login`")
	case "":
		return nil, false, false, fmt.Errorf("authenticate endpoint returned %d", status)
	default:
		return nil, false, false, fmt.Errorf("authenticate endpoint %s: %s", tr.Error, tr.ErrorDescription)
	}
}

// refreshAccessToken trades a refresh_token for a new access_token against the
// WorkOS User Management authenticate endpoint. On success the returned struct
// carries the same WorkosAPIBaseURL/ClientID so the caller can save it back to
// disk without re-doing discovery.
//
// WorkOS rotates refresh tokens by default (each call returns a new
// refresh_token and invalidates the one we just used), so the result MUST be
// persisted promptly — see makeOAuthTokenProvider.
//
// Defensive note: if the server happens to omit `refresh_token` from the
// rotation response, we preserve the caller's existing refresh_token rather
// than wiping it.
func refreshAccessToken(ctx context.Context, tok *oauthTokens) (*oauthTokens, error) {
	if tok == nil || tok.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh_token on file; re-run `retab auth login`")
	}
	if tok.ClientID == "" {
		return nil, fmt.Errorf("config is missing client_id; re-run `retab auth login`")
	}
	base := workosBaseURL(tok.WorkosAPIBaseURL)
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tok.RefreshToken)
	form.Set("client_id", tok.ClientID)

	endpoint := base + "/user_management/authenticate"
	tr, status, err := postForm(ctx, endpoint, form)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || tr.Error != "" {
		if tr.Error == "invalid_grant" {
			return nil, fmt.Errorf("refresh failed: %s — run `retab auth login` to re-authenticate", tr.ErrorDescription)
		}
		if tr.Error != "" {
			return nil, fmt.Errorf("token endpoint %s: %s", tr.Error, tr.ErrorDescription)
		}
		return nil, fmt.Errorf("token endpoint returned %d", status)
	}
	disc := &cliOAuthDiscovery{ClientID: tok.ClientID, WorkosAPIBaseURL: tok.WorkosAPIBaseURL}
	refreshed, err := tokensFromResponse(tr, disc)
	if err != nil {
		return nil, err
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = tok.RefreshToken
	}
	return refreshed, nil
}

// workosBaseURL normalizes a discovered/persisted WorkOS API base URL,
// defaulting to api.workos.com and trimming a trailing slash so endpoint
// concatenation is clean.
func workosBaseURL(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = defaultWorkosAPIBaseURL
	}
	return strings.TrimRight(base, "/")
}

// tokensFromResponse converts a successful authenticate response into the
// persisted oauthTokens shape, stamping the discovery fields so refresh keeps
// working without re-discovery.
func tokensFromResponse(tr *tokenResponse, disc *cliOAuthDiscovery) (*oauthTokens, error) {
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned no access_token")
	}
	ttl := time.Duration(tr.ExpiresIn) * time.Second
	if ttl <= 0 {
		// WorkOS access tokens default to ~5-10 minutes; if the server didn't
		// report a TTL, assume short.
		ttl = 10 * time.Minute
	}
	return &oauthTokens{
		AccessToken:      tr.AccessToken,
		RefreshToken:     tr.RefreshToken,
		TokenType:        tr.TokenType,
		ExpiresAt:        time.Now().Add(ttl),
		ClientID:         disc.ClientID,
		WorkosAPIBaseURL: disc.WorkosAPIBaseURL,
		OrganizationID:   tr.OrganizationID,
	}, nil
}

// tokenHTTPClient is the HTTP client used for WorkOS endpoint POSTs. Tests
// override it to trust the self-signed cert served by httptest.NewTLSServer;
// production never assigns to it.
var tokenHTTPClient = &http.Client{Timeout: 30 * time.Second}

// postForm is the shared form-POST → JSON helper for the device poll and the
// refresh grant. It returns the decoded tokenResponse plus the HTTP status so
// callers can branch on OAuth error codes carried in the JSON body.
func postForm(ctx context.Context, endpoint string, form url.Values) (*tokenResponse, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := tokenHTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("POST %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode token response (status %d): %w", resp.StatusCode, err)
	}
	return &tr, resp.StatusCode, nil
}

// accessTokenOrgID reads the `org_id` claim from a WorkOS access token WITHOUT
// verifying the signature — WorkOS just minted it and the backend re-verifies on
// every request; here we only need to confirm which organization the session
// landed in after an org switch. Any decode failure returns "" (the caller then
// trusts the requested target rather than failing a working login).
func accessTokenOrgID(accessToken string) string {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		OrgID string `json:"org_id"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.OrgID
}

// browserOpener is the function flows use to open the user's browser. It is a
// package var (defaulting to openBrowser) so tests can stub the browser step of
// flows driven through a cobra RunE without shelling out.
var browserOpener = openBrowser

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
