//go:build !retab_oagen_cli_workflows_evals

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// decodeJSONInto round-trips a parsed JSON object map back into a strongly
// typed SDK struct. Used to populate the typed Target / Source / Assertion
// fields on workflow-evals create/update requests from the CLI's loose
// map-based JSON descriptors.
func decodeJSONInto(field string, raw map[string]any, out any) error {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("--%s: %w", field, err)
	}
	if err := json.Unmarshal(encoded, out); err != nil {
		return fmt.Errorf("--%s: %w", field, err)
	}
	return nil
}

// resolveWorkflowIDArg implements the positional/deprecated-flag shim used by
// the workflows subcommands that historically required `--workflow-id` and
// have since standardized on a positional first argument.
//
// Resolution order:
//
//  1. positional `args[0]` (if present) — wins outright. If a deprecated
//     `--workflow-id` flag was ALSO provided, a one-line warning is emitted
//     to stderr.
//  2. `--workflow-id <id>` — accepted with a deprecation warning to stderr.
//  3. neither — returns an error "workflow id required".
//
// The warning sink is parameterized so tests can capture it without
// hijacking os.Stderr.
func resolveWorkflowIDArg(cmd *cobra.Command, args []string) (string, error) {
	return resolveWorkflowIDArgTo(cmd, args, cmd.ErrOrStderr())
}

func resolveWorkflowIDArgTo(cmd *cobra.Command, args []string, warnTo io.Writer) (string, error) {
	flagVal, _ := cmd.Flags().GetString("workflow-id")
	flagSet := cmd.Flags().Changed("workflow-id")
	if warnTo == nil {
		warnTo = os.Stderr
	}
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		if flagSet {
			if _, err := fmt.Fprintln(warnTo, "warning: --workflow-id is deprecated; positional argument takes precedence"); err != nil {
				return "", err
			}
		}
		return args[0], nil
	}
	if flagSet && strings.TrimSpace(flagVal) != "" {
		if _, err := fmt.Fprintln(warnTo, "warning: --workflow-id is deprecated; pass the workflow id as the first positional argument"); err != nil {
			return "", err
		}
		return flagVal, nil
	}
	return "", fmt.Errorf("workflow id required")
}

// evalIDArg resolves the eval id from a command that optionally accepts a
// leading <workflow-id>. `evals get`/`update`/`delete` are eval-scoped — the
// eval id (wfnodeeval_…) is globally unique, so the workflow id is not
// required — but users coming from `evals create <workflow-id>` and
// `evals runs create <workflow-id> [eval-id]` habitually prefix the workflow
// id and used to hit a cryptic "accepts 1 arg(s), received 2". Accept it
// forgivingly: with two args the first is the (ignored) workflow id and the
// last is the eval id; with one arg the lone arg is the eval id. Pairs with
// cobra.RangeArgs(1, 2).
func evalIDArg(args []string) string {
	return strings.TrimSpace(args[len(args)-1])
}

var workflowsEvalsCmd = &cobra.Command{
	Use:   "evals",
	Short: "Manage workflow evals",
	Long: `Declarative regression evals for workflow blocks.

An eval pins a target block's expected output against a recorded input
(typically captured from a successful past run). Run the eval suite
after any change — config update, model swap, prompt edit — to
catch silent drift in extraction quality or classification accuracy.

An eval has three pieces:

  * ` + "`target`" + `   — which block is under eval
  * ` + "`source`" + `   — the input the block should consume (often a
    pointer to an existing run step)
  * ` + "`assertion`" + ` — the output handle/path and expected condition

Use ` + "`workflows evals runs create`" + ` to run a single eval or the whole
suite. Inspect runs via ` + "`workflows evals runs list`" + `.

For exploring multiple alternative block configurations side-by-side
(A/B-style), use ` + "`retab workflows experiments --help`" + ` instead.`,
	Example: `  # List a workflow's evals
  retab workflows evals list wf_abc123

  # Create an eval pinning a block's expected output
  retab workflows evals create wf_abc123 \
    --name "Invoice 17 baseline" \
    --target-file ./target.json \
    --source-file ./source.json \
    --assertion-file ./assertion.json

  # Create a run for every eval in the workflow
  retab workflows evals runs create wf_abc123`,
}

func resolveJSONMap(cmd *cobra.Command, flag string) (map[string]any, error) {
	path, _ := cmd.Flags().GetString(flag)
	if path == "" {
		return nil, nil
	}
	obj, err := readJSONMap(path)
	if err != nil {
		return nil, fmt.Errorf("--%s: %w", flag, err)
	}
	return obj, nil
}

// resolveEvalComponent picks a single eval component (target / source /
// assertion) from EITHER its `--<flag>-file` JSON file OR the inline flag
// form, enforcing that the two forms are not mixed for the same component.
// Returns nil when neither form was supplied so the caller can report which
// piece is missing.
func resolveEvalComponent(cmd *cobra.Command, fileFlag, inlineFlagDesc string, inline map[string]any, inlineSet bool) (map[string]any, error) {
	fromFile, err := resolveJSONMap(cmd, fileFlag)
	if err != nil {
		return nil, err
	}
	if fromFile != nil && inlineSet {
		return nil, fmt.Errorf("--%s and %s are mutually exclusive", fileFlag, inlineFlagDesc)
	}
	if inlineSet {
		return inline, nil
	}
	return fromFile, nil
}

// inlineEvalTarget builds the target descriptor from --block-id. Returns
// (nil, false) when --block-id was not supplied.
func inlineEvalTarget(cmd *cobra.Command) (map[string]any, bool) {
	blockID, _ := cmd.Flags().GetString("block-id")
	if strings.TrimSpace(blockID) == "" {
		return nil, false
	}
	return map[string]any{"type": "block", "block_id": blockID}, true
}

// inlineEvalSource builds a run_step source from --run-id (and optional
// --step-id). Returns (nil, false) when --run-id was not supplied.
func inlineEvalSource(cmd *cobra.Command) (map[string]any, bool) {
	runID, _ := cmd.Flags().GetString("run-id")
	if strings.TrimSpace(runID) == "" {
		return nil, false
	}
	source := map[string]any{"type": "run_step", "run_id": runID}
	if stepID, _ := cmd.Flags().GetString("step-id"); strings.TrimSpace(stepID) != "" {
		source["step_id"] = stepID
	}
	return source, true
}

// inlineEvalAssertion builds an `equals` assertion from --output-handle-id /
// --path / --equals. It is considered "set" when any of those flags is
// provided; --equals is then mandatory (the assertion needs an expected
// value), and --output-handle-id defaults to the conventional single-JSON
// output handle. The expected value is parsed as a JSON literal when possible
// (so `--equals 300000` is the number 300000, `--equals true` the boolean)
// and falls back to the raw string otherwise.
func inlineEvalAssertion(cmd *cobra.Command) (map[string]any, bool, error) {
	handleSet := cmd.Flags().Changed("output-handle-id")
	pathSet := cmd.Flags().Changed("path")
	equalsSet := cmd.Flags().Changed("equals")
	if !handleSet && !pathSet && !equalsSet {
		return nil, false, nil
	}
	if !equalsSet {
		return nil, true, fmt.Errorf("inline assertion requires --equals (the expected value)")
	}
	handle, _ := cmd.Flags().GetString("output-handle-id")
	if strings.TrimSpace(handle) == "" {
		handle = "output-json-0"
	}
	target := map[string]any{"output_handle_id": handle}
	if pathSet {
		path, _ := cmd.Flags().GetString("path")
		target["path"] = path
	}
	raw, _ := cmd.Flags().GetString("equals")
	return map[string]any{
		"target":    target,
		"condition": map[string]any{"kind": "equals", "expected": parseInlineExpected(raw)},
	}, true, nil
}

// assertionSpecToMap round-trips a typed AssertionSpec into a generic map so the
// update path can merge inline overrides onto the existing assertion. A nil spec
// yields an empty map.
func assertionSpecToMap(spec *retab.AssertionSpec) (map[string]any, error) {
	if spec == nil {
		return map[string]any{}, nil
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// applyInlineAssertionOverrides applies ONLY the explicitly-set inline assertion
// flags onto an existing assertion map, preserving every field the user did not
// touch — notably target.path. `evals update --equals X` is a partial re-pin of
// the expected value, not a full assertion replacement; rebuilding the assertion
// from inline flags alone silently dropped target.path, turning a field-scoped
// assertion into a whole-handle one.
func applyInlineAssertionOverrides(cmd *cobra.Command, base map[string]any) map[string]any {
	merged := base
	if merged == nil {
		merged = map[string]any{}
	}
	target, _ := merged["target"].(map[string]any)
	if target == nil {
		target = map[string]any{}
	}
	if cmd.Flags().Changed("output-handle-id") {
		handle, _ := cmd.Flags().GetString("output-handle-id")
		if strings.TrimSpace(handle) == "" {
			handle = "output-json-0"
		}
		target["output_handle_id"] = handle
	}
	if _, ok := target["output_handle_id"]; !ok {
		target["output_handle_id"] = "output-json-0"
	}
	if cmd.Flags().Changed("path") {
		path, _ := cmd.Flags().GetString("path")
		target["path"] = path
	}
	merged["target"] = target
	if cmd.Flags().Changed("equals") {
		raw, _ := cmd.Flags().GetString("equals")
		merged["condition"] = map[string]any{"kind": "equals", "expected": parseInlineExpected(raw)}
	}
	return merged
}

// parseInlineExpected interprets the --equals value as a JSON literal
// (number, boolean, null, quoted string, array, object) when it parses, and
// otherwise treats it as a bare string. This keeps the common
// `--equals 300000` numeric while still allowing `--equals "in review"`.
func parseInlineExpected(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}
	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		return parsed
	}
	return raw
}

// validateWorkflowEvalName trims surrounding whitespace and rejects names
// that are empty after trimming. Returns the cleaned value so callers can
// propagate the trimmed string into the outgoing request body — without
// this transformation `workflows evals create --name "  padded  "` would
// silently persist with the padding (bug #3). Sibling validators
// (validateWorkflowName, validateExperimentName) share the same shape.
func validateWorkflowEvalName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("eval name must not be blank")
	}
	return trimmed, nil
}

func validateWorkflowEvalAssertion(assertion map[string]any) error {
	target, ok := assertion["target"].(map[string]any)
	if !ok || target == nil {
		return fmt.Errorf("assertion.target is required")
	}
	if outputHandleID, ok := target["output_handle_id"].(string); !ok || outputHandleID == "" {
		return fmt.Errorf("assertion.target.output_handle_id is required")
	}
	if _, ok := assertion["condition"].(map[string]any); !ok {
		return fmt.Errorf("assertion.condition is required")
	}
	return nil
}

func validateWorkflowEvalTarget(target map[string]any) error {
	targetType, _ := target["type"].(string)
	if targetType != "" && targetType != "block" {
		return fmt.Errorf("target.type must be block")
	}
	blockID, _ := target["block_id"].(string)
	if strings.TrimSpace(blockID) == "" {
		return fmt.Errorf("target.block_id is required")
	}
	return nil
}

func validateWorkflowEvalSource(source map[string]any) error {
	sourceType, _ := source["type"].(string)
	switch sourceType {
	case "manual":
		if handleInputs, ok := source["handle_inputs"]; ok {
			if _, ok := handleInputs.(map[string]any); !ok {
				return fmt.Errorf("source.handle_inputs must be an object")
			}
		}
		return nil
	case "run_step":
		runID, _ := source["run_id"].(string)
		if strings.TrimSpace(runID) == "" {
			return fmt.Errorf("source.run_id is required")
		}
		if stepID, ok := source["step_id"]; ok && stepID != nil {
			if _, ok := stepID.(string); !ok {
				return fmt.Errorf("source.step_id must be a string")
			}
		}
		return nil
	default:
		return fmt.Errorf("source.type must be manual or run_step")
	}
}

type consensusFlagValue struct{ value string }

func (v *consensusFlagValue) String() string {
	if v.value == "" {
		return "0"
	}
	return v.value
}

func (v *consensusFlagValue) Type() string { return "int" }

func (v *consensusFlagValue) Set(raw string) error {
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}
	switch parsed {
	case 3, 5, 7:
		v.value = raw
		return nil
	default:
		return fmt.Errorf("must be 3, 5, or 7")
	}
}

var workflowsEvalsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> [flags]",
	Short: "Create a workflow eval",
	Long: `Create a regression eval pinning a block's expected output.
You'll typically capture the three files from a successful past run:

  ` + "`--target-file`" + `     — which block to eval (e.g.
  ` + `{"type": "block", "block_id": "block_extract_1"}` + `).

  ` + "`--source-file`" + `     — the input the target block will see.
  Two accepted shapes (discriminated by ` + "`type`" + `):

    1. ` + `{"type": "manual", "handle_inputs": {...}}` + ` — a
       fully-specified input payload, keyed by handle id. Use this
       to feed literal values that don't come from a previous run.
       ` + "`handle_inputs`" + ` defaults to ` + `{}` + ` when omitted.

    2. ` + `{"type": "run_step", "run_id": "run_xxx"}` + ` — a
       pointer to a past run. ` + "`run_id`" + ` is required; the
       server resolves which step within that run by matching
       ` + "`target.block_id`" + ` from your ` + "`--target-file`" + `.
       An optional ` + "`step_id`" + ` may be passed to pin to an
       exact step. Both source models use ` + "`extra=\"forbid\"`" + `,
       so any other top-level field is rejected with
       ` + `"Extra inputs are not permitted"` + `.

  ` + "`--assertion-file`" + `  — the output handle/path and condition
  (e.g. ` + `{"target":{"output_handle_id":"output-json-0","path":"total"},"condition":{"kind":"equals","expected":120}}` + `).

For the common "assert block X's field Y equals Z from run R" case you
can skip the three files and pass the pieces inline instead:

  ` + "`--block-id`" + `          — target block (same as ` + "`--target-file`" + `).
  ` + "`--run-id`" + ` / ` + "`--step-id`" + ` — a ` + "`run_step`" + ` source (` + "`--step-id`" + ` optional).
  ` + "`--path`" + ` / ` + "`--equals`" + `    — the field path and expected value for an
  ` + "`equals`" + ` assertion. ` + "`--equals`" + ` is parsed as a JSON literal when it
  can be (so ` + "`--equals 300000`" + ` is numeric), else as a string.
  ` + "`--output-handle-id`" + `   — defaults to ` + "`output-json-0`" + `; override for blocks
  with a differently-named output handle.

The file and inline forms are alternatives — passing both for the same
component (e.g. ` + "`--target-file`" + ` and ` + "`--block-id`" + `) is an error.

After creation, run with ` + "`workflows evals runs create`" + `.`,
	Example: `  # Create an eval from JSON files
  retab workflows evals create wf_abc123 \
    --name "Invoice 17 baseline" \
    --target-file ./target.json \
    --source-file ./source.json \
    --assertion-file ./assertion.json

  # Same eval, inline — no JSON files needed
  retab workflows evals create wf_abc123 \
    --name "Invoice 17 baseline" \
    --block-id extract_1 \
    --run-id run_xxx \
    --path net_amount_payable_usd \
    --equals 300000`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, err := resolveWorkflowIDArg(cmd, args)
		if err != nil {
			return err
		}
		req := retab.WorkflowEvalsCreateParams{}
		req.WorkflowID = workflowID
		rawName, _ := cmd.Flags().GetString("name")
		trimmedName, err := validateWorkflowEvalName(rawName)
		if err != nil {
			return err
		}
		req.Name = &trimmedName
		// Each component may be supplied as a JSON file OR via the inline
		// flag form — but not both for the same component.
		inlineTarget, inlineTargetSet := inlineEvalTarget(cmd)
		inlineSource, inlineSourceSet := inlineEvalSource(cmd)
		inlineAssertion, inlineAssertionSet, err := inlineEvalAssertion(cmd)
		if err != nil {
			return err
		}
		target, err := resolveEvalComponent(cmd, "target-file", "--block-id", inlineTarget, inlineTargetSet)
		if err != nil {
			return err
		}
		source, err := resolveEvalComponent(cmd, "source-file", "--run-id/--step-id", inlineSource, inlineSourceSet)
		if err != nil {
			return err
		}
		assertion, err := resolveEvalComponent(cmd, "assertion-file", "--output-handle-id/--path/--equals", inlineAssertion, inlineAssertionSet)
		if err != nil {
			return err
		}
		if target == nil || source == nil || assertion == nil {
			return fmt.Errorf("each of target, source, and assertion is required — supply it as a file " +
				"(--target-file / --source-file / --assertion-file) or inline " +
				"(--block-id; --run-id [--step-id]; --path --equals)")
		}
		if err := validateWorkflowEvalTarget(target); err != nil {
			return fmt.Errorf("--target-file: %w", err)
		}
		if err := validateWorkflowEvalSource(source); err != nil {
			return fmt.Errorf("--source-file: %w", err)
		}
		if err := validateWorkflowEvalAssertion(assertion); err != nil {
			return fmt.Errorf("--assertion-file: %w", err)
		}
		if err := decodeJSONInto("target-file", target, &req.Target); err != nil {
			return err
		}
		if err := decodeJSONInto("source-file", source, &req.Source); err != nil {
			return err
		}
		if err := decodeJSONInto("assertion-file", assertion, &req.Assertion); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Create(ctx, &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEvalsGetCmd = &cobra.Command{
	Use:   "get [workflow-id] <eval-id>",
	Short: "Get an eval",
	Long: `Fetch an eval's definition: target block, source input, assertion
output, name, timestamps. The eval id is sufficient; a leading workflow id is
accepted (and ignored) so the call mirrors 'evals create <workflow-id>'.`,
	Example: `  # Inspect an eval
  retab workflows evals get wfnodeeval_jkl012`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Get(ctx, evalIDArg(args))
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEvalsListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List evals",
	Long: `List every eval attached to a workflow. Filter by
` + "`--target-block-id`" + ` to focus on the regression suite for a
particular block.

Name the workflow either positionally (` + "`list <workflow-id>`" + `) or with
the ` + "`--workflow-id`" + ` flag — the two forms are equivalent. Passing both
is accepted when they agree; an error is raised only when they disagree. The
workflow id is required: evals have no org-wide listing.`,
	Example: `  # All evals in a workflow (positional)
  retab workflows evals list wf_abc123

  # Same, with the flag form
  retab workflows evals list --workflow-id wf_abc123

  # Just the evals guarding one block
  retab workflows evals list wf_abc123 --target-block-id block_extract_1`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Workflow id positionally OR via --workflow-id (co-equal forms);
		// required here — evals have no org-wide listing.
		workflowID, err := resolveWorkflowScope(cmd, args, true)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.WorkflowEvalsListParams{WorkflowID: workflowID}
		if v, _ := cmd.Flags().GetString("target-block-id"); v != "" {
			req.TargetBlockID = ptr(v)
		}
		if v := getIntFlagOrDefault(cmd, "limit", 50); v > 0 {
			req.Limit = ptr(v)
		}
		result, err := client.Workflows.Evals.List(ctx, &req)
		if err != nil {
			return err
		}
		return printWorkflowEvalsListResult(cmd, result)
	}),
}

func printWorkflowEvalsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowEval]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputTable) {
			return RenderList(os.Stdout, OutputTable, result, workflowEvalColumns)
		}
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputCSV) {
			return RenderList(os.Stdout, OutputCSV, result, workflowEvalColumns)
		}
	}
	return printJSON(result)
}

var workflowEvalColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowEvalCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return workflowEvalCell(row, "name") }},
	{Header: "TARGET_BLOCK_ID", Extract: func(row any) string { return workflowEvalCell(row, "target.block_id") }},
	{Header: "FRESHNESS", Extract: artifactFreshnessCell},
	{Header: "SCHEMA_DRIFT", Extract: func(row any) string { return workflowEvalCell(row, "schema_drift") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowEvalCell(row, "created_at") }},
}

func workflowEvalCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

// workflowEvalRunColumns is the dedicated TableColumn spec for
// `workflows evals runs list --output table/csv`. The generic auto-renderer only
// surfaced ID + a TYPE column that confusingly showed the lifecycle status; these
// columns show what a user reads off an eval-suite run: status and the pass/fail
// tally.
var workflowEvalRunColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowEvalCell(row, "id") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowEvalCell(row, "lifecycle.status") }},
	{Header: "TESTS", Extract: func(row any) string { return workflowEvalCell(row, "total_evals") }},
	{Header: "PASSED", Extract: func(row any) string { return workflowEvalCell(row, "counts.outcome.passed") }},
	{Header: "FAILED", Extract: func(row any) string { return workflowEvalCell(row, "counts.outcome.failed") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowEvalCell(row, "timing.created_at") }},
}

func printWorkflowEvalsRunsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowEvalRun]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			switch f.Value.String() {
			case string(OutputTable), string(OutputCSV):
				return RenderList(os.Stdout, OutputFormat(f.Value.String()), result, workflowEvalRunColumns)
			}
		}
	}
	return printJSON(result)
}

// workflowEvalResultColumns is the dedicated TableColumn spec for
// `workflows evals results list --output table/csv`. Eval-result rows are flat
// (block_id/block_type/eval_id), not nested target/source descriptors, so these
// columns mirror the public result payload instead of rendering empty cells.
var workflowEvalResultColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowEvalCell(row, "id") }},
	{Header: "VERDICT", Extract: func(row any) string { return workflowEvalCell(row, "verdict") }},
	{Header: "TARGET", Extract: func(row any) string { return workflowEvalCell(row, "block_id") }},
	{Header: "BLOCK_KIND", Extract: func(row any) string { return workflowEvalCell(row, "block_type") }},
	{Header: "TEST", Extract: func(row any) string { return workflowEvalCell(row, "eval_id") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowEvalCell(row, "lifecycle.status") }},
}

func printWorkflowEvalResultsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowEvalResult]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			switch f.Value.String() {
			case string(OutputTable), string(OutputCSV):
				return RenderList(os.Stdout, OutputFormat(f.Value.String()), result, workflowEvalResultColumns)
			}
		}
	}
	return printJSON(result)
}

var workflowsEvalsUpdateCmd = &cobra.Command{
	Use:   "update [workflow-id] <eval-id>",
	Short: "Update an eval",
	Long: `Re-pin an eval's expected output or source input. Use this when
a deliberate schema or prompt change makes the old assertion stale — the
intent is to ratify the new output as the new baseline, not to silence
flaky runs.`,
	Example: `  # Refresh the assertion after a deliberate schema change
  retab workflows evals update wfnodeeval_jkl012 \
    --assertion-file ./new-assertion.json

  # Same, inline — no JSON file needed (mirrors 'evals create')
  retab workflows evals update wfnodeeval_jkl012 \
    --path net_amount_payable_usd --equals 300000

  # Re-pin the source run inline
  retab workflows evals update wfnodeeval_jkl012 --run-id run_xyz789

  # Rename an eval
  retab workflows evals update wfnodeeval_jkl012 \
    --name "Invoice 17 baseline (v2 schema)"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		evalID := evalIDArg(args)
		// The assertion and source may each be supplied as a JSON file OR via the
		// same inline flag form `evals create` accepts (mutually exclusive per
		// component, enforced by resolveEvalComponent).
		inlineSource, inlineSourceSet := inlineEvalSource(cmd)
		// Detect inline assertion flags and enforce --equals (the expected value)
		// without building the full assertion here — the update path merges inline
		// overrides onto the existing assertion below, so we must not synthesize a
		// fresh assertion that drops untouched fields.
		_, inlineAssertionSet, err := inlineEvalAssertion(cmd)
		if err != nil {
			return err
		}
		// File presence is value-based (an empty path means "not supplied"),
		// matching resolveJSONMap — Changed() would spuriously fire on a flag a
		// prior call set back to "".
		assertionFilePath, _ := cmd.Flags().GetString("assertion-file")
		fileAssertionSet := strings.TrimSpace(assertionFilePath) != ""
		if fileAssertionSet && inlineAssertionSet {
			return fmt.Errorf("--assertion-file and --output-handle-id/--path/--equals are mutually exclusive")
		}
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") &&
			!fileAssertionSet && !inlineAssertionSet &&
			!cmd.Flags().Changed("source-file") && !inlineSourceSet {
			return fmt.Errorf("nothing to update: pass at least one of --name, the assertion " +
				"(--assertion-file or --output-handle-id/--path/--equals), or the source " +
				"(--source-file or --run-id/--step-id)")
		}
		req := retab.WorkflowEvalsUpdateParams{}
		if cmd.Flags().Changed("name") {
			rawName, _ := cmd.Flags().GetString("name")
			trimmed, err := validateWorkflowEvalName(rawName)
			if err != nil {
				return err
			}
			req.Name = &trimmed
		}
		source, err := resolveEvalComponent(cmd, "source-file", "--run-id/--step-id", inlineSource, inlineSourceSet)
		if err != nil {
			return err
		}
		// Read any local files (source-file above, assertion-file here) BEFORE
		// touching credentials so a bad path fails with a file error, not a
		// credentials error.
		var fileAssertion map[string]any
		if fileAssertionSet {
			fileAssertion, err = resolveJSONMap(cmd, "assertion-file")
			if err != nil {
				return err
			}
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		// Resolve the assertion: a --assertion-file fully replaces it, while inline
		// flags merge onto the EXISTING assertion (fetched first) so a --equals-only
		// re-pin preserves target.path / target.output_handle_id.
		var assertion map[string]any
		switch {
		case fileAssertionSet:
			assertion = fileAssertion
		case inlineAssertionSet:
			existing, gerr := client.Workflows.Evals.Get(ctx, evalID)
			if gerr != nil {
				return gerr
			}
			base, merr := assertionSpecToMap(existing.Assertion)
			if merr != nil {
				return merr
			}
			assertion = applyInlineAssertionOverrides(cmd, base)
		}
		if assertion != nil {
			if err := validateWorkflowEvalAssertion(assertion); err != nil {
				return fmt.Errorf("assertion: %w", err)
			}
			req.Assertion = &retab.AssertionSpec{}
			if err := decodeJSONInto("assertion", assertion, req.Assertion); err != nil {
				return err
			}
		}
		if source != nil {
			if err := validateWorkflowEvalSource(source); err != nil {
				return fmt.Errorf("--source-file: %w", err)
			}
			req.Source = &retab.WorkflowEvalSource{}
			if err := decodeJSONInto("source", source, req.Source); err != nil {
				return err
			}
		}
		result, err := client.Workflows.Evals.Update(ctx, evalID, &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEvalsDeleteCmd = &cobra.Command{
	Use:   "delete [workflow-id] <eval-id>",
	Short: "Delete an eval",
	Long: `Permanently delete a regression eval and its run history.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal. Run history is removed alongside the eval definition. The
eval id is sufficient; a leading workflow id is accepted (and ignored).`,
	Example: `  # Drop a stale eval (interactive, asks to confirm)
  retab workflows evals delete wfnodeeval_jkl012

  # Skip the prompt in scripts
  retab workflows evals delete wfnodeeval_jkl012 --yes`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		evalID := evalIDArg(args)
		if err := confirmDestructive(cmd, "eval", evalID); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Evals.Delete(ctx, evalID); err != nil {
			return err
		}
		confirmDeleted("eval", evalID)
		return nil
	}),
}

// ---- eval runs subgroup ----
//
// SDK-backed canonical routes: "/v1/workflows/evals/runs" and
// "/v1/workflows/evals/results".

var workflowsEvalsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Create and inspect eval runs",
	Long: `Create workflow-eval runs, poll their status, and inspect child
result records.`,
	Example: `  # Create a run for the whole workflow
  retab workflows evals runs create wf_abc123

  # Poll a workflow-eval run
  retab workflows evals runs get wfevalrun_mno345

  # Fetch child result rows for a workflow-eval run
  retab workflows evals results list wfevalrun_mno345

  # Recent runs for one workflow (narrow to an eval with --eval-id)
  retab workflows evals runs list wf_abc123 --eval-id wfnodeeval_jkl012`,
}

var workflowsEvalsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> [eval-id] [flags]",
	Short: "Create a workflow-eval run",
	Long: `Start a workflow-eval run. The first positional argument is the
` + "`workflow-id`" + ` — by default the run executes every eval attached to
the workflow. Pass a second positional ` + "`eval-id`" + ` (or ` + "`--eval-id`" + `)
to run just that one eval.

To run a single eval pass ` + "`--eval-id`" + `; to run every saved eval
for one block pass ` + "`--target-file`" + ` with a JSON target such as
` + `{"type":"block","block_id":"extract_1"}` + `.

Pass ` + "`--wait`" + ` to block until the run reaches a terminal status
(` + "`completed`" + `/` + "`error`" + `/` + "`cancelled`" + `) and print the
final run — saving you from scripting a poll loop around
` + "`runs get`" + `. With ` + "`--wait`" + ` the command exits non-zero if the
run ends in ` + "`error`" + `/` + "`cancelled`" + ` or the timeout elapses.
Without it, poll progress with ` + "`workflows evals runs get`" + ` and fetch
per-eval results with ` + "`workflows evals results list`" + `.`,
	Example: `  # Run every eval in the workflow
  retab workflows evals runs create wf_abc123

  # Run a single eval
  retab workflows evals runs create wf_abc123 \
    --eval-id wfnodeeval_jkl012

  # Run every eval for one block
  retab workflows evals runs create wf_abc123 \
    --target-file ./target.json

  # Create and block until the suite finishes (2s polls, 10m timeout)
  retab workflows evals runs create wf_abc123 --wait

  # Tune the polling cadence and ceiling
  retab workflows evals runs create wf_abc123 \
    --wait --poll-interval-ms 1000 --timeout-seconds 1800`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := &retab.WorkflowEvalRunsCreateParams{
			WorkflowID: args[0],
		}
		evalID, _ := cmd.Flags().GetString("eval-id")
		// Tolerate `evals runs create <workflow-id> <eval-id>`: a second
		// positional is a convenience alias for --eval-id (run that one eval),
		// matching how users carry over the `blocks/edges <wf-id> <child-id>`
		// shape. The flag and the positional are mutually exclusive.
		if len(args) == 2 {
			if evalID != "" {
				return fmt.Errorf("pass the eval id once: either the second positional argument or --eval-id, not both")
			}
			evalID = args[1]
		}
		target, err := resolveJSONMap(cmd, "target-file")
		if err != nil {
			return err
		}
		if evalID != "" && target != nil {
			return fmt.Errorf("--eval-id and --target-file are mutually exclusive")
		}
		if evalID != "" {
			params.Scope = &retab.WorkflowEvalRunScope{
				Type:   retab.WorkflowEvalRunScopeTypeSingle,
				EvalID: ptr(evalID),
			}
		}
		if target != nil {
			if err := validateWorkflowEvalTarget(target); err != nil {
				return fmt.Errorf("--target-file: %w", err)
			}
			blockID, _ := target["block_id"].(string)
			params.Scope = &retab.WorkflowEvalRunScope{
				Type:    retab.WorkflowEvalRunScopeTypeBlock,
				BlockID: ptr(blockID),
			}
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Runs.Create(ctx, params)
		if err != nil {
			return err
		}
		if wait, _ := cmd.Flags().GetBool("wait"); !wait {
			return printResult(cmd, result)
		}
		// --wait: poll the freshly-created run until it settles, matching the
		// contract already on `experiments runs create --wait`. Removes the
		// hand-rolled `until` loop around `runs get` that the help previously
		// pointed callers to.
		pollInterval, timeout := experimentWaitDurations(cmd)
		final, waitErr := waitForWorkflowEvalRun(ctx, client, result.ID, pollInterval, timeout)
		if final != nil {
			if err := printResult(cmd, final); err != nil {
				return err
			}
		}
		if waitErr != nil {
			return waitErr
		}
		return workflowEvalRunTerminalError(final)
	}),
}

// workflowEvalRunStatus reads the lifecycle discriminator off a workflow-eval
// run. The lifecycle is a discriminated-union envelope, so the status string
// comes from its `Status()` getter rather than a plain field.
func workflowEvalRunStatus(run *retab.WorkflowEvalRun) string {
	if run == nil {
		return ""
	}
	return run.Lifecycle.Status()
}

// isTerminalWorkflowEvalRun reports whether the run has settled. A
// workflow-eval run can't pause for review, so the terminal set is just
// completed / error / cancelled (pending/queued/running are still in flight).
func isTerminalWorkflowEvalRun(run *retab.WorkflowEvalRun) bool {
	switch workflowEvalRunStatus(run) {
	case "completed", "error", "cancelled":
		return true
	default:
		return false
	}
}

// workflowEvalRunTerminalError maps a settled-but-unsuccessful run to a
// non-zero exit. completed (or an unknown/empty status) is success;
// error/cancelled is failure.
func workflowEvalRunTerminalError(run *retab.WorkflowEvalRun) error {
	if run == nil {
		return nil
	}
	switch status := workflowEvalRunStatus(run); status {
	case "", "completed":
		return nil
	case "error", "cancelled":
		if run.ID == "" {
			return fmt.Errorf("workflow-eval run ended with status %s", status)
		}
		return fmt.Errorf("workflow-eval run %s ended with status %s", run.ID, status)
	default:
		return nil
	}
}

// waitForWorkflowEvalRun polls Runs.Get until the run reaches a terminal
// status or the timeout elapses. On timeout it returns the last-observed
// (non-final) run alongside the error so callers can still surface partial
// state — mirroring waitForExperimentRun.
func waitForWorkflowEvalRun(
	ctx context.Context,
	client *retab.Client,
	runID string,
	pollInterval time.Duration,
	timeout time.Duration,
) (*retab.WorkflowEvalRun, error) {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		run, err := client.Workflows.Evals.Runs.Get(ctx, runID)
		if err != nil {
			return nil, err
		}
		if isTerminalWorkflowEvalRun(run) {
			return run, nil
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return run, fmt.Errorf("timed out waiting for workflow-eval run %s: %w", runID, ctx.Err())
		case <-timer.C:
		}
	}
}

// workflowsEvalsRunsWaitCmd is the standalone poller for an already-created
// workflow-eval run, mirroring `experiments runs wait` so the
// `create --wait` / standalone-`wait` pair is consistent across the run
// families.
var workflowsEvalsRunsWaitCmd = &cobra.Command{
	Use:   "wait <run-id>",
	Short: "Poll until a workflow-eval run reaches a terminal status",
	Long: `Block until a workflow-eval run hits a terminal status
(` + "`completed`" + `, ` + "`error`" + `, or ` + "`cancelled`" + `),
polling on a configurable interval. Defaults: 2-second polls, 10-minute
timeout.

Cleaner than scripting a poll loop around ` + "`runs get`" + ` — the CLI
handles the interval and timeout, and exits non-zero if the run ends in
` + "`error`" + `/` + "`cancelled`" + ` or the timeout elapses. Pair with
` + "`runs create --wait`" + ` to create and block in a single step.`,
	Example: `  # Wait with defaults (2s polls, 600s timeout)
  retab workflows evals runs wait wfevalrun_mno345

  # Faster polls, longer ceiling
  retab workflows evals runs wait wfevalrun_mno345 \
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
		final, waitErr := waitForWorkflowEvalRun(ctx, client, args[0], pollInterval, timeout)
		if final != nil {
			if err := printResult(cmd, final); err != nil {
				return err
			}
		}
		if waitErr != nil {
			return waitErr
		}
		return workflowEvalRunTerminalError(final)
	}),
}

var workflowsEvalsRunsListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List workflow-eval runs",
	Long: `List workflow-eval runs. Filter by workflow, eval, target block,
status, trigger, date, or cursor.

Name the workflow either positionally (` + "`list <workflow-id>`" + `) or with
the ` + "`--workflow-id`" + ` flag — the two forms are equivalent. Passing both
is accepted when they agree; an error is raised only when they disagree. The
workflow id is optional here: with neither form, runs are listed
workspace-wide.`,
	Example: `  # Recent runs for one workflow (positional)
  retab workflows evals runs list wf_abc123 --limit 50

  # Narrow to a single eval (flag form still accepted)
  retab workflows evals runs list wf_abc123 --eval-id wfnodeeval_jkl012 --limit 50`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateWorkflowEvalsRunsListFilters(cmd); err != nil {
			return err
		}
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		// Workflow id positionally OR via --workflow-id (co-equal forms);
		// optional here — eval runs have a workspace-wide listing.
		workflowID, err := resolveWorkflowScope(cmd, args, false)
		if err != nil {
			return err
		}
		params := &retab.WorkflowEvalRunsListParams{
			PaginationParams: retab.PaginationParams{
				Limit: ptr(getIntFlagOrDefault(cmd, "limit", 20)),
			},
		}
		if workflowID != "" {
			params.WorkflowID = ptr(workflowID)
		}
		if v, _ := cmd.Flags().GetString("eval-id"); v != "" {
			params.EvalID = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("target-block-id"); v != "" {
			params.TargetBlockID = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			params.Status = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("exclude-status"); v != "" {
			params.ExcludeStatus = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("trigger-type"); v != "" {
			params.TriggerType = ptr(v)
		}
		dateQuery := url.Values{}
		if v, _ := cmd.Flags().GetString("from-date"); v != "" {
			parsed, formatted, err := workflowEvalRunDateQueryValue(v, false)
			if err != nil {
				return err
			}
			params.FromDate = &parsed
			dateQuery.Set("from_date", formatted)
		}
		if v, _ := cmd.Flags().GetString("to-date"); v != "" {
			parsed, formatted, err := workflowEvalRunDateQueryValue(v, true)
			if err != nil {
				return err
			}
			params.ToDate = &parsed
			dateQuery.Set("to_date", formatted)
		}
		if v, _ := cmd.Flags().GetString("sort-by"); v != "" {
			params.SortBy = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("before"); v != "" {
			params.Before = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("after"); v != "" {
			params.After = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("order"); v != "" {
			params.Order = ptr(v)
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Runs.List(ctx, params, retab.WithRequestParams(dateQuery))
		if err != nil {
			return err
		}
		return printWorkflowEvalsRunsListResult(cmd, result)
	}),
}

// workflowEvalRunStatusValues mirrors `WorkflowEvalPublicRunStatus` in
// `backend/.../evals/models.py`. A workflow-eval run can't pause for
// review, so `awaiting_review` is intentionally absent.
var allowedWorkflowEvalRunStatuses = map[string]bool{
	"pending":   true,
	"running":   true,
	"completed": true,
	"error":     true,
	"cancelled": true,
}

const workflowEvalRunStatusValues = "pending, running, completed, error, cancelled"

func validateWorkflowEvalsRunsListFilters(cmd *cobra.Command) error {
	if err := validateEnumFlag(cmd, "status", allowedWorkflowEvalRunStatuses, workflowEvalRunStatusValues); err != nil {
		return err
	}
	if err := validateEnumFlag(cmd, "exclude-status", allowedWorkflowEvalRunStatuses, workflowEvalRunStatusValues); err != nil {
		return err
	}
	if err := validateOrderFlag(cmd, "order"); err != nil {
		return err
	}
	if err := validateDateFlag(cmd, "from-date"); err != nil {
		return err
	}
	return validateDateFlag(cmd, "to-date")
}

func workflowEvalRunDateQueryValue(raw string, endOfDay bool) (time.Time, string, error) {
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, "", err
	}
	parsed = parsed.UTC()
	if endOfDay {
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, time.UTC)
	}
	return parsed, parsed.Format(time.RFC3339), nil
}

var workflowsEvalsRunsGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get a workflow-eval run",
	Long:  `Fetch a workflow-eval run by run id.`,
	Example: `  # Poll a workflow-eval run
  retab workflows evals runs get wfevalrun_mno345`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Runs.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEvalsRunsCancelCmd = &cobra.Command{
	Use:   "cancel <run-id>",
	Short: "Cancel a workflow-eval run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Runs.Cancel(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEvalsResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Inspect workflow-eval run results",
}

var workflowsEvalsResultsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List child results for a workflow-eval run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := &retab.WorkflowEvalRunResultsListParams{
			RunID: args[0],
			PaginationParams: retab.PaginationParams{
				Limit: ptr(getIntFlagOrDefault(cmd, "limit", 20)),
			},
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Results.List(ctx, params)
		if err != nil {
			return err
		}
		return printWorkflowEvalResultsListResult(cmd, result)
	}),
}

var workflowsEvalsResultsGetCmd = &cobra.Command{
	Use:   "get <result-id>",
	Short: "Get a workflow-eval result",
	Long:  `Fetch one workflow-eval result by flat result id.`,
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Evals.Results.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsEvalsCreateCmd.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	workflowsEvalsCreateCmd.Flags().String("name", "", "eval name")
	workflowsEvalsCreateCmd.Flags().String("target-file", "", "JSON file with the target object (or - for stdin). Alternative to --block-id.")
	workflowsEvalsCreateCmd.Flags().String("source-file", "", "JSON file with the source object (or - for stdin). Two accepted shapes: {\"type\":\"manual\",\"handle_inputs\":{...}} or {\"type\":\"run_step\",\"run_id\":\"run_xxx\",\"step_id\":\"...\"} (step_id optional). Alternative to --run-id. See 'evals create --help' for the full schema.")
	workflowsEvalsCreateCmd.Flags().String("assertion-file", "", "JSON file with the assertion object (or - for stdin). Alternative to --path/--equals.")
	// Inline alternative to the three JSON files for the common
	// "assert block X's field Y equals Z from run R" case.
	workflowsEvalsCreateCmd.Flags().String("block-id", "", "target block id (inline alternative to --target-file)")
	workflowsEvalsCreateCmd.Flags().String("run-id", "", "source run id for a run_step source (inline alternative to --source-file)")
	workflowsEvalsCreateCmd.Flags().String("step-id", "", "source step id within --run-id (optional; pairs with --run-id)")
	workflowsEvalsCreateCmd.Flags().String("output-handle-id", "", "assertion output handle id (inline; defaults to output-json-0)")
	workflowsEvalsCreateCmd.Flags().String("path", "", "assertion field path inside the output handle (inline; pairs with --equals)")
	workflowsEvalsCreateCmd.Flags().String("equals", "", "expected value for an inline equals assertion (JSON literal or string; alternative to --assertion-file)")
	// Keep the flag hidden but DO NOT use MarkDeprecated — cobra's auto warning
	// duplicates the more-specific message emitted by resolveWorkflowIDArg.
	_ = workflowsEvalsCreateCmd.Flags().MarkHidden("workflow-id")

	workflowsEvalsListCmd.Flags().String("workflow-id", "", "workflow id (alternative to the positional form)")
	workflowsEvalsListCmd.Flags().String("target-block-id", "", "filter by target block id")
	workflowsEvalsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 50)")

	workflowsEvalsUpdateCmd.Flags().String("name", "", "new eval name")
	workflowsEvalsUpdateCmd.Flags().String("assertion-file", "", "JSON file with new assertion (or - for stdin). Alternative to --output-handle-id/--path/--equals.")
	workflowsEvalsUpdateCmd.Flags().String("source-file", "", "JSON file with new source (or - for stdin). Alternative to --run-id/--step-id.")
	// Inline assertion/source flags, mirroring `evals create`, so an eval's
	// assertion or source can be re-pinned without hand-authoring JSON.
	workflowsEvalsUpdateCmd.Flags().String("output-handle-id", "", "new assertion output handle id (inline; defaults to output-json-0)")
	workflowsEvalsUpdateCmd.Flags().String("path", "", "new assertion field path inside the output handle (inline; pairs with --equals)")
	workflowsEvalsUpdateCmd.Flags().String("equals", "", "new expected value for an inline equals assertion (JSON literal or string; alternative to --assertion-file)")
	workflowsEvalsUpdateCmd.Flags().String("run-id", "", "new source run id for a run_step source (inline alternative to --source-file)")
	workflowsEvalsUpdateCmd.Flags().String("step-id", "", "new source step id within --run-id (optional; pairs with --run-id)")

	workflowsEvalsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsEvalsRunsListCmd.Flags().String("workflow-id", "", "workflow id (alternative to the positional form)")
	workflowsEvalsRunsListCmd.Flags().String("eval-id", "", "filter by eval id")
	workflowsEvalsRunsListCmd.Flags().String("target-block-id", "", "filter by target block id")
	workflowsEvalsRunsListCmd.Flags().String("status", "", "filter by lifecycle status")
	workflowsEvalsRunsListCmd.Flags().String("exclude-status", "", "exclude lifecycle status")
	workflowsEvalsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsEvalsRunsListCmd.Flags().String("from-date", "", "created on or after YYYY-MM-DD")
	workflowsEvalsRunsListCmd.Flags().String("to-date", "", "created on or before YYYY-MM-DD")
	workflowsEvalsRunsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsEvalsRunsListCmd.Flags().String("before", "", "page before cursor (mutually exclusive with --after)")
	workflowsEvalsRunsListCmd.Flags().String("after", "", "page after cursor (mutually exclusive with --before)")
	workflowsEvalsRunsListCmd.Flags().String("order", "", "asc or desc")
	workflowsEvalsRunsCreateCmd.Flags().String("eval-id", "", "single eval to run")
	workflowsEvalsRunsCreateCmd.Flags().String("target-file", "", "JSON file with target (or - for stdin)")
	// --wait blocks until the run settles (completed/error/cancelled);
	// --poll-interval-ms / --timeout-seconds tune the poll loop, matching
	// `experiments runs create --wait`.
	workflowsEvalsRunsCreateCmd.Flags().Bool("wait", false, "block until the run reaches a terminal status (completed/error/cancelled), then print the final run")
	workflowsEvalsRunsCreateCmd.Flags().Int("poll-interval-ms", 2000, "poll cadence in milliseconds while --wait is set")
	workflowsEvalsRunsCreateCmd.Flags().Int("timeout-seconds", 600, "max seconds to wait while --wait is set")
	// Standalone poller for an already-running eval run; tuning flags match
	// the create knobs and the experiment/primitive wait commands.
	addPrimitiveWaitTuningFlags(workflowsEvalsRunsWaitCmd, false)
	workflowsEvalsResultsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")

	workflowsEvalsResultsCmd.AddCommand(workflowsEvalsResultsListCmd, workflowsEvalsResultsGetCmd)
	workflowsEvalsRunsCmd.AddCommand(workflowsEvalsRunsCreateCmd, workflowsEvalsRunsListCmd, workflowsEvalsRunsGetCmd, workflowsEvalsRunsCancelCmd, workflowsEvalsRunsWaitCmd)
	workflowsEvalsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsEvalsCmd.AddCommand(workflowsEvalsCreateCmd, workflowsEvalsGetCmd, workflowsEvalsListCmd, workflowsEvalsUpdateCmd, workflowsEvalsDeleteCmd, workflowsEvalsRunsCmd, workflowsEvalsResultsCmd)
	workflowsCmd.AddCommand(workflowsEvalsCmd)
}
