package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeLocalDoc writes a small text doc and returns its path.
func writeLocalDoc(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "n.txt")
	if err := os.WriteFile(path, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestFilesParseHonorsExplicitOutputJSON pins that `files parse --output json`
// emits the structured JSON view, matching its siblings `files grep`/`files
// inspect`. An explicit global --output json must win over the text default.
func TestFilesParseHonorsExplicitOutputJSON(t *testing.T) {
	path := writeLocalDoc(t)
	resetParseFlags(t)
	if err := rootCmd.PersistentFlags().Set("output", "json"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, _ := captureStd(t, func() {
		if err := filesParseCmd.RunE(filesParseCmd, []string{path}); err != nil {
			t.Fatalf("parse: %v", err)
		}
	})
	if !strings.Contains(stdout, `"document_type"`) {
		t.Fatalf("expected JSON output with --output json, got %q", stdout)
	}
}

// TestFilesParseTextDefaultUnaffectedByAutoOutput guards the fix's edge: with no
// explicit --output, parse must stay text even though auto-detect resolves to
// json on a non-TTY (the captured test buffer). Only an EXPLICIT --output json
// flips it.
func TestFilesParseTextDefaultUnaffectedByAutoOutput(t *testing.T) {
	path := writeLocalDoc(t)
	resetParseFlags(t)
	stdout, _ := captureStd(t, func() {
		if err := filesParseCmd.RunE(filesParseCmd, []string{path}); err != nil {
			t.Fatalf("parse: %v", err)
		}
	})
	if stdout != "hello\nworld\n" {
		t.Fatalf("parse default output = %q, want plain text", stdout)
	}
}

// TestFilesParseFileFlag pins that --file supplies the path when no positional
// is given (parity with the document primitives' --file).
func TestFilesParseFileFlag(t *testing.T) {
	path := writeLocalDoc(t)
	resetParseFlags(t)
	if err := filesParseCmd.Flags().Set("file", path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filesParseCmd.Flags().Set("file", "") })

	stdout, _ := captureStd(t, func() {
		if err := filesParseCmd.RunE(filesParseCmd, []string{}); err != nil {
			t.Fatalf("parse: %v", err)
		}
	})
	if stdout != "hello\nworld\n" {
		t.Fatalf("parse --file output = %q", stdout)
	}
}

// TestFilesParseRequiresPath pins the error when neither a positional path nor
// --file is provided.
func TestFilesParseRequiresPath(t *testing.T) {
	resetParseFlags(t)
	_ = filesParseCmd.Flags().Set("file", "")
	if err := filesParseCmd.RunE(filesParseCmd, []string{}); err == nil {
		t.Fatal("expected an error when no path and no --file")
	}
}

// TestFilesGrepFileFlag pins that grep accepts --file for the path with the
// pattern as the sole positional.
func TestFilesGrepFileFlag(t *testing.T) {
	path := writeLocalDoc(t)
	resetGrepFlags(t)
	if err := filesGrepCmd.Flags().Set("file", path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filesGrepCmd.Flags().Set("file", "") })

	stdout, _ := captureStd(t, func() {
		if err := filesGrepCmd.RunE(filesGrepCmd, []string{"world"}); err != nil {
			t.Fatalf("grep: %v", err)
		}
	})
	if !strings.Contains(stdout, "world") {
		t.Fatalf("grep --file output = %q, want a match for 'world'", stdout)
	}
}

// TestTablesGetJSONPrintsTableObject pins that `tables get --output json` prints
// the single table object flat (unwrapping the `{table: {...}}` envelope), the
// same shape `tables create` already returns.
func TestTablesGetJSONPrintsTableObject(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/v1/tables/") {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.RequestURI())
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"table": tableFixture()})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "get", "tbl_bank", "--output", "json"); err != nil {
			t.Fatalf("tables get: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode stdout %q: %v", stdout, err)
	}
	if got["id"] != "tbl_bank" {
		t.Fatalf("id = %v, want tbl_bank (flat); stdout=%s", got["id"], stdout)
	}
	if _, wrapped := got["table"]; wrapped {
		t.Fatalf("get JSON should be a single table object, got wrapped envelope: %s", stdout)
	}
}
