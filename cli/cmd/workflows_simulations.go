package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsSimulationsCmd = &cobra.Command{
	Use:   "simulations",
	Short: "Run and inspect workflow block simulations",
	Long: `Create and list simulations for one block within a workflow run.

A simulation replays a block with the current draft configuration against
inputs from an existing run. Use this to verify a block change before
starting another full workflow run.`,
	Example: `  # Re-run one block using inputs from a prior run
  retab workflows simulations create run_xyz789 --block-id blk_extract_1

  # List recent simulations for that run and block
  retab workflows simulations list run_xyz789 --block-id blk_extract_1`,
}

var workflowsSimulationsCreateCmd = &cobra.Command{
	Use:   "create <run-id>",
	Short: "Create a workflow block simulation",
	Long: `Create a simulation by replaying a block with the current draft
configuration against inputs captured from an existing workflow run.

The run id is positional; ` + "`--block-id`" + ` selects the block to replay.
For for_each blocks, ` + "`--source-step-id`" + ` can pin a concrete iteration
step.`,
	Example: `  # Re-run one block
  retab workflows simulations create run_xyz789 --block-id blk_extract_1

  # Pin a for_each iteration source step
  retab workflows simulations create run_xyz789 \
    --block-id blk_extract_1 \
    --source-step-id step_iter_0_blk_extract_1`,
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
		if err := validateSimulationNConsensus(nConsensus); err != nil {
			return err
		}
		request := retab.CreateWorkflowSimulationRequest{RunID: args[0]}
		request.BlockID, _ = cmd.Flags().GetString("block-id")
		request.SourceStepID, _ = cmd.Flags().GetString("source-step-id")
		request.NConsensus = nConsensus
		noCheckEligibility, _ := cmd.Flags().GetBool("no-check-eligibility")
		if noCheckEligibility {
			checkEligibility := false
			request.CheckEligibility = &checkEligibility
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Simulations.Create(ctx, request)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsSimulationsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List workflow block simulations",
	Long: `List simulations for a block within one workflow run.

The run id is positional; ` + "`--block-id`" + ` selects the block whose
simulation history should be returned.`,
	Example: `  # List recent simulations
  retab workflows simulations list run_xyz789 --block-id blk_extract_1

  # Limit the response size
  retab workflows simulations list run_xyz789 --block-id blk_extract_1 --limit 10`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowSimulationsParams{RunID: args[0]}
		params.BlockID, _ = cmd.Flags().GetString("block-id")
		params.Limit, _ = cmd.Flags().GetInt("limit")
		result, err := client.Workflows.Simulations.List(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func validateSimulationNConsensus(value int) error {
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
	workflowsSimulationsCreateCmd.Flags().String("block-id", "", "block id to simulate (required)")
	workflowsSimulationsCreateCmd.Flags().String("source-step-id", "", "specific iteration step id to source inputs from")
	workflowsSimulationsCreateCmd.Flags().String("n-consensus", "", "override n_consensus for extract, split, or classifier blocks (3, 5, or 7)")
	workflowsSimulationsCreateCmd.Flags().Bool("no-check-eligibility", false, "skip upstream drift eligibility checks")
	_ = workflowsSimulationsCreateCmd.MarkFlagRequired("block-id")

	workflowsSimulationsListCmd.Flags().String("block-id", "", "block id to list simulations for (required)")
	workflowsSimulationsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	_ = workflowsSimulationsListCmd.MarkFlagRequired("block-id")

	workflowsSimulationsCmd.AddCommand(workflowsSimulationsCreateCmd, workflowsSimulationsListCmd)
	workflowsCmd.AddCommand(workflowsSimulationsCmd)
}
