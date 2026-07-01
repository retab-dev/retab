package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// Regression: when the upstream server closes the connection mid-response
// (e.g. an uvicorn worker crashing on a publish), the bare “net/http“ error
// surface was “error: Post "...": EOF“ — actionless. “runE“ now translates
// any connection-drop signal into a clearer message that names the symptom.
func TestRenderConnectionDropForCLI(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "bare io.EOF", err: io.EOF, want: true},
		{name: "bare io.ErrUnexpectedEOF", err: io.ErrUnexpectedEOF, want: true},
		{name: "wrapped EOF (errors.Is path)", err: fmt.Errorf("transport: %w", io.EOF), want: true},
		{name: "net/http style suffix", err: errors.New(`Post "http://localhost:4000/v1/workflows/wf/publish": EOF`), want: true},
		{name: "unexpected EOF substring", err: errors.New(`read body: unexpected EOF`), want: true},
		{name: "unrelated error stays nil", err: errors.New("validation failed"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderConnectionDropForCLI(tc.err)
			if tc.want {
				if got == "" {
					t.Fatalf("expected a connection-drop message, got empty string")
				}
				if !strings.HasPrefix(got, "error: ") {
					t.Fatalf("message should start with 'error: ', got %q", got)
				}
				if !strings.Contains(got, "upstream server closed the connection") {
					t.Fatalf("message should name the symptom; got %q", got)
				}
				return
			}
			if got != "" {
				t.Fatalf("expected empty string for non-connection-drop error %v, got %q", tc.err, got)
			}
		})
	}
}

// TestValidateBaseURL pins the user-supplied base URL guard. Empty (use
// default) is allowed; http/https URLs with a host are allowed; everything
// else surfaces a CLI-shaped error before any HTTP call so users see the
// flag spelling instead of a net/http "unsupported protocol scheme" leak.
func TestValidateBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
		errLike string
	}{
		{name: "empty ok", in: "", wantErr: false},
		{name: "localhost http ok", in: "http://localhost:4000/v1", wantErr: false},
		{name: "https ok", in: "https://api.retab.com", wantErr: false},
		{name: "missing scheme", in: "not-a-url", wantErr: true, errLike: "missing scheme"},
		{name: "ftp scheme rejected", in: "ftp://x.com", wantErr: true, errLike: "not http or https"},
		{name: "scheme without host", in: "http://", wantErr: true, errLike: "missing host"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBaseURL(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tc.in)
				}
				if !strings.Contains(err.Error(), "--base-url") {
					t.Fatalf("error should mention --base-url, got: %v", err)
				}
				if tc.errLike != "" && !strings.Contains(err.Error(), tc.errLike) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.errLike)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.in, err)
			}
		})
	}
}

func TestParseKVStringList(t *testing.T) {
	cases := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr bool
	}{
		{name: "empty", input: nil, want: nil},
		{name: "single", input: []string{"a=1"}, want: map[string]string{"a": "1"}},
		{name: "multi", input: []string{"a=1", "b=2"}, want: map[string]string{"a": "1", "b": "2"}},
		{name: "empty value", input: []string{"a="}, want: map[string]string{"a": ""}},
		{name: "missing eq", input: []string{"a"}, wantErr: true},
		{name: "missing key", input: []string{"=1"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseKVStringList(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len=%d want=%d", len(got), len(tc.want))
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("key %q got %q want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestSplitKV(t *testing.T) {
	k, v, ok := splitKV("foo=bar")
	if !ok || k != "foo" || v != "bar" {
		t.Fatalf("got %q %q %v", k, v, ok)
	}
	k, v, ok = splitKV("foo")
	if ok || k != "foo" || v != "" {
		t.Fatalf("no-eq case wrong: %q %q %v", k, v, ok)
	}
	k, v, ok = splitKV("foo=bar=baz")
	if !ok || k != "foo" || v != "bar=baz" {
		t.Fatalf("multi-eq case wrong: %q %q %v", k, v, ok)
	}
}

func TestConfirmDeletedHonorsJSONOutputFlag(t *testing.T) {
	if err := rootCmd.PersistentFlags().Set("output", "json"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		confirmDeleted("parse", "prs_123")
	})
	if stderr != "" {
		t.Fatalf("expected no human stderr confirmation, got %q", stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("expected JSON stdout, got parse error %v for:\n%s", err, stdout)
	}
	if got["id"] != "prs_123" {
		t.Fatalf("id = %#v, want prs_123", got["id"])
	}
	if got["deleted"] != true {
		t.Fatalf("deleted = %#v, want true", got["deleted"])
	}
}

// Regression for CLI probing 2026-05: commands routed through
// `cliJSONRequest` (workflows runs restart, experiments runs *, tests
// runs *) used to surface the raw HTTP envelope on non-2xx — e.g.
// `error: GET http://… failed with status 404: {"detail":{"code":"HTTP_EXCEPTION",…}}`.
// SDK-backed commands rendered the same shape as `404 — Workflow not found`.
// cliJSONRequest now returns an *APIError so both paths render identically.
func TestCLIJSONRequestSurfacesAPIErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":{"code":"HTTP_EXCEPTION","message":"An HTTP exception occurred.","details":{"error":"Experiment run not found: exprun_404"}}}`))
	}))
	defer server.Close()

	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "fake"}
	rootCmd.AddCommand(cmd)
	t.Cleanup(func() { rootCmd.RemoveCommand(cmd) })

	_, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/experiments/runs/exprun_404", nil, nil)
	if err == nil {
		t.Fatal("expected an error on 404")
	}
	apiErr, ok := err.(*retab.APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Error(), "Experiment run not found") {
		t.Fatalf("APIError did not surface the server message: %s", apiErr.Error())
	}
	// The pre-fix shape was a bare wrapped string containing "failed with status".
	if strings.Contains(apiErr.Error(), "failed with status") {
		t.Fatalf("APIError message still resembles the raw HTTP envelope: %s", apiErr.Error())
	}
}

func TestRenderAPIErrorForCLIDefaultHidesRawHTTPDebug(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, false)
	apiErr := &retab.APIError{
		StatusCode: http.StatusBadRequest,
		Code:       "HTTP_EXCEPTION",
		Message:    "An HTTP exception occurred.",
		Method:     http.MethodPost,
		URL:        "http://localhost:4000/v1/workflows/blocks?workflow_id=wf_1",
		Details: map[string]any{
			"error": "Invalid config for for_each block. Problem: review is only available for map_method='split_by_key'.",
		},
		Body: `{"detail":{"code":"HTTP_EXCEPTION","message":"An HTTP exception occurred.","details":{"error":"Invalid config for for_each block."}}}`,
	}

	got := renderAPIErrorForCLI(cmd, apiErr)

	if !strings.Contains(got, "Invalid config for for_each block") {
		t.Fatalf("rendered error should surface the useful message:\n%s", got)
	}
	for _, raw := range []string{"URL:", "Code:", "Details:", "Body:", "HTTP_EXCEPTION", "/workflows/blocks?workflow_id=wf_1"} {
		if strings.Contains(got, raw) {
			t.Fatalf("default CLI error should hide raw HTTP detail %q:\n%s", raw, got)
		}
	}
}

func TestRenderAPIErrorForCLISurfacesNestedErrorMessage(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, false)
	apiErr := &retab.APIError{
		StatusCode: http.StatusConflict,
		Code:       "HTTP_EXCEPTION",
		Message:    "An HTTP exception occurred.",
		Method:     http.MethodPost,
		URL:        "http://localhost:4000/v1/workflows/runs",
		Details: map[string]any{
			"error": map[string]any{
				"code":    "production_required_but_unpublished",
				"message": "Workflow wf_1 has no published version. Publish it, or pass version='draft' to run the current draft.",
			},
		},
		Body: `{"detail":{"code":"HTTP_EXCEPTION","message":"An HTTP exception occurred.","details":{"error":{"message":"Workflow wf_1 has no published version."}}}}`,
	}

	got := renderAPIErrorForCLI(cmd, apiErr)

	if !strings.Contains(got, "has no published version") || !strings.Contains(got, "version='draft'") {
		t.Fatalf("rendered error should surface nested server message:\n%s", got)
	}
	for _, raw := range []string{"URL:", "Details:", "Body:", "HTTP_EXCEPTION", "production_required_but_unpublished"} {
		if strings.Contains(got, raw) {
			t.Fatalf("default CLI error should hide raw HTTP detail %q:\n%s", raw, got)
		}
	}
}

func TestRenderAPIErrorForCLIDebugKeepsRawHTTPDebug(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, true)
	apiErr := &retab.APIError{
		StatusCode: http.StatusBadRequest,
		Code:       "HTTP_EXCEPTION",
		Message:    "An HTTP exception occurred.",
		Method:     http.MethodPost,
		URL:        "http://localhost:4000/v1/workflows/blocks?workflow_id=wf_1",
		Details:    map[string]any{"error": "Invalid config for for_each block."},
		Body:       `{"detail":{"code":"HTTP_EXCEPTION"}}`,
	}

	got := renderAPIErrorForCLI(cmd, apiErr)

	for _, raw := range []string{"URL:", "Code:", "Details:", "Body:", "HTTP_EXCEPTION", "/workflows/blocks?workflow_id=wf_1"} {
		if !strings.Contains(got, raw) {
			t.Fatalf("--debug CLI error should include raw HTTP detail %q:\n%s", raw, got)
		}
	}
}

func TestRenderAPIErrorForCLIFormatsFlatValidationEnvelope(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, false)
	apiErr := &retab.APIError{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    `[{"type":"missing","loc":["body","blocks",0,"label"],"msg":"Field required"},{"type":"missing","loc":["body","edges",1,"source_block"],"msg":"Field required"}]`,
		Method:     http.MethodPost,
		URL:        "http://localhost:4000/v1/workflows/blocks",
		Body:       `{"status_code":10422,"message":"[...]","data":null}`,
	}

	got := renderAPIErrorForCLI(cmd, apiErr)

	for _, want := range []string{
		"422 — Invalid request.",
		"blocks[0].label: Field required",
		"edges[1].source_block: Field required",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered validation error should contain %q:\n%s", want, got)
		}
	}
	for _, raw := range []string{"\"loc\"", "\"type\"", "Body:", "URL:"} {
		if strings.Contains(got, raw) {
			t.Fatalf("default validation error should hide raw detail %q:\n%s", raw, got)
		}
	}
}

// Regression: the Workflow Evals create call hits a server that routes
// /v1/workflows/evals but does not register POST there yet (405, Allow omits
// POST). The raw envelope rendered "405 — 405 method not allowed", which reads
// like a broken CLI. The renderer now explains it's a client-ahead-of-server
// rollout and names the endpoint without leaking the host or query string.
func TestRenderAPIErrorForCLIExplains405AsNotEnabled(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, false)
	apiErr := &retab.APIError{
		StatusCode: http.StatusMethodNotAllowed,
		Message:    "405 method not allowed",
		Method:     http.MethodPost,
		URL:        "https://api.retab.com/v1/workflows/evals?workflow_id=wrk_1",
		RequestID:  "req_123",
		Body:       "405 method not allowed",
	}

	got := renderAPIErrorForCLI(cmd, apiErr)

	for _, want := range []string{
		"405 — This operation is not available on this environment yet.",
		"POST /v1/workflows/evals",
		"not be enabled on this environment yet",
		"Request-ID: req_123",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered 405 should contain %q:\n%s", want, got)
		}
	}
	// The query string and host are raw HTTP detail — keep them hidden, like the
	// other non-debug renders. The redundant raw body must not leak either.
	for _, raw := range []string{"workflow_id=wrk_1", "api.retab.com", "405 method not allowed"} {
		if strings.Contains(got, raw) {
			t.Fatalf("non-debug 405 render should hide %q:\n%s", raw, got)
		}
	}
}

// --debug must still win: a 405 dumps the full wire envelope, not the friendly
// summary, so operators can inspect the Allow header and request URL.
func TestRenderAPIErrorForCLI405DebugKeepsRawEnvelope(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, true)
	apiErr := &retab.APIError{
		StatusCode: http.StatusMethodNotAllowed,
		Message:    "405 method not allowed",
		Method:     http.MethodPost,
		URL:        "https://api.retab.com/v1/workflows/evals",
		Body:       "405 method not allowed",
	}

	got := renderAPIErrorForCLI(cmd, apiErr)

	if strings.Contains(got, "not available on this environment") {
		t.Fatalf("--debug 405 should show the raw envelope, not the summary:\n%s", got)
	}
	for _, want := range []string{"URL:", "/v1/workflows/evals", "Body:"} {
		if !strings.Contains(got, want) {
			t.Fatalf("--debug 405 render should include raw HTTP detail %q:\n%s", want, got)
		}
	}
}

// A 404 "Workflow not found" is intentionally left as-is: a genuinely mistyped
// workflow id produces the same status, so it must not be reworded as a
// not-enabled rollout gap.
func TestRenderAPIErrorForCLI404NotRewordedAsNotEnabled(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, false)
	apiErr := &retab.APIError{
		StatusCode: http.StatusNotFound,
		Message:    "Workflow not found",
		Method:     http.MethodGet,
		URL:        "https://api.retab.com/v1/workflows/evals?workflow_id=wrk_1",
	}

	got := renderAPIErrorForCLI(cmd, apiErr)

	if !strings.Contains(got, "404 — Workflow not found") {
		t.Fatalf("404 should render its server message verbatim:\n%s", got)
	}
	if strings.Contains(got, "not available on this environment") {
		t.Fatalf("404 must not be reworded as a not-enabled rollout gap:\n%s", got)
	}
}

func commandWithDebugFlagForTest(t *testing.T, debug bool) *cobra.Command {
	t.Helper()
	root := &cobra.Command{Use: "retab"}
	root.PersistentFlags().Bool("debug", false, "")
	cmd := &cobra.Command{Use: "test"}
	root.AddCommand(cmd)
	if debug {
		if err := root.PersistentFlags().Set("debug", "true"); err != nil {
			t.Fatal(err)
		}
	}
	return cmd
}

func TestReadJSONMapAndArray(t *testing.T) {
	dir := t.TempDir()
	mapPath := filepath.Join(dir, "obj.json")
	arrPath := filepath.Join(dir, "arr.json")
	if err := os.WriteFile(mapPath, []byte(`{"x":1,"y":"two"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(arrPath, []byte(`[1,2,3]`), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := readJSONMap(mapPath)
	if err != nil {
		t.Fatalf("readJSONMap: %v", err)
	}
	if m["x"].(float64) != 1 || m["y"].(string) != "two" {
		t.Fatalf("got %#v", m)
	}
	arr, err := readJSONArray(arrPath)
	if err != nil {
		t.Fatalf("readJSONArray: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("len=%d", len(arr))
	}
	// Wrong shape → wrong helper.
	if _, err := readJSONMap(arrPath); err == nil {
		t.Fatalf("expected error for array → readJSONMap")
	}
	if _, err := readJSONArray(mapPath); err == nil {
		t.Fatalf("expected error for object → readJSONArray")
	}

	inline, err := parseJSONMap(`{"inline":true}`)
	if err != nil {
		t.Fatalf("parseJSONMap: %v", err)
	}
	if inline["inline"] != true {
		t.Fatalf("inline = %#v", inline)
	}
	if _, err := parseJSONMap(`[1,2,3]`); err == nil {
		t.Fatalf("expected error for array → parseJSONMap")
	}
}

func TestReadInputNormalizesBOMAndUTF16(t *testing.T) {
	dir := t.TempDir()

	// UTF-8 BOM + JSON object (as written by PowerShell `Out-File -Encoding utf8`).
	u8 := filepath.Join(dir, "u8bom.json")
	if err := os.WriteFile(u8, append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"x":1}`)...), 0o600); err != nil {
		t.Fatal(err)
	}
	if m, err := readJSONMap(u8); err != nil || m["x"].(float64) != 1 {
		t.Fatalf("utf-8 BOM: m=%#v err=%v", m, err)
	}

	// UTF-16 LE BOM + JSON object (as written by PowerShell's default `>` redirect).
	u16 := []byte{0xFF, 0xFE}
	for _, c := range []byte(`{"x":1}`) {
		u16 = append(u16, c, 0x00)
	}
	u16Path := filepath.Join(dir, "u16le.json")
	if err := os.WriteFile(u16Path, u16, 0o600); err != nil {
		t.Fatal(err)
	}
	if m, err := readJSONMap(u16Path); err != nil || m["x"].(float64) != 1 {
		t.Fatalf("utf-16 LE: m=%#v err=%v", m, err)
	}

	// The free-text reader strips the BOM too (e.g. --instructions-file).
	txt := filepath.Join(dir, "instr.txt")
	if err := os.WriteFile(txt, append([]byte{0xEF, 0xBB, 0xBF}, []byte("be concise\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := readTextFileOrStdin(txt); err != nil || got != "be concise" {
		t.Fatalf("text BOM: got=%q err=%v", got, err)
	}
}

func TestResolveDocumentURL(t *testing.T) {
	cmd := &cobra.Command{}
	addDocumentFlags(cmd)
	if err := cmd.ParseFlags([]string{"--url", "https://example.com/doc.pdf"}); err != nil {
		t.Fatal(err)
	}
	doc, err := resolveDocument(cmd)
	if err != nil {
		t.Fatal(err)
	}
	mime, ok := doc.(retab.MIMEData)
	if !ok {
		t.Fatalf("got %T", doc)
	}
	if mime.URL != "https://example.com/doc.pdf" {
		t.Fatalf("url=%q", mime.URL)
	}
	// The server requires `filename` on every document descriptor — pin
	// the rule that --url derives one from the URL path's last segment.
	if mime.Filename != "doc.pdf" {
		t.Fatalf("filename=%q, want %q", mime.Filename, "doc.pdf")
	}
}

// filenameFromURL is the offline core of the --url → filename derivation,
// table-driven so edge cases (no path, query string, root, etc.) are all
// pinned without paying for a real HTTP roundtrip.
func TestFilenameFromURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://example.com/doc.pdf", "doc.pdf"},
		{"https://example.com/nested/path/invoice.PDF", "invoice.PDF"},
		{"https://example.com/", "document"},
		{"https://example.com", "document"},
		{"https://example.com/path/", "path"},
		{"https://example.com/file.pdf?signed=1&token=abc", "file.pdf"},
		{"not-a-url", "not-a-url"}, // url.Parse accepts this; path.Base returns the literal
		{"", "document"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := filenameFromURL(tc.in); got != tc.want {
				t.Errorf("filenameFromURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// resolveDocument with --file-id now triggers a server lookup to fetch
// the file's filename + a fresh download URL (the SDK's `FileRef{ID: ...}`
// shape is rejected by the server post-API-change). Without RETAB_API_KEY
// set, the lookup fails fast on missing credentials — the failure mode
// we pin here. The happy path requires a live API and is exercised via
// dogfooding rather than as a unit test.
func TestResolveDocumentFileID_RequiresCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	home := t.TempDir() // ensure no ~/.retab/config.json bleeds through
	t.Setenv("HOME", home)
	// configDir resolves via os.UserHomeDir(), which is %USERPROFILE% on
	// Windows — isolate it too or a sibling test's leaked config.json (api_key
	// + base_url) bleeds through and the call reaches the network.
	t.Setenv("USERPROFILE", home)

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("api-key", "", "")
	cmd.PersistentFlags().String("base-url", "", "")
	addDocumentFlags(cmd)
	if err := cmd.ParseFlags([]string{"--file-id", "file_123"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveDocument(cmd)
	if err == nil {
		t.Fatal("expected error without credentials, got nil")
	}
	if !strings.Contains(err.Error(), "credentials") && !strings.Contains(err.Error(), "auth") {
		t.Errorf("error should mention credentials/auth, got: %v", err)
	}
}

func TestResolveDocumentFileIDUsesDurableMIMEData(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"download_url": "https://storage.googleapis.com/bucket/org_1/file/file_123.pdf?signed=1",
			"expires_in": "60 minutes",
			"filename": "invoice.pdf",
			"mime_data": {
				"filename": "invoice.pdf",
				"url": "https://storage.retab.com/org_1/file_123.pdf"
			}
		}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("api-key", "", "")
	cmd.PersistentFlags().String("base-url", "", "")
	addDocumentFlags(cmd)
	if err := cmd.ParseFlags([]string{"--file-id", "file_123"}); err != nil {
		t.Fatal(err)
	}

	doc, err := resolveDocument(cmd)
	if err != nil {
		t.Fatal(err)
	}
	mime, ok := doc.(retab.MIMEData)
	if !ok {
		t.Fatalf("got %T", doc)
	}
	if seenPath != "/v1/files/file_123/download-link" {
		t.Fatalf("path = %q", seenPath)
	}
	if mime.Filename != "invoice.pdf" {
		t.Fatalf("filename = %q", mime.Filename)
	}
	if mime.URL != "https://storage.retab.com/org_1/file_123.pdf" {
		t.Fatalf("url = %q", mime.URL)
	}
	if mime.ID() != "file_123" {
		t.Fatalf("id = %q", mime.ID())
	}
}

// A bad --file path used to bubble up as "retab: unsupported MIME input
// string" because the path was handed straight to retab.InferMIMEData,
// which fails the same way a binary blob with no detectable mime would.
// Stat the path upfront and surface a clear "file not found:" error so
// users can spot the typo without thinking about MIME machinery.
func TestResolveDocumentFileMissing(t *testing.T) {
	cmd := &cobra.Command{}
	addDocumentFlags(cmd)
	missing := "/definitely/does/not/exist.pdf"
	if err := cmd.ParseFlags([]string{"--file", missing}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveDocument(cmd)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	want := "file not found: " + missing
	if !strings.HasPrefix(err.Error(), "file not found: ") {
		t.Fatalf("error should start with %q, got: %v", "file not found: ", err)
	}
	if err.Error() != want {
		t.Errorf("error mismatch:\n got: %q\nwant: %q", err.Error(), want)
	}
}

// Same treatment for --document-file — a missing JSON descriptor used to
// surface as a json-unmarshal error wrapping "no such file or directory",
// which is confusing for a typo'd path.
func TestResolveDocumentDocumentFileMissing(t *testing.T) {
	cmd := &cobra.Command{}
	addDocumentFlags(cmd)
	missing := "/definitely/does/not/exist.json"
	if err := cmd.ParseFlags([]string{"--document-file", missing}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveDocument(cmd)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.HasPrefix(err.Error(), "file not found: ") {
		t.Fatalf("error should start with %q, got: %v", "file not found: ", err)
	}
}

// inferFileMIMEData is the shared guard behind every command that takes a
// local-file document flag (resolveDocument, schemas generate, workflows
// runs create --document). A missing path must surface as "file not
// found:" rather than the SDK's cryptic "unsupported MIME input string".
func TestInferFileMIMEDataMissing(t *testing.T) {
	missing := "/definitely/does/not/exist.pdf"
	_, err := inferFileMIMEData(missing)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	want := "file not found: " + missing
	if err.Error() != want {
		t.Fatalf("error mismatch:\n got: %q\nwant: %q", err.Error(), want)
	}
}

// A directory passed to --file used to slip past inferFileMIMEData's
// os.Stat guard (stat succeeds on a directory) and fall through to the
// SDK's MIME sniffing, resurfacing as exactly the cryptic "unsupported
// MIME input string" the guard exists to prevent. The guard must reject
// directories with a clear error too.
func TestInferFileMIMEDataDirectory(t *testing.T) {
	dir := t.TempDir()
	_, err := inferFileMIMEData(dir)
	if err == nil {
		t.Fatal("expected error for directory path, got nil")
	}
	if strings.Contains(err.Error(), "unsupported MIME input string") {
		t.Fatalf("directory error leaked the cryptic SDK message: %v", err)
	}
	want := "not a file (is a directory): " + dir
	if err.Error() != want {
		t.Fatalf("error mismatch:\n got: %q\nwant: %q", err.Error(), want)
	}
}

// resolveDocument with --file pointing at a directory must surface the
// same clear error rather than the cryptic MIME message.
func TestResolveDocumentFileIsDirectory(t *testing.T) {
	cmd := &cobra.Command{}
	addDocumentFlags(cmd)
	dir := t.TempDir()
	if err := cmd.ParseFlags([]string{"--file", dir}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveDocument(cmd)
	if err == nil {
		t.Fatal("expected error for directory path, got nil")
	}
	if !strings.HasPrefix(err.Error(), "not a file (is a directory): ") {
		t.Fatalf("error should start with %q, got: %v", "not a file (is a directory): ", err)
	}
}

func TestResolveDocumentMutex(t *testing.T) {
	cmd := &cobra.Command{}
	addDocumentFlags(cmd)
	if err := cmd.ParseFlags([]string{"--url", "x", "--file-id", "y"}); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveDocument(cmd); err == nil {
		t.Fatalf("expected mutex error")
	}
}

func TestResolveDocumentNoneRequired(t *testing.T) {
	cmd := &cobra.Command{}
	addDocumentFlags(cmd)
	if _, err := resolveDocument(cmd); err == nil {
		t.Fatalf("expected required error")
	}
}

func TestResolveDocumentNoneOptional(t *testing.T) {
	cmd := &cobra.Command{}
	addDocumentFlags(cmd)
	doc, err := resolveOptionalDocument(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if doc != nil {
		t.Fatalf("expected nil document, got %T", doc)
	}
}

func TestResolveSchemaLiteral(t *testing.T) {
	cmd := &cobra.Command{}
	addSchemaFlags(cmd)
	if err := cmd.ParseFlags([]string{"--json-schema", `{"type":"object"}`}); err != nil {
		t.Fatal(err)
	}
	got, err := resolveSchema(cmd)
	if err != nil {
		t.Fatal(err)
	}
	obj, ok := got.(map[string]any)
	if !ok || obj["type"] != "object" {
		t.Fatalf("got %#v", got)
	}
}

func TestResolveSchemaFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.json")
	if err := os.WriteFile(path, []byte(`{"type":"number"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	addSchemaFlags(cmd)
	if err := cmd.ParseFlags([]string{"--json-schema-file", path}); err != nil {
		t.Fatal(err)
	}
	got, err := resolveSchema(cmd)
	if err != nil {
		t.Fatal(err)
	}
	obj := got.(map[string]any)
	if obj["type"] != "number" {
		t.Fatalf("got %#v", obj)
	}
}

func TestCollectListParams(t *testing.T) {
	cmd := &cobra.Command{}
	addListFlags(cmd, false)
	if err := cmd.ParseFlags([]string{
		"--before", "b1",
		"--after", "a1",
		"--limit", "42",
		"--order", "asc",
		"--filename", "doc.pdf",
		"--from-date", "2024-01-02T03:04:05Z",
	}); err != nil {
		t.Fatal(err)
	}
	got := collectListParams(cmd)
	if got.Before == nil || *got.Before != "b1" || got.After == nil || *got.After != "a1" ||
		got.Limit == nil || *got.Limit != 42 || got.Order == nil || *got.Order != "asc" {
		t.Fatalf("got=%+v", got)
	}
}

func TestAddListFlagsRejectsInvalidDate(t *testing.T) {
	cmd := &cobra.Command{}
	addListFlags(cmd, false)

	err := cmd.ParseFlags([]string{"--from-date", "not-a-date"})
	if err == nil {
		t.Fatal("expected parse error for invalid --from-date, got nil")
	}
	if !strings.Contains(err.Error(), "RFC3339") {
		t.Fatalf("error should mention RFC3339, got: %v", err)
	}
}

// TestAddListFlagsAcceptsBareDate guards a staging-dogfood bug: the file-shaped
// list filters (files/parses/classifications/...) rejected the natural
// `--from-date 2026-06-13`, demanding a full RFC3339 timestamp — even though
// the backend's StartOfDayUTC/EndOfDayUTC (parseISO) accepts a bare YYYY-MM-DD
// and pins it to the start/end of the UTC day. Both forms must parse.
func TestAddListFlagsAcceptsBareDate(t *testing.T) {
	for _, val := range []string{"2026-06-13", "2026-06-13T00:00:00Z"} {
		cmd := &cobra.Command{}
		addListFlags(cmd, false)
		if err := cmd.ParseFlags([]string{"--from-date", val, "--to-date", val}); err != nil {
			t.Fatalf("--from-date/--to-date %q should be accepted, got: %v", val, err)
		}
	}
}

func TestAddListFlagsRejectsNegativeLimit(t *testing.T) {
	cmd := &cobra.Command{}
	addListFlags(cmd, false)

	err := cmd.ParseFlags([]string{"--limit", "-1"})
	if err == nil {
		t.Fatal("expected parse error for negative --limit, got nil")
	}
	if !strings.Contains(err.Error(), "between 1 and 100") {
		t.Fatalf("error should mention backend limit range, got: %v", err)
	}
}

func TestAddListFlagsRejectsInvalidOrder(t *testing.T) {
	cmd := &cobra.Command{}
	addListFlags(cmd, false)

	err := cmd.ParseFlags([]string{"--order", "sideways"})
	if err == nil {
		t.Fatal("expected parse error for invalid --order, got nil")
	}
	if !strings.Contains(err.Error(), "asc") || !strings.Contains(err.Error(), "desc") {
		t.Fatalf("error should mention asc/desc, got: %v", err)
	}
}

func TestRedactKey(t *testing.T) {
	if got := redactKey("retab_sk_abcd1234"); !strings.HasPrefix(got, "reta") || !strings.HasSuffix(got, "1234") {
		t.Fatalf("got %q", got)
	}
	if got := redactKey("short"); got != "*****" {
		t.Fatalf("got %q", got)
	}
}

// TestRedactKeyMaskIsFixedWidth pins that a long credential (e.g. an OAuth
// JWT, ~1000 chars) is masked with a short, bounded asterisk run rather
// than one asterisk per hidden character. The old len-8 mask flooded
// --debug output with a screenful of asterisks and reproduced the exact
// length of the secret.
func TestRedactKeyMaskIsFixedWidth(t *testing.T) {
	short := redactKey("retab_sk_abcd1234")                       // 17 chars
	long := redactKey("eyJ" + strings.Repeat("x", 1000) + "_abc") // ~1006 chars

	if !strings.HasPrefix(long, "eyJx") || !strings.HasSuffix(long, "_abc") {
		t.Fatalf("long key lost its prefix/suffix preview: %q", long)
	}
	if len(long) != len(short) {
		t.Fatalf("redacted length leaks the secret size: short=%d long=%d", len(short), len(long))
	}
	if stars := strings.Count(long, "*"); stars > 8 {
		t.Fatalf("mask width not capped: %d asterisks", stars)
	}
}

// TestCommandTreeShape walks the registered commands and checks every leaf
// has either RunE or Run, and that no two siblings share a name.
func TestCommandTreeShape(t *testing.T) {
	var walk func(c *cobra.Command, path string)
	walk = func(c *cobra.Command, path string) {
		names := map[string]bool{}
		for _, child := range c.Commands() {
			name := child.Name()
			if names[name] {
				t.Errorf("duplicate child name %q under %s", name, path)
			}
			names[name] = true
			if len(child.Commands()) == 0 {
				// Leaf — must be invokable, UNLESS it's a Hidden command
				// (no help-topic system currently registers any, but we
				// keep the escape hatch for future opt-in help-only nodes).
				if child.RunE == nil && child.Run == nil && !child.Hidden {
					// Built-in cobra commands (help, completion) are fine.
					if !isBuiltin(child.Name()) {
						t.Errorf("leaf command %s/%s has no Run/RunE", path, name)
					}
				}
			}
			walk(child, path+"/"+name)
		}
	}
	walk(rootCmd, "")
}

func isBuiltin(name string) bool {
	switch name {
	case "help", "completion":
		return true
	default:
		return false
	}
}

// TestConfigRoundTrip exercises save/load/delete against a temp HOME.
func TestConfigRoundTrip(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	// On some platforms USERPROFILE is also consulted; align them.
	t.Setenv("USERPROFILE", tmpHome)

	cfg := retabConfig{APIKey: "retab_sk_test", BaseURL: "https://api.test/v1"}
	if err := saveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	got, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	// saveConfig stamps the current schema version on every write.
	cfg.Version = configVersion
	if !reflect.DeepEqual(got, cfg) {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, cfg)
	}
	if err := deleteConfig(); err != nil {
		t.Fatal(err)
	}
	got, err = loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, retabConfig{}) {
		t.Fatalf("expected empty after delete, got %+v", got)
	}
}

// TestPrintJSONRoundTrip — light sanity check that printJSON output is valid
// JSON. Captures via a pipe.
func TestPrintJSONRoundTrip(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	go func() {
		_ = printJSON(map[string]any{"hello": "world"})
		_ = w.Close()
	}()
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	var got map[string]any
	if err := json.Unmarshal(buf[:n], &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, string(buf[:n]))
	}
	if got["hello"] != "world" {
		t.Fatalf("got=%v", got)
	}
}

// confirmDestructive must accept --confirm as also satisfying the destructive
// gate. A production-env delete is already gated by productionGate's --confirm;
// demanding an ADDITIONAL --yes made delete the only production-mutating command
// needing two ack flags. --confirm is a strictly explicit acknowledgment, so it
// unifies the two gates: a production delete needs only --confirm (like publish/
// runs create), while --yes still works on its own for non-production deletes.
func TestConfirmDestructiveAcceptsConfirmFlag(t *testing.T) {
	newCmd := func() *cobra.Command {
		c := &cobra.Command{Use: "delete"}
		c.Flags().BoolP("yes", "y", false, "")
		addConfirmFlag(c)
		c.SetIn(strings.NewReader("")) // non-terminal stdin → no interactive prompt
		return c
	}

	// Neither flag, non-interactive → refuses.
	if err := confirmDestructive(newCmd(), "workflow", "wrk_1"); err == nil {
		t.Fatal("expected refusal without --yes/--confirm in non-interactive mode")
	}

	// --yes alone → passes (existing behavior, unchanged).
	cy := newCmd()
	if err := cy.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	if err := confirmDestructive(cy, "workflow", "wrk_1"); err != nil {
		t.Fatalf("--yes should satisfy the destructive gate: %v", err)
	}

	// --confirm alone → passes (new unified behavior).
	cc := newCmd()
	if err := cc.Flags().Set("confirm", "true"); err != nil {
		t.Fatal(err)
	}
	if err := confirmDestructive(cc, "workflow", "wrk_1"); err != nil {
		t.Fatalf("--confirm should satisfy the destructive gate: %v", err)
	}
}

func TestScopedResourceID(t *testing.T) {
	// Single arg: returned as-is.
	if got, err := scopedResourceID([]string{"run_abc"}, "run id"); err != nil || got != "run_abc" {
		t.Fatalf("single arg = %q, %v; want run_abc, nil", got, err)
	}
	if _, err := scopedResourceID([]string{""}, "run id"); err == nil {
		t.Fatal("empty single arg should error")
	}
	// Optional leading workflow id is tolerated and dropped.
	if got, err := scopedResourceID([]string{"wrk_123", "run_abc"}, "run id"); err != nil || got != "run_abc" {
		t.Fatalf("workflow-prefixed = %q, %v; want run_abc, nil", got, err)
	}
	if _, err := scopedResourceID([]string{"wrk_123", ""}, "run id"); err == nil {
		t.Fatal("empty scoped resource id should error")
	}
	// A second arg that is NOT a workflow id is a real mistake → error.
	if _, err := scopedResourceID([]string{"run_abc", "run_def"}, "run id"); err == nil {
		t.Fatal("two non-workflow args should error")
	}
}

func TestDocumentPathHint(t *testing.T) {
	cases := []struct {
		path string
		want string // substring expected in the hint ("" = no hint)
	}{
		{"file_abc123", "--document-id"},
		{"--file-id file_abc123", "--document-id"},
		{"https://example.com/x.pdf", "--document-url"},
		{"./invoice.pdf", ""},
		{"invoice.pdf", ""},
	}
	for _, c := range cases {
		got := documentPathHint("start", c.path)
		if c.want == "" {
			if got != "" {
				t.Fatalf("documentPathHint(%q) = %q; want no hint", c.path, got)
			}
			continue
		}
		if !strings.Contains(got, c.want) {
			t.Fatalf("documentPathHint(%q) = %q; want substring %q", c.path, got, c.want)
		}
	}
}

func TestMaterializeInlineMIMEData(t *testing.T) {
	ctx := context.Background()
	// Already-inline data: documents pass through untouched.
	inline := retab.MIMEData{Filename: "a.pdf", URL: "data:application/pdf;base64,JVBERg=="}
	if got, err := materializeInlineMIMEData(ctx, inline); err != nil || got.URL != inline.URL {
		t.Fatalf("inline passthrough = %q, %v; want unchanged", got.URL, err)
	}
	// Content-only (no URL) descriptors pass through.
	if got, err := materializeInlineMIMEData(ctx, retab.MIMEData{Filename: "a.pdf"}); err != nil || got.URL != "" {
		t.Fatalf("content-only passthrough = %q, %v", got.URL, err)
	}
	// A remote URL document is downloaded and re-inlined as a data: URL.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("%PDF-1.4\n%%EOF\n"))
	}))
	defer srv.Close()
	got, err := materializeInlineMIMEData(ctx, retab.MIMEData{Filename: "invoice.pdf", URL: srv.URL + "/file.pdf"})
	if err != nil {
		t.Fatalf("remote materialize error: %v", err)
	}
	if !strings.HasPrefix(got.URL, "data:application/pdf;base64,") {
		t.Fatalf("materialized URL = %q; want inline data:application/pdf", got.URL)
	}
	if got.Filename != "invoice.pdf" {
		t.Fatalf("filename = %q; want invoice.pdf preserved", got.Filename)
	}
}
