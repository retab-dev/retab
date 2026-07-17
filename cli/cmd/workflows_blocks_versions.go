//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsBlocksVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Inspect and restore workflow block versions",
}

var workflowsBlocksVersionsListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List block versions",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		blockID, _ := cmd.Flags().GetString("block-id")
		workflowVersionID, _ := cmd.Flags().GetString("workflow-version-id")
		limit, _ := cmd.Flags().GetInt("limit")
		params := &retab.WorkflowBlocksListVersionsParams{WorkflowID: args[0]}
		if blockID != "" {
			params.BlockID = ptr(blockID)
		}
		if workflowVersionID != "" {
			params.WorkflowVersionID = ptr(workflowVersionID)
		}
		if limit > 0 {
			params.Limit = ptr(limit)
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.ListVersions(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksVersionsGetCmd = &cobra.Command{
	Use:   "get <block-version-id>",
	Short: "Get a block version",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.GetVersion(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksVersionsDiffCmd = &cobra.Command{
	Use:   "diff <from-block-version-id> <to-block-version-id>",
	Short: "Diff two block versions",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.ListDiff(ctx, &retab.WorkflowBlocksListDiffParams{
			FromBlockVersionID: args[0],
			ToBlockVersionID:   args[1],
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksVersionsRestoreCmd = &cobra.Command{
	Use:   "restore <block-version-id>",
	Short: "Restore a block version into the draft",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructiveVerb(cmd, "overwrite", "workflow block draft from version", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.CreateVersionRestore(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsBlocksVersionsListCmd.Flags().String("block-id", "", "stable block id filter")
	workflowsBlocksVersionsListCmd.Flags().String("workflow-version-id", "", "workflow version id filter")
	workflowsBlocksVersionsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	workflowsBlocksVersionsRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	workflowsBlocksVersionsCmd.AddCommand(workflowsBlocksVersionsListCmd, workflowsBlocksVersionsGetCmd, workflowsBlocksVersionsDiffCmd, workflowsBlocksVersionsRestoreCmd)
	workflowsBlocksCmd.AddCommand(workflowsBlocksVersionsCmd)
}
