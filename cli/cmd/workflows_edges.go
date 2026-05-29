//go:build !retab_oagen_cli_workflows_edges

package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsEdgesCmd = &cobra.Command{
	Use:   "edges",
	Short: "Manage workflow edges",
	Long: `Wire blocks together. Data flows from ` + "`source_block.output`" + `
into ` + "`target_block.input`" + `, with optional handles
(` + "`source_handle`" + `, ` + "`target_handle`" + `) for blocks that expose
multiple ports.

Most workflows don't need direct edge management: when you add a block
from a start_document block in the visual editor, edges are auto-created. Reach
for ` + "`workflows edges create`" + ` when scaffolding a graph from JSON,
re-wiring after a refactor, or fixing a disconnected node flagged by
` + "`workflows diagnose`" + `.

Dynamic split, classifier, and conditional outputs use ` + "`handle_key`" + `
values, not display labels. For example, a subdocument named "booking
confirmation" exposes ` + "`output-file-booking-confirmation`" + `. Use
` + "`workflows blocks get`" + ` to inspect the block config before wiring
dynamic ports.`,
	Example: `  # Inspect every edge
  retab workflows edges list wf_abc123

  # Wire two blocks; "start" resolves to the workflow's single start_document block
  retab workflows edges create wf_abc123 \
    --source-block start --target-block blk_extract_1`,
}

func parseEdgeCreate(obj map[string]any) retab.WorkflowEdgesCreateParams {
	req := retab.WorkflowEdgesCreateParams{}
	if v, ok := obj["id"].(string); ok && v != "" {
		req.ID = ptr(v)
	}
	if v, ok := obj["source_block"].(string); ok {
		req.SourceBlock = v
	}
	if v, ok := obj["target_block"].(string); ok {
		req.TargetBlock = v
	}
	if v, ok := obj["source_handle"].(string); ok && v != "" {
		req.SourceHandle = ptr(v)
	}
	if v, ok := obj["target_handle"].(string); ok && v != "" {
		req.TargetHandle = ptr(v)
	}
	ensureWorkflowEdgeID(&req)
	return req
}

func validateWorkflowEdgeCreate(req retab.WorkflowEdgesCreateParams) error {
	if strings.TrimSpace(req.SourceBlock) == "" {
		return fmt.Errorf("source_block is required")
	}
	if strings.TrimSpace(req.TargetBlock) == "" {
		return fmt.Errorf("target_block is required")
	}
	return nil
}

func resolveWorkflowEdgeAliases(ctx context.Context, client *retab.Client, workflowID string, req *retab.WorkflowEdgesCreateParams) error {
	needsBlockLookup := req.SourceBlock == "start" || req.TargetBlock == "start" || req.TargetHandle != nil
	if !needsBlockLookup {
		return nil
	}
	blocks, err := client.Workflows.Blocks.List(ctx, &retab.WorkflowBlocksListParams{WorkflowID: workflowID})
	if err != nil {
		return err
	}
	if err := resolveWorkflowEdgeStartAliasesFromBlocks(blocks.Data, req); err != nil {
		return err
	}
	resolveWorkflowEdgeTargetHandleAliasFromBlocks(blocks.Data, req)
	return nil
}

func resolveWorkflowEdgeStartAliasesFromBlocks(blocks []retab.WorkflowBlock, req *retab.WorkflowEdgesCreateParams) error {
	if req.SourceBlock != "start" && req.TargetBlock != "start" {
		return nil
	}
	hasLiteralStartID := false
	startBlockIDs := []string{}
	for _, block := range blocks {
		if block.ID == "start" {
			hasLiteralStartID = true
		}
		if isStartDocumentBlock(block) {
			startBlockIDs = append(startBlockIDs, block.ID)
		}
	}
	if hasLiteralStartID {
		return nil
	}
	if len(startBlockIDs) == 0 {
		return fmt.Errorf("start alias requested, but workflow has no start_document block")
	}
	if len(startBlockIDs) > 1 {
		return fmt.Errorf("start alias is ambiguous: workflow has multiple start_document blocks")
	}
	if req.SourceBlock == "start" {
		req.SourceBlock = startBlockIDs[0]
	}
	if req.TargetBlock == "start" {
		req.TargetBlock = startBlockIDs[0]
	}
	return nil
}

func resolveWorkflowEdgeTargetHandleAliasFromBlocks(blocks []retab.WorkflowBlock, req *retab.WorkflowEdgesCreateParams) {
	if req.TargetHandle == nil || *req.TargetHandle == "" || strings.HasPrefix(*req.TargetHandle, "input-") {
		return
	}
	for _, block := range blocks {
		if block.ID != req.TargetBlock {
			continue
		}
		if resolved := targetInputHandleAlias(string(block.Type), block.Config, *req.TargetHandle); resolved != "" {
			req.TargetHandle = ptr(resolved)
		}
		return
	}
}

func targetInputHandleAlias(blockType string, config map[string]any, inputName string) string {
	rawInputs, ok := config["inputs"].([]any)
	if ok {
		for _, rawInput := range rawInputs {
			input, ok := rawInput.(map[string]any)
			if !ok {
				continue
			}
			name, _ := input["name"].(string)
			inputType, _ := input["type"].(string)
			if name == inputName && inputType != "" {
				return "input-" + inputType + "-" + name
			}
		}
	}
	if rawInput, ok := config["input"].(map[string]any); ok {
		name, _ := rawInput["name"].(string)
		inputType, _ := rawInput["type"].(string)
		if name == inputName && inputType != "" {
			return "input-" + inputType + "-" + name
		}
	}
	if inputName == "document" {
		switch blockType {
		case "extract", "classifier":
			return "input-file-document"
		case "split":
			return "input-file-0"
		}
	}
	return ""
}

func ensureWorkflowEdgeID(req *retab.WorkflowEdgesCreateParams) {
	if req.ID != nil && *req.ID != "" {
		return
	}
	req.ID = ptr(defaultWorkflowEdgeID(*req))
}

func defaultWorkflowEdgeID(req retab.WorkflowEdgesCreateParams) string {
	parts := []string{req.SourceBlock, derefString(req.SourceHandle), req.TargetBlock, derefString(req.TargetHandle)}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	encoded := base64.RawURLEncoding.EncodeToString(sum[:])
	return "edge_" + encoded[:22]
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

var workflowsEdgesListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List edges in a workflow",
	Long: `List edges in a workflow's draft graph. Filter by either endpoint
to focus on a single block's wiring.

Name the workflow either positionally (` + "`list <workflow-id>`" + `) or with
the ` + "`--workflow-id`" + ` flag — the two forms are equivalent. Passing both
is accepted when they agree; an error is raised only when they disagree. The
workflow id is required: edges have no org-wide listing.

Paginate by passing the cursor from a previous response's
` + "`list_metadata`" + `: ` + "`--after`" + ` for the next page,
` + "`--before`" + ` for the previous one. The two are mutually exclusive.`,
	Example: `  # All edges (positional)
  retab workflows edges list wf_abc123

  # Same, with the flag form
  retab workflows edges list --workflow-id wf_abc123

  # Just the edges that fan out of one block
  retab workflows edges list wf_abc123 --source-block blk_def456

  # First page of 50
  retab workflows edges list wf_abc123 --limit 50

  # Next page
  retab workflows edges list wf_abc123 --limit 50 \
    --after $(retab workflows edges list wf_abc123 --limit 50 --output json | jq -r '.list_metadata.after')`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Workflow id positionally OR via --workflow-id (co-equal forms);
		// required here — edges have no org-wide listing.
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
		params := retab.WorkflowEdgesListParams{
			PaginationParams: collectListParams(cmd),
			WorkflowID:       workflowID,
		}
		if v, _ := cmd.Flags().GetString("source-block"); v != "" {
			params.SourceBlock = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("target-block"); v != "" {
			params.TargetBlock = ptr(v)
		}
		result, err := client.Workflows.Edges.List(ctx, &params)
		if err != nil {
			return err
		}
		return printWorkflowEdgesListResult(cmd, result)
	}),
}

func printWorkflowEdgesListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.WorkflowEdgeDoc]) error {
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputTable) {
			return RenderList(os.Stdout, OutputTable, result, workflowEdgeColumns)
		}
	}
	return printJSON(result)
}

var workflowEdgeColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowEdgeCell(row, "id") }},
	{Header: "SOURCE_BLOCK", Extract: func(row any) string { return workflowEdgeCell(row, "source_block") }},
	{Header: "SOURCE_HANDLE", Extract: func(row any) string { return workflowEdgeCell(row, "source_handle") }},
	{Header: "TARGET_BLOCK", Extract: func(row any) string { return workflowEdgeCell(row, "target_block") }},
	{Header: "TARGET_HANDLE", Extract: func(row any) string { return workflowEdgeCell(row, "target_handle") }},
	{Header: "UPDATED_AT", Extract: func(row any) string { return workflowEdgeCell(row, "updated_at") }},
}

func workflowEdgeCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

var workflowsEdgesGetCmd = &cobra.Command{
	Use:   "get <edge-id>",
	Short: "Get an edge",
	Long:  `Fetch a single edge: source, target, handles.`,
	Example: `  # Inspect an edge
  retab workflows edges get edge_ghi789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Edges.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEdgesCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Create an edge",
	Long: `Connect two blocks: data flows from ` + "`--source-block`" + ` into
` + "`--target-block`" + `. Use ` + "`--source-handle`" + ` /
` + "`--target-handle`" + ` for blocks with multiple named ports (e.g. a
` + "`conditional`" + ` block exposing ` + "`true`" + ` / ` + "`false`" + `
branches).

For dynamic routing blocks, source handles are built from route
` + "`handle_key`" + ` values: ` + "`output-file-booking-confirmation`" + `
for split/classifier file routes and ` + "`output-json-needs-review`" + `
for conditional JSON routes. If a route omits ` + "`handle_key`" + `, Retab
derives one from the human label by lowercasing it and replacing spaces with
hyphens.

For ` + "`--source-block start`" + ` or ` + "`--target-block start`" + `, ` + "`start_document`" + `
is an alias for the workflow's single block of type ` + "`start_document`" + ` unless a
block with id ` + "`start_document`" + ` already exists. For ` + "`--target-handle`" + `,
you may pass the friendly input name from the block config, such as
` + "`document`" + `. The CLI resolves ` + "`document`" + ` to
` + "`input-file-document`" + ` for extract/classifier blocks and
` + "`input-file-0`" + ` for split blocks.`,
	Example: `  # Connect the start document to an extractor
  retab workflows edges create wf_abc123 \
    --source-block start \
    --source-handle output-file-0 \
    --target-handle document \
    --target-block blk_extract_1

  # Connect extractor output to a classifier
  retab workflows edges create wf_abc123 \
    --source-block blk_extract_1 \
    --target-block blk_classify_1

  # Connect a conditional branch whose handle_key is "needs-review"
  retab workflows edges create wf_abc123 \
    --source-block blk_cond_1 --source-handle output-json-needs-review \
    --target-block blk_extract_2`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.WorkflowEdgesCreateParams{WorkflowID: args[0]}
		req.SourceBlock, _ = cmd.Flags().GetString("source-block")
		req.TargetBlock, _ = cmd.Flags().GetString("target-block")
		if sourceHandle, _ := cmd.Flags().GetString("source-handle"); sourceHandle != "" {
			req.SourceHandle = ptr(sourceHandle)
		}
		if targetHandle, _ := cmd.Flags().GetString("target-handle"); targetHandle != "" {
			req.TargetHandle = ptr(targetHandle)
		}
		if id, _ := cmd.Flags().GetString("id"); id != "" {
			req.ID = ptr(id)
		}
		// Trim whitespace on block ids so values like "start  " (e.g. from
		// shell completion or copy/paste) don't hit the server verbatim and
		// produce a confusing 404. The "is required" check still runs on
		// the trimmed value below so "   " is rejected with the existing
		// message rather than slipping through as a valid id.
		req.SourceBlock = strings.TrimSpace(req.SourceBlock)
		req.TargetBlock = strings.TrimSpace(req.TargetBlock)
		idWasExplicit := req.ID != nil && strings.TrimSpace(*req.ID) != ""
		if err := validateWorkflowEdgeCreate(req); err != nil {
			return err
		}
		if err := resolveWorkflowEdgeAliases(ctx, client, args[0], &req); err != nil {
			return err
		}
		ensureWorkflowEdgeID(&req)
		result, err := client.Workflows.Edges.Create(ctx, &req)
		if err != nil {
			return rewrapAutoEdgeIDConflict(err, req, idWasExplicit)
		}
		return printResult(cmd, result)
	}),
}

// rewrapAutoEdgeIDConflict rewrites the server's misleading
// "Edge ID already exists" 409 when the CLI generated the edge id
// deterministically from the endpoint tuple (i.e. the user did not pass
// --id). In that case the real cause is the duplicate (source, target,
// handles) tuple, not an id picked by the user, so the message should say
// so. When --id was explicit, the server message is accurate; pass it
// through unchanged.
func rewrapAutoEdgeIDConflict(err error, req retab.WorkflowEdgesCreateParams, idWasExplicit bool) error {
	if idWasExplicit {
		return err
	}
	var apiErr *retab.APIError
	if !errors.As(err, &apiErr) {
		return err
	}
	if apiErr.StatusCode != http.StatusConflict {
		return err
	}
	if !strings.Contains(apiErr.Message, "Edge ID already exists") {
		// Server now sends a clearer tuple-aware 409 in this case — let
		// it through verbatim. We only override the legacy id-only message.
		return err
	}
	src := req.SourceBlock
	if req.SourceHandle != nil && *req.SourceHandle != "" {
		src = fmt.Sprintf("%s[:%s]", req.SourceBlock, *req.SourceHandle)
	}
	tgt := req.TargetBlock
	if req.TargetHandle != nil && *req.TargetHandle != "" {
		tgt = fmt.Sprintf("%s[:%s]", req.TargetBlock, *req.TargetHandle)
	}
	apiErr.Message = fmt.Sprintf(
		"an edge from %s to %s already exists (auto-generated edge id collided)",
		src, tgt,
	)
	return apiErr
}

var workflowsEdgesDeleteCmd = &cobra.Command{
	Use:   "delete <edge-id>",
	Short: "Delete an edge",
	Long: `Remove a single edge. The blocks remain — only the wiring is
severed.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal.`,
	Example: `  # Disconnect two blocks (interactive, asks to confirm)
  retab workflows edges delete edge_ghi789

  # Skip the prompt in scripts
  retab workflows edges delete edge_ghi789 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "edge", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Edges.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("edge", args[0])
		return nil
	}),
}

func init() {
	workflowsEdgesListCmd.Flags().String("workflow-id", "", "workflow id (alternative to the positional form)")
	workflowsEdgesListCmd.Flags().String("source-block", "", "filter by source block")
	workflowsEdgesListCmd.Flags().String("target-block", "", "filter by target block")
	workflowsEdgesListCmd.Flags().String("before", "", "edge id: return the page before this id (mutually exclusive with --after)")
	workflowsEdgesListCmd.Flags().String("after", "", "edge id: return the page after this id (mutually exclusive with --before)")
	workflowsEdgesListCmd.MarkFlagsMutuallyExclusive("before", "after")
	workflowsEdgesListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")

	workflowsEdgesCreateCmd.Flags().String("id", "", "edge id (optional)")
	workflowsEdgesCreateCmd.Flags().String("source-block", "", "source block id (required) (use start for the single start_document block)")
	workflowsEdgesCreateCmd.Flags().String("target-block", "", "target block id (required) (use start for the single start_document block)")
	workflowsEdgesCreateCmd.Flags().String("source-handle", "", "source handle")
	workflowsEdgesCreateCmd.Flags().String("target-handle", "", "target handle")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("source-block")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("target-block")

	workflowsEdgesDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsEdgesCmd.AddCommand(workflowsEdgesListCmd, workflowsEdgesGetCmd, workflowsEdgesCreateCmd, workflowsEdgesDeleteCmd)
	workflowsCmd.AddCommand(workflowsEdgesCmd)
}
