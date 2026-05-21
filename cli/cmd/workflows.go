package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// warnIfEmptyWorkflowOnPublish prints a stderr warning when the workflow has
// 0 or 1 blocks and the single block (if any) is the auto-added `start-document`
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
	blocks, err := client.Workflows.Blocks.List(ctx, workflowID)
	if err != nil {
		return
	}
	if !isEffectivelyEmptyDraft(blocks.Data) {
		return
	}
	fmt.Fprintln(w, "warning: workflow has only a start-document block — publishing an empty workflow.")
	fmt.Fprintln(w, "warning: add blocks with `retab workflows blocks create` before publishing.")
}

// isEffectivelyEmptyDraft returns true for the two empty-ish shapes that
// produce a no-op published version: a fully empty block list (shouldn't
// happen via the UI but is possible via the API), and the canonical
// freshly-created-workflow shape of exactly one `start-document` block.
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
	return block.Type == "start-document" || block.Type == "start"
}

func workflowGraphObjects(body map[string]any, key string) ([]map[string]any, error) {
	raw, ok := body[key]
	if !ok {
		return nil, nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("--graph-file: %s must be an array", key)
	}
	out := make([]map[string]any, 0, len(arr))
	for i, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("--graph-file: %s[%d] must be an object", key, i)
		}
		out = append(out, normalizeWorkflowGraphObject(key, obj))
	}
	return out, nil
}

func normalizeWorkflowGraphObject(key string, obj map[string]any) map[string]any {
	switch key {
	case "blocks":
		out := map[string]any{}
		copyWorkflowGraphField(out, obj, "id")
		copyWorkflowGraphField(out, obj, "type")
		copyWorkflowGraphField(out, obj, "label")
		copyWorkflowGraphField(out, obj, "config")
		copyWorkflowGraphField(out, obj, "width")
		copyWorkflowGraphField(out, obj, "height")
		copyWorkflowGraphField(out, obj, "parent_id")
		if _, ok := obj["position"]; ok {
			copyWorkflowGraphField(out, obj, "position")
		} else if _, hasX := obj["position_x"]; hasX {
			out["position"] = map[string]any{"x": obj["position_x"], "y": obj["position_y"]}
		} else if _, hasY := obj["position_y"]; hasY {
			out["position"] = map[string]any{"x": obj["position_x"], "y": obj["position_y"]}
		}
		return out
	case "edges":
		out := map[string]any{}
		copyWorkflowGraphField(out, obj, "id")
		copyWorkflowGraphEndpoint(out, obj, "source", "source_block")
		copyWorkflowGraphEndpoint(out, obj, "target", "target_block")
		copyWorkflowGraphField(out, obj, "source_handle")
		copyWorkflowGraphField(out, obj, "target_handle")
		return out
	default:
		return obj
	}
}

func copyWorkflowGraphField(dst map[string]any, src map[string]any, key string) {
	if value, ok := src[key]; ok {
		dst[key] = value
	}
}

func copyWorkflowGraphEndpoint(dst map[string]any, src map[string]any, key string, fallbackKey string) {
	if value, ok := src[key]; ok {
		dst[key] = value
		return
	}
	if value, ok := src[fallbackKey]; ok {
		dst[key] = value
	}
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
	Long: `List workflows in your project, with id pagination and field
selection. Use ` + "`--fields`" + ` to trim large list payloads, and
` + "`--after`" + ` / ` + "`--before`" + ` to walk through pages.`,
	Example: `  # First page
  retab workflows list --limit 20

  # Next page (use the last workflow id from the previous response)
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
		return printWorkflowList(cmd, result)
	}),
}

// printWorkflowList renders a workflow list response. For JSON output it
// rebuilds the envelope from each Workflow's Raw bytes (the exact server
// payload) so server-side `--fields` trimming survives to stdout —
// re-marshaling the typed struct would re-inflate zero-valued fields
// (name, description, email_trigger) and defeat the flag's whole purpose.
// Table output keeps the shared renderer: it only surfaces id/name/
// created_at, so struct re-inflation is harmless there.
func printWorkflowList(cmd *cobra.Command, result *retab.PaginatedList[retab.Workflow]) error {
	if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputTable) {
		return printResultTable(result)
	}
	rows := make([]json.RawMessage, len(result.Data))
	for i, wf := range result.Data {
		if len(wf.Raw) > 0 {
			rows[i] = wf.Raw
			continue
		}
		// Defensive: an element with no captured Raw (shouldn't happen for
		// a decoded list response) falls back to the typed marshal.
		encoded, err := json.Marshal(wf)
		if err != nil {
			return err
		}
		rows[i] = encoded
	}
	return printJSON(struct {
		Data         []json.RawMessage      `json:"data"`
		ListMetadata retab.PaginationCursor `json:"list_metadata"`
		HasMore      bool                   `json:"has_more,omitempty"`
		Total        int                    `json:"total,omitempty"`
	}{
		Data:         rows,
		ListMetadata: result.ListMetadata,
		HasMore:      result.HasMore,
		Total:        result.Total,
	})
}

var workflowsGetCmd = &cobra.Command{
	Use:   "get <workflow-id>",
	Short: "Get a workflow",
	Long: `Fetch the workflow envelope: name, description, current draft and
published versions, email-trigger settings, timestamps.

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
start-document block; add processing blocks (` + "`workflows blocks create`" + `) and
wire them with edges (` + "`workflows edges create`" + `) to make the workflow
functional.`,
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
		if cmd.Flags().Changed("name") {
			if err := validateWorkflowName(name); err != nil {
				return err
			}
		}
		description, _ := cmd.Flags().GetString("description")
		result, err := client.Workflows.Create(ctx, retab.CreateWorkflowRequest{
			Name:        name,
			Description: description,
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
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
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") &&
			!cmd.Flags().Changed("allowed-sender") && !cmd.Flags().Changed("allowed-domain") {
			return fmt.Errorf("nothing to update: pass at least one of --name, --description, --allowed-sender, or --allowed-domain")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		var req retab.UpdateWorkflowRequest
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			if err := validateWorkflowName(v); err != nil {
				return err
			}
			req.Name = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
		}
		senders, _ := cmd.Flags().GetStringArray("allowed-sender")
		domains, _ := cmd.Flags().GetStringArray("allowed-domain")
		if cmd.Flags().Changed("allowed-sender") || cmd.Flags().Changed("allowed-domain") {
			if err := validateWorkflowEmailAllowlistValues("--allowed-sender", senders); err != nil {
				return err
			}
			if err := validateWorkflowEmailAllowlistValues("--allowed-domain", domains); err != nil {
				return err
			}
			req.EmailTrigger = &retab.WorkflowEmailTrigger{
				AllowedSenders: senders,
				AllowedDomains: domains,
			}
		}
		result, err := client.Workflows.Update(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func validateWorkflowName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("workflow name must not be blank")
	}
	return nil
}

func validateWorkflowEmailAllowlistValues(flagName string, values []string) error {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s must not be blank", flagName)
		}
	}
	return nil
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

By default, publishing a draft that contains only the auto-added ` + "`start-document`" + `
block (i.e. an effectively empty workflow) prints a warning to stderr but
proceeds — the publish itself succeeds. Pass ` + "`--force`" + ` to skip
the warning.`,
	Example: `  # Publish with a release note
  retab workflows publish wf_abc123 \
    --description "v3: tighter line-item schema"

  # Pin a run to a specific published version
  retab workflows runs create wf_abc123 \
    --version v3 --document start=./invoice.pdf

  # Skip the "empty workflow" warning when intentionally publishing a stub
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
			// stderr is the warning sink: it doesn't pollute the JSON
			// payload on stdout that callers pipe into `jq`.
			warnIfEmptyWorkflowOnPublish(ctx, client, args[0], cmd.ErrOrStderr())
		}
		result, err := client.Workflows.Publish(ctx, args[0], retab.PublishWorkflowRequest{Description: description})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsDiagnoseCmd = &cobra.Command{
	Use:   "diagnose <workflow-id>",
	Short: "Diagnose the persisted draft graph (use --graph-file to send an in-memory graph)",
	Long: `Validate a workflow's draft graph and report structural/config issues:
disconnected blocks, type mismatches across edges, incomplete configs,
review configuration warnings, unreachable paths.

By default diagnoses the persisted draft. Pass ` + "`--graph-file`" + ` to
diagnose an in-memory graph (e.g. before persisting changes) — the file
must be a JSON object with ` + "`blocks`" + `, ` + "`edges`" + `, and optional
` + "`re_propagate`" + ` fields.`,
	Example: `  # Diagnose the persisted draft
  retab workflows diagnose wf_abc123

  # Diagnose a proposed graph before persisting
  retab workflows diagnose wf_abc123 --graph-file ./graph.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		graphFile, _ := cmd.Flags().GetString("graph-file")
		var graphReq *retab.DiagnoseWorkflowGraphRequest
		if graphFile != "" {
			body, err := readJSONMap(graphFile)
			if err != nil {
				return err
			}
			req := retab.DiagnoseWorkflowGraphRequest{RePropagate: true}
			blocks, err := workflowGraphObjects(body, "blocks")
			if err != nil {
				return err
			}
			req.Blocks = blocks
			edges, err := workflowGraphObjects(body, "edges")
			if err != nil {
				return err
			}
			req.Edges = edges
			if v, ok := body["re_propagate"].(bool); ok {
				req.RePropagate = v
			}
			graphReq = &req
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if graphReq != nil {
			result, err := client.Workflows.DiagnoseGraph(ctx, args[0], *graphReq)
			if err != nil {
				return err
			}
			return printResult(cmd, result)
		}
		result, err := client.Workflows.Diagnose(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsListCmd.Flags().String("before", "", "workflow id: return items before this id")
	workflowsListCmd.Flags().String("after", "", "workflow id: return items after this id")
	workflowsListCmd.Flags().Var(&boundedIntFlagValue{min: 0, max: 100}, "limit", "max items to return (1-100)")
	workflowsListCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
	workflowsListCmd.Flags().Var(newEnumStringFlagValue("--sort-by", "updated_at"), "sort-by", "sort field: updated_at")
	workflowsListCmd.Flags().String("fields", "", "comma-separated field list to return")

	workflowsCreateCmd.Flags().String("name", "", "workflow name")
	workflowsCreateCmd.Flags().String("description", "", "workflow description")

	workflowsUpdateCmd.Flags().String("name", "", "new name")
	workflowsUpdateCmd.Flags().String("description", "", "new description")
	workflowsUpdateCmd.Flags().StringArray("allowed-sender", nil, "email trigger allowed sender (repeatable)")
	workflowsUpdateCmd.Flags().StringArray("allowed-domain", nil, "email trigger allowed domain (repeatable)")

	workflowsPublishCmd.Flags().String("description", "", "publish description")
	workflowsPublishCmd.Flags().Bool("force", false, "skip the empty-workflow warning")

	workflowsDiagnoseCmd.Flags().String("graph-file", "", "JSON file with {blocks, edges, re_propagate} to diagnose without persisting")

	workflowsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsCmd.AddCommand(workflowsListCmd, workflowsGetCmd, workflowsCreateCmd, workflowsUpdateCmd, workflowsDeleteCmd, workflowsPublishCmd, workflowsDiagnoseCmd)
	rootCmd.AddCommand(workflowsCmd)
}
