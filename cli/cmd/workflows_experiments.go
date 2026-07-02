//go:build !retab_oagen_cli_workflows_experiments

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
` + "`retab workflows evals --help`" + `.`,
	Example: `  # List experiments for a workflow
  retab workflows experiments list wf_abc123

  # Create an experiment on one block, capturing documents from real runs
  retab workflows experiments create wf_abc123 \
    --block-id block_extract_1 \
    --name "Tighter schema v2" \
    --captures-file ./captures.json

  # Inspect quality metrics
  retab workflows experiments metrics get exprun_abc --view summary`,
}

func parseExperimentDocs(cmd *cobra.Command) ([]*retab.ExperimentDocumentCaptureRequest, []*retab.ExplicitExperimentDocumentRequest, error) {
	var captures []*retab.ExperimentDocumentCaptureRequest
	var explicit []*retab.ExplicitExperimentDocumentRequest
	// --capture is the inline shortcut for the common case: a production run id
	// with an optional ":step_id" suffix, repeatable. It builds the same
	// capture entries as --captures-file so the two forms can coexist.
	if specs, _ := cmd.Flags().GetStringArray("capture"); len(specs) > 0 {
		for i, raw := range specs {
			runID, stepID, hasStep := strings.Cut(strings.TrimSpace(raw), ":")
			cap := &retab.ExperimentDocumentCaptureRequest{RunID: strings.TrimSpace(runID)}
			if cap.RunID == "" {
				return nil, nil, fmt.Errorf("--capture[%d]: run id is required (run-id or run-id:step-id)", i)
			}
			if hasStep {
				if stepID = strings.TrimSpace(stepID); stepID != "" {
					cap.StepID = &stepID
				}
			}
			captures = append(captures, cap)
		}
	}
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
			if v, ok := obj["run_id"].(string); ok {
				cap.RunID = v
			}
			if cap.RunID == "" {
				return nil, nil, fmt.Errorf("--captures-file[%d]: run_id is required", i)
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
				inputs, err := experimentHandleInputsFromMap(v)
				if err != nil {
					return nil, nil, fmt.Errorf("--documents-file[%d].handle_inputs: %w", i, err)
				}
				doc.HandleInputs = inputs
			}
			if doc.HandleInputs == nil {
				return nil, nil, fmt.Errorf("--documents-file[%d]: handle_inputs is required", i)
			}
			if v, ok := obj["provenance"].(map[string]any); ok {
				prov := &retab.ExperimentDocumentProvenance{}
				if s, ok := v["run_id"].(string); ok && s != "" {
					runID := s
					prov.RunID = &runID
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
func experimentHandleInputsFromMap(raw map[string]any) (map[string]retab.HandleInputType, error) {
	if raw == nil {
		return nil, nil
	}
	out := make(map[string]retab.HandleInputType, len(raw))
	for key, value := range raw {
		input := retab.JSONHandleInput{}
		if obj, ok := value.(map[string]any); ok {
			if t, ok := obj["type"].(string); ok && t != "" {
				switch t {
				case "file":
					fileInput, err := experimentFileHandleInputFromMap(obj)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", key, err)
					}
					out[key] = retab.HandleInputTypeFromFileHandleInput(fileInput)
					continue
				case "json":
					if data, ok := obj["data"]; ok {
						dataCopy := data
						input.Data = &dataCopy
					}
					out[key] = retab.HandleInputTypeFromJSONHandleInput(input)
					continue
				default:
					return nil, fmt.Errorf("%s: unsupported handle input type %q (want: file | json)", key, t)
				}
			}
		}
		dataCopy := value
		input.Data = &dataCopy
		out[key] = retab.HandleInputTypeFromJSONHandleInput(input)
	}
	return out, nil
}

func experimentFileHandleInputFromMap(obj map[string]any) (retab.FileHandleInput, error) {
	rawDocument, ok := obj["document"].(map[string]any)
	if !ok {
		return retab.FileHandleInput{}, fmt.Errorf("file handle input requires document")
	}
	id, _ := rawDocument["id"].(string)
	filename, _ := rawDocument["filename"].(string)
	mimeType, _ := rawDocument["mime_type"].(string)
	if strings.TrimSpace(id) == "" {
		return retab.FileHandleInput{}, fmt.Errorf("file handle document.id is required")
	}
	if strings.TrimSpace(filename) == "" {
		return retab.FileHandleInput{}, fmt.Errorf("file handle document.filename is required")
	}
	if strings.TrimSpace(mimeType) == "" {
		return retab.FileHandleInput{}, fmt.Errorf("file handle document.mime_type is required")
	}
	return retab.FileHandleInput{
		Document: retab.ResultFileRef{
			ID:       id,
			Filename: filename,
			MIMEType: mimeType,
		},
	}, nil
}

// resolveExperimentIDArg implements the uniform positional contract shared
// across the `experiments` command group. Sibling commands had diverged:
// `runs create` (RangeArgs 1-2) and `runs list` (MaximumNArgs 2) already
// accepted the convenience `<workflow-id> <experiment-id>` form, while
// `get` / `update` / `delete` were ExactArgs(1) and rejected it with cobra's
// opaque "accepts 1 arg(s), received 2". That made the two-positional form
// non-uniform across the group and surfaced as a confusing failure — and,
// when piped (`... 2>/dev/null | jq`), as a silent empty result.
//
// The experiment id is globally unique and is the SOLE selector the SDK
// Get/Update/Delete methods accept, so it is always the LAST positional.
// A leading `<workflow-id>`, when supplied, is accepted purely for symmetry
// with `runs create`: unlike `runs list` (where workflow-id is a server-side
// FILTER that changes which rows return) the experiment id alone fully
// determines the resource here, so a workflow id in the first slot cannot
// misroute to a different experiment and is intentionally not forwarded.
func resolveExperimentIDArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("experiment id required")
	}
	id := strings.TrimSpace(args[len(args)-1])
	if id == "" {
		return "", fmt.Errorf("experiment id required")
	}
	return id, nil
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
  ` + `{"run_id": ..., "step_id": ...}` + ` entries pointing at
  existing production runs. The exact input that block saw will be
  replayed.

  ` + "`--documents-file`" + ` — a JSON array of explicit
  ` + `{"handle_inputs": ..., "provenance": ...}` + ` entries.

After creation, create a run with
` + "`workflows experiments runs create`" + `.`,
	Example: `  # Capture documents from real production runs
  retab workflows experiments create wf_abc123 \
    --block-id block_extract_1 \
    --name "Try gpt-4o-mini" \
    --captures-file ./captures.json \
    --n-consensus 3`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, err := resolveWorkflowIDArg(cmd, args)
		if err != nil {
			return err
		}
		req := retab.WorkflowExperimentsCreateParams{}
		req.WorkflowID = workflowID

		fromRaw, _ := cmd.Flags().GetString("from-experiment")
		if sourceID := strings.TrimSpace(fromRaw); sourceID != "" {
			// The server duplicates the source experiment wholesale (block,
			// name + "(Copy)", consensus, documents) and rejects any other
			// field. Reject the conflicting flags here with a clear message
			// instead of forwarding a request the server is bound to 400.
			for _, f := range []string{"block-id", "name", "n-consensus", "capture", "captures-file", "documents-file"} {
				if cmd.Flags().Changed(f) {
					return fmt.Errorf("--from-experiment clones an existing experiment and can't be combined with --%s", f)
				}
			}
			req.SourceExperimentID = &sourceID
		} else {
			blockID, err := requireNonBlankFlag(cmd, "block-id")
			if err != nil {
				return fmt.Errorf("--block-id is required (or pass --from-experiment to clone an existing experiment)")
			}
			req.BlockID = &blockID
			rawName, _ := cmd.Flags().GetString("name")
			trimmedName, err := validateExperimentName(rawName)
			if err != nil {
				return err
			}
			req.Name = &trimmedName
			// --captures-file and --documents-file are described in the help
			// text as alternatives ("Provide the input documents in one of two
			// ways"). Reject the combination client-side before any file I/O
			// or network call so users don't silently get only one of the two
			// sources used.
			if err := validateMutuallyExclusiveChangedFlags(cmd, "captures-file", "documents-file"); err != nil {
				return err
			}
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
			if len(captures) == 0 && len(explicit) == 0 {
				return fmt.Errorf("at least one document is required: pass --capture <run-id[:step-id]>, --captures-file, or --documents-file (or --from-experiment to clone)")
			}
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
		// Without --run, creating the experiment definition is the whole job.
		if run, _ := cmd.Flags().GetBool("run"); !run {
			return printResult(cmd, result)
		}
		// --run: immediately launch a run against the new experiment, reusing
		// the same wait / terminal-status machinery as `runs create --wait` so
		// "create it and run it" is a single command.
		runResult, err := client.Workflows.Experiments.Runs.Create(ctx, &retab.ExperimentRunsCreateParams{ExperimentID: result.ID})
		if err != nil {
			return err
		}
		if wait, _ := cmd.Flags().GetBool("wait"); !wait {
			return printResult(cmd, runResult)
		}
		pollInterval, timeout := experimentWaitDurations(cmd)
		final, waitErr := waitForExperimentRun(ctx, client, runResult.ID, pollInterval, timeout)
		if final != nil {
			if err := printResult(cmd, final); err != nil {
				return err
			}
		}
		if waitErr != nil {
			return waitErr
		}
		return experimentRunTerminalError(final)
	}),
}

var workflowsExperimentsListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List experiments for a workflow",
	Long: `List every experiment attached to a workflow, across all its
blocks.

Name the workflow either positionally (` + "`list <workflow-id>`" + `) or with
the ` + "`--workflow-id`" + ` flag — the two forms are equivalent. Passing both
is accepted when they agree; an error is raised only when they disagree. The
workflow id is required: experiments have no org-wide listing.`,
	Example: `  # List experiments in a workflow (positional)
  retab workflows experiments list wf_abc123

  # Same, with the flag form
  retab workflows experiments list --workflow-id wf_abc123`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Workflow id positionally OR via --workflow-id (co-equal forms);
		// required here — experiments have no org-wide listing.
		workflowID, err := resolveWorkflowScope(cmd, args, true)
		if err != nil {
			return err
		}
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := &retab.WorkflowExperimentsListParams{WorkflowID: workflowID}
		if before, _ := cmd.Flags().GetString("before"); before != "" {
			params.Before = ptr(before)
		}
		if after, _ := cmd.Flags().GetString("after"); after != "" {
			params.After = ptr(after)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			params.Limit = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("order"); v != "" {
			params.Order = ptr(v)
		}
		result, err := client.Workflows.Experiments.List(ctx, params)
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
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			switch f.Value.String() {
			case string(OutputTable), string(OutputCSV):
				return RenderList(os.Stdout, OutputFormat(f.Value.String()), result, workflowExperimentColumns)
			}
		}
	}
	return printJSON(result)
}

// workflowExperimentColumns is the dedicated TableColumn spec for
// `workflows experiments list --output table`. BLOCK_KIND replaces the
// always-empty TYPE column the generic auto-renderer used to render
// before bug #12 was fixed; the value comes from the `block_type` field
// on each Experiment record (extract / split / classifier / …). (The field
// is `block_type`, not `block_kind` — the latter was always empty.)
var workflowExperimentColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowExperimentCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return workflowExperimentCell(row, "name") }},
	{Header: "BLOCK_KIND", Extract: func(row any) string { return workflowExperimentCell(row, "block_type") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowExperimentCell(row, "status") }},
	{Header: "FRESHNESS", Extract: artifactFreshnessCell},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowExperimentCell(row, "created_at") }},
}

func workflowExperimentCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func artifactFreshnessCell(row any) string {
	value, ok := rowField(row, "freshness.status")
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

// experimentResultColumns is the dedicated TableColumn spec for
// `workflows experiments results list --output table/csv`. The generic
// auto-renderer only surfaced ID + a confusing TYPE column; these columns show
// the per-document result fields a user actually compares (which document, its
// lifecycle status, and how long it took). ExperimentResult has no verdict —
// experiments measure outputs for comparison, they don't pass/fail like tests.
var experimentResultColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowExperimentCell(row, "id") }},
	{Header: "DOCUMENT", Extract: func(row any) string { return workflowExperimentCell(row, "document_id") }},
	{Header: "BLOCK_KIND", Extract: func(row any) string { return workflowExperimentCell(row, "block_type") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowExperimentCell(row, "lifecycle.status") }},
	{Header: "DURATION_MS", Extract: func(row any) string { return workflowExperimentCell(row, "timing.duration_ms") }},
}

// printExperimentResultsList renders the results page through the dedicated
// column spec for table/csv, falling back to JSON otherwise (matching the
// experiments-list renderer).
func printExperimentResultsList(cmd *cobra.Command, result *retab.PaginatedList[retab.ExperimentResult]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			switch f.Value.String() {
			case string(OutputTable), string(OutputCSV):
				return RenderList(os.Stdout, OutputFormat(f.Value.String()), result, experimentResultColumns)
			}
		}
	}
	return printJSON(result)
}

// experimentRunColumns is the dedicated TableColumn spec for
// `workflows experiments runs list --output table/csv`. The generic renderer
// picked ID + TYPE + CREATED_AT and hid run status/counts, which are the fields
// users need when monitoring experiment progress.
var experimentRunColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowExperimentCell(row, "id") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowExperimentCell(row, "lifecycle.status") }},
	{Header: "BLOCK_KIND", Extract: func(row any) string { return workflowExperimentCell(row, "block_type") }},
	{Header: "DOCS", Extract: experimentRunDocumentCountCell},
	{Header: "DONE", Extract: func(row any) string { return workflowExperimentCell(row, "completed_document_count") }},
	{Header: "ERRORS", Extract: func(row any) string { return workflowExperimentCell(row, "error_count") }},
	{Header: "SCORE", Extract: func(row any) string { return workflowExperimentCell(row, "score") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowExperimentCell(row, "timing.created_at") }},
}

func experimentRunDocumentCountCell(row any) string {
	if value := workflowExperimentCell(row, "total_document_count"); value != "" {
		return value
	}
	return workflowExperimentCell(row, "document_count")
}

func printExperimentRunsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.ExperimentRun]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			switch f.Value.String() {
			case string(OutputTable), string(OutputCSV):
				return RenderList(os.Stdout, OutputFormat(f.Value.String()), result, experimentRunColumns)
			}
		}
	}
	return printJSON(result)
}

var workflowsExperimentsGetCmd = &cobra.Command{
	Use:   "get [workflow-id] <experiment-id>",
	Short: "Get an experiment",
	Long: `Fetch an experiment's definition: target block, document set,
consensus count, recent run status.

The experiment id is the only required positional. A leading
` + "`<workflow-id>`" + ` is also accepted for symmetry with
` + "`workflows experiments runs create`" + `, so the same
` + "`<workflow-id> <experiment-id>`" + ` pair works across the group.`,
	Example: `  # Inspect an experiment
  retab workflows experiments get exp_pqr678

  # Two-positional form (matches ` + "`runs create`" + `) also works
  retab workflows experiments get wf_abc123 exp_pqr678`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		experimentID, err := resolveExperimentIDArg(args)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Experiments.Get(ctx, experimentID)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsUpdateCmd = &cobra.Command{
	Use:   "update [workflow-id] <experiment-id>",
	Short: "Update an experiment",
	Long: `Rename an experiment, adjust its consensus count, or replace
its document set. Note that updating the document set invalidates
previously-captured results for that experiment.

The experiment id is the only required positional. A leading
` + "`<workflow-id>`" + ` is also accepted for symmetry with
` + "`workflows experiments runs create`" + `.`,
	Example: `  # Increase consensus to measure stability
  retab workflows experiments update exp_pqr678 --n-consensus 5

  # Add more documents from production
  retab workflows experiments update exp_pqr678 \
    --captures-file ./more-captures.json`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		experimentID, err := resolveExperimentIDArg(args)
		if err != nil {
			return err
		}
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("n-consensus") &&
			!cmd.Flags().Changed("capture") && !cmd.Flags().Changed("captures-file") && !cmd.Flags().Changed("documents-file") {
			return fmt.Errorf("nothing to update: pass at least one of --name, --n-consensus, --capture, --captures-file, or --documents-file")
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
		result, err := client.Workflows.Experiments.Update(ctx, experimentID, &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsExperimentsDeleteCmd = &cobra.Command{
	Use:   "delete [workflow-id] <experiment-id>",
	Short: "Delete an experiment",
	Long: `Permanently delete an experiment and its run history. Captured
production runs and artifacts are unaffected.

The experiment id is the only required positional. A leading
` + "`<workflow-id>`" + ` is also accepted for symmetry with
` + "`workflows experiments runs create`" + `.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal. Run history is removed alongside the experiment
definition.`,
	Example: `  # Drop an experiment (interactive, asks to confirm)
  retab workflows experiments delete exp_pqr678

  # Skip the prompt in scripts
  retab workflows experiments delete exp_pqr678 --yes`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		experimentID, err := resolveExperimentIDArg(args)
		if err != nil {
			return err
		}
		if err := confirmDestructive(cmd, "experiment", experimentID); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Experiments.Delete(ctx, experimentID); err != nil {
			return err
		}
		confirmDeleted("experiment", experimentID)
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
of being silently dropped.

By default the command returns as soon as the run is queued. Pass
` + "`--wait`" + ` to block until the run reaches a terminal status
(` + "`completed`" + `, ` + "`error`" + `, or ` + "`cancelled`" + `) and
print the final run record — saving you from scripting a poll loop around
` + "`runs get`" + `. With ` + "`--wait`" + ` the command exits non-zero if
the run ends in ` + "`error`" + `/` + "`cancelled`" + ` or the timeout
elapses.`,
	Example: `  # Create an experiment run
  retab workflows experiments runs create exp_pqr678

  # Create and block until it finishes (2s polls, 10m timeout)
  retab workflows experiments runs create exp_pqr678 --wait

  # Tune the polling cadence and ceiling
  retab workflows experiments runs create exp_pqr678 \
    --wait --poll-interval-ms 1000 --timeout-seconds 1800`,
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
		if wait, _ := cmd.Flags().GetBool("wait"); !wait {
			return printResult(cmd, result)
		}
		// --wait: poll the freshly-created run until it settles. A
		// consensus run over a document set takes 30s–2min, so blocking
		// here removes the most common hand-rolled `until` loop.
		pollInterval, timeout := experimentWaitDurations(cmd)
		final, waitErr := waitForExperimentRun(ctx, client, result.ID, pollInterval, timeout)
		if final != nil {
			if err := printResult(cmd, final); err != nil {
				return err
			}
		}
		if waitErr != nil {
			return waitErr
		}
		return experimentRunTerminalError(final)
	}),
}

var workflowsExperimentsRunsWaitCmd = &cobra.Command{
	Use:   "wait <run-id>",
	Short: "Poll until an experiment run reaches a terminal status",
	Long: `Block until an experiment run hits a terminal status
(` + "`completed`" + `, ` + "`error`" + `, or ` + "`cancelled`" + `),
polling on a configurable interval. Defaults: 2-second polls, 10-minute
timeout.

Cleaner than scripting a poll loop around ` + "`runs get`" + ` — the CLI
handles the interval and timeout, and exits non-zero if the run ends in
` + "`error`" + `/` + "`cancelled`" + ` or the timeout elapses. Pair with
` + "`runs create --wait`" + ` to create and block in a single step.`,
	Example: `  # Wait with defaults (2s polls, 600s timeout)
  retab workflows experiments runs wait exprun_aaa

  # Faster polls, longer ceiling
  retab workflows experiments runs wait exprun_aaa \
    --poll-interval-ms 1000 --timeout-seconds 1800`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		pollInterval, timeout := experimentWaitDurations(cmd)
		final, waitErr := waitForExperimentRun(ctx, client, args[0], pollInterval, timeout)
		if final != nil {
			if err := printResult(cmd, final); err != nil {
				return err
			}
		}
		if waitErr != nil {
			return waitErr
		}
		return experimentRunTerminalError(final)
	}),
}

// experimentWaitDurations resolves the poll cadence and timeout from the
// shared --poll-interval-ms / --timeout-seconds flags, applying the same
// defaults as experiment run polling (2s polls, 10m timeout).
func experimentWaitDurations(cmd *cobra.Command) (pollInterval, timeout time.Duration) {
	pollMS, _ := cmd.Flags().GetInt("poll-interval-ms")
	timeoutS, _ := cmd.Flags().GetInt("timeout-seconds")
	pollInterval = 2 * time.Second
	if pollMS > 0 {
		pollInterval = time.Duration(pollMS) * time.Millisecond
	}
	timeout = 10 * time.Minute
	if timeoutS > 0 {
		timeout = time.Duration(timeoutS) * time.Second
	}
	return pollInterval, timeout
}

// experimentRunStatus reads the lifecycle discriminator off an experiment
// run. The lifecycle is a discriminated-union envelope, so the status string
// comes from its `Status()` getter rather than a plain field.
func experimentRunStatus(run *retab.ExperimentRun) string {
	if run == nil {
		return ""
	}
	return run.Lifecycle.Status()
}

// isTerminalExperimentRun reports whether the run has settled — completed,
// error, or cancelled. pending/queued/running are still in flight.
func isTerminalExperimentRun(run *retab.ExperimentRun) bool {
	switch experimentRunStatus(run) {
	case "completed", "error", "cancelled":
		return true
	default:
		return false
	}
}

// experimentRunTerminalError maps a settled-but-unsuccessful run to a non-zero
// exit. completed (or an unknown/empty status) is success; error/cancelled is
// failure.
func experimentRunTerminalError(run *retab.ExperimentRun) error {
	if run == nil {
		return nil
	}
	status := experimentRunStatus(run)
	switch status {
	case "", "completed":
		return nil
	case "error", "cancelled":
		if run.ID == "" {
			return fmt.Errorf("experiment run ended with status %s", status)
		}
		return fmt.Errorf("experiment run %s ended with status %s", run.ID, status)
	default:
		return nil
	}
}

// waitForExperimentRun polls Runs.Get until the run reaches a terminal status
// or the timeout elapses. On timeout it returns the last-observed (non-final)
// run alongside the error so callers can still surface partial state.
func waitForExperimentRun(
	ctx context.Context,
	client *retab.Client,
	runID string,
	pollInterval time.Duration,
	timeout time.Duration,
) (*retab.ExperimentRun, error) {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		run, err := client.Workflows.Experiments.Runs.Get(ctx, runID)
		if err != nil {
			return nil, err
		}
		if isTerminalExperimentRun(run) {
			return run, nil
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return run, fmt.Errorf("timed out waiting for experiment run %s: %w", runID, ctx.Err())
		case <-timer.C:
		}
	}
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
		resolvedWorkflowID := flagWorkflowID
		resolvedExperimentID := flagExperimentID
		if len(args) == 1 {
			switch {
			case strings.HasPrefix(args[0], "exp_"):
				if resolvedExperimentID != "" && resolvedExperimentID != args[0] {
					return fmt.Errorf(
						"experiment id specified twice (positional %q, --experiment-id %q)",
						args[0], resolvedExperimentID,
					)
				}
				resolvedExperimentID = args[0]
			default:
				if resolvedWorkflowID != "" && resolvedWorkflowID != args[0] {
					return fmt.Errorf(
						"workflow id specified twice (positional %q, --workflow-id %q)",
						args[0], resolvedWorkflowID,
					)
				}
				resolvedWorkflowID = args[0]
			}
		}
		// Reject conflict: positional + flag form disagreeing is a
		// silent-misroute hazard (see the F-bug fix for the block create
		// route's body/path workflow_id mismatch — same shape).
		if len(args) >= 2 {
			positionalWorkflowID := args[0]
			if positionalWorkflowID != "" && flagWorkflowID != "" && positionalWorkflowID != flagWorkflowID {
				return fmt.Errorf(
					"workflow id specified twice (positional %q, --workflow-id %q)",
					positionalWorkflowID, flagWorkflowID,
				)
			}
			positionalExperimentID := args[1]
			if positionalExperimentID != "" && flagExperimentID != "" && positionalExperimentID != flagExperimentID {
				return fmt.Errorf(
					"experiment id specified twice (positional %q, --experiment-id %q)",
					positionalExperimentID, flagExperimentID,
				)
			}
			resolvedWorkflowID = positionalWorkflowID
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
		if value, _ := cmd.Flags().GetString("exclude-status"); value != "" {
			status := retab.WorkflowExperimentsExcludeStatus(value)
			params.ExcludeStatus = &status
		}
		if value, _ := cmd.Flags().GetString("trigger-type"); value != "" {
			params.TriggerType = &value
		}
		fromDate, _ := cmd.Flags().GetString("from-date")
		if fromDate != "" {
			params.FromDate = &fromDate
		}
		toDate, _ := cmd.Flags().GetString("to-date")
		if toDate != "" {
			params.ToDate = &toDate
		}
		if err := validateDateRange("from-date", "to-date", fromDate, toDate); err != nil {
			return err
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
		return printExperimentRunsListResult(cmd, result)
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
		return printExperimentResultsList(cmd, result)
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
view with ` + "`--view`" + ` to drill from headline numbers down to
individual fields. Each view requires different flags:

  summary      headline numbers           (no extra flags)
  by_document  per-document breakdown      --document-id
  by_target    per-field breakdown         --target-path
  votes        raw per-pass consensus      --document-id AND --target-path

Compare against a prior run with ` + "`--prior-run-id`" + `.`,
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
		//   by_target   — --target-path required
		//   votes       — BOTH --document-id and --target-path required
		//                 (the server rejects either one missing; check
		//                 --document-id first so the error matches the
		//                 most-likely-forgotten flag)
		switch view {
		case "by_document":
			if strings.TrimSpace(documentID) == "" {
				return fmt.Errorf("--document-id is required when --view is %s", view)
			}
		case "votes":
			if strings.TrimSpace(documentID) == "" {
				return fmt.Errorf("--document-id is required when --view is %s", view)
			}
			if strings.TrimSpace(targetPath) == "" {
				return fmt.Errorf("--target-path is required when --view is %s", view)
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
		c.Flags().StringArray("capture", nil, "capture a document from a production run: run-id[:step-id] (repeatable; inline alternative to --captures-file)")
		c.Flags().String("captures-file", "", "JSON array of {run_id, step_id} captures (or - for stdin)")
		c.Flags().String("documents-file", "", "JSON array of {handle_inputs, provenance} (or - for stdin)")
	}

	workflowsExperimentsCreateCmd.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	workflowsExperimentsCreateCmd.Flags().String("block-id", "", "block id (required unless --from-experiment)")
	workflowsExperimentsCreateCmd.Flags().String("name", "", "experiment name (required unless --from-experiment)")
	workflowsExperimentsCreateCmd.Flags().String("from-experiment", "", "clone an existing experiment by id (block, name+\"(Copy)\", consensus, and documents are inherited; takes no other flags)")
	workflowsExperimentsCreateCmd.Flags().Var(&consensusFlagValue{}, "n-consensus", "consensus count (3, 5, or 7)")
	workflowsExperimentsCreateCmd.Flags().Bool("run", false, "create a run immediately after the experiment is created")
	workflowsExperimentsCreateCmd.Flags().Bool("wait", false, "with --run, block until the run reaches a terminal status, then print the final run")
	// The --run --wait path reuses runs-create's wait machinery
	// (experimentWaitDurations), so expose the same cadence/timeout knobs here
	// — otherwise `create --run --wait --timeout-seconds N` errored "unknown flag".
	workflowsExperimentsCreateCmd.Flags().Int("poll-interval-ms", 2000, "poll cadence in milliseconds while --run --wait is set")
	workflowsExperimentsCreateCmd.Flags().Int("timeout-seconds", 600, "max seconds to wait while --run --wait is set")
	addExperimentDocFlags(workflowsExperimentsCreateCmd)
	// Keep the flag hidden but DO NOT use MarkDeprecated — cobra's auto warning
	// duplicates the more-specific message emitted by resolveWorkflowIDArg.
	_ = workflowsExperimentsCreateCmd.Flags().MarkHidden("workflow-id")
	// block-id and name are required for a fresh experiment but inherited when
	// cloning via --from-experiment, so they're enforced in RunE rather than
	// with cobra's unconditional MarkFlagRequired.

	workflowsExperimentsListCmd.Flags().String("workflow-id", "", "workflow id (alternative to the positional form)")
	workflowsExperimentsListCmd.Flags().String("before", "", "experiment id: return items before this id (mutually exclusive with --after)")
	workflowsExperimentsListCmd.Flags().String("after", "", "experiment id: return items after this id (mutually exclusive with --before)")
	workflowsExperimentsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100; default 50)")
	workflowsExperimentsListCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")

	workflowsExperimentsUpdateCmd.Flags().String("name", "", "new name")
	workflowsExperimentsUpdateCmd.Flags().Var(&consensusFlagValue{}, "n-consensus", "new consensus count (3, 5, or 7)")
	addExperimentDocFlags(workflowsExperimentsUpdateCmd)

	workflowsExperimentsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsExperimentsRunsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsExperimentsRunsListCmd.Flags().String("experiment-id", "", "filter by experiment id")
	workflowsExperimentsRunsListCmd.Flags().String("block-id", "", "filter by block id")
	workflowsExperimentsRunsListCmd.Flags().String("status", "", "filter by lifecycle status")
	workflowsExperimentsRunsListCmd.Flags().String("exclude-status", "", "exclude lifecycle status")
	workflowsExperimentsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsExperimentsRunsListCmd.Flags().Var(&rfc3339FlagValue{}, "from-date", "created on or after this date (YYYY-MM-DD or RFC3339)")
	workflowsExperimentsRunsListCmd.Flags().Var(&rfc3339FlagValue{}, "to-date", "created on or before this date (YYYY-MM-DD or RFC3339)")
	workflowsExperimentsRunsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsExperimentsRunsListCmd.Flags().String("before", "", "page before cursor (mutually exclusive with --after)")
	workflowsExperimentsRunsListCmd.Flags().String("after", "", "page after cursor (mutually exclusive with --before)")
	workflowsExperimentsRunsListCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
	workflowsExperimentsResultsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsExperimentsMetricsGetCmd.Flags().String("view", "summary", "view: summary | by_document (needs --document-id) | by_target (needs --target-path) | votes (needs both)")
	workflowsExperimentsMetricsGetCmd.Flags().String("document-id", "", "document id (required for by_document and votes views)")
	workflowsExperimentsMetricsGetCmd.Flags().String("target-path", "", "target path (required for by_target and votes views)")
	workflowsExperimentsMetricsGetCmd.Flags().String("prior-run-id", "", "prior run id")
	workflowsExperimentsMetricsGetCmd.Flags().Bool("include-prior", true, "include prior run metrics")

	// `runs create --wait` and the standalone `runs wait` share the same
	// poll/timeout knobs.
	workflowsExperimentsRunsCreateCmd.Flags().Bool("wait", false, "block until the run reaches a terminal status, then print the final run")
	workflowsExperimentsRunsCreateCmd.Flags().Int("poll-interval-ms", 2000, "poll cadence in milliseconds while --wait is set")
	workflowsExperimentsRunsCreateCmd.Flags().Int("timeout-seconds", 600, "max seconds to wait while --wait is set")
	workflowsExperimentsRunsWaitCmd.Flags().Int("poll-interval-ms", 2000, "poll cadence in milliseconds")
	workflowsExperimentsRunsWaitCmd.Flags().Int("timeout-seconds", 600, "max seconds to wait before giving up")

	workflowsExperimentsResultsCmd.AddCommand(workflowsExperimentsResultsListCmd, workflowsExperimentsResultsGetCmd)
	workflowsExperimentsMetricsCmd.AddCommand(workflowsExperimentsMetricsGetCmd)
	workflowsExperimentsRunsCmd.AddCommand(workflowsExperimentsRunsCreateCmd, workflowsExperimentsRunsListCmd, workflowsExperimentsRunsGetCmd, workflowsExperimentsRunsCancelCmd, workflowsExperimentsRunsWaitCmd)
	workflowsExperimentsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsExperimentsCmd.AddCommand(workflowsExperimentsCreateCmd, workflowsExperimentsListCmd, workflowsExperimentsGetCmd, workflowsExperimentsUpdateCmd, workflowsExperimentsDeleteCmd, workflowsExperimentsRunsCmd, workflowsExperimentsResultsCmd, workflowsExperimentsMetricsCmd)
	workflowsCmd.AddCommand(workflowsExperimentsCmd)
}
