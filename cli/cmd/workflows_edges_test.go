package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
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
