//go:build !retab_oagen_cli_workflows

package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Inspect and restore workflow versions",
	Long:  `Inspect immutable workflow graph versions and restore one into the current draft.`,
}

var workflowsVersionsListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List workflow versions",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		params := &retab.WorkflowsListVersionsParams{WorkflowID: args[0]}
		if limit > 0 {
			params.Limit = ptr(limit)
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
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

var workflowsVersionsGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <workflow-version-id>",
	Short: "Get a workflow version",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.GetVersion(ctx, args[1], &retab.WorkflowsGetVersionParams{WorkflowID: args[0]})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionsDiffCmd = &cobra.Command{
	Use:   "diff <workflow-id> <from-version-id> <to-version-id>",
	Short: "Diff two workflow versions",
	Args:  cobra.ExactArgs(3),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.ListDiff(ctx, &retab.WorkflowsListDiffParams{
			WorkflowID:            args[0],
			FromWorkflowVersionID: args[1],
			ToWorkflowVersionID:   args[2],
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionsRestoreCmd = &cobra.Command{
	Use:   "restore <workflow-id> <workflow-version-id>",
	Short: "Restore a workflow version into the draft",
	Long: `Restore an immutable workflow graph version into the workflow's current draft.
This overwrites the editable draft graph with a fresh draft created from the
selected version. Pass ` + "`--yes`" + ` to confirm.`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "workflow draft", fmt.Sprintf("%s from %s", args[0], args[1])); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.CreateVersionRestore(ctx, args[1], &retab.WorkflowsCreateVersionRestoreParams{WorkflowID: args[0]})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsVersionsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	workflowsVersionsRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	workflowsVersionsCmd.AddCommand(workflowsVersionsListCmd, workflowsVersionsGetCmd, workflowsVersionsDiffCmd, workflowsVersionsRestoreCmd)
	workflowsCmd.AddCommand(workflowsVersionsCmd)
}
