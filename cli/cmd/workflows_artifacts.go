//go:build !retab_oagen_cli_workflows_artifacts

package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var allowedWorkflowArtifactOperations = map[string]bool{
	"extraction":                true,
	"split":                     true,
	"classification":            true,
	"parse":                     true,
	"edit":                      true,
	"partition":                 true,
	"conditional_evaluation":    true,
	"review_trigger_evaluation": true,
	"while_loop_termination":    true,
	"api_call_invocation":       true,
	"function_invocation":       true,
}

const workflowArtifactOperationValues = "extraction, split, classification, parse, edit, partition, conditional_evaluation, review_trigger_evaluation, while_loop_termination, api_call_invocation, function_invocation"

func validateWorkflowArtifactOperation(operation string) error {
	if operation == "" {
		return nil
	}
	if !allowedWorkflowArtifactOperations[operation] {
		return fmt.Errorf("invalid operation %q (want: %s)", operation, workflowArtifactOperationValues)
	}
	return nil
}

var workflowsArtifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Inspect workflow artifacts",
	Long: `Persistent, addressable outputs produced by workflow runs:
parsed documents, extracted JSON blobs, classification results, edited
files, etc.

Artifacts are separate objects from the run that produced them: they
survive ` + "`workflows runs delete`" + ` so you can reference outputs long
after their run is gone. List every artifact tied to a run, or fetch one
by id with ` + "`get`" + `.`,
	Example: `  # All artifacts produced by a run
  retab workflows artifacts list run_xyz789

  # Just the extract block's artifacts
  retab workflows artifacts list run_xyz789 --block-id block_extract_1

  # Fetch one artifact by id
  retab workflows artifacts get extr_lz1_abc`,
}

var workflowsArtifactsGetCmd = &cobra.Command{
	Use:   "get <artifact-id>",
	Short: "Get one workflow artifact by id",
	Long: `Fetch a single artifact produced by a workflow run.

The artifact id appears on each ` + "`workflows steps list`" + ` row as
` + "`artifact.id`" + ` (e.g. ` + "`extr_lz1_…`" + ` for an extraction,
` + "`clss_…`" + ` for a classification).

Backing route: ` + "`GET /workflows/artifacts/{artifact_id}`" + `. The server
derives the backing collection from the id prefix.`,
	Example: `  # Fetch one extraction artifact
  retab workflows artifacts get extr_lz1_abc

  # Fetch one classification artifact
  retab workflows artifacts get clss_xyz_123`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		artifactID := args[0]
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Artifacts.Get(ctx, artifactID)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsArtifactsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List artifacts for a run",
	Long: `List artifacts produced by a workflow run. Filter by
` + "`--block-id`" + ` to focus on one node's outputs, or by
` + "`--operation`" + ` to scope to one artifact type.

The run id is positional, matching the other ` + "`workflows <X> list`" + `
commands (` + "`workflows tests list <wf-id>`" + `,
` + "`workflows steps list <run-id>`" + `, …): the positional slot is
always the parent id, and flags are reserved for filters.

Paginate by passing the cursor from a previous response's
` + "`list_metadata`" + ` (the producing step's ` + "`step_id`" + `):
` + "`--after`" + ` for the next page, ` + "`--before`" + ` for the previous
one. The two are mutually exclusive.`,
	Example: `  # All artifacts from a run
  retab workflows artifacts list run_xyz789

  # Just one block's outputs
  retab workflows artifacts list run_xyz789 --block-id block_extract_1

  # Just parse-block artifacts
  retab workflows artifacts list run_xyz789 --operation parse

  # First page of 50
  retab workflows artifacts list run_xyz789 --limit 50

  # Next page
  retab workflows artifacts list run_xyz789 --limit 50 \
    --after $(retab workflows artifacts list run_xyz789 --limit 50 --output json | jq -r '.list_metadata.after')`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.WorkflowArtifactsListParams{PaginationParams: collectListParams(cmd)}
		params.RunID = ptr(args[0])
		operation, _ := cmd.Flags().GetString("operation")
		if err := validateWorkflowArtifactOperation(operation); err != nil {
			return err
		}
		if operation != "" {
			op := retab.WorkflowArtifactsOperation(operation)
			params.Operation = &op
		}
		if blockID, _ := cmd.Flags().GetString("block-id"); blockID != "" {
			params.BlockID = ptr(blockID)
		}
		result, err := client.Workflows.Artifacts.List(ctx, &params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsArtifactsListCmd.Flags().String("operation", "", "filter by operation")
	workflowsArtifactsListCmd.Flags().String("block-id", "", "filter by block id")
	workflowsArtifactsListCmd.Flags().String("before", "", "step id: return the page before this step (mutually exclusive with --after)")
	workflowsArtifactsListCmd.Flags().String("after", "", "step id: return the page after this step (mutually exclusive with --before)")
	workflowsArtifactsListCmd.MarkFlagsMutuallyExclusive("before", "after")
	workflowsArtifactsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")

	workflowsArtifactsCmd.AddCommand(workflowsArtifactsGetCmd, workflowsArtifactsListCmd)
	workflowsCmd.AddCommand(workflowsArtifactsCmd)
}
