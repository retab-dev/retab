//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsBlocksVersionsCmd = &cobra.Command{
	Use:   "versions <workflow-id>",
	Short: "List block versions",
	Long: `List immutable block versions within a workflow.

Filter by stable block id or workflow version id when you need a narrower
history.`,
	Example: `  retab workflows blocks versions wf_abc123
  retab workflows blocks versions wf_abc123 --block-id block_extract
  retab workflows blocks versions wf_abc123 --workflow-version-id wfv_456`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		params := &retab.WorkflowBlocksListVersionsParams{WorkflowID: args[0]}
		if blockID, _ := cmd.Flags().GetString("block-id"); blockID != "" {
			params.BlockID = ptr(blockID)
		}
		if workflowVersionID, _ := cmd.Flags().GetString("workflow-version-id"); workflowVersionID != "" {
			params.WorkflowVersionID = ptr(workflowVersionID)
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			params.Limit = ptr(limit)
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

var workflowsBlocksDiffCmd = &cobra.Command{
	Use:     "diff <from-block-version-id> <to-block-version-id>",
	Short:   "Diff block versions",
	Long:    `Diff two immutable block versions.`,
	Example: `  retab workflows blocks diff bv_old bv_new`,
	Args:    cobra.ExactArgs(2),
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

var workflowsBlocksVersionCmd = &cobra.Command{
	Use:     "version <block-version-id>",
	Short:   "Get a block version",
	Long:    `Fetch one immutable block version.`,
	Example: `  retab workflows blocks version bv_456`,
	Args:    cobra.ExactArgs(1),
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

var workflowsBlocksVersionRestoreCmd = &cobra.Command{
	Use:   "version-restore <block-version-id>",
	Short: "Restore a block version",
	Long: `Restore an immutable block version into the workflow draft.

This mutates the draft block state. The historical block version remains
immutable.`,
	Example: `  retab workflows blocks version-restore bv_456 --yes`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "block version", args[0]); err != nil {
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
	workflowsBlocksVersionsCmd.Flags().String("block-id", "", "filter by stable block id")
	workflowsBlocksVersionsCmd.Flags().String("workflow-version-id", "", "filter by workflow version id")
	workflowsBlocksVersionsCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")

	workflowsBlocksVersionRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsBlocksCmd.AddCommand(workflowsBlocksVersionsCmd, workflowsBlocksDiffCmd, workflowsBlocksVersionCmd, workflowsBlocksVersionRestoreCmd)
}
