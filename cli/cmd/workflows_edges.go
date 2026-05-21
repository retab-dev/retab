package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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

func parseEdgeCreate(obj map[string]any) retab.WorkflowEdgeCreateRequest {
	req := retab.WorkflowEdgeCreateRequest{}
	if v, ok := obj["id"].(string); ok {
		req.ID = v
	}
	if v, ok := obj["source_block"].(string); ok {
		req.SourceBlock = v
	}
	if v, ok := obj["target_block"].(string); ok {
		req.TargetBlock = v
	}
	if v, ok := obj["source_handle"].(string); ok {
		req.SourceHandle = v
	}
	if v, ok := obj["target_handle"].(string); ok {
		req.TargetHandle = v
	}
	ensureWorkflowEdgeID(&req)
	return req
}

func validateWorkflowEdgeCreate(req retab.WorkflowEdgeCreateRequest) error {
	if strings.TrimSpace(req.SourceBlock) == "" {
		return fmt.Errorf("source_block is required")
	}
	if strings.TrimSpace(req.TargetBlock) == "" {
		return fmt.Errorf("target_block is required")
	}
	return nil
}

func resolveWorkflowEdgeAliases(ctx context.Context, client *retab.Client, workflowID string, req *retab.WorkflowEdgeCreateRequest) error {
	needsBlockLookup := req.SourceBlock == "start" || req.TargetBlock == "start" || req.TargetHandle != ""
	if !needsBlockLookup {
		return nil
	}
	blocks, err := client.Workflows.Blocks.List(ctx, workflowID)
	if err != nil {
		return err
	}
	if err := resolveWorkflowEdgeStartAliasesFromBlocks(blocks.Data, req); err != nil {
		return err
	}
	resolveWorkflowEdgeTargetHandleAliasFromBlocks(blocks.Data, req)
	return nil
}

func resolveWorkflowEdgeStartAliasesFromBlocks(blocks []retab.WorkflowBlock, req *retab.WorkflowEdgeCreateRequest) error {
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

func resolveWorkflowEdgeTargetHandleAliasFromBlocks(blocks []retab.WorkflowBlock, req *retab.WorkflowEdgeCreateRequest) {
	if req.TargetHandle == "" || strings.HasPrefix(req.TargetHandle, "input-") {
		return
	}
	for _, block := range blocks {
		if block.ID != req.TargetBlock {
			continue
		}
		if resolved := targetInputHandleAlias(block.Type, block.Config, req.TargetHandle); resolved != "" {
			req.TargetHandle = resolved
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

func ensureWorkflowEdgeID(req *retab.WorkflowEdgeCreateRequest) {
	if req.ID != "" {
		return
	}
	req.ID = defaultWorkflowEdgeID(*req)
}

func defaultWorkflowEdgeID(req retab.WorkflowEdgeCreateRequest) string {
	parts := []string{req.SourceBlock, req.SourceHandle, req.TargetBlock, req.TargetHandle}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	encoded := base64.RawURLEncoding.EncodeToString(sum[:])
	return "edge_" + encoded[:22]
}

var workflowsEdgesListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List edges in a workflow",
	Long: `List edges in a workflow's draft graph. Filter by either endpoint
to focus on a single block's wiring.`,
	Example: `  # All edges
  retab workflows edges list wf_abc123

  # Just the edges that fan out of one block
  retab workflows edges list wf_abc123 --source-block blk_def456`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowEdgesParams{}
		params.SourceBlock, _ = cmd.Flags().GetString("source-block")
		params.TargetBlock, _ = cmd.Flags().GetString("target-block")
		result, err := client.Workflows.Edges.List(ctx, args[0], &params)
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
  retab workflows edges get edg_ghi789`,
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
		req := retab.WorkflowEdgeCreateRequest{}
		req.SourceBlock, _ = cmd.Flags().GetString("source-block")
		req.TargetBlock, _ = cmd.Flags().GetString("target-block")
		req.SourceHandle, _ = cmd.Flags().GetString("source-handle")
		req.TargetHandle, _ = cmd.Flags().GetString("target-handle")
		req.ID, _ = cmd.Flags().GetString("id")
		if err := validateWorkflowEdgeCreate(req); err != nil {
			return err
		}
		if err := resolveWorkflowEdgeAliases(ctx, client, args[0], &req); err != nil {
			return err
		}
		ensureWorkflowEdgeID(&req)
		result, err := client.Workflows.Edges.Create(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsEdgesDeleteCmd = &cobra.Command{
	Use:   "delete <edge-id>",
	Short: "Delete an edge",
	Long: `Remove a single edge. The blocks remain — only the wiring is
severed.`,
	Example: `  # Disconnect two blocks
  retab workflows edges delete edg_ghi789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
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
	workflowsEdgesListCmd.Flags().String("source-block", "", "filter by source block")
	workflowsEdgesListCmd.Flags().String("target-block", "", "filter by target block")

	workflowsEdgesCreateCmd.Flags().String("id", "", "edge id (optional)")
	workflowsEdgesCreateCmd.Flags().String("source-block", "", "source block id (required) (use start for the single start_document block)")
	workflowsEdgesCreateCmd.Flags().String("target-block", "", "target block id (required) (use start for the single start_document block)")
	workflowsEdgesCreateCmd.Flags().String("source-handle", "", "source handle")
	workflowsEdgesCreateCmd.Flags().String("target-handle", "", "target handle")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("source-block")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("target-block")

	workflowsEdgesCmd.AddCommand(workflowsEdgesListCmd, workflowsEdgesGetCmd, workflowsEdgesCreateCmd, workflowsEdgesDeleteCmd)
	workflowsCmd.AddCommand(workflowsEdgesCmd)
}
