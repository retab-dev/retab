package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Manage workflow runs",
	Long: `Execute, inspect, cancel, and replay workflow runs.

A run is one execution of a workflow against a set of inputs. Use this
subgroup to start runs (` + "`create`" + `), watch their lifecycle
(` + "`get`" + `, ` + "`steps list`" + `), restart failed runs
(` + "`restart`" + `), or submit human-in-the-loop decisions
(` + "`submit-hil`" + `).

Human-in-the-loop: when a run hits a ` + "`hil`" + ` block it pauses with
status ` + "`awaiting_review`" + `. Use ` + "`get-hil`" + ` to see what's
pending, then ` + "`submit-hil`" + ` to approve, reject, or send back
edited data — the run resumes from there.

For declarative regression testing of workflow outputs, see
` + "`retab workflows tests --help`" + `.`,
	Example: `  # Start a run by uploading a document into the start block
  retab workflows runs create wf_abc123 \
    --document-file start=./invoice.pdf

  # Inspect a run's lifecycle (status, errors, timing)
  retab workflows runs get run_xyz789

  # Stream per-block execution records
  retab workflows runs steps list run_xyz789

  # Cancel an in-flight run
  retab workflows runs cancel run_xyz789

  # Approve a paused HIL block
  retab workflows runs submit-hil run_xyz789 \
    --block-id blk_review_1 --approved`,
}

var workflowsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Start a workflow run",
	Long: `Start one run of a workflow against the supplied inputs.

Inputs are keyed by the id of the input block they target. There are
three ways to supply them, mix and match as needed:

  ` + "`--document-file BLOCK=PATH`" + ` — upload a local file. The MIME
  type is inferred from the file extension. Repeatable.

  ` + "`--document-url BLOCK=URL`" + ` — pass a remote URL the server will
  fetch. Repeatable.

  ` + "`--documents-file PATH`" + ` — a JSON object mapping block ids to
  pre-built MIME-data payloads. Use this for advanced cases or when the
  documents come from upstream pipelines.

Add ` + "`--json-inputs-file PATH`" + ` for blocks that accept structured
JSON instead of (or alongside) documents.

By default the workflow's latest published version runs; pin a specific
version with ` + "`--version`" + `. Inspect the resulting run with
` + "`workflows runs get`" + ` or ` + "`workflows runs steps list`" + `.`,
	Example: `  # Upload a local file into the start block
  retab workflows runs create wf_abc123 \
    --document-file start=./invoice.pdf

  # Multiple inputs across different blocks
  retab workflows runs create wf_abc123 \
    --document-file start=./invoice.pdf \
    --document-url reference=https://acme.com/po.pdf

  # JSON-only inputs (e.g. an api_call block)
  retab workflows runs create wf_abc123 \
    --json-inputs-file ./inputs.json

  # Pin to a specific published version
  retab workflows runs create wf_abc123 \
    --version v3 --document-file start=./invoice.pdf`,
	Args: cobra.ExactArgs(1),
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
	Long: `Fetch a run's metadata: status, trigger type, timestamps,
duration, cost, error info, final outputs. For per-block detail use
` + "`workflows runs steps list`" + `.`,
	Example: `  # Inspect a run
  retab workflows runs get run_xyz789

  # Poll until done
  while [ "$(retab workflows runs get run_xyz789 | jq -r '.status')" = "running" ]; do
    sleep 2
  done`,
	Args: cobra.ExactArgs(1),
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
	Long: `List runs across workflows, with filters for workflow id, status,
trigger type, date range, cost, and duration. Use cursor pagination
(` + "`--after`" + ` / ` + "`--before`" + ` / ` + "`--limit`" + `) and
` + "`--fields`" + ` to keep responses small on busy projects.`,
	Example: `  # Failed runs in a workflow over the last day
  retab workflows runs list \
    --workflow-id wf_abc123 \
    --status failed \
    --from-date 2026-05-13

  # Expensive runs only
  retab workflows runs list --workflow-id wf_abc123 --min-cost 1.0

  # Walk pages
  retab workflows runs list --workflow-id wf_abc123 --limit 50 --order desc`,
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
	Long: `Permanently delete a run and its step records. Artifacts
produced by the run are preserved separately (see
` + "`workflows artifacts`" + `).`,
	Example: `  # Delete a run
  retab workflows runs delete run_xyz789`,
	Args: cobra.ExactArgs(1),
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
	Long: `Cancel an in-flight run. Already-completed blocks keep their
outputs; in-progress blocks are terminated. Use ` + "`--command-id`" + ` to
make the cancel idempotent if you may retry the request.`,
	Example: `  # Cancel a stuck run
  retab workflows runs cancel run_xyz789

  # Idempotent cancel (safe to retry on network errors)
  retab workflows runs cancel run_xyz789 --command-id cancel-once-123`,
	Args: cobra.ExactArgs(1),
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
	Long: `Re-execute a failed or cancelled run, reusing the original
inputs. Useful after fixing a transient infra issue or tweaking block
config. Use ` + "`--command-id`" + ` for idempotency.`,
	Example: `  # Restart a failed run
  retab workflows runs restart run_xyz789`,
	Args: cobra.ExactArgs(1),
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
	Long: `Resume a run paused at a ` + "`hil`" + ` (human-in-the-loop) block.
Approve the proposed output, reject it, or submit edited data that
replaces the block's output before flow continues downstream.

To see what's pending, query ` + "`workflows runs get-hil`" + ` or
` + "`workflows runs get-agent-hil`" + ` (managed-agent reviews).`,
	Example: `  # Approve the block's proposed output as-is
  retab workflows runs submit-hil run_xyz789 \
    --block-id blk_review_1 --approved

  # Submit edited data — the file replaces the block's output
  retab workflows runs submit-hil run_xyz789 \
    --block-id blk_review_1 --approved \
    --modified-data-file ./fixed.json

  # Reject (run continues with rejection downstream)
  retab workflows runs submit-hil run_xyz789 \
    --block-id blk_review_1`,
	Args: cobra.ExactArgs(1),
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
	Long: `Inspect the pending human-in-the-loop review for a block:
proposed output, any reviewer-provided edits, decision timestamps. Pair
with ` + "`workflows runs submit-hil`" + ` to resolve it.`,
	Example: `  # Inspect the pending HIL state
  retab workflows runs get-hil run_xyz789 blk_review_1`,
	Args: cobra.ExactArgs(2),
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
	Long: `Fetch the managed-agent review record for a HIL block. When the
HIL is delegated to a managed agent (rather than a human), this returns
the agent's reasoning, score, and proposed decision.`,
	Example: `  # Inspect the managed-agent's review
  retab workflows runs get-agent-hil run_xyz789 blk_review_1`,
	Args: cobra.ExactArgs(2),
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
	Long: `Return the frozen workflow config (blocks, edges, per-block
config) as it was at the moment the run started. Useful for forensics
when a workflow has changed since the run executed and you want to
understand what produced a given output.`,
	Example: `  # Audit the config that produced an old run
  retab workflows runs config run_xyz789`,
	Args: cobra.ExactArgs(1),
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
	Long: `Return the topological order in which the run's blocks were (or
will be) executed. Helpful for understanding scheduling when a run
contains branches, parallel paths, or conditionals.`,
	Example: `  # See the order blocks ran in
  retab workflows runs execution-order run_xyz789`,
	Args: cobra.ExactArgs(1),
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
	Long: `Return a time-limited URL to the document a block consumed (or
produced) during a run. Useful for downloading the exact bytes the
workflow saw.`,
	Example: `  # Fetch the original input document
  retab workflows runs document-url run_xyz789 blk_start_1`,
	Args: cobra.ExactArgs(2),
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
	Long: `Bulk-export the outputs of many runs of a workflow as CSV,
focused on one block's output. Filter by status, date range, trigger
type, or an explicit set of run ids. Use ` + "`--preferred-column`" + ` to
pin column order.`,
	Example: `  # Export every successful run of a block to CSV
  retab workflows runs export \
    --workflow-id wf_abc123 --block-id blk_extract_1 \
    --status succeeded \
    --from-date 2026-05-01

  # Export a specific set of runs with custom columns
  retab workflows runs export \
    --workflow-id wf_abc123 --block-id blk_extract_1 \
    --run-id run_xyz789 --run-id run_aaa000 \
    --preferred-column invoice_id --preferred-column total`,
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
	Long: `A step is the execution record of one block within a run —
input it saw, output it produced, error if any, timing, cost. Steps are
the entry point for debugging: when a run output looks wrong, list its
steps and pull the offending one with ` + "`workflows runs steps get`" + `
to see exactly what the block did.`,
	Example: `  # List every step in a run
  retab workflows runs steps list run_xyz789

  # Pull the full execution record for one block
  retab workflows runs steps get run_xyz789 blk_extract_1`,
}

var workflowsRunsStepsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List run steps",
	Long: `List every step in a run — one record per block execution.
Includes status, timing, and a summary of input/output sizes. For the
full input/output payload of one step use ` + "`steps get`" + `.`,
	Example: `  # List steps
  retab workflows runs steps list run_xyz789

  # Find the first failed step
  retab workflows runs steps list run_xyz789 \
    | jq '.[] | select(.status == "failed") | .block_id' | head -1`,
	Args: cobra.ExactArgs(1),
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
	Long: `Return everything about one block's execution in a run: the
exact input payload, the produced output, any error, timing, cost, and
model usage if applicable.

This is the canonical entry point when debugging a run that produced
the wrong output — find the offending block, inspect its inputs, and
correlate against the block's config.`,
	Example: `  # Pull the full record for a single step
  retab workflows runs steps get run_xyz789 blk_extract_1

  # Save the input payload for offline replay
  retab workflows runs steps get run_xyz789 blk_extract_1 \
    | jq '.input' > input.json`,
	Args: cobra.ExactArgs(2),
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
