//go:build !retab_oagen_cli_workflows_blocks_executions

package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsBlockExecutionsCmd = &cobra.Command{
	Use:   "executions",
	Short: "Run and inspect workflow block executions",
	Long: `Create and list block executions for one block within a workflow run.

A block execution replays a block with the current draft configuration against
inputs from an existing run. Use this to verify a block change before
starting another full workflow run.`,
	Example: `  # Re-run one block using inputs from a prior run
  retab workflows blocks executions create run_xyz789 --block-id blk_extract_1

  # List recent block executions for that run and block
  retab workflows blocks executions list run_xyz789 --block-id blk_extract_1`,
}

var workflowsBlockExecutionsCreateCmd = &cobra.Command{
	Use:   "create <run-id>",
	Short: "Create a workflow block execution",
	Long: `Create a block execution by replaying a block with the current draft
configuration against inputs captured from an existing workflow run.

The run id is positional; ` + "`--block-id`" + ` selects the block to replay.
For for_each blocks, ` + "`--step-id`" + ` can pin a concrete iteration
step.`,
	Example: `  # Re-run one block
  retab workflows blocks executions create run_xyz789 --block-id blk_extract_1

  # Pin a for_each iteration source step
  retab workflows blocks executions create run_xyz789 \
    --block-id blk_extract_1 \
    --step-id step_iter_0_blk_extract_1`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		nConsensus := 0
		if raw, _ := cmd.Flags().GetString("n-consensus"); raw != "" {
			switch raw {
			case "3":
				nConsensus = 3
			case "5":
				nConsensus = 5
			case "7":
				nConsensus = 7
			default:
				return fmt.Errorf("--n-consensus must be 3, 5, or 7")
			}
		}
		if err := validateBlockExecutionNConsensus(nConsensus); err != nil {
			return err
		}
		request := retab.WorkflowBlockExecutionsCreateParams{RunID: args[0]}
		request.BlockID, _ = cmd.Flags().GetString("block-id")
		if stepID, _ := cmd.Flags().GetString("step-id"); stepID != "" {
			request.StepID = ptr(stepID)
		}
		if nConsensus != 0 {
			request.NConsensus = ptr(nConsensus)
		}
		noCheckEligibility, _ := cmd.Flags().GetBool("no-check-eligibility")
		if noCheckEligibility {
			request.CheckEligibility = ptr(false)
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.Executions.Create(ctx, &request)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlockExecutionsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List workflow block executions",
	Long: `List block executions for a block within one workflow run.

The run id is positional; ` + "`--block-id`" + ` selects the block whose
block execution history should be returned.`,
	Example: `  # List recent block executions
  retab workflows blocks executions list run_xyz789 --block-id blk_extract_1

  # Limit the response size
  retab workflows blocks executions list run_xyz789 --block-id blk_extract_1 --limit 10`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.WorkflowBlockExecutionsListParams{
			PaginationParams: collectListParams(cmd),
			RunID:            args[0],
		}
		params.BlockID, _ = cmd.Flags().GetString("block-id")
		result, err := client.Workflows.Blocks.Executions.List(ctx, &params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func validateBlockExecutionNConsensus(value int) error {
	if value == 0 {
		return nil
	}
	switch value {
	case 3, 5, 7:
		return nil
	default:
		return fmt.Errorf("--n-consensus must be 3, 5, or 7, got %d", value)
	}
}

func init() {
	workflowsBlockExecutionsCreateCmd.Flags().String("block-id", "", "block id to execute (required)")
	workflowsBlockExecutionsCreateCmd.Flags().String("step-id", "", "specific iteration step id to source inputs from")
	workflowsBlockExecutionsCreateCmd.Flags().String("n-consensus", "", "override n_consensus for extract, split, or classifier blocks (3, 5, or 7)")
	workflowsBlockExecutionsCreateCmd.Flags().Bool("no-check-eligibility", false, "skip upstream drift eligibility checks")
	_ = workflowsBlockExecutionsCreateCmd.MarkFlagRequired("block-id")

	workflowsBlockExecutionsListCmd.Flags().String("block-id", "", "block id to list block executions for (required)")
	workflowsBlockExecutionsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	_ = workflowsBlockExecutionsListCmd.MarkFlagRequired("block-id")

	workflowsBlockExecutionsCmd.AddCommand(workflowsBlockExecutionsCreateCmd, workflowsBlockExecutionsListCmd)
	workflowsBlocksCmd.AddCommand(workflowsBlockExecutionsCmd)
}
