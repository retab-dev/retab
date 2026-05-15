package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNonNegativeNumericFlagsRejectNegativeValuesLocally(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		flag string
	}{
		{name: "parses dpi", cmd: parsesCreateCmd, flag: "image-resolution-dpi"},
		{name: "extractions create dpi", cmd: extractionsCreateCmd, flag: "image-resolution-dpi"},
		{name: "extractions create consensus", cmd: extractionsCreateCmd, flag: "n-consensus"},
		{name: "extractions stream dpi", cmd: extractionsStreamCmd, flag: "image-resolution-dpi"},
		{name: "extractions stream consensus", cmd: extractionsStreamCmd, flag: "n-consensus"},
		{name: "classifications consensus", cmd: classificationsCreateCmd, flag: "n-consensus"},
		{name: "classifications first pages", cmd: classificationsCreateCmd, flag: "first-n-pages"},
		{name: "splits consensus", cmd: splitsCreateCmd, flag: "n-consensus"},
		{name: "partitions consensus", cmd: partitionsCreateCmd, flag: "n-consensus"},
		{name: "workflow tests limit", cmd: workflowsTestsListCmd, flag: "limit"},
		{name: "workflow test runs limit", cmd: workflowsTestsRunsListCmd, flag: "limit"},
		{name: "workflow tests consensus", cmd: workflowsTestsExecuteCmd, flag: "n-consensus"},
		{name: "workflow experiments create consensus", cmd: workflowsExperimentsCreateCmd, flag: "n-consensus"},
		{name: "workflow experiments update consensus", cmd: workflowsExperimentsUpdateCmd, flag: "n-consensus"},
		{name: "workflow block simulate consensus", cmd: workflowsBlocksSimulateCmd, flag: "n-consensus"},
		{name: "workflow runs min duration", cmd: workflowsRunsListCmd, flag: "min-duration"},
		{name: "workflow runs max duration", cmd: workflowsRunsListCmd, flag: "max-duration"},
		{name: "workflow runs min cost", cmd: workflowsRunsListCmd, flag: "min-cost"},
		{name: "workflow runs max cost", cmd: workflowsRunsListCmd, flag: "max-cost"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set(tc.flag, "-1")
			if err == nil {
				t.Fatalf("expected local parse error for --%s=-1", tc.flag)
			}
			if !strings.Contains(err.Error(), "non-negative") && !strings.Contains(err.Error(), "0, 3, 5, or 7") {
				t.Fatalf("error %q does not contain a numeric validation hint", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set(tc.flag, "0"); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}
