package cmd

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// TestEnsureSingleStdinFlag pins the shared guard that rejects an invocation in
// which two or more flags each try to read stdin ("-"). stdin can only be
// consumed once, so the second reader would otherwise get EOF and a misleading
// "empty input" error far from the real cause.
func TestEnsureSingleStdinFlag(t *testing.T) {
	newCmd := func() *cobra.Command {
		c := &cobra.Command{Use: "x"}
		c.Flags().String("a", "", "")
		c.Flags().String("b", "", "")
		c.Flags().String("c", "", "")
		return c
	}

	// No stdin flags → ok.
	if err := ensureSingleStdinFlag(newCmd(), "a", "b", "c"); err != nil {
		t.Fatalf("no stdin flags must be allowed, got %v", err)
	}

	// Exactly one "-" → ok (the common, valid case).
	c := newCmd()
	_ = c.Flags().Set("a", "-")
	if err := ensureSingleStdinFlag(c, "a", "b", "c"); err != nil {
		t.Fatalf("a single stdin flag must be allowed, got %v", err)
	}

	// Two "-" → "both".
	_ = c.Flags().Set("b", "-")
	err := ensureSingleStdinFlag(c, "a", "b", "c")
	if err == nil || !strings.Contains(err.Error(), "cannot both read from stdin") {
		t.Fatalf("two stdin flags must be rejected with a 'both' message, got %v", err)
	}

	// Three "-" → "all".
	_ = c.Flags().Set("c", "-")
	err = ensureSingleStdinFlag(c, "a", "b", "c")
	if err == nil || !strings.Contains(err.Error(), "cannot all read from stdin") {
		t.Fatalf("three stdin flags must be rejected with an 'all' message, got %v", err)
	}

	// A missing flag name is skipped (no panic), and a non-"-" value is ignored,
	// so a superset flag list stays safe for shared builders.
	c2 := newCmd()
	_ = c2.Flags().Set("a", "-")
	_ = c2.Flags().Set("b", "file.json")
	if err := ensureSingleStdinFlag(c2, "a", "b", "missing-flag"); err != nil {
		t.Fatalf("only one real stdin flag here; missing flag must be skipped, got %v", err)
	}
}

// TestExtractionsCreateRejectsDocumentPlusSchemaStdin covers the gap the earlier
// two-flag guard missed: --json-schema-file and --document-file both read stdin,
// so `--json-schema-file - --document-file -` used to slip past the guard and
// fail later with a confusing "empty JSON input".
func TestExtractionsCreateRejectsDocumentPlusSchemaStdin(t *testing.T) {
	create := findLeafCommandForTest(t, "extractions", "create")
	t.Cleanup(func() {
		for _, n := range []string{"json-schema-file", "document-file", "messages-file", "model"} {
			_ = create.Flags().Set(n, "")
		}
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = runRootForTest(t,
			"extractions", "create",
			"--model", "test-model",
			"--json-schema-file", "-",
			"--document-file", "-",
		)
	})
	if err == nil || !strings.Contains(stderr, "cannot both read from stdin") {
		t.Fatalf("expected schema+document double-stdin rejection, got err=%v stderr=%q", err, stderr)
	}
}

// TestWorkflowsRunsCreateRejectsDualStdin covers one of the newly-guarded
// workflow commands: --documents-file and --json-inputs-file both read stdin.
func TestWorkflowsRunsCreateRejectsDualStdin(t *testing.T) {
	create := findLeafCommandForTest(t, "workflows", "runs", "create")
	t.Cleanup(func() {
		for _, n := range []string{"documents-file", "json-inputs-file"} {
			_ = create.Flags().Set(n, "")
		}
	})

	// --confirm keeps the assertion independent of the safety confirm-gate,
	// which (depending on leaked "production" env state from earlier tests in
	// the suite) can otherwise fire before RunE reaches the stdin guard.
	var err error
	_, stderr := captureStd(t, func() {
		err = runRootForTest(t,
			"workflows", "runs", "create", "wf_test123",
			"--confirm",
			"--documents-file", "-",
			"--json-inputs-file", "-",
		)
	})
	if err == nil || !strings.Contains(stderr, "cannot both read from stdin") {
		t.Fatalf("expected documents+json-inputs double-stdin rejection, got err=%v stderr=%q", err, stderr)
	}
}

// TestPrintResultCSVDoesNotTruncate pins that the generic auto-column CSV path
// emits full cell values. It previously reused the table columns' Extract funcs
// verbatim, so any cell longer than the grid width cap (40 runes trailing / 80
// interior) was silently cut and had an ellipsis injected — corrupting the CSV
// on re-import.
func TestPrintResultCSVDoesNotTruncate(t *testing.T) {
	longName := strings.Repeat("A", 60) // > autoTableTruncate (40)
	list := map[string]any{
		"data": []any{
			map[string]any{"id": "wrk_1", "name": longName},
		},
	}

	stdout, _ := captureStd(t, func() {
		if err := printResultCSV(list); err != nil {
			t.Fatalf("printResultCSV failed: %v", err)
		}
	})
	if !strings.Contains(stdout, longName) {
		t.Fatalf("CSV must contain the full untruncated value, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "…") {
		t.Fatalf("CSV must not inject an ellipsis (…), got:\n%s", stdout)
	}

	// Sanity check the scope: the table path still truncates the same value, so
	// this is a CSV-only change, not a regression of the grid's readability.
	tableOut, _ := captureStd(t, func() {
		if err := printResultTable(list); err != nil {
			t.Fatalf("printResultTable failed: %v", err)
		}
	})
	if strings.Contains(tableOut, longName) {
		t.Fatalf("table path should still truncate; unexpectedly held the full value:\n%s", tableOut)
	}
}

// TestWaitPollTrackerNonConsecutive404Reset pins that the 404 read-after-write
// grace is measured in CONSECUTIVE not-found polls, as its name and doc promise.
// A non-404 outcome (5xx or a transport blip) between 404s breaks the streak, so
// a resource that flaps 404/5xx during a redeploy must not fast-abort earlier
// than the documented grace.
func TestWaitPollTrackerNonConsecutive404Reset(t *testing.T) {
	apiErr := func(code int) error { return &retab.APIError{StatusCode: code} }

	// 404, 500, 404, 500, 404 — three NON-consecutive 404s. The interleaved 5xx
	// resets the streak, so no step may declare the wait fatal.
	var tr waitPollTracker
	for i, code := range []int{404, 500, 404, 500, 404} {
		if tr.fatal(apiErr(code)) {
			t.Fatalf("step %d (code %d): non-consecutive 404s must not abort", i, code)
		}
	}

	// A transport (non-API) error likewise breaks the streak.
	var tr2 waitPollTracker
	_ = tr2.fatal(apiErr(http.StatusNotFound))
	_ = tr2.fatal(apiErr(http.StatusNotFound))
	if tr2.fatal(errors.New("dial tcp: connection reset by peer")) {
		t.Fatal("a transport error must never be fatal on its own")
	}
	if tr2.fatal(apiErr(http.StatusNotFound)) {
		t.Fatal("the transport error reset the streak; a single following 404 must not abort")
	}

	// Guard against over-correction: a genuinely stuck id (uninterrupted 404s)
	// must still abort at the documented threshold.
	var tr3 waitPollTracker
	var aborted bool
	for i := 0; i < waitMaxConsecutive404s; i++ {
		aborted = tr3.fatal(apiErr(http.StatusNotFound))
	}
	if !aborted {
		t.Fatalf("%d consecutive 404s must still abort", waitMaxConsecutive404s)
	}
}
