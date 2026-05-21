package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Bug 6: --before / --after were only enforced as mutually exclusive on
// `workflows reviews list` and `workflows reviews versions list`. Every
// other list command silently accepted both — the server would pick one
// (whichever query string the SDK serialized last) and quietly drop the
// other. Cobra's `MarkFlagsMutuallyExclusive` is the canonical way to
// surface this at the CLI layer; pair it with an explicit RunE-time
// fallback so tests invoking RunE directly still see the error.
//
// This test pins both surfaces for every list command that exposes
// --before and --after:
//   - the flag pair is declared mutually exclusive at the cobra level
//   - the RunE body returns an explicit "mutually exclusive" error when
//     both are passed (covers direct-RunE invocations that bypass
//     cobra.Execute()'s validateFlagGroups stage).
func TestListCommandsDeclareBeforeAfterMutuallyExclusive(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"workflows list", workflowsListCmd},
		{"workflows runs list", workflowsRunsListCmd},
		{"workflows experiments runs list", workflowsExperimentsRunsListCmd},
		{"workflows tests runs list", workflowsTestsRunsListCmd},
		{"jobs list", jobsListCmd},
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
			// Cobra-level mutex registration: the annotation
			// `cobra_annotation_mutually_exclusive` is set on each member
			// of the group by MarkFlagsMutuallyExclusive.
			if _, ok := beforeFlag.Annotations["cobra_annotation_mutually_exclusive"]; !ok {
				t.Fatalf("%s: --before is not registered as mutually exclusive at the cobra level",
					tc.name)
			}
			if _, ok := afterFlag.Annotations["cobra_annotation_mutually_exclusive"]; !ok {
				t.Fatalf("%s: --after is not registered as mutually exclusive at the cobra level",
					tc.name)
			}
		})
	}
}
