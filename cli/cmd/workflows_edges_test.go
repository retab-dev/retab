package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

func TestParseEdgeCreateGeneratesIDWhenMissing(t *testing.T) {
	req := parseEdgeCreate(map[string]any{
		"source_block": "start_1",
		"target_block": "extract_1",
	})

	if req.ID == "" {
		t.Fatal("expected parseEdgeCreate to generate an id when none is provided")
	}
}

func TestParseEdgeCreatePreservesExplicitID(t *testing.T) {
	req := parseEdgeCreate(map[string]any{
		"id":           "edge_custom",
		"source_block": "start_1",
		"target_block": "extract_1",
	})

	if req.ID != "edge_custom" {
		t.Fatalf("id = %q, want edge_custom", req.ID)
	}
}

func TestWorkflowsEdgesHelpExplainsDynamicHandleKeys(t *testing.T) {
	helpText := workflowsEdgesCmd.Long + "\n" + workflowsEdgesCreateCmd.Long + "\n" + workflowsEdgesCreateCmd.Example
	for _, want := range []string{
		"handle_key",
		"output-file-booking-confirmation",
		"output-json-needs-review",
		"workflows blocks get",
	} {
		if !strings.Contains(helpText, want) {
			t.Fatalf("edges help should contain %q:\n%s", want, helpText)
		}
	}
	if strings.Contains(workflowsEdgesCreateCmd.Example, "--source-handle true") {
		t.Fatalf("edge create example should use a full canonical output handle:\n%s", workflowsEdgesCreateCmd.Example)
	}
}

func TestWorkflowsEdgesCreateResolvesSourceStartAliasBeforeGeneratingID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"blk_generated_start","type":"start-document"},{"id":"extract_1","type":"extract"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"edge_created","workflow_id":"wf_123","source_block":"blk_generated_start","target_block":"extract_1"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	resetWorkflowEdgesCreateFlags(t)
	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", "start"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "extract_1"); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if got := posted["source_block"]; got != "blk_generated_start" {
		t.Fatalf("source_block = %v, want blk_generated_start", got)
	}
	if got := posted["target_block"]; got != "extract_1" {
		t.Fatalf("target_block = %v, want extract_1", got)
	}
	wantID := defaultWorkflowEdgeID(retabWorkflowEdgeCreateRequestForTest("blk_generated_start", "extract_1"))
	if got := posted["id"]; got != wantID {
		t.Fatalf("id = %v, want %s", got, wantID)
	}
}

func TestWorkflowsEdgesGetHonorsTableOutputFallback(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/workflows/wf_123/edges/edge_1" {
			t.Fatalf("path = %s, want edge get", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "edge_1",
			"workflow_id":   "wf_123",
			"source_block":  "start",
			"target_block":  "split",
			"source_handle": "output-file-0",
			"target_handle": "input-file-0",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsEdgesGetCmd.RunE(workflowsEdgesGetCmd, []string{"wf_123", "edge_1"}); err != nil {
			t.Fatalf("edges get: %v", err)
		}
	})
	if !strings.Contains(stderr, "falling back to json") {
		t.Fatalf("expected table fallback warning, got stderr %q", stderr)
	}
	if !strings.Contains(stdout, `"id": "edge_1"`) {
		t.Fatalf("expected JSON fallback payload, got:\n%s", stdout)
	}
}

func TestWorkflowsEdgesCreateResolvesTargetStartAlias(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"extract_1","type":"extract"},{"id":"blk_generated_start","type":"start-document"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"edge_created","workflow_id":"wf_123","source_block":"extract_1","target_block":"blk_generated_start"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	resetWorkflowEdgesCreateFlags(t)
	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", "extract_1"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "start"); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if got := posted["target_block"]; got != "blk_generated_start" {
		t.Fatalf("target_block = %v, want blk_generated_start", got)
	}
}

func TestWorkflowsEdgesCreateResolvesFriendlyTargetInputHandle(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"id":"blk_generated_start","type":"start-document"},
				{"id":"extract_1","type":"extract","config":{"inputs":[{"name":"document","type":"file","is_primary":true}]}}
			]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"edge_created","workflow_id":"wf_123","source_block":"blk_generated_start","target_block":"extract_1"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	resetWorkflowEdgesCreateFlags(t)
	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", "start"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("source-handle", "output-file-0"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "extract_1"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-handle", "document"); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if got := posted["source_block"]; got != "blk_generated_start" {
		t.Fatalf("source_block = %v, want blk_generated_start", got)
	}
	if got := posted["target_handle"]; got != "input-file-document" {
		t.Fatalf("target_handle = %v, want input-file-document", got)
	}
	wantID := defaultWorkflowEdgeID(retab.WorkflowEdgeCreateRequest{
		SourceBlock:  "blk_generated_start",
		SourceHandle: "output-file-0",
		TargetBlock:  "extract_1",
		TargetHandle: "input-file-document",
	})
	if got := posted["id"]; got != wantID {
		t.Fatalf("id = %v, want %s", got, wantID)
	}
}

func TestWorkflowsEdgesCreateResolvesDocumentHandleForDefaultFileInput(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"id":"blk_generated_start","type":"start-document"},
				{"id":"split_1","type":"split","config":{"subdocuments":[{"name":"invoice"}]}}
			]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"edge_created","workflow_id":"wf_123","source_block":"blk_generated_start","target_block":"split_1"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	resetWorkflowEdgesCreateFlags(t)
	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", "start"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("source-handle", "output-file-0"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "split_1"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-handle", "document"); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if got := posted["target_handle"]; got != "input-file-0" {
		t.Fatalf("target_handle = %v, want input-file-0", got)
	}
}

func TestWorkflowsEdgesCreateResolvesDocumentHandleForClassifier(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"id":"blk_generated_start","type":"start-document"},
				{"id":"classifier_1","type":"classifier","config":{"categories":[{"name":"invoice"}]}}
			]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"edge_created","workflow_id":"wf_123","source_block":"blk_generated_start","target_block":"classifier_1"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	resetWorkflowEdgesCreateFlags(t)
	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", "start"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("source-handle", "output-file-0"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "classifier_1"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-handle", "document"); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if got := posted["target_handle"]; got != "input-file-document" {
		t.Fatalf("target_handle = %v, want input-file-document", got)
	}
}

func TestWorkflowsEdgesCreateHelpExplainsDocumentHandleAliases(t *testing.T) {
	help := workflowsEdgesCreateCmd.Long + "\n" + workflowsEdgesCreateCmd.Example

	for _, want := range []string{
		"extract/classifier",
		"split/classifier",
		"input-file-document",
		"split",
		"input-file-0",
		"--target-handle document",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("edge create help should mention %q, got:\n%s", want, help)
		}
	}
	if strings.Contains(help, "split/classify") {
		t.Fatalf("edge create help should use classifier block naming, got:\n%s", help)
	}
}

func TestWorkflowsEdgesListTableIncludesHandles(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/edges":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"data": [
					{
						"id": "edge_1",
						"workflow_id": "wf_123",
						"source_block": "start_1",
						"source_handle": "output-file-0",
						"target_block": "classifier_1",
						"target_handle": "input-file-document",
						"updated_at": "2026-05-20T16:00:00Z"
					}
				],
				"list_metadata": {}
			}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsEdgesListCmd.RunE(workflowsEdgesListCmd, []string{"wf_123"}); err != nil {
			t.Fatalf("edges list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	for _, want := range []string{
		"ID", "SOURCE_BLOCK", "SOURCE_HANDLE", "TARGET_BLOCK", "TARGET_HANDLE", "UPDATED_AT",
		"edge_1", "start_1", "output-file-0", "classifier_1", "input-file-document",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in edges table:\n%s", want, stdout)
		}
	}
}

func TestWorkflowsEdgesCreateBatchResolvesFriendlyAliasesBeforeGeneratingIDs(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	edgesPath := filepath.Join(t.TempDir(), "edges.json")
	if err := os.WriteFile(edgesPath, []byte(`[
		{"source_block":"start","source_handle":"output-file-0","target_block":"extract_1","target_handle":"document"},
		{"source_block":"start","source_handle":"output-file-0","target_block":"split_1","target_handle":"document"}
	]`), 0o600); err != nil {
		t.Fatal(err)
	}

	var posted []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"id":"blk_generated_start","type":"start-document"},
				{"id":"extract_1","type":"extract","config":{"inputs":[{"name":"document","type":"file","is_primary":true}]}},
				{"id":"split_1","type":"split","config":{"subdocuments":[{"name":"invoice"}]}}
			]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges/batch":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := workflowsEdgesCreateBatchCmd.Flags().Set("edges-file", edgesPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsEdgesCreateBatchCmd.Flags().Set("edges-file", "") })

	var err error
	captureStd(t, func() {
		err = workflowsEdgesCreateBatchCmd.RunE(workflowsEdgesCreateBatchCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if len(posted) != 2 {
		t.Fatalf("posted %d edges, want 2: %#v", len(posted), posted)
	}
	if got := posted[0]["source_block"]; got != "blk_generated_start" {
		t.Fatalf("edge 0 source_block = %v, want blk_generated_start", got)
	}
	if got := posted[0]["target_handle"]; got != "input-file-document" {
		t.Fatalf("edge 0 target_handle = %v, want input-file-document", got)
	}
	if got := posted[1]["source_block"]; got != "blk_generated_start" {
		t.Fatalf("edge 1 source_block = %v, want blk_generated_start", got)
	}
	if got := posted[1]["target_handle"]; got != "input-file-0" {
		t.Fatalf("edge 1 target_handle = %v, want input-file-0", got)
	}
	for i, req := range []retab.WorkflowEdgeCreateRequest{
		{
			SourceBlock:  "blk_generated_start",
			SourceHandle: "output-file-0",
			TargetBlock:  "extract_1",
			TargetHandle: "input-file-document",
		},
		{
			SourceBlock:  "blk_generated_start",
			SourceHandle: "output-file-0",
			TargetBlock:  "split_1",
			TargetHandle: "input-file-0",
		},
	} {
		if got, want := posted[i]["id"], defaultWorkflowEdgeID(req); got != want {
			t.Fatalf("edge %d id = %v, want %s", i, got, want)
		}
	}
}

func TestWorkflowsEdgesCreateRejectsAmbiguousStartAlias(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"start_a","type":"start-document"},{"id":"start_b","type":"start-document"},{"id":"extract_1","type":"extract"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges":
			postHits.Add(1)
			http.Error(w, "server should not be reached", http.StatusInternalServerError)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	resetWorkflowEdgesCreateFlags(t)
	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", "start"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "extract_1"); err != nil {
		t.Fatal(err)
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected ambiguous start alias error")
	}
	if !strings.Contains(stderr, "multiple start-document blocks") {
		t.Fatalf("stderr %q does not mention multiple start-document blocks", stderr)
	}
	if got := postHits.Load(); got != 0 {
		t.Fatalf("edge create endpoint was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsEdgesCreateKeepsLiteralStartBlockID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"start","type":"document"},{"id":"generated_start","type":"start-document"},{"id":"extract_1","type":"extract"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/edges":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"edge_created","workflow_id":"wf_123","source_block":"start","target_block":"extract_1"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	resetWorkflowEdgesCreateFlags(t)
	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", "start"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "extract_1"); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if got := posted["source_block"]; got != "start" {
		t.Fatalf("source_block = %v, want literal start", got)
	}
}

func TestWorkflowsEdgesCreateRejectsEmptyEndpointsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := workflowsEdgesCreateCmd.Flags().Set("source-block", ""); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateCmd.Flags().Set("target-block", "block_b"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsEdgesCreateCmd.Flags().Set("source-block", "")
		_ = workflowsEdgesCreateCmd.Flags().Set("target-block", "")
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsEdgesCreateCmd.RunE(workflowsEdgesCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected empty source block error")
	}
	if !strings.Contains(stderr, "source_block is required") {
		t.Fatalf("stderr %q does not mention source_block", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func resetWorkflowEdgesCreateFlags(t *testing.T) {
	t.Helper()
	for name, value := range map[string]string{
		"id":            "",
		"source-block":  "",
		"target-block":  "",
		"source-handle": "",
		"target-handle": "",
	} {
		if err := workflowsEdgesCreateCmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for name := range map[string]string{
			"id":            "",
			"source-block":  "",
			"target-block":  "",
			"source-handle": "",
			"target-handle": "",
		} {
			_ = workflowsEdgesCreateCmd.Flags().Set(name, "")
		}
	})
}

func retabWorkflowEdgeCreateRequestForTest(sourceBlock string, targetBlock string) retab.WorkflowEdgeCreateRequest {
	return retab.WorkflowEdgeCreateRequest{
		SourceBlock: sourceBlock,
		TargetBlock: targetBlock,
	}
}

func TestWorkflowsEdgesCreateBatchRejectsEmptyEndpointsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	edgesFile := t.TempDir() + "/edges.json"
	if err := os.WriteFile(edgesFile, []byte(`[{"source_block":"","target_block":"block_b"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEdgesCreateBatchCmd.Flags().Set("edges-file", edgesFile); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsEdgesCreateBatchCmd.Flags().Set("edges-file", "") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsEdgesCreateBatchCmd.RunE(workflowsEdgesCreateBatchCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected empty source block error")
	}
	if !strings.Contains(stderr, "--edges-file[0]: source_block is required") {
		t.Fatalf("stderr %q does not mention indexed source_block", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsEdgesCreateBatchReadsLocalFileBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-edges.json")
	if err := workflowsEdgesCreateBatchCmd.Flags().Set("edges-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsEdgesCreateBatchCmd.Flags().Set("edges-file", "") })

	err := workflowsEdgesCreateBatchCmd.RunE(workflowsEdgesCreateBatchCmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected missing edges file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "missing-edges.json") {
		t.Fatalf("error should mention missing edges file, got %q", err.Error())
	}
}
