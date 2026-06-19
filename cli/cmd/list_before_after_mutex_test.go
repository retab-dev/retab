//go:build !retab_oagen_cli_workflows && !retab_oagen_cli_workflows_experiments && !retab_oagen_cli_workflows_runs && !retab_oagen_cli_workflows_evals

package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Bug 6 (original): --before / --after were silently accepted together on
// most list commands — the server would pick whichever query string the
// SDK serialized last and drop the other.
//
// Bug 5 (follow-up): the original fix layered cobra's
// `MarkFlagsMutuallyExclusive` on top of the handwritten RunE check. Cobra's
// validateFlagGroups stage fires first, so the user always saw the noisy
// default message ("if any flags in the group [before after] are set
// none of the others can be ...") and never the concise handwritten one.
// We now rely solely on the RunE-time check via `validateBeforeAfterMutex`
// so the user gets `--before and --after are mutually exclusive`.
//
// This test pins the surfaces that survive:
//   - the flag pair is NOT registered with cobra's mutex helper (otherwise
//     the noisy message would shadow the handwritten one again)
//   - the help text still tells the user the two are mutually exclusive
//   - the RunE-level check is exercised by
//     TestListCommandsRunERejectsBeforeAndAfterTogether in
//     list_fields_shape_test.go.
func TestListCommandsDeclareBeforeAfterMutuallyExclusive(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"workflows list", workflowsListCmd},
		{"workflows runs list", workflowsRunsListCmd},
		{"workflows experiments runs list", workflowsExperimentsRunsListCmd},
		{"workflows evals runs list", workflowsEvalsRunsListCmd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			beforeFlag := tc.cmd.Flag("before")
			if beforeFlag == nil {
				t.Fatalf("%s: missing --before flag", tc.name)
			}
			afterFlag := tc.cmd.Flag("after")
			if afterFlag == nil {
				t.Fatalf("%s: missing --after flag", tc.name)
			}
			// Help text should mention they are mutually exclusive, so the
			// user notices the constraint before running the command.
			if !strings.Contains(beforeFlag.Usage, "mutually exclusive") {
				t.Fatalf("%s: --before usage %q does not mention mutually exclusive constraint",
					tc.name, beforeFlag.Usage)
			}
			if !strings.Contains(afterFlag.Usage, "mutually exclusive") {
				t.Fatalf("%s: --after usage %q does not mention mutually exclusive constraint",
					tc.name, afterFlag.Usage)
			}
			// Cobra-level mutex registration must NOT be present. Cobra's
			// validateFlagGroups stage runs before RunE and emits a noisy
			// "if any flags in the group ..." message that hides our
			// handwritten one. The RunE-time `validateBeforeAfterMutex`
			// helper is the single source of the user-facing error.
			if _, ok := beforeFlag.Annotations["cobra_annotation_mutually_exclusive"]; ok {
				t.Fatalf("%s: --before should not have cobra mutex annotation; the handwritten RunE check is the single source of the error message",
					tc.name)
			}
			if _, ok := afterFlag.Annotations["cobra_annotation_mutually_exclusive"]; ok {
				t.Fatalf("%s: --after should not have cobra mutex annotation; the handwritten RunE check is the single source of the error message",
					tc.name)
			}
		})
	}
}
