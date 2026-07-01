//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func TestWorkflowsBlocksRunsListsStepsForBlock(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/steps" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("block_id") != "blk_extract" || query.Get("limit") != "25" {
			t.Fatalf("query = %s, want block_id=blk_extract&limit=25", r.URL.RawQuery)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		sawRequest = true
		writeBlockRunsTestResponse(t, w)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksRunsCmd.Flags().Set("limit", "25"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowBlockRunsFlag(t, workflowsBlocksRunsCmd, "limit") })

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksRunsCmd.RunE(workflowsBlocksRunsCmd, []string{"blk_extract"}); err != nil {
			t.Fatalf("blocks runs: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected steps request")
	}
	if !strings.Contains(stdout, `"step_id": "step_1"`) || !strings.Contains(stdout, `"block_id": "blk_extract"`) {
		t.Fatalf("stdout did not include step response:\n%s", stdout)
	}
}

func TestWorkflowsBlocksRunsListPassesRunAndStatusFilters(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/steps" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeBlockRunsTestResponse(t, w)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksRunsListCmd.Flags().Set("run-id", "run_123"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksRunsListCmd.Flags().Set("status", "completed"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		resetWorkflowBlockRunsFlag(t, workflowsBlocksRunsListCmd, "run-id")
		resetWorkflowBlockRunsFlag(t, workflowsBlocksRunsListCmd, "status")
	})

	captureStd(t, func() {
		if err := workflowsBlocksRunsListCmd.RunE(workflowsBlocksRunsListCmd, []string{"blk_extract"}); err != nil {
			t.Fatalf("blocks runs list: %v", err)
		}
	})
	for _, want := range []string{"block_id=blk_extract", "run_id=run_123", "status=completed"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %s, missing %s", gotQuery, want)
		}
	}
}

func TestWorkflowsBlocksRunsRejectsInvalidStatus(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksRunsListCmd.Flags().Set("status", "definitely_not_a_status"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowBlockRunsFlag(t, workflowsBlocksRunsListCmd, "status") })

	err := workflowsBlocksRunsListCmd.RunE(workflowsBlocksRunsListCmd, []string{"blk_extract"})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
	if !strings.Contains(err.Error(), "invalid --status") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestWorkflowsBlocksRunsVerifiesWorkflowScopedBlock(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var blockGetCount atomic.Int32
	var stepsCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/blocks/blk_extract":
			blockGetCount.Add(1)
			if r.URL.Query().Get("workflow_id") != "wf_123" {
				t.Fatalf("block get query = %s, want workflow_id=wf_123", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_extract",
				"workflow_id": "wf_123",
				"type":        "extract",
				"label":       "Extract",
				"updated_at":  "2026-06-01T00:00:00Z",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/steps":
			stepsCount.Add(1)
			if r.URL.Query().Get("block_id") != "blk_extract" {
				t.Fatalf("steps query = %s, want block_id=blk_extract", r.URL.RawQuery)
			}
			if r.URL.Query().Has("workflow_id") {
				t.Fatalf("steps query = %s, did not expect workflow_id", r.URL.RawQuery)
			}
			writeBlockRunsTestResponse(t, w)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	captureStd(t, func() {
		if err := workflowsBlocksRunsListCmd.RunE(workflowsBlocksRunsListCmd, []string{"wf_123", "blk_extract"}); err != nil {
			t.Fatalf("blocks runs list: %v", err)
		}
	})
	if blockGetCount.Load() == 0 {
		t.Fatal("expected workflow-scoped block get before listing steps")
	}
	if stepsCount.Load() == 0 {
		t.Fatal("expected steps request")
	}
}

func TestWorkflowsBlocksRunsRejectsConflictingWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksRunsCmd.PersistentFlags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksRunsCmd.PersistentFlags().Set("workflow-id", "") })

	err := workflowsBlocksRunsListCmd.RunE(workflowsBlocksRunsListCmd, []string{"wf_pos", "blk_extract"})
	if err == nil {
		t.Fatal("expected conflicting workflow id error")
	}
	if !strings.Contains(err.Error(), "conflicting workflow id") {
		t.Fatalf("error = %q", err.Error())
	}
}

func writeBlockRunsTestResponse(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": []map[string]any{{
			"block_id":     "blk_extract",
			"step_id":      "step_1",
			"block_type":   "extract",
			"block_label":  "Extract",
			"lifecycle":    map[string]any{"status": "completed"},
			"run_id":       "run_123",
			"started_at":   "2026-06-01T00:00:00Z",
			"completed_at": "2026-06-01T00:00:05Z",
			"artifact":     map[string]any{"operation": "extraction", "id": "extr_123"},
			"retry_count":  1,
		}},
		"list_metadata": map[string]any{},
	})
}

func resetWorkflowBlockRunsFlag(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing flag %s", name)
	}
	if value, ok := flag.Value.(*boundedIntFlagValue); ok {
		value.value = ""
		flag.Changed = false
		return
	}
	if err := flag.Value.Set(flag.DefValue); err != nil {
		t.Fatalf("reset --%s: %v", name, err)
	}
	flag.Changed = false
}
