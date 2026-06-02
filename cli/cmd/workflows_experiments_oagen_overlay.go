//go:build retab_oagen_cli_workflows_experiments

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func resolveExperimentIDArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("experiment id required")
	}
	id := strings.TrimSpace(args[len(args)-1])
	if id == "" {
		return "", fmt.Errorf("experiment id required")
	}
	return id, nil
}

var workflowsExperimentsRunsWaitCmd = &cobra.Command{
	Use:   "wait <run-id>",
	Short: "Poll until an experiment run reaches a terminal status",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("generated prototype does not implement workflows experiments runs wait")
	}),
}

func init() {
	workflowsExperimentsRunsWaitCmd.Flags().Int("poll-interval-ms", 2000, "poll cadence in milliseconds")
	workflowsExperimentsRunsWaitCmd.Flags().Int("timeout-seconds", 600, "max seconds to wait before giving up")
}
