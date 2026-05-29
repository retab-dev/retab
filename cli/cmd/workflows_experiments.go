//go:build !retab_oagen_cli_workflows_experiments

package cmd

import (
	"fmt"
	"os"
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
` + "`workflows experiments metrics get`" + `. Create a run inside an
experiment via ` + "`workflows experiments runs create`" + `.

For deterministic regression testing of a single pinned assertion, see
` + "`retab workflows tests --help`" + `.`,
	Example: `  # List experiments for a workflow
  retab workflows experiments list wf_abc123

  # Create an experiment on one block, capturing documents from real runs
  retab workflows experiments create wf_abc123 \
    --block-id blk_extract_1 \
    --name "Tighter schema v2" \
    --captures-file ./captures.json

  # Inspect quality metrics
  retab workflows experiments metrics get exprun_abc --view summary`,
}

func parseExperimentDocs(cmd *cobra.Command) ([]*retab.ExperimentDocumentCaptureRequest, []*retab.ExplicitExperimentDocumentRequest, error) {
	var captures []*retab.ExperimentDocumentCaptureRequest
	var explicit []*retab.ExplicitExperimentDocumentRequest
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
			cap := &retab.ExperimentDocumentCaptureRequest{}
			if v, ok := obj["workflow_run_id"].(string); ok {
				cap.WorkflowRunID = v
			}
			if cap.WorkflowRunID == "" {
				return nil, nil, fmt.Errorf("--captures-file[%d]: workflow_run_id is required", i)
			}
			if v, ok := obj["step_id"].(string); ok && v != "" {
				stepID := v
				cap.StepID = &stepID
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
			doc := &retab.ExplicitExperimentDocumentRequest{}
			if v, ok := obj["handle_inputs"].(map[string]any); ok {
				doc.HandleInputs = experimentHandleInputsFromMap(v)
			}
			if doc.HandleInputs == nil {
				return nil, nil, fmt.Errorf("--documents-file[%d]: handle_inputs is required", i)
			}
			if v, ok := obj["provenance"].(map[string]any); ok {
				prov := &retab.ExperimentDocumentProvenance{}
				if s, ok := v["workflow_run_id"].(string); ok && s != "" {
					runID := s
					prov.WorkflowRunID = &runID
				}
				if s, ok := v["step_id"].(string); ok && s != "" {
					stepID := s
					prov.StepID = &stepID
				}
				doc.Provenance = prov
			}
			explicit = append(explicit, doc)
		}
	}
	return captures, explicit, nil
}

// experimentHandleInputsFromMap turns a parsed JSON object
// ({"handle_name": <any JSON value>}) into the typed shape the SDK expects.
// The wire shape carries an optional `type` discriminator on each value, but
// the CLI's JSON descriptors have historically used the raw value form; both
// shapes are normalized here so legacy descriptor files keep working.
func experimentHandleInputsFromMap(raw map[string]any) map[string]retab.HandleInput {
	if raw == nil {
		return nil
	}
	out := make(map[string]retab.HandleInput, len(raw))
	for key, value := range raw {
		input := retab.JSONHandleInput{}
		if obj, ok := value.(map[string]any); ok {
			if t, ok := obj["type"].(string); ok && t != "" {
				typeCopy := t
				input.Type = &typeCopy
				if data, ok := obj["data"]; ok {
					dataCopy := data
					input.Data = &dataCopy
				}
				out[key] = retab.HandleInputFromJSONHandleInput(input)
				continue
			}
		}
		dataCopy := value
		input.Data = &dataCopy
		out[key] = retab.HandleInputFromJSONHandleInput(input)
	}
	return out
}

func validateExperimentMetricsView(view string) error {
	switch view {
	case "", "summary", "by_document", "by_target", "votes":
		return nil
	default:
		return fmt.Errorf("invalid --view %q (want: summary | by_document | by_target | votes)", view)
	}
}

// validateExperimentName trims surrounding whitespace, rejects blank-only
// values, and returns the cleaned name. Callers MUST use the returned
// value when populating the outgoing request — otherwise
// “experiments create --name "  padded  "“ persists with the padding
// (bug #3, same shape as the workflow-name fix).
func validateExperimentName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("experiment name is required")
	}
	return trimmed, nil
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
		// --captures-file and --documents-file are described in the help
		// text as alternatives ("Provide the input documents in one of two
		// ways"). Reject the combination client-side before any file I/O
		// or network call so users don't silently get only one of the two
		// sources used.
		if err := validateMutuallyExclusiveChangedFlags(cmd, "captures-file", "documents-file"); err != nil {
			return err
		}
		req := retab.WorkflowExperimentsCreateParams{}
		req.WorkflowID = workflowID
		req.BlockID = &blockID
		rawName, _ := cmd.Flags().GetString("name")
		trimmedName, err := validateExperimentName(rawName)
		if err != nil {
			return err
		}
		req.Name = &trimmedName
		if cmd.Flags().Changed("n-consensus") {
			v, _ := cmd.Flags().GetInt("n-consensus")
			n := retab.CreateExperimentRequestNConsensus(v)
			req.NConsensus = &n
		}
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
		result, err := client.Workflows.Experiments.Create(ctx, &req)
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
		result, err := client.Workflows.Experiments.List(ctx, &retab.WorkflowExperimentsListParams{WorkflowID: args[0]})
		if err != nil {
			return err
		}
		return printWorkflowExperimentsListResult(cmd, result)
	}),
}

// printWorkflowExperimentsListResult routes experiment list rendering to
// a dedicated TableColumn spec when --output table is selected. The
// generic auto-column picker maps the TYPE column to whichever alias
// fires first (`status` wins over `block_kind` on experiment payloads),
// which produces a column that varies confusingly between resource
// shapes. The explicit BLOCK_KIND column makes the source field obvious
// (bug #12) — extract / split / classifier / … — and aligns with how
// the SDK names the field on the wire.
func printWorkflowExperimentsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowExperiment]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputTable) {
			return RenderList(os.Stdout, OutputTable, result, workflowExperimentColumns)
		}
	}
	return printJSON(result)
}

// workflowExperimentColumns is the dedicated TableColumn spec for
// `workflows experiments list --output table`. BLOCK_KIND replaces the
// always-empty TYPE column the generic auto-renderer used to render
// before bug #12 was fixed; the value comes from the `block_kind` field
// on each Experiment record (extract / split / classifier / …).
var workflowExperimentColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowExperimentCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return workflowExperimentCell(row, "name") }},
	{Header: "BLOCK_KIND", Extract: func(row any) string { return workflowExperimentCell(row, "block_kind") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowExperimentCell(row, "status") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowExperimentCell(row, "created_at") }},
}

func workflowExperimentCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

var workflowsExperimentsGetCmd = &cobra.Command{
	Use:   "get <experiment-id>",
	Short: "Get an experiment",
	Long: `Fetch an experiment's definition: target block, document set,
consensus count, recent run status.`,
	Example: `  # Inspect an experiment
  retab workflows experiments get exp_pqr678`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsUpdateCmd = &cobra.Command{
	Use:   "update <experiment-id>",
	Short: "Update an experiment",
	Long: `Rename an experiment, adjust its consensus count, or replace
its document set. Note that updating the document set invalidates
previously-captured results for that experiment.`,
	Example: `  # Increase consensus to measure stability
  retab workflows experiments update exp_pqr678 --n-consensus 5

  # Add more documents from production
  retab workflows experiments update exp_pqr678 \
    --captures-file ./more-captures.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("n-consensus") &&
			!cmd.Flags().Changed("captures-file") && !cmd.Flags().Changed("documents-file") {
			return fmt.Errorf("nothing to update: pass at least one of --name, --n-consensus, --captures-file, or --documents-file")
		}
		req := retab.WorkflowExperimentsUpdateParams{}
		if cmd.Flags().Changed("name") {
			rawName, _ := cmd.Flags().GetString("name")
			trimmed, err := validateExperimentName(rawName)
			if err != nil {
				return err
			}
			req.Name = &trimmed
		}
		if cmd.Flags().Changed("n-consensus") {
			v, _ := cmd.Flags().GetInt("n-consensus")
			n := retab.UpdateExperimentRequestNConsensus(v)
			req.NConsensus = &n
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
		result, err := client.Workflows.Experiments.Update(ctx, args[0], &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsDeleteCmd = &cobra.Command{
	Use:   "delete <experiment-id>",
	Short: "Delete an experiment",
	Long: `Permanently delete an experiment and its run history. Captured
production runs and artifacts are unaffected.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal. Run history is removed alongside the experiment
definition.`,
	Example: `  # Drop an experiment (interactive, asks to confirm)
  retab workflows experiments delete exp_pqr678

  # Skip the prompt in scripts
  retab workflows experiments delete exp_pqr678 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "experiment", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Experiments.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("experiment", args[0])
		return nil
	}),
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
  retab workflows experiments runs create exp_pqr678

  # Inspect a run
  retab workflows experiments runs get exprun_aaa

  # Inspect run results
  retab workflows experiments results list exprun_aaa`,
}

var workflowsExperimentsRunsCreateCmd = &cobra.Command{
	Use:   "create <experiment-id>",
	Short: "Create a new experiment run",
	Long: `Create an experiment run that evaluates the candidate config across every
document in its set.

The workflow id is derived server-side from the experiment record, so
only the ` + "`<experiment-id>`" + ` positional is required. The legacy
two-argument form (` + "`<workflow-id> <experiment-id>`" + `) is still
accepted for backward compatibility — when supplied, the workflow id is
forwarded to the server for cross-checking against the experiment's own
` + "`workflow_id`" + ` so a mismatched pairing surfaces as a 404 instead
of being silently dropped.`,
	Example: `  # Create an experiment run
  retab workflows experiments runs create exp_pqr678`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Backward compat: the historical shape was
		//   `runs create <workflow-id> <experiment-id>`
		// The server derives `workflow_id` from the experiment, but it
		// ALSO validates a client-supplied `workflow_id` matches when one
		// is sent — so we forward the leading positional when present.
		// Without this, a typo in the workflow id slot was silently
		// dropped (the experiment id alone determined which workflow ran).
		experimentID := args[0]
		params := retab.ExperimentRunsCreateParams{ExperimentID: experimentID}
		if len(args) == 2 {
			experimentID = args[1]
			workflowID := args[0]
			params.ExperimentID = experimentID
			params.WorkflowID = &workflowID
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Runs.Create(ctx, &params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsRunsListCmd = &cobra.Command{
	Use:   "list [workflow-id] [experiment-id]",
	Short: "List experiment runs",
	Long: `List historical executions of one experiment, in reverse
chronological order.

The two positional arguments mirror ` + "`workflows experiments runs create`" + `:
the parent ids ` + "`<workflow-id>`" + ` and ` + "`<experiment-id>`" + ` are
positional, filters are flags — same convention as the rest of the
` + "`workflows <X> list`" + ` commands. The flag forms (` + "`--workflow-id`" + `,
` + "`--experiment-id`" + `) are still accepted for back-compat.`,
	Example: `  # Positional (recommended; matches ` + "`runs create`" + ` signature)
  retab workflows experiments runs list wf_abc123 exp_pqr678

  # Flag form (still accepted)
  retab workflows experiments runs list --workflow-id wf_abc123 --experiment-id exp_pqr678`,
	Args: cobra.MaximumNArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		flagWorkflowID, _ := cmd.Flags().GetString("workflow-id")
		flagExperimentID, _ := cmd.Flags().GetString("experiment-id")
		positionalWorkflowID := ""
		positionalExperimentID := ""
		if len(args) >= 1 {
			positionalWorkflowID = args[0]
		}
		if len(args) >= 2 {
			positionalExperimentID = args[1]
		}
		// Reject conflict: positional + flag form disagreeing is a
		// silent-misroute hazard (see the F-bug fix for the block create
		// route's body/path workflow_id mismatch — same shape).
		if positionalWorkflowID != "" && flagWorkflowID != "" && positionalWorkflowID != flagWorkflowID {
			return fmt.Errorf(
				"workflow id specified twice (positional %q, --workflow-id %q)",
				positionalWorkflowID, flagWorkflowID,
			)
		}
		if positionalExperimentID != "" && flagExperimentID != "" && positionalExperimentID != flagExperimentID {
			return fmt.Errorf(
				"experiment id specified twice (positional %q, --experiment-id %q)",
				positionalExperimentID, flagExperimentID,
			)
		}
		resolvedWorkflowID := flagWorkflowID
		if positionalWorkflowID != "" {
			resolvedWorkflowID = positionalWorkflowID
		}
		resolvedExperimentID := flagExperimentID
		if positionalExperimentID != "" {
			resolvedExperimentID = positionalExperimentID
		}
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		params := retab.ExperimentRunsListParams{
			PaginationParams: collectListParams(cmd),
		}
		if resolvedWorkflowID != "" {
			params.WorkflowID = &resolvedWorkflowID
		}
		if resolvedExperimentID != "" {
			params.ExperimentID = &resolvedExperimentID
		}
		if value, _ := cmd.Flags().GetString("block-id"); value != "" {
			params.BlockID = &value
		}
		if value, _ := cmd.Flags().GetString("status"); value != "" {
			status := retab.WorkflowExperimentsStatus(value)
			params.Status = &status
		}
		if value, _ := cmd.Flags().GetString("statuses"); value != "" {
			params.Statuses = &value
		}
		if value, _ := cmd.Flags().GetString("exclude-status"); value != "" {
			status := retab.WorkflowExperimentsExcludeStatus(value)
			params.ExcludeStatus = &status
		}
		if value, _ := cmd.Flags().GetString("trigger-type"); value != "" {
			params.TriggerType = &value
		}
		if value, _ := cmd.Flags().GetString("trigger-types"); value != "" {
			params.TriggerTypes = &value
		}
		if value, _ := cmd.Flags().GetString("from-date"); value != "" {
			params.FromDate = &value
		}
		if value, _ := cmd.Flags().GetString("to-date"); value != "" {
			params.ToDate = &value
		}
		if value, _ := cmd.Flags().GetString("sort-by"); value != "" {
			params.SortBy = &value
		}
		limit := getIntFlagOrDefault(cmd, "limit", 20)
		params.Limit = &limit
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Runs.List(ctx, &params)
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
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Runs.Get(ctx, args[0])
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
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Runs.Cancel(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Inspect experiment run results",
}

var workflowsExperimentsResultsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List per-document results for an experiment run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		limit := getIntFlagOrDefault(cmd, "limit", 20)
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Results.List(ctx, &retab.ExperimentRunResultsListParams{
			RunID: args[0],
			PaginationParams: retab.PaginationParams{
				Limit: &limit,
			},
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsResultsGetCmd = &cobra.Command{
	Use:   "get <result-id>",
	Short: "Get an experiment result",
	Long:  `Fetch one experiment result by flat result id.`,
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Results.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Inspect experiment run metrics",
}

var workflowsExperimentsMetricsGetCmd = &cobra.Command{
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
		documentID, _ := cmd.Flags().GetString("document-id")
		targetPath, _ := cmd.Flags().GetString("target-path")
		// Catch missing dependent flags client-side so users don't see
		// the server's raw 400. The legal combinations are:
		//   summary     — no extra requirement
		//   by_document — --document-id required
		//   votes       — --document-id required
		//   by_target   — --target-path required
		switch view {
		case "by_document", "votes":
			if strings.TrimSpace(documentID) == "" {
				return fmt.Errorf("--document-id is required when --view is %s", view)
			}
		case "by_target":
			if strings.TrimSpace(targetPath) == "" {
				return fmt.Errorf("--target-path is required when --view is by_target")
			}
		}
		params := retab.ExperimentRunMetricsGetParams{
			RunID: args[0],
		}
		if view != "" {
			typedView := retab.ExperimentRunMetricsView(view)
			params.View = &typedView
		}
		if documentID != "" {
			params.DocumentID = &documentID
		}
		if targetPath != "" {
			params.TargetPath = &targetPath
		}
		if value, _ := cmd.Flags().GetString("prior-run-id"); value != "" {
			params.PriorRunID = &value
		}
		includePrior, _ := cmd.Flags().GetBool("include-prior")
		params.IncludePrior = &includePrior
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Metrics.Get(ctx, &params)
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

	workflowsExperimentsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")
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
	workflowsExperimentsRunsListCmd.Flags().String("before", "", "page before cursor (mutually exclusive with --after)")
	workflowsExperimentsRunsListCmd.Flags().String("after", "", "page after cursor (mutually exclusive with --before)")
	workflowsExperimentsRunsListCmd.Flags().String("order", "", "asc or desc")
	workflowsExperimentsResultsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsExperimentsMetricsGetCmd.Flags().String("view", "summary", "view (summary | by_document | by_target | votes)")
	workflowsExperimentsMetricsGetCmd.Flags().String("document-id", "", "document id")
	workflowsExperimentsMetricsGetCmd.Flags().String("target-path", "", "target path")
	workflowsExperimentsMetricsGetCmd.Flags().String("prior-run-id", "", "prior run id")
	workflowsExperimentsMetricsGetCmd.Flags().Bool("include-prior", true, "include prior run metrics")

	workflowsExperimentsResultsCmd.AddCommand(workflowsExperimentsResultsListCmd, workflowsExperimentsResultsGetCmd)
	workflowsExperimentsMetricsCmd.AddCommand(workflowsExperimentsMetricsGetCmd)
	workflowsExperimentsRunsCmd.AddCommand(workflowsExperimentsRunsCreateCmd, workflowsExperimentsRunsListCmd, workflowsExperimentsRunsGetCmd, workflowsExperimentsRunsCancelCmd)
	workflowsExperimentsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsExperimentsCmd.AddCommand(workflowsExperimentsCreateCmd, workflowsExperimentsListCmd, workflowsExperimentsGetCmd, workflowsExperimentsUpdateCmd, workflowsExperimentsDeleteCmd, workflowsExperimentsRunsCmd, workflowsExperimentsResultsCmd, workflowsExperimentsMetricsCmd)
	workflowsCmd.AddCommand(workflowsExperimentsCmd)
}
