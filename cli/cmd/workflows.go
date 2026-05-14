package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Manage workflows",
	Long: `Build and run multi-step document-processing pipelines.

A workflow is a DAG of blocks (` + "`extract`" + `, ` + "`split`" + `,
` + "`classify`" + `, ` + "`edit`" + `, ` + "`hil`" + `, ` + "`conditional`" + `,
` + "`api_call`" + `, ` + "`function`" + `, …) wired together by edges. Documents
or JSON inputs flow through the graph; each block contributes to the final
output. Workflows are versioned — drafts are mutable, published versions are
immutable.

Typical lifecycle:

  1. ` + "`workflows create`" + ` — scaffold a draft
  2. ` + "`workflows blocks create`" + ` / ` + "`workflows edges create`" + ` — wire the graph
  3. ` + "`workflows blocks update`" + ` — configure each block
  4. ` + "`workflows runs create`" + ` — execute against inputs
  5. ` + "`workflows runs steps list`" + ` — inspect per-block output
  6. ` + "`workflows tests create`" + ` — pin the expected output for regression`,
	Example: `  # List your workflows
  retab workflows list

  # Inspect a workflow's graph (blocks + edges in one call)
  retab workflows entities wf_abc123

  # Run a workflow against an uploaded file
  retab workflows runs create wf_abc123 \
    --document-file start=./invoice.pdf

  # Publish the current draft as an immutable version
  retab workflows publish wf_abc123 --description "v1: invoice extraction"`,
}

var workflowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflows",
	Long: `List workflows in your project, with cursor pagination and field
selection. Use ` + "`--fields`" + ` to trim large list payloads, and
` + "`--after`" + ` / ` + "`--before`" + ` to walk through pages.`,
	Example: `  # First page
  retab workflows list --limit 20

  # Next page (cursor returned in the previous response)
  retab workflows list --after wf_abc123 --limit 20

  # Project just the fields you need
  retab workflows list --fields id,name,updated_at`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowsParams{}
		params.Before, _ = cmd.Flags().GetString("before")
		params.After, _ = cmd.Flags().GetString("after")
		params.Limit, _ = cmd.Flags().GetInt("limit")
		params.Order, _ = cmd.Flags().GetString("order")
		params.SortBy, _ = cmd.Flags().GetString("sort-by")
		params.Fields, _ = cmd.Flags().GetString("fields")
		result, err := client.Workflows.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsGetCmd = &cobra.Command{
	Use:   "get <workflow-id>",
	Short: "Get a workflow",
	Long: `Fetch the workflow envelope: name, description, current draft and
published versions, email-trigger settings, timestamps.

For the resolved graph (blocks + edges), use ` + "`workflows entities`" + `.
For per-block I/O schemas, use ` + "`workflows resolved-schemas`" + `.`,
	Example: `  # Inspect a workflow
  retab workflows get wf_abc123

  # Pretty-print just the published version
  retab workflows get wf_abc123 | jq '.published_version'`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a workflow",
	Long: `Scaffold a new draft workflow. The returned workflow has an empty
graph — add blocks (` + "`workflows blocks create`" + `) and wire them with
edges (` + "`workflows edges create`" + `) to make it functional.`,
	Example: `  # Create a draft workflow
  retab workflows create --name "Invoice extraction" \
    --description "Parse vendor invoices into structured JSON"

  # Capture the returned id for piping into subsequent calls
  WF=$(retab workflows create --name "Quick demo" | jq -r '.id')`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		result, err := client.Workflows.Create(ctx, retab.CreateWorkflowRequest{
			Name:        name,
			Description: description,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsUpdateCmd = &cobra.Command{
	Use:   "update <workflow-id>",
	Short: "Update a workflow",
	Long: `Patch the workflow envelope. Use this to rename, re-describe, or
configure the email trigger (allowlist of sender addresses or domains that
can drop a document into the workflow by email).

This does NOT modify the graph — for that see ` + "`workflows blocks`" + `
and ` + "`workflows edges`" + `.`,
	Example: `  # Rename
  retab workflows update wf_abc123 --name "Invoice extraction v2"

  # Configure the email trigger allowlist
  retab workflows update wf_abc123 \
    --allowed-domain acme.com \
    --allowed-sender ops@vendor.io`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		var req retab.UpdateWorkflowRequest
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			req.Name = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
		}
		senders, _ := cmd.Flags().GetStringArray("allowed-sender")
		domains, _ := cmd.Flags().GetStringArray("allowed-domain")
		if cmd.Flags().Changed("allowed-sender") || cmd.Flags().Changed("allowed-domain") {
			req.EmailTrigger = &retab.WorkflowEmailTrigger{
				AllowedSenders: senders,
				AllowedDomains: domains,
			}
		}
		result, err := client.Workflows.Update(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id>",
	Short: "Delete a workflow",
	Long: `Permanently delete a workflow, including its draft graph,
published versions, and run history. Artifacts produced by past runs are
preserved as separate objects (see ` + "`workflows artifacts`" + `).`,
	Example: `  # Delete a workflow
  retab workflows delete wf_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("workflow", args[0])
		return nil
	}),
}

var workflowsPublishCmd = &cobra.Command{
	Use:   "publish <workflow-id>",
	Short: "Publish the current draft",
	Long: `Snapshot the current draft as an immutable published version.
Subsequent ` + "`workflows runs create`" + ` calls without an explicit
` + "`--version`" + ` use the latest published version.

The draft remains editable after publishing; iterate freely, then publish
again to cut a new version.`,
	Example: `  # Publish with a release note
  retab workflows publish wf_abc123 \
    --description "v3: tighter line-item schema"

  # Pin a run to a specific published version
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
		description, _ := cmd.Flags().GetString("description")
		result, err := client.Workflows.Publish(ctx, args[0], retab.PublishWorkflowRequest{Description: description})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsDuplicateCmd = &cobra.Command{
	Use:   "duplicate <workflow-id>",
	Short: "Duplicate a workflow",
	Long: `Deep-copy a workflow's draft graph, blocks, and edges into a new
workflow with a fresh id. Useful for forking a production workflow to
experiment without risk.

Published versions and run history are NOT carried over.`,
	Example: `  # Fork a workflow for tweaking
  retab workflows duplicate wf_abc123

  # Capture the new id
  NEW=$(retab workflows duplicate wf_abc123 | jq -r '.id')`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Duplicate(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEntitiesCmd = &cobra.Command{
	Use:   "entities <workflow-id>",
	Short: "Get the workflow with its blocks and edges",
	Long: `Return the workflow envelope plus the full graph in one call:
blocks, edges, and connections. Use this when you need the complete
picture in a single request — for inspection, exporting, or feeding into
` + "`workflows diagnose --graph-file`" + `.`,
	Example: `  # Dump the full graph
  retab workflows entities wf_abc123

  # Save to disk for offline editing or diffs
  retab workflows entities wf_abc123 > wf_abc123.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.GetEntities(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsResolvedSchemasCmd = &cobra.Command{
	Use:   "resolved-schemas <workflow-id>",
	Short: "Get the resolved input/output schemas for every block",
	Long: `Compute the resolved input and output JSON schemas for every
block in the draft graph, after applying type propagation across edges.
Useful for confirming a block sees the shape you expect before running.

For a single block, prefer ` + "`workflows blocks resolved-schemas`" + `.`,
	Example: `  # Inspect schemas across the whole graph
  retab workflows resolved-schemas wf_abc123

  # Grep for a specific block's output
  retab workflows resolved-schemas wf_abc123 \
    | jq '.blocks["blk_extract_1"].output_schema'`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.GetResolvedSchemas(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsDiagnoseCmd = &cobra.Command{
	Use:   "diagnose <workflow-id>",
	Short: "Diagnose the persisted draft graph (use --graph-file to send an in-memory graph)",
	Long: `Validate a workflow's draft graph and report structural issues:
disconnected blocks, type mismatches across edges, missing required
config, unreachable paths.

By default diagnoses the persisted draft. Pass ` + "`--graph-file`" + ` to
diagnose an in-memory graph (e.g. before persisting changes) — the file
must be a JSON object with ` + "`blocks`" + `, ` + "`edges`" + `, and optional
` + "`re_propagate`" + ` fields, in the same shape as
` + "`workflows entities`" + ` output.`,
	Example: `  # Diagnose the persisted draft
  retab workflows diagnose wf_abc123

  # Diagnose a proposed graph before persisting
  retab workflows entities wf_abc123 > graph.json
  # edit graph.json
  retab workflows diagnose wf_abc123 --graph-file ./graph.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		graphFile, _ := cmd.Flags().GetString("graph-file")
		if graphFile != "" {
			body, err := readJSONMap(graphFile)
			if err != nil {
				return err
			}
			req := retab.DiagnoseWorkflowGraphRequest{RePropagate: true}
			if blocks, ok := body["blocks"].([]any); ok {
				for _, b := range blocks {
					if obj, ok := b.(map[string]any); ok {
						req.Blocks = append(req.Blocks, obj)
					}
				}
			}
			if edges, ok := body["edges"].([]any); ok {
				for _, e := range edges {
					if obj, ok := e.(map[string]any); ok {
						req.Edges = append(req.Edges, obj)
					}
				}
			}
			if v, ok := body["re_propagate"].(bool); ok {
				req.RePropagate = v
			}
			result, err := client.Workflows.DiagnoseGraph(ctx, args[0], req)
			if err != nil {
				return err
			}
			return printJSON(result)
		}
		result, err := client.Workflows.Diagnose(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	workflowsListCmd.Flags().String("before", "", "cursor: items before this id")
	workflowsListCmd.Flags().String("after", "", "cursor: items after this id")
	workflowsListCmd.Flags().Int("limit", 0, "max items to return")
	workflowsListCmd.Flags().String("order", "", "asc | desc")
	workflowsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsListCmd.Flags().String("fields", "", "comma-separated field list to return")

	workflowsCreateCmd.Flags().String("name", "", "workflow name")
	workflowsCreateCmd.Flags().String("description", "", "workflow description")

	workflowsUpdateCmd.Flags().String("name", "", "new name")
	workflowsUpdateCmd.Flags().String("description", "", "new description")
	workflowsUpdateCmd.Flags().StringArray("allowed-sender", nil, "email trigger allowed sender (repeatable)")
	workflowsUpdateCmd.Flags().StringArray("allowed-domain", nil, "email trigger allowed domain (repeatable)")

	workflowsPublishCmd.Flags().String("description", "", "publish description")

	workflowsDiagnoseCmd.Flags().String("graph-file", "", "JSON file with {blocks, edges, re_propagate} to diagnose without persisting")

	workflowsCmd.AddCommand(workflowsListCmd, workflowsGetCmd, workflowsCreateCmd, workflowsUpdateCmd, workflowsDeleteCmd, workflowsPublishCmd, workflowsDuplicateCmd, workflowsEntitiesCmd, workflowsResolvedSchemasCmd, workflowsDiagnoseCmd)
	rootCmd.AddCommand(workflowsCmd)
}
