package cmd

import (
	"encoding/json"
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
	t.Setenv("RETAB_BASE_URL", "")
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

func TestRedactKey(t *testing.T) {
	if got := redactKey("retab_sk_abcd1234"); !strings.HasPrefix(got, "reta") || !strings.HasSuffix(got, "1234") {
		t.Fatalf("got %q", got)
	}
	if got := redactKey("short"); got != "*****" {
		t.Fatalf("got %q", got)
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
