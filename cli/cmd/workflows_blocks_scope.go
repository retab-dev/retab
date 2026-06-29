//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func resolveWorkflowBlockScope(cmd *cobra.Command, args []string, requireWorkflow bool, resourceName string) (string, string, error) {
	flagWorkflowID := workflowBlockWorkflowID(cmd)
	var workflowID string
	var blockID string
	switch len(args) {
	case 1:
		workflowID = flagWorkflowID
		blockID = strings.TrimSpace(args[0])
	case 2:
		workflowID = strings.TrimSpace(args[0])
		blockID = strings.TrimSpace(args[1])
		if workflowID == "" {
			return "", "", fmt.Errorf("workflow-id positional argument is empty")
		}
		if flagWorkflowID != "" && flagWorkflowID != workflowID {
			return "", "", fmt.Errorf("conflicting workflow id: positional %q vs --workflow-id %q", workflowID, flagWorkflowID)
		}
	default:
		return "", "", fmt.Errorf("expected 1 or 2 positional arguments, got %d", len(args))
	}
	if blockID == "" {
		return "", "", fmt.Errorf("block-id positional argument is empty")
	}
	if requireWorkflow && workflowID == "" {
		return "", "", fmt.Errorf("workflow id is required for block %s; pass `%s <workflow-id> <block-id>` or `--workflow-id <workflow-id>`", resourceName, resourceName)
	}
	return workflowID, blockID, nil
}

func workflowBlockWorkflowID(cmd *cobra.Command) string {
	for current := cmd; current != nil; current = current.Parent() {
		if f := current.Flags().Lookup("workflow-id"); f != nil {
			value, _ := current.Flags().GetString("workflow-id")
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
		if f := current.PersistentFlags().Lookup("workflow-id"); f != nil {
			value, _ := current.PersistentFlags().GetString("workflow-id")
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}
