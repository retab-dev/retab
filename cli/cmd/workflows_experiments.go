package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsExperimentsCmd = &cobra.Command{
	Use:   "experiments",
	Short: "Manage block experiments",
	Long: `A/B-style alternate block configurations layered on a workflow
for measuring quality differences.

An experiment owns a candidate block config and a set of input
documents (either captured from past production runs, or supplied
explicitly). Running the experiment executes the candidate config
against every document and stores the outputs in isolation —
production traffic is never affected.

Compare experiment outputs to a baseline via
` + "`workflows experiments metrics`" + `. Run experiments for a whole
block in one shot with ` + "`run-batch`" + `, or trigger a single run
inside an experiment via ` + "`workflows experiments runs create`" + `.

For deterministic regression testing of a single pinned assertion, see
` + "`retab workflows tests --help`" + `.`,
	Example: `  # See which blocks support experiments
  retab workflows experiments eligible-blocks wf_abc123

  # Create an experiment on one block, capturing documents from real runs
  retab workflows experiments create \
    --workflow-id wf_abc123 --block-id blk_extract_1 \
    --name "Tighter schema v2" \
    --captures-file ./captures.json

  # Inspect quality metrics
  retab workflows experiments metrics wf_abc123 exp_pqr678 --view summary`,
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
	Long: `Create an experiment scoped to one block. Provide the input
documents in one of two ways:

  ` + "`--captures-file`" + `  — a JSON array of
  ` + `{"workflow_run_id": ..., "step_id": ...}` + ` entries pointing at
  existing production runs. The exact input that block saw will be
  replayed.

  ` + "`--documents-file`" + ` — a JSON array of explicit
  ` + `{"handle_inputs": ..., "provenance": ...}` + ` entries.

After creation, trigger a run with
` + "`workflows experiments runs create`" + ` or run the whole block's
experiments together via ` + "`workflows experiments run-batch`" + `.`,
	Example: `  # Capture documents from real production runs
  retab workflows experiments create \
    --workflow-id wf_abc123 --block-id blk_extract_1 \
    --name "Try gpt-4o-mini" \
    --captures-file ./captures.json \
    --n-consensus 3`,
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
	Long: `List every experiment attached to a workflow, across all its
blocks.`,
	Example: `  # List experiments in a workflow
  retab workflows experiments list wf_abc123`,
	Args: cobra.ExactArgs(1),
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
	Long: `Fetch an experiment's definition: target block, document set,
consensus count, recent run status.`,
	Example: `  # Inspect an experiment
  retab workflows experiments get wf_abc123 exp_pqr678`,
	Args: cobra.ExactArgs(2),
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
	Long: `Rename an experiment, adjust its consensus count, or replace
its document set. Note that updating the document set invalidates
previously-captured results for that experiment.`,
	Example: `  # Increase consensus to measure stability
  retab workflows experiments update wf_abc123 exp_pqr678 --n-consensus 5

  # Add more documents from production
  retab workflows experiments update wf_abc123 exp_pqr678 \
    --captures-file ./more-captures.json`,
	Args: cobra.ExactArgs(2),
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
	Long: `Permanently delete an experiment and its run history. Captured
production runs and artifacts are unaffected.`,
	Example: `  # Drop an experiment
  retab workflows experiments delete wf_abc123 exp_pqr678`,
	Args: cobra.ExactArgs(2),
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
	Long: `Fork an existing experiment with the same document set and
config — convenient when iterating on minor variations without losing
the previous experiment's results.`,
	Example: `  # Fork an experiment for tweaking
  retab workflows experiments duplicate wf_abc123 exp_pqr678`,
	Args: cobra.ExactArgs(2),
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
	Long: `Stop the most recent experiment run mid-flight. Documents
already processed retain their outputs; pending documents are skipped.`,
	Example: `  # Stop a long-running experiment
  retab workflows experiments cancel wf_abc123 exp_pqr678`,
	Args: cobra.ExactArgs(2),
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
	Long: `Aggregate quality metrics for an experiment's runs. Pivot the
view with ` + "`--view`" + ` (` + "`summary`" + ` | ` + "`per_run`" + ` |
` + "`per_document`" + ` | ` + "`per_field`" + `) to drill from headline
numbers down to individual fields. Compare against a prior run with
` + "`--prior-run-id`" + `.`,
	Example: `  # Headline numbers
  retab workflows experiments metrics wf_abc123 exp_pqr678 --view summary

  # Per-field accuracy
  retab workflows experiments metrics wf_abc123 exp_pqr678 --view per_field

  # Compare two runs of the same experiment
  retab workflows experiments metrics wf_abc123 exp_pqr678 \
    --run-id exprun_aaa --prior-run-id exprun_bbb --view per_field`,
	Args: cobra.ExactArgs(2),
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
	Long: `List the blocks in a workflow that can host an experiment.
Not every block type is eligible — typically ` + "`extract`" + `,
` + "`classify`" + `, and other LLM-backed blocks are supported. Use this
to discover what to experiment on before calling
` + "`workflows experiments create`" + `.`,
	Example: `  # See which blocks support experiments
  retab workflows experiments eligible-blocks wf_abc123`,
	Args: cobra.ExactArgs(1),
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
	Long: `Trigger a run for every experiment attached to one block, in one
call. Use when iterating across multiple candidate configs for the same
block — kick off the whole comparison sweep at once, then read metrics.`,
	Example: `  # Run every experiment on a block
  retab workflows experiments run-batch \
    --workflow-id wf_abc123 --block-id blk_extract_1 --n-consensus 3`,
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
	Long: `Trigger, list, and inspect individual experiment executions.
Each run processes every document in the experiment's document set
against the candidate block config and stores per-document outputs.

These runs are isolated from production workflow runs — they don't
appear in ` + "`retab workflows runs list`" + ` and don't affect downstream
consumers.`,
	Example: `  # Trigger a new run on an experiment
  retab workflows experiments runs create wf_abc123 exp_pqr678 \
    --n-consensus 3

  # Inspect the latest run's per-document content
  retab workflows experiments runs get wf_abc123 exp_pqr678`,
}

var workflowsExperimentsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> <experiment-id>",
	Short: "Trigger a new experiment run",
	Long: `Execute the experiment's candidate config across every
document in its set. Use ` + "`--retry-failed-only`" + ` to re-run only
the documents that failed in the previous run — cheaper than a full
re-run when iterating on a flaky config.`,
	Example: `  # Full run with 3-sample consensus
  retab workflows experiments runs create wf_abc123 exp_pqr678 \
    --n-consensus 3

  # Retry just the failed documents from the last run
  retab workflows experiments runs create wf_abc123 exp_pqr678 \
    --retry-failed-only`,
	Args: cobra.ExactArgs(2),
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
	Long: `List historical executions of one experiment, in reverse
chronological order.`,
	Example: `  # See the run history
  retab workflows experiments runs list wf_abc123 exp_pqr678`,
	Args: cobra.ExactArgs(2),
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
	Long: `Return per-document outputs for one experiment run — the actual
candidate-config output for every document in the experiment's set.
Without ` + "`--run-id`" + ` the latest run's content is returned.`,
	Example: `  # Latest run content
  retab workflows experiments runs get wf_abc123 exp_pqr678

  # A specific run
  retab workflows experiments runs get wf_abc123 exp_pqr678 \
    --run-id exprun_aaa`,
	Args: cobra.ExactArgs(2),
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
