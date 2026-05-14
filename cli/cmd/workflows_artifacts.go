package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsArtifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Inspect workflow artifacts",
}

var workflowsArtifactsGetCmd = &cobra.Command{
	Use:   "get <operation> <artifact-id>",
	Short: "Get an artifact by operation and id",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
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
	Use:   "list",
	Short: "List artifacts for a run",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowArtifactsParams{}
		params.RunID, _ = cmd.Flags().GetString("run-id")
		params.Operation, _ = cmd.Flags().GetString("operation")
		params.BlockID, _ = cmd.Flags().GetString("block-id")
		result, err := client.Workflows.Artifacts.List(ctx, params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	workflowsArtifactsListCmd.Flags().String("run-id", "", "run id (required)")
	workflowsArtifactsListCmd.Flags().String("operation", "", "filter by operation")
	workflowsArtifactsListCmd.Flags().String("block-id", "", "filter by block id")
	_ = workflowsArtifactsListCmd.MarkFlagRequired("run-id")

	workflowsArtifactsCmd.AddCommand(workflowsArtifactsGetCmd, workflowsArtifactsListCmd)
	workflowsCmd.AddCommand(workflowsArtifactsCmd)
}
