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
		if r.URL.Path != "/workflows/runs/run_123" {
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
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs/run_pending/cancel" {
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
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/blocks" && r.URL.Query().Get("workflow_id") == "wf_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "block_generated", "type": "start_document", "label": "Document"},
					{"id": "parse", "type": "parse", "label": "Parse"},
				},
				"list_metadata": map[string]any{},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs":
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
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs" {
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
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs" {
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

func TestWorkflowsRunsCreateResolvesStartAliasFromDocumentsFile(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/blocks" && r.URL.Query().Get("workflow_id") == "wf_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "block_generated", "type": "start_document", "label": "Document"},
					{"id": "extract", "type": "extract", "label": "Extract"},
				},
				"list_metadata": map[string]any{},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs":
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
		if r.Method != http.MethodGet || r.URL.Path != "/workflows/blocks" || r.URL.Query().Get("workflow_id") != "wf_123" {
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

func TestWorkflowsRunsListRejectsInvalidListFlagsLocally(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
		reset     string
	}{
		{name: "negative limit", flag: "limit", value: "-1", wantError: "between 0 and 100", reset: "0"},
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
	if !strings.Contains(err.Error(), "between 0 and 100") {
		t.Fatalf("error %q does not mention backend limit range", err.Error())
	}
	if resetErr := workflowsRunsListCmd.Flags().Set("limit", "0"); resetErr != nil {
		t.Fatalf("reset --limit: %v", resetErr)
	}
}

func TestWorkflowsRunsListRejectsBlankFieldsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached for blank fields flag, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsListCmd.Flags().Set("fields", "   "); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, "fields") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsRunsListCmd.RunE(workflowsRunsListCmd, nil)
	})
	if err == nil {
		t.Fatal("expected blank fields error")
	}
	if !strings.Contains(stderr, "--fields must not be blank") {
		t.Fatalf("stderr %q does not mention blank fields", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
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
		{name: "list invalid statuses", cmd: workflowsRunsListCmd, flag: "statuses", value: "running,banana", wantError: "invalid --statuses"},
		{name: "list invalid exclude status", cmd: workflowsRunsListCmd, flag: "exclude-status", value: "banana", wantError: "invalid --exclude-status"},
		{name: "list invalid trigger type", cmd: workflowsRunsListCmd, flag: "trigger-type", value: "banana", wantError: "invalid --trigger-type"},
		{name: "list invalid trigger types", cmd: workflowsRunsListCmd, flag: "trigger-types", value: "api,banana", wantError: "invalid --trigger-types"},
		{name: "export invalid status", cmd: workflowsRunsExportCmd, flag: "status", value: "banana", wantError: "invalid --status"},
		{name: "export invalid exclude status", cmd: workflowsRunsExportCmd, flag: "exclude-status", value: "banana", wantError: "invalid --exclude-status"},
		{name: "export invalid trigger types", cmd: workflowsRunsExportCmd, flag: "trigger-types", value: "api,banana", wantError: "invalid --trigger-types"},
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

func TestWorkflowsRunsRestartSendsDefaultConfigSource(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "run_456",
			"workflow_id": "wf_123",
			"workflow": map[string]any{
				"workflow_id":       "wf_123",
				"version_id":        "ver_123",
				"name_at_run_time":  "Workflow",
				"requested_version": "production",
			},
			"trigger": map[string]any{"type": "api"},
			"lifecycle": map[string]any{
				"status": "running",
			},
			"timing": map[string]any{
				"created_at": "2026-05-15T00:00:00Z",
			},
			"inputs": map[string]any{
				"documents": map[string]any{},
				"json_data": map[string]any{},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsRestartCmd.Flags().Set("command-id", "cmd_restart"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsRunsRestartCmd.Flags().Set("config-source", "published"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsRunsRestartCmd.Flags().Set("command-id", "")
		_ = workflowsRunsRestartCmd.Flags().Set("config-source", "published")
	})

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
	if body["restart_of"] != "run_123" || body["command_id"] != "cmd_restart" || body["config_source"] != "published" {
		t.Fatalf("restart body = %#v", body)
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
		return
	}
	// Boolean flags reject "" — restore them by writing their default.
	if flag.Value.Type() == "bool" {
		if err := cmd.Flags().Set(name, flag.DefValue); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
		return
	}
	if err := cmd.Flags().Set(name, ""); err != nil {
		t.Fatalf("reset --%s: %v", name, err)
	}
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

func TestWorkflowsStepsGetUsesStepIDRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/workflows/steps/step_123" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"step_id":         "step_123",
			"workflow_run_id": "run_123",
			"block_id":        "blk_123",
			"handle_inputs":   map[string]any{},
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

func TestWorkflowsRunsExportSplitsCommaSeparatedTriggerTypes(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs/export-payload" {
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
		"workflow-id":   "wf_123",
		"block-id":      "blk_123",
		"trigger-types": "api, email",
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
		t.Fatalf("runs export: %v\nstderr:\n%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	triggerTypes, ok := body["trigger_types"].([]any)
	if !ok {
		t.Fatalf("trigger_types = %#v", body["trigger_types"])
	}
	want := []string{"api", "email"}
	if len(triggerTypes) != len(want) {
		t.Fatalf("trigger_types = %#v, want %#v", triggerTypes, want)
	}
	for i, value := range want {
		if triggerTypes[i] != value {
			t.Fatalf("trigger_types = %#v, want %#v", triggerTypes, want)
		}
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

func TestParseDocumentArgs_NewFlagWinsOnCollision(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./new.pdf"},
		[]string{"start=./legacy.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./new.pdf" {
		t.Fatalf("got %v, want {start: ./new.pdf} (--document overrides --document-file)", got)
	}
	// Still exactly one warning line because the legacy flag was used.
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
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
