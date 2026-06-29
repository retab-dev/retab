//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkflowsBlocksStatsUsesDirectEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_classifier/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("workflow_id") != "wf_123" {
			t.Fatalf("query = %s, want workflow_id=wf_123", r.URL.RawQuery)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"block_id":     "blk_classifier",
			"workflow_id":  "wf_123",
			"block_type":   "classifier",
			"query_source": "bigquery",
			"query_status": "ready",
			"classifier_stats": map[string]any{
				"total_executions":         5,
				"uncategorized_executions": 1,
				"categories": []map[string]any{{
					"category":          "Invoice",
					"handle_key":        "invoice",
					"execution_count":   4,
					"execution_percent": 0.8,
				}},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksStatsCmd.RunE(workflowsBlocksStatsCmd, []string{"wf_123", "blk_classifier"}); err != nil {
			t.Fatalf("blocks stats: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected stats request")
	}
	if !strings.Contains(stdout, `"query_status": "ready"`) || !strings.Contains(stdout, `"total_executions": 5`) {
		t.Fatalf("stdout did not include stats response:\n%s", stdout)
	}
}

func TestWorkflowsBlocksStatsGetAcceptsWorkflowIDFlag(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_extract/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"block_id":     "blk_extract",
			"workflow_id":  "wf_flag",
			"block_type":   "extract",
			"query_source": "bigquery",
			"query_status": "unsupported",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksStatsCmd.PersistentFlags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksStatsCmd.PersistentFlags().Set("workflow-id", "") })

	captureStd(t, func() {
		if err := workflowsBlocksStatsGetCmd.RunE(workflowsBlocksStatsGetCmd, []string{"blk_extract"}); err != nil {
			t.Fatalf("blocks stats get: %v", err)
		}
	})
	if !strings.Contains(gotQuery, "workflow_id=wf_flag") {
		t.Fatalf("query = %s, want workflow_id=wf_flag", gotQuery)
	}
}

func TestWorkflowsBlocksStatsRequiresWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	err := workflowsBlocksStatsCmd.RunE(workflowsBlocksStatsCmd, []string{"blk_extract"})
	if err == nil {
		t.Fatal("expected missing workflow id error")
	}
	if !strings.Contains(err.Error(), "workflow id is required") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestWorkflowsBlocksStatsRejectsConflictingWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksStatsCmd.PersistentFlags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksStatsCmd.PersistentFlags().Set("workflow-id", "") })

	err := workflowsBlocksStatsGetCmd.RunE(workflowsBlocksStatsGetCmd, []string{"wf_pos", "blk_extract"})
	if err == nil {
		t.Fatal("expected conflicting workflow id error")
	}
	if !strings.Contains(err.Error(), "conflicting workflow id") {
		t.Fatalf("error = %q", err.Error())
	}
}
