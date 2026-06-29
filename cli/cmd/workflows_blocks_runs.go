//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"os"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsBlocksRunsCmd = &cobra.Command{
	Use:   "runs [<workflow-id>] <block-id>",
	Short: "List workflow run steps for a block",
	Long: `List workflow step records for one block.

This is the block-centric view of ` + "`workflows steps list`" + `: it filters the
public steps API by block_id and optionally by run_id/status. When a workflow id
is supplied, the CLI first verifies that the block resolves in that workflow and
then lists the matching step executions.`,
	Example: `  # List executions of one block across runs
  retab workflows blocks runs blk_extract

  # Verify the block in a workflow before listing its executions
  retab workflows blocks runs wf_abc123 blk_extract

  # Filter to one run
  retab workflows blocks runs list blk_extract --run-id run_xyz789`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(runWorkflowsBlocksRunsList),
}

var workflowsBlocksRunsListCmd = &cobra.Command{
	Use:   "list [<workflow-id>] <block-id>",
	Short: "List workflow run steps for a block",
	Long:  workflowsBlocksRunsCmd.Long,
	Example: `  retab workflows blocks runs list blk_extract
  retab workflows blocks runs list wf_abc123 blk_extract --status completed`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(runWorkflowsBlocksRunsList),
}

func runWorkflowsBlocksRunsList(cmd *cobra.Command, args []string) error {
	if err := validateBeforeAfterMutex(cmd); err != nil {
		return err
	}
	workflowID, blockID, err := resolveWorkflowBlockScope(cmd, args, false, "runs")
	if err != nil {
		return err
	}
	client, err := newClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	if workflowID != "" {
		if _, err := client.Workflows.Blocks.Get(ctx, blockID, &retab.WorkflowBlocksGetParams{WorkflowID: &workflowID}); err != nil {
			return err
		}
	}
	params := workflowBlockRunsListParams(cmd, blockID)
	result, err := client.Workflows.Steps.List(ctx, params)
	if err != nil {
		return err
	}
	return printBlockRunsListResult(cmd, result)
}

func workflowBlockRunsListParams(cmd *cobra.Command, blockID string) *retab.WorkflowStepsListParams {
	params := &retab.WorkflowStepsListParams{
		PaginationParams: collectListParams(cmd),
		BlockID:          ptr(blockID),
	}
	if runID, _ := cmd.Flags().GetString("run-id"); strings.TrimSpace(runID) != "" {
		params.RunID = ptr(strings.TrimSpace(runID))
	}
	if status, _ := cmd.Flags().GetString("status"); strings.TrimSpace(status) != "" {
		params.Status = []string{strings.TrimSpace(status)}
	}
	return params
}

var blockRunColumns = []TableColumn{
	{Header: "RUN", Extract: func(row any) string { return blockRunCell(row, "run_id") }},
	{Header: "STEP", Extract: func(row any) string { return blockRunCell(row, "step_id") }},
	{Header: "STATUS", Extract: func(row any) string { return blockRunCell(row, "lifecycle.status") }},
	{Header: "TYPE", Extract: func(row any) string { return blockRunCell(row, "block_type") }},
	{Header: "LABEL", Extract: func(row any) string { return blockRunCell(row, "block_label") }},
	{Header: "STARTED_AT", Extract: func(row any) string { return normalizeTimestampCell(blockRunCell(row, "started_at")) }},
	{Header: "COMPLETED_AT", Extract: func(row any) string { return normalizeTimestampCell(blockRunCell(row, "completed_at")) }},
	{Header: "ARTIFACT", Extract: blockRunArtifactCell},
	{Header: "RETRIES", Extract: func(row any) string { return blockRunCell(row, "retry_count") }},
}

func printBlockRunsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowRunStep]) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		return RenderList(os.Stdout, format, result, blockRunColumns)
	}
	return printJSON(result)
}

func blockRunCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func blockRunArtifactCell(row any) string {
	id := blockRunCell(row, "artifact.id")
	operation := blockRunCell(row, "artifact.operation")
	switch {
	case operation != "" && id != "":
		return operation + ":" + id
	case id != "":
		return id
	default:
		return operation
	}
}

func addBlockRunsListFlags(cmd *cobra.Command) {
	cmd.Flags().String("run-id", "", "filter by workflow run id")
	cmd.Flags().String("status", "", "filter by step lifecycle status")
	cmd.Flags().String("before", "", "step id: return the page before this id (mutually exclusive with --after)")
	cmd.Flags().String("after", "", "step id: return the page after this id (mutually exclusive with --before)")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 1000}, "limit", "max items to return (1-1000)")
}

func init() {
	workflowsBlocksRunsCmd.PersistentFlags().String("workflow-id", "", "workflow id to verify block membership")
	addBlockRunsListFlags(workflowsBlocksRunsCmd)
	addBlockRunsListFlags(workflowsBlocksRunsListCmd)
	workflowsBlocksRunsCmd.AddCommand(workflowsBlocksRunsListCmd)
	workflowsBlocksCmd.AddCommand(workflowsBlocksRunsCmd)
}
