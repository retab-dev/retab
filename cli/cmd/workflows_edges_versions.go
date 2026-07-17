//go:build !retab_oagen_cli_workflows_edges

package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsEdgesVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Inspect and restore workflow edge versions",
}

var workflowsEdgesVersionsListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List edge versions",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		edgeID, _ := cmd.Flags().GetString("edge-id")
		workflowVersionID, _ := cmd.Flags().GetString("workflow-version-id")
		limit, _ := cmd.Flags().GetInt("limit")
		params := &retab.WorkflowEdgesListVersionsParams{WorkflowID: args[0]}
		if edgeID != "" {
			params.EdgeID = ptr(edgeID)
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
		result, err := client.Workflows.Edges.ListVersions(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEdgesVersionsGetCmd = &cobra.Command{
	Use:   "get <edge-version-id>",
	Short: "Get an edge version",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Edges.GetVersion(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEdgesVersionsDiffCmd = &cobra.Command{
	Use:   "diff <from-edge-version-id> <to-edge-version-id>",
	Short: "Diff two edge versions",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Edges.ListDiff(ctx, &retab.WorkflowEdgesListDiffParams{
			FromEdgeVersionID: args[0],
			ToEdgeVersionID:   args[1],
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEdgesVersionsRestoreCmd = &cobra.Command{
	Use:   "restore <edge-version-id>",
	Short: "Restore an edge version into the draft",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructiveVerb(cmd, "overwrite", "workflow edge draft from version", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Edges.CreateVersionRestore(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsEdgesVersionsListCmd.Flags().String("edge-id", "", "stable edge id filter")
	workflowsEdgesVersionsListCmd.Flags().String("workflow-version-id", "", "workflow version id filter")
	workflowsEdgesVersionsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	workflowsEdgesVersionsRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	workflowsEdgesVersionsCmd.AddCommand(workflowsEdgesVersionsListCmd, workflowsEdgesVersionsGetCmd, workflowsEdgesVersionsDiffCmd, workflowsEdgesVersionsRestoreCmd)
	workflowsEdgesCmd.AddCommand(workflowsEdgesVersionsCmd)
}
