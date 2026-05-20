package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
			_, _ = w.Write([]byte(`[{"id":"blk_generated_start","type":"start"},{"id":"extract_1","type":"extract"}]`))
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

func TestWorkflowsEdgesCreateResolvesTargetStartAlias(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"extract_1","type":"extract"},{"id":"blk_generated_start","type":"start"}]`))
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

func TestWorkflowsEdgesCreateRejectsAmbiguousStartAlias(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"start_a","type":"start"},{"id":"start_b","type":"start"},{"id":"extract_1","type":"extract"}]`))
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
	if !strings.Contains(stderr, "multiple start blocks") {
		t.Fatalf("stderr %q does not mention multiple start blocks", stderr)
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
			_, _ = w.Write([]byte(`[{"id":"start","type":"document"},{"id":"generated_start","type":"start"},{"id":"extract_1","type":"extract"}]`))
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
