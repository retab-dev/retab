package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// makeCmdWithWorkflowIDFlag mimics the flag wiring used by the workflows
// subcommands that have migrated workflow-id from a required flag to a
// positional argument with a hidden --workflow-id fallback. The flag is
// MarkHidden (not MarkDeprecated) so cobra does NOT emit its own
// auto-deprecation warning that would duplicate the one in
// resolveWorkflowIDArgTo.
func makeCmdWithWorkflowIDFlag() *cobra.Command {
	c := &cobra.Command{Use: "fake"}
	c.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	_ = c.Flags().MarkHidden("workflow-id")
	return c
}

func TestResolveWorkflowIDArg_PositionalAlone(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags(nil); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, []string{"wf_abc"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_abc" {
		t.Fatalf("got %q want wf_abc", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestResolveWorkflowIDArg_FlagAloneEmitsWarning(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags([]string{"--workflow-id", "wf_flag"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_flag" {
		t.Fatalf("got %q want wf_flag", got)
	}
	// Exactly one line, with the "pass as first positional" wording.
	lines := nonEmptyLines(warn.String())
	if len(lines) != 1 {
		t.Fatalf("expected exactly one warning line, got %d: %q", len(lines), warn.String())
	}
	want := "warning: --workflow-id is deprecated; pass the workflow id as the first positional argument"
	if lines[0] != want {
		t.Fatalf("warning mismatch:\n  got:  %q\n  want: %q", lines[0], want)
	}
}

func TestResolveWorkflowIDArg_PositionalWinsOverFlag(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags([]string{"--workflow-id", "wf_flag"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, []string{"wf_pos"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_pos" {
		t.Fatalf("got %q want wf_pos (positional must win)", got)
	}
	lines := nonEmptyLines(warn.String())
	if len(lines) != 1 {
		t.Fatalf("expected exactly one warning line, got %d: %q", len(lines), warn.String())
	}
	want := "warning: --workflow-id is deprecated; positional argument takes precedence"
	if lines[0] != want {
		t.Fatalf("warning mismatch:\n  got:  %q\n  want: %q", lines[0], want)
	}
}

func TestResolveWorkflowIDArg_NeitherErrors(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags(nil); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	_, err := resolveWorkflowIDArgTo(c, nil, &warn)
	if err == nil {
		t.Fatalf("expected error when neither positional nor flag is set")
	}
	if !strings.Contains(err.Error(), "workflow id required") {
		t.Fatalf("expected 'workflow id required' message, got %q", err.Error())
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning when neither is set, got %q", warn.String())
	}
}

// TestWorkflowsTestsCreateCmd_NoCobraAutoDeprecation verifies that the real
// workflowsTestsCreateCmd does NOT trigger cobra's auto-emitted
// "Flag --workflow-id has been deprecated..." message. The custom warning
// in resolveWorkflowIDArg is the single source of truth.
func TestWorkflowsTestsCreateCmd_NoCobraAutoDeprecation(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{"workflowsTestsCreateCmd", workflowsTestsCreateCmd},
		{"workflowsTestsExecuteCmd", workflowsTestsExecuteCmd},
		{"workflowsExperimentsCreateCmd", workflowsExperimentsCreateCmd},
		{"workflowsExperimentsRunBatchCmd", workflowsExperimentsRunBatchCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flag := tc.cmd.Flags().Lookup("workflow-id")
			if flag == nil {
				t.Fatalf("%s: --workflow-id flag is missing — it must still be registered as a hidden fallback", tc.name)
			}
			if flag.Deprecated != "" {
				t.Fatalf("%s: --workflow-id has flag.Deprecated=%q — cobra will auto-emit a duplicate warning. Use MarkHidden, not MarkDeprecated.", tc.name, flag.Deprecated)
			}
			if !flag.Hidden {
				t.Fatalf("%s: --workflow-id should be hidden from --help", tc.name)
			}
		})
	}

	// End-to-end: parse `--workflow-id wf_xyz` through real cobra plumbing
	// on workflowsTestsCreateCmd and capture stderr. We can't actually run
	// the RunE (it would dial the API), so we ParseFlags + invoke the
	// resolver directly the same way RunE does, then assert the captured
	// buffer has only the custom warning.
	cmd := workflowsTestsCreateCmd
	// Suppress cobra usage output during the test.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	// Reset the flag value so a prior test doesn't leak in.
	if err := cmd.Flags().Set("workflow-id", ""); err != nil {
		t.Fatalf("reset workflow-id: %v", err)
	}
	cmd.Flags().Lookup("workflow-id").Changed = false

	if err := cmd.ParseFlags([]string{"--workflow-id", "wf_xyz"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	// Cobra emits its auto-deprecation message during execution
	// (HasFlags + deprecation check). To pin behavior precisely, set the
	// command's stderr sink to a buffer and call the production helper
	// resolveWorkflowIDArgTo with the same buffer. If MarkDeprecated were
	// still in effect, the cobra auto-message would surface on subsequent
	// command execution. Here we verify the static config instead — the
	// table-driven checks above are sufficient to guarantee no auto-message.
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	got, err := resolveWorkflowIDArgTo(cmd, nil, &buf)
	if err != nil {
		t.Fatalf("resolveWorkflowIDArgTo: %v", err)
	}
	if got != "wf_xyz" {
		t.Fatalf("got %q want wf_xyz", got)
	}
	out := buf.String()
	if strings.Contains(out, "Flag --workflow-id has been deprecated") {
		t.Fatalf("cobra auto-deprecation message leaked through:\n%s", out)
	}
	if !strings.Contains(out, "pass the workflow id as the first positional argument") {
		t.Fatalf("expected the custom deprecation warning, got %q", out)
	}
	// Cleanup so other tests in the package don't inherit the flag value.
	if err := cmd.Flags().Set("workflow-id", ""); err != nil {
		t.Fatalf("cleanup reset workflow-id: %v", err)
	}
	cmd.Flags().Lookup("workflow-id").Changed = false
}

func TestWorkflowsTestsExecuteRejectsUnsupportedConsensusLocally(t *testing.T) {
	err := workflowsTestsExecuteCmd.Flags().Set("n-consensus", "2")
	if err == nil {
		t.Fatal("expected local parse error for --n-consensus=2")
	}
	if !strings.Contains(err.Error(), "3, 5, or 7") {
		t.Fatalf("error %q does not mention allowed consensus counts", err.Error())
	}
	if resetErr := workflowsTestsExecuteCmd.Flags().Set("n-consensus", "0"); resetErr != nil {
		t.Fatalf("reset --n-consensus: %v", resetErr)
	}
}

func TestWorkflowsTestsListCommandsRejectNegativeLimitLocally(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "tests list", cmd: workflowsTestsListCmd},
		{name: "test runs list", cmd: workflowsTestsRunsListCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("limit", "-1")
			if err == nil {
				t.Fatal("expected local parse error for --limit=-1")
			}
			if !strings.Contains(err.Error(), "non-negative") {
				t.Fatalf("error %q does not mention non-negative", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("limit", "0"); resetErr != nil {
				t.Fatalf("reset --limit: %v", resetErr)
			}
		})
	}
}

func TestWorkflowsExperimentsConsensusFlagsMatchBackendContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "create", cmd: workflowsExperimentsCreateCmd},
		{name: "update", cmd: workflowsExperimentsUpdateCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("n-consensus", "2")
			if err == nil {
				t.Fatal("expected local parse error for --n-consensus=2")
			}
			if !strings.Contains(err.Error(), "3, 5, or 7") {
				t.Fatalf("error %q does not mention allowed consensus counts", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("n-consensus", "0"); resetErr != nil {
				t.Fatalf("reset --n-consensus: %v", resetErr)
			}
		})
	}
}

func TestWorkflowsExperimentsUnsupportedRunOverrideFlagsAreNotRegistered(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "run-batch", cmd: workflowsExperimentsRunBatchCmd},
		{name: "runs create", cmd: workflowsExperimentsRunsCreateCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if flag := tc.cmd.Flags().Lookup("n-consensus"); flag != nil {
				t.Fatalf("%s should not expose unsupported --n-consensus flag", tc.name)
			}
		})
	}
	if flag := workflowsExperimentsRunsCreateCmd.Flags().Lookup("retry-failed-only"); flag != nil {
		t.Fatalf("runs create should not expose unsupported --retry-failed-only flag")
	}
}

func TestWorkflowsExperimentsMetricsRejectsInvalidViewBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	workflowsExperimentsMetricsCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsExperimentsMetricsCmd.SetContext(nil) })
	if err := workflowsExperimentsMetricsCmd.Flags().Set("view", "banana"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsExperimentsMetricsCmd.Flags().Set("view", "") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsExperimentsMetricsCmd.RunE(workflowsExperimentsMetricsCmd, []string{"wf_123", "exp_123"})
	})
	if err == nil {
		t.Fatal("expected invalid view error")
	}
	if !strings.Contains(stderr, "invalid --view") {
		t.Fatalf("stderr %q does not mention invalid --view", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

func TestWorkflowsExperimentsMetricsViewsMatchBackendContract(t *testing.T) {
	for _, view := range []string{"", "summary", "by_document", "by_target", "votes"} {
		if err := validateExperimentMetricsView(view); err != nil {
			t.Fatalf("view %q should be valid, got %v", view, err)
		}
	}
	for _, view := range []string{"per_run", "per_document", "per_field"} {
		err := validateExperimentMetricsView(view)
		if err == nil {
			t.Fatalf("view %q should be rejected", view)
		}
		if !strings.Contains(err.Error(), "by_document") || !strings.Contains(err.Error(), "votes") {
			t.Fatalf("error %q does not mention backend view names", err.Error())
		}
	}
}

func TestWorkflowsTestsCreateRejectsAssertionMissingTargetBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	sourcePath := filepath.Join(dir, "source.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	if err := os.WriteFile(targetPath, []byte(`{"type":"block","block_id":"block_123"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte(`{"type":"manual","handle_inputs":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assertionPath, []byte(`{"condition":{"kind":"exists"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	for flag, path := range map[string]string{
		"name":           "missing target",
		"target-file":    targetPath,
		"source-file":    sourcePath,
		"assertion-file": assertionPath,
	} {
		if err := workflowsTestsCreateCmd.Flags().Set(flag, path); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = workflowsTestsCreateCmd.Flags().Set(flag, "") })
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsTestsCreateCmd.RunE(workflowsTestsCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected assertion validation error")
	}
	if !strings.Contains(stderr, "--assertion-file: assertion.target is required") {
		t.Fatalf("stderr %q does not mention assertion target", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

// nonEmptyLines splits s on "\n" and returns the non-empty entries. We use
// this to assert exact line counts on warning output without being tripped
// up by the trailing newline from fmt.Fprintln.
func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
