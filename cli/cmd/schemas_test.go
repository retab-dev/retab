//go:build !retab_oagen_cli_schemas

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// TestSchemasGenerateRetriesTransientServerFailure pins that a --background
// --wait generation whose first attempt fails with a transient server error
// (the orchestrator "context deadline exceeded") is automatically resubmitted,
// and the recovered schema is emitted.
func TestSchemasGenerateRetriesTransientServerFailure(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			n := posts.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": fmt.Sprintf("sch_%d", n), "status": "queued"})
			return
		}
		// GET poll: first job fails transiently, second completes.
		if strings.HasSuffix(r.URL.Path, "sch_1") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "sch_1", "status": "failed",
				"error": map[string]any{"code": "deadline", "message": `Post "https://orchestrator/...": context deadline exceeded`},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "sch_2", "status": "completed",
			"json_schema": map[string]any{"type": "object"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	set := func(name, val string) {
		if err := schemasGenerateCmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })
	set("url", "https://example.com/doc.pdf")
	set("background", "true")
	set("wait", "true")
	set("max-retries", "2")
	set("retry-delay-ms", "1")
	set("poll-interval-ms", "1")
	t.Cleanup(func() {
		for name, def := range map[string]string{"background": "false", "wait": "false", "max-retries": "3", "retry-delay-ms": "3000", "poll-interval-ms": "2000"} {
			_ = schemasGenerateCmd.Flags().Set(name, def)
		}
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})

	var err error
	stdout, _ := captureStd(t, func() { err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posts.Load() != 2 {
		t.Fatalf("expected 2 submissions (1 retry), got %d", posts.Load())
	}
	var got map[string]any
	if jerr := json.Unmarshal([]byte(stdout), &got); jerr != nil {
		t.Fatalf("stdout not schema JSON: %v\n%s", jerr, stdout)
	}
	if got["type"] != "object" {
		t.Fatalf("final output should be the recovered schema, got:\n%s", stdout)
	}
}

// TestSchemasGenerateDoesNotRetryNonTransientFailure pins that a terminal
// failure that is NOT transient (e.g. a real document/schema error) is
// surfaced immediately without resubmitting.
func TestSchemasGenerateDoesNotRetryNonTransientFailure(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			posts.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "sch_x", "status": "queued"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "sch_x", "status": "failed",
			"error": map[string]any{"code": "bad_document", "message": "document is not a supported type"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	set := func(name, val string) {
		if err := schemasGenerateCmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })
	set("url", "https://example.com/doc.pdf")
	set("background", "true")
	set("wait", "true")
	set("max-retries", "3")
	set("retry-delay-ms", "1")
	set("poll-interval-ms", "1")
	t.Cleanup(func() {
		for name, def := range map[string]string{"background": "false", "wait": "false", "max-retries": "3", "retry-delay-ms": "3000", "poll-interval-ms": "2000"} {
			_ = schemasGenerateCmd.Flags().Set(name, def)
		}
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})

	var err error
	_, _ = captureStd(t, func() { err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil) })
	if err == nil {
		t.Fatal("expected a terminal-failure error, got nil")
	}
	if posts.Load() != 1 {
		t.Fatalf("non-transient failure must not be retried; got %d submissions", posts.Load())
	}
}

// TestSchemasGenerateSyncTimesOutAttemptAndRetries pins that --timeout-seconds
// actually bounds a *synchronous* (default, non-background) attempt: when the
// first submission stalls past the per-attempt deadline the CLI abandons it and
// resubmits, rather than blocking on the SDK client's long default timeout. This
// is the behaviour the help text promises ("Each attempt waits up to
// --timeout-seconds ... to abandon a stuck job and resubmit sooner").
func TestSchemasGenerateSyncTimesOutAttemptAndRetries(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := posts.Add(1)
		if n == 1 {
			// First attempt: stall past --timeout-seconds so the per-attempt
			// deadline fires. Honour cancellation so the handler unblocks as
			// soon as the client gives up.
			select {
			case <-r.Context().Done():
			case <-time.After(2 * time.Second):
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"json_schema": map[string]any{"type": "object"}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	set := func(name, val string) {
		if err := schemasGenerateCmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(context.Background()) })
	set("url", "https://example.com/doc.pdf")
	set("timeout-seconds", "1")
	set("max-retries", "1")
	set("retry-delay-ms", "1")
	t.Cleanup(func() {
		for name, def := range map[string]string{"timeout-seconds": "600", "max-retries": "3", "retry-delay-ms": "3000"} {
			_ = schemasGenerateCmd.Flags().Set(name, def)
		}
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})

	var err error
	stdout, _ := captureStd(t, func() { err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posts.Load() != 2 {
		t.Fatalf("expected the stalled first attempt to be abandoned and resubmitted (2 posts), got %d", posts.Load())
	}
	var got map[string]any
	if jerr := json.Unmarshal([]byte(stdout), &got); jerr != nil {
		t.Fatalf("stdout not schema JSON: %v\n%s", jerr, stdout)
	}
	if got["type"] != "object" {
		t.Fatalf("final output should be the recovered schema, got:\n%s", stdout)
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

func TestIsTransientGenFailure(t *testing.T) {
	transient := []string{
		"context deadline exceeded",
		"upstream deadline exceeded",
		"request timed out",
		"gateway timeout",
		"read tcp: connection reset by peer",
		"dial: connection refused",
		"service unavailable",
		"backend temporarily unavailable",
		"HTTP 503 from orchestrator",
		"received status 502",
		"unexpected EOF",
	}
	for _, d := range transient {
		if !isTransientGenFailure(d) {
			t.Errorf("expected transient: %q", d)
		}
	}

	// Genuine terminal failures whose message merely contains the characters of
	// a transient token must NOT be retried.
	terminal := []string{
		"",
		"   ",
		"invalid schema: value 1503 is out of range",
		"validation failed on field neofield",
		"field amount_503 is required",
		"unauthorized: invalid api key",
		"document 5040 not found",
		"unsupported file type",
	}
	for _, d := range terminal {
		if isTransientGenFailure(d) {
			t.Errorf("expected non-transient: %q", d)
		}
	}
}
