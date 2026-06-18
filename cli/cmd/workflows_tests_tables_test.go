package cmd

import (
	"strings"
	"testing"
)

// TestWorkflowTestsRunsListTableRendersTally pins that the dedicated runs-list
// column spec surfaces status + the pass/fail tally instead of the old generic
// ID + (mislabeled) TYPE pair.
func TestWorkflowTestsRunsListTableRendersTally(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":          "wftestrun_x",
				"lifecycle":   map[string]any{"status": "completed"},
				"total_tests": 3,
				"counts":      map[string]any{"outcome": map[string]any{"passed": 2, "failed": 1}},
				"timing":      map[string]any{"created_at": "2026-06-14T10:17:07Z"},
			},
		},
	}
	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, workflowTestRunColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"STATUS", "TESTS", "PASSED", "FAILED", "completed", "wftestrun_x"} {
		if !strings.Contains(out, want) {
			t.Fatalf("runs table missing %q:\n%s", want, out)
		}
	}
}

// TestWorkflowTestResultsListTableRendersVerdict pins that the dedicated
// results-list column spec surfaces VERDICT (the key field) — the generic
// renderer dropped it entirely.
func TestWorkflowTestResultsListTableRendersVerdict(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":         "wfnodetestrun_y",
				"verdict":    "passed",
				"test_id":    "wfnodetest_abc",
				"block_id":   "block_g10h",
				"block_type": "extract",
				"lifecycle":  map[string]any{"status": "completed"},
			},
		},
	}
	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, workflowTestResultColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"VERDICT", "TARGET", "BLOCK_KIND", "TEST", "passed", "block_g10h", "extract", "wfnodetest_abc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("results table missing %q:\n%s", want, out)
		}
	}
}
