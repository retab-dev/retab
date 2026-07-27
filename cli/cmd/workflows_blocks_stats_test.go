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

func TestWorkflowsStatsUsesWorkflowEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/wf_123/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer rt_test_key" {
			t.Fatalf("Authorization = %q, want Bearer rt_test_key", got)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflowStatsFixture("wf_123"))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := workflowsStatsCmd.RunE(workflowsStatsCmd, []string{"wf_123"}); err != nil {
			t.Fatalf("workflow stats: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected workflow stats request")
	}
	if !strings.Contains(stdout, `"workflow_id": "wf_123"`) || !strings.Contains(stdout, `"document_shape"`) || !strings.Contains(stdout, `"run_volume"`) {
		t.Fatalf("stdout did not include workflow stats response:\n%s", stdout)
	}
}

func TestWorkflowsStatsGetForwardsWindowFlags(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/wf_123/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflowStatsFixture("wf_123"))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsStatsCmd.PersistentFlags().Set("from", "2026-06-01"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsStatsCmd.PersistentFlags().Set("to", "2026-06-29"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsStatsCmd.PersistentFlags().Set("granularity", "week"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsStatsCmd.PersistentFlags().Set("from", "")
		_ = workflowsStatsCmd.PersistentFlags().Set("to", "")
		_ = workflowsStatsCmd.PersistentFlags().Set("granularity", "")
	})

	captureStd(t, func() {
		if err := workflowsStatsGetCmd.RunE(workflowsStatsGetCmd, []string{"wf_123"}); err != nil {
			t.Fatalf("workflow stats get: %v", err)
		}
	})
	for _, want := range []string{"from=2026-06-01", "to=2026-06-29", "granularity=week"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %s, want %s", gotQuery, want)
		}
	}
}

func TestWorkflowsStatsTableSummarizesWorkflowShape(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflowStatsFixture("wf_123"))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, _ := captureStd(t, func() {
		if err := workflowsStatsCmd.RunE(workflowsStatsCmd, []string{"wf_123"}); err != nil {
			t.Fatalf("workflow stats: %v", err)
		}
	})
	for _, want := range []string{"WORKFLOW", "RUNS", "TOP_FORMAT", "PAGES", "wf_123", "12", "pdf (9)", "2-5 (7)"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	for _, unwanted := range []string{"ERROR", "COST"} {
		if strings.Contains(strings.ToUpper(stdout), unwanted) {
			t.Fatalf("stdout should not expose %s:\n%s", unwanted, stdout)
		}
	}
}

func TestWorkflowsStatsBlocksNamespaceUsesBlockEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_classifier/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("workflow_id") != "wf_123" {
			t.Fatalf("query = %s, want workflow_id=wf_123", r.URL.RawQuery)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(classifierBlockStatsFixture("wf_123", "blk_classifier"))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := workflowsStatsBlocksGetCmd.RunE(workflowsStatsBlocksGetCmd, []string{"wf_123", "blk_classifier"}); err != nil {
			t.Fatalf("workflow stats blocks get: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected block stats request")
	}
	if !strings.Contains(stdout, `"block_type": "classifier"`) || !strings.Contains(stdout, `"classification_categories"`) {
		t.Fatalf("stdout did not include classifier block stats:\n%s", stdout)
	}
}

func TestWorkflowsStatsBlocksNamespaceForwardsWindowFlags(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_classifier/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(classifierBlockStatsFixture("wf_123", "blk_classifier"))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsStatsCmd.PersistentFlags().Set("from", "2026-06-01"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsStatsCmd.PersistentFlags().Set("to", "2026-06-29"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsStatsCmd.PersistentFlags().Set("granularity", "week"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsStatsCmd.PersistentFlags().Set("from", "")
		_ = workflowsStatsCmd.PersistentFlags().Set("to", "")
		_ = workflowsStatsCmd.PersistentFlags().Set("granularity", "")
	})

	captureStd(t, func() {
		if err := workflowsStatsBlocksGetCmd.RunE(workflowsStatsBlocksGetCmd, []string{"wf_123", "blk_classifier"}); err != nil {
			t.Fatalf("workflow stats blocks get: %v", err)
		}
	})
	for _, want := range []string{"workflow_id=wf_123", "from=2026-06-01", "to=2026-06-29", "granularity=week"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %s, want %s", gotQuery, want)
		}
	}
}

func TestWorkflowsStatsBlocksNamespaceAcceptsWorkflowIDFlag(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_classifier/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(classifierBlockStatsFixture("wf_flag", "blk_classifier"))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsStatsBlocksCmd.PersistentFlags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsStatsBlocksCmd.PersistentFlags().Set("workflow-id", "") })

	captureStd(t, func() {
		if err := workflowsStatsBlocksGetCmd.RunE(workflowsStatsBlocksGetCmd, []string{"blk_classifier"}); err != nil {
			t.Fatalf("workflow stats blocks get: %v", err)
		}
	})
	if !strings.Contains(gotQuery, "workflow_id=wf_flag") {
		t.Fatalf("query = %s, want workflow_id=wf_flag", gotQuery)
	}
}

func TestWorkflowsStatsBlocksNamespaceRequiresWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	err := workflowsStatsBlocksCmd.RunE(workflowsStatsBlocksCmd, []string{"blk_classifier"})
	if err == nil {
		t.Fatal("expected missing workflow id error")
	}
	if !strings.Contains(err.Error(), "workflow id is required") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestWorkflowsStatsBlocksNamespaceRejectsConflictingWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsStatsBlocksCmd.PersistentFlags().Set("workflow-id", "wf_flag"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsStatsBlocksCmd.PersistentFlags().Set("workflow-id", "") })

	err := workflowsStatsBlocksGetCmd.RunE(workflowsStatsBlocksGetCmd, []string{"wf_pos", "blk_classifier"})
	if err == nil {
		t.Fatal("expected conflicting workflow id error")
	}
	if !strings.Contains(err.Error(), "conflicting workflow id") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestWorkflowsBlocksStatsGetAcceptsWorkflowIDFlag(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_extract/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(extractBlockStatsFixture("wf_flag", "blk_extract"))
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
	for _, want := range []string{"workflow_id=wf_flag", "from=2026-06-01", "to=2026-06-29", "granularity=week"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %s, want %s", gotQuery, want)
		}
	}
}

func TestWorkflowsBlocksStatsTableUsesOutputShape(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_extract/stats" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(extractBlockStatsFixture("wf_123", "blk_extract"))
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
	for _, want := range []string{"BLOCK", "RUNS", "DETAIL", "TOTAL", "AVG", "blk_extract", "7", "extract", "6", "4.2"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	for _, unwanted := range []string{"ERROR", "P95", "COST"} {
		if strings.Contains(strings.ToUpper(stdout), unwanted) {
			t.Fatalf("stdout should not expose %s:\n%s", unwanted, stdout)
		}
	}
}

func TestWorkflowsBlocksStatsTableSummarizesClassifierSplitAndForEachShapes(t *testing.T) {
	cases := []struct {
		name     string
		blockID  string
		fixture  map[string]any
		command  *cobra.Command
		args     []string
		expected []string
	}{
		{
			name:     "classifier",
			blockID:  "blk_classifier",
			fixture:  classifierBlockStatsFixture("wf_123", "blk_classifier"),
			command:  workflowsBlocksStatsCmd,
			args:     []string{"wf_123", "blk_classifier"},
			expected: []string{"blk_classifier", "classifier", "5", "1"},
		},
		{
			name:     "split",
			blockID:  "blk_split",
			fixture:  splitBlockStatsFixture("wf_123", "blk_split"),
			command:  workflowsBlocksStatsCmd,
			args:     []string{"wf_123", "blk_split"},
			expected: []string{"blk_split", "split", "4", "11", "2.8"},
		},
		{
			name:     "for_each via stats namespace",
			blockID:  "blk_for_each",
			fixture:  forEachBlockStatsFixture("wf_123", "blk_for_each"),
			command:  workflowsStatsBlocksCmd,
			args:     []string{"wf_123", "blk_for_each"},
			expected: []string{"blk_for_each", "for_each", "6", "18", "3"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "rt_test_key")
			t.Setenv("HOME", t.TempDir())

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/"+tc.blockID+"/stats" {
					t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(tc.fixture)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

			stdout, _ := captureStd(t, func() {
				if err := tc.command.RunE(tc.command, tc.args); err != nil {
					t.Fatalf("block stats table: %v", err)
				}
			})
			for _, want := range tc.expected {
				if !strings.Contains(stdout, want) {
					t.Fatalf("stdout missing %q:\n%s", want, stdout)
				}
			}
			for _, unwanted := range []string{"ERROR", "P95", "COST", "99", "1234", "4321"} {
				if strings.Contains(strings.ToUpper(stdout), unwanted) {
					t.Fatalf("stdout should not expose %s:\n%s", unwanted, stdout)
				}
			}
		})
	}
}

func TestWorkflowsBlocksStatsRequiresWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
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
	t.Setenv("RETAB_API_KEY", "rt_test_key")
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

func workflowStatsFixture(workflowID string) map[string]any {
	return map[string]any{
		"workflow_id":  workflowID,
		"query_source": "bigquery",
		"query_status": "ready",
		"analytics": map[string]any{
			"generated_at": "2026-06-29T15:00:00Z",
			"time_range": map[string]any{
				"from":        "2026-06-01T00:00:00Z",
				"to":          "2026-06-29T00:00:00Z",
				"granularity": "day",
			},
			"run_volume": map[string]any{
				"total_runs": 12,
				"series": []map[string]any{{
					"bucket_start": "2026-06-29T00:00:00Z",
					"runs":         3,
				}},
			},
			"document_shape": map[string]any{
				"document_format_distribution": []map[string]any{{
					"bucket":     "pdf",
					"count":      9,
					"percentage": 0.75,
				}},
				"pages_per_document_distribution": []map[string]any{{
					"bucket":     "2-5",
					"count":      7,
					"percentage": 0.58,
				}},
			},
		},
	}
}

func TestWorkflowStatsTopBucketPicksMaxCountNotFirst(t *testing.T) {
	// The distribution is intentionally NOT sorted by count descending; the
	// TOP_FORMAT / PAGES column must still surface the most frequent bucket
	// rather than blindly rendering element [0].
	row := map[string]any{
		"dist": []any{
			map[string]any{"bucket": "png", "count": 3},
			map[string]any{"bucket": "pdf", "count": 9},
			map[string]any{"bucket": "tiff", "count": 5},
		},
	}
	if got := workflowStatsTopBucketCell(row, "dist"); got != "pdf (9)" {
		t.Fatalf("workflowStatsTopBucketCell = %q, want %q", got, "pdf (9)")
	}

	// Counts decoded from JSON arrive as float64 — max-selection must work for
	// that shape too.
	rowF := map[string]any{
		"dist": []any{
			map[string]any{"bucket": "a", "count": float64(2)},
			map[string]any{"bucket": "b", "count": float64(8)},
		},
	}
	if got := workflowStatsTopBucketCell(rowF, "dist"); got != "b (8)" {
		t.Fatalf("workflowStatsTopBucketCell (float) = %q, want %q", got, "b (8)")
	}
}

func classifierBlockStatsFixture(workflowID string, blockID string) map[string]any {
	return map[string]any{
		"block_id":     blockID,
		"workflow_id":  workflowID,
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
			"run_volume": map[string]any{"total_runs": 5, "series": []map[string]any{}},
			"details": map[string]any{
				"block_type":  "classifier",
				"error_count": 99,
				"error_rate":  0.99,
				"cost_value":  1234,
				"p95_ms":      4321,
				"classification_categories": []map[string]any{{
					"category":   "Invoice",
					"count":      4,
					"percentage": 0.8,
				}},
				"category_series": []map[string]any{},
			},
		},
	}
}

func extractBlockStatsFixture(workflowID string, blockID string) map[string]any {
	return map[string]any{
		"block_id":     blockID,
		"workflow_id":  workflowID,
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
			"run_volume": map[string]any{"total_runs": 7, "series": []map[string]any{}},
			"details": map[string]any{
				"block_type":  "extract",
				"error_count": 99,
				"error_rate":  0.99,
				"cost_value":  1234,
				"p95_ms":      4321,
				"output_shape": map[string]any{
					"avg_filled_fields_per_run": 4.2,
					"schema_field_count":        6,
					"filled_field_distribution": []map[string]any{},
				},
				"field_coverage": []map[string]any{},
			},
		},
	}
}

func splitBlockStatsFixture(workflowID string, blockID string) map[string]any {
	return map[string]any{
		"block_id":     blockID,
		"workflow_id":  workflowID,
		"block_type":   "split",
		"query_source": "bigquery",
		"query_status": "ready",
		"analytics": map[string]any{
			"generated_at": "2026-06-29T15:00:00Z",
			"time_range": map[string]any{
				"from":        "2026-06-01T00:00:00Z",
				"to":          "2026-06-29T00:00:00Z",
				"granularity": "day",
			},
			"run_volume": map[string]any{"total_runs": 4, "series": []map[string]any{}},
			"details": map[string]any{
				"block_type":  "split",
				"error_count": 99,
				"error_rate":  0.99,
				"cost_value":  1234,
				"p95_ms":      4321,
				"output_shape": map[string]any{
					"total_subdocuments":                 11,
					"avg_subdocuments_per_run":           2.8,
					"subdocuments_per_run_distribution":  []map[string]any{},
					"pages_per_subdocument_distribution": []map[string]any{},
				},
				"subdocument_categories": []map[string]any{},
			},
		},
	}
}

func forEachBlockStatsFixture(workflowID string, blockID string) map[string]any {
	return map[string]any{
		"block_id":     blockID,
		"workflow_id":  workflowID,
		"block_type":   "for_each",
		"query_source": "bigquery",
		"query_status": "ready",
		"analytics": map[string]any{
			"generated_at": "2026-06-29T15:00:00Z",
			"time_range": map[string]any{
				"from":        "2026-06-01T00:00:00Z",
				"to":          "2026-06-29T00:00:00Z",
				"granularity": "day",
			},
			"run_volume": map[string]any{"total_runs": 6, "series": []map[string]any{}},
			"details": map[string]any{
				"block_type":  "for_each",
				"error_count": 99,
				"error_rate":  0.99,
				"cost_value":  1234,
				"p95_ms":      4321,
				"item_volume": map[string]any{
					"total_items":                18,
					"avg_items_per_run":          3,
					"items_per_run_distribution": []map[string]any{},
					"series":                     []map[string]any{},
				},
			},
		},
	}
}
