package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// listAllWorkflowBlocks returns every block for a workflow, walking all pages
// via AutoPaging. Block-resolution callers (start/alias/target-handle
// resolution) must see the full set: Blocks.List returns a single page and the
// server's default ordering / page size are not guaranteed, so a block on a
// later page would otherwise be invisible — producing wrong alias resolution,
// false "no/ambiguous start_document block" errors, or unresolved handles. Same
// failure mode as the review-version resolver, which already auto-pages.
func listAllWorkflowBlocks(ctx context.Context, client *retab.Client, workflowID string) ([]retab.WorkflowBlock, error) {
	page, err := client.Workflows.Blocks.List(ctx, &retab.WorkflowBlocksListParams{WorkflowID: workflowID})
	if err != nil {
		return nil, err
	}
	var blocks []retab.WorkflowBlock
	if err := page.AutoPaging(ctx, func(b retab.WorkflowBlock) error {
		blocks = append(blocks, b)
		return nil
	}); err != nil {
		return nil, err
	}
	return blocks, nil
}

// runE wraps a command body so APIErrors render as concise user-facing
// messages by default and as full HTTP diagnostics with --debug. Other errors
// render as a single line. Returned sentinel errors keep the process exit
// non-zero without asking cobra to print the same message a second time.
func runE(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := fn(cmd, args)
		if err == nil {
			return nil
		}
		var apiErr *retab.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintln(os.Stderr, renderAPIErrorForCLI(cmd, apiErr))
			return errSilent
		}
		if msg := renderConnectionDropForCLI(err); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
			return renderedError{err: err}
		}
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		return renderedError{err: err}
	}
}

// renderConnectionDropForCLI replaces bare “net/http“ errors of the
// shape “Post "...": EOF“ / “unexpected EOF“ with an actionable
// message. These appear when the upstream server closes the connection
// mid-response (e.g. an uvicorn worker crashed handling the request),
// and the raw Go error tells the user nothing about what happened or
// what to do next.
//
// Returns an empty string when the error is not a connection drop —
// the caller falls back to the default “error: <err>“ render.
func renderConnectionDropForCLI(err error) string {
	if err == nil {
		return ""
	}
	if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		// Some Go HTTP transports wrap EOF in a non-errors.Is-comparable
		// wrapper. Fall back to a substring check on the rendered text so
		// the heuristic still catches `Post "...": EOF` from net/http.
		text := err.Error()
		if !strings.HasSuffix(text, ": EOF") && !strings.Contains(text, ": unexpected EOF") {
			return ""
		}
	}
	return "error: upstream server closed the connection unexpectedly — the request likely crashed server-side. Retry, and if it persists check the server logs."
}

func renderAPIErrorForCLI(cmd *cobra.Command, apiErr *retab.APIError) string {
	if debug, _ := cmd.Root().PersistentFlags().GetBool("debug"); debug {
		return apiErr.String()
	}
	if validationLines := formatValidationErrorLines(apiErr.Message); len(validationLines) > 0 {
		lines := []string{fmt.Sprintf("%d — Invalid request.", apiErr.StatusCode)}
		lines = append(lines, validationLines...)
		if apiErr.RequestID != "" {
			lines = append(lines, "  Request-ID: "+apiErr.RequestID)
		}
		return strings.Join(lines, "\n")
	}
	message := apiErr.Message
	if detail := usefulAPIErrorDetail(apiErr.Details); detail != "" && isGenericAPIErrorMessage(message) {
		message = detail
	}
	if strings.TrimSpace(message) == "" {
		message = fmt.Sprintf("Request failed (%d)", apiErr.StatusCode)
	}
	lines := []string{fmt.Sprintf("%d — %s", apiErr.StatusCode, message)}
	if apiErr.RequestID != "" {
		lines = append(lines, "  Request-ID: "+apiErr.RequestID)
	}
	return strings.Join(lines, "\n")
}

func usefulAPIErrorDetail(details map[string]any) string {
	raw, ok := details["error"]
	if !ok {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case map[string]any:
		if message, ok := value["message"].(string); ok {
			return strings.TrimSpace(message)
		}
	}
	return ""
}

type validationErrorLine struct {
	Location []any  `json:"loc"`
	Message  string `json:"msg"`
}

func formatValidationErrorLines(raw string) []string {
	var errors []validationErrorLine
	if err := json.Unmarshal([]byte(raw), &errors); err != nil || len(errors) == 0 {
		return nil
	}

	lines := make([]string, 0, len(errors))
	for _, err := range errors {
		location := formatValidationLocation(err.Location)
		message := strings.TrimSpace(err.Message)
		if message == "" {
			message = "Invalid value"
		}
		if location == "" {
			lines = append(lines, "  - "+message)
			continue
		}
		lines = append(lines, fmt.Sprintf("  - %s: %s", location, message))
	}
	return lines
}

func formatValidationLocation(location []any) string {
	parts := make([]string, 0, len(location))
	for _, value := range location {
		switch typed := value.(type) {
		case string:
			if typed == "body" || typed == "query" || typed == "path" {
				continue
			}
			if typed != "" {
				parts = append(parts, typed)
			}
		case float64:
			if len(parts) == 0 {
				parts = append(parts, strconv.Itoa(int(typed)))
				continue
			}
			parts[len(parts)-1] = fmt.Sprintf("%s[%d]", parts[len(parts)-1], int(typed))
		}
	}
	return strings.Join(parts, ".")
}

func isGenericAPIErrorMessage(message string) bool {
	switch strings.TrimSpace(message) {
	case "", "An HTTP exception occurred.":
		return true
	default:
		return strings.HasPrefix(message, "Request failed (")
	}
}

// errSilent signals that an API error was already rendered to stderr.
var errSilent = errors.New("")

const defaultAPIBaseURL = "https://api.retab.com"

type renderedError struct {
	err error
}

func (e renderedError) Error() string {
	return e.err.Error()
}

func (e renderedError) Unwrap() error {
	return e.err
}

// resolveCLIRequest unifies credential and base-URL resolution for newClient
// and the raw cliJSON/Raw/Multipart helpers, so every request path honors the
// same precedence as the production safety gate (resolveCredential) —
// crucially including the --env / --live / default_environment profile steps,
// which the per-call resolution used to skip. Routing both through one resolver
// guarantees the gate guards the same credential the request actually uses.
//
// cfg is returned for the OAuth dashboard-context provider. The returned
// baseURL is "" when the SDK/default should apply; callers that need a concrete
// URL apply defaultAPIBaseURL themselves.
func resolveCLIRequest(cmd *cobra.Command) (resolvedCredential, string, retabConfig, error) {
	cred, err := resolveCredential(cmd)
	if err != nil {
		return resolvedCredential{}, "", retabConfig{}, err
	}
	cfg, _ := loadConfig()

	// Base-URL override precedence: --base-url flag > RETAB_API_BASE_URL >
	// RETAB_BASE_URL. When none is set, fall back to the profile/config base
	// URL that resolveCredential already resolved. Only the user-supplied
	// flag/env value is validated; the config/profile value is trusted. The
	// "/v1" API version prefix now lives in individual request paths.
	flagBaseURL, _ := cmd.Root().PersistentFlags().GetString("base-url")
	override := flagBaseURL
	if override == "" {
		override = os.Getenv("RETAB_API_BASE_URL")
	}
	if override == "" {
		override = os.Getenv("RETAB_BASE_URL")
	}
	if err := validateBaseURL(override); err != nil {
		return resolvedCredential{}, "", retabConfig{}, err
	}
	baseURL := override
	if baseURL == "" {
		baseURL = cred.BaseURL
	}
	baseURL = stripLegacyV1Suffix(baseURL)
	return cred, baseURL, cfg, nil
}

func newClient(cmd *cobra.Command) (*retab.Client, error) {
	// Resolution order (first match wins), shared with the safety gate via
	// resolveCLIRequest -> resolveCredential:
	//   1. `--api-key` flag                       -> Bearer
	//   2. `RETAB_API_KEY` env                    -> Bearer
	//   3. `--env` / `--live` / default profile   -> Bearer
	//   4. Stored access token                     -> Bearer
	//   5. Stored OAuth tokens                     -> Bearer, transparent refresh
	//   6. Stored legacy api_key                   -> Bearer
	//   7. nothing                                 -> error
	cred, baseURL, cfg, err := resolveCLIRequest(cmd)
	if err != nil {
		return nil, err
	}

	var opts []retab.Option
	if baseURL != "" {
		opts = append(opts, retab.WithBaseURL(baseURL))
	}

	// --debug wires a logging RoundTripper into the SDK's HTTP client so
	// every wire-level request/response is dumped to stderr. The dump
	// uses httputil so headers + body land in a copy-pasteable format
	// (curl-equivalent). Bodies stay in memory; for large uploads this
	// adds RAM pressure — that's fine for a debugging flag.
	authHTTPClient := debugHTTPClient(cmd)
	if authHTTPClient != http.DefaultClient {
		opts = append(opts, retab.WithHTTPClient(authHTTPClient))
	}

	// API-key credential (flag, env, --env/--live/default profile, or legacy).
	if cred.APIKey != "" {
		return retab.NewClient(cred.APIKey, opts...)
	}

	if cred.AccessToken != "" {
		opts = append(opts, retab.WithBearerToken(cred.AccessToken))
		return retab.NewClient("", opts...)
	}

	// OAuth path. WithBearerTokenProvider is invoked on every request, so
	// a command that straddles token expiry still gets a fresh token
	// without rebuilding the Client.
	if cred.OAuth != nil && cred.OAuth.AccessToken != "" {
		rawOAuthProvider := makeOAuthTokenProvider(cred.OAuth)
		opts = append(opts, retab.WithBearerTokenProvider(
			makeCLIAuthTokenProvider(cmd, cfg, baseURL, rawOAuthProvider, authHTTPClient),
		))
		return retab.NewClient("", opts...)
	}

	return nil, fmt.Errorf("no credentials configured. Run `retab auth login` or set RETAB_API_KEY")
}

func bearerTokenForCredential(
	ctx context.Context,
	cmd *cobra.Command,
	cfg retabConfig,
	baseURL string,
	cred resolvedCredential,
	httpClient *http.Client,
) (string, error) {
	if cred.AccessToken != "" {
		return cred.AccessToken, nil
	}
	if cred.OAuth != nil && cred.OAuth.AccessToken != "" {
		rawOAuthProvider := makeOAuthTokenProvider(cred.OAuth)
		return makeCLIAuthTokenProvider(cmd, cfg, baseURL, rawOAuthProvider, httpClient)(ctx)
	}
	return "", nil
}

func selectedEnvironmentID(cmd *cobra.Command, cfg retabConfig) string {
	environmentID, _ := selectedEnvironmentIDWithSource(cmd, cfg)
	return environmentID
}

func selectedEnvironmentIDWithSource(cmd *cobra.Command, cfg retabConfig) (string, string) {
	flagValue := ""
	if cmd != nil && cmd.Root() != nil {
		flagValue, _ = cmd.Root().PersistentFlags().GetString("environment-id")
	}
	if strings.TrimSpace(flagValue) != "" {
		return strings.TrimSpace(flagValue), "--environment-id flag"
	}
	if envValue := strings.TrimSpace(os.Getenv("RETAB_ENVIRONMENT_ID")); envValue != "" {
		return envValue, "RETAB_ENVIRONMENT_ID env"
	}
	if strings.TrimSpace(cfg.EnvironmentID) != "" {
		return strings.TrimSpace(cfg.EnvironmentID), "~/.retab/config.json"
	}
	return "", ""
}

func selectedEnvironmentContextID(cmd *cobra.Command, cfg retabConfig) string {
	if isEnvironmentManagementCommand(cmd) {
		return ""
	}
	return selectedEnvironmentID(cmd, cfg)
}

func isEnvironmentManagementCommand(cmd *cobra.Command) bool {
	for current := cmd; current != nil; current = current.Parent() {
		if current.Name() == "env" {
			return true
		}
	}
	return false
}

func makeCLIAuthTokenProvider(
	cmd *cobra.Command,
	cfg retabConfig,
	baseURL string,
	rawOAuthProvider func(context.Context) (string, error),
	httpClient *http.Client,
) func(context.Context) (string, error) {
	environmentID := selectedEnvironmentContextID(cmd, cfg)
	if environmentID == "" {
		return rawOAuthProvider
	}
	provider := &dashboardContextTokenProvider{
		baseURL:          canonicalAPIBaseURL(baseURL),
		environmentID:    environmentID,
		rawOAuthProvider: rawOAuthProvider,
		httpClient:       httpClient,
	}
	return provider.Token
}

func canonicalAPIBaseURL(baseURL string) string {
	trimmed := strings.TrimSpace(stripLegacyV1Suffix(baseURL))
	if trimmed == "" {
		return defaultAPIBaseURL
	}
	return trimmed
}

type dashboardContextTokenProvider struct {
	baseURL          string
	environmentID    string
	rawOAuthProvider func(context.Context) (string, error)
	httpClient       *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

type dashboardContextTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
}

func (p *dashboardContextTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token != "" && time.Now().Before(p.expiresAt.Add(-refreshLeeway)) {
		return p.token, nil
	}
	rawOAuthToken, err := p.rawOAuthProvider(ctx)
	if err != nil {
		return "", err
	}
	token, expiresAt, err := mintDashboardContextToken(
		ctx,
		p.httpClient,
		p.baseURL,
		rawOAuthToken,
		p.environmentID,
	)
	if err != nil {
		return "", err
	}
	p.token = token
	p.expiresAt = expiresAt
	return p.token, nil
}

func mintDashboardContextToken(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	rawOAuthToken string,
	environmentID string,
) (string, time.Time, error) {
	if strings.TrimSpace(rawOAuthToken) == "" {
		return "", time.Time{}, fmt.Errorf("OAuth access token is empty")
	}
	if strings.TrimSpace(environmentID) == "" {
		return "", time.Time{}, fmt.Errorf("environment id is required")
	}
	requestURL, err := buildCLIRequestURL(baseURL, "/v1/auth/dashboard-context", nil)
	if err != nil {
		return "", time.Time{}, err
	}
	bodyBytes, err := json.Marshal(map[string]any{
		"environment_id": environmentID,
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode dashboard context request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rawOAuthToken)

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", time.Time{}, retab.ParseAPIError(resp, respBody)
	}
	var result dashboardContextTokenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", time.Time{}, fmt.Errorf("decode dashboard context response: %w", err)
	}
	if result.Token == "" {
		return "", time.Time{}, fmt.Errorf("dashboard context response did not include a token")
	}
	if result.TokenType != "" && !strings.EqualFold(result.TokenType, "Bearer") {
		return "", time.Time{}, fmt.Errorf("dashboard context token type %q is not supported", result.TokenType)
	}
	if result.ExpiresAt.IsZero() {
		return "", time.Time{}, fmt.Errorf("dashboard context response did not include expires_at")
	}
	return result.Token, result.ExpiresAt, nil
}

func buildCLIRequestURL(baseURL string, requestPath string, query url.Values) (*url.URL, error) {
	base, err := url.Parse(strings.TrimRight(canonicalAPIBaseURL(baseURL), "/") + "/")
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	relative, err := url.Parse(strings.TrimLeft(requestPath, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid request path: %w", err)
	}
	requestURL := base.ResolveReference(relative)
	if query != nil {
		requestURL.RawQuery = query.Encode()
	}
	return requestURL, nil
}

// debugHTTPClient returns an *http.Client that dumps wire-level request and
// response traffic to stderr when --debug is set, or http.DefaultClient
// otherwise. The 30-minute timeout matches the SDK's default client
// (clients/go/retab.go) so long-running calls (e.g. schema generation) aren't
// silently killed only under --debug. Centralizing this here keeps newClient
// and the raw cli*Request helpers from each re-deriving the same block.
func debugHTTPClient(cmd *cobra.Command) *http.Client {
	if debug, _ := cmd.Root().PersistentFlags().GetBool("debug"); debug {
		return &http.Client{
			Timeout:   30 * time.Minute,
			Transport: &debugTransport{wrapped: http.DefaultTransport},
		}
	}
	return http.DefaultClient
}

// resolveCLIRequestAuth resolves the API key / bearer token pair for the raw
// cli*Request helpers and enforces the shared "no credentials configured"
// guard. Exactly one of apiKey / bearerToken is non-empty on success, matching
// the precedence newClient documents. Folding the three identical credential
// preambles into one resolver guarantees they can't drift apart.
func resolveCLIRequestAuth(
	ctx context.Context,
	cmd *cobra.Command,
	cred resolvedCredential,
	cfg retabConfig,
	baseURL string,
	httpClient *http.Client,
) (apiKey string, bearerToken string, err error) {
	apiKey = cred.APIKey
	if apiKey == "" {
		bearerToken, err = bearerTokenForCredential(ctx, cmd, cfg, baseURL, cred, httpClient)
		if err != nil {
			return "", "", err
		}
	}
	if apiKey == "" && bearerToken == "" {
		return "", "", fmt.Errorf("no credentials configured. Run `retab auth login` or set RETAB_API_KEY")
	}
	return apiKey, bearerToken, nil
}

func cliJSONRequest(cmd *cobra.Command, method string, requestPath string, query url.Values, body any) (any, error) {
	var result any
	if err := cliJSONRequestInto(cmd, method, requestPath, query, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func cliJSONRequestInto(cmd *cobra.Command, method string, requestPath string, query url.Values, body any, result any) error {
	cred, baseURL, cfg, err := resolveCLIRequest(cmd)
	if err != nil {
		return err
	}
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	httpClient := debugHTTPClient(cmd)

	ctx, cancel := ctxFor(cmd)
	defer cancel()

	apiKey, bearerToken, err := resolveCLIRequestAuth(ctx, cmd, cred, cfg, baseURL, httpClient)
	if err != nil {
		return err
	}

	return doCLIJSONRequest(ctx, httpClient, baseURL, method, requestPath, query, body, apiKey, bearerToken, result)
}

func cliRawRequestBytes(
	cmd *cobra.Command,
	method string,
	requestPath string,
	query url.Values,
	body any,
	accept string,
) ([]byte, error) {
	cred, baseURL, cfg, err := resolveCLIRequest(cmd)
	if err != nil {
		return nil, err
	}
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	httpClient := debugHTTPClient(cmd)

	ctx, cancel := ctxFor(cmd)
	defer cancel()

	apiKey, bearerToken, err := resolveCLIRequestAuth(ctx, cmd, cred, cfg, baseURL, httpClient)
	if err != nil {
		return nil, err
	}

	return doCLIRawRequest(ctx, httpClient, baseURL, method, requestPath, query, body, accept, apiKey, bearerToken)
}

func cliMultipartRequestInto(
	cmd *cobra.Command,
	method string,
	requestPath string,
	query url.Values,
	fields map[string]string,
	fileField string,
	filePath string,
	result any,
) error {
	cred, baseURL, cfg, err := resolveCLIRequest(cmd)
	if err != nil {
		return err
	}
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	httpClient := debugHTTPClient(cmd)

	ctx, cancel := ctxFor(cmd)
	defer cancel()

	apiKey, bearerToken, err := resolveCLIRequestAuth(ctx, cmd, cred, cfg, baseURL, httpClient)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("write multipart field %s: %w", key, err)
		}
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	part, err := writer.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("create multipart file field %s: %w", fileField, err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finalize multipart body: %w", err)
	}

	return doCLIBodyRequest(ctx, httpClient, baseURL, method, requestPath, query, &body, writer.FormDataContentType(), apiKey, bearerToken, result)
}

func doCLIJSONRequest(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	method string,
	requestPath string,
	query url.Values,
	body any,
	apiKey string,
	bearerToken string,
	result any,
) error {
	if body == nil && method != http.MethodGet {
		body = map[string]any{}
	}
	var reader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(bodyBytes)
	}
	contentType := ""
	if body != nil {
		contentType = "application/json"
	}

	return doCLIBodyRequest(ctx, httpClient, baseURL, method, requestPath, query, reader, contentType, apiKey, bearerToken, result)
}

func doCLIRawRequest(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	method string,
	requestPath string,
	query url.Values,
	body any,
	accept string,
	apiKey string,
	bearerToken string,
) ([]byte, error) {
	if body == nil && method != http.MethodGet {
		body = map[string]any{}
	}
	var reader io.Reader
	contentType := ""
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(bodyBytes)
		contentType = "application/json"
	}
	requestURL, err := buildCLIRequestURL(baseURL, requestPath, query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), reader)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if accept == "" {
		accept = "*/*"
	}
	req.Header.Set("Accept", accept)
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, retab.ParseAPIError(resp, respBody)
	}
	return respBody, nil
}

func doCLIBodyRequest(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	method string,
	requestPath string,
	query url.Values,
	reader io.Reader,
	contentType string,
	apiKey string,
	bearerToken string,
	result any,
) error {
	requestURL, err := buildCLIRequestURL(baseURL, requestPath, query)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), reader)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Surface the same `APIError` shape the SDK clients return so
		// `runE` renders it through `renderAPIErrorForCLI`. Commands that
		// route through this helper (experiments runs, evals runs list,
		// runs restart, …) used to dump the raw HTTP envelope here while
		// SDK-backed commands rendered "404 — Workflow not found"; that
		// inconsistency confused users probing the CLI for status codes.
		return retab.ParseAPIError(resp, respBody)
	}
	if len(strings.TrimSpace(string(respBody))) == 0 || result == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// makeOAuthTokenProvider returns a closure that yields a current access
// token on demand, refreshing transparently as expiry approaches. A
// successful refresh is persisted to disk (atomically — see saveConfig)
// so subsequent CLI invocations pick up the rotated refresh_token.
//
// Two failure modes are handled specifically:
//
//  1. saveConfig fails. The in-memory token still completes the current
//     request, but the rotated refresh_token never makes it to disk;
//     since WorkOS has already invalidated the previous one, the NEXT
//     CLI invocation would be forced to re-login. We warn loudly on
//     stderr so the user can fix the underlying disk problem before
//     the access_token expires.
//
//  2. refresh returns `invalid_grant`. The likeliest cause is a
//     concurrent CLI invocation that refreshed first and rotated the
//     refresh_token out from under us. We re-read the config file and,
//     if it now holds a different refresh_token, switch to those
//     freshly-rotated credentials and try again — exactly once, to
//     avoid loops if something deeper is wrong.
func makeOAuthTokenProvider(initial *oauthTokens) func(ctx context.Context) (string, error) {
	tok := *initial
	// The SDK calls this provider once per outbound request and fires
	// requests concurrently (parallel uploads, batched runs). On the
	// pure-OAuth path it is used as the bearer provider directly (no
	// dashboardContextTokenProvider mutex in front of it), so without this
	// lock concurrent goroutines would race on `tok` and interleave
	// loadConfig/saveConfig writes. Serializing also collapses a thundering
	// herd of refreshes into one: the first goroutine refreshes, the rest
	// take the fast path on the now-valid token.
	var mu sync.Mutex
	return func(ctx context.Context) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		if time.Now().Before(tok.ExpiresAt.Add(-refreshLeeway)) {
			return tok.AccessToken, nil
		}
		// Serialize the whole refresh+persist across processes. The in-process
		// mutex above only protects this process; a cross-process lock stops
		// two concurrent `retab` invocations from both refreshing and
		// clobbering each other's rotated refresh_token.
		var accessToken string
		var refreshErr error
		_ = withConfigLock(func() error {
			// Re-read under the lock: a process we waited behind may have
			// already minted a fresh, still-valid token. Adopt it and skip our
			// own refresh entirely — this is what prevents the clobber.
			if disk, ldErr := loadConfig(); ldErr == nil && disk.OAuth != nil &&
				disk.OAuth.AccessToken != "" &&
				time.Now().Before(disk.OAuth.ExpiresAt.Add(-refreshLeeway)) {
				tok = *disk.OAuth
				accessToken = tok.AccessToken
				return nil
			}
			refreshed, err := refreshAccessToken(ctx, &tok)
			if err != nil {
				// The refresh_token may have just been rotated out from under
				// us. Re-read disk and try once with whatever is there.
				if isInvalidGrantError(err) {
					if disk, ldErr := loadConfig(); ldErr == nil && disk.OAuth != nil &&
						disk.OAuth.RefreshToken != "" && disk.OAuth.RefreshToken != tok.RefreshToken {
						tok = *disk.OAuth
						if time.Now().Before(tok.ExpiresAt.Add(-refreshLeeway)) {
							accessToken = tok.AccessToken
							return nil
						}
						refreshed, err = refreshAccessToken(ctx, &tok)
					}
				}
				if err != nil {
					refreshErr = err
					return nil
				}
			}
			tok = *refreshed
			accessToken = tok.AccessToken
			// Re-read the rest of the config so we only swap the OAuth block and
			// preserve Environments/BaseURL/legacy key. If the read fails, do
			// NOT persist: writing a zero-value config would wipe those fields.
			// The in-memory tok still serves this request.
			cfg, ldErr := loadConfig()
			if ldErr != nil {
				fmt.Fprintf(os.Stderr,
					"warning: refreshed OAuth token but could not re-read %s to persist it: %v\n"+
						"  current command will succeed; next CLI invocation may require re-login.\n",
					configPathOrEmpty(), ldErr)
				return nil
			}
			cfg.OAuth = &tok
			if err := saveConfig(cfg); err != nil {
				// The in-memory tok works for this request, but if we lose
				// the rotated refresh_token before saving, the next process
				// is forced to re-login. Surface loudly — silent failure
				// here is exactly the "long-lived token mysteriously expires"
				// bug we want to avoid.
				fmt.Fprintf(os.Stderr,
					"warning: refreshed OAuth token but failed to persist to %s: %v\n"+
						"  current command will succeed; next CLI invocation may require re-login.\n",
					configPathOrEmpty(), err)
			}
			return nil
		})
		if refreshErr != nil {
			return "", refreshErr
		}
		return accessToken, nil
	}
}

// configPathOrEmpty returns the config path for diagnostic output, swallowing
// the unlikely error from configPath() — we're already in an error branch.
func configPathOrEmpty() string {
	p, _ := configPath()
	return p
}

// isInvalidGrantError detects the specific OAuth-spec error code that
// refresh_token rotation collisions surface as. We string-match against
// the message produced by postTokenEndpoint — coupling the two via a
// constant would be cleaner, but a string check keeps the surface area
// of this defensive path minimal.
func isInvalidGrantError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "refresh failed:") ||
		strings.Contains(err.Error(), "invalid_grant")
}

func ctxFor(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	parent := context.Background()
	if cmd != nil && cmd.Context() != nil && cmd.Context().Err() == nil {
		parent = cmd.Context()
	}
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}

// printJSON writes v to stdout as indented JSON followed by a newline.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// printNDJSON writes one JSON object per line — used by streaming output.
func printNDJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// normalizeInputBytes converts raw bytes read from a file or stdin into clean
// UTF-8. It transparently strips a leading byte-order mark and decodes UTF-16
// content — the two encodings Windows tooling silently produces. PowerShell's
// `Out-File -Encoding utf8` prepends a UTF-8 BOM, and its default `>` redirection
// writes UTF-16 LE; without this normalization a perfectly valid JSON file fails
// to parse with "invalid character 'ï'", and free-text inputs keep stray BOM
// bytes. Content without a recognized BOM is returned unchanged.
func normalizeInputBytes(raw []byte) []byte {
	switch {
	case len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF:
		return raw[3:] // UTF-8 BOM
	case len(raw) >= 2 && raw[0] == 0xFF && raw[1] == 0xFE:
		return decodeUTF16(raw[2:], binary.LittleEndian)
	case len(raw) >= 2 && raw[0] == 0xFE && raw[1] == 0xFF:
		return decodeUTF16(raw[2:], binary.BigEndian)
	default:
		return raw
	}
}

// decodeUTF16 decodes BOM-stripped UTF-16 bytes (in the given byte order) to
// UTF-8. A stray trailing byte from a truncated code unit is dropped rather than
// failing the read.
func decodeUTF16(b []byte, order binary.ByteOrder) []byte {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u16 := make([]uint16, len(b)/2)
	for i := range u16 {
		u16[i] = order.Uint16(b[i*2:])
	}
	return []byte(string(utf16.Decode(u16)))
}

// readTextFileOrStdin reads raw UTF-8 text from path, or stdin when path is
// "-". Trailing whitespace/newlines are trimmed so a file authored in an
// editor (which appends a final newline) and an inline flag behave the same.
// Unlike readJSON it does not parse the content — it's for free-text inputs
// such as schema-generation instructions.
func readTextFileOrStdin(path string) (string, error) {
	var raw []byte
	var err error
	if path == "-" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	raw = normalizeInputBytes(raw)
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return "", fmt.Errorf("empty text input")
	}
	return text, nil
}

// readJSON reads JSON from path, or stdin when path is "-" or empty.
func readJSON(path string) (any, error) {
	var raw []byte
	var err error
	if path == "" || path == "-" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	raw = normalizeInputBytes(raw)
	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil, fmt.Errorf("empty JSON input")
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return value, nil
}

// readJSONMap decodes JSON into a map[string]any.
func readJSONMap(path string) (map[string]any, error) {
	value, err := readJSON(path)
	if err != nil {
		return nil, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected JSON object")
	}
	return obj, nil
}

// parseJSONMap decodes an inline JSON object literal into a map[string]any.
func parseJSONMap(raw string) (map[string]any, error) {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected JSON object")
	}
	return obj, nil
}

// readJSONArray decodes JSON into a []any.
func readJSONArray(path string) ([]any, error) {
	value, err := readJSON(path)
	if err != nil {
		return nil, err
	}
	arr, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected JSON array")
	}
	return arr, nil
}

// parseKVStringList parses k=v pairs into a map[string]string.
func parseKVStringList(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := map[string]string{}
	for _, raw := range values {
		key, value, ok := strings.Cut(raw, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid key=value pair %q", raw)
		}
		out[key] = value
	}
	return out, nil
}

// validateBaseURL rejects malformed user-supplied base URLs early so the
// failure surfaces as a clear CLI message instead of leaking a Go-internal
// "Get \"/not-a-url/...\": unsupported protocol scheme \"\"" message from
// net/http. Called against the flag / env value before it's plumbed into
// the SDK or http.Request — empty means "fall back to default", which is
// always valid and short-circuits to nil.
//
// Rules: must parse as a URL, must have scheme http or https, must have a
// host. Anything else returns an error with the user-visible flag spelling
// preserved.
func validateBaseURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("--base-url %q is not a valid http(s) URL: %s", raw, err)
	}
	if parsed.Scheme == "" {
		return fmt.Errorf("--base-url %q is not a valid http(s) URL: missing scheme", raw)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("--base-url %q is not a valid http(s) URL: scheme %q is not http or https", raw, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("--base-url %q is not a valid http(s) URL: missing host", raw)
	}
	return nil
}

// stripLegacyV1Suffix removes a trailing "/v1" (with or without trailing
// slash) from baseURL. The SDK used to default baseURL to
// "https://api.retab.com/v1" with paths sent without a version prefix;
// after the WorkOS-pattern regen the default is "https://api.retab.com"
// and every path includes "/v1/..." explicitly. Stored configs and shell
// env vars from before that migration still carry the "/v1" suffix —
// stripping it here keeps those configs working without forcing the user
// to re-run `retab auth login`. New paths like ".../v1/v2" or
// "/v1/anything-but-empty-segment" are left alone.
func stripLegacyV1Suffix(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, "/v1") {
		return strings.TrimSuffix(trimmed, "/v1")
	}
	return baseURL
}

func requireNonBlankFlag(cmd *cobra.Command, name string) (string, error) {
	value, _ := cmd.Flags().GetString(name)
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("--%s must not be blank", name)
	}
	return value, nil
}

// splitKV splits "name=value" into (name, value, true). When no '=' is present,
// it returns (raw, "", false). Used for repeatable k=v flags where the value
// half is optional (compare parseKVStringList, which is strict).
func splitKV(raw string) (string, string, bool) {
	return strings.Cut(raw, "=")
}

// addListFlags attaches the id pagination + filter flags shared by
// most list commands. baseOnly skips filename/from-date/to-date (which only
// apply to file-shaped resources).
//
// `--before` and `--after` are declared mutually exclusive at the cobra
// level — cobra.Execute() will reject any invocation that passes both
// before the RunE body fires. Commands invoked directly from tests via
// `cmd.RunE(...)` bypass that validation; those code paths must also
// run an explicit `validateBeforeAfterMutex` check, otherwise the two
// flags silently end up in the outbound query string together (and the
// server's planner uses whichever the SDK serialized last, dropping the
// other).
func addListFlags(cmd *cobra.Command, baseOnly bool) {
	cmd.Flags().String("before", "", "item id: return items before this id (mutually exclusive with --after)")
	cmd.Flags().String("after", "", "item id: return items after this id (mutually exclusive with --before)")
	cmd.MarkFlagsMutuallyExclusive("before", "after")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	cmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
	if !baseOnly {
		cmd.Flags().String("filename", "", "filter by filename")
		cmd.Flags().Var(&rfc3339FlagValue{}, "from-date", "filter from this date (YYYY-MM-DD or RFC3339)")
		cmd.Flags().Var(&rfc3339FlagValue{}, "to-date", "filter to this date (YYYY-MM-DD or RFC3339)")
	}
}

// validateBeforeAfterMutex returns the canonical mutual-exclusion error
// when both --before and --after were passed. Most workflow-family list
// commands call this from RunE instead of registering cobra's
// `MarkFlagsMutuallyExclusive`: the cobra default message is noisy
// ("if any flags in the group [before after] are set none of the others
// can be ..."), and we prefer the concise "--before and --after are
// mutually exclusive" string. Tests that invoke RunE directly also see
// the error through this path.
func validateBeforeAfterMutex(cmd *cobra.Command) error {
	before, _ := cmd.Flags().GetString("before")
	after, _ := cmd.Flags().GetString("after")
	if before != "" && after != "" {
		return fmt.Errorf("--before and --after are mutually exclusive")
	}
	return nil
}

type nonNegativeIntFlagValue struct{ value string }

func (v *nonNegativeIntFlagValue) String() string {
	if v.value == "" {
		return "0"
	}
	return v.value
}

func (v *nonNegativeIntFlagValue) Type() string { return "int" }

func (v *nonNegativeIntFlagValue) Set(raw string) error {
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}
	if parsed < 0 {
		return fmt.Errorf("must be non-negative")
	}
	v.value = raw
	return nil
}

type positiveIntFlagValue struct{ value string }

func (v *positiveIntFlagValue) String() string {
	if v.value == "" {
		return "0"
	}
	return v.value
}

func (v *positiveIntFlagValue) Type() string { return "int" }

func (v *positiveIntFlagValue) Set(raw string) error {
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}
	if parsed <= 0 {
		return fmt.Errorf("must be positive")
	}
	v.value = raw
	return nil
}

type nonNegativeInt64FlagValue struct{ value string }

func (v *nonNegativeInt64FlagValue) String() string {
	if v.value == "" {
		return "0"
	}
	return v.value
}

func (v *nonNegativeInt64FlagValue) Type() string { return "int64" }

func (v *nonNegativeInt64FlagValue) Set(raw string) error {
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	if parsed < 0 {
		return fmt.Errorf("must be non-negative")
	}
	v.value = raw
	return nil
}

type boundedIntFlagValue struct {
	value string
	min   int
	max   int
}

func (v *boundedIntFlagValue) String() string {
	if v.value == "" {
		return "0"
	}
	return v.value
}

func (v *boundedIntFlagValue) Type() string { return "int" }

func (v *boundedIntFlagValue) Set(raw string) error {
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}
	if parsed < v.min || parsed > v.max {
		if v.min == 0 && parsed < 0 {
			return fmt.Errorf("must be non-negative and between %d and %d", v.min, v.max)
		}
		return fmt.Errorf("must be between %d and %d", v.min, v.max)
	}
	v.value = raw
	return nil
}

type nonNegativeFloatFlagValue struct{ value string }

func (v *nonNegativeFloatFlagValue) String() string {
	if v.value == "" {
		return "0"
	}
	return v.value
}

func (v *nonNegativeFloatFlagValue) Type() string { return "float64" }

func (v *nonNegativeFloatFlagValue) Set(raw string) error {
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return err
	}
	if parsed < 0 {
		return fmt.Errorf("must be non-negative")
	}
	v.value = raw
	return nil
}

// rfc3339FlagValue accepts either a date-only (YYYY-MM-DD) bound or a full
// RFC3339 timestamp for --from-date/--to-date filters. The file-shaped list
// backends (files/parses/classifications/...) parse the bound with parseISO
// (StartOfDayUTC/EndOfDayUTC), which pins a bare date to the start/end of the
// UTC day, so the natural `--from-date 2026-06-13` is valid at the API. Earlier
// this flag demanded full RFC3339 and rejected the bare date even though the
// server accepts it; the lenient set of layouts mirrors the backend's parseISO.
type rfc3339FlagValue struct{ value string }

func (v *rfc3339FlagValue) String() string { return v.value }

func (v *rfc3339FlagValue) Type() string { return "string" }

func (v *rfc3339FlagValue) Set(raw string) error {
	if raw != "" {
		parsed := false
		for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05"} {
			if _, err := time.Parse(layout, raw); err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			return fmt.Errorf("must be a YYYY-MM-DD date or RFC3339 timestamp")
		}
	}
	v.value = raw
	return nil
}

type dateFlagValue struct{ value string }

func (v *dateFlagValue) String() string { return v.value }

func (v *dateFlagValue) Type() string { return "string" }

func (v *dateFlagValue) Set(raw string) error {
	if raw != "" {
		if _, err := time.Parse("2006-01-02", raw); err != nil {
			return fmt.Errorf("must use YYYY-MM-DD date format: %w", err)
		}
	}
	v.value = raw
	return nil
}

// validateDateRange rejects reversed --from-date / --to-date pairs. Without
// this, a typo (e.g. swapping the two values) silently returns the empty set
// — indistinguishable from "no runs match the filter", which is a common
// source of confusion when triaging "where did my runs go?".
//
// Both args are YYYY-MM-DD strings already validated by dateFlagValue.Set
// (or equivalent), so this function trusts the format. Empty strings mean
// the user didn't pass that side of the range — the function returns nil
// in that case so the existing partial-range semantics keep working.
func validateDateRange(fromDateFlag, toDateFlag, fromVal, toVal string) error {
	if fromVal == "" || toVal == "" {
		return nil
	}
	from, fromErr := time.Parse("2006-01-02", fromVal)
	to, toErr := time.Parse("2006-01-02", toVal)
	if fromErr != nil || toErr != nil {
		// Shouldn't happen — dateFlagValue.Set rejects bad input. If it
		// does, defer to the server: this validator is not the right
		// place to surface a format error a second time.
		return nil
	}
	if from.After(to) {
		return fmt.Errorf(
			"--%s %s is after --%s %s (date range is reversed; pass the older date as --%s)",
			fromDateFlag, fromVal, toDateFlag, toVal, fromDateFlag,
		)
	}
	return nil
}

// validateOrderFlag and validateDateFlag let plain `String` flags reuse
// the same parsing as `orderFlagValue` / `dateFlagValue` without changing
// flag registration. Useful for commands like `workflows evals runs list`
// that historically forwarded raw query strings to the server and need
// retrofitted client-side validation.

func validateOrderFlag(cmd *cobra.Command, name string) error {
	value, _ := cmd.Flags().GetString(name)
	switch value {
	case "", "asc", "desc":
		return nil
	default:
		return fmt.Errorf("--%s %q must be asc or desc", name, value)
	}
}

func validateDateFlag(cmd *cobra.Command, name string) error {
	value, _ := cmd.Flags().GetString(name)
	if value == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("--%s must use YYYY-MM-DD date format: %w", name, err)
	}
	return nil
}

type orderFlagValue struct{ value string }

func (v *orderFlagValue) String() string { return v.value }

func (v *orderFlagValue) Type() string { return "string" }

func (v *orderFlagValue) Set(raw string) error {
	switch raw {
	case "", "asc", "desc":
		v.value = raw
		return nil
	default:
		return fmt.Errorf("must be asc or desc")
	}
}

type enumStringFlagValue struct {
	value   string
	allowed map[string]bool
	values  []string
	label   string
}

func newEnumStringFlagValue(label string, allowedValues ...string) *enumStringFlagValue {
	allowed := map[string]bool{"": true}
	for _, value := range allowedValues {
		allowed[value] = true
	}
	return &enumStringFlagValue{allowed: allowed, values: allowedValues, label: label}
}

func (v *enumStringFlagValue) String() string { return v.value }

func (v *enumStringFlagValue) Type() string { return "string" }

func (v *enumStringFlagValue) Set(raw string) error {
	if v.allowed[raw] {
		v.value = raw
		return nil
	}
	return fmt.Errorf("invalid %s %q (want: %s)", v.label, raw, strings.Join(v.values, " | "))
}

var sha256HexPattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

type sha256FlagValue struct{ value string }

func (v *sha256FlagValue) String() string { return v.value }

func (v *sha256FlagValue) Type() string { return "string" }

func (v *sha256FlagValue) Set(raw string) error {
	if raw != "" && !sha256HexPattern.MatchString(raw) {
		return fmt.Errorf("must be a 64-character hex SHA-256 digest")
	}
	v.value = raw
	return nil
}

func ptr[T any](value T) *T {
	return &value
}

// resolveWorkflowScope reconciles the two co-equal ways a workflow-scoped
// `list` command names its workflow: a positional `[workflow-id]` argument and
// a `--workflow-id` flag. Both forms are first-class — neither is deprecated —
// so a user can pick whichever reads better. This is the list-command
// counterpart to resolveWorkflowIDArg, which keeps a deprecation warning on the
// flag for create-style commands.
//
// Resolution rules:
//
//   - both set, agreeing → that id.
//   - both set, disagreeing → error (never silently pick one; a mismatch is
//     almost always a typo, and masking it would query the wrong workflow).
//   - exactly one set → that id (whitespace-trimmed).
//   - an explicitly-blank flag (`--workflow-id ""`) → error, since pflag
//     returns "" which would otherwise silently widen the query.
//   - neither set:
//   - required=false → "" (the caller lists workspace-wide, e.g. runs/reviews).
//   - required=true  → error (the command has no org-wide view, e.g.
//     blocks/edges/tests/experiments).
func resolveWorkflowScope(cmd *cobra.Command, args []string, required bool) (string, error) {
	flagID, _ := cmd.Flags().GetString("workflow-id")
	if cmd.Flags().Changed("workflow-id") && strings.TrimSpace(flagID) == "" {
		return "", fmt.Errorf("--workflow-id must not be blank")
	}
	flagID = strings.TrimSpace(flagID)

	posID := ""
	if len(args) > 0 {
		posID = strings.TrimSpace(args[0])
	}

	switch {
	case posID != "" && flagID != "" && posID != flagID:
		return "", fmt.Errorf("workflow id specified twice (positional %q, --workflow-id %q)", posID, flagID)
	case posID != "":
		return posID, nil
	case flagID != "":
		return flagID, nil
	}
	if required {
		return "", fmt.Errorf("workflow id required")
	}
	return "", nil
}

func collectListParams(cmd *cobra.Command) retab.PaginationParams {
	params := retab.PaginationParams{}
	if v, _ := cmd.Flags().GetString("before"); v != "" {
		params.Before = ptr(v)
	}
	if v, _ := cmd.Flags().GetString("after"); v != "" {
		params.After = ptr(v)
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Limit = ptr(v)
	}
	if v, _ := cmd.Flags().GetString("order"); v != "" {
		params.Order = ptr(v)
	}
	return params
}

func collectFileDateListFilters(cmd *cobra.Command) (filename, fromDate, toDate *string) {
	if v, _ := cmd.Flags().GetString("filename"); v != "" {
		filename = ptr(v)
	}
	if v, _ := cmd.Flags().GetString("from-date"); v != "" {
		fromDate = ptr(v)
	}
	if v, _ := cmd.Flags().GetString("to-date"); v != "" {
		toDate = ptr(v)
	}
	return filename, fromDate, toDate
}

// validateListDateRange rejects a reversed --from-date / --to-date pair on the
// file-shaped list commands, matching the check `workflows runs list` already
// performs. Without it a swapped pair silently returns the empty set —
// indistinguishable from "nothing matched", a common "where did my items go?"
// trap. Safe to call unconditionally: validateDateRange returns nil when either
// bound is empty or not a plain date (e.g. an RFC3339 timestamp).
func validateListDateRange(cmd *cobra.Command) error {
	fromDate, _ := cmd.Flags().GetString("from-date")
	toDate, _ := cmd.Flags().GetString("to-date")
	return validateDateRange("from-date", "to-date", fromDate, toDate)
}

func getIntFlagOrDefault(cmd *cobra.Command, name string, defaultValue int) int {
	if !cmd.Flags().Changed(name) {
		return defaultValue
	}
	value, _ := cmd.Flags().GetInt(name)
	if value > 0 {
		return value
	}
	return defaultValue
}

// addDocumentFlags adds the mutually-exclusive document source flags shared
// by every command that takes a document body.
func addDocumentFlags(cmd *cobra.Command) {
	cmd.Flags().String("file", "", "path to a local document")
	cmd.Flags().String("url", "", "https URL of a document")
	cmd.Flags().String("file-id", "", "Retab file id")
	cmd.Flags().String("document-file", "", "path to a JSON file describing the document (or - for stdin)")
}

// scopedResourceID resolves the trailing resource id from args while tolerating
// an optional leading workflow id (wrk_...). Run-, step-, and artifact-scoped
// commands address their resource globally (the API routes don't carry a
// workflow id), so they take a single id — unlike blocks/edges, which are
// workflow-scoped and take `<workflow-id> <child-id>`. Users routinely carry
// the blocks/edges habit over and type `<workflow-id> <run-id>`; rather than
// failing with cobra's bare "accepts 1 arg(s), received 2", accept the extra
// leading arg when it is unmistakably a workflow id and use the real id. A
// second arg that is NOT a workflow id is a genuine mistake and still errors.
func scopedResourceID(args []string, resourceLabel string) (string, error) {
	switch len(args) {
	case 1:
		id := strings.TrimSpace(args[0])
		if id == "" {
			return "", fmt.Errorf("expected the %s", resourceLabel)
		}
		return id, nil
	case 2:
		if strings.HasPrefix(strings.TrimSpace(args[0]), "wrk_") {
			id := strings.TrimSpace(args[1])
			if id == "" {
				return "", fmt.Errorf("expected the %s", resourceLabel)
			}
			return id, nil
		}
		return "", fmt.Errorf("this command takes only the %s (no workflow id); got %q and %q", resourceLabel, args[0], args[1])
	default:
		return "", fmt.Errorf("expected the %s", resourceLabel)
	}
}

// inferFileMIMEData turns a local file path into MIMEData, statting the
// path upfront so a bad path surfaces as a clear "file not found" instead
// of being routed through the SDK's MIME inference and resurfacing as the
// cryptic "unsupported MIME input string" (which is what InferMIMEData
// reports when it can't open the file for sniffing). Directories stat
// fine but break MIME sniffing the same way, so they get an explicit
// check too. Shared by every command that accepts a local-file document
// flag.
func inferFileMIMEData(path string) (retab.MIMEData, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return retab.MIMEData{}, fmt.Errorf("file not found: %s", path)
		}
		return retab.MIMEData{}, fmt.Errorf("cannot read %s: %w", path, err)
	}
	if info.IsDir() {
		return retab.MIMEData{}, fmt.Errorf("not a file (is a directory): %s", path)
	}
	return retab.InferMIMEData(path)
}

// resolveDocument turns the document flags into a value the SDK can marshal.
// At most one of file / url / file-id / document-file must be set.
func resolveDocument(cmd *cobra.Command) (any, error) {
	file, _ := cmd.Flags().GetString("file")
	urlStr, _ := cmd.Flags().GetString("url")
	fileID, _ := cmd.Flags().GetString("file-id")
	docFile, _ := cmd.Flags().GetString("document-file")
	count := 0
	for _, v := range []string{file, urlStr, fileID, docFile} {
		if v != "" {
			count++
		}
	}
	if count == 0 {
		return nil, fmt.Errorf("one of --file, --url, --file-id, or --document-file is required")
	}
	if count > 1 {
		return nil, fmt.Errorf("--file, --url, --file-id, and --document-file are mutually exclusive")
	}
	switch {
	case file != "":
		mime, err := inferFileMIMEData(file)
		if err != nil {
			return nil, err
		}
		return mime, nil
	case urlStr != "":
		// Server requires `filename` on every document descriptor —
		// `{"url": "..."}` alone returns HTTP 422. Derive from the URL
		// path's last segment; fall back to "document" for path-less URLs.
		return retab.MIMEData{Filename: filenameFromURL(urlStr), URL: urlStr}, nil
	case fileID != "":
		// Same filename-required constraint. The SDK's `FileRef{ID: ...}`
		// shape no longer satisfies the server contract on its own. Look
		// the file up to fetch its filename and a fresh download URL,
		// then send a full MIMEData. One extra GET per command — fine
		// for the readability win on every --file-id callsite.
		return resolveFileIDToMIMEData(cmd, fileID)
	case docFile != "":
		// Same upfront stat as --file: a missing JSON descriptor would
		// otherwise bubble out as a json-unmarshal error wrapped around
		// "no such file or directory" — confusing for users who just
		// fat-fingered the path. Skip the stat for "-" (stdin), which
		// readJSON handles.
		if docFile != "-" {
			if _, err := os.Stat(docFile); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return nil, fmt.Errorf("file not found: %s", docFile)
				}
				return nil, fmt.Errorf("cannot read %s: %w", docFile, err)
			}
		}
		return readJSON(docFile)
	}
	return nil, fmt.Errorf("unreachable")
}

func validateDocumentFlagSelection(cmd *cobra.Command) error {
	file, _ := cmd.Flags().GetString("file")
	urlStr, _ := cmd.Flags().GetString("url")
	fileID, _ := cmd.Flags().GetString("file-id")
	docFile, _ := cmd.Flags().GetString("document-file")
	count := 0
	for _, v := range []string{file, urlStr, fileID, docFile} {
		if v != "" {
			count++
		}
	}
	if count == 0 {
		return fmt.Errorf("one of --file, --url, --file-id, or --document-file is required")
	}
	if count > 1 {
		return fmt.Errorf("--file, --url, --file-id, and --document-file are mutually exclusive")
	}
	return nil
}

// filenameFromURL returns the basename of a URL path, or "document" when
// the path is empty / root. Used to satisfy the server's `filename`
// requirement on document descriptors when only a `--url` was given.
func filenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err == nil && u.Path != "" {
		base := path.Base(u.Path)
		if base != "/" && base != "." && base != "" {
			return base
		}
	}
	return "document"
}

// resolveFileIDToMIMEData fetches the file metadata for fileID and
// returns a MIMEData populated with filename + a fresh download URL.
// Costs one round trip (sometimes two — Files.Get for filename + a
// signed download link for the URL). Kept in one place so all the
// document-taking commands behave identically.
func resolveFileIDToMIMEData(cmd *cobra.Command, fileID string) (retab.MIMEData, error) {
	client, err := newClient(cmd)
	if err != nil {
		return retab.MIMEData{}, err
	}
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	mimeData, err := resolveFileIDToMIMEDataWithClient(ctx, client, fileID)
	if err != nil {
		return retab.MIMEData{}, fmt.Errorf("resolving --file-id %s: %w", fileID, err)
	}
	return mimeData, nil
}

// resolveFileIDToMIMEDataWithClient resolves a stored file id into MIMEData
// populated with a filename + a fresh signed download URL, reusing a client and
// context the caller already holds. The workflow-runs and single-document
// routes accept MIMEData only — a bare FileRef{id} body is rejected (422) — so
// every file-id input is resolved here before it is sent. Callers add their own
// flag/field context to the returned error.
func resolveFileIDToMIMEDataWithClient(ctx context.Context, client *retab.Client, fileID string) (retab.MIMEData, error) {
	link, err := client.Files.GetDownloadLink(ctx, fileID)
	if err != nil {
		return retab.MIMEData{}, err
	}
	if link.DownloadURL == "" {
		return retab.MIMEData{}, fmt.Errorf("server returned no download URL")
	}
	if link.MIMEData != nil && link.MIMEData.URL != "" {
		if link.MIMEData.Filename == "" {
			link.MIMEData.Filename = link.Filename
		}
		if link.MIMEData.Filename == "" {
			link.MIMEData.Filename = "document"
		}
		return *link.MIMEData, nil
	}
	filename := link.Filename
	if filename == "" {
		filename = "document"
	}
	return retab.MIMEData{Filename: filename, URL: link.DownloadURL}, nil
}

func resolveFileIDToSignedDownloadMIMEData(cmd *cobra.Command, fileID string) (retab.MIMEData, error) {
	client, err := newClient(cmd)
	if err != nil {
		return retab.MIMEData{}, err
	}
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	link, err := client.Files.GetDownloadLink(ctx, fileID)
	if err != nil {
		return retab.MIMEData{}, fmt.Errorf("resolving --file-id %s: %w", fileID, err)
	}
	if link.DownloadURL == "" {
		return retab.MIMEData{}, fmt.Errorf("--file-id %s: server returned no download URL", fileID)
	}
	filename := link.Filename
	if filename == "" && link.MIMEData != nil {
		filename = link.MIMEData.Filename
	}
	if filename == "" {
		filename = "document"
	}
	return retab.MIMEData{Filename: filename, URL: link.DownloadURL}, nil
}

// resolveOptionalDocument is like resolveDocument but returns (nil, nil)
// when no flag is set.
func resolveOptionalDocument(cmd *cobra.Command) (any, error) {
	file, _ := cmd.Flags().GetString("file")
	urlStr, _ := cmd.Flags().GetString("url")
	fileID, _ := cmd.Flags().GetString("file-id")
	docFile, _ := cmd.Flags().GetString("document-file")
	if file == "" && urlStr == "" && fileID == "" && docFile == "" {
		return nil, nil
	}
	return resolveDocument(cmd)
}

func mimeDataFromDocument(doc any) (retab.MIMEData, error) {
	switch value := doc.(type) {
	case retab.MIMEData:
		// Return as-is: a content-only descriptor (Content + MIMEType, no
		// URL) is a valid inline document. Copying only Filename/URL drops
		// the inline content and serializes an empty document over the wire.
		return value, nil
	case *retab.MIMEData:
		if value == nil {
			return retab.MIMEData{}, fmt.Errorf("document is required")
		}
		return *value, nil
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return retab.MIMEData{}, fmt.Errorf("encode document: %w", err)
	}
	var result retab.MIMEData
	if err := json.Unmarshal(body, &result); err != nil {
		return retab.MIMEData{}, fmt.Errorf("decode document: %w", err)
	}
	return result, nil
}

// materializeInlineMIMEData ensures a document is carried as inline base64 bytes
// rather than a remote URL reference. Routes that persist the document
// server-side via the inline-only persist seam (edit templates, Excel sample
// docs) cannot dereference a remote URL — a `--url` or a `--file-id` (which
// resolves to a signed storage URL) would otherwise 500 with "failed to persist
// document". An already-inline `data:` document (e.g. from `--file`) and a
// content-only descriptor are returned unchanged; a remote-URL document is
// downloaded once and re-inlined with the SDK's own MIME detection.
func materializeInlineMIMEData(ctx context.Context, doc retab.MIMEData) (retab.MIMEData, error) {
	if doc.URL == "" || strings.HasPrefix(doc.URL, "data:") {
		return doc, nil
	}
	if !strings.HasPrefix(doc.URL, "http://") && !strings.HasPrefix(doc.URL, "https://") {
		return doc, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, doc.URL, nil)
	if err != nil {
		return retab.MIMEData{}, err
	}
	resp, err := fileDownloadClient.Do(req)
	if err != nil {
		return retab.MIMEData{}, fmt.Errorf("download document for inline upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return retab.MIMEData{}, fmt.Errorf("download document failed: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return retab.MIMEData{}, err
	}
	inline, err := retab.InferMIMEData(data)
	if err != nil {
		return retab.MIMEData{}, fmt.Errorf("inline document: %w", err)
	}
	if doc.Filename != "" {
		inline.Filename = doc.Filename
	}
	return inline, nil
}

// resolveSchema reads a JSON schema from --json-schema (JSON literal) or
// --json-schema-file (path to JSON file, or - for stdin).
func resolveSchema(cmd *cobra.Command) (any, error) {
	literal, _ := cmd.Flags().GetString("json-schema")
	path, _ := cmd.Flags().GetString("json-schema-file")
	if literal != "" && path != "" {
		return nil, fmt.Errorf("--json-schema and --json-schema-file are mutually exclusive")
	}
	if literal == "" && path == "" {
		return nil, fmt.Errorf("one of --json-schema or --json-schema-file is required")
	}
	if literal != "" {
		var v any
		if err := json.Unmarshal([]byte(literal), &v); err != nil {
			return nil, fmt.Errorf("invalid --json-schema: %w", err)
		}
		return v, nil
	}
	value, err := readJSON(path)
	if err != nil {
		return nil, fmt.Errorf("--json-schema-file: %w", err)
	}
	return value, nil
}

// addSchemaFlags adds the JSON-schema source flags used by extractions.
func addSchemaFlags(cmd *cobra.Command) {
	cmd.Flags().String("json-schema", "", "JSON schema literal")
	cmd.Flags().String("json-schema-file", "", "path to JSON schema file (or - for stdin)")
}

// debugTransport wraps an http.RoundTripper and dumps every request +
// response to stderr. Activated by `--debug` on the root command. Output
// is in HTTP wire format so it's pasteable into other tools (httpie,
// requestbin, etc.) without translation.
//
// Sensitive auth headers are redacted in the
// dump. Without this guard, sharing a `--debug` log with another engineer
// for a bug report would leak the user's full API key, which is the exact
// scenario where the redaction matters.
type debugTransport struct {
	wrapped http.RoundTripper
}

// sensitiveHeaders is the list of HTTP header names whose values must be
// replaced with a redacted preview in --debug output. Lowercase because
// the comparison is case-insensitive (Go's http.Header normalises).
var sensitiveHeaders = map[string]bool{
	"api-key":       true,
	"authorization": true,
}

// maxDebugDumpBody bounds how many bytes of a request/response body --debug
// copies into the dump string. The wire body is never affected — only the
// human-readable dump is capped — so dumping an extraction request that
// carries a large inline base64 document no longer doubles its size in
// memory just to print it.
const maxDebugDumpBody = 256 << 10 // 256 KiB

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var requestBody []byte
	if req.Body != nil {
		body, readErr := io.ReadAll(req.Body)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "--- HTTP debug read error ---\n%v\n", readErr)
		} else {
			_ = req.Body.Close()
			requestBody = body
			req.Body = io.NopCloser(bytes.NewReader(requestBody))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(requestBody)), nil
			}
			req.ContentLength = int64(len(body))
		}
	}

	// Clone after restoring the request body so the debug dump cannot consume
	// the body that still needs to go over the wire.
	dumpReq := req.Clone(req.Context())
	redactSensitiveHeaders(dumpReq.Header)
	if requestBody != nil {
		debugRequestBody := redactSensitiveDebugBody(requestBody, req.Header.Get("Content-Type"), req.URL.Path)
		dumpReq.Body = io.NopCloser(bytes.NewReader(debugRequestBody))
		dumpReq.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(debugRequestBody)), nil
		}
		dumpReq.ContentLength = int64(len(debugRequestBody))
	}
	includeReqBody := len(requestBody) <= maxDebugDumpBody
	if dump, err := httputil.DumpRequestOut(dumpReq, includeReqBody); err == nil {
		fmt.Fprintf(os.Stderr, "\n--- HTTP request ---\n%s\n", dump)
		if !includeReqBody {
			fmt.Fprintf(os.Stderr, "<request body omitted from --debug: %d bytes>\n", len(requestBody))
		}
	}
	resp, err := t.wrapped.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "--- HTTP error ---\n%v\n", err)
		return nil, err
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "--- HTTP debug read error ---\n%v\n", readErr)
		return resp, nil
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	includeRespBody := len(body) <= maxDebugDumpBody
	dumpResp := *resp
	debugResponseBody := redactSensitiveDebugBody(body, resp.Header.Get("Content-Type"), req.URL.Path)
	dumpResp.Body = io.NopCloser(bytes.NewReader(debugResponseBody))
	dumpResp.ContentLength = int64(len(debugResponseBody))
	if dump, err := httputil.DumpResponse(&dumpResp, includeRespBody); err == nil {
		fmt.Fprintf(os.Stderr, "--- HTTP response ---\n%s\n", dump)
		if !includeRespBody {
			fmt.Fprintf(os.Stderr, "<response body omitted from --debug: %d bytes>\n", len(body))
		}
	}
	return resp, nil
}

var sensitiveJSONFields = map[string]bool{
	"api_key":       true,
	"access_token":  true,
	"refresh_token": true,
	"id_token":      true,
	"token":         true,
}

func redactSensitiveDebugBody(body []byte, contentType, path string) []byte {
	if len(body) == 0 {
		return body
	}
	// Redact any secrets-resource body wholesale. Match both the collection
	// path (".../secrets", e.g. POST create whose body carries a value) and any
	// sub-resource (".../secrets/<name>[/value]"). The trailing-slash-only check
	// missed the bare collection path, which would dump a create body in clear.
	// HasSuffix/"/secrets/" avoids matching unrelated paths like "/secretsfoo".
	if strings.Contains(path, "/secrets/") || strings.HasSuffix(path, "/secrets") {
		return []byte("[REDACTED]")
	}
	if !strings.Contains(strings.ToLower(contentType), "json") {
		return body
	}
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return body
	}
	redactSensitiveJSONValue(value)
	redacted, err := json.Marshal(value)
	if err != nil {
		return body
	}
	return redacted
}

func redactSensitiveJSONValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if sensitiveJSONFields[strings.ToLower(key)] {
				typed[key] = "[REDACTED]"
				continue
			}
			redactSensitiveJSONValue(child)
		}
	case []any:
		for _, child := range typed {
			redactSensitiveJSONValue(child)
		}
	}
}

// redactSensitiveHeaders replaces credential-carrying header values in
// place with a short prefix+suffix preview (using the same redactKey
// shape as `retab auth status`'s `api_key_preview`). Idempotent on
// headers that don't appear in the request.
func redactSensitiveHeaders(h http.Header) {
	for name := range h {
		if !sensitiveHeaders[strings.ToLower(name)] {
			continue
		}
		for i, v := range h[name] {
			// Authorization: "Bearer <token>" — preserve the scheme so
			// users debugging an auth flow still see WHICH scheme the
			// CLI selected; only the credential body is redacted.
			if scheme, rest, ok := strings.Cut(v, " "); ok && rest != "" &&
				(scheme == "Bearer" || scheme == "Basic") {
				h[name][i] = scheme + " " + redactKey(rest)
				continue
			}
			h[name][i] = redactKey(v)
		}
	}
}

// confirmDestructive gates a destructive command behind an explicit user
// confirmation. Returns nil if the caller passed --yes; otherwise prompts on
// stderr and reads the answer from stdin (TTY only). When stdin is not a TTY
// — pipes, redirected files, CI — refuses with a clear error rather than
// auto-accepting a stray newline.
func confirmDestructive(cmd *cobra.Command, kind, id string) error {
	if yes, _ := cmd.Flags().GetBool("yes"); yes {
		return nil
	}
	// --confirm (the production-mutation acknowledgment shared by every high-risk
	// command) also satisfies the destructive gate, so a production delete needs
	// only --confirm rather than both --confirm and --yes. --yes still stands on
	// its own for deletes in non-production environments, where productionGate is
	// a no-op and --confirm is not required.
	if confirm, _ := cmd.Flags().GetBool(confirmFlagName); confirm {
		return nil
	}
	stdin, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(stdin.Fd())) {
		return fmt.Errorf("refusing to delete %s %q without --yes (stdin is not a terminal)", kind, id)
	}
	if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Permanently delete %s %s? Type the id to confirm: ", kind, id); err != nil {
		return err
	}
	answer, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read confirmation: %w", err)
	}
	if strings.TrimSpace(answer) != id {
		return fmt.Errorf("aborted: %s %q not deleted", kind, id)
	}
	return nil
}

// confirmDeleted writes a one-line confirmation to stderr after a
// successful delete. Stderr (not stdout) so users piping a delete
// command into another process don't get the confirmation in their
// data stream — the JSON / no-content body still goes to stdout per
// the rest of the CLI's convention.
//
// Quiet by design: a single line, no decoration. The user is mostly
// interested in "did the right thing happen?" and the resource id is
// the load-bearing piece — fat-fingered ids are the failure mode this
// guards against.
func confirmDeleted(kind, id string) {
	if f := rootCmd.PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputJSON) {
		if err := printJSON(map[string]any{"id": id, "deleted": true}); err != nil {
			fmt.Fprintln(os.Stderr, "error: "+err.Error())
		}
		return
	}
	fmt.Fprintf(os.Stderr, "deleted %s: %s\n", kind, id)
}
