package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

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
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		return renderedError{err: err}
	}
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

type renderedError struct {
	err error
}

func (e renderedError) Error() string {
	return e.err.Error()
}

func (e renderedError) Unwrap() error {
	return e.err
}

func newClient(cmd *cobra.Command) (*retab.Client, error) {
	// Resolution order (first match wins):
	//   1. `--api-key` flag        -> Api-Key header
	//   2. `RETAB_API_KEY` env     -> Api-Key header
	//   3. Stored OAuth tokens     -> Bearer header, with transparent refresh
	//   4. Stored legacy api_key   -> Api-Key header
	//   5. nothing                 -> error
	flagKey, _ := cmd.Root().PersistentFlags().GetString("api-key")
	flagBaseURL, _ := cmd.Root().PersistentFlags().GetString("base-url")

	apiKey := flagKey
	baseURL := flagBaseURL
	if apiKey == "" {
		apiKey = os.Getenv("RETAB_API_KEY")
	}
	if baseURL == "" {
		baseURL = os.Getenv("RETAB_API_BASE_URL")
	}
	if baseURL == "" {
		baseURL = os.Getenv("RETAB_BASE_URL")
	}

	cfg, _ := loadConfig()
	if baseURL == "" {
		baseURL = cfg.BaseURL
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
	if debug, _ := cmd.Root().PersistentFlags().GetBool("debug"); debug {
		opts = append(opts, retab.WithHTTPClient(&http.Client{
			Timeout:   60 * time.Second,
			Transport: &debugTransport{wrapped: http.DefaultTransport},
		}))
	}

	// Flag/env API key wins outright — the documented CI escape hatch.
	if apiKey != "" {
		return retab.NewClient(apiKey, opts...)
	}

	// OAuth path. WithBearerTokenProvider is invoked on every request, so
	// a command that straddles token expiry still gets a fresh token
	// without rebuilding the Client.
	if cfg.OAuth != nil && cfg.OAuth.AccessToken != "" {
		opts = append(opts, retab.WithBearerTokenProvider(makeOAuthTokenProvider(cfg.OAuth)))
		return retab.NewClient("", opts...)
	}

	// Legacy `api_key` field from ~/.retab/config.json.
	if cfg.APIKey != "" {
		return retab.NewClient(cfg.APIKey, opts...)
	}

	return nil, fmt.Errorf("no credentials configured. Run `retab auth login` or set RETAB_API_KEY")
}

func cliJSONRequest(cmd *cobra.Command, method string, requestPath string, query url.Values, body any) (any, error) {
	flagKey, _ := cmd.Root().PersistentFlags().GetString("api-key")
	flagBaseURL, _ := cmd.Root().PersistentFlags().GetString("base-url")

	apiKey := flagKey
	baseURL := flagBaseURL
	if apiKey == "" {
		apiKey = os.Getenv("RETAB_API_KEY")
	}
	if baseURL == "" {
		baseURL = os.Getenv("RETAB_API_BASE_URL")
	}
	if baseURL == "" {
		baseURL = os.Getenv("RETAB_BASE_URL")
	}

	cfg, _ := loadConfig()
	if baseURL == "" {
		baseURL = cfg.BaseURL
	}
	if baseURL == "" {
		baseURL = "https://api.retab.com/v1"
	}

	var bearerToken string
	if apiKey == "" {
		if cfg.OAuth != nil && cfg.OAuth.AccessToken != "" {
			token, err := makeOAuthTokenProvider(cfg.OAuth)(cmd.Context())
			if err != nil {
				return nil, err
			}
			bearerToken = token
		} else if cfg.APIKey != "" {
			apiKey = cfg.APIKey
		}
	}
	if apiKey == "" && bearerToken == "" {
		return nil, fmt.Errorf("no credentials configured. Run `retab auth login` or set RETAB_API_KEY")
	}

	if body == nil && method != http.MethodGet {
		body = map[string]any{}
	}
	var reader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(bodyBytes)
	}

	base, err := url.Parse(strings.TrimRight(baseURL, "/") + "/")
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

	ctx, cancel := ctxFor(cmd)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	} else {
		req.Header.Set("Api-Key", apiKey)
	}

	httpClient := http.DefaultClient
	if debug, _ := cmd.Root().PersistentFlags().GetBool("debug"); debug {
		httpClient = &http.Client{
			Timeout:   60 * time.Second,
			Transport: &debugTransport{wrapped: http.DefaultTransport},
		}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s failed with status %d: %s", method, requestURL.String(), resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if len(strings.TrimSpace(string(respBody))) == 0 {
		return nil, nil
	}
	var result any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
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
	return func(ctx context.Context) (string, error) {
		if time.Now().Before(tok.ExpiresAt.Add(-refreshLeeway)) {
			return tok.AccessToken, nil
		}
		refreshed, err := refreshAccessToken(ctx, &tok)
		if err != nil {
			// Concurrent-refresh race: someone else may have rotated
			// the refresh_token out from under us. Re-read disk and try
			// once with whatever's there.
			if isInvalidGrantError(err) {
				if disk, ldErr := loadConfig(); ldErr == nil && disk.OAuth != nil &&
					disk.OAuth.RefreshToken != "" && disk.OAuth.RefreshToken != tok.RefreshToken {
					tok = *disk.OAuth
					// Disk's access_token may already be valid — use it
					// directly if so, otherwise refresh with the new RT.
					if time.Now().Before(tok.ExpiresAt.Add(-refreshLeeway)) {
						return tok.AccessToken, nil
					}
					refreshed, err = refreshAccessToken(ctx, &tok)
				}
			}
			if err != nil {
				return "", err
			}
		}
		tok = *refreshed
		cfg, _ := loadConfig()
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
		return tok.AccessToken, nil
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
	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil, fmt.Errorf("empty JSON input")
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return value, nil
}

// readJSONAs reads JSON and decodes into out.
func readJSONAs(path string, out any) error {
	value, err := readJSON(path)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
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
func addListFlags(cmd *cobra.Command, baseOnly bool) {
	cmd.Flags().String("before", "", "item id: return items before this id")
	cmd.Flags().String("after", "", "item id: return items after this id")
	cmd.Flags().Var(&boundedIntFlagValue{min: 0, max: 100}, "limit", "max items to return (1-100)")
	cmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
	if !baseOnly {
		cmd.Flags().String("filename", "", "filter by filename")
		cmd.Flags().Var(&rfc3339FlagValue{}, "from-date", "filter from this RFC3339 date")
		cmd.Flags().Var(&rfc3339FlagValue{}, "to-date", "filter to this RFC3339 date")
	}
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

type minIntFlagValue struct {
	value string
	min   int
}

func (v *minIntFlagValue) String() string {
	if v.value == "" {
		return "0"
	}
	return v.value
}

func (v *minIntFlagValue) Type() string { return "int" }

func (v *minIntFlagValue) Set(raw string) error {
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}
	if parsed < v.min {
		return fmt.Errorf("must be at least %d", v.min)
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

type rfc3339FlagValue struct{ value string }

func (v *rfc3339FlagValue) String() string { return v.value }

func (v *rfc3339FlagValue) Type() string { return "string" }

func (v *rfc3339FlagValue) Set(raw string) error {
	if raw != "" {
		if _, err := time.Parse(time.RFC3339, raw); err != nil {
			return fmt.Errorf("must be RFC3339 timestamp: %w", err)
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

func collectListParams(cmd *cobra.Command) retab.ListParams {
	params := retab.ListParams{}
	if v, _ := cmd.Flags().GetString("before"); v != "" {
		params.Before = v
	}
	if v, _ := cmd.Flags().GetString("after"); v != "" {
		params.After = v
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Limit = v
	}
	if v, _ := cmd.Flags().GetString("order"); v != "" {
		params.Order = v
	}
	if cmd.Flags().Lookup("filename") != nil {
		if v, _ := cmd.Flags().GetString("filename"); v != "" {
			params.Filename = v
		}
	}
	if cmd.Flags().Lookup("from-date") != nil {
		if v, _ := cmd.Flags().GetString("from-date"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				params.FromDate = &t
			}
		}
	}
	if cmd.Flags().Lookup("to-date") != nil {
		if v, _ := cmd.Flags().GetString("to-date"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				params.ToDate = &t
			}
		}
	}
	return params
}

func getIntFlagOrDefault(cmd *cobra.Command, name string, defaultValue int) int {
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
	link, err := client.Files.GetDownloadLink(ctx, fileID)
	if err != nil {
		return retab.MIMEData{}, fmt.Errorf("resolving --file-id %s: %w", fileID, err)
	}
	if link.DownloadURL == "" {
		return retab.MIMEData{}, fmt.Errorf("--file-id %s: server returned no download URL", fileID)
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
// Sensitive headers — `Api-Key`, `Authorization` — are redacted in the
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
		dumpReq.Body = io.NopCloser(bytes.NewReader(requestBody))
		dumpReq.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(requestBody)), nil
		}
		dumpReq.ContentLength = int64(len(requestBody))
	}
	if dump, err := httputil.DumpRequestOut(dumpReq, true); err == nil {
		fmt.Fprintf(os.Stderr, "\n--- HTTP request ---\n%s\n", dump)
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
	dumpResp := *resp
	dumpResp.Body = io.NopCloser(bytes.NewReader(body))
	dumpResp.ContentLength = int64(len(body))
	if dump, err := httputil.DumpResponse(&dumpResp, true); err == nil {
		fmt.Fprintf(os.Stderr, "--- HTTP response ---\n%s\n", dump)
	}
	return resp, nil
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
