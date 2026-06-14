//go:build !retab_oagen_cli_workflows_tests

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
// fields on workflow-tests create/update requests from the CLI's loose
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

var workflowsTestsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Manage workflow tests",
	Long: `Declarative regression tests for workflow blocks.

A test pins a target block's expected output against a recorded input
(typically captured from a successful past run). Run the test suite
after any change — config update, model swap, prompt edit — to
catch silent drift in extraction quality or classification accuracy.

A test has three pieces:

  * ` + "`target`" + `   — which block is under test
  * ` + "`source`" + `   — the input the block should consume (often a
    pointer to an existing run step)
  * ` + "`assertion`" + ` — the output handle/path and expected condition

Use ` + "`workflows tests runs create`" + ` to run a single test or the whole
suite. Inspect runs via ` + "`workflows tests runs list`" + `.

For exploring multiple alternative block configurations side-by-side
(A/B-style), use ` + "`retab workflows experiments --help`" + ` instead.`,
	Example: `  # List a workflow's tests
  retab workflows tests list wf_abc123

  # Create a test pinning a block's expected output
  retab workflows tests create wf_abc123 \
    --name "Invoice 17 baseline" \
    --target-file ./target.json \
    --source-file ./source.json \
    --assertion-file ./assertion.json

  # Create a run for every test in the workflow
  retab workflows tests runs create wf_abc123`,
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

// resolveTestComponent picks a single test component (target / source /
// assertion) from EITHER its `--<flag>-file` JSON file OR the inline flag
// form, enforcing that the two forms are not mixed for the same component.
// Returns nil when neither form was supplied so the caller can report which
// piece is missing.
func resolveTestComponent(cmd *cobra.Command, fileFlag, inlineFlagDesc string, inline map[string]any, inlineSet bool) (map[string]any, error) {
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

// inlineTestTarget builds the target descriptor from --block-id. Returns
// (nil, false) when --block-id was not supplied.
func inlineTestTarget(cmd *cobra.Command) (map[string]any, bool) {
	blockID, _ := cmd.Flags().GetString("block-id")
	if strings.TrimSpace(blockID) == "" {
		return nil, false
	}
	return map[string]any{"type": "block", "block_id": blockID}, true
}

// inlineTestSource builds a run_step source from --run-id (and optional
// --step-id). Returns (nil, false) when --run-id was not supplied.
func inlineTestSource(cmd *cobra.Command) (map[string]any, bool) {
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

// inlineTestAssertion builds an `equals` assertion from --output-handle-id /
// --path / --equals. It is considered "set" when any of those flags is
// provided; --equals is then mandatory (the assertion needs an expected
// value), and --output-handle-id defaults to the conventional single-JSON
// output handle. The expected value is parsed as a JSON literal when possible
// (so `--equals 300000` is the number 300000, `--equals true` the boolean)
// and falls back to the raw string otherwise.
func inlineTestAssertion(cmd *cobra.Command) (map[string]any, bool, error) {
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

// validateWorkflowTestName trims surrounding whitespace and rejects names
// that are empty after trimming. Returns the cleaned value so callers can
// propagate the trimmed string into the outgoing request body — without
// this transformation `workflows tests create --name "  padded  "` would
// silently persist with the padding (bug #3). Sibling validators
// (validateWorkflowName, validateExperimentName) share the same shape.
func validateWorkflowTestName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("test name must not be blank")
	}
	return trimmed, nil
}

func validateWorkflowTestAssertion(assertion map[string]any) error {
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

func validateWorkflowTestTarget(target map[string]any) error {
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

func validateWorkflowTestSource(source map[string]any) error {
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

var workflowsTestsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> [flags]",
	Short: "Create a workflow test",
	Long: `Create a regression test pinning a block's expected output.
You'll typically capture the three files from a successful past run:

  ` + "`--target-file`" + `     — which block to test (e.g.
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

After creation, run with ` + "`workflows tests runs create`" + `.`,
	Example: `  # Create a test from JSON files
  retab workflows tests create wf_abc123 \
    --name "Invoice 17 baseline" \
    --target-file ./target.json \
    --source-file ./source.json \
    --assertion-file ./assertion.json

  # Same test, inline — no JSON files needed
  retab workflows tests create wf_abc123 \
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
		req := retab.WorkflowTestsCreateParams{}
		req.WorkflowID = workflowID
		rawName, _ := cmd.Flags().GetString("name")
		trimmedName, err := validateWorkflowTestName(rawName)
		if err != nil {
			return err
		}
		req.Name = &trimmedName
		// Each component may be supplied as a JSON file OR via the inline
		// flag form — but not both for the same component.
		inlineTarget, inlineTargetSet := inlineTestTarget(cmd)
		inlineSource, inlineSourceSet := inlineTestSource(cmd)
		inlineAssertion, inlineAssertionSet, err := inlineTestAssertion(cmd)
		if err != nil {
			return err
		}
		target, err := resolveTestComponent(cmd, "target-file", "--block-id", inlineTarget, inlineTargetSet)
		if err != nil {
			return err
		}
		source, err := resolveTestComponent(cmd, "source-file", "--run-id/--step-id", inlineSource, inlineSourceSet)
		if err != nil {
			return err
		}
		assertion, err := resolveTestComponent(cmd, "assertion-file", "--output-handle-id/--path/--equals", inlineAssertion, inlineAssertionSet)
		if err != nil {
			return err
		}
		if target == nil || source == nil || assertion == nil {
			return fmt.Errorf("each of target, source, and assertion is required — supply it as a file " +
				"(--target-file / --source-file / --assertion-file) or inline " +
				"(--block-id; --run-id [--step-id]; --path --equals)")
		}
		if err := validateWorkflowTestTarget(target); err != nil {
			return fmt.Errorf("--target-file: %w", err)
		}
		if err := validateWorkflowTestSource(source); err != nil {
			return fmt.Errorf("--source-file: %w", err)
		}
		if err := validateWorkflowTestAssertion(assertion); err != nil {
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
		result, err := client.Workflows.Tests.Create(ctx, &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsGetCmd = &cobra.Command{
	Use:   "get <test-id>",
	Short: "Get a test",
	Long: `Fetch a test's definition: target block, source input, assertion
output, name, timestamps.`,
	Example: `  # Inspect a test
  retab workflows tests get wfnodetest_jkl012`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List tests",
	Long: `List every test attached to a workflow. Filter by
` + "`--target-block-id`" + ` to focus on the regression suite for a
particular block.

Name the workflow either positionally (` + "`list <workflow-id>`" + `) or with
the ` + "`--workflow-id`" + ` flag — the two forms are equivalent. Passing both
is accepted when they agree; an error is raised only when they disagree. The
workflow id is required: tests have no org-wide listing.`,
	Example: `  # All tests in a workflow (positional)
  retab workflows tests list wf_abc123

  # Same, with the flag form
  retab workflows tests list --workflow-id wf_abc123

  # Just the tests guarding one block
  retab workflows tests list wf_abc123 --target-block-id block_extract_1`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Workflow id positionally OR via --workflow-id (co-equal forms);
		// required here — tests have no org-wide listing.
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
		req := retab.WorkflowTestsListParams{WorkflowID: workflowID}
		if v, _ := cmd.Flags().GetString("target-block-id"); v != "" {
			req.TargetBlockID = ptr(v)
		}
		if v := getIntFlagOrDefault(cmd, "limit", 50); v > 0 {
			req.Limit = ptr(v)
		}
		result, err := client.Workflows.Tests.List(ctx, &req)
		if err != nil {
			return err
		}
		return printWorkflowTestsListResult(cmd, result)
	}),
}

func printWorkflowTestsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowTest]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputTable) {
			return RenderList(os.Stdout, OutputTable, result, workflowTestColumns)
		}
	}
	return printJSON(result)
}

var workflowTestColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowTestCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return workflowTestCell(row, "name") }},
	{Header: "TARGET_BLOCK_ID", Extract: func(row any) string { return workflowTestCell(row, "target.block_id") }},
	{Header: "FRESHNESS", Extract: artifactFreshnessCell},
	{Header: "SCHEMA_DRIFT", Extract: func(row any) string { return workflowTestCell(row, "schema_drift") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowTestCell(row, "created_at") }},
}

func workflowTestCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

// workflowTestRunColumns is the dedicated TableColumn spec for
// `workflows tests runs list --output table/csv`. The generic auto-renderer only
// surfaced ID + a TYPE column that confusingly showed the lifecycle status; these
// columns show what a user reads off a test-suite run: status and the pass/fail
// tally.
var workflowTestRunColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowTestCell(row, "id") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowTestCell(row, "lifecycle.status") }},
	{Header: "TESTS", Extract: func(row any) string { return workflowTestCell(row, "total_tests") }},
	{Header: "PASSED", Extract: func(row any) string { return workflowTestCell(row, "counts.outcome.passed") }},
	{Header: "FAILED", Extract: func(row any) string { return workflowTestCell(row, "counts.outcome.failed") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowTestCell(row, "timing.created_at") }},
}

func printWorkflowTestsRunsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowTestRun]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			switch f.Value.String() {
			case string(OutputTable), string(OutputCSV):
				return RenderList(os.Stdout, OutputFormat(f.Value.String()), result, workflowTestRunColumns)
			}
		}
	}
	return printJSON(result)
}

// workflowTestResultColumns is the dedicated TableColumn spec for
// `workflows tests results list --output table/csv`. The generic auto-renderer
// dropped VERDICT — the single most important field of a test result
// (passed/failed) — and mislabeled the status as TYPE.
var workflowTestResultColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowTestCell(row, "id") }},
	{Header: "VERDICT", Extract: func(row any) string { return workflowTestCell(row, "verdict") }},
	{Header: "TARGET", Extract: func(row any) string { return workflowTestCell(row, "target.block_id") }},
	{Header: "SOURCE", Extract: func(row any) string { return workflowTestCell(row, "source.type") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowTestCell(row, "lifecycle.status") }},
}

func printWorkflowTestResultsListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowTestResult]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			switch f.Value.String() {
			case string(OutputTable), string(OutputCSV):
				return RenderList(os.Stdout, OutputFormat(f.Value.String()), result, workflowTestResultColumns)
			}
		}
	}
	return printJSON(result)
}

var workflowsTestsUpdateCmd = &cobra.Command{
	Use:   "update <test-id>",
	Short: "Update a test",
	Long: `Re-pin a test's expected output or source input. Use this when
a deliberate schema or prompt change makes the old assertion stale — the
intent is to ratify the new output as the new baseline, not to silence
flaky runs.`,
	Example: `  # Refresh the assertion after a deliberate schema change
  retab workflows tests update wfnodetest_jkl012 \
    --assertion-file ./new-assertion.json

  # Same, inline — no JSON file needed (mirrors 'tests create')
  retab workflows tests update wfnodetest_jkl012 \
    --path net_amount_payable_usd --equals 300000

  # Re-pin the source run inline
  retab workflows tests update wfnodetest_jkl012 --run-id run_xyz789

  # Rename a test
  retab workflows tests update wfnodetest_jkl012 \
    --name "Invoice 17 baseline (v2 schema)"`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// The assertion and source may each be supplied as a JSON file OR via the
		// same inline flag form `tests create` accepts (mutually exclusive per
		// component, enforced by resolveTestComponent).
		inlineSource, inlineSourceSet := inlineTestSource(cmd)
		inlineAssertion, inlineAssertionSet, err := inlineTestAssertion(cmd)
		if err != nil {
			return err
		}
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") &&
			!cmd.Flags().Changed("assertion-file") && !inlineAssertionSet &&
			!cmd.Flags().Changed("source-file") && !inlineSourceSet {
			return fmt.Errorf("nothing to update: pass at least one of --name, the assertion " +
				"(--assertion-file or --output-handle-id/--path/--equals), or the source " +
				"(--source-file or --run-id/--step-id)")
		}
		req := retab.WorkflowTestsUpdateParams{}
		if cmd.Flags().Changed("name") {
			rawName, _ := cmd.Flags().GetString("name")
			trimmed, err := validateWorkflowTestName(rawName)
			if err != nil {
				return err
			}
			req.Name = &trimmed
		}
		assertion, err := resolveTestComponent(cmd, "assertion-file", "--output-handle-id/--path/--equals", inlineAssertion, inlineAssertionSet)
		if err != nil {
			return err
		}
		source, err := resolveTestComponent(cmd, "source-file", "--run-id/--step-id", inlineSource, inlineSourceSet)
		if err != nil {
			return err
		}
		if assertion != nil {
			if err := validateWorkflowTestAssertion(assertion); err != nil {
				return fmt.Errorf("assertion: %w", err)
			}
			req.Assertion = &retab.AssertionSpec{}
			if err := decodeJSONInto("assertion", assertion, req.Assertion); err != nil {
				return err
			}
		}
		if source != nil {
			if err := validateWorkflowTestSource(source); err != nil {
				return fmt.Errorf("--source-file: %w", err)
			}
			req.Source = &retab.WorkflowTestSource{}
			if err := decodeJSONInto("source", source, req.Source); err != nil {
				return err
			}
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Update(ctx, args[0], &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsDeleteCmd = &cobra.Command{
	Use:   "delete <test-id>",
	Short: "Delete a test",
	Long: `Permanently delete a regression test and its run history.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal. Run history is removed alongside the test definition.`,
	Example: `  # Drop a stale test (interactive, asks to confirm)
  retab workflows tests delete wfnodetest_jkl012

  # Skip the prompt in scripts
  retab workflows tests delete wfnodetest_jkl012 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "test", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Tests.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("test", args[0])
		return nil
	}),
}

// ---- test runs subgroup ----
//
// SDK-backed canonical routes: "/v1/workflows/tests/runs" and
// "/v1/workflows/tests/results".

var workflowsTestsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Create and inspect test runs",
	Long: `Create workflow-test runs, poll their status, and inspect child
result records.`,
	Example: `  # Create a run for the whole workflow
  retab workflows tests runs create wf_abc123

  # Poll a workflow-test run
  retab workflows tests runs get wftestrun_mno345

  # Fetch child result rows for a workflow-test run
  retab workflows tests results list wftestrun_mno345

  # Recent runs for one workflow (narrow to a test with --test-id)
  retab workflows tests runs list wf_abc123 --test-id wfnodetest_jkl012`,
}

var workflowsTestsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> [test-id] [flags]",
	Short: "Create a workflow-test run",
	Long: `Start a workflow-test run. The first positional argument is the
` + "`workflow-id`" + ` — by default the run executes every test attached to
the workflow. Pass a second positional ` + "`test-id`" + ` (or ` + "`--test-id`" + `)
to run just that one test.

To run a single test pass ` + "`--test-id`" + `; to run every saved test
for one block pass ` + "`--target-file`" + ` with a JSON target such as
` + `{"type":"block","block_id":"extract_1"}` + `.

Pass ` + "`--wait`" + ` to block until the run reaches a terminal status
(` + "`completed`" + `/` + "`error`" + `/` + "`cancelled`" + `) and print the
final run — saving you from scripting a poll loop around
` + "`runs get`" + `. With ` + "`--wait`" + ` the command exits non-zero if the
run ends in ` + "`error`" + `/` + "`cancelled`" + ` or the timeout elapses.
Without it, poll progress with ` + "`workflows tests runs get`" + ` and fetch
per-test results with ` + "`workflows tests results list`" + `.`,
	Example: `  # Run every test in the workflow
  retab workflows tests runs create wf_abc123

  # Run a single test
  retab workflows tests runs create wf_abc123 \
    --test-id wfnodetest_jkl012

  # Run every test for one block
  retab workflows tests runs create wf_abc123 \
    --target-file ./target.json

  # Create and block until the suite finishes (2s polls, 10m timeout)
  retab workflows tests runs create wf_abc123 --wait

  # Tune the polling cadence and ceiling
  retab workflows tests runs create wf_abc123 \
    --wait --poll-interval-ms 1000 --timeout-seconds 1800`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := &retab.WorkflowTestRunsCreateParams{
			WorkflowID: args[0],
		}
		testID, _ := cmd.Flags().GetString("test-id")
		// Tolerate `tests runs create <workflow-id> <test-id>`: a second
		// positional is a convenience alias for --test-id (run that one test),
		// matching how users carry over the `blocks/edges <wf-id> <child-id>`
		// shape. The flag and the positional are mutually exclusive.
		if len(args) == 2 {
			if testID != "" {
				return fmt.Errorf("pass the test id once: either the second positional argument or --test-id, not both")
			}
			testID = args[1]
		}
		target, err := resolveJSONMap(cmd, "target-file")
		if err != nil {
			return err
		}
		if testID != "" && target != nil {
			return fmt.Errorf("--test-id and --target-file are mutually exclusive")
		}
		if testID != "" {
			params.Scope = &retab.WorkflowTestRunScope{
				Type:   retab.WorkflowTestRunScopeTypeSingle,
				TestID: ptr(testID),
			}
		}
		if target != nil {
			if err := validateWorkflowTestTarget(target); err != nil {
				return fmt.Errorf("--target-file: %w", err)
			}
			blockID, _ := target["block_id"].(string)
			params.Scope = &retab.WorkflowTestRunScope{
				Type:    retab.WorkflowTestRunScopeTypeBlock,
				BlockID: ptr(blockID),
			}
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Runs.Create(ctx, params)
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
		final, waitErr := waitForWorkflowTestRun(ctx, client, result.ID, pollInterval, timeout)
		if final != nil {
			if err := printResult(cmd, final); err != nil {
				return err
			}
		}
		if waitErr != nil {
			return waitErr
		}
		return workflowTestRunTerminalError(final)
	}),
}

// workflowTestRunStatus reads the lifecycle discriminator off a workflow-test
// run. The lifecycle is a discriminated-union envelope, so the status string
// comes from its `Status()` getter rather than a plain field.
func workflowTestRunStatus(run *retab.WorkflowTestRun) string {
	if run == nil {
		return ""
	}
	return run.Lifecycle.Status()
}

// isTerminalWorkflowTestRun reports whether the run has settled. A
// workflow-test run can't pause for review, so the terminal set is just
// completed / error / cancelled (pending/queued/running are still in flight).
func isTerminalWorkflowTestRun(run *retab.WorkflowTestRun) bool {
	switch workflowTestRunStatus(run) {
	case "completed", "error", "cancelled":
		return true
	default:
		return false
	}
}

// workflowTestRunTerminalError maps a settled-but-unsuccessful run to a
// non-zero exit. completed (or an unknown/empty status) is success;
// error/cancelled is failure.
func workflowTestRunTerminalError(run *retab.WorkflowTestRun) error {
	if run == nil {
		return nil
	}
	switch status := workflowTestRunStatus(run); status {
	case "", "completed":
		return nil
	case "error", "cancelled":
		if run.ID == "" {
			return fmt.Errorf("workflow-test run ended with status %s", status)
		}
		return fmt.Errorf("workflow-test run %s ended with status %s", run.ID, status)
	default:
		return nil
	}
}

// waitForWorkflowTestRun polls Runs.Get until the run reaches a terminal
// status or the timeout elapses. On timeout it returns the last-observed
// (non-final) run alongside the error so callers can still surface partial
// state — mirroring waitForExperimentRun.
func waitForWorkflowTestRun(
	ctx context.Context,
	client *retab.Client,
	runID string,
	pollInterval time.Duration,
	timeout time.Duration,
) (*retab.WorkflowTestRun, error) {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		run, err := client.Workflows.Tests.Runs.Get(ctx, runID)
		if err != nil {
			return nil, err
		}
		if isTerminalWorkflowTestRun(run) {
			return run, nil
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return run, fmt.Errorf("timed out waiting for workflow-test run %s: %w", runID, ctx.Err())
		case <-timer.C:
		}
	}
}

// workflowsTestsRunsWaitCmd is the standalone poller for an already-created
// workflow-test run, mirroring `experiments runs wait` so the
// `create --wait` / standalone-`wait` pair is consistent across the run
// families.
var workflowsTestsRunsWaitCmd = &cobra.Command{
	Use:   "wait <run-id>",
	Short: "Poll until a workflow-test run reaches a terminal status",
	Long: `Block until a workflow-test run hits a terminal status
(` + "`completed`" + `, ` + "`error`" + `, or ` + "`cancelled`" + `),
polling on a configurable interval. Defaults: 2-second polls, 10-minute
timeout.

Cleaner than scripting a poll loop around ` + "`runs get`" + ` — the CLI
handles the interval and timeout, and exits non-zero if the run ends in
` + "`error`" + `/` + "`cancelled`" + ` or the timeout elapses. Pair with
` + "`runs create --wait`" + ` to create and block in a single step.`,
	Example: `  # Wait with defaults (2s polls, 600s timeout)
  retab workflows tests runs wait wftestrun_mno345

  # Faster polls, longer ceiling
  retab workflows tests runs wait wftestrun_mno345 \
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
		final, waitErr := waitForWorkflowTestRun(ctx, client, args[0], pollInterval, timeout)
		if final != nil {
			if err := printResult(cmd, final); err != nil {
				return err
			}
		}
		if waitErr != nil {
			return waitErr
		}
		return workflowTestRunTerminalError(final)
	}),
}

var workflowsTestsRunsListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List workflow-test runs",
	Long: `List workflow-test runs. Filter by workflow, test, target block,
status, trigger, date, or cursor.

Name the workflow either positionally (` + "`list <workflow-id>`" + `) or with
the ` + "`--workflow-id`" + ` flag — the two forms are equivalent. Passing both
is accepted when they agree; an error is raised only when they disagree. The
workflow id is optional here: with neither form, runs are listed
workspace-wide.`,
	Example: `  # Recent runs for one workflow (positional)
  retab workflows tests runs list wf_abc123 --limit 50

  # Narrow to a single test (flag form still accepted)
  retab workflows tests runs list wf_abc123 --test-id wfnodetest_jkl012 --limit 50`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateWorkflowTestsRunsListFilters(cmd); err != nil {
			return err
		}
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		// Workflow id positionally OR via --workflow-id (co-equal forms);
		// optional here — test runs have a workspace-wide listing.
		workflowID, err := resolveWorkflowScope(cmd, args, false)
		if err != nil {
			return err
		}
		params := &retab.WorkflowTestRunsListParams{
			PaginationParams: retab.PaginationParams{
				Limit: ptr(getIntFlagOrDefault(cmd, "limit", 20)),
			},
		}
		if workflowID != "" {
			params.WorkflowID = ptr(workflowID)
		}
		if v, _ := cmd.Flags().GetString("test-id"); v != "" {
			params.TestID = ptr(v)
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
			parsed, _ := time.Parse("2006-01-02", v)
			params.FromDate = &parsed
			dateQuery.Set("from_date", v)
		}
		if v, _ := cmd.Flags().GetString("to-date"); v != "" {
			parsed, _ := time.Parse("2006-01-02", v)
			params.ToDate = &parsed
			dateQuery.Set("to_date", v)
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
		result, err := client.Workflows.Tests.Runs.List(ctx, params, retab.WithRequestParams(dateQuery))
		if err != nil {
			return err
		}
		return printWorkflowTestsRunsListResult(cmd, result)
	}),
}

// workflowTestRunStatusValues mirrors `WorkflowTestPublicRunStatus` in
// `backend/.../tests/models.py`. A workflow-test run can't pause for
// review, so `awaiting_review` is intentionally absent.
var allowedWorkflowTestRunStatuses = map[string]bool{
	"pending":   true,
	"running":   true,
	"completed": true,
	"error":     true,
	"cancelled": true,
}

const workflowTestRunStatusValues = "pending, running, completed, error, cancelled"

func validateWorkflowTestsRunsListFilters(cmd *cobra.Command) error {
	if err := validateEnumFlag(cmd, "status", allowedWorkflowTestRunStatuses, workflowTestRunStatusValues); err != nil {
		return err
	}
	if err := validateEnumFlag(cmd, "exclude-status", allowedWorkflowTestRunStatuses, workflowTestRunStatusValues); err != nil {
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

var workflowsTestsRunsGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get a workflow-test run",
	Long:  `Fetch a workflow-test run by run id.`,
	Example: `  # Poll a workflow-test run
  retab workflows tests runs get wftestrun_mno345`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Runs.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsRunsCancelCmd = &cobra.Command{
	Use:   "cancel <run-id>",
	Short: "Cancel a workflow-test run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Runs.Cancel(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Inspect workflow-test run results",
}

var workflowsTestsResultsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List child results for a workflow-test run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := &retab.WorkflowTestRunResultsListParams{
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
		result, err := client.Workflows.Tests.Results.List(ctx, params)
		if err != nil {
			return err
		}
		return printWorkflowTestResultsListResult(cmd, result)
	}),
}

var workflowsTestsResultsGetCmd = &cobra.Command{
	Use:   "get <result-id>",
	Short: "Get a workflow-test result",
	Long:  `Fetch one workflow-test result by flat result id.`,
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Results.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsTestsCreateCmd.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	workflowsTestsCreateCmd.Flags().String("name", "", "test name")
	workflowsTestsCreateCmd.Flags().String("target-file", "", "JSON file with the target object (or - for stdin). Alternative to --block-id.")
	workflowsTestsCreateCmd.Flags().String("source-file", "", "JSON file with the source object (or - for stdin). Two accepted shapes: {\"type\":\"manual\",\"handle_inputs\":{...}} or {\"type\":\"run_step\",\"run_id\":\"run_xxx\",\"step_id\":\"...\"} (step_id optional). Alternative to --run-id. See 'tests create --help' for the full schema.")
	workflowsTestsCreateCmd.Flags().String("assertion-file", "", "JSON file with the assertion object (or - for stdin). Alternative to --path/--equals.")
	// Inline alternative to the three JSON files for the common
	// "assert block X's field Y equals Z from run R" case.
	workflowsTestsCreateCmd.Flags().String("block-id", "", "target block id (inline alternative to --target-file)")
	workflowsTestsCreateCmd.Flags().String("run-id", "", "source run id for a run_step source (inline alternative to --source-file)")
	workflowsTestsCreateCmd.Flags().String("step-id", "", "source step id within --run-id (optional; pairs with --run-id)")
	workflowsTestsCreateCmd.Flags().String("output-handle-id", "", "assertion output handle id (inline; defaults to output-json-0)")
	workflowsTestsCreateCmd.Flags().String("path", "", "assertion field path inside the output handle (inline; pairs with --equals)")
	workflowsTestsCreateCmd.Flags().String("equals", "", "expected value for an inline equals assertion (JSON literal or string; alternative to --assertion-file)")
	// Keep the flag hidden but DO NOT use MarkDeprecated — cobra's auto warning
	// duplicates the more-specific message emitted by resolveWorkflowIDArg.
	_ = workflowsTestsCreateCmd.Flags().MarkHidden("workflow-id")

	workflowsTestsListCmd.Flags().String("workflow-id", "", "workflow id (alternative to the positional form)")
	workflowsTestsListCmd.Flags().String("target-block-id", "", "filter by target block id")
	workflowsTestsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 50)")

	workflowsTestsUpdateCmd.Flags().String("name", "", "new test name")
	workflowsTestsUpdateCmd.Flags().String("assertion-file", "", "JSON file with new assertion (or - for stdin). Alternative to --output-handle-id/--path/--equals.")
	workflowsTestsUpdateCmd.Flags().String("source-file", "", "JSON file with new source (or - for stdin). Alternative to --run-id/--step-id.")
	// Inline assertion/source flags, mirroring `tests create`, so a test's
	// assertion or source can be re-pinned without hand-authoring JSON.
	workflowsTestsUpdateCmd.Flags().String("output-handle-id", "", "new assertion output handle id (inline; defaults to output-json-0)")
	workflowsTestsUpdateCmd.Flags().String("path", "", "new assertion field path inside the output handle (inline; pairs with --equals)")
	workflowsTestsUpdateCmd.Flags().String("equals", "", "new expected value for an inline equals assertion (JSON literal or string; alternative to --assertion-file)")
	workflowsTestsUpdateCmd.Flags().String("run-id", "", "new source run id for a run_step source (inline alternative to --source-file)")
	workflowsTestsUpdateCmd.Flags().String("step-id", "", "new source step id within --run-id (optional; pairs with --run-id)")

	workflowsTestsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsTestsRunsListCmd.Flags().String("workflow-id", "", "workflow id (alternative to the positional form)")
	workflowsTestsRunsListCmd.Flags().String("test-id", "", "filter by test id")
	workflowsTestsRunsListCmd.Flags().String("target-block-id", "", "filter by target block id")
	workflowsTestsRunsListCmd.Flags().String("status", "", "filter by lifecycle status")
	workflowsTestsRunsListCmd.Flags().String("exclude-status", "", "exclude lifecycle status")
	workflowsTestsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsTestsRunsListCmd.Flags().String("from-date", "", "created on or after YYYY-MM-DD")
	workflowsTestsRunsListCmd.Flags().String("to-date", "", "created on or before YYYY-MM-DD")
	workflowsTestsRunsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsTestsRunsListCmd.Flags().String("before", "", "page before cursor (mutually exclusive with --after)")
	workflowsTestsRunsListCmd.Flags().String("after", "", "page after cursor (mutually exclusive with --before)")
	workflowsTestsRunsListCmd.Flags().String("order", "", "asc or desc")
	workflowsTestsRunsCreateCmd.Flags().String("test-id", "", "single test to run")
	workflowsTestsRunsCreateCmd.Flags().String("target-file", "", "JSON file with target (or - for stdin)")
	// --wait blocks until the run settles (completed/error/cancelled);
	// --poll-interval-ms / --timeout-seconds tune the poll loop, matching
	// `experiments runs create --wait`.
	workflowsTestsRunsCreateCmd.Flags().Bool("wait", false, "block until the run reaches a terminal status (completed/error/cancelled), then print the final run")
	workflowsTestsRunsCreateCmd.Flags().Int("poll-interval-ms", 2000, "poll cadence in milliseconds while --wait is set")
	workflowsTestsRunsCreateCmd.Flags().Int("timeout-seconds", 600, "max seconds to wait while --wait is set")
	// Standalone poller for an already-running test run; tuning flags match
	// the create knobs and the experiment/primitive wait commands.
	addPrimitiveWaitTuningFlags(workflowsTestsRunsWaitCmd, false)
	workflowsTestsResultsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")

	workflowsTestsResultsCmd.AddCommand(workflowsTestsResultsListCmd, workflowsTestsResultsGetCmd)
	workflowsTestsRunsCmd.AddCommand(workflowsTestsRunsCreateCmd, workflowsTestsRunsListCmd, workflowsTestsRunsGetCmd, workflowsTestsRunsCancelCmd, workflowsTestsRunsWaitCmd)
	workflowsTestsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsTestsCmd.AddCommand(workflowsTestsCreateCmd, workflowsTestsGetCmd, workflowsTestsListCmd, workflowsTestsUpdateCmd, workflowsTestsDeleteCmd, workflowsTestsRunsCmd, workflowsTestsResultsCmd)
	workflowsCmd.AddCommand(workflowsTestsCmd)
}
