package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsExperimentsCmd = &cobra.Command{
	Use:   "experiments",
	Short: "Manage block experiments",
}

func parseExperimentDocs(cmd *cobra.Command) ([]retab.ExperimentDocumentCaptureRequest, []retab.ExplicitExperimentDocumentRequest, error) {
	var captures []retab.ExperimentDocumentCaptureRequest
	var explicit []retab.ExplicitExperimentDocumentRequest
	if path, _ := cmd.Flags().GetString("captures-file"); path != "" {
		arr, err := readJSONArray(path)
		if err != nil {
			return nil, nil, fmt.Errorf("--captures-file: %w", err)
		}
		for i, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("--captures-file[%d]: must be a JSON object", i)
			}
			cap := retab.ExperimentDocumentCaptureRequest{}
			if v, ok := obj["workflow_run_id"].(string); ok {
				cap.WorkflowRunID = v
			}
			if v, ok := obj["step_id"].(string); ok {
				cap.StepID = v
			}
			captures = append(captures, cap)
		}
	}
	if path, _ := cmd.Flags().GetString("documents-file"); path != "" {
		arr, err := readJSONArray(path)
		if err != nil {
			return nil, nil, fmt.Errorf("--documents-file: %w", err)
		}
		for i, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("--documents-file[%d]: must be a JSON object", i)
			}
			doc := retab.ExplicitExperimentDocumentRequest{}
			if v, ok := obj["handle_inputs"].(map[string]any); ok {
				doc.HandleInputs = v
			}
			if v, ok := obj["provenance"].(map[string]any); ok {
				prov := &retab.ExperimentDocumentProvenance{}
				if s, ok := v["workflow_run_id"].(string); ok {
					prov.WorkflowRunID = s
				}
				if s, ok := v["step_id"].(string); ok {
					prov.StepID = s
				}
				doc.Provenance = prov
			}
			explicit = append(explicit, doc)
		}
	}
	return captures, explicit, nil
}

var workflowsExperimentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an experiment",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.CreateExperimentRequest{}
		req.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		req.BlockID, _ = cmd.Flags().GetString("block-id")
		req.Name, _ = cmd.Flags().GetString("name")
		req.NConsensus, _ = cmd.Flags().GetInt("n-consensus")
		captures, explicit, err := parseExperimentDocs(cmd)
		if err != nil {
			return err
		}
		req.DocumentCaptures = captures
		req.Documents = explicit
		result, err := client.Workflows.Experiments.Create(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List experiments for a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.List(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <experiment-id>",
	Short: "Get an experiment",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsUpdateCmd = &cobra.Command{
	Use:   "update <workflow-id> <experiment-id>",
	Short: "Update an experiment",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.UpdateExperimentRequest{}
		req.Name, _ = cmd.Flags().GetString("name")
		if cmd.Flags().Changed("n-consensus") {
			v, _ := cmd.Flags().GetInt("n-consensus")
			req.NConsensus = &v
		}
		captures, explicit, err := parseExperimentDocs(cmd)
		if err != nil {
			return err
		}
		req.DocumentCaptures = captures
		req.Documents = explicit
		result, err := client.Workflows.Experiments.Update(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id> <experiment-id>",
	Short: "Delete an experiment",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Workflows.Experiments.Delete(ctx, args[0], args[1])
	}),
}

var workflowsExperimentsDuplicateCmd = &cobra.Command{
	Use:   "duplicate <workflow-id> <experiment-id>",
	Short: "Duplicate an experiment",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Duplicate(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsCancelCmd = &cobra.Command{
	Use:   "cancel <workflow-id> <experiment-id>",
	Short: "Cancel the latest in-flight run",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Cancel(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsMetricsCmd = &cobra.Command{
	Use:   "metrics <workflow-id> <experiment-id>",
	Short: "Get experiment metrics",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := &retab.GetExperimentMetricsParams{}
		params.View, _ = cmd.Flags().GetString("view")
		params.RunID, _ = cmd.Flags().GetString("run-id")
		params.DocumentID, _ = cmd.Flags().GetString("document-id")
		params.TargetPath, _ = cmd.Flags().GetString("target-path")
		params.PriorRunID, _ = cmd.Flags().GetString("prior-run-id")
		if cmd.Flags().Changed("include-prior") {
			v, _ := cmd.Flags().GetBool("include-prior")
			params.IncludePrior = &v
		}
		result, err := client.Workflows.Experiments.GetMetrics(ctx, args[0], args[1], params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsEligibleBlocksCmd = &cobra.Command{
	Use:   "eligible-blocks <workflow-id>",
	Short: "List blocks eligible for experiments",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.ListEligibleBlocks(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsRunBatchCmd = &cobra.Command{
	Use:   "run-batch",
	Short: "Run every experiment attached to a block",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.RunBatchExperimentsRequest{}
		req.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		req.BlockID, _ = cmd.Flags().GetString("block-id")
		req.NConsensus, _ = cmd.Flags().GetInt("n-consensus")
		result, err := client.Workflows.Experiments.RunBatch(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

// ---- experiment runs subgroup ----

var workflowsExperimentsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Manage experiment runs",
}

var workflowsExperimentsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> <experiment-id>",
	Short: "Trigger a new experiment run",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := &retab.RunExperimentOptions{}
		params.NConsensus, _ = cmd.Flags().GetInt("n-consensus")
		params.RetryFailedOnly, _ = cmd.Flags().GetBool("retry-failed-only")
		result, err := client.Workflows.Experiments.Runs.Create(ctx, args[0], args[1], params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsRunsListCmd = &cobra.Command{
	Use:   "list <workflow-id> <experiment-id>",
	Short: "List runs for an experiment",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Runs.List(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsExperimentsRunsGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <experiment-id>",
	Short: "Get per-document content for an experiment run (latest by default)",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		runID, _ := cmd.Flags().GetString("run-id")
		result, err := client.Workflows.Experiments.Runs.Get(ctx, args[0], args[1], runID)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	addExperimentDocFlags := func(c *cobra.Command) {
		c.Flags().String("captures-file", "", "JSON array of {workflow_run_id, step_id} captures (or - for stdin)")
		c.Flags().String("documents-file", "", "JSON array of {handle_inputs, provenance} (or - for stdin)")
	}

	workflowsExperimentsCreateCmd.Flags().String("workflow-id", "", "workflow id (required)")
	workflowsExperimentsCreateCmd.Flags().String("block-id", "", "block id (required)")
	workflowsExperimentsCreateCmd.Flags().String("name", "", "experiment name (required)")
	workflowsExperimentsCreateCmd.Flags().Int("n-consensus", 0, "consensus count")
	addExperimentDocFlags(workflowsExperimentsCreateCmd)
	_ = workflowsExperimentsCreateCmd.MarkFlagRequired("workflow-id")
	_ = workflowsExperimentsCreateCmd.MarkFlagRequired("block-id")
	_ = workflowsExperimentsCreateCmd.MarkFlagRequired("name")

	workflowsExperimentsUpdateCmd.Flags().String("name", "", "new name")
	workflowsExperimentsUpdateCmd.Flags().Int("n-consensus", 0, "new consensus count")
	addExperimentDocFlags(workflowsExperimentsUpdateCmd)

	workflowsExperimentsMetricsCmd.Flags().String("view", "", "view (summary | per_run | per_document | per_field)")
	workflowsExperimentsMetricsCmd.Flags().String("run-id", "", "run id")
	workflowsExperimentsMetricsCmd.Flags().String("document-id", "", "document id")
	workflowsExperimentsMetricsCmd.Flags().String("target-path", "", "target path")
	workflowsExperimentsMetricsCmd.Flags().String("prior-run-id", "", "prior run id")
	workflowsExperimentsMetricsCmd.Flags().Bool("include-prior", true, "include prior run metrics")

	workflowsExperimentsRunBatchCmd.Flags().String("workflow-id", "", "workflow id (required)")
	workflowsExperimentsRunBatchCmd.Flags().String("block-id", "", "block id (required)")
	workflowsExperimentsRunBatchCmd.Flags().Int("n-consensus", 0, "consensus count")
	_ = workflowsExperimentsRunBatchCmd.MarkFlagRequired("workflow-id")
	_ = workflowsExperimentsRunBatchCmd.MarkFlagRequired("block-id")

	workflowsExperimentsRunsCreateCmd.Flags().Int("n-consensus", 0, "consensus count")
	workflowsExperimentsRunsCreateCmd.Flags().Bool("retry-failed-only", false, "retry only failed documents")
	workflowsExperimentsRunsGetCmd.Flags().String("run-id", "", "specific run id (default: latest)")

	workflowsExperimentsRunsCmd.AddCommand(workflowsExperimentsRunsCreateCmd, workflowsExperimentsRunsListCmd, workflowsExperimentsRunsGetCmd)
	workflowsExperimentsCmd.AddCommand(workflowsExperimentsCreateCmd, workflowsExperimentsListCmd, workflowsExperimentsGetCmd, workflowsExperimentsUpdateCmd, workflowsExperimentsDeleteCmd, workflowsExperimentsDuplicateCmd, workflowsExperimentsCancelCmd, workflowsExperimentsMetricsCmd, workflowsExperimentsEligibleBlocksCmd, workflowsExperimentsRunBatchCmd, workflowsExperimentsRunsCmd)
	workflowsCmd.AddCommand(workflowsExperimentsCmd)
}
