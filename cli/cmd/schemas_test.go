//go:build !retab_oagen_cli_schemas

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
)

func TestWriteGeneratedSchemaDefaultEmitsReusableSchema(t *testing.T) {
	response := retab.SchemaGeneration{
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
	createdAt := time.Date(2026, 5, 15, 11, 42, 6, 0, time.UTC)
	response := retab.SchemaGeneration{
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

// TestSchemasGenerateBackgroundSendsFlagAndPrintsRecord pins that
// `schemas generate --background` forwards background:true and, without --wait,
// prints the queued record (which has no schema yet) rather than erroring on the
// missing json_schema.
func TestSchemasGenerateBackgroundSendsFlagAndPrintsRecord(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotBackground atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if bg, _ := body["background"].(bool); bg {
			gotBackground.Store(true)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "sch_queued1", "status": "queued"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })
	if err := schemasGenerateCmd.Flags().Set("url", "https://example.com/doc.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := schemasGenerateCmd.Flags().Set("background", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = schemasGenerateCmd.Flags().Set("background", "false")
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})

	var err error
	stdout, _ := captureStd(t, func() { err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotBackground.Load() {
		t.Fatal("request body did not carry background:true")
	}
	for _, want := range []string{"sch_queued1", "queued"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("queued record output missing %q:\n%s", want, stdout)
		}
	}
}

// TestSchemasGenerateForwardsInstructionsAndDPI pins that --instructions and
// --image-resolution-dpi are wired into the request body (the SDK supported
// them before the CLI exposed them).
func TestSchemasGenerateForwardsInstructionsAndDPI(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotInstr atomic.Value
	var gotDPI atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if s, ok := body["instructions"].(string); ok {
			gotInstr.Store(s)
		}
		if d, ok := body["image_resolution_dpi"].(float64); ok {
			gotDPI.Store(int64(d))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"json_schema": map[string]any{"type": "object"}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })
	if err := schemasGenerateCmd.Flags().Set("url", "https://example.com/doc.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := schemasGenerateCmd.Flags().Set("instructions", "Only model deeds; nullable optionals."); err != nil {
		t.Fatal(err)
	}
	if err := schemasGenerateCmd.Flags().Set("image-resolution-dpi", "96"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = schemasGenerateCmd.Flags().Set("instructions", "")
		_ = schemasGenerateCmd.Flags().Set("image-resolution-dpi", "0")
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})

	var err error
	_, _ = captureStd(t, func() { err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, _ := gotInstr.Load().(string); got != "Only model deeds; nullable optionals." {
		t.Fatalf("instructions not forwarded, got %q", got)
	}
	if gotDPI.Load() != 96 {
		t.Fatalf("image_resolution_dpi not forwarded, got %d", gotDPI.Load())
	}
}

// TestSchemasGenerateInstructionsFileForwardedAndConflicts pins that
// --instructions-file is read (and trimmed) into the request body, and that
// passing both --instructions and --instructions-file is rejected before any
// request is made.
func TestSchemasGenerateInstructionsFileForwardedAndConflicts(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotInstr atomic.Value
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if s, ok := body["instructions"].(string); ok {
			gotInstr.Store(s)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"json_schema": map[string]any{"type": "object"}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "instr.txt")
	if err := os.WriteFile(path, []byte("  Scope to deeds only.\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })
	if err := schemasGenerateCmd.Flags().Set("url", "https://example.com/doc.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := schemasGenerateCmd.Flags().Set("instructions-file", path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = schemasGenerateCmd.Flags().Set("instructions-file", "")
		_ = schemasGenerateCmd.Flags().Set("instructions", "")
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})

	var err error
	_, _ = captureStd(t, func() { err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, _ := gotInstr.Load().(string); got != "Scope to deeds only." {
		t.Fatalf("instructions-file content not forwarded/trimmed, got %q", got)
	}

	// Passing both flags must fail before any request.
	hitsBefore := hits.Load()
	if err := schemasGenerateCmd.Flags().Set("instructions", "inline"); err != nil {
		t.Fatal(err)
	}
	_, _ = captureStd(t, func() { err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil) })
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutual-exclusion error, got %v", err)
	}
	if hits.Load() != hitsBefore {
		t.Fatalf("conflict path made a request (hits %d -> %d)", hitsBefore, hits.Load())
	}
}

// TestSchemasGetAndCancelHitGenerationSubpaths pins that the new get/cancel
// commands target /v1/schemas/generate/{id}[/cancel] (the background-primitive
// routes), not /v1/schemas/{id}.
func TestSchemasGetAndCancelHitGenerationSubpaths(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotPath, gotMethod atomic.Value
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath.Store(r.URL.Path)
		gotMethod.Store(r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "sch_x", "status": "cancelled"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, _ = captureStd(t, func() {
		if err := schemasGetCmd.RunE(schemasGetCmd, []string{"sch_x"}); err != nil {
			t.Errorf("get: %v", err)
		}
	})
	if gotPath.Load() != "/v1/schemas/generate/sch_x" || gotMethod.Load() != http.MethodGet {
		t.Fatalf("get hit %v %v, want GET /v1/schemas/generate/sch_x", gotMethod.Load(), gotPath.Load())
	}

	_, _ = captureStd(t, func() {
		if err := schemasCancelCmd.RunE(schemasCancelCmd, []string{"sch_x"}); err != nil {
			t.Errorf("cancel: %v", err)
		}
	})
	if gotPath.Load() != "/v1/schemas/generate/sch_x/cancel" || gotMethod.Load() != http.MethodPost {
		t.Fatalf("cancel hit %v %v, want POST /v1/schemas/generate/sch_x/cancel", gotMethod.Load(), gotPath.Load())
	}
}
