package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsTestsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Manage block tests for a workflow",
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
	Args:  cobra.ExactArgs(2),
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
	Args:  cobra.ExactArgs(1),
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
	Args:  cobra.ExactArgs(2),
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
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Workflows.Tests.Delete(ctx, args[0], args[1])
	}),
}

var workflowsTestsExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute one or more block tests",
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
}

var workflowsTestsRunsListCmd = &cobra.Command{
	Use:   "list <workflow-id> <test-id>",
	Short: "List runs for a test",
	Args:  cobra.ExactArgs(2),
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
	Args:  cobra.ExactArgs(3),
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
