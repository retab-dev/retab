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
			"analytics": map[string]any{
				"generated_at": "2026-06-29T15:00:00Z",
				"time_range": map[string]any{
					"from":        "2026-06-01T00:00:00Z",
					"to":          "2026-06-29T00:00:00Z",
					"granularity": "day",
				},
				"summary": map[string]any{
					"total_executions": 5,
					"completed_count":  4,
					"error_count":      1,
					"skipped_count":    0,
					"cancelled_count":  0,
					"running_count":    0,
					"completion_rate":  0.8,
					"error_rate":       0.2,
				},
				"time_series":      []map[string]any{},
				"status_breakdown": []map[string]any{},
				"config_versions":  []map[string]any{},
				"error_groups":     []map[string]any{},
			},
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
	if !strings.Contains(stdout, `"query_status": "ready"`) || !strings.Contains(stdout, `"analytics"`) || !strings.Contains(stdout, `"time_series": []`) {
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
	if err := workflowsBlocksStatsCmd.PersistentFlags().Set("from", "2026-06-01"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksStatsCmd.PersistentFlags().Set("to", "2026-06-29"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksStatsCmd.PersistentFlags().Set("granularity", "week"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksStatsCmd.PersistentFlags().Set("workflow-id", "")
		_ = workflowsBlocksStatsCmd.PersistentFlags().Set("from", "")
		_ = workflowsBlocksStatsCmd.PersistentFlags().Set("to", "")
		_ = workflowsBlocksStatsCmd.PersistentFlags().Set("granularity", "")
	})

	captureStd(t, func() {
		if err := workflowsBlocksStatsGetCmd.RunE(workflowsBlocksStatsGetCmd, []string{"blk_extract"}); err != nil {
			t.Fatalf("blocks stats get: %v", err)
		}
	})
	if !strings.Contains(gotQuery, "workflow_id=wf_flag") {
		t.Fatalf("query = %s, want workflow_id=wf_flag", gotQuery)
	}
	if !strings.Contains(gotQuery, "from=2026-06-01") || !strings.Contains(gotQuery, "to=2026-06-29") || !strings.Contains(gotQuery, "granularity=week") {
		t.Fatalf("query = %s, want analytics time-window filters", gotQuery)
	}
}

func TestWorkflowsBlocksStatsTableUsesAnalyticsSummary(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_extract/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"block_id":     "blk_extract",
			"workflow_id":  "wf_123",
			"block_type":   "extract",
			"query_source": "bigquery",
			"query_status": "ready",
			"analytics": map[string]any{
				"generated_at": "2026-06-29T15:00:00Z",
				"time_range": map[string]any{
					"from":        "2026-06-01T00:00:00Z",
					"to":          "2026-06-29T00:00:00Z",
					"granularity": "day",
				},
				"summary": map[string]any{
					"total_executions":    7,
					"completed_count":     5,
					"error_count":         1,
					"skipped_count":       1,
					"cancelled_count":     0,
					"running_count":       0,
					"completion_rate":     0.7142857142857143,
					"error_rate":          0.14285714285714285,
					"p95_duration_ms":     2400,
					"latest_completed_at": "2026-06-29T12:34:56Z",
				},
				"time_series":      []map[string]any{},
				"status_breakdown": []map[string]any{},
				"config_versions":  []map[string]any{},
				"error_groups":     []map[string]any{},
				"extract_stats": map[string]any{
					"fields": []map[string]any{{
						"field_path":    "borrower",
						"present_count": 4,
						"missing_count": 3,
						"fill_rate":     0.5714285714285714,
					}},
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksStatsCmd.RunE(workflowsBlocksStatsCmd, []string{"wf_123", "blk_extract"}); err != nil {
			t.Fatalf("blocks stats: %v", err)
		}
	})
	for _, want := range []string{"EXECUTIONS", "COMPLETED", "ERRORS", "P95_MS", "LATEST_COMPLETED_AT", "7", "5", "1", "2400", "2026-06-29T12:34:56Z"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
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
