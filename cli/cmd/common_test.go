package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

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
		URL:        "http://localhost:4000/v1/workflows/wf_1/diagnose-graph",
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
	t.Setenv("HOME", t.TempDir()) // ensure no ~/.retab/config.json bleeds through

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
	if seenPath != "/files/file_123/download-link" {
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
	if got.Before != "b1" || got.After != "a1" || got.Limit != 42 || got.Order != "asc" || got.Filename != "doc.pdf" {
		t.Fatalf("got=%+v", got)
	}
	if got.FromDate == nil || got.FromDate.Year() != 2024 {
		t.Fatalf("from_date=%v", got.FromDate)
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

func TestAddListFlagsRejectsNegativeLimit(t *testing.T) {
	cmd := &cobra.Command{}
	addListFlags(cmd, false)

	err := cmd.ParseFlags([]string{"--limit", "-1"})
	if err == nil {
		t.Fatal("expected parse error for negative --limit, got nil")
	}
	if !strings.Contains(err.Error(), "between 0 and 100") {
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
	if got != cfg {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, cfg)
	}
	if err := deleteConfig(); err != nil {
		t.Fatal(err)
	}
	got, err = loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if (got != retabConfig{}) {
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
		w.Close()
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
