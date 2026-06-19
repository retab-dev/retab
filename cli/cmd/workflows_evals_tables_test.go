package cmd

import (
	"strings"
	"testing"
)

// TestWorkflowEvalsRunsListTableRendersTally pins that the dedicated runs-list
// column spec surfaces status + the pass/fail tally instead of the old generic
// ID + (mislabeled) TYPE pair.
func TestWorkflowEvalsRunsListTableRendersTally(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":          "wfevalrun_x",
				"lifecycle":   map[string]any{"status": "completed"},
				"total_evals": 3,
				"counts":      map[string]any{"outcome": map[string]any{"passed": 2, "failed": 1}},
				"timing":      map[string]any{"created_at": "2026-06-14T10:17:07Z"},
			},
		},
	}
	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, workflowEvalRunColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"STATUS", "TESTS", "PASSED", "FAILED", "completed", "wfevalrun_x"} {
		if !strings.Contains(out, want) {
			t.Fatalf("runs table missing %q:\n%s", want, out)
		}
	}
}

// TestWorkflowEvalResultsListTableRendersVerdict pins that the dedicated
// results-list column spec surfaces VERDICT (the key field) — the generic
// renderer dropped it entirely.
func TestWorkflowEvalResultsListTableRendersVerdict(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":         "wfnodeevalrun_y",
				"verdict":    "passed",
				"eval_id":    "wfnodeeval_abc",
				"block_id":   "block_g10h",
				"block_type": "extract",
				"lifecycle":  map[string]any{"status": "completed"},
			},
		},
	}
	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, workflowEvalResultColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"VERDICT", "TARGET", "BLOCK_KIND", "TEST", "passed", "block_g10h", "extract", "wfnodeeval_abc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("results table missing %q:\n%s", want, out)
		}
	}
}
