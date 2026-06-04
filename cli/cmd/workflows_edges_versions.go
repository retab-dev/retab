//go:build !retab_oagen_cli_workflows_edges

package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsEdgesVersionsCmd = &cobra.Command{
	Use:   "versions <workflow-id>",
	Short: "List edge versions",
	Long: `List immutable edge versions within a workflow.

Filter by stable edge id or workflow version id when you need a narrower
history.`,
	Example: `  retab workflows edges versions wf_abc123
  retab workflows edges versions wf_abc123 --edge-id edge_123
  retab workflows edges versions wf_abc123 --workflow-version-id wfv_456`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		params := &retab.WorkflowEdgesListVersionsParams{WorkflowID: args[0]}
		if edgeID, _ := cmd.Flags().GetString("edge-id"); edgeID != "" {
			params.EdgeID = ptr(edgeID)
		}
		if workflowVersionID, _ := cmd.Flags().GetString("workflow-version-id"); workflowVersionID != "" {
			params.WorkflowVersionID = ptr(workflowVersionID)
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			params.Limit = ptr(limit)
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

var workflowsEdgesDiffCmd = &cobra.Command{
	Use:     "diff <from-edge-version-id> <to-edge-version-id>",
	Short:   "Diff edge versions",
	Long:    `Diff two immutable edge versions.`,
	Example: `  retab workflows edges diff ev_old ev_new`,
	Args:    cobra.ExactArgs(2),
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

var workflowsEdgesVersionCmd = &cobra.Command{
	Use:     "version <edge-version-id>",
	Short:   "Get an edge version",
	Long:    `Fetch one immutable edge version.`,
	Example: `  retab workflows edges version ev_456`,
	Args:    cobra.ExactArgs(1),
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

var workflowsEdgesVersionRestoreCmd = &cobra.Command{
	Use:   "version-restore <edge-version-id>",
	Short: "Restore an edge version",
	Long: `Restore an immutable edge version into the workflow draft.

This mutates the draft edge state. The historical edge version remains
immutable.`,
	Example: `  retab workflows edges version-restore ev_456 --yes`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "edge version", args[0]); err != nil {
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
	workflowsEdgesVersionsCmd.Flags().String("edge-id", "", "filter by stable edge id")
	workflowsEdgesVersionsCmd.Flags().String("workflow-version-id", "", "filter by workflow version id")
	workflowsEdgesVersionsCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")

	workflowsEdgesVersionRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsEdgesCmd.AddCommand(workflowsEdgesVersionsCmd, workflowsEdgesDiffCmd, workflowsEdgesVersionCmd, workflowsEdgesVersionRestoreCmd)
}
