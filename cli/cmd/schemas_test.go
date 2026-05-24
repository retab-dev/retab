package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

func TestWriteGeneratedSchemaDefaultEmitsReusableSchema(t *testing.T) {
	response := retab.PartialSchema{
		Object: "schema",
		JSONSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"invoice_number": map[string]any{"type": "string"},
			},
		},
	}

	var buf bytes.Buffer
	if err := writeGeneratedSchema(&buf, &response, "schema"); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
	}
	if got["type"] != "object" {
		t.Fatalf("type = %#v, want object", got["type"])
	}
	if _, ok := got["json_schema"]; ok {
		t.Fatalf("default output should be the reusable schema body, got envelope:\n%s", buf.String())
	}
}

func TestWriteGeneratedSchemaJSONFormatPreservesEnvelope(t *testing.T) {
	createdAt := "2026-05-15T11:42:06Z"
	response := retab.PartialSchema{
		Object:     "schema",
		CreatedAt:  &createdAt,
		JSONSchema: map[string]any{"type": "object"},
	}

	var buf bytes.Buffer
	if err := writeGeneratedSchema(&buf, &response, "json"); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
	}
	if got["object"] != "schema" {
		t.Fatalf("object = %#v, want schema", got["object"])
	}
	if _, ok := got["json_schema"]; !ok {
		t.Fatalf("json format should preserve envelope:\n%s", buf.String())
	}
}

// A missing --file path used to reach retab.InferMIMEData directly and
// surface as the cryptic "retab: unsupported MIME input string" — the
// same failure a binary blob with no detectable mime produces. schemas
// generate should stat the path upfront like resolveDocument does and
// report a clear "file not found:" instead.
func TestSchemasGenerateFileMissingReportsClearError(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"json_schema": map[string]any{"type": "object"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })

	missing := "/definitely/does/not/exist.pdf"
	if err := schemasGenerateCmd.Flags().Set("file", missing); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if slice, ok := schemasGenerateCmd.Flags().Lookup("file").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})

	var err error
	_, _ = captureStd(t, func() {
		err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil)
	})
	if err == nil {
		t.Fatal("expected error for missing --file, got nil")
	}
	if !strings.Contains(err.Error(), "file not found: "+missing) {
		t.Fatalf("error should report a clear file-not-found, got: %v", err)
	}
	if strings.Contains(err.Error(), "unsupported MIME") {
		t.Fatalf("error leaks the cryptic MIME failure: %v", err)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

// `schemas generate` uniquely uses `--format` while every other command
// uses `--output`. Per the UX papercut audit, the two flags are kept
// distinct (different semantics: --format chooses WHICH payload to
// return, --output chooses HOW to render it) but the Long: help text
// must explicitly call out the distinction so users don't see them as
// duplicate. This test pins the disambiguation text — losing it means
// the help reverts to looking like it has two names for one concept.
func TestSchemasGenerateLongDisambiguatesFormatVsOutput(t *testing.T) {
	long := schemasGenerateCmd.Long
	if !strings.Contains(long, "--format") {
		t.Errorf("Long: text missing --format mention:\n%s", long)
	}
	if !strings.Contains(long, "--output") {
		t.Errorf("Long: text missing --output mention:\n%s", long)
	}
	// The whole point is to explain the distinction. Look for either
	// "which" (which payload) or "how" (how to format) phrasing so the
	// help can be reworded without churning this test.
	lowered := strings.ToLower(long)
	hasWhich := strings.Contains(lowered, "which")
	hasHow := strings.Contains(lowered, "how")
	if !hasWhich || !hasHow {
		t.Errorf("Long: text should explain --format selects WHICH payload and --output selects HOW to print:\n%s", long)
	}
}

func TestSchemasGenerateRejectsUnknownFormatBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"json_schema": map[string]any{"type": "object"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })
	if err := schemasGenerateCmd.Flags().Set("url", "https://example.com/sample.pdf"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})
	if err := schemasGenerateCmd.Flags().Set("format", "yaml"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = schemasGenerateCmd.Flags().Set("format", "schema") })

	var err error
	_, stderr := captureStd(t, func() {
		err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil)
	})
	if err == nil {
		t.Fatalf("expected invalid format error")
	}
	if !strings.Contains(stderr, "invalid --format") {
		t.Fatalf("stderr %q does not mention invalid --format", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}
