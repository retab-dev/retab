package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var allowedWorkflowArtifactOperations = map[string]bool{
	"extraction":             true,
	"split":                  true,
	"classification":         true,
	"parse":                  true,
	"edit":                   true,
	"partition":              true,
	"conditional_evaluation": true,
	"hil_evaluation":         true,
	"while_loop_termination": true,
	"api_call_invocation":    true,
	"function_invocation":    true,
}

const workflowArtifactOperationValues = "extraction, split, classification, parse, edit, partition, conditional_evaluation, hil_evaluation, while_loop_termination, api_call_invocation, function_invocation"

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
after their run is gone. Look up an artifact by ` + "`<operation>`" + ` +
` + "`<artifact-id>`" + `, or list every artifact tied to a run.`,
	Example: `  # All artifacts produced by a run
  retab workflows artifacts list run_xyz789

  # Just the extract block's artifacts
  retab workflows artifacts list run_xyz789 --block-id blk_extract_1

  # Fetch one artifact
  retab workflows artifacts get extraction art_stu901`,
}

var workflowsArtifactsGetCmd = &cobra.Command{
	Use:   "get <operation> <artifact-id>",
	Short: "Get an artifact by operation and id",
	Long: `Fetch one artifact. ` + "`<operation>`" + ` is the persisted artifact
operation (e.g. ` + "`extraction`" + `, ` + "`parse`" + `,
` + "`classification`" + `) and addresses the artifact namespace.`,
	Example: `  # Fetch an extracted JSON artifact
  retab workflows artifacts get extraction art_stu901

  # Fetch a parsed document
  retab workflows artifacts get parse art_vwx234`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateWorkflowArtifactOperation(args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Artifacts.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
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
` + "`workflows runs steps list <run-id>`" + `, …): the positional slot is
always the parent id, and flags are reserved for filters.`,
	Example: `  # All artifacts from a run
  retab workflows artifacts list run_xyz789

  # Just one block's outputs
  retab workflows artifacts list run_xyz789 --block-id blk_extract_1

  # Just parse-block artifacts
  retab workflows artifacts list run_xyz789 --operation parse`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowArtifactsParams{}
		params.RunID = args[0]
		params.Operation, _ = cmd.Flags().GetString("operation")
		if err := validateWorkflowArtifactOperation(params.Operation); err != nil {
			return err
		}
		params.BlockID, _ = cmd.Flags().GetString("block-id")
		result, err := client.Workflows.Artifacts.List(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsArtifactsListCmd.Flags().String("operation", "", "filter by operation")
	workflowsArtifactsListCmd.Flags().String("block-id", "", "filter by block id")

	workflowsArtifactsCmd.AddCommand(workflowsArtifactsGetCmd, workflowsArtifactsListCmd)
	workflowsCmd.AddCommand(workflowsArtifactsCmd)
}
