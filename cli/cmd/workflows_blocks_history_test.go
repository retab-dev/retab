//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestWorkflowsBlocksHistoryUsesDirectEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_extract/config-history" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("workflow_id") != "wf_123" || query.Get("limit") != "20" {
			t.Fatalf("query = %s, want workflow_id=wf_123&limit=20", r.URL.RawQuery)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		sawRequest = true
		writeBlockConfigHistoryTestResponse(t, w)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksHistoryCmd.Flags().Set("limit", "20"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowBlockHistoryFlag(t, workflowsBlocksHistoryCmd, "limit") })

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksHistoryCmd.RunE(workflowsBlocksHistoryCmd, []string{"wf_123", "blk_extract"}); err != nil {
			t.Fatalf("blocks history: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected history request")
	}
	if !strings.Contains(stdout, `"block_config_version_fingerprint": "fp_123"`) || !strings.Contains(stdout, `"run_count": 3`) {
		t.Fatalf("stdout did not include history response:\n%s", stdout)
	}
}

func TestWorkflowsBlocksHistoryListAcceptsWorkflowIDFlagAndCursor(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_extract/config-history" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeBlockConfigHistoryTestResponse(t, w)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksHistoryCmd.PersistentFlags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksHistoryListCmd.Flags().Set("after", "fp_prev"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksHistoryCmd.PersistentFlags().Set("workflow-id", "")
		resetWorkflowBlockHistoryFlag(t, workflowsBlocksHistoryListCmd, "after")
	})

	captureStd(t, func() {
		if err := workflowsBlocksHistoryListCmd.RunE(workflowsBlocksHistoryListCmd, []string{"blk_extract"}); err != nil {
			t.Fatalf("blocks history list: %v", err)
		}
	})
	if !strings.Contains(gotQuery, "workflow_id=wf_flag") || !strings.Contains(gotQuery, "after=fp_prev") {
		t.Fatalf("query = %s, want workflow_id=wf_flag and after=fp_prev", gotQuery)
	}
}

func TestWorkflowsBlocksHistoryRequiresWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	err := workflowsBlocksHistoryCmd.RunE(workflowsBlocksHistoryCmd, []string{"blk_extract"})
	if err == nil {
		t.Fatal("expected missing workflow id error")
	}
	if !strings.Contains(err.Error(), "workflow id is required") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestWorkflowsBlocksHistoryRejectsConflictingWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksHistoryCmd.PersistentFlags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksHistoryCmd.PersistentFlags().Set("workflow-id", "") })

	err := workflowsBlocksHistoryListCmd.RunE(workflowsBlocksHistoryListCmd, []string{"wf_pos", "blk_extract"})
	if err == nil {
		t.Fatal("expected conflicting workflow id error")
	}
	if !strings.Contains(err.Error(), "conflicting workflow id") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestWorkflowsBlocksHistoryRunCountCellPreservesOmission(t *testing.T) {
	if got := blockConfigHistoryCell(workflowBlockConfigHistoryVersion{}, "run_count"); got != "" {
		t.Fatalf("omitted run_count cell = %q, want blank", got)
	}
	if got := blockConfigHistoryCell(workflowBlockConfigHistoryVersion{RunCount: ptr(3)}, "run_count"); got != "3" {
		t.Fatalf("run_count cell = %q, want 3", got)
	}
}

func writeBlockConfigHistoryTestResponse(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": []map[string]any{{
			"id":                               "fp_123",
			"block_config_version_fingerprint": "fp_123",
			"block_type":                       "extract",
			"block_label":                      "Extract",
			"first_seen_at":                    "2026-06-01T00:00:00Z",
			"last_seen_at":                     "2026-06-02T00:00:00Z",
			"run_count":                        3,
			"matches_current_draft":            true,
			"is_current_published":             false,
			"publish_epochs": []map[string]any{{
				"id":                  "epoch_1",
				"publish_version":     1,
				"published_at":        "2026-06-01T00:00:00Z",
				"workflow_version_id": "wver_1",
			}},
		}},
		"list_metadata": map[string]any{},
	})
}

func resetWorkflowBlockHistoryFlag(t *testing.T, cmd *cobra.Command, name string) {
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
