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

// TestWorkflowEvalsRunsListTableSurfacesBlockedOutcome pins that a run whose
// only outcome is `blocked` is distinguishable in the table. Blocked is its own
// bucket in evals.ComputeVerdict — the block executed but the assertion could
// not be evaluated (bad relative path after a schema rename, unparseable regex,
// missing output handle) — and it gates CI with a non-zero exit exactly like a
// failure. With only PASSED/FAILED columns such a run rendered as TESTS=1 with
// both tallies blank, reading as "nothing ran" rather than "needs attention".
func TestWorkflowEvalsRunsListTableSurfacesBlockedOutcome(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":          "wfevalrun_blocked",
				"lifecycle":   map[string]any{"status": "completed"},
				"total_evals": 1,
				"counts":      map[string]any{"outcome": map[string]any{"blocked": 1}},
				"timing":      map[string]any{"created_at": "2026-08-25T00:10:53Z"},
			},
		},
	}
	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, workflowEvalRunColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "BLOCKED") {
		t.Fatalf("runs table has no BLOCKED column:\n%s", out)
	}
	// The tally itself must render, not just the header — otherwise the row is
	// still indistinguishable from a run that evaluated nothing.
	header, row, found := strings.Cut(out, "\n")
	if !found {
		t.Fatalf("expected a data row:\n%s", out)
	}
	blockedCol := strings.Index(header, "BLOCKED")
	if blockedCol < 0 || blockedCol >= len(row) {
		t.Fatalf("BLOCKED column %d not present in row %q", blockedCol, row)
	}
	if got := strings.Fields(row[blockedCol:]); len(got) == 0 || got[0] != "1" {
		t.Fatalf("BLOCKED cell = %v, want 1:\n%s", got, out)
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
