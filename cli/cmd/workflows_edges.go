package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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
from a start block in the visual editor, edges are auto-created. Reach
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

  # Wire two blocks; "start" resolves to the workflow's single start block
  retab workflows edges create wf_abc123 \
    --source-block start --target-block blk_extract_1

  # Clear the entire graph wiring (blocks remain)
  retab workflows edges delete-all wf_abc123`,
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

func resolveWorkflowEdgeStartAliases(ctx context.Context, client *retab.Client, workflowID string, req *retab.WorkflowEdgeCreateRequest) error {
	if req.SourceBlock != "start" && req.TargetBlock != "start" {
		return nil
	}
	blocks, err := client.Workflows.Blocks.List(ctx, workflowID)
	if err != nil {
		return err
	}
	hasLiteralStartID := false
	startBlockIDs := []string{}
	for _, block := range blocks.Data {
		if block.ID == "start" {
			hasLiteralStartID = true
		}
		if block.Type == "start" {
			startBlockIDs = append(startBlockIDs, block.ID)
		}
	}
	if hasLiteralStartID {
		return nil
	}
	if len(startBlockIDs) == 0 {
		return fmt.Errorf("start alias requested, but workflow has no start block")
	}
	if len(startBlockIDs) > 1 {
		return fmt.Errorf("start alias is ambiguous: workflow has multiple start blocks")
	}
	if req.SourceBlock == "start" {
		req.SourceBlock = startBlockIDs[0]
	}
	if req.TargetBlock == "start" {
		req.TargetBlock = startBlockIDs[0]
	}
	return nil
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
		return printResult(cmd, result)
	}),
}

var workflowsEdgesGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <edge-id>",
	Short: "Get an edge",
	Long:  `Fetch a single edge: source, target, handles.`,
	Example: `  # Inspect an edge
  retab workflows edges get wf_abc123 edg_ghi789`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Edges.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
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
for split/classify file routes and ` + "`output-json-needs-review`" + `
for conditional JSON routes. If a route omits ` + "`handle_key`" + `, Retab
derives one from the human label by lowercasing it and replacing spaces with
hyphens.

For ` + "`--source-block start`" + ` or ` + "`--target-block start`" + `, ` + "`start`" + `
is an alias for the workflow's single block of type ` + "`start`" + ` unless a
block with id ` + "`start`" + ` already exists.`,
	Example: `  # Connect the start document to an extractor
  retab workflows edges create wf_abc123 \
    --source-block start \
    --source-handle document \
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
		if err := resolveWorkflowEdgeStartAliases(ctx, client, args[0], &req); err != nil {
			return err
		}
		ensureWorkflowEdgeID(&req)
		result, err := client.Workflows.Edges.Create(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEdgesCreateBatchCmd = &cobra.Command{
	Use:   "create-batch <workflow-id>",
	Short: "Create multiple edges from --edges-file (JSON array)",
	Long: `Wire many edges in one call. The file is a JSON array of edge
objects with ` + "`source_block`" + `, ` + "`target_block`" + `, and optional
` + "`source_handle`" + `, ` + "`target_handle`" + `, ` + "`id`" + `.`,
	Example: `  # Bulk-wire a graph from a manifest
  retab workflows edges create-batch wf_abc123 \
    --edges-file ./graph/edges.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		path, _ := cmd.Flags().GetString("edges-file")
		if path == "" {
			return fmt.Errorf("--edges-file is required")
		}
		arr, err := readJSONArray(path)
		if err != nil {
			return err
		}
		if len(arr) == 0 {
			return fmt.Errorf("--edges-file: empty JSON array")
		}
		var reqs []retab.WorkflowEdgeCreateRequest
		for i, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("--edges-file[%d]: must be a JSON object", i)
			}
			req := parseEdgeCreate(obj)
			if err := validateWorkflowEdgeCreate(req); err != nil {
				return fmt.Errorf("--edges-file[%d]: %w", i, err)
			}
			reqs = append(reqs, req)
		}
		result, err := client.Workflows.Edges.CreateBatch(ctx, args[0], reqs)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEdgesDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id> <edge-id>",
	Short: "Delete an edge",
	Long: `Remove a single edge. The blocks remain — only the wiring is
severed.`,
	Example: `  # Disconnect two blocks
  retab workflows edges delete wf_abc123 edg_ghi789`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Edges.Delete(ctx, args[0], args[1]); err != nil {
			return err
		}
		confirmDeleted("edge", args[1])
		return nil
	}),
}

var workflowsEdgesDeleteAllCmd = &cobra.Command{
	Use:   "delete-all <workflow-id>",
	Short: "Delete all edges in a workflow",
	Long: `Sever every edge in the draft graph at once. Blocks remain;
re-wire from scratch with ` + "`workflows edges create`" + ` or
` + "`workflows edges create-batch`" + `. Useful when scripting a graph
rewrite.`,
	Example: `  # Reset the wiring before re-creating from a manifest
  retab workflows edges delete-all wf_abc123
  retab workflows edges create-batch wf_abc123 --edges-file ./edges.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Edges.DeleteAll(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("all edges in workflow", args[0])
		return nil
	}),
}

func init() {
	workflowsEdgesListCmd.Flags().String("source-block", "", "filter by source block")
	workflowsEdgesListCmd.Flags().String("target-block", "", "filter by target block")

	workflowsEdgesCreateCmd.Flags().String("id", "", "edge id (optional)")
	workflowsEdgesCreateCmd.Flags().String("source-block", "", "source block id (required) (use start for the single start block)")
	workflowsEdgesCreateCmd.Flags().String("target-block", "", "target block id (required) (use start for the single start block)")
	workflowsEdgesCreateCmd.Flags().String("source-handle", "", "source handle")
	workflowsEdgesCreateCmd.Flags().String("target-handle", "", "target handle")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("source-block")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("target-block")

	workflowsEdgesCreateBatchCmd.Flags().String("edges-file", "", "JSON array of edges (or - for stdin)")

	workflowsEdgesCmd.AddCommand(workflowsEdgesListCmd, workflowsEdgesGetCmd, workflowsEdgesCreateCmd, workflowsEdgesCreateBatchCmd, workflowsEdgesDeleteCmd, workflowsEdgesDeleteAllCmd)
	workflowsCmd.AddCommand(workflowsEdgesCmd)
}
