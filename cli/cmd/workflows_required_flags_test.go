package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkflowsBlocksSimulateRejectsBlankRequiredStringsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		t.Fatalf("server should not be reached for blank simulate flag, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := workflowsBlocksSimulateCmd.Flags().Set("run-id", "   "); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksSimulateCmd.Flags().Set("block-id", "blk_123"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksSimulateCmd.Flags().Set("run-id", "")
		_ = workflowsBlocksSimulateCmd.Flags().Set("block-id", "")
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsBlocksSimulateCmd.RunE(workflowsBlocksSimulateCmd, nil)
	})
	if err == nil {
		t.Fatal("expected blank run-id error")
	}
	if !strings.Contains(stderr, "--run-id must not be blank") {
		t.Fatalf("stderr %q does not mention blank run-id", stderr)
	}
	if hits != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits)
	}
}

func TestWorkflowsExperimentsRunBatchRejectsBlankBlockIDBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		t.Fatalf("server should not be reached for blank experiment flag, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := workflowsExperimentsRunBatchCmd.Flags().Set("block-id", "   "); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsExperimentsRunBatchCmd.Flags().Set("block-id", "")
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsExperimentsRunBatchCmd.RunE(workflowsExperimentsRunBatchCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected blank block-id error")
	}
	if !strings.Contains(stderr, "--block-id must not be blank") {
		t.Fatalf("stderr %q does not mention blank block-id", stderr)
	}
	if hits != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits)
	}
}
