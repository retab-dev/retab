package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Cover for warnExportSkippedRunIDs, the reconciliation between the run ids a
// caller named with --run-id and the rows the export could actually contain.
//
// The failure mode this guards is not a crash: it is a CSV that looks complete.
// The envelope's row/column counts cannot reveal a missing run (one run can emit
// several rows), so anything absent has to be named on stderr or it is invisible.

// exportSkippedRun is one run the fake API will serve.
type exportSkippedRun struct {
	workflowID string
	status     string
}

// exportSkippedEnv stands up a fake runs API and returns the stderr produced by
// reconciling runIDs against workflowID.
func exportSkippedEnv(t *testing.T, workflowID string, runs map[string]exportSkippedRun, runIDs []string) string {
	t.Helper()
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		runID := strings.TrimPrefix(r.URL.Path, "/v1/workflows/runs/")
		run, ok := runs[runID]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "Workflow run not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                  runID,
			"workflow_id":         run.workflowID,
			"workflow_version_id": "ver_1",
			"trigger":             map[string]any{"type": "manual"},
			"lifecycle":           map[string]any{"status": run.status},
		})
	}))
	t.Cleanup(server.Close)
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{}
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	client, err := newClient(cmd)
	if err != nil {
		t.Fatal(err)
	}
	warnExportSkippedRunIDs(context.Background(), cmd, client, workflowID, runIDs)
	return stderr.String()
}

// Every distinct reason a requested run can be absent must appear together in one
// note, with a count the caller can reconcile against what they asked for.
func TestExportSkippedRunIDsReportsEveryReasonAtOnce(t *testing.T) {
	note := exportSkippedEnv(t, "wf_mine", map[string]exportSkippedRun{
		"run_ok":        {workflowID: "wf_mine", status: "completed"},
		"run_failed":    {workflowID: "wf_mine", status: "error"},
		"run_cancelled": {workflowID: "wf_mine", status: "cancelled"},
		"run_running":   {workflowID: "wf_mine", status: "running"},
		"run_review":    {workflowID: "wf_mine", status: "awaiting_review"},
		"run_foreign":   {workflowID: "wf_other", status: "completed"},
	}, []string{"run_ok", "run_failed", "run_cancelled", "run_running", "run_review", "run_foreign", "run_missing"})

	for _, want := range []string{
		"run_failed (error)",
		"run_cancelled (cancelled)",
		"run_running (running)",
		"run_review (awaiting_review)",
		"run_foreign (belongs to wf_other)",
		"run_missing (not readable)",
		"6 of 7",
	} {
		if !strings.Contains(note, want) {
			t.Fatalf("note missing %q, got: %q", want, note)
		}
	}
	if strings.Contains(note, "run_ok") {
		t.Fatalf("the one run actually in the export must not be named as skipped, got: %q", note)
	}
}

// A foreign run that is ALSO non-completed must be reported once, and as foreign:
// "belongs to wf_other" tells the caller they pasted an id from the wrong
// workflow, which is actionable; "(error)" would send them to look at a run that
// was never in scope.
func TestExportSkippedRunIDsPrefersTheForeignWorkflowReason(t *testing.T) {
	note := exportSkippedEnv(t, "wf_mine", map[string]exportSkippedRun{
		"run_foreign_failed": {workflowID: "wf_other", status: "error"},
	}, []string{"run_foreign_failed"})

	if !strings.Contains(note, "belongs to wf_other") {
		t.Fatalf("a foreign run should be reported as foreign, got: %q", note)
	}
	if strings.Contains(note, "(error)") {
		t.Fatalf("a foreign run should be reported once, as foreign, not also by status: %q", note)
	}
	if !strings.Contains(note, "1 of 1") {
		t.Fatalf("note should count it exactly once, got: %q", note)
	}
}

// The whole selection being foreign is the paste-the-wrong-ids case: the CSV
// comes back with a header and nothing else, so the note is the ONLY signal.
func TestExportSkippedRunIDsReportsAnEntirelyForeignSelection(t *testing.T) {
	note := exportSkippedEnv(t, "wf_mine", map[string]exportSkippedRun{
		"run_a": {workflowID: "wf_other", status: "completed"},
		"run_b": {workflowID: "wf_other", status: "completed"},
	}, []string{"run_a", "run_b"})

	if !strings.Contains(note, "2 of 2") {
		t.Fatalf("an all-foreign selection must be fully reported, got: %q", note)
	}
	for _, want := range []string{"run_a (belongs to wf_other)", "run_b (belongs to wf_other)"} {
		if !strings.Contains(note, want) {
			t.Fatalf("note missing %q, got: %q", want, note)
		}
	}
}

// A clean selection stays silent. A note printed on every export trains the
// reader to ignore it, which would defeat the whole mechanism.
func TestExportSkippedRunIDsStaysSilentWhenNothingIsSkipped(t *testing.T) {
	note := exportSkippedEnv(t, "wf_mine", map[string]exportSkippedRun{
		"run_a": {workflowID: "wf_mine", status: "completed"},
		"run_b": {workflowID: "wf_mine", status: "completed"},
	}, []string{"run_a", "run_b"})
	if note != "" {
		t.Fatalf("no note expected when every requested run is in the export, got: %q", note)
	}
}

// No --run-id means no explicit selection to reconcile against.
func TestExportSkippedRunIDsStaysSilentWithoutASelection(t *testing.T) {
	for _, runIDs := range [][]string{nil, {}} {
		note := exportSkippedEnv(t, "wf_mine", map[string]exportSkippedRun{}, runIDs)
		if note != "" {
			t.Fatalf("no note expected without --run-id, got: %q", note)
		}
	}
}

// The check costs one lookup per id, so it bows out past
// exportSkippedRunIDCheckLimit. That bound is deliberate, but it means a large
// selection silently loses the reconciliation — pin the boundary so the tradeoff
// stays visible and cannot drift unnoticed.
func TestExportSkippedRunIDsBowsOutPastTheCheckLimit(t *testing.T) {
	runs := map[string]exportSkippedRun{}
	atLimit := make([]string, 0, exportSkippedRunIDCheckLimit)
	for index := 0; index < exportSkippedRunIDCheckLimit; index++ {
		runID := fmt.Sprintf("run_%03d", index)
		runs[runID] = exportSkippedRun{workflowID: "wf_other", status: "completed"}
		atLimit = append(atLimit, runID)
	}

	note := exportSkippedEnv(t, "wf_mine", runs, atLimit)
	if !strings.Contains(note, fmt.Sprintf("%d of %d", exportSkippedRunIDCheckLimit, exportSkippedRunIDCheckLimit)) {
		t.Fatalf("a selection exactly at the limit must still be checked, got: %q", note)
	}

	overLimit := append(append([]string{}, atLimit...), "run_over")
	runs["run_over"] = exportSkippedRun{workflowID: "wf_other", status: "completed"}
	if note := exportSkippedEnv(t, "wf_mine", runs, overLimit); note != "" {
		t.Fatalf("past the limit the check bows out entirely; a partial note would misreport how many runs were skipped, got: %q", note)
	}
}

// The note must name the export's scope accurately. It previously said "covers
// completed runs only", which does not explain a foreign run at all.
func TestExportSkippedRunIDsNoteNamesTheWorkflowScope(t *testing.T) {
	note := exportSkippedEnv(t, "wf_mine", map[string]exportSkippedRun{
		"run_foreign": {workflowID: "wf_other", status: "completed"},
	}, []string{"run_foreign"})

	if !strings.Contains(note, "this workflow's completed runs only") {
		t.Fatalf("note should say the export is scoped to THIS workflow's completed runs, got: %q", note)
	}
}

// A lookup failure must not turn a good CSV into an error: the export already
// succeeded by the time this runs, and the note is a courtesy on top of it.
func TestExportSkippedRunIDsSurvivesAnUnreadableRun(t *testing.T) {
	note := exportSkippedEnv(t, "wf_mine", map[string]exportSkippedRun{
		"run_ok": {workflowID: "wf_mine", status: "completed"},
	}, []string{"run_ok", "run_gone"})

	if !strings.Contains(note, "run_gone (not readable)") {
		t.Fatalf("an unreadable run should be named, got: %q", note)
	}
	if !strings.Contains(note, "1 of 2") {
		t.Fatalf("note should count only the unreadable run, got: %q", note)
	}
}
