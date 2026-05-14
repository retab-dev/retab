package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsTestsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Manage block tests for a workflow",
	Long: `Declarative regression tests for workflow blocks.

A test pins a target block's expected output against a recorded input
(typically captured from a successful past run). Re-execute the test
suite after any change — config update, model swap, prompt edit — to
catch silent drift in extraction quality or classification accuracy.

A test has three pieces:

  * ` + "`target`" + `   — which block is under test
  * ` + "`source`" + `   — the input the block should consume (often a
    pointer to an existing run step)
  * ` + "`assertion`" + ` — the expected output, what we assert against

Use ` + "`workflows tests execute`" + ` to run a single test or the whole
suite. Inspect history via ` + "`workflows tests runs list`" + `.

For exploring multiple alternative block configurations side-by-side
(A/B-style), use ` + "`retab workflows experiments --help`" + ` instead.`,
	Example: `  # List a workflow's tests
  retab workflows tests list wf_abc123

  # Create a test pinning a block's expected output
  retab workflows tests create \
    --workflow-id wf_abc123 --name "Invoice 17 baseline" \
    --target-file ./target.json \
    --source-file ./source.json \
    --assertion-file ./assertion.json

  # Execute every test in the workflow
  retab workflows tests execute --workflow-id wf_abc123`,
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

var workflowsTestsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a block test",
	Long: `Create a regression test pinning a block's expected output.
You'll typically capture the three files from a successful past run:

  ` + "`--target-file`" + `     — which block to test (e.g.
  ` + `{"block_id": "blk_extract_1"}` + `).

  ` + "`--source-file`" + `     — the input. Usually a reference to a
  run/step that supplied known-good input.

  ` + "`--assertion-file`" + `  — the expected output JSON.

After creation, run with ` + "`workflows tests execute`" + `.`,
	Example: `  # Create a test from JSON files
  retab workflows tests create \
    --workflow-id wf_abc123 \
    --name "Invoice 17 baseline" \
    --target-file ./target.json \
    --source-file ./source.json \
    --assertion-file ./assertion.json`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.WorkflowTestCreateRequest{}
		req.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
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
		req.Target = retab.Resource(target)
		req.Source = retab.Resource(source)
		req.Assertion = retab.Resource(assertion)
		result, err := client.Workflows.Tests.Create(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsTestsGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <test-id>",
	Short: "Get a test",
	Long: `Fetch a test's definition: target block, source input, assertion
output, name, timestamps.`,
	Example: `  # Inspect a test
  retab workflows tests get wf_abc123 tst_jkl012`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
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
		req.Limit, _ = cmd.Flags().GetInt("limit")
		result, err := client.Workflows.Tests.List(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsTestsUpdateCmd = &cobra.Command{
	Use:   "update <workflow-id> <test-id>",
	Short: "Update a test",
	Long: `Re-pin a test's expected output or source input. Use this when
a deliberate schema or prompt change makes the old assertion stale — the
intent is to ratify the new output as the new baseline, not to silence
flaky runs.`,
	Example: `  # Refresh the assertion after a deliberate schema change
  retab workflows tests update wf_abc123 tst_jkl012 \
    --assertion-file ./new-assertion.json

  # Rename a test
  retab workflows tests update wf_abc123 tst_jkl012 \
    --name "Invoice 17 baseline (v2 schema)"`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
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
			req.Assertion = retab.Resource(assertion)
		}
		if source != nil {
			req.Source = retab.Resource(source)
		}
		result, err := client.Workflows.Tests.Update(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsTestsDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id> <test-id>",
	Short: "Delete a test",
	Long:  `Permanently delete a regression test and its run history.`,
	Example: `  # Drop a stale test
  retab workflows tests delete wf_abc123 tst_jkl012`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Tests.Delete(ctx, args[0], args[1]); err != nil {
			return err
		}
		confirmDeleted("test", args[1])
		return nil
	}),
}

var workflowsTestsExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute one or more block tests",
	Long: `Re-run regression tests and compare current output to the pinned
assertions. Without ` + "`--test-id`" + ` every test in the workflow runs;
with ` + "`--test-id`" + ` only that one runs.

Use ` + "`--n-consensus`" + ` to sample the target block N times and report
variance — useful for non-deterministic models.

Inspect results with ` + "`workflows tests runs list`" + ` and
` + "`workflows tests runs get`" + `.`,
	Example: `  # Run the whole regression suite
  retab workflows tests execute --workflow-id wf_abc123

  # Run just one test
  retab workflows tests execute \
    --workflow-id wf_abc123 --test-id tst_jkl012

  # Sample 5 times to measure stability
  retab workflows tests execute \
    --workflow-id wf_abc123 --test-id tst_jkl012 --n-consensus 5`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.ExecuteBlockTestsRequest{}
		req.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		req.TestID, _ = cmd.Flags().GetString("test-id")
		req.NConsensus, _ = cmd.Flags().GetInt("n-consensus")
		target, err := resolveJSONMap(cmd, "target-file")
		if err != nil {
			return err
		}
		if target != nil {
			req.Target = retab.Resource(target)
		}
		result, err := client.Workflows.Tests.Execute(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

// ---- test runs subgroup ----

var workflowsTestsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Inspect test runs",
	Long: `History of past test executions. Each ` + "`workflows tests execute`" + `
call creates one run; this subgroup retrieves the records.`,
	Example: `  # Recent runs for one test
  retab workflows tests runs list wf_abc123 tst_jkl012

  # Full record for one run
  retab workflows tests runs get wf_abc123 tst_jkl012 trun_mno345`,
}

var workflowsTestsRunsListCmd = &cobra.Command{
	Use:   "list <workflow-id> <test-id>",
	Short: "List runs for a test",
	Long: `List historical executions of one regression test. Defaults to
the 20 most recent.`,
	Example: `  # Recent runs for one test
  retab workflows tests runs list wf_abc123 tst_jkl012 --limit 50`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		limit, _ := cmd.Flags().GetInt("limit")
		result, err := client.Workflows.Tests.Runs.List(ctx, args[0], args[1], limit)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsTestsRunsGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <test-id> <run-id>",
	Short: "Get a test run",
	Long: `Fetch the full record for one test execution: actual output,
pass/fail per-assertion, diff against the pinned expectation, timing.`,
	Example: `  # Inspect a test run
  retab workflows tests runs get wf_abc123 tst_jkl012 trun_mno345`,
	Args: cobra.ExactArgs(3),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Tests.Runs.Get(ctx, args[0], args[1], args[2])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	workflowsTestsCreateCmd.Flags().String("workflow-id", "", "workflow id (required)")
	workflowsTestsCreateCmd.Flags().String("name", "", "test name")
	workflowsTestsCreateCmd.Flags().String("target-file", "", "JSON file with the target object (or - for stdin)")
	workflowsTestsCreateCmd.Flags().String("source-file", "", "JSON file with the source object (or - for stdin)")
	workflowsTestsCreateCmd.Flags().String("assertion-file", "", "JSON file with the assertion object (or - for stdin)")
	_ = workflowsTestsCreateCmd.MarkFlagRequired("workflow-id")

	workflowsTestsListCmd.Flags().String("target-block-id", "", "filter by target block id")
	workflowsTestsListCmd.Flags().Int("limit", 0, "max items (default 50)")

	workflowsTestsUpdateCmd.Flags().String("name", "", "new test name")
	workflowsTestsUpdateCmd.Flags().String("assertion-file", "", "JSON file with new assertion (or - for stdin)")
	workflowsTestsUpdateCmd.Flags().String("source-file", "", "JSON file with new source (or - for stdin)")

	workflowsTestsExecuteCmd.Flags().String("workflow-id", "", "workflow id (required)")
	workflowsTestsExecuteCmd.Flags().String("test-id", "", "single test to execute")
	workflowsTestsExecuteCmd.Flags().Int("n-consensus", 0, "consensus count")
	workflowsTestsExecuteCmd.Flags().String("target-file", "", "JSON file with target (or - for stdin)")
	_ = workflowsTestsExecuteCmd.MarkFlagRequired("workflow-id")

	workflowsTestsRunsListCmd.Flags().Int("limit", 0, "max items (default 20)")

	workflowsTestsRunsCmd.AddCommand(workflowsTestsRunsListCmd, workflowsTestsRunsGetCmd)
	workflowsTestsCmd.AddCommand(workflowsTestsCreateCmd, workflowsTestsGetCmd, workflowsTestsListCmd, workflowsTestsUpdateCmd, workflowsTestsDeleteCmd, workflowsTestsExecuteCmd, workflowsTestsRunsCmd)
	workflowsCmd.AddCommand(workflowsTestsCmd)
}
