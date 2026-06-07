//go:build !retab_oagen_cli_files && !retab_oagen_cli_workflows && !retab_oagen_cli_workflows_runs

package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSortByFlagsRejectUnsupportedFieldsLocally(t *testing.T) {
	cases := []struct {
		name  string
		cmd   *cobra.Command
		valid string
	}{
		{name: "files", cmd: filesListCmd, valid: "created_at"},
		{name: "workflows", cmd: workflowsListCmd, valid: "updated_at"},
		{name: "workflow runs", cmd: workflowsRunsListCmd, valid: "timing.created_at"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("sort-by", "banana")
			if err == nil {
				t.Fatal("expected local parse error for unsupported --sort-by")
			}
			if !strings.Contains(err.Error(), "sort") {
				t.Fatalf("error %q does not mention sort", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("sort-by", tc.valid); resetErr != nil {
				t.Fatalf("reset --sort-by: %v", resetErr)
			}
			if resetErr := tc.cmd.Flags().Set("sort-by", ""); resetErr != nil {
				t.Fatalf("clear --sort-by: %v", resetErr)
			}
		})
	}
}
