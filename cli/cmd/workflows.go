//go:build !retab_oagen_cli_workflows

package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// warnIfEmptyWorkflowOnPublish prints a stderr warning when the workflow has
// 0 or 1 blocks and the single block (if any) is the auto-added `start_document`
// placeholder. A freshly-`workflows create`d draft is exactly that shape —
// publishing it produces a version that does nothing. The user might still
// want to (CI stub, scaffolding), so the warning never blocks; --force on
// the publish command suppresses it entirely.
//
// Failures fetching the block list are deliberately swallowed: this is a
// best-effort UX nicety, not a precondition. A network blip here must not
// keep the user from publishing — the publish call itself will surface any
// real auth/server failure with its own error.
func warnIfEmptyWorkflowOnPublish(ctx context.Context, client *retab.Client, workflowID string, w io.Writer) {
	blocks, err := client.Workflows.Blocks.List(ctx, &retab.WorkflowBlocksListParams{WorkflowID: workflowID})
	if err != nil {
		return
	}
	if !isEffectivelyEmptyDraft(blocks.Data) {
		return
	}
	_, _ = fmt.Fprintln(w, "warning: workflow has only a start_document block — publishing an empty workflow.")
	_, _ = fmt.Fprintln(w, "warning: add blocks with `retab workflows blocks create` before publishing.")
}

// isEffectivelyEmptyDraft returns true for the two empty-ish shapes that
// produce a no-op published version: a fully empty block list (shouldn't
// happen via the UI but is possible via the API), and the canonical
// freshly-created-workflow shape of exactly one `start_document` block.
func isEffectivelyEmptyDraft(blocks []retab.WorkflowBlock) bool {
	switch len(blocks) {
	case 0:
		return true
	case 1:
		return isStartDocumentBlock(blocks[0])
	default:
		return false
	}
}

func isStartDocumentBlock(block retab.WorkflowBlock) bool {
	return block.Type == "start_document"
}

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Manage workflows",
	Long: `Build and run multi-step document-processing pipelines.

A workflow is a DAG of blocks (` + "`extract`" + `, ` + "`split`" + `,
` + "`classifier`" + `, ` + "`edit`" + `, ` + "`conditional`" + `,
` + "`api_call`" + `, ` + "`function`" + `, …) wired together by edges. Documents
or JSON inputs flow through the graph; each block contributes to the final
output. Add ` + "`config.review`" + ` to reviewable blocks when a run should
pause for review. Workflows are versioned — drafts are mutable, published
versions are immutable.

Review is configured on the block (` + "`config.review`" + `), not as a
standalone block. A reviewed run pauses with status
` + "`awaiting_review`" + ` and is resumed through
` + "`retab workflows reviews approve --version-id ...`" + ` or failed through
` + "`retab workflows reviews reject --version-id ... --reason ...`" + `.

Typical lifecycle:

  1. ` + "`workflows create`" + ` — scaffold a draft
  2. ` + "`workflows blocks create`" + ` / ` + "`workflows edges create`" + ` — wire the graph
  3. ` + "`workflows blocks update`" + ` — configure each block
  4. ` + "`workflows runs create`" + ` — execute against inputs
  5. ` + "`workflows steps list`" + ` — inspect per-block output
  6. ` + "`workflows tests create`" + ` — pin the expected output for regression`,
	Example: `  # List your workflows
  retab workflows list

  # Inspect a workflow's graph
  retab workflows view wf_abc123

  # Run a workflow against an uploaded file
  retab workflows runs create wf_abc123 \
    --document start=./invoice.pdf

  # See runs paused for review
  retab workflows runs list --status awaiting_review

  # Publish the current draft as an immutable version
  retab workflows publish wf_abc123 --description "v1: invoice extraction"`,
}

var workflowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflows",
	Long: `List workflows in your project, with id pagination.
Use ` + "`--after`" + ` / ` + "`--before`" + ` to walk through pages.`,
	Example: `  # First page
  retab workflows list --limit 20

  # Next page (use the last workflow id from the previous response)
  retab workflows list --after wf_abc123 --limit 20`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		before, _ := cmd.Flags().GetString("before")
		after, _ := cmd.Flags().GetString("after")
		limit, _ := cmd.Flags().GetInt("limit")
		order, _ := cmd.Flags().GetString("order")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		projectID, _ := cmd.Flags().GetString("project-id")

		params := &retab.WorkflowsListParams{}
		if before != "" {
			params.Before = ptr(before)
		}
		if after != "" {
			params.After = ptr(after)
		}
		if limit > 0 {
			params.Limit = ptr(limit)
		}
		if order != "" {
			params.Order = ptr(order)
		}
		if sortBy != "" {
			params.SortBy = ptr(sortBy)
		}
		if projectID != "" {
			params.ProjectID = ptr(projectID)
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.List(ctx, params)
		if err != nil {
			return err
		}
		return printWorkflowList(cmd, result)
	}),
}

func printWorkflowList(cmd *cobra.Command, result any) error {
	if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputTable) {
		return printResultTable(result)
	}
	return printJSON(result)
}

var workflowsGetCmd = &cobra.Command{
	Use:   "get <workflow-id>",
	Short: "Get a workflow",
	Long: `Fetch the workflow envelope: name, description, current draft and
published versions, timestamps.

For a graph-shaped view, use ` + "`workflows view`" + `.
For blocks and edges as JSON, use ` + "`workflows blocks list`" + ` and ` + "`workflows edges list`" + `.`,
	Example: `  # Inspect a workflow
  retab workflows get wf_abc123

  # Pretty-print just the published version
  retab workflows get wf_abc123 | jq '.published.version_id'`,
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
		return printResult(cmd, result)
	}),
}

var workflowsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a workflow",
	Long: `Scaffold a new draft workflow. Fresh drafts include the default
start_document block; add processing blocks (` + "`workflows blocks create`" + `) and
wire them with edges (` + "`workflows edges create`" + `) to make the workflow
functional.`,
	Example: `  # Create a draft workflow (a workflow must belong to a project)
  retab workflows create --name "Invoice extraction" \
    --project-id proj_abc123 \
    --description "Parse vendor invoices into structured JSON"

  # Capture the returned id for piping into subsequent calls
  WF=$(retab workflows create --name "Quick demo" --project-id proj_abc123 | jq -r '.id')`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		name, _ := cmd.Flags().GetString("name")
		if cmd.Flags().Changed("name") {
			trimmed, err := validateWorkflowName(name)
			if err != nil {
				return err
			}
			name = trimmed
		}
		description, _ := cmd.Flags().GetString("description")
		projectID, _ := cmd.Flags().GetString("project-id")
		if projectID == "" {
			return fmt.Errorf("--project-id is required: a workflow must belong to a project")
		}
		params := retab.WorkflowsCreateParams{ProjectID: projectID}
		if name != "" || cmd.Flags().Changed("name") {
			params.Name = ptr(name)
		}
		if description != "" || cmd.Flags().Changed("description") {
			params.Description = ptr(description)
		}
		result, err := client.Workflows.Create(ctx, &params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsUpdateCmd = &cobra.Command{
	Use:   "update <workflow-id>",
	Short: "Update a workflow",
	Long: `Patch the workflow envelope. Use this to rename or re-describe the
workflow.

This does NOT modify the graph — for that see ` + "`workflows blocks`" + `
and ` + "`workflows edges`" + `.`,
	Example: `  # Rename
  retab workflows update wf_abc123 --name "Invoice extraction v2"`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") {
			return fmt.Errorf("nothing to update: pass at least one of --name or --description")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		body := map[string]any{}
		var req retab.WorkflowsUpdateParams
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			trimmed, err := validateWorkflowName(v)
			if err != nil {
				return err
			}
			req.Name = &trimmed
			body["name"] = trimmed
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
			body["description"] = v
		}
		result, err := client.Workflows.Update(ctx, args[0], &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

// validateWorkflowName trims surrounding whitespace from the user-supplied
// workflow name, rejects values that are blank after trimming, and returns
// the cleaned value. Callers MUST use the returned string when building the
// outgoing request body — otherwise “workflows create --name "  padded  "“
// silently stores the padded version on the server (bug #3).
func validateWorkflowName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("workflow name must not be blank")
	}
	return trimmed, nil
}

var workflowsDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id>",
	Short: "Delete a workflow",
	Long: `Permanently delete a workflow, including its draft graph,
published versions, and run history. Artifacts produced by past runs are
preserved as separate objects (see ` + "`workflows artifacts`" + `).

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal.`,
	Example: `  # Delete a workflow (interactive, asks to confirm)
  retab workflows delete wf_abc123

  # Skip the prompt in scripts
  retab workflows delete wf_abc123 --yes
  # => { "id": "wf_abc123", "deleted": true }`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "workflow", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Delete(ctx, args[0]); err != nil {
			return err
		}
		return printResult(cmd, map[string]any{"id": args[0], "deleted": true})
	}),
}

var workflowsPublishCmd = &cobra.Command{
	Use:   "publish <workflow-id>",
	Short: "Publish the current draft",
	Long: `Snapshot the current draft as an immutable published version.
Subsequent ` + "`workflows runs create`" + ` calls without an explicit
` + "`--version`" + ` use the latest published version.

The draft remains editable after publishing; iterate freely, then publish
again to cut a new version.

Before publishing, the CLI warns when the draft appears empty (a draft with only the auto-added
` + "`start_document`" + ` block produces a no-op published version);
` + "`--force`" + ` suppresses that warning.`,
	Example: `  # Publish with a release note
  retab workflows publish wf_abc123 \
    --description "v3: tighter line-item schema"

  # Pin a run to a specific published version (production, draft, or a ver_... id)
  retab workflows runs create wf_abc123 \
    --version ver_xxx --document start=./invoice.pdf

  # Skip the empty-workflow warning
  retab workflows publish wf_abc123 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		description, _ := cmd.Flags().GetString("description")
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			warnIfEmptyWorkflowOnPublish(ctx, client, args[0], cmd.ErrOrStderr())
		}
		result, err := client.Workflows.Publish(ctx, args[0], &retab.WorkflowsPublishParams{
			Body: retab.PublishWorkflowRequest{Description: ptr(description)},
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsDiscardDraftCmd = &cobra.Command{
	Use:   "discard-draft <workflow-id>",
	Short: "Discard draft changes",
	Long: `Discard all unpublished draft changes and restore the workflow to its
last published state. The draft blocks and edges are recreated from the
published version, so any in-progress edits are lost.

This requires the workflow to be published (it must have a published
version). It is destructive to draft edits. Pass ` + "`--yes`" + ` to skip the
confirmation prompt in scripts and CI — otherwise the command refuses to
discard when stdin is not a terminal.`,
	Example: `  # Revert the draft to the published version (interactive, asks to confirm)
  retab workflows discard-draft wf_abc123

  # Skip the prompt in scripts
  retab workflows discard-draft wf_abc123 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "workflow draft", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.DiscardDraft(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsListCmd.Flags().String("before", "", "workflow id: return items before this id (mutually exclusive with --after)")
	workflowsListCmd.Flags().String("after", "", "workflow id: return items after this id (mutually exclusive with --before)")
	workflowsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	workflowsListCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
	workflowsListCmd.Flags().Var(newEnumStringFlagValue("--sort-by", "updated_at"), "sort-by", "sort field: updated_at")
	workflowsListCmd.Flags().String("project-id", "", "only return workflows belonging to this project")

	workflowsCreateCmd.Flags().String("name", "", "workflow name")
	workflowsCreateCmd.Flags().String("description", "", "workflow description")
	workflowsCreateCmd.Flags().String("project-id", "", "project that should own the workflow (required)")
	_ = workflowsCreateCmd.MarkFlagRequired("project-id")

	workflowsUpdateCmd.Flags().String("name", "", "new name")
	workflowsUpdateCmd.Flags().String("description", "", "new description")

	workflowsPublishCmd.Flags().String("description", "", "publish description")
	workflowsPublishCmd.Flags().Bool("force", false, "skip the empty-workflow warning")

	workflowsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsDiscardDraftCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsCmd.AddCommand(workflowsListCmd, workflowsGetCmd, workflowsCreateCmd, workflowsUpdateCmd, workflowsDeleteCmd, workflowsPublishCmd, workflowsDiscardDraftCmd)
	rootCmd.AddCommand(workflowsCmd)
}
