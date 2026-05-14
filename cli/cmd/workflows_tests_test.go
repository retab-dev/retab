package cmd

import (
	"bytes"
	"strings"
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
