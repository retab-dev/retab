package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Manage workflow runs",
}

var workflowsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Start a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.CreateWorkflowRunRequest{WorkflowID: args[0]}
		req.Version, _ = cmd.Flags().GetString("version")
		docsFile, _ := cmd.Flags().GetString("documents-file")
		if docsFile != "" {
			docs, err := readJSONMap(docsFile)
			if err != nil {
				return fmt.Errorf("--documents-file: %w", err)
			}
			req.Documents = docs
		}
		fileFlags, _ := cmd.Flags().GetStringArray("document-file")
		urlFlags, _ := cmd.Flags().GetStringArray("document-url")
		if len(fileFlags) > 0 || len(urlFlags) > 0 {
			if req.Documents == nil {
				req.Documents = map[string]any{}
			}
			for _, raw := range fileFlags {
				key, path, ok := splitKV(raw)
				if !ok || path == "" {
					return fmt.Errorf("--document-file expects block-id=path, got %q", raw)
				}
				mime, err := retab.InferMIMEData(path)
				if err != nil {
					return fmt.Errorf("--document-file %s: %w", raw, err)
				}
				req.Documents[key] = mime
			}
			for _, raw := range urlFlags {
				key, url, ok := splitKV(raw)
				if !ok || url == "" {
					return fmt.Errorf("--document-url expects block-id=url, got %q", raw)
				}
				req.Documents[key] = retab.MIMEData{URL: url}
			}
		}
		jsonInputsFile, _ := cmd.Flags().GetString("json-inputs-file")
		if jsonInputsFile != "" {
			inputs, err := readJSONMap(jsonInputsFile)
			if err != nil {
				return fmt.Errorf("--json-inputs-file: %w", err)
			}
			req.JSONInputs = inputs
		}
		result, err := client.Workflows.Runs.Create(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflow runs",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowRunsParams{}
		params.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		params.Status, _ = cmd.Flags().GetString("status")
		params.Statuses, _ = cmd.Flags().GetStringArray("statuses")
		params.ExcludeStatus, _ = cmd.Flags().GetString("exclude-status")
		params.TriggerType, _ = cmd.Flags().GetString("trigger-type")
		params.TriggerTypes, _ = cmd.Flags().GetStringArray("trigger-types")
		params.FromDate, _ = cmd.Flags().GetString("from-date")
		params.ToDate, _ = cmd.Flags().GetString("to-date")
		params.Search, _ = cmd.Flags().GetString("search")
		params.SortBy, _ = cmd.Flags().GetString("sort-by")
		params.Fields, _ = cmd.Flags().GetStringArray("fields")
		params.Before, _ = cmd.Flags().GetString("before")
		params.After, _ = cmd.Flags().GetString("after")
		params.Limit, _ = cmd.Flags().GetInt("limit")
		params.Order, _ = cmd.Flags().GetString("order")
		if cmd.Flags().Changed("min-cost") {
			v, _ := cmd.Flags().GetFloat64("min-cost")
			params.MinCost = &v
		}
		if cmd.Flags().Changed("max-cost") {
			v, _ := cmd.Flags().GetFloat64("max-cost")
			params.MaxCost = &v
		}
		if cmd.Flags().Changed("min-duration") {
			v, _ := cmd.Flags().GetInt("min-duration")
			params.MinDuration = &v
		}
		if cmd.Flags().Changed("max-duration") {
			v, _ := cmd.Flags().GetInt("max-duration")
			params.MaxDuration = &v
		}
		result, err := client.Workflows.Runs.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsDeleteCmd = &cobra.Command{
	Use:   "delete <run-id>",
	Short: "Delete a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Workflows.Runs.Delete(ctx, args[0])
	}),
}

var workflowsRunsCancelCmd = &cobra.Command{
	Use:   "cancel <run-id>",
	Short: "Cancel a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		commandID, _ := cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Runs.Cancel(ctx, args[0], retab.WorkflowRunCommandRequest{CommandID: commandID})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsRestartCmd = &cobra.Command{
	Use:   "restart <run-id>",
	Short: "Restart a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		commandID, _ := cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Runs.Restart(ctx, args[0], retab.WorkflowRunCommandRequest{CommandID: commandID})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsSubmitHILCmd = &cobra.Command{
	Use:   "submit-hil <run-id>",
	Short: "Submit a human-in-the-loop decision",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.SubmitHILDecisionRequest{}
		req.BlockID, _ = cmd.Flags().GetString("block-id")
		req.Approved, _ = cmd.Flags().GetBool("approved")
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		if path, _ := cmd.Flags().GetString("modified-data-file"); path != "" {
			data, err := readJSONMap(path)
			if err != nil {
				return fmt.Errorf("--modified-data-file: %w", err)
			}
			req.ModifiedData = data
		}
		result, err := client.Workflows.Runs.SubmitHILDecision(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsGetHILCmd = &cobra.Command{
	Use:   "get-hil <run-id> <block-id>",
	Short: "Get HIL decision state for a block",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.GetHILDecision(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsGetAgentHILCmd = &cobra.Command{
	Use:   "get-agent-hil <run-id> <block-id>",
	Short: "Get the managed-agent HIL review for a block",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.GetAgentHILReview(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsConfigCmd = &cobra.Command{
	Use:   "config <run-id>",
	Short: "Get the workflow config used by a run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.GetConfig(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsExecutionOrderCmd = &cobra.Command{
	Use:   "execution-order <run-id>",
	Short: "Get the execution order for a run",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.ExecutionOrder(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsDocumentURLCmd = &cobra.Command{
	Use:   "document-url <run-id> <block-id>",
	Short: "Get the document URL stored at a run/block",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.GetDocumentURL(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export workflow runs as CSV",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.ExportWorkflowRunsRequest{}
		req.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		req.BlockID, _ = cmd.Flags().GetString("block-id")
		req.ExportSource, _ = cmd.Flags().GetString("export-source")
		req.SelectedRunIDs, _ = cmd.Flags().GetStringArray("run-id")
		req.Status, _ = cmd.Flags().GetString("status")
		req.ExcludeStatus, _ = cmd.Flags().GetString("exclude-status")
		req.FromDate, _ = cmd.Flags().GetString("from-date")
		req.ToDate, _ = cmd.Flags().GetString("to-date")
		req.TriggerTypes, _ = cmd.Flags().GetStringArray("trigger-types")
		req.PreferredColumns, _ = cmd.Flags().GetStringArray("preferred-column")
		result, err := client.Workflows.Runs.Export(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

// ---- steps subgroup ----

var workflowsRunsStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "Inspect workflow run steps",
}

var workflowsRunsStepsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List run steps",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.Steps.List(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsStepsGetCmd = &cobra.Command{
	Use:   "get <run-id> <block-id>",
	Short: "Get the full step execution record",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.Steps.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	workflowsRunsCreateCmd.Flags().String("version", "", "workflow version (defaults to production)")
	workflowsRunsCreateCmd.Flags().String("documents-file", "", "JSON object {block-id: document} (or - for stdin)")
	workflowsRunsCreateCmd.Flags().StringArray("document-file", nil, "document file as block-id=path (repeatable)")
	workflowsRunsCreateCmd.Flags().StringArray("document-url", nil, "document url as block-id=url (repeatable)")
	workflowsRunsCreateCmd.Flags().String("json-inputs-file", "", "JSON inputs object (or - for stdin)")

	workflowsRunsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsRunsListCmd.Flags().String("status", "", "filter by status")
	workflowsRunsListCmd.Flags().StringArray("statuses", nil, "filter by status (repeatable)")
	workflowsRunsListCmd.Flags().String("exclude-status", "", "exclude status")
	workflowsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsRunsListCmd.Flags().StringArray("trigger-types", nil, "filter by trigger types (repeatable)")
	workflowsRunsListCmd.Flags().String("from-date", "", "filter from this date")
	workflowsRunsListCmd.Flags().String("to-date", "", "filter to this date")
	workflowsRunsListCmd.Flags().String("search", "", "search query")
	workflowsRunsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsRunsListCmd.Flags().StringArray("fields", nil, "include only these fields (repeatable)")
	workflowsRunsListCmd.Flags().String("before", "", "cursor: items before this id")
	workflowsRunsListCmd.Flags().String("after", "", "cursor: items after this id")
	workflowsRunsListCmd.Flags().Int("limit", 0, "max items to return")
	workflowsRunsListCmd.Flags().String("order", "", "asc | desc")
	workflowsRunsListCmd.Flags().Float64("min-cost", 0, "min cost")
	workflowsRunsListCmd.Flags().Float64("max-cost", 0, "max cost")
	workflowsRunsListCmd.Flags().Int("min-duration", 0, "min duration (ms)")
	workflowsRunsListCmd.Flags().Int("max-duration", 0, "max duration (ms)")

	workflowsRunsCancelCmd.Flags().String("command-id", "", "idempotency command id")
	workflowsRunsRestartCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsRunsSubmitHILCmd.Flags().String("block-id", "", "block id (required)")
	workflowsRunsSubmitHILCmd.Flags().Bool("approved", false, "approve the block output")
	workflowsRunsSubmitHILCmd.Flags().String("modified-data-file", "", "JSON file with modified data (or - for stdin)")
	workflowsRunsSubmitHILCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsRunsSubmitHILCmd.MarkFlagRequired("block-id")

	workflowsRunsExportCmd.Flags().String("workflow-id", "", "workflow id (required)")
	workflowsRunsExportCmd.Flags().String("block-id", "", "block id (required)")
	workflowsRunsExportCmd.Flags().String("export-source", "", "export source (default outputs)")
	workflowsRunsExportCmd.Flags().StringArray("run-id", nil, "filter to selected run ids (repeatable)")
	workflowsRunsExportCmd.Flags().String("status", "", "status filter")
	workflowsRunsExportCmd.Flags().String("exclude-status", "", "exclude status")
	workflowsRunsExportCmd.Flags().String("from-date", "", "from date")
	workflowsRunsExportCmd.Flags().String("to-date", "", "to date")
	workflowsRunsExportCmd.Flags().StringArray("trigger-types", nil, "trigger types (repeatable)")
	workflowsRunsExportCmd.Flags().StringArray("preferred-column", nil, "preferred CSV column (repeatable)")
	_ = workflowsRunsExportCmd.MarkFlagRequired("workflow-id")
	_ = workflowsRunsExportCmd.MarkFlagRequired("block-id")

	workflowsRunsStepsCmd.AddCommand(workflowsRunsStepsListCmd, workflowsRunsStepsGetCmd)
	workflowsRunsCmd.AddCommand(workflowsRunsCreateCmd, workflowsRunsGetCmd, workflowsRunsListCmd, workflowsRunsDeleteCmd, workflowsRunsCancelCmd, workflowsRunsRestartCmd, workflowsRunsSubmitHILCmd, workflowsRunsGetHILCmd, workflowsRunsGetAgentHILCmd, workflowsRunsConfigCmd, workflowsRunsExecutionOrderCmd, workflowsRunsDocumentURLCmd, workflowsRunsExportCmd, workflowsRunsStepsCmd)
	workflowsCmd.AddCommand(workflowsRunsCmd)
}
