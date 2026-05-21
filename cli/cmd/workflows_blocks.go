package cmd

import (
	"fmt"
	"net/url"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsBlocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Manage workflow blocks",
	Long: `Add, configure, and inspect the nodes of a workflow graph.

A block is one processing step — ` + "`extract`" + `, ` + "`split`" + `,
` + "`classifier`" + `, ` + "`edit`" + `, ` + "`conditional`" + `,
` + "`api_call`" + `, ` + "`function`" + `, etc. Each block has a typed input,
typed output, and a JSON ` + "`config`" + ` blob shaped by its type.

The workhorse here is ` + "`update`" + `. Once a block is on the graph,
swap its full config with ` + "`workflows blocks update --config-file ./cfg.json`" + `
(REPLACE) or patch a slice with ` + "`--merge-config-file ./patch.json`" + `
(deep merge, RFC 7396) rather than deleting and re-creating. Use review config
inside supported block configs instead of adding a separate review block.`,
	Example: `  # List blocks
  retab workflows blocks list wf_abc123

  # Add a block from a JSON definition
  retab workflows blocks create wf_abc123 --block-file ./extract.json

  # Tune just the config of an existing block
  retab workflows blocks update blk_def456 \
    --config-file ./new-config.json`,
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
	if err := rejectLegacyReviewConfig(req.Config); err != nil {
		return req, err
	}
	// id is optional: when omitted the server generates an opaque
	// `blk_<nanoid>` via default_factory. Client-side requiring a user-chosen
	// id forced collisions because block ids are org-globally unique, and the
	// server's own 409 message points users at server-generated ids — which
	// this client used to make unreachable.
	if req.Type == "" {
		return req, fmt.Errorf("block type is required")
	}
	if req.Type == "review" || req.Type == "hil" {
		return req, fmt.Errorf("standalone review blocks are no longer supported; add config.review to a reviewable block instead")
	}
	return req, nil
}

func parseBlockCreateForWorkflow(workflowID string, obj map[string]any) (retab.WorkflowBlockCreateRequest, error) {
	if bodyWorkflowID, ok := obj["workflow_id"].(string); ok && bodyWorkflowID != "" && bodyWorkflowID != workflowID {
		return retab.WorkflowBlockCreateRequest{}, fmt.Errorf("block-file workflow_id %q does not match positional workflow id %q", bodyWorkflowID, workflowID)
	}
	return parseBlockCreate(obj)
}

func rejectLegacyReviewConfig(config map[string]any) error {
	if config == nil {
		return nil
	}
	if _, ok := config["hil"]; ok {
		return fmt.Errorf("legacy config.hil is no longer supported; use config.review instead")
	}
	return nil
}

func mergeWorkflowBlockConfig(existing map[string]any, patch map[string]any) map[string]any {
	merged := make(map[string]any, len(existing)+len(patch))
	for key, value := range existing {
		merged[key] = value
	}
	for key, patchValue := range patch {
		existingChild, existingIsMap := merged[key].(map[string]any)
		patchChild, patchIsMap := patchValue.(map[string]any)
		if existingIsMap && patchIsMap {
			merged[key] = mergeWorkflowBlockConfig(existingChild, patchChild)
			continue
		}
		merged[key] = patchValue
	}
	return merged
}

var workflowsBlocksListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List blocks in a workflow",
	Long: `List every block in a workflow's draft graph, including id, type,
label, position, and config.`,
	Example: `  # List all blocks
  retab workflows blocks list wf_abc123

  # Get the ids only
  retab workflows blocks list wf_abc123 | jq -r '.data[].id'`,
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
	Use:   "get <block-id>",
	Short: "Get a workflow block",
	Long: `Fetch a single block's full definition: type, label, position,
parent group, and the typed config blob.

Pass ` + "`--workflow-id`" + ` only when the block id is not unique within
your organization — typically pre-uniqueness dev / staging data where
custom block ids (` + "`block_split`" + `, ` + "`block_extract`" + `) were
created in multiple workflows. The server's 409 response on a duplicate
lists the colliding workflow_ids so you know what to pass.`,
	Example: `  # Inspect a block
  retab workflows blocks get blk_def456

  # Disambiguate a legacy duplicate id
  retab workflows blocks get block_split --workflow-id wf_abc123

  # Save a block's config for offline editing
  retab workflows blocks get blk_def456 \
    | jq '.config' > cfg.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		opts := workflowBlockLookupOpts(cmd)
		result, err := client.Workflows.Blocks.Get(ctx, args[0], opts...)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

// workflowBlockLookupOpts builds the optional “workflow_id“ query
// parameter for blocks get/update/delete. Empty flag → no option, so
// the URL stays identical to the original flat-resource shape.
func workflowBlockLookupOpts(cmd *cobra.Command) []retab.RequestOption {
	workflowID, _ := cmd.Flags().GetString("workflow-id")
	workflowID = strings.TrimSpace(workflowID)
	if workflowID == "" {
		return nil
	}
	return []retab.RequestOption{
		retab.WithRequestParams(url.Values{"workflow_id": []string{workflowID}}),
	}
}

var workflowsBlocksCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Create a workflow block from --block-file",
	Long: `Add a block to a workflow's draft graph. The block file is a
JSON object with the keys ` + "`id`" + ` (required), ` + "`type`" + ` (required),
` + "`label`" + `, ` + "`position_x`" + `, ` + "`position_y`" + `, ` + "`width`" + `,
` + "`height`" + `, ` + "`parent_id`" + `, and ` + "`config`" + `.

Block IDs are unique per ORGANIZATION, not per workflow. Reusing a
human-friendly id like ` + "`block_extract`" + ` across two workflows in
the same org will fail with 409. Prefer the server-generated
` + "`blk_<nanoid>`" + ` form (omit ` + "`id`" + ` to get one) unless you
have a reason to pin a stable name.

Review is configured inside the block's typed config as
` + "`config.review.predicate`" + `. For example, an extract block can pause
every run with ` + "`{\"review\":{\"predicate\":{\"kind\":\"always\"}}}`" + `.
Reviewable block types are ` + "`extract`" + `, ` + "`split`" + `, ` + "`classifier`" + `,
and ` + "`for_each`" + ` only when ` + "`config.map_method`" + ` is ` + "`split_by_key`" + ` and
` + "`config.key`" + ` is set.
Common review predicates are ` + "`always`" + ` and ` + "`validation_failed`" + `.
Extract also supports ` + "`any_required_field_null`" + ` and ` + "`field_confidence_lt`" + `;
split and split-by-key ` + "`for_each`" + ` support ` + "`split_count_neq`" + `,
` + "`any_split_pages_lt`" + `, and ` + "`boundary_confidence_lt`" + `; classifier
supports ` + "`category_in`" + ` and ` + "`top_margin_lt`" + `.
Review is not a standalone block type.`,

	Example: `  # Add one block from a JSON file
  retab workflows blocks create wf_abc123 --block-file ./extract.json

  # Minimal extract block with review (replace the id with one unique
  # to your organization, or drop the field entirely to let the server
  # generate an opaque blk_<nanoid>)
  cat > extract-review.json <<'JSON'
  {
    "id": "your-extract-block-id",
    "type": "extract",
    "label": "Extract with review",
    "position_x": 420,
    "position_y": 180,
    "config": {
      "model": "retab-small",
      "inputs": [{"name": "document", "type": "file", "is_primary": true}],
      "json_schema": {
        "type": "object",
        "properties": {
          "document_type": {"type": "string"}
        },
        "required": ["document_type"],
        "additionalProperties": false
      },
      "review": {
        "predicate": {"kind": "always"}
      }
    }
  }
  JSON
  retab workflows blocks create wf_abc123 --block-file ./extract-review.json

  # Pipe a block definition from stdin
  cat block.json | retab workflows blocks create wf_abc123 --block-file -`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("block-file")
		if path == "" {
			return fmt.Errorf("--block-file is required")
		}
		obj, err := readJSONMap(path)
		if err != nil {
			return err
		}
		req, err := parseBlockCreateForWorkflow(args[0], obj)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.Create(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksUpdateCmd = &cobra.Command{
	Use:   "update <block-id>",
	Short: "Update a workflow block",
	Long: `Tune an existing block in place. The two config flags map to
explicit server-side modes:

` + "`--config-file`" + ` REPLACES the typed config — the file is treated
as the full new config and any existing key not in the file is dropped.
Use this for a clean swap (new prompt, new schema, etc.).
Only use ` + "`--config-file`" + ` for review changes when replacing the whole typed config.

` + "`--merge-config-file`" + ` DEEP-MERGES a patch into the existing
config (RFC 7396 JSON Merge Patch). Dicts recurse, arrays/scalars replace,
` + "`null`" + ` deletes the key. Use this when you only want to touch a
slice of the config, for example adding review:

  printf '{"review":{"predicate":{"kind":"always"}}}' |
    retab workflows blocks update BLK --merge-config-file -

Pass ` + "`{\"review\":null}`" + ` to remove review without touching anything else.

The flags are mutually exclusive. Layout fields (` + "`position-*`" + `,
` + "`width`" + `, ` + "`height`" + `, ` + "`parent-id`" + `) only affect
the visual editor.`,
	Example: `  # Swap the config blob
  retab workflows blocks update blk_def456 \
    --config-file ./new-config.json

  # Add review to an existing block
  printf '{"review":{"predicate":{"kind":"always"}}}' |
    retab workflows blocks update blk_def456 --merge-config-file -

  # Rename a block's label
  retab workflows blocks update blk_def456 \
    --label "Extract line items"`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Reject an empty invocation before issuing a no-op PATCH that
		// would round-trip to the server and silently bump updated_at.
		configPath, _ := cmd.Flags().GetString("config-file")
		mergeConfigPath, _ := cmd.Flags().GetString("merge-config-file")
		if configPath != "" && mergeConfigPath != "" {
			return fmt.Errorf("--config-file and --merge-config-file are mutually exclusive")
		}
		if !cmd.Flags().Changed("label") && !cmd.Flags().Changed("position-x") &&
			!cmd.Flags().Changed("position-y") && !cmd.Flags().Changed("width") &&
			!cmd.Flags().Changed("height") && !cmd.Flags().Changed("parent-id") &&
			!cmd.Flags().Changed("config-file") && !cmd.Flags().Changed("merge-config-file") {
			return fmt.Errorf("nothing to update: pass at least one of --label, --position-x, --position-y, --width, --height, --parent-id, --config-file, or --merge-config-file")
		}
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
		if configPath != "" {
			cfg, err := readJSONMap(configPath)
			if err != nil {
				return fmt.Errorf("--config-file: %w", err)
			}
			if err := rejectLegacyReviewConfig(cfg); err != nil {
				return fmt.Errorf("--config-file: %w", err)
			}
			req.Config = cfg
			// Tell the server this is a full replacement, not a merge.
			// Without this, the route keeps any existing keys that aren't
			// in cfg (e.g. ``review``), which silently defeats
			// ``--config-file``'s documented "replace" semantic.
			req.ConfigMode = "replace"
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if mergeConfigPath != "" {
			patch, err := readJSONMap(mergeConfigPath)
			if err != nil {
				return fmt.Errorf("--merge-config-file: %w", err)
			}
			if err := rejectLegacyReviewConfig(patch); err != nil {
				return fmt.Errorf("--merge-config-file: %w", err)
			}
			// Empty patches must be rejected client-side. Go's
			// `json.Marshal` drops `Config map[string]any` via `omitempty`
			// when the map is empty, so we'd send just
			// `{"config_mode":"merge"}` to the server and trip a 422
			// (`config_mode is only meaningful when 'config' is also
			// provided`). That error is opaque to the user — surface the
			// real reason here.
			if len(patch) == 0 {
				return fmt.Errorf("--merge-config-file %s: patch contains no keys; nothing to merge", mergeConfigPath)
			}
			// Send the raw patch and let the server deep-merge it. The
			// route now implements RFC 7396 (dicts recurse, arrays/scalars
			// replace, null deletes the key), which is what the help text
			// has always claimed. Client-side pre-merging would (a) need
			// its own null-as-delete pass to stay consistent and (b)
			// double-merge against pre-config_mode servers in subtle ways
			// — easier to make the server authoritative.
			req.Config = patch
			req.ConfigMode = "merge"
		}
		opts := workflowBlockLookupOpts(cmd)
		result, err := client.Workflows.Blocks.Update(ctx, args[0], req, opts...)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksDeleteCmd = &cobra.Command{
	Use:   "delete <block-id>",
	Short: "Delete a workflow block",
	Long: `Remove a block from the draft graph. Edges that referenced this
block are also deleted. Past runs that used this block remain intact —
deletion only affects the draft.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal.`,
	Example: `  # Remove a block (interactive, asks to confirm)
  retab workflows blocks delete blk_def456

  # Skip the prompt in scripts
  retab workflows blocks delete blk_def456 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "block", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		opts := workflowBlockLookupOpts(cmd)
		if err := client.Workflows.Blocks.Delete(ctx, args[0], opts...); err != nil {
			return err
		}
		confirmDeleted("block", args[0])
		return nil
	}),
}

func init() {
	workflowsBlocksCreateCmd.Flags().String("block-file", "", "JSON file describing the block (or - for stdin)")

	workflowsBlocksUpdateCmd.Flags().String("label", "", "update label")
	workflowsBlocksUpdateCmd.Flags().Float64("position-x", 0, "update X position")
	workflowsBlocksUpdateCmd.Flags().Float64("position-y", 0, "update Y position")
	workflowsBlocksUpdateCmd.Flags().Var(&nonNegativeFloatFlagValue{}, "width", "update width")
	workflowsBlocksUpdateCmd.Flags().Var(&nonNegativeFloatFlagValue{}, "height", "update height")
	workflowsBlocksUpdateCmd.Flags().String("parent-id", "", "update parent id")
	workflowsBlocksUpdateCmd.Flags().String("config-file", "", "JSON file to use as the full new config — REPLACES the existing config (or - for stdin)")
	workflowsBlocksUpdateCmd.Flags().String("merge-config-file", "", "JSON file to deep-merge into the existing config; nulls delete keys (RFC 7396) (or - for stdin)")

	workflowsBlocksDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	// `--workflow-id` is an optional disambiguator for legacy duplicate
	// block ids. New blocks use server-generated opaque ids and never
	// need it; pre-uniqueness data does. See `workflowBlockLookupOpts`.
	workflowsBlocksGetCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksUpdateCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksDeleteCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")

	workflowsBlocksCmd.AddCommand(workflowsBlocksListCmd, workflowsBlocksGetCmd, workflowsBlocksCreateCmd, workflowsBlocksUpdateCmd, workflowsBlocksDeleteCmd)
	workflowsCmd.AddCommand(workflowsBlocksCmd)
}
