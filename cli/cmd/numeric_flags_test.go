package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNonNegativeNumericFlagsRejectNegativeValuesLocally(t *testing.T) {
	cases := []struct {
		name      string
		cmd       *cobra.Command
		flag      string
		reset     string
		wantError string
	}{
		{name: "parses dpi", cmd: parsesCreateCmd, flag: "image-resolution-dpi", reset: "96", wantError: "between"},
		{name: "extractions create dpi", cmd: extractionsCreateCmd, flag: "image-resolution-dpi", reset: "96", wantError: "between"},
		{name: "extractions create consensus", cmd: extractionsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "extractions stream dpi", cmd: extractionsStreamCmd, flag: "image-resolution-dpi", reset: "96", wantError: "between"},
		{name: "extractions stream consensus", cmd: extractionsStreamCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "classifications consensus", cmd: classificationsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "classifications first pages", cmd: classificationsCreateCmd, flag: "first-n-pages", reset: "1", wantError: "positive"},
		{name: "splits consensus", cmd: splitsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "partitions consensus", cmd: partitionsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "files create-upload size", cmd: filesCreateUploadCmd, flag: "size-bytes"},
		{name: "workflow tests limit", cmd: workflowsTestsListCmd, flag: "limit", reset: "1", wantError: "between"},
		{name: "workflow test runs limit", cmd: workflowsTestsRunsListCmd, flag: "limit", reset: "1", wantError: "between"},
		{name: "workflow tests consensus", cmd: workflowsTestsRunsCreateCmd, flag: "n-consensus"},
		{name: "workflow experiments create consensus", cmd: workflowsExperimentsCreateCmd, flag: "n-consensus"},
		{name: "workflow experiments update consensus", cmd: workflowsExperimentsUpdateCmd, flag: "n-consensus"},
		{name: "workflow block width", cmd: workflowsBlocksUpdateCmd, flag: "width"},
		{name: "workflow block height", cmd: workflowsBlocksUpdateCmd, flag: "height"},
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
			wantError := tc.wantError
			if wantError == "" {
				wantError = "non-negative"
			}
			if !strings.Contains(err.Error(), wantError) && !strings.Contains(err.Error(), "3, 5, or 7") {
				t.Fatalf("error %q does not contain a numeric validation hint", err.Error())
			}
			if _, isConsensus := tc.cmd.Flags().Lookup(tc.flag).Value.(*consensusFlagValue); isConsensus {
				resetConsensusFlag(t, tc.cmd)
				return
			}
			reset := tc.reset
			if reset == "" {
				reset = "0"
			}
			if resetErr := tc.cmd.Flags().Set(tc.flag, reset); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestDPIFlagsRejectValuesOutsideBackendRangeLocally(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "parses create", cmd: parsesCreateCmd},
		{name: "extractions create", cmd: extractionsCreateCmd},
		{name: "extractions stream", cmd: extractionsStreamCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, value := range []string{"1", "999"} {
				err := tc.cmd.Flags().Set("image-resolution-dpi", value)
				if err == nil {
					t.Fatalf("expected local parse error for --image-resolution-dpi=%s", value)
				}
				if !strings.Contains(err.Error(), "between 96 and 300") {
					t.Fatalf("error %q does not mention DPI range", err.Error())
				}
			}
			if resetErr := tc.cmd.Flags().Set("image-resolution-dpi", "96"); resetErr != nil {
				t.Fatalf("reset --image-resolution-dpi: %v", resetErr)
			}
		})
	}
}

func TestPositiveNumericFlagsRejectZeroValuesLocally(t *testing.T) {
	cases := []struct {
		name      string
		cmd       *cobra.Command
		flag      string
		reset     string
		wantError string
	}{
		{name: "extractions create consensus", cmd: extractionsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "extractions stream consensus", cmd: extractionsStreamCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "classifications consensus", cmd: classificationsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "classifications first pages", cmd: classificationsCreateCmd, flag: "first-n-pages", reset: "1", wantError: "positive"},
		{name: "splits consensus", cmd: splitsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
		{name: "partitions consensus", cmd: partitionsCreateCmd, flag: "n-consensus", reset: "1", wantError: "between"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set(tc.flag, "0")
			if err == nil {
				t.Fatalf("expected local parse error for --%s=0", tc.flag)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if resetErr := tc.cmd.Flags().Set(tc.flag, tc.reset); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestConsensusFlagsRejectValuesAboveBackendRangeLocally(t *testing.T) {
	cases := []struct {
		name  string
		cmd   *cobra.Command
		value string
		reset string
	}{
		{name: "extractions create consensus", cmd: extractionsCreateCmd, value: "17", reset: "1"},
		{name: "extractions stream consensus", cmd: extractionsStreamCmd, value: "17", reset: "1"},
		{name: "classifications consensus", cmd: classificationsCreateCmd, value: "17", reset: "1"},
		{name: "splits consensus", cmd: splitsCreateCmd, value: "9", reset: "1"},
		{name: "partitions consensus", cmd: partitionsCreateCmd, value: "9", reset: "1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("n-consensus", tc.value)
			if err == nil {
				t.Fatalf("expected local parse error for --n-consensus=%s", tc.value)
			}
			if !strings.Contains(err.Error(), "between") {
				t.Fatalf("error %q does not mention bounded validation", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("n-consensus", tc.reset); resetErr != nil {
				t.Fatalf("reset --n-consensus: %v", resetErr)
			}
		})
	}
}

func TestWorkflowConsensusFlagErrorMatchesHelp(t *testing.T) {
	// The flag's help text on all three consumers reads "(3, 5, or 7)" — the
	// validator's error should match. Passing 0 explicitly should be rejected
	// (Cobra's Changed() handles the "not provided" path without needing 0 as
	// a sentinel inside Set).
	for _, cmd := range []*cobra.Command{
		workflowsTestsRunsCreateCmd,
		workflowsExperimentsCreateCmd,
		workflowsExperimentsUpdateCmd,
	} {
		err := cmd.Flags().Set("n-consensus", "0")
		if err == nil {
			t.Fatalf("%s: expected --n-consensus=0 to be rejected", cmd.Name())
		}
		if strings.Contains(err.Error(), "0, 3, 5, or 7") {
			t.Fatalf("%s: error %q still lists 0 as a valid value", cmd.Name(), err.Error())
		}
		if !strings.Contains(err.Error(), "3, 5, or 7") {
			t.Fatalf("%s: error %q does not match help text \"3, 5, or 7\"", cmd.Name(), err.Error())
		}
	}
}

func TestSharedListLimitFlagsRejectValuesAboveBackendRangeLocally(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "files", cmd: filesListCmd},
		{name: "extractions", cmd: extractionsListCmd},
		{name: "classifications", cmd: classificationsListCmd},
		{name: "splits", cmd: splitsListCmd},
		{name: "partitions", cmd: partitionsListCmd},
		{name: "parses", cmd: parsesListCmd},
		{name: "edits", cmd: editsListCmd},
		{name: "edit templates", cmd: editsTemplatesListCmd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("limit", "101")
			if err == nil {
				t.Fatal("expected local parse error for --limit=101")
			}
			if !strings.Contains(err.Error(), "between 1 and 100") {
				t.Fatalf("error %q does not mention backend limit range", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("limit", "1"); resetErr != nil {
				t.Fatalf("reset --limit: %v", resetErr)
			}
		})
	}
}
