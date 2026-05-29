package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestParseDocumentArgs_DocumentFlagOnly_NoWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs([]string{"start=./a.pdf"}, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./a.pdf" {
		t.Fatalf("got %v, want {start: ./a.pdf}", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestWorkflowsRunsCreateReadsDocumentBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-document.pdf")
	if err := workflowsRunsCreateCmd.Flags().Set("document", "start="+missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsCreateCmd, "document") })

	err := workflowsRunsCreateCmd.RunE(workflowsRunsCreateCmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected missing document file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "--document start=") || !strings.Contains(err.Error(), "missing-document.pdf") {
		t.Fatalf("error should mention missing document file, got %q", err.Error())
	}
}

func TestWorkflowsRunsGetHonorsTableOutputFallback(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/workflows/runs/run_123" {
			t.Fatalf("path = %s, want run get", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "run_123",
			"lifecycle": map[string]any{
				"status": "completed",
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsRunsGetCmd.RunE(workflowsRunsGetCmd, []string{"run_123"}); err != nil {
			t.Fatalf("runs get: %v", err)
		}
	})
	if !strings.Contains(stderr, "falling back to json") {
		t.Fatalf("expected table fallback warning, got stderr %q", stderr)
	}
	if !strings.Contains(stdout, `"id": "run_123"`) {
		t.Fatalf("expected JSON fallback payload, got:\n%s", stdout)
	}
}

func TestWorkflowsRunsCancelSurfacesPendingCancellationStatusToStderr(t *testing.T) {
	// Regression: the cancel endpoint returns 200 even when the cancel
	// signal has not yet been confirmed by Temporal. Previously the CLI
	// just printed the run object, making the response look like the
	// run was already cancelled when it was still ``cancellation_requested``.
	// The fix prints a one-line note to stderr so the user knows to poll.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs/run_pending/cancel" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"run": map[string]any{
				"id":        "run_pending",
				"lifecycle": map[string]any{"status": "running"},
			},
			"redis_available":     true,
			"cancellation_status": "cancellation_requested",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, stderr := captureStd(t, func() {
		if err := workflowsRunsCancelCmd.RunE(workflowsRunsCancelCmd, []string{"run_pending"}); err != nil {
			t.Fatalf("runs cancel: %v", err)
		}
	})
	for _, want := range []string{
		"cancellation_status=\"cancellation_requested\"",
		"not yet reached a terminal state",
		"runs get run_pending",
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("expected stderr to contain %q, got:\n%s", want, stderr)
		}
	}
}

func TestWorkflowsRunsCancelStaysSilentWhenCancellationIsFinalized(t *testing.T) {
	// If the server reports ``cancellation_status=cancelled`` (synchronous
	// terminal state) the CLI should NOT emit the pending-state note —
	// that would be misleading the other direction.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"run": map[string]any{
				"id":        "run_done",
				"lifecycle": map[string]any{"status": "cancelled"},
			},
			"redis_available":     true,
			"cancellation_status": "cancelled",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, stderr := captureStd(t, func() {
		if err := workflowsRunsCancelCmd.RunE(workflowsRunsCancelCmd, []string{"run_done"}); err != nil {
			t.Fatalf("runs cancel: %v", err)
		}
	})
	if strings.Contains(stderr, "not yet reached a terminal state") {
		t.Fatalf("did not expect pending-state note when cancellation_status=cancelled, got:\n%s", stderr)
	}
}

func TestWorkflowsRunsCreateResolvesStartAliasToGeneratedStartBlock(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/blocks" && r.URL.Query().Get("workflow_id") == "wf_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "block_generated", "type": "start_document", "label": "Document"},
					{"id": "parse", "type": "parse", "label": "Parse"},
				},
				"list_metadata": map[string]any{},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["workflow_id"] != "wf_123" {
				t.Fatalf("workflow_id = %#v, want wf_123", body["workflow_id"])
			}
			postedDocuments, _ = body["documents"].(map[string]any)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "run_123",
				"status": "running",
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	docPath := filepath.Join(dir, "invoice.txt")
	if err := os.WriteFile(docPath, []byte("invoice\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := workflowsRunsCreateCmd.Flags().Set("document", "start="+docPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsCreateCmd, "document") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsRunsCreateCmd.RunE(workflowsRunsCreateCmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_123") {
		t.Fatalf("expected run response on stdout, got:\n%s", stdout)
	}
	if _, ok := postedDocuments["block_generated"]; !ok {
		t.Fatalf("documents posted under keys %#v, want block_generated", keysOfAnyMap(postedDocuments))
	}
	if _, ok := postedDocuments["start"]; ok {
		t.Fatalf("friendly alias leaked into request body: %#v", postedDocuments)
	}
}

func TestWorkflowsRunsCreateSendsDocumentURLPayload(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["workflow_id"] != "wf_123" {
			t.Fatalf("workflow_id = %#v, want wf_123", body["workflow_id"])
		}
		postedDocuments, _ = body["documents"].(map[string]any)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "run_123",
			"status": "running",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("document-url", "block_start=https://example.com/invoice.pdf"); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_123") {
		t.Fatalf("expected run response on stdout, got:\n%s", stdout)
	}
	startDocument, ok := postedDocuments["block_start"].(map[string]any)
	if !ok {
		t.Fatalf("block_start document = %#v", postedDocuments["block_start"])
	}
	if startDocument["filename"] != "invoice.pdf" || startDocument["url"] != "https://example.com/invoice.pdf" {
		t.Fatalf("start document = %#v", startDocument)
	}
	if _, ok := startDocument["content"]; ok {
		t.Fatalf("document-url should send url-backed payload, got %#v", startDocument)
	}
}

func TestWorkflowsRunsCreateAcceptsDocumentsFileDescriptors(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["workflow_id"] != "wf_123" {
			t.Fatalf("workflow_id = %#v, want wf_123", body["workflow_id"])
		}
		postedDocuments, _ = body["documents"].(map[string]any)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "run_123",
			"status": "running",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	docsPath := filepath.Join(dir, "documents.json")
	if err := os.WriteFile(
		docsPath,
		[]byte(`{"block_start":{"filename":"invoice.pdf","url":"https://example.com/invoice.pdf"}}`),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("documents-file", docsPath); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_123") {
		t.Fatalf("expected run response on stdout, got:\n%s", stdout)
	}
	startDocument, ok := postedDocuments["block_start"].(map[string]any)
	if !ok {
		t.Fatalf("block_start document = %#v", postedDocuments["block_start"])
	}
	if startDocument["filename"] != "invoice.pdf" || startDocument["url"] != "https://example.com/invoice.pdf" {
		t.Fatalf("start document = %#v", startDocument)
	}
}

func TestWorkflowsRunsCreateResolvesDocumentIDToFileRef(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/files/file_abc":
			// Resolve file metadata for the FileRef.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":        "file_abc",
				"filename":  "deed.tiff",
				"mime_type": "image/tiff",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["workflow_id"] != "wf_123" {
				t.Fatalf("workflow_id = %#v, want wf_123", body["workflow_id"])
			}
			postedDocuments, _ = body["documents"].(map[string]any)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "run_123",
				"status": "running",
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().StringArray("document-id", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("document-id", "block_start=file_abc"); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_123") {
		t.Fatalf("expected run response on stdout, got:\n%s", stdout)
	}
	startDocument, ok := postedDocuments["block_start"].(map[string]any)
	if !ok {
		t.Fatalf("block_start document = %#v", postedDocuments["block_start"])
	}
	if startDocument["id"] != "file_abc" {
		t.Fatalf("expected FileRef id file_abc, got %#v", startDocument)
	}
	if startDocument["filename"] != "deed.tiff" || startDocument["mime_type"] != "image/tiff" {
		t.Fatalf("FileRef should carry resolved filename/mime_type, got %#v", startDocument)
	}
	if _, ok := startDocument["url"]; ok {
		t.Fatalf("FileRef should reference by id, not send a url; got %#v", startDocument)
	}
	if _, ok := startDocument["content"]; ok {
		t.Fatalf("FileRef should not inline content; got %#v", startDocument)
	}
}

func TestWorkflowsRunsCreatePreservesFileRefIDFromDocumentsFile(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		postedDocuments, _ = body["documents"].(map[string]any)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "run_123",
			"status": "running",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	docsPath := filepath.Join(dir, "documents.json")
	if err := os.WriteFile(
		docsPath,
		[]byte(`{"block_start":{"id":"file_abc","filename":"deed.tiff","mime_type":"image/tiff"}}`),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().StringArray("document-id", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("documents-file", docsPath); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create: %v", err)
		}
	})
	if !strings.Contains(stdout, "run_123") {
		t.Fatalf("expected run response on stdout, got:\n%s", stdout)
	}
	startDocument, ok := postedDocuments["block_start"].(map[string]any)
	if !ok {
		t.Fatalf("block_start document = %#v", postedDocuments["block_start"])
	}
	// The id key must survive the --documents-file passthrough (regression:
	// it used to be silently dropped, making file references impossible).
	if startDocument["id"] != "file_abc" {
		t.Fatalf("expected documents-file id to be preserved, got %#v", startDocument)
	}
	if startDocument["filename"] != "deed.tiff" || startDocument["mime_type"] != "image/tiff" {
		t.Fatalf("FileRef descriptor = %#v", startDocument)
	}
}

func TestWorkflowsRunsCreateRejectsConflictingDocumentIDAndURL(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().StringArray("document-id", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("document-id", "block_start=file_abc"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("document-url", "block_start=https://example.com/x.pdf"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected conflict error for block claimed by both --document-id and --document-url")
	}
	if !strings.Contains(err.Error(), "block_start") || !strings.Contains(err.Error(), "--document-id") {
		t.Fatalf("error should name the block and conflicting flags, got: %v", err)
	}
}

func TestWorkflowsRunsCreateResolvesStartAliasFromDocumentsFile(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/blocks" && r.URL.Query().Get("workflow_id") == "wf_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "block_generated", "type": "start_document", "label": "Document"},
					{"id": "extract", "type": "extract", "label": "Extract"},
				},
				"list_metadata": map[string]any{},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["workflow_id"] != "wf_123" {
				t.Fatalf("workflow_id = %#v, want wf_123", body["workflow_id"])
			}
			postedDocuments, _ = body["documents"].(map[string]any)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "run_123",
				"status": "running",
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	docsPath := filepath.Join(t.TempDir(), "documents.json")
	if err := os.WriteFile(
		docsPath,
		[]byte(`{"start":{"filename":"invoice.pdf","url":"https://example.com/invoice.pdf"}}`),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")
	if err := cmd.Flags().Set("documents-file", docsPath); err != nil {
		t.Fatal(err)
	}

	if err := cmd.RunE(cmd, []string{"wf_123"}); err != nil {
		t.Fatalf("runs create: %v", err)
	}
	if _, ok := postedDocuments["block_generated"]; !ok {
		t.Fatalf("documents posted under keys %#v, want block_generated", keysOfAnyMap(postedDocuments))
	}
	if _, ok := postedDocuments["start"]; ok {
		t.Fatalf("friendly alias leaked into request body: %#v", postedDocuments)
	}
}

func TestWorkflowsRunsCreateValidatesJSONInputsBeforeResolvingStartAlias(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks" || r.URL.Query().Get("workflow_id") != "wf_123" {
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "block_generated", "type": "start_document", "label": "Document"},
			},
			"list_metadata": map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	_ = cmd.Flags().Set("document-url", "start=https://example.com/invoice.pdf")
	_ = cmd.Flags().Set("json-inputs-file", "/does/not/exist.json")

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected json-inputs-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--json-inputs-file") {
		t.Fatalf("error %q does not mention --json-inputs-file", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want json inputs validation before start alias resolution", got)
	}
}

// Regression for 2026-05 CLI probing: passing both --document and
// --document-url for the SAME block id used to silently overwrite one
// with the other, and (depending on order) the lost --document-url
// produced a 500 from the server's document-fetch path. Reject the
// conflict up-front with a clear message instead.
func TestWorkflowsRunsCreateRejectsConflictingDocumentSourcesForSameBlock(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached for conflicting document sources, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	tempDir := t.TempDir()
	docPath := tempDir + "/invoice.pdf"
	if err := os.WriteFile(docPath, []byte("%PDF-1.4 fake"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("document", "start="+docPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("document-url", "start=https://example.com/invoice.pdf"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected conflict error for same-block --document + --document-url")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "both --document and --document-url") {
		t.Fatalf("error %q should mention the conflicting flags", err.Error())
	}
	if !strings.Contains(err.Error(), "start") {
		t.Fatalf("error %q should name the offending block id", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsRunsCreateRejectsDuplicateDocumentFlag(t *testing.T) {
	// `retab workflows runs create wf_x --document start=a --document start=b`
	// used to silently keep the last entry. Now the second occurrence is a
	// hard error so users notice the misconfiguration.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached for duplicate --document, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	tempDir := t.TempDir()
	docPathA := filepath.Join(tempDir, "a.pdf")
	docPathB := filepath.Join(tempDir, "b.pdf")
	for _, p := range []string{docPathA, docPathB} {
		if err := os.WriteFile(p, []byte("%PDF-1.4 fake"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("document", "start="+docPathA); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("document", "start="+docPathB); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected duplicate --document error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "start") {
		t.Fatalf("error %q should name the offending block id", err.Error())
	}
	if !strings.Contains(err.Error(), "--document") {
		t.Fatalf("error %q should name the --document flag", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsRunsCreateRejectsDuplicateAcrossDocumentAndDocumentUrl(t *testing.T) {
	// Pin the existing cross-source guard: when the same block id appears
	// in both `--document` and `--document-url`, the command must refuse
	// before any server call.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached for cross-flag duplicate, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	tempDir := t.TempDir()
	docPath := filepath.Join(tempDir, "invoice.pdf")
	if err := os.WriteFile(docPath, []byte("%PDF-1.4 fake"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("document", "start="+docPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("document-url", "start=https://example.com/invoice.pdf"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected cross-flag duplicate error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--document") || !strings.Contains(err.Error(), "--document-url") {
		t.Fatalf("error %q should name both conflicting flags", err.Error())
	}
	if !strings.Contains(err.Error(), "start") {
		t.Fatalf("error %q should name the offending block id", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsRunsCreateRejectsEmptyDocumentURLBlockIDBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached for invalid document-url, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-run-create", RunE: workflowsRunsCreateCmd.RunE}
	cmd.Flags().String("version", "", "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().StringArray("document", nil, "")
	cmd.Flags().StringArray("document-file", nil, "")
	cmd.Flags().StringArray("document-url", nil, "")
	cmd.Flags().String("json-inputs-file", "", "")

	if err := cmd.Flags().Set("document-url", "=https://example.com/invoice.pdf"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected document-url validation error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--document-url expects block-id=url") {
		t.Fatalf("error %q does not mention --document-url shape", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func keysOfAnyMap(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

// Regression: “retab workflows runs list --workflow-id ""“ used to silently
// return runs from every workflow in the org. pflag returns "" for an empty
// string, the filter check “flagID != ""“ was false, and the listing
// proceeded with no workflow filter. The positional form already errors
// (the SDK guard rejects empty workflowIDs), so the two forms used to
// disagree on what an empty arg means.
func TestWorkflowsRunsListRejectsEmptyWorkflowIDFlag(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached for empty --workflow-id, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsListCmd.Flags().Set("workflow-id", ""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, "workflow-id") })

	var err error
	_, _ = captureStd(t, func() {
		err = workflowsRunsListCmd.RunE(workflowsRunsListCmd, nil)
	})
	if err == nil {
		t.Fatal("expected --workflow-id must not be blank error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--workflow-id must not be blank") {
		t.Fatalf("error %q does not mention --workflow-id", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsRunsListRejectsInvalidListFlagsLocally(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
		reset     string
	}{
		{name: "negative limit", flag: "limit", value: "-1", wantError: "between 1 and 100", reset: "1"},
		{name: "invalid order", flag: "order", value: "sideways", wantError: "asc", reset: ""},
		{name: "invalid from date", flag: "from-date", value: "not-a-date", wantError: "YYYY-MM-DD", reset: ""},
		{name: "invalid to date", flag: "to-date", value: "not-a-date", wantError: "YYYY-MM-DD", reset: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := workflowsRunsListCmd.Flags().Set(tc.flag, tc.value)
			if err == nil {
				t.Fatalf("expected local parse error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if resetErr := workflowsRunsListCmd.Flags().Set(tc.flag, tc.reset); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestWorkflowsRunsListRejectsOverLimitLocally(t *testing.T) {
	err := workflowsRunsListCmd.Flags().Set("limit", "101")
	if err == nil {
		t.Fatal("expected local parse error for --limit=101")
	}
	if !strings.Contains(err.Error(), "between 1 and 100") {
		t.Fatalf("error %q does not mention backend limit range", err.Error())
	}
	if resetErr := workflowsRunsListCmd.Flags().Set("limit", "1"); resetErr != nil {
		t.Fatalf("reset --limit: %v", resetErr)
	}
}

func TestWorkflowsRunsCommandsRejectInvalidEnumFiltersBeforeRequest(t *testing.T) {
	cases := []struct {
		name      string
		cmd       *cobra.Command
		args      []string
		flag      string
		value     string
		wantError string
	}{
		{name: "list invalid status", cmd: workflowsRunsListCmd, flag: "status", value: "banana", wantError: "invalid --status"},
		{name: "list invalid exclude status", cmd: workflowsRunsListCmd, flag: "exclude-status", value: "banana", wantError: "invalid --exclude-status"},
		{name: "list invalid trigger type", cmd: workflowsRunsListCmd, flag: "trigger-type", value: "banana", wantError: "invalid --trigger-type"},
		{name: "export invalid status", cmd: workflowsRunsExportCmd, flag: "status", value: "banana", wantError: "invalid --status"},
		{name: "export invalid exclude status", cmd: workflowsRunsExportCmd, flag: "exclude-status", value: "banana", wantError: "invalid --exclude-status"},
		{name: "export invalid trigger type", cmd: workflowsRunsExportCmd, flag: "trigger-type", value: "banana", wantError: "invalid --trigger-type"},
		{name: "export invalid export source", cmd: workflowsRunsExportCmd, flag: "export-source", value: "banana", wantError: "invalid --export-source"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			hits := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				t.Fatalf("server should not be reached for invalid local filter, got %s %s", r.Method, r.URL.String())
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := tc.cmd.Flags().Set(tc.flag, tc.value); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { resetWorkflowRunsFlag(t, tc.cmd, tc.flag) })

			var err error
			_, stderr := captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, tc.args)
			})
			if err == nil {
				t.Fatalf("expected local validation error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if hits != 0 {
				t.Fatalf("server was hit %d time(s), want 0", hits)
			}
		})
	}
}

func TestWorkflowsRunsRestartCreatesFreshRunFromSourceInputs(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	if flag := workflowsRunsRestartCmd.Flags().Lookup("command-id"); flag != nil {
		t.Fatalf("restart should not expose --command-id")
	}

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs/run_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "run_123",
				"workflow": map[string]any{
					"workflow_id":       "wf_123",
					"version_id":        "ver_old",
					"name_at_run_time":  "Workflow",
					"requested_version": "production",
				},
				"trigger":   map[string]any{"type": "manual"},
				"lifecycle": map[string]any{"status": "error"},
				"inputs": map[string]any{
					"documents": map[string]any{
						"start_doc": map[string]any{
							"id":        "file_123",
							"filename":  "invoice.pdf",
							"mime_type": "application/pdf",
						},
					},
					"json_data": map[string]any{
						"start_json": map[string]any{"invoice_id": "INV-1"},
					},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "run_456",
				"workflow": map[string]any{
					"workflow_id":       "wf_123",
					"version_id":        "ver_123",
					"name_at_run_time":  "Workflow",
					"requested_version": "production",
				},
				"trigger":   map[string]any{"type": "api"},
				"lifecycle": map[string]any{"status": "running"},
				"timing":    map[string]any{"created_at": "2026-05-15T00:00:00Z"},
				"inputs": map[string]any{
					"documents": map[string]any{},
					"json_data": map[string]any{},
				},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsRestartCmd.Flags().Set("config-source", "published"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsRunsRestartCmd.Flags().Set("config-source", "published") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsRunsRestartCmd.RunE(workflowsRunsRestartCmd, []string{"run_123"}); err != nil {
			t.Fatalf("runs restart: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_456") {
		t.Fatalf("expected restart response on stdout, got:\n%s", stdout)
	}
	if body["workflow_id"] != "wf_123" || body["version"] != "production" {
		t.Fatalf("create body = %#v", body)
	}
	if _, ok := body["restart_of"]; ok {
		t.Fatalf("restart_of leaked into composed create body: %#v", body)
	}
	if _, ok := body["command_id"]; ok {
		t.Fatalf("command_id leaked into composed create body: %#v", body)
	}
	if _, ok := body["config_source"]; ok {
		t.Fatalf("config_source leaked into composed create body: %#v", body)
	}
	documents, ok := body["documents"].(map[string]any)
	if !ok {
		t.Fatalf("documents = %#v", body["documents"])
	}
	startDoc, ok := documents["start_doc"].(map[string]any)
	if !ok {
		t.Fatalf("start_doc = %#v", documents["start_doc"])
	}
	if startDoc["id"] != "file_123" || startDoc["filename"] != "invoice.pdf" || startDoc["mime_type"] != "application/pdf" {
		t.Fatalf("start_doc = %#v", startDoc)
	}
	jsonInputs, ok := body["json_inputs"].(map[string]any)
	if !ok {
		t.Fatalf("json_inputs = %#v", body["json_inputs"])
	}
	startJSON, ok := jsonInputs["start_json"].(map[string]any)
	if !ok || startJSON["invoice_id"] != "INV-1" {
		t.Fatalf("start_json = %#v", jsonInputs["start_json"])
	}
}

func TestWorkflowsRunsRestartMapsDraftConfigSourceToDraftVersion(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs/run_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "run_123",
				"workflow": map[string]any{
					"workflow_id":       "wf_123",
					"version_id":        "ver_old",
					"name_at_run_time":  "Workflow",
					"requested_version": "production",
				},
				"trigger":   map[string]any{"type": "manual"},
				"lifecycle": map[string]any{"status": "error"},
				"inputs": map[string]any{
					"documents": map[string]any{},
					"json_data": map[string]any{},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "run_456",
				"workflow": map[string]any{
					"workflow_id":       "wf_123",
					"version_id":        "ver_draft",
					"name_at_run_time":  "Workflow",
					"requested_version": "draft",
				},
				"trigger":   map[string]any{"type": "api"},
				"lifecycle": map[string]any{"status": "running"},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsRestartCmd.Flags().Set("config-source", "draft"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsRunsRestartCmd.Flags().Set("config-source", "published") })

	_, stderr := captureStd(t, func() {
		if err := workflowsRunsRestartCmd.RunE(workflowsRunsRestartCmd, []string{"run_123"}); err != nil {
			t.Fatalf("runs restart: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if body["version"] != "draft" {
		t.Fatalf("create body = %#v", body)
	}
}

func TestWorkflowsRunsRestartRejectsInvalidConfigSourceLocally(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		t.Fatalf("server should not be reached for invalid config source, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsRestartCmd.Flags().Set("config-source", "preview"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsRunsRestartCmd.Flags().Set("config-source", "published") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsRunsRestartCmd.RunE(workflowsRunsRestartCmd, []string{"run_123"})
	})
	if err == nil {
		t.Fatal("expected invalid config source error")
	}
	if !strings.Contains(errors.Unwrap(err).Error(), "--config-source must be published or draft") {
		t.Fatalf("error %q does not mention valid config sources", err.Error())
	}
	if !strings.Contains(stderr, "--config-source must be published or draft") {
		t.Fatalf("stderr %q does not mention valid config sources", stderr)
	}
	if hits != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits)
	}
}

func resetWorkflowRunsFlag(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing workflow runs flag %q", name)
	}
	if slice, ok := flag.Value.(pflag.SliceValue); ok {
		if err := slice.Replace(nil); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
		flag.Changed = false
		return
	}
	// Boolean and numeric flags reject "" — restore them by writing their default.
	switch flag.Value.Type() {
	case "bool", "float", "float64", "int", "int64", "uint", "uint64":
		if err := cmd.Flags().Set(name, flag.DefValue); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
		flag.Changed = false
		return
	}
	if err := cmd.Flags().Set(name, ""); err != nil {
		t.Fatalf("reset --%s: %v", name, err)
	}
	flag.Changed = false
}

func TestWorkflowsStepsListExampleUsesPaginatedEnvelope(t *testing.T) {
	example := workflowsStepsListCmd.Example
	if !strings.Contains(example, ".data[]") {
		t.Fatalf("steps list example should iterate over .data[], got:\n%s", example)
	}
	if !strings.Contains(example, ".lifecycle.status") {
		t.Fatalf("steps list example should read lifecycle status, got:\n%s", example)
	}
}

// Regression guard for the steps-get help example. The pre-cutover hint
// piped `jq '.input'` but the server response field is `handle_inputs`
// (StepResponse in routes.py). A doc that names a non-existent field on
// the live response is worse than no example — users copy-paste and get
// empty output.
func TestWorkflowsStepsGetExampleNamesARealResponseField(t *testing.T) {
	example := workflowsStepsGetCmd.Example
	if strings.Contains(example, "'.input'") || strings.Contains(example, ".input ") || strings.HasSuffix(strings.TrimSpace(example), ".input") {
		t.Fatalf("steps get example references stale .input — StepResponse exposes .handle_inputs / .handle_outputs:\n%s", example)
	}
	if !strings.Contains(example, "handle_inputs") && !strings.Contains(example, "handle_outputs") {
		t.Fatalf("steps get example should reference a real StepResponse field (handle_inputs / handle_outputs):\n%s", example)
	}
}

// `runs delete` is destructive in the same way `workflows delete` is:
// it permanently removes the run document and every step record. It used
// to delete silently with no confirmation flag, breaking the convention
// established by the rest of the destructive delete commands
// (`workflows delete`, `workflows blocks delete`, `workflows edges delete`,
// `workflows experiments delete`, `workflows tests delete`). This pin
// keeps `runs delete` aligned: without --yes and without a TTY stdin,
// the command refuses, and the server is never contacted.
func TestWorkflowsRunsDeleteRefusesWithoutYesWhenStdinNotATTY(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached when --yes is not passed and stdin is not a TTY, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// Reset --yes between tests so a previous test's setting can't leak in.
	if err := workflowsRunsDeleteCmd.Flags().Set("yes", "false"); err != nil {
		t.Fatal(err)
	}

	err := workflowsRunsDeleteCmd.RunE(workflowsRunsDeleteCmd, []string{"run_xyz"})
	if err == nil {
		t.Fatal("expected refusal error when --yes not passed and stdin not a TTY")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	for _, want := range []string{"--yes", "run", "run_xyz"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not contain %q", err.Error(), want)
		}
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsRunsDeleteWithYesSendsDelete(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawDelete atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/run_xyz") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		sawDelete.Store(true)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsRunsDeleteCmd.Flags().Set("yes", "false")
	})

	if err := workflowsRunsDeleteCmd.RunE(workflowsRunsDeleteCmd, []string{"run_xyz"}); err != nil {
		t.Fatalf("delete with --yes should succeed, got %v", err)
	}
	if !sawDelete.Load() {
		t.Fatal("expected DELETE request to be issued")
	}
}

func TestWorkflowsStepsGetUsesStepIDRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/steps/step_123" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"step_id":       "step_123",
			"run_id":        "run_123",
			"block_id":      "blk_123",
			"handle_inputs": map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsStepsGetCmd.RunE(workflowsStepsGetCmd, []string{"step_123"}); err != nil {
			t.Fatalf("steps get: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "step_123") {
		t.Fatalf("expected step response on stdout, got:\n%s", stdout)
	}
}

func TestWorkflowsStepsGetHelpUsesStepID(t *testing.T) {
	for _, text := range []string{workflowsStepsCmd.Example, workflowsStepsGetCmd.Use, workflowsStepsGetCmd.Long, workflowsStepsGetCmd.Example} {
		if strings.Contains(text, "run_xyz789 blk_extract_1") || strings.Contains(text, "<run-id> <block-id>") {
			t.Fatalf("steps get help should use step id, got:\n%s", text)
		}
	}
	if !strings.Contains(workflowsStepsGetCmd.Use, "<step-id>") {
		t.Fatalf("steps get usage should mention <step-id>, got %q", workflowsStepsGetCmd.Use)
	}
}

func TestWorkflowsStepsCommandIsTopLevelOnly(t *testing.T) {
	if commandByName(workflowsCmd, "steps") != workflowsStepsCmd {
		t.Fatal("workflows steps command is not registered at the top level")
	}
	if commandByName(workflowsRunsCmd, "steps") != nil {
		t.Fatal("nested steps command should not be registered under runs")
	}
}

func commandByName(parent *cobra.Command, name string) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func TestWorkflowsRunsGetHelpUsesLifecycleStatusPath(t *testing.T) {
	if !strings.Contains(workflowsRunsGetCmd.Example, ".lifecycle.status") {
		t.Fatalf("runs get example should read lifecycle.status:\n%s", workflowsRunsGetCmd.Example)
	}
	if strings.Contains(workflowsRunsGetCmd.Example, "jq -r '.status'") {
		t.Fatalf("runs get example should not read top-level status:\n%s", workflowsRunsGetCmd.Example)
	}
}

func TestWorkflowsRunsListExamplesUseBackendStatusNames(t *testing.T) {
	if strings.Contains(workflowsRunsListCmd.Example, "--status failed") {
		t.Fatalf("runs list example should use backend status name error, got:\n%s", workflowsRunsListCmd.Example)
	}
	if !strings.Contains(workflowsRunsListCmd.Example, "--status error") {
		t.Fatalf("runs list example should include --status error, got:\n%s", workflowsRunsListCmd.Example)
	}
}

func TestWorkflowRunsListTableUsesStatusColumn(t *testing.T) {
	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	result := map[string]any{
		"data": []any{
			map[string]any{
				"id": "run_1",
				"workflow": map[string]any{
					"name_at_run_time": "Invoice workflow",
				},
				"lifecycle": map[string]any{
					"status": "awaiting_review",
				},
				"timing": map[string]any{
					"created_at": "2026-05-20T15:31:28Z",
				},
			},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printWorkflowRunListResult(workflowsRunsListCmd, result); err != nil {
			t.Fatalf("printWorkflowRunListResult: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	for _, want := range []string{"ID", "NAME", "STATUS", "CREATED_AT", "run_1", "Invoice workflow", "awaiting_review"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in runs table:\n%s", want, stdout)
		}
	}
	if strings.Contains(strings.SplitN(stdout, "\n", 2)[0], "TYPE") {
		t.Fatalf("runs list table should call lifecycle values STATUS, not TYPE:\n%s", stdout)
	}
}

func TestWorkflowsRunsHelpUsesBackendReviewStatusName(t *testing.T) {
	staleBackendStatus := "waiting_for" + "_human"
	if strings.Contains(workflowsRunsCmd.Long, staleBackendStatus) {
		t.Fatalf("runs help should not mention stale legacy review status:\n%s", workflowsRunsCmd.Long)
	}
	if strings.Contains(workflowsRunsCmd.Long, "escalate") {
		t.Fatalf("runs help should not advertise unsupported review escalation:\n%s", workflowsRunsCmd.Long)
	}
	if !strings.Contains(workflowsRunsCmd.Long, "awaiting_review") {
		t.Fatalf("runs help should mention backend review status awaiting_review:\n%s", workflowsRunsCmd.Long)
	}
}

func TestWorkflowsRunsExportRejectsInvalidDateFlagsLocally(t *testing.T) {
	cases := []struct {
		name string
		flag string
	}{
		{name: "invalid from date", flag: "from-date"},
		{name: "invalid to date", flag: "to-date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := workflowsRunsExportCmd.Flags().Set(tc.flag, "not-a-date")
			if err == nil {
				t.Fatalf("expected local parse error for --%s=not-a-date", tc.flag)
			}
			if !strings.Contains(err.Error(), "YYYY-MM-DD") {
				t.Fatalf("error %q does not contain YYYY-MM-DD", err.Error())
			}
			if resetErr := workflowsRunsExportCmd.Flags().Set(tc.flag, ""); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestWorkflowsRunsExportSendsTriggerType(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs/export" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"csv_data": "id\nrun_123\n",
			"rows":     1,
			"columns":  1,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	flags := map[string]string{
		"block-id":     "blk_123",
		"trigger-type": "api",
	}
	for flag, value := range flags {
		if err := workflowsRunsExportCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsExportCmd, flag) })
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsRunsExportCmd.RunE(workflowsRunsExportCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("runs export: %v\nstderr:\n%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	triggerType, ok := body["trigger_type"].(string)
	if !ok {
		t.Fatalf("trigger_type = %#v", body["trigger_type"])
	}
	if triggerType != "api" {
		t.Fatalf("trigger_type = %q, want %q", triggerType, "api")
	}
}

func TestWorkflowsRunsExportRawFlagWritesPlainCSVToStdout(t *testing.T) {
	// Regression: by default, ``export`` returned a JSON envelope
	// (``{"csv_data": "...", "rows": N, "columns": M}``) regardless of
	// --output. Users had to ``jq -r .csv_data`` to get usable CSV. The
	// --raw flag dumps the CSV body straight to stdout.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"csv_data": "id;name\nrun_123;Alice\nrun_456;Bob",
			"rows":     2,
			"columns":  2,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	flags := map[string]string{
		"workflow-id": "wf_123",
		"block-id":    "blk_123",
		"raw":         "true",
	}
	for flag, value := range flags {
		if err := workflowsRunsExportCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsExportCmd, flag) })
	}

	var runErr error
	stdout, stderr := captureStd(t, func() {
		runErr = workflowsRunsExportCmd.RunE(workflowsRunsExportCmd, nil)
	})
	if runErr != nil {
		t.Fatalf("runs export --raw: %v\nstderr:\n%s", runErr, stderr)
	}
	if !strings.HasPrefix(stdout, "id;name\nrun_123;Alice\nrun_456;Bob") {
		t.Fatalf("--raw stdout should be the unwrapped CSV body, got:\n%q", stdout)
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("--raw stdout should end with a trailing newline (terminal friendliness), got:\n%q", stdout)
	}
	// Crucially the JSON envelope keys must NOT appear in stdout.
	for _, leak := range []string{`"csv_data"`, `"rows"`, `"columns"`} {
		if strings.Contains(stdout, leak) {
			t.Fatalf("--raw stdout leaked JSON envelope key %q:\n%s", leak, stdout)
		}
	}
}

func TestWorkflowsRunsListRejectsReversedDateRange(t *testing.T) {
	// Regression: ``runs list --from-date 2026-05-01 --to-date 2026-04-01``
	// used to silently return an empty data array — indistinguishable
	// from "no runs match", which is the most common trigger of the
	// "where did my runs go?" support thread. The fix rejects the
	// reversed range at the CLI before any request leaves the client.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should NOT be reached for reversed date range, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	flags := map[string]string{
		"from-date": "2026-05-01",
		"to-date":   "2026-04-01",
	}
	for flag, value := range flags {
		if err := workflowsRunsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, flag) })
	}

	err := workflowsRunsListCmd.RunE(workflowsRunsListCmd, nil)
	if err == nil {
		t.Fatal("expected reversed-date error, got nil")
	}
	if !strings.Contains(err.Error(), "reversed") {
		t.Fatalf("expected the error to call out the reversed range, got %q", err.Error())
	}
	for _, mustMention := range []string{"--from-date", "--to-date", "2026-05-01", "2026-04-01"} {
		if !strings.Contains(err.Error(), mustMention) {
			t.Fatalf("error %q should mention %q", err.Error(), mustMention)
		}
	}
}

func TestWorkflowsRunsListRejectsReversedCostRange(t *testing.T) {
	// Regression: ``runs list --min-cost 100 --max-cost 1`` previously
	// returned rows verbatim — the server-side cost filter silently
	// accepts swapped bounds. Reject the impossible range at the CLI
	// so the user sees the typo immediately.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should NOT be reached for reversed cost range, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{"min-cost": "100", "max-cost": "1"} {
		if err := workflowsRunsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, flag) })
	}

	err := workflowsRunsListCmd.RunE(workflowsRunsListCmd, nil)
	if err == nil {
		t.Fatal("expected reversed cost-range error, got nil")
	}
	for _, mustMention := range []string{"--min-cost", "--max-cost"} {
		if !strings.Contains(err.Error(), mustMention) {
			t.Fatalf("error %q should mention %q", err.Error(), mustMention)
		}
	}
}

func TestWorkflowsRunsListRejectsReversedDurationRange(t *testing.T) {
	// Mirror of the cost-range test for duration. ``--min-duration 1000
	// --max-duration 100`` is impossible — surface it client-side so
	// the user can fix the typo immediately.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should NOT be reached for reversed duration range, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{"min-duration": "1000", "max-duration": "100"} {
		if err := workflowsRunsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, flag) })
	}

	err := workflowsRunsListCmd.RunE(workflowsRunsListCmd, nil)
	if err == nil {
		t.Fatal("expected reversed duration-range error, got nil")
	}
	for _, mustMention := range []string{"--min-duration", "--max-duration"} {
		if !strings.Contains(err.Error(), mustMention) {
			t.Fatalf("error %q should mention %q", err.Error(), mustMention)
		}
	}
}

func TestWorkflowsRunsListAcceptsEqualDateRange(t *testing.T) {
	// Equal from/to is a legitimate "single-day window" filter — must
	// NOT be rejected as a reversed range. Pin so the validator can't
	// drift to a stricter < (strictly less) check.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "list_metadata": map[string]any{}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	flags := map[string]string{
		"from-date": "2026-05-01",
		"to-date":   "2026-05-01",
	}
	for flag, value := range flags {
		if err := workflowsRunsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, flag) })
	}

	_, _ = captureStd(t, func() {
		if err := workflowsRunsListCmd.RunE(workflowsRunsListCmd, nil); err != nil {
			t.Fatalf("equal from/to should be accepted, got %v", err)
		}
	})
	if hits != 1 {
		t.Fatalf("expected server to be hit once, got %d", hits)
	}
}

func TestWorkflowsRunsExportRejectsBlankStringArrayFlagsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "blank run id", flag: "run-id", wantError: "--run-id must not be blank"},
		{name: "blank preferred column", flag: "preferred-column", wantError: "--preferred-column must not be blank"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for invalid export flag, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			flags := map[string]string{
				"workflow-id": "wf_123",
				"block-id":    "blk_123",
				tc.flag:       "   ",
			}
			for flag, value := range flags {
				if err := workflowsRunsExportCmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("set --%s: %v", flag, err)
				}
				t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsExportCmd, flag) })
			}

			var err error
			_, stderr := captureStd(t, func() {
				err = workflowsRunsExportCmd.RunE(workflowsRunsExportCmd, nil)
			})
			if err == nil {
				t.Fatal("expected blank export flag error")
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
	}
}

func TestParseDocumentArgs_MultipleDocumentFlags(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./a.pdf", "classify=./b.pdf"},
		nil,
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want both keys", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestParseDocumentArgs_LegacyFlagEmitsOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(nil, []string{"start=./a.pdf"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./a.pdf" {
		t.Fatalf("got %v, want {start: ./a.pdf}", got)
	}
	if !strings.Contains(warn.String(), "--document-file is deprecated") {
		t.Fatalf("expected deprecation warning, got %q", warn.String())
	}
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_LegacyMultipleEntriesOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		nil,
		[]string{"start=./a.pdf", "classify=./b.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want both keys", got)
	}
	// Two legacy entries must still produce exactly one warning line.
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_MixedFlagsUnionOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./a.pdf"},
		[]string{"classify=./b.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want union of both keys", got)
	}
	if !strings.Contains(warn.String(), "--document-file is deprecated") {
		t.Fatalf("expected deprecation warning, got %q", warn.String())
	}
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_RejectsCollisionBetweenDocumentAndDocumentFile(t *testing.T) {
	// A given block-id must appear in at most one of `--document` /
	// `--document-file`. Previously the new flag silently won on collision
	// (the loser was discarded with no error), which was surprising and
	// hid user typos. The fix errors instead, naming the offending block.
	var warn bytes.Buffer
	_, err := parseDocumentArgs(
		[]string{"start=./new.pdf"},
		[]string{"start=./legacy.pdf"},
		&warn,
	)
	if err == nil {
		t.Fatal("expected error when --document and --document-file collide on the same block id")
	}
	if !strings.Contains(err.Error(), "start") {
		t.Fatalf("error %q should name the offending block id", err.Error())
	}
	if !strings.Contains(err.Error(), "--document") || !strings.Contains(err.Error(), "--document-file") {
		t.Fatalf("error %q should name both conflicting flags", err.Error())
	}
}

func TestParseDocumentArgs_RejectsDuplicateDocumentFlag(t *testing.T) {
	// Two `--document` flags with the same block id silently overrode each
	// other before the fix. Now the second occurrence is a hard error so
	// the user notices the typo / scripting bug.
	var warn bytes.Buffer
	_, err := parseDocumentArgs(
		[]string{"start=./a.pdf", "start=./b.pdf"},
		nil,
		&warn,
	)
	if err == nil {
		t.Fatal("expected error for duplicate --document block-id")
	}
	if !strings.Contains(err.Error(), "start") {
		t.Fatalf("error %q should name the offending block id", err.Error())
	}
	if !strings.Contains(err.Error(), "--document") {
		t.Fatalf("error %q should name the --document flag", err.Error())
	}
}

func TestParseDocumentArgs_RejectsDuplicateDocumentFileFlag(t *testing.T) {
	var warn bytes.Buffer
	_, err := parseDocumentArgs(
		nil,
		[]string{"start=./a.pdf", "start=./b.pdf"},
		&warn,
	)
	if err == nil {
		t.Fatal("expected error for duplicate --document-file block-id")
	}
	if !strings.Contains(err.Error(), "start") {
		t.Fatalf("error %q should name the offending block id", err.Error())
	}
	if !strings.Contains(err.Error(), "--document-file") {
		t.Fatalf("error %q should name the --document-file flag", err.Error())
	}
}

func TestParseDocumentArgs_NoFlagsEmptyMap(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(nil, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty map", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestParseDocumentArgs_BadShapes(t *testing.T) {
	cases := []struct {
		name string
		docs []string
		legs []string
	}{
		{name: "missing equals on --document", docs: []string{"./a.pdf"}},
		{name: "empty key on --document", docs: []string{"=./a.pdf"}},
		{name: "empty value on --document", docs: []string{"start="}},
		{name: "missing equals on --document-file", legs: []string{"./a.pdf"}},
		{name: "empty key on --document-file", legs: []string{"=./a.pdf"}},
		{name: "empty value on --document-file", legs: []string{"start="}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var warn bytes.Buffer
			_, err := parseDocumentArgs(tc.docs, tc.legs, &warn)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestParseDocumentArgs_NilWarnSinkDoesNotPanic(t *testing.T) {
	// Smoke test: when the legacy flag is used but warnTo is nil (e.g. tests
	// that don't care about warnings), the helper must not panic.
	_, err := parseDocumentArgs(nil, []string{"start=./a.pdf"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Bug B: `workflows runs export` is the only `runs` command that historically
// required `--workflow-id` instead of a positional argument. Migrate the
// command to a positional first argument, keeping the flag as a hidden
// deprecated fallback.
func TestWorkflowsRunsExportAcceptsPositionalWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs/export" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"csv_data": "id\nrun_123\n",
			"rows":     1,
			"columns":  1,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsExportCmd.Flags().Set("block-id", "blk_def"); err != nil {
		t.Fatalf("set --block-id: %v", err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsExportCmd, "block-id") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsRunsExportCmd.RunE(workflowsRunsExportCmd, []string{"wf_abc"})
	})
	if err != nil {
		t.Fatalf("runs export wf_abc: %v\nstderr:\n%s", err, stderr)
	}
	if body["workflow_id"] != "wf_abc" {
		t.Fatalf("workflow_id = %#v, want wf_abc", body["workflow_id"])
	}
	if body["block_id"] != "blk_def" {
		t.Fatalf("block_id = %#v, want blk_def", body["block_id"])
	}
	if strings.Contains(stderr, "deprecated") {
		t.Fatalf("positional invocation should not warn about deprecation, stderr:\n%s", stderr)
	}
}

func TestWorkflowsRunsExportFlagFormStillWorksWithDeprecationWarning(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs/export" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"csv_data": "id\nrun_123\n",
			"rows":     1,
			"columns":  1,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	flags := map[string]string{
		"workflow-id": "wf_abc",
		"block-id":    "blk_def",
	}
	for flag, value := range flags {
		if err := workflowsRunsExportCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsExportCmd, flag) })
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsRunsExportCmd.RunE(workflowsRunsExportCmd, nil)
	})
	if err != nil {
		t.Fatalf("runs export --workflow-id: %v\nstderr:\n%s", err, stderr)
	}
	if body["workflow_id"] != "wf_abc" {
		t.Fatalf("workflow_id = %#v, want wf_abc", body["workflow_id"])
	}
	if !strings.Contains(stderr, "--workflow-id is deprecated") {
		t.Fatalf("stderr should warn about --workflow-id deprecation, got:\n%s", stderr)
	}
}

// Bug C: indentation typo in the second example for `workflows runs restart`.
func TestWorkflowsRunsRestartExampleIndentsSecondCommandLineConsistently(t *testing.T) {
	example := workflowsRunsRestartCmd.Example
	// One-space indented "retab" prefix used to live in the second example line.
	if strings.Contains(example, "\n retab workflows runs restart") {
		t.Fatalf("restart example has a single-space indent on a command line:\n%s", example)
	}
	if !strings.Contains(example, "\n  retab workflows runs restart run_xyz789 --config-source draft") {
		t.Fatalf("restart example should align the draft-config-source example with two spaces:\n%s", example)
	}
}

// Issue 6: “runs create“ and “publish“ help previously showed
// “--version v3“ as the example value. The API rejects “v3“ —
// version must be “production“, “draft“, or a pinned “ver_...“
// id. The example must use one of the accepted forms so a copy-paste
// works against the live API instead of bouncing with a 400.
func TestWorkflowsRunsCreateExampleDoesNotUseInvalidVersionV3(t *testing.T) {
	example := workflowsRunsCreateCmd.Example
	if strings.Contains(example, "--version v3") {
		t.Fatalf("runs create example still shows --version v3, which the API rejects (want production / draft / ver_...):\n%s", example)
	}
	// Sanity: the example must still demonstrate --version on a line
	// somewhere — otherwise the entire "pin a version" example is gone.
	if !strings.Contains(example, "--version") {
		t.Fatalf("runs create example no longer shows --version at all:\n%s", example)
	}
	// The replacement must be one of the three accepted forms.
	accepted := []string{"--version production", "--version draft", "--version ver_"}
	matched := false
	for _, want := range accepted {
		if strings.Contains(example, want) {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("runs create example must use --version production, --version draft, or --version ver_...; got:\n%s", example)
	}
}

func TestWorkflowsPublishExampleDoesNotUseInvalidVersionV3(t *testing.T) {
	example := workflowsPublishCmd.Example
	if strings.Contains(example, "--version v3") {
		t.Fatalf("publish example still shows --version v3, which the API rejects (want production / draft / ver_...):\n%s", example)
	}
	if !strings.Contains(example, "--version") {
		t.Fatalf("publish example no longer shows --version at all:\n%s", example)
	}
	accepted := []string{"--version production", "--version draft", "--version ver_"}
	matched := false
	for _, want := range accepted {
		if strings.Contains(example, want) {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("publish example must use --version production, --version draft, or --version ver_...; got:\n%s", example)
	}
}

// Issue 4: “runs list“ help previously said the positional and
// “--workflow-id“ flag forms could not be passed together, but the
// implementation actually accepts both when they match — only the
// disagreement case errors. Pin the help text to the real behavior so
// the doc stops contradicting the code.
func TestWorkflowsRunsListHelpReflectsImplementedBehavior(t *testing.T) {
	long := workflowsRunsListCmd.Long
	if strings.Contains(long, "not both") {
		t.Fatalf("runs list Long still says \"not both\"; the implementation accepts both when they match, so help must not forbid it:\n%s", long)
	}
	// The help must positively call out that the disagreeing case is
	// the only one that errors — otherwise users still won't know what
	// the rule is. ``disagree`` is what the code comment uses; the
	// help should reflect the same word so doc and code stay in sync.
	if !strings.Contains(long, "disagree") {
		t.Fatalf("runs list Long should explain that only disagreement between positional and --workflow-id errors:\n%s", long)
	}
}

// Issue 4 behavioural pin: passing positional and “--workflow-id“
// with the SAME value must be accepted (no error, no fallback to
// workspace-wide). This is the case the help text used to forbid.
func TestWorkflowsRunsListAcceptsPositionalAndFlagWhenTheyMatch(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var captured string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query().Get("workflow_id")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "list_metadata": map[string]any{}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsListCmd.Flags().Set("workflow-id", "wf_match"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, "workflow-id") })

	_, _ = captureStd(t, func() {
		if err := workflowsRunsListCmd.RunE(workflowsRunsListCmd, []string{"wf_match"}); err != nil {
			t.Fatalf("matching positional + flag should be accepted, got: %v", err)
		}
	})
	if captured != "wf_match" {
		t.Fatalf("server saw workflow_id = %q, want wf_match", captured)
	}
}

// Issue 4 behavioural pin: disagreeing positional + “--workflow-id“
// must still error. The fix only updates the help text; we want a
// regression guard so a future "simplify" pass doesn't drop the
// disagreement check and then quietly start picking one over the other.
func TestWorkflowsRunsListRejectsPositionalAndFlagWhenTheyDisagree(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should NOT be reached for disagreeing positional + flag, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsListCmd.Flags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, "workflow-id") })

	err := workflowsRunsListCmd.RunE(workflowsRunsListCmd, []string{"wf_positional"})
	if err == nil {
		t.Fatal("expected error when positional and --workflow-id disagree")
	}
	if !strings.Contains(err.Error(), "workflow id specified twice") {
		t.Fatalf("error %q should explain the disagreement", err.Error())
	}
}
