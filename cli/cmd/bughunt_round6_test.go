// Regression tests for the round-6 bug-hunt fixes:
//   - `--json` / `--output-table` shortcuts stamping the output flag's
//     Changed bit                                          (root.go)
//   - Unicode-dash normalization honoring the `--` terminator and leaving
//     a bare Unicode dash alone                            (root.go)
//   - `--where` picking the earliest-occurring operator     (tables.go)
//   - waitPollTracker fatal/transient classification        (primitive_wait.go)
//   - `experiments create --wait` requiring `--run`         (workflows_experiments.go)
//   - `consensus create --inputs ""` not silently reading stdin
//   - `extractions create` rejecting schema+messages both on stdin
//
// The command-level tests exercise commands that live behind oagen build
// tags, so this file carries the matching negations to stay out of those
// prototype builds.
//go:build !retab_oagen_cli_workflows_experiments && !retab_oagen_cli_consensus && !retab_oagen_cli_extractions && !retab_oagen_cli_tables

package cmd

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// findLeafCommandForTest walks the live rootCmd tree by name so tests can
// reset a command's flags without referencing its (build-tagged) symbol.
func findLeafCommandForTest(t *testing.T, path ...string) *cobra.Command {
	t.Helper()
	current := rootCmd
	for _, name := range path {
		var next *cobra.Command
		for _, sub := range current.Commands() {
			if sub.Name() == name {
				next = sub
				break
			}
		}
		if next == nil {
			t.Fatalf("command %q not found under %q", name, current.CommandPath())
		}
		current = next
	}
	return current
}

// TestOutputShortcutsStampChangedBit: the `--json` and `--output-table`
// shortcuts must go through pflag's FlagSet.Set on the real `output` flag so
// its Changed bit is stamped — consumers like explicitOutputJSON gate on
// Changed("output"), and mutating the Value directly left `--json` silently
// ignored there while `--output json` worked.
func TestOutputShortcutsStampChangedBit(t *testing.T) {
	flags := rootCmd.PersistentFlags()
	t.Cleanup(func() { _ = flags.Set("output", "") })

	if err := flags.Set("json", "true"); err != nil {
		t.Fatalf("Set(json, true): %v", err)
	}
	if !flags.Changed("output") {
		t.Fatal("--json must stamp the output flag's Changed bit")
	}
	if got := flags.Lookup("output").Value.String(); got != string(OutputJSON) {
		t.Fatalf("--json should select json output, got %q", got)
	}

	if err := flags.Set("output-table", "true"); err != nil {
		t.Fatalf("Set(output-table, true): %v", err)
	}
	if got := flags.Lookup("output").Value.String(); got != string(OutputTable) {
		t.Fatalf("--output-table should select table output, got %q", got)
	}
}

// TestUnicodeDashNormalizationRespectsTerminator: args after `--` are
// positional by definition and must not be rewritten, and a bare Unicode dash
// must not become a second `--` terminator.
func TestUnicodeDashNormalizationRespectsTerminator(t *testing.T) {
	got := normalizeUnicodeDashArgs([]string{"—output", "json", "--", "–report.txt", "—", "—flag"})
	want := []string{"--output", "json", "--", "–report.txt", "—", "—flag"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeUnicodeDashArgs = %#v, want %#v", got, want)
	}

	if got := normalizeUnicodeDashArg("—"); got != "—" {
		t.Fatalf("a bare em dash must stay untouched, got %q", got)
	}
	if got := normalizeUnicodeDashArg("–"); got != "–" {
		t.Fatalf("a bare en dash must stay untouched, got %q", got)
	}
}

// TestParseTableWhereFlagEarliestOperatorWins: the operator that occurs
// earliest in the string is the intended one — an operator-shaped word later
// in the string belongs to the value, not the column.
func TestParseTableWhereFlagEarliestOperatorWins(t *testing.T) {
	got, err := parseTableWhereFlag("note eq meeting between 3 and 4")
	if err != nil {
		t.Fatalf("parseTableWhereFlag: %v", err)
	}
	want := map[string]any{"column": "note", "operator": "eq", "value": "meeting between 3 and 4"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parse = %#v, want %#v", got, want)
	}

	// The longer variant of an operator still wins over its substring.
	got, err = parseTableWhereFlag("desc not-contains draft")
	if err != nil {
		t.Fatalf("parseTableWhereFlag: %v", err)
	}
	if got["operator"] != "not_contains" || got["column"] != "desc" {
		t.Fatalf("parse = %#v, want not_contains on desc", got)
	}

	// A plain filter still parses as before.
	got, err = parseTableWhereFlag("amount between 3,4")
	if err != nil {
		t.Fatalf("parseTableWhereFlag: %v", err)
	}
	if got["operator"] != "between" {
		t.Fatalf("parse = %#v, want between", got)
	}
}

// TestWaitPollTrackerClassification: auth/validation rejections abort a wait
// immediately, 5xx/transport errors never do, and a 404 gets a short grace
// window that a successful poll resets.
func TestWaitPollTrackerClassification(t *testing.T) {
	apiErr := func(code int) error { return &retab.APIError{StatusCode: code} }

	var tracker waitPollTracker
	if !tracker.fatal(apiErr(http.StatusUnauthorized)) {
		t.Fatal("401 must abort the wait immediately")
	}
	tracker = waitPollTracker{}
	if !tracker.fatal(apiErr(http.StatusUnprocessableEntity)) {
		t.Fatal("422 must abort the wait immediately")
	}

	tracker = waitPollTracker{}
	for i := 0; i < 10; i++ {
		if tracker.fatal(apiErr(http.StatusBadGateway)) {
			t.Fatal("5xx must stay transient")
		}
	}

	tracker = waitPollTracker{}
	if tracker.fatal(apiErr(http.StatusNotFound)) || tracker.fatal(apiErr(http.StatusNotFound)) {
		t.Fatal("the first 404s are read-after-write grace and must not abort")
	}
	if !tracker.fatal(apiErr(http.StatusNotFound)) {
		t.Fatalf("%d consecutive 404s must abort", waitMaxConsecutive404s)
	}

	tracker = waitPollTracker{}
	_ = tracker.fatal(apiErr(http.StatusNotFound))
	_ = tracker.fatal(apiErr(http.StatusNotFound))
	tracker.sawSuccess()
	if tracker.fatal(apiErr(http.StatusNotFound)) {
		t.Fatal("a successful poll must reset the 404 grace window")
	}
}

// TestExperimentsCreateWaitRequiresRun: `--wait` (and the poll-tuning flags)
// without `--run` used to be parsed and silently dropped — the command exited
// 0 without launching any run to wait on.
func TestExperimentsCreateWaitRequiresRun(t *testing.T) {
	create := findLeafCommandForTest(t, "workflows", "experiments", "create")
	t.Cleanup(func() { _ = create.Flags().Set("wait", "false") })

	// runE renders the error to stderr and returns a silent sentinel, so the
	// guidance is asserted on the captured stderr, not on err.Error().
	var err error
	_, stderr := captureStd(t, func() {
		err = runRootForTest(t, "workflows", "experiments", "create", "wf_test123", "--wait")
	})
	if err == nil || !strings.Contains(stderr, "--wait requires --run") {
		t.Fatalf("expected the --wait-without---run guidance, got err=%v stderr=%q", err, stderr)
	}
}

// TestConsensusCreateRejectsBlankInputs: readJSON treats "" like "-" (stdin),
// so `--inputs ""` used to silently block on a terminal waiting for input.
func TestConsensusCreateRejectsBlankInputs(t *testing.T) {
	create := findLeafCommandForTest(t, "consensus", "create")
	t.Cleanup(func() { _ = create.Flags().Set("inputs", "") })

	var err error
	_, stderr := captureStd(t, func() {
		err = runRootForTest(t, "consensus", "create", "--inputs", "")
	})
	if err == nil || !strings.Contains(stderr, "--inputs cannot be blank") {
		t.Fatalf("expected the blank --inputs rejection, got err=%v stderr=%q", err, stderr)
	}
}

// TestExtractionsCreateRejectsDoubleStdin: --json-schema-file and
// --messages-file both accepting "-" used to race for the one stdin, with the
// loser reporting a misleading "empty JSON input".
func TestExtractionsCreateRejectsDoubleStdin(t *testing.T) {
	create := findLeafCommandForTest(t, "extractions", "create")
	t.Cleanup(func() {
		for _, name := range []string{"json-schema-file", "messages-file", "model"} {
			_ = create.Flags().Set(name, "")
		}
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = runRootForTest(t,
			"extractions", "create",
			"--model", "test-model",
			"--json-schema-file", "-",
			"--messages-file", "-",
		)
	})
	if err == nil || !strings.Contains(stderr, "cannot both read from stdin") {
		t.Fatalf("expected the double-stdin rejection, got err=%v stderr=%q", err, stderr)
	}
}
