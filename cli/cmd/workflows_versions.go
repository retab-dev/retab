//go:build !retab_oagen_cli_workflows

package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsVersionsCmd = &cobra.Command{
	Use:   "versions <workflow-id>",
	Short: "List workflow versions",
	Long: `List immutable versions for a workflow.

Use this to find version ids for diffing, inspection, or restoring a prior
published graph back into the draft.`,
	Example: `  retab workflows versions wf_abc123
  retab workflows versions wf_abc123 --limit 20`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		params := &retab.WorkflowsListVersionsParams{WorkflowID: args[0]}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			params.Limit = ptr(limit)
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.ListVersions(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsDiffCmd = &cobra.Command{
	Use:   "diff <workflow-id>",
	Short: "Diff workflow versions",
	Long: `Diff two immutable workflow graph versions for the same workflow.

Pass version ids from ` + "`retab workflows versions <workflow-id>`" + `.`,
	Example: `  retab workflows diff wf_abc123 \
    --from-version-id wfv_old --to-version-id wfv_new`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		fromVersionID, _ := cmd.Flags().GetString("from-version-id")
		toVersionID, _ := cmd.Flags().GetString("to-version-id")
		if fromVersionID == "" {
			return fmt.Errorf("--from-version-id is required")
		}
		if toVersionID == "" {
			return fmt.Errorf("--to-version-id is required")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.ListDiff(ctx, &retab.WorkflowsListDiffParams{
			WorkflowID:            args[0],
			FromWorkflowVersionID: fromVersionID,
			ToWorkflowVersionID:   toVersionID,
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionCmd = &cobra.Command{
	Use:   "version <workflow-id> <workflow-version-id>",
	Short: "Get a workflow version",
	Long: `Fetch one immutable workflow graph version.

The workflow id disambiguates content-addressed version ids that may be reused
across workflows.`,
	Example: `  retab workflows version wf_abc123 wfv_456`,
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.GetVersion(ctx, args[1], &retab.WorkflowsGetVersionParams{
			WorkflowID: args[0],
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionRestoreCmd = &cobra.Command{
	Use:   "version-restore <workflow-id> <workflow-version-id>",
	Short: "Restore a workflow version",
	Long: `Restore an immutable workflow graph version into the workflow draft.

This mutates the draft graph. The restored version remains immutable; only the
current draft is replaced.`,
	Example: `  retab workflows version-restore wf_abc123 wfv_456 --yes`,
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "workflow draft", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.CreateVersionRestore(ctx, args[1], &retab.WorkflowsCreateVersionRestoreParams{
			WorkflowID: args[0],
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsVersionsCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")

	workflowsDiffCmd.Flags().String("from-version-id", "", "base workflow version id")
	workflowsDiffCmd.Flags().String("to-version-id", "", "target workflow version id")

	workflowsVersionRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsCmd.AddCommand(workflowsVersionsCmd, workflowsDiffCmd, workflowsVersionCmd, workflowsVersionRestoreCmd)
}
