package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// runE wraps a command body so APIErrors render with full context
// (request id, method, url, status, code, body) and other errors render as a
// single line. Non-nil return propagates a non-zero exit through cobra.
func runE(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := fn(cmd, args)
		if err == nil {
			return nil
		}
		var apiErr *retab.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintln(os.Stderr, apiErr.String())
			return errSilent
		}
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		return errSilent
	}
}

// errSilent signals that the error was already rendered to stderr.
var errSilent = errors.New("")

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

// makeOAuthTokenProvider returns a closure that yields a current access
// token on demand, refreshing transparently as expiry approaches. A
// successful refresh is persisted to disk so subsequent CLI invocations
// pick up the new tokens. Save errors are swallowed: the in-memory token
// still completes the current request.
func makeOAuthTokenProvider(initial *oauthTokens) func(ctx context.Context) (string, error) {
	tok := *initial
	return func(ctx context.Context) (string, error) {
		if time.Now().Before(tok.ExpiresAt.Add(-refreshLeeway)) {
			return tok.AccessToken, nil
		}
		refreshed, err := refreshAccessToken(ctx, &tok)
		if err != nil {
			return "", err
		}
		tok = *refreshed
		cfg, _ := loadConfig()
		cfg.OAuth = &tok
		_ = saveConfig(cfg)
		return tok.AccessToken, nil
	}
}

func ctxFor(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	if ctx.Err() != nil {
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	}
	return ctx, cancel
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

// splitKV splits "name=value" into (name, value, true). When no '=' is present,
// it returns (raw, "", false). Used for repeatable k=v flags where the value
// half is optional (compare parseKVStringList, which is strict).
func splitKV(raw string) (string, string, bool) {
	return strings.Cut(raw, "=")
}

// addListFlags attaches the cursor pagination + filter flags shared by
// most list commands. baseOnly skips filename/from-date/to-date (which only
// apply to file-shaped resources).
func addListFlags(cmd *cobra.Command, baseOnly bool) {
	cmd.Flags().String("before", "", "cursor: items before this id")
	cmd.Flags().String("after", "", "cursor: items after this id")
	cmd.Flags().Int("limit", 0, "max items to return")
	cmd.Flags().String("order", "", "asc | desc")
	if !baseOnly {
		cmd.Flags().String("filename", "", "filter by filename")
		cmd.Flags().String("from-date", "", "filter from this RFC3339 date")
		cmd.Flags().String("to-date", "", "filter to this RFC3339 date")
	}
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

// addDocumentFlags adds the mutually-exclusive document source flags shared
// by every command that takes a document body.
func addDocumentFlags(cmd *cobra.Command) {
	cmd.Flags().String("file", "", "path to a local document")
	cmd.Flags().String("url", "", "https URL of a document")
	cmd.Flags().String("file-id", "", "Retab file id")
	cmd.Flags().String("document-file", "", "path to a JSON file describing the document (or - for stdin)")
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
		mime, err := retab.InferMIMEData(file)
		if err != nil {
			return nil, err
		}
		return mime, nil
	case urlStr != "":
		return retab.MIMEData{URL: urlStr}, nil
	case fileID != "":
		return retab.FileRef{ID: fileID}, nil
	case docFile != "":
		return readJSON(docFile)
	}
	return nil, fmt.Errorf("unreachable")
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
	return readJSON(path)
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
type debugTransport struct {
	wrapped http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if dump, err := httputil.DumpRequestOut(req, true); err == nil {
		fmt.Fprintf(os.Stderr, "\n--- HTTP request ---\n%s\n", dump)
	}
	resp, err := t.wrapped.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "--- HTTP error ---\n%v\n", err)
		return nil, err
	}
	if dump, err := httputil.DumpResponse(resp, true); err == nil {
		fmt.Fprintf(os.Stderr, "--- HTTP response ---\n%s\n", dump)
	}
	return resp, nil
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
	fmt.Fprintf(os.Stderr, "deleted %s: %s\n", kind, id)
}
