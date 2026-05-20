package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
` + "`workflows experiments runs metrics get`" + `. Create a run inside an
experiment via ` + "`workflows experiments runs create`" + `.

For deterministic regression testing of a single pinned assertion, see
` + "`retab workflows tests --help`" + `.`,
	Example: `  # See which blocks support experiments
  retab workflows experiments eligible-blocks wf_abc123

  # Create an experiment on one block, capturing documents from real runs
  retab workflows experiments create wf_abc123 \
    --block-id blk_extract_1 \
    --name "Tighter schema v2" \
    --captures-file ./captures.json

  # Inspect quality metrics
  retab workflows experiments runs metrics get exprun_abc --view summary`,
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
			if cap.WorkflowRunID == "" {
				return nil, nil, fmt.Errorf("--captures-file[%d]: workflow_run_id is required", i)
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
			if doc.HandleInputs == nil {
				return nil, nil, fmt.Errorf("--documents-file[%d]: handle_inputs is required", i)
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

func validateExperimentMetricsView(view string) error {
	switch view {
	case "", "summary", "by_document", "by_target", "votes":
		return nil
	default:
		return fmt.Errorf("invalid --view %q (want: summary | by_document | by_target | votes)", view)
	}
}

func validateExperimentName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("experiment name is required")
	}
	return nil
}

var workflowsExperimentsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> [flags]",
	Short: "Create an experiment",
	Long: `Create an experiment scoped to one block. Provide the input
documents in one of two ways:

  ` + "`--captures-file`" + `  — a JSON array of
  ` + `{"workflow_run_id": ..., "step_id": ...}` + ` entries pointing at
  existing production runs. The exact input that block saw will be
  replayed.

  ` + "`--documents-file`" + ` — a JSON array of explicit
  ` + `{"handle_inputs": ..., "provenance": ...}` + ` entries.

After creation, create a run with
` + "`workflows experiments runs create`" + `.`,
	Example: `  # Capture documents from real production runs
  retab workflows experiments create wf_abc123 \
    --block-id blk_extract_1 \
    --name "Try gpt-4o-mini" \
    --captures-file ./captures.json \
    --n-consensus 3`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, err := resolveWorkflowIDArg(cmd, args)
		if err != nil {
			return err
		}
		blockID, err := requireNonBlankFlag(cmd, "block-id")
		if err != nil {
			return err
		}
		req := retab.CreateExperimentRequest{}
		req.WorkflowID = workflowID
		req.BlockID = blockID
		req.Name, _ = cmd.Flags().GetString("name")
		if err := validateExperimentName(req.Name); err != nil {
			return err
		}
		req.NConsensus, _ = cmd.Flags().GetInt("n-consensus")
		captures, explicit, err := parseExperimentDocs(cmd)
		if err != nil {
			return err
		}
		req.DocumentCaptures = captures
		req.Documents = explicit
		if len(req.DocumentCaptures) == 0 && len(req.Documents) == 0 {
			return fmt.Errorf("at least one document or document capture is required (--captures-file or --documents-file)")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Create(ctx, req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
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
		return printResult(cmd, result)
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
		return printResult(cmd, result)
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
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("n-consensus") &&
			!cmd.Flags().Changed("captures-file") && !cmd.Flags().Changed("documents-file") {
			return fmt.Errorf("nothing to update: pass at least one of --name, --n-consensus, --captures-file, or --documents-file")
		}
		req := retab.UpdateExperimentRequest{}
		req.Name, _ = cmd.Flags().GetString("name")
		if cmd.Flags().Changed("name") {
			if err := validateExperimentName(req.Name); err != nil {
				return err
			}
		}
		if cmd.Flags().Changed("n-consensus") {
			v, _ := cmd.Flags().GetInt("n-consensus")
			if v == 0 {
				return fmt.Errorf("invalid --n-consensus 0 (want: 3 | 5 | 7)")
			}
			req.NConsensus = &v
		}
		captures, explicit, err := parseExperimentDocs(cmd)
		if err != nil {
			return err
		}
		req.DocumentCaptures = captures
		req.Documents = explicit
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Update(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
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
		if err := client.Workflows.Experiments.Delete(ctx, args[0], args[1]); err != nil {
			return err
		}
		confirmDeleted("experiment", args[1])
		return nil
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
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsEligibleBlocksCmd = &cobra.Command{
	Use:   "eligible-blocks <workflow-id>",
	Short: "List blocks eligible for experiments",
	Long: `List the blocks in a workflow that can host an experiment.
Not every block type is eligible — typically ` + "`extract`" + `,
` + "`classifier`" + `, and other LLM-backed blocks are supported. Use this
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
		return printEligibleBlocksResult(cmd, result)
	}),
}

func printEligibleBlocksResult(cmd *cobra.Command, result *retab.EligibleBlockListResponse) error {
	var raw string
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	if raw != string(OutputTable) {
		return printResult(cmd, result)
	}
	return printResult(cmd, map[string]any{"data": result.Blocks})
}

// ---- experiment runs subgroup ----

var workflowsExperimentsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Manage experiment runs",
	Long: `Create, list, and inspect individual experiment runs.
Each run processes every document in the experiment's document set
against the candidate block config and stores per-document outputs.

These runs are isolated from production workflow runs — they don't
appear in ` + "`retab workflows runs list`" + ` and don't affect downstream
consumers.`,
	Example: `  # Create a new experiment run
  retab workflows experiments runs create wf_abc123 exp_pqr678

  # Inspect a run
  retab workflows experiments runs get exprun_aaa

  # Inspect run results
  retab workflows experiments runs results list exprun_aaa`,
}

var workflowsExperimentsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> <experiment-id>",
	Short: "Create a new experiment run",
	Long: `Create an experiment run that evaluates the candidate config across every
document in its set.`,
	Example: `  # Create an experiment run
  retab workflows experiments runs create wf_abc123 exp_pqr678`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		result, err := cliJSONRequest(cmd, http.MethodPost, "/workflows/"+url.PathEscape(args[0])+"/experiments/"+url.PathEscape(args[1])+"/runs", nil, map[string]any{})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsRunsListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "List experiment runs",
	Long: `List historical executions of one experiment, in reverse
chronological order.`,
	Example: `  # See the run history
  retab workflows experiments runs list --experiment-id exp_pqr678`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		query := url.Values{}
		for _, name := range []string{"workflow-id", "experiment-id", "block-id", "status", "statuses", "exclude-status", "trigger-type", "trigger-types", "from-date", "to-date", "sort-by", "fields", "before", "after", "order"} {
			if value, _ := cmd.Flags().GetString(name); value != "" {
				query.Set(strings.ReplaceAll(name, "-", "_"), value)
			}
		}
		limit := getIntFlagOrDefault(cmd, "limit", 20)
		query.Set("limit", strconv.Itoa(limit))
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/experiments/runs", query, nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsRunsGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get an experiment run",
	Long:  `Fetch an experiment run by run id.`,
	Example: `  # A specific run
  retab workflows experiments runs get exprun_aaa`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(args[0]), nil, nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsRunsCancelCmd = &cobra.Command{
	Use:   "cancel <run-id>",
	Short: "Cancel an experiment run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		result, err := cliJSONRequest(cmd, http.MethodPost, "/workflows/experiments/runs/"+url.PathEscape(args[0])+"/cancel", nil, map[string]any{})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsRunsResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Inspect experiment run results",
}

var workflowsExperimentsRunsResultsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List per-document results for an experiment run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		query := url.Values{}
		limit := getIntFlagOrDefault(cmd, "limit", 20)
		query.Set("limit", strconv.Itoa(limit))
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(args[0])+"/results", query, nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsRunsResultsGetCmd = &cobra.Command{
	Use:   "get <run-id> <document-id>",
	Short: "Get one document result from an experiment run",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(args[0])+"/results/"+url.PathEscape(args[1]), nil, nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsRunsMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Inspect experiment run metrics",
}

var workflowsExperimentsRunsMetricsGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get metrics for an experiment run",
	Long: `Aggregate quality metrics for an experiment run. Pivot the
view with ` + "`--view`" + ` (` + "`summary`" + ` | ` + "`by_document`" + ` |
` + "`by_target`" + ` | ` + "`votes`" + `) to drill from headline
numbers down to individual fields. Compare against a prior run with
` + "`--prior-run-id`" + `.`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		view, _ := cmd.Flags().GetString("view")
		if err := validateExperimentMetricsView(view); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("view", view)
		if value, _ := cmd.Flags().GetString("document-id"); value != "" {
			query.Set("document_id", value)
		}
		if value, _ := cmd.Flags().GetString("target-path"); value != "" {
			query.Set("target_path", value)
		}
		if value, _ := cmd.Flags().GetString("prior-run-id"); value != "" {
			query.Set("prior_run_id", value)
		}
		includePrior, _ := cmd.Flags().GetBool("include-prior")
		query.Set("include_prior", strconv.FormatBool(includePrior))
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(args[0])+"/metrics", query, nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	addExperimentDocFlags := func(c *cobra.Command) {
		c.Flags().String("captures-file", "", "JSON array of {workflow_run_id, step_id} captures (or - for stdin)")
		c.Flags().String("documents-file", "", "JSON array of {handle_inputs, provenance} (or - for stdin)")
	}

	workflowsExperimentsCreateCmd.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	workflowsExperimentsCreateCmd.Flags().String("block-id", "", "block id (required)")
	workflowsExperimentsCreateCmd.Flags().String("name", "", "experiment name (required)")
	workflowsExperimentsCreateCmd.Flags().Var(&consensusFlagValue{}, "n-consensus", "consensus count (3, 5, or 7)")
	addExperimentDocFlags(workflowsExperimentsCreateCmd)
	// Keep the flag hidden but DO NOT use MarkDeprecated — cobra's auto warning
	// duplicates the more-specific message emitted by resolveWorkflowIDArg.
	_ = workflowsExperimentsCreateCmd.Flags().MarkHidden("workflow-id")
	_ = workflowsExperimentsCreateCmd.MarkFlagRequired("block-id")
	_ = workflowsExperimentsCreateCmd.MarkFlagRequired("name")

	workflowsExperimentsUpdateCmd.Flags().String("name", "", "new name")
	workflowsExperimentsUpdateCmd.Flags().Var(&consensusFlagValue{}, "n-consensus", "new consensus count (3, 5, or 7)")
	addExperimentDocFlags(workflowsExperimentsUpdateCmd)

	workflowsExperimentsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 0, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsExperimentsRunsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsExperimentsRunsListCmd.Flags().String("experiment-id", "", "filter by experiment id")
	workflowsExperimentsRunsListCmd.Flags().String("block-id", "", "filter by block id")
	workflowsExperimentsRunsListCmd.Flags().String("status", "", "filter by lifecycle status")
	workflowsExperimentsRunsListCmd.Flags().String("statuses", "", "comma-separated lifecycle statuses")
	workflowsExperimentsRunsListCmd.Flags().String("exclude-status", "", "exclude lifecycle status")
	workflowsExperimentsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsExperimentsRunsListCmd.Flags().String("trigger-types", "", "comma-separated trigger types")
	workflowsExperimentsRunsListCmd.Flags().String("from-date", "", "created on or after YYYY-MM-DD")
	workflowsExperimentsRunsListCmd.Flags().String("to-date", "", "created on or before YYYY-MM-DD")
	workflowsExperimentsRunsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsExperimentsRunsListCmd.Flags().String("fields", "", "comma-separated fields")
	workflowsExperimentsRunsListCmd.Flags().String("before", "", "page before cursor")
	workflowsExperimentsRunsListCmd.Flags().String("after", "", "page after cursor")
	workflowsExperimentsRunsListCmd.Flags().String("order", "", "asc or desc")
	workflowsExperimentsRunsResultsListCmd.Flags().Var(&boundedIntFlagValue{min: 0, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsExperimentsRunsMetricsGetCmd.Flags().String("view", "summary", "view (summary | by_document | by_target | votes)")
	workflowsExperimentsRunsMetricsGetCmd.Flags().String("document-id", "", "document id")
	workflowsExperimentsRunsMetricsGetCmd.Flags().String("target-path", "", "target path")
	workflowsExperimentsRunsMetricsGetCmd.Flags().String("prior-run-id", "", "prior run id")
	workflowsExperimentsRunsMetricsGetCmd.Flags().Bool("include-prior", true, "include prior run metrics")

	workflowsExperimentsRunsResultsCmd.AddCommand(workflowsExperimentsRunsResultsListCmd, workflowsExperimentsRunsResultsGetCmd)
	workflowsExperimentsRunsMetricsCmd.AddCommand(workflowsExperimentsRunsMetricsGetCmd)
	workflowsExperimentsRunsCmd.AddCommand(workflowsExperimentsRunsCreateCmd, workflowsExperimentsRunsListCmd, workflowsExperimentsRunsGetCmd, workflowsExperimentsRunsCancelCmd, workflowsExperimentsRunsResultsCmd, workflowsExperimentsRunsMetricsCmd)
	workflowsExperimentsCmd.AddCommand(workflowsExperimentsCreateCmd, workflowsExperimentsListCmd, workflowsExperimentsGetCmd, workflowsExperimentsUpdateCmd, workflowsExperimentsDeleteCmd, workflowsExperimentsDuplicateCmd, workflowsExperimentsEligibleBlocksCmd, workflowsExperimentsRunsCmd)
	workflowsCmd.AddCommand(workflowsExperimentsCmd)
}
