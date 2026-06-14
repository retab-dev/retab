//go:build !retab_oagen_cli_workflows_steps

package cmd

import (
	"fmt"
	"strconv"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "Inspect workflow run steps",
	Long: `A step is the execution record of one block within a run —
input it saw, output it produced, error if any, timing, cost. Steps are
the entry point for debugging: when a run output looks wrong, list its
steps and pull the offending one with ` + "`workflows steps get`" + `
to see exactly what the block did.`,
	Example: `  # List every step in a run
  retab workflows steps list run_xyz789

  # Pull the full execution record for one step
  retab workflows steps get step_extract_1`,
}

var workflowsStepsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List run steps",
	Long: `List every step in a run — one record per block execution.
Includes status, timing, and a summary of input/output sizes. For the
full input/output payload of one step use ` + "`steps get`" + `.

Paginate by passing the cursor from a previous response's
` + "`list_metadata`" + `: ` + "`--after`" + ` for the next page,
` + "`--before`" + ` for the previous one. The two are mutually exclusive.`,
	Example: `  # List steps
  retab workflows steps list run_xyz789

  # First page of 100
  retab workflows steps list run_xyz789 --limit 100

  # Next page
  retab workflows steps list run_xyz789 --limit 100 \
    --after $(retab workflows steps list run_xyz789 --limit 100 --output json | jq -r '.list_metadata.after')

  # Find the first failed step
  retab workflows steps list run_xyz789 \
    | jq '.data[] | select(.lifecycle.status == "error") | .block_id' | head -1`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		runID, err := scopedResourceID(args, "run id")
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		params, err := workflowStepsListParams(cmd, runID)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Steps.List(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func workflowStepsListParams(cmd *cobra.Command, runID string) (*retab.WorkflowStepsListParams, error) {
	params := &retab.WorkflowStepsListParams{}
	if runID != "" {
		params.RunID = ptr(runID)
	}
	before, _ := cmd.Flags().GetString("before")
	after, _ := cmd.Flags().GetString("after")
	if before != "" {
		params.Before = ptr(before)
	}
	if after != "" {
		params.After = ptr(after)
	}
	if f := cmd.Flags().Lookup("limit"); f != nil && f.Changed {
		raw := f.Value.String()
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --limit %q", raw)
		}
		if parsed > 0 {
			params.Limit = ptr(parsed)
		}
	}
	return params, nil
}

var workflowsStepsGetCmd = &cobra.Command{
	Use:   "get <step-id>",
	Short: "Get the full step execution record",
	Long: `Return everything about one step execution in a run: the
exact input payload, the produced output, any error, timing, cost, and
model usage if applicable.

This is the canonical entry point when debugging a run that produced
the wrong output — find the offending step id, inspect its inputs, and
correlate against the step's block config.`,
	Example: `  # Pull the full record for a single step
  retab workflows steps get step_extract_1

  # Save the input payload for offline replay
  retab workflows steps get step_extract_1 \
    | jq '.handle_inputs' > inputs.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Steps.Get(ctx, args[0], nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsStepsListCmd.Flags().String("before", "", "step id: return the page before this id (mutually exclusive with --after)")
	workflowsStepsListCmd.Flags().String("after", "", "step id: return the page after this id (mutually exclusive with --before)")
	// Mutex enforced inside RunE via validateBeforeAfterMutex (concise
	// handwritten message; see workflowsListCmd for the rationale).
	workflowsStepsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 1000}, "limit", "max items to return (1-1000)")

	workflowsStepsCmd.AddCommand(workflowsStepsListCmd, workflowsStepsGetCmd)
	workflowsCmd.AddCommand(workflowsStepsCmd)
}
