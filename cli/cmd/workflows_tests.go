package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

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
			fmt.Fprintln(warnTo, "warning: --workflow-id is deprecated; positional argument takes precedence")
		}
		return args[0], nil
	}
	if flagSet && strings.TrimSpace(flagVal) != "" {
		fmt.Fprintln(warnTo, "warning: --workflow-id is deprecated; pass the workflow id as the first positional argument")
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
  ` + `{"type": "block", "block_id": "blk_extract_1"}` + `).

  ` + "`--source-file`" + `     — the input. Usually a reference to a
  run/step that supplied known-good input.

  ` + "`--assertion-file`" + `  — the output handle/path and condition
  (e.g. ` + `{"target":{"output_handle_id":"output-json-0","path":"total"},"condition":{"kind":"equals","expected":120}}` + `).

After creation, run with ` + "`workflows tests runs create`" + `.`,
	Example: `  # Create a test from JSON files
  retab workflows tests create wf_abc123 \
    --name "Invoice 17 baseline" \
    --target-file ./target.json \
    --source-file ./source.json \
    --assertion-file ./assertion.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, err := resolveWorkflowIDArg(cmd, args)
		if err != nil {
			return err
		}
		req := retab.WorkflowTestCreateRequest{}
		req.WorkflowID = workflowID
		req.Name, _ = cmd.Flags().GetString("name")
		target, err := resolveJSONMap(cmd, "target-file")
		if err != nil {
			return err
		}
		source, err := resolveJSONMap(cmd, "source-file")
		if err != nil {
			return err
		}
		assertion, err := resolveJSONMap(cmd, "assertion-file")
		if err != nil {
			return err
		}
		if target == nil || source == nil || assertion == nil {
			return fmt.Errorf("--target-file, --source-file, and --assertion-file are required")
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
		req.Target = retab.Resource(target)
		req.Source = retab.Resource(source)
		req.Assertion = retab.Resource(assertion)
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Create(ctx, req)
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
  retab workflows tests get tst_jkl012`,
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
	Use:   "list <workflow-id>",
	Short: "List tests",
	Long: `List every test attached to a workflow. Filter by
` + "`--target-block-id`" + ` to focus on the regression suite for a
particular block.`,
	Example: `  # All tests in a workflow
  retab workflows tests list wf_abc123

  # Just the tests guarding one block
  retab workflows tests list wf_abc123 --target-block-id blk_extract_1`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.ListWorkflowTestsRequest{WorkflowID: args[0]}
		req.TargetBlockID, _ = cmd.Flags().GetString("target-block-id")
		req.Limit = getIntFlagOrDefault(cmd, "limit", 50)
		result, err := client.Workflows.Tests.List(ctx, req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsUpdateCmd = &cobra.Command{
	Use:   "update <test-id>",
	Short: "Update a test",
	Long: `Re-pin a test's expected output or source input. Use this when
a deliberate schema or prompt change makes the old assertion stale — the
intent is to ratify the new output as the new baseline, not to silence
flaky runs.`,
	Example: `  # Refresh the assertion after a deliberate schema change
  retab workflows tests update tst_jkl012 \
    --assertion-file ./new-assertion.json

  # Rename a test
  retab workflows tests update tst_jkl012 \
    --name "Invoice 17 baseline (v2 schema)"`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("assertion-file") &&
			!cmd.Flags().Changed("source-file") {
			return fmt.Errorf("nothing to update: pass at least one of --name, --assertion-file, or --source-file")
		}
		req := retab.WorkflowTestUpdateRequest{}
		req.Name, _ = cmd.Flags().GetString("name")
		assertion, err := resolveJSONMap(cmd, "assertion-file")
		if err != nil {
			return err
		}
		source, err := resolveJSONMap(cmd, "source-file")
		if err != nil {
			return err
		}
		if assertion != nil {
			if err := validateWorkflowTestAssertion(assertion); err != nil {
				return fmt.Errorf("--assertion-file: %w", err)
			}
			req.Assertion = retab.Resource(assertion)
		}
		if source != nil {
			if err := validateWorkflowTestSource(source); err != nil {
				return fmt.Errorf("--source-file: %w", err)
			}
			req.Source = retab.Resource(source)
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Update(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsDeleteCmd = &cobra.Command{
	Use:   "delete <test-id>",
	Short: "Delete a test",
	Long:  `Permanently delete a regression test and its run history.`,
	Example: `  # Drop a stale test
  retab workflows tests delete tst_jkl012`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
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
  retab workflows tests runs results list wftestrun_mno345

  # Recent runs for one test
  retab workflows tests runs list --workflow-id wf_abc123 --test-id tst_jkl012`,
}

var workflowsTestsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id> [flags]",
	Short: "Create a workflow-test run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if testID, _ := cmd.Flags().GetString("test-id"); testID != "" {
			body["test_id"] = testID
		}
		if nConsensus, _ := cmd.Flags().GetInt("n-consensus"); nConsensus != 0 {
			body["n_consensus"] = nConsensus
		}
		target, err := resolveJSONMap(cmd, "target-file")
		if err != nil {
			return err
		}
		if target != nil {
			if err := validateWorkflowTestTarget(target); err != nil {
				return fmt.Errorf("--target-file: %w", err)
			}
			body["target"] = target
		}
		body["workflow_id"] = args[0]
		result, err := cliJSONRequest(cmd, http.MethodPost, "/workflows/tests/runs", nil, body)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsRunsListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "List workflow-test runs",
	Long: `List workflow-test runs. Filter by workflow, test, target block,
status, trigger, date, or cursor.`,
	Example: `  # Recent runs for one test
  retab workflows tests runs list --workflow-id wf_abc123 --test-id tst_jkl012 --limit 50`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateWorkflowTestsRunsListFilters(cmd); err != nil {
			return err
		}
		query := url.Values{}
		for _, name := range []string{"workflow-id", "test-id", "target-block-id", "status", "statuses", "exclude-status", "trigger-type", "trigger-types", "from-date", "to-date", "sort-by", "fields", "before", "after", "order"} {
			if value, _ := cmd.Flags().GetString(name); value != "" {
				query.Set(strings.ReplaceAll(name, "-", "_"), value)
			}
		}
		limit := getIntFlagOrDefault(cmd, "limit", 20)
		query.Set("limit", strconv.Itoa(limit))
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/tests/runs", query, nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
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
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/tests/runs/"+url.PathEscape(args[0]), nil, nil)
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
		result, err := cliJSONRequest(cmd, http.MethodPost, "/workflows/tests/runs/"+url.PathEscape(args[0])+"/cancel", nil, map[string]any{})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsTestsRunsResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Inspect workflow-test run results",
}

var workflowsTestsRunsResultsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List child results for a workflow-test run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		query := url.Values{}
		limit := getIntFlagOrDefault(cmd, "limit", 20)
		query.Set("limit", strconv.Itoa(limit))
		result, err := cliJSONRequest(cmd, http.MethodGet, "/workflows/tests/runs/"+url.PathEscape(args[0])+"/results", query, nil)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsTestsCreateCmd.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	workflowsTestsCreateCmd.Flags().String("name", "", "test name")
	workflowsTestsCreateCmd.Flags().String("target-file", "", "JSON file with the target object (or - for stdin) (required)")
	workflowsTestsCreateCmd.Flags().String("source-file", "", "JSON file with the source object (or - for stdin) (required)")
	workflowsTestsCreateCmd.Flags().String("assertion-file", "", "JSON file with the assertion object (or - for stdin) (required)")
	// Keep the flag hidden but DO NOT use MarkDeprecated — cobra's auto warning
	// duplicates the more-specific message emitted by resolveWorkflowIDArg.
	_ = workflowsTestsCreateCmd.Flags().MarkHidden("workflow-id")

	workflowsTestsListCmd.Flags().String("target-block-id", "", "filter by target block id")
	workflowsTestsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 50)")

	workflowsTestsUpdateCmd.Flags().String("name", "", "new test name")
	workflowsTestsUpdateCmd.Flags().String("assertion-file", "", "JSON file with new assertion (or - for stdin)")
	workflowsTestsUpdateCmd.Flags().String("source-file", "", "JSON file with new source (or - for stdin)")

	workflowsTestsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")
	workflowsTestsRunsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsTestsRunsListCmd.Flags().String("test-id", "", "filter by test id")
	workflowsTestsRunsListCmd.Flags().String("target-block-id", "", "filter by target block id")
	workflowsTestsRunsListCmd.Flags().String("status", "", "filter by lifecycle status")
	workflowsTestsRunsListCmd.Flags().String("statuses", "", "comma-separated lifecycle statuses")
	workflowsTestsRunsListCmd.Flags().String("exclude-status", "", "exclude lifecycle status")
	workflowsTestsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsTestsRunsListCmd.Flags().String("trigger-types", "", "comma-separated trigger types")
	workflowsTestsRunsListCmd.Flags().String("from-date", "", "created on or after YYYY-MM-DD")
	workflowsTestsRunsListCmd.Flags().String("to-date", "", "created on or before YYYY-MM-DD")
	workflowsTestsRunsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsTestsRunsListCmd.Flags().String("fields", "", "comma-separated fields")
	workflowsTestsRunsListCmd.Flags().String("before", "", "page before cursor")
	workflowsTestsRunsListCmd.Flags().String("after", "", "page after cursor")
	workflowsTestsRunsListCmd.Flags().String("order", "", "asc or desc")
	workflowsTestsRunsCreateCmd.Flags().String("test-id", "", "single test to run")
	workflowsTestsRunsCreateCmd.Flags().Var(&consensusFlagValue{}, "n-consensus", "consensus count (3, 5, or 7)")
	workflowsTestsRunsCreateCmd.Flags().String("target-file", "", "JSON file with target (or - for stdin)")
	workflowsTestsRunsResultsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items (1-100; default 20)")

	workflowsTestsRunsResultsCmd.AddCommand(workflowsTestsRunsResultsListCmd)
	workflowsTestsRunsCmd.AddCommand(workflowsTestsRunsCreateCmd, workflowsTestsRunsListCmd, workflowsTestsRunsGetCmd, workflowsTestsRunsCancelCmd, workflowsTestsRunsResultsCmd)
	workflowsTestsCmd.AddCommand(workflowsTestsCreateCmd, workflowsTestsGetCmd, workflowsTestsListCmd, workflowsTestsUpdateCmd, workflowsTestsDeleteCmd, workflowsTestsRunsCmd)
	workflowsCmd.AddCommand(workflowsTestsCmd)
}
