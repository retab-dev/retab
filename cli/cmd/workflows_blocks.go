package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsBlocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Manage workflow blocks",
	Long: `Add, configure, and inspect the nodes of a workflow graph.

A block is one processing step — ` + "`extract`" + `, ` + "`split`" + `,
` + "`classify`" + `, ` + "`edit`" + `, ` + "`hil`" + `, ` + "`conditional`" + `,
` + "`api_call`" + `, ` + "`function`" + `, etc. Each block has a typed input,
typed output, and a JSON ` + "`config`" + ` blob shaped by its type.

The workhorse here is ` + "`update`" + ` — once a block is on the graph, tune
its config with ` + "`workflows blocks update --config-file ./cfg.json`" + `
rather than deleting and re-creating. To test a config change without
re-running the whole workflow, use ` + "`workflows blocks simulate`" + ` to
replay one block against a past run's input.`,
	Example: `  # List blocks
  retab workflows blocks list wf_abc123

  # Add a block from a JSON definition
  retab workflows blocks create wf_abc123 --block-file ./extract.json

  # Tune just the config of an existing block
  retab workflows blocks update wf_abc123 blk_def456 \
    --config-file ./new-config.json

  # Test that config change against an existing run, without a full re-run
  retab workflows blocks simulate \
    --run-id run_xyz789 --block-id blk_def456`,
}

func parseBlockCreate(obj map[string]any) (retab.WorkflowBlockCreateRequest, error) {
	req := retab.WorkflowBlockCreateRequest{}
	if v, ok := obj["id"].(string); ok {
		req.ID = v
	}
	if v, ok := obj["type"].(string); ok {
		req.Type = v
	}
	if v, ok := obj["label"].(string); ok {
		req.Label = v
	}
	if v, ok := obj["position_x"].(float64); ok {
		req.PositionX = v
	}
	if v, ok := obj["position_y"].(float64); ok {
		req.PositionY = v
	}
	if v, ok := obj["width"].(float64); ok {
		req.Width = &v
	}
	if v, ok := obj["height"].(float64); ok {
		req.Height = &v
	}
	if v, ok := obj["parent_id"].(string); ok {
		req.ParentID = v
	}
	if v, ok := obj["config"].(map[string]any); ok {
		req.Config = v
	}
	if req.ID == "" {
		return req, fmt.Errorf("block id is required")
	}
	if req.Type == "" {
		return req, fmt.Errorf("block type is required")
	}
	return req, nil
}

var workflowsBlocksListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List blocks in a workflow",
	Long: `List every block in a workflow's draft graph, including id, type,
label, position, and config.`,
	Example: `  # List all blocks
  retab workflows blocks list wf_abc123

  # Get the ids only
  retab workflows blocks list wf_abc123 | jq -r '.[].id'`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.List(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <block-id>",
	Short: "Get a workflow block",
	Long: `Fetch a single block's full definition: type, label, position,
parent group, and the typed config blob.`,
	Example: `  # Inspect a block
  retab workflows blocks get wf_abc123 blk_def456

  # Save a block's config for offline editing
  retab workflows blocks get wf_abc123 blk_def456 \
    | jq '.config' > cfg.json`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsBlocksResolvedSchemasCmd = &cobra.Command{
	Use:   "resolved-schemas <workflow-id> <block-id>",
	Short: "Get the resolved schema for a single block",
	Long: `Resolve the input and output JSON schemas for one block after
propagating types from its upstream edges. Use this to confirm a block
sees the shape you expect before running.

For all blocks at once, use ` + "`workflows resolved-schemas`" + `.`,
	Example: `  # Inspect schemas for one block
  retab workflows blocks resolved-schemas wf_abc123 blk_def456`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.GetResolvedSchemas(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsBlocksCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Create a workflow block from --block-file",
	Long: `Add a block to a workflow's draft graph. The block file is a
JSON object with the keys ` + "`id`" + ` (required), ` + "`type`" + ` (required),
` + "`label`" + `, ` + "`position_x`" + `, ` + "`position_y`" + `, ` + "`width`" + `,
` + "`height`" + `, ` + "`parent_id`" + `, and ` + "`config`" + `.

For batch creation, see ` + "`workflows blocks create-batch`" + `.`,
	Example: `  # Add one block from a JSON file
  retab workflows blocks create wf_abc123 --block-file ./extract.json

  # Pipe a block definition from stdin
  cat block.json | retab workflows blocks create wf_abc123 --block-file -`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		path, _ := cmd.Flags().GetString("block-file")
		if path == "" {
			return fmt.Errorf("--block-file is required")
		}
		obj, err := readJSONMap(path)
		if err != nil {
			return err
		}
		req, err := parseBlockCreate(obj)
		if err != nil {
			return err
		}
		result, err := client.Workflows.Blocks.Create(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsBlocksCreateBatchCmd = &cobra.Command{
	Use:   "create-batch <workflow-id>",
	Short: "Create multiple workflow blocks from --blocks-file (JSON array)",
	Long: `Add many blocks in one call. The file is a JSON array of block
objects, each shaped like the ` + "`--block-file`" + ` payload accepted by
` + "`workflows blocks create`" + `.

Preferred when scaffolding an entire workflow programmatically — fewer
round-trips and atomic from the caller's perspective.`,
	Example: `  # Bulk-create a graph from a manifest
  retab workflows blocks create-batch wf_abc123 \
    --blocks-file ./graph/blocks.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		path, _ := cmd.Flags().GetString("blocks-file")
		if path == "" {
			return fmt.Errorf("--blocks-file is required")
		}
		arr, err := readJSONArray(path)
		if err != nil {
			return err
		}
		var reqs []retab.WorkflowBlockCreateRequest
		for i, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("--blocks-file[%d]: must be a JSON object", i)
			}
			r, err := parseBlockCreate(obj)
			if err != nil {
				return fmt.Errorf("--blocks-file[%d]: %w", i, err)
			}
			reqs = append(reqs, r)
		}
		result, err := client.Workflows.Blocks.CreateBatch(ctx, args[0], reqs)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsBlocksUpdateCmd = &cobra.Command{
	Use:   "update <workflow-id> <block-id>",
	Short: "Update a workflow block",
	Long: `Tune an existing block in place. The most common use is swapping
the typed config blob (` + "`--config-file`" + `) to adjust the prompt,
schema, model, or thresholds without re-creating the block — your
upstream/downstream wiring is preserved.

Layout fields (` + "`position-*`" + `, ` + "`width`" + `, ` + "`height`" + `,
` + "`parent-id`" + `) only affect the visual editor.`,
	Example: `  # Swap the config blob
  retab workflows blocks update wf_abc123 blk_def456 \
    --config-file ./new-config.json

  # Rename a block's label
  retab workflows blocks update wf_abc123 blk_def456 \
    --label "Extract line items"`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.WorkflowBlockUpdateRequest{}
		if cmd.Flags().Changed("label") {
			v, _ := cmd.Flags().GetString("label")
			req.Label = &v
		}
		if cmd.Flags().Changed("position-x") {
			v, _ := cmd.Flags().GetFloat64("position-x")
			req.PositionX = &v
		}
		if cmd.Flags().Changed("position-y") {
			v, _ := cmd.Flags().GetFloat64("position-y")
			req.PositionY = &v
		}
		if cmd.Flags().Changed("width") {
			v, _ := cmd.Flags().GetFloat64("width")
			req.Width = &v
		}
		if cmd.Flags().Changed("height") {
			v, _ := cmd.Flags().GetFloat64("height")
			req.Height = &v
		}
		if cmd.Flags().Changed("parent-id") {
			v, _ := cmd.Flags().GetString("parent-id")
			req.ParentID = &v
		}
		if path, _ := cmd.Flags().GetString("config-file"); path != "" {
			cfg, err := readJSONMap(path)
			if err != nil {
				return fmt.Errorf("--config-file: %w", err)
			}
			req.Config = cfg
		}
		result, err := client.Workflows.Blocks.Update(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsBlocksDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id> <block-id>",
	Short: "Delete a workflow block",
	Long: `Remove a block from the draft graph. Edges that referenced this
block are also deleted. Past runs that used this block remain intact —
deletion only affects the draft.`,
	Example: `  # Remove a block
  retab workflows blocks delete wf_abc123 blk_def456`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Blocks.Delete(ctx, args[0], args[1]); err != nil {
			return err
		}
		confirmDeleted("block", args[1])
		return nil
	}),
}

var workflowsBlocksSimulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Replay one block against an existing run",
	Long: `Re-execute a single block using a past run's stored input,
without re-running the whole workflow. The fastest feedback loop for
tuning a block's config: update the config, simulate against a known
input, compare outputs.

Pass ` + "`--n-consensus`" + ` to draw multiple samples for variance
testing.`,
	Example: `  # Replay block blk_def456 against an existing run
  retab workflows blocks simulate \
    --run-id run_xyz789 --block-id blk_def456

  # Replay with 3-sample consensus
  retab workflows blocks simulate \
    --run-id run_xyz789 --block-id blk_def456 --n-consensus 3`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		runID, _ := cmd.Flags().GetString("run-id")
		blockID, _ := cmd.Flags().GetString("block-id")
		stepID, _ := cmd.Flags().GetString("step-id")
		nConsensus, _ := cmd.Flags().GetInt("n-consensus")
		req := retab.SimulateBlockRequest{
			RunID:      runID,
			BlockID:    blockID,
			StepID:     stepID,
			NConsensus: nConsensus,
		}
		if cmd.Flags().Changed("check-eligibility") {
			v, _ := cmd.Flags().GetBool("check-eligibility")
			req.CheckEligibility = &v
		}
		result, err := client.Workflows.Blocks.Simulate(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	workflowsBlocksCreateCmd.Flags().String("block-file", "", "JSON file describing the block (or - for stdin)")
	workflowsBlocksCreateBatchCmd.Flags().String("blocks-file", "", "JSON array of blocks (or - for stdin)")

	workflowsBlocksUpdateCmd.Flags().String("label", "", "update label")
	workflowsBlocksUpdateCmd.Flags().Float64("position-x", 0, "update X position")
	workflowsBlocksUpdateCmd.Flags().Float64("position-y", 0, "update Y position")
	workflowsBlocksUpdateCmd.Flags().Float64("width", 0, "update width")
	workflowsBlocksUpdateCmd.Flags().Float64("height", 0, "update height")
	workflowsBlocksUpdateCmd.Flags().String("parent-id", "", "update parent id")
	workflowsBlocksUpdateCmd.Flags().String("config-file", "", "JSON file with new config (or - for stdin)")

	workflowsBlocksSimulateCmd.Flags().String("run-id", "", "run id (required)")
	workflowsBlocksSimulateCmd.Flags().String("block-id", "", "block id (required)")
	workflowsBlocksSimulateCmd.Flags().String("step-id", "", "step id")
	workflowsBlocksSimulateCmd.Flags().Int("n-consensus", 0, "consensus count")
	workflowsBlocksSimulateCmd.Flags().Bool("check-eligibility", true, "check block eligibility")
	_ = workflowsBlocksSimulateCmd.MarkFlagRequired("run-id")
	_ = workflowsBlocksSimulateCmd.MarkFlagRequired("block-id")

	workflowsBlocksCmd.AddCommand(workflowsBlocksListCmd, workflowsBlocksGetCmd, workflowsBlocksResolvedSchemasCmd, workflowsBlocksCreateCmd, workflowsBlocksCreateBatchCmd, workflowsBlocksUpdateCmd, workflowsBlocksDeleteCmd, workflowsBlocksSimulateCmd)
	workflowsCmd.AddCommand(workflowsBlocksCmd)
}
