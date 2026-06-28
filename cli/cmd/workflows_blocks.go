//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"fmt"
	"maps"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

const (
	defaultNoteBlockWidth     = 280.0
	defaultNoteBlockHeight    = 120.0
	defaultContainerBlockSize = 800.0
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
inside supported block configs.`,
	Example: `  # List blocks
  retab workflows blocks list wf_abc123

  # Add a block from a JSON definition
  retab workflows blocks create wf_abc123 --block-file ./extract.json

  # Tune just the config of an existing block
  retab workflows blocks update block_def456 \
    --config-file ./new-config.json`,
}

func parseBlockCreate(obj map[string]any) (retab.WorkflowBlocksCreateParams, error) {
	req := retab.WorkflowBlocksCreateParams{}
	if v, ok := obj["id"].(string); ok && v != "" {
		req.ID = ptr(v)
	}
	if v, ok := obj["type"].(string); ok {
		req.Type = retab.WorkflowBlockCreateRequestType(v)
	}
	if v, ok := obj["label"].(string); ok {
		req.Label = ptr(v)
	}
	if v, ok := obj["position_x"].(float64); ok {
		req.PositionX = ptr(v)
	}
	if v, ok := obj["position_y"].(float64); ok {
		req.PositionY = ptr(v)
	}
	if v, ok := obj["width"].(float64); ok {
		req.Width = &v
	}
	if v, ok := obj["height"].(float64); ok {
		req.Height = &v
	}
	if v, ok := obj["parent_id"].(string); ok {
		req.ParentID = ptr(v)
	}
	if v, ok := obj["config"].(map[string]any); ok {
		req.Config = &v
	}
	if err := rejectLegacyReviewConfig(req.Config); err != nil {
		return req, err
	}
	// id is optional: when omitted the server generates an opaque
	// `block_<nanoid>` via default_factory. Client-side requiring a user-chosen
	// id forced collisions because block ids are org-globally unique, and the
	// server's own 409 message points users at server-generated ids — which
	// this client used to make unreachable.
	if req.Type == "" {
		return req, fmt.Errorf("block type is required")
	}
	if req.Type == "hil" {
		return req, fmt.Errorf("legacy hil blocks are no longer supported; add config.review to a reviewable block instead")
	}
	defaultBlockCreateDimensions(&req)
	return req, nil
}

func defaultBlockCreateDimensions(req *retab.WorkflowBlocksCreateParams) {
	switch req.Type {
	case "note":
		req.Width = defaultPositiveBlockDimension(req.Width, defaultNoteBlockWidth)
		req.Height = defaultPositiveBlockDimension(req.Height, defaultNoteBlockHeight)
	case "for_each", "while_loop":
		req.Width = clampMinimumBlockDimension(req.Width, defaultContainerBlockSize)
		req.Height = clampMinimumBlockDimension(req.Height, defaultContainerBlockSize)
	}
}

func defaultPositiveBlockDimension(value *float64, fallback float64) *float64 {
	if value != nil && *value > 0 {
		return value
	}
	v := fallback
	return &v
}

func clampMinimumBlockDimension(value *float64, minimum float64) *float64 {
	if value != nil && *value >= minimum {
		return value
	}
	v := minimum
	return &v
}

func parseBlockCreateForWorkflow(workflowID string, obj map[string]any) (retab.WorkflowBlocksCreateParams, error) {
	if bodyWorkflowID, ok := obj["workflow_id"].(string); ok && bodyWorkflowID != "" && bodyWorkflowID != workflowID {
		return retab.WorkflowBlocksCreateParams{}, fmt.Errorf("block-file workflow_id %q does not match positional workflow id %q", bodyWorkflowID, workflowID)
	}
	return parseBlockCreate(obj)
}

func rejectLegacyReviewConfig(config *map[string]any) error {
	if config == nil {
		return nil
	}
	if _, ok := (*config)["hil"]; ok {
		return fmt.Errorf("legacy config.hil is no longer supported; use config.review instead")
	}
	return nil
}

func mergeWorkflowBlockConfig(existing map[string]any, patch map[string]any) map[string]any {
	merged := make(map[string]any, len(existing)+len(patch))
	maps.Copy(merged, existing)
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
	Use:   "list [workflow-id]",
	Short: "List blocks in a workflow",
	Long: `List every block in a workflow's draft graph, including id, type,
label, position, and config.

Name the workflow either positionally (` + "`list <workflow-id>`" + `) or with
the ` + "`--workflow-id`" + ` flag — the two forms are equivalent. Passing both
is accepted when they agree; an error is raised only when they disagree, so a
typo isn't silently masked. The workflow id is required: blocks have no
org-wide listing.

Paginate by passing the cursor from a previous response's
` + "`list_metadata`" + `: ` + "`--after`" + ` for the next page,
` + "`--before`" + ` for the previous one. The two are mutually exclusive.`,
	Example: `  # List all blocks (positional)
  retab workflows blocks list wf_abc123

  # Same, with the flag form
  retab workflows blocks list --workflow-id wf_abc123

  # First page of 50
  retab workflows blocks list wf_abc123 --limit 50

  # Next page, using the cursor from the previous response
  retab workflows blocks list wf_abc123 --limit 50 \
    --after $(retab workflows blocks list wf_abc123 --limit 50 --output json | jq -r '.list_metadata.after')

  # Get the ids only
  retab workflows blocks list wf_abc123 | jq -r '.data[].id'`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// Workflow id positionally OR via --workflow-id (co-equal forms);
		// required here — blocks have no org-wide listing.
		workflowID, err := resolveWorkflowScope(cmd, args, true)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		params := workflowBlocksListParams(cmd, workflowID)
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.List(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func workflowBlocksListParams(cmd *cobra.Command, workflowID string) *retab.WorkflowBlocksListParams {
	return &retab.WorkflowBlocksListParams{
		PaginationParams: collectListParams(cmd),
		WorkflowID:       workflowID,
	}
}

var workflowsBlocksGetCmd = &cobra.Command{
	Use:   "get [<workflow-id>] <block-id>",
	Short: "Get a workflow block",
	Long: `Fetch a single block's full definition: type, label, position,
parent group, and the typed config blob.

Block IDs are unique per organization, so the canonical form is
` + "`get <block-id>`" + `. As a convenience for users whose mental model is
"block X in workflow Y" you can also pass two positionals
(` + "`get <workflow-id> <block-id>`" + `); the workflow id is then used as
the disambiguator described below.

Pass ` + "`--workflow-id`" + ` (or the two-positional form) only when the
block id is not unique within your organization — typically pre-uniqueness
dev / staging data where custom block ids (` + "`block_split`" + `,
` + "`block_extract`" + `) were created in multiple workflows. The server's
409 response on a duplicate lists the colliding workflow_ids so you know
what to pass.`,
	Example: `  # Inspect a block
  retab workflows blocks get block_def456

  # Two-positional convenience form (same effect as --workflow-id)
  retab workflows blocks get wf_abc123 block_def456

  # Disambiguate a legacy duplicate id
  retab workflows blocks get block_split --workflow-id wf_abc123

  # Save a block's config for offline editing
  retab workflows blocks get block_def456 \
    | jq '.config' > cfg.json`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, blockID, err := resolveBlockPositionalWorkflowID(cmd, args)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := &retab.WorkflowBlocksGetParams{WorkflowID: workflowID}
		result, err := client.Workflows.Blocks.Get(ctx, blockID, params)
		if err != nil {
			return err
		}
		// Surface the block's connectable handle ids. Previously the only
		// way to learn an edge's handle names was to inspect an existing
		// workflow's edges; now `blocks get` is self-describing.
		merged, err := primitiveMap(result)
		if err != nil {
			return err
		}
		if handles := deriveBlockHandles(merged); handles != nil {
			merged["handles"] = handles
		}
		return printResult(cmd, merged)
	}),
}

// deriveBlockHandles computes the connectable handle ids for a block from
// its type and config. Input handles are fully determined by config.inputs
// (`input-<type>-<name>`). Output handles are emitted only for block types
// whose single output handle is deterministic; routing blocks (classifier,
// split, conditional, for_each) name their outputs per-route as
// `output-file-<handle_key>` / `output-json-<branch>` — see
// `workflows edges create --help` — and are intentionally not guessed here.
func deriveBlockHandles(block map[string]any) map[string]any {
	blockType, _ := block["type"].(string)
	config, _ := block["config"].(map[string]any)
	inputs := blockInputHandles(config)
	outputs := blockOutputHandles(blockType)
	if len(inputs) == 0 && len(outputs) == 0 {
		return nil
	}
	handles := map[string]any{}
	if len(inputs) > 0 {
		handles["input"] = inputs
	}
	if len(outputs) > 0 {
		handles["output"] = outputs
	}
	return handles
}

func blockInputHandles(config map[string]any) []string {
	if config == nil {
		return nil
	}
	// Function blocks consume a single JSON payload at the runtime-level
	// handle input-json-0. Server-bound function configs do not accept
	// config.inputs; local bundles derive input typing from resolved schemas.
	if _, ok := config["code"].(string); ok {
		if _, hasOutputSchema := config["output_schema"].(map[string]any); hasOutputSchema {
			return []string{"input-json-0"}
		}
	}
	var out []string
	appendInput := func(m map[string]any) {
		name, _ := m["name"].(string)
		t, _ := m["type"].(string)
		if name != "" && t != "" {
			out = append(out, "input-"+t+"-"+name)
		}
	}
	if rawInputs, ok := config["inputs"].([]any); ok {
		for _, ri := range rawInputs {
			if m, ok := ri.(map[string]any); ok {
				appendInput(m)
			}
		}
	}
	if m, ok := config["input"].(map[string]any); ok {
		appendInput(m)
	}
	return out
}

func blockOutputHandles(blockType string) []string {
	switch blockType {
	case "start_document", "edit":
		return []string{"output-file-0"}
	case "start_json", "extract", "merge_dicts", "function", "api_call":
		return []string{"output-json-0"}
	default:
		return nil
	}
}

// workflowBlockLookupWorkflowID returns the optional “workflow_id“ query
// parameter for blocks get/update/delete. Empty flag → nil so the URL stays
// identical to the original flat-resource shape.
func workflowBlockLookupWorkflowID(cmd *cobra.Command) *string {
	workflowID, _ := cmd.Flags().GetString("workflow-id")
	workflowID = strings.TrimSpace(workflowID)
	if workflowID == "" {
		return nil
	}
	return ptr(workflowID)
}

// resolveBlockPositionalWorkflowID lets `blocks get/update/delete` accept
// either `<block-id>` (the canonical, server-resolved form) or
// `<workflow-id> <block-id>` (matches the user mental model of "block X in
// workflow Y" and the shape of `blocks create`/`blocks list`). When the
// two-positional form is used, the workflow id takes the same disambiguator
// path as the `--workflow-id` flag. Returns the resolved (workflowID,
// blockID), where workflowID is nil unless either the flag or the positional
// supplied one. Errors if both are set with conflicting values.
func resolveBlockPositionalWorkflowID(cmd *cobra.Command, args []string) (*string, string, error) {
	flagWorkflowID := workflowBlockLookupWorkflowID(cmd)
	switch len(args) {
	case 1:
		return flagWorkflowID, args[0], nil
	case 2:
		positionalWorkflowID := strings.TrimSpace(args[0])
		blockID := args[1]
		if positionalWorkflowID == "" {
			return nil, "", fmt.Errorf("workflow-id positional argument is empty")
		}
		if flagWorkflowID != nil && *flagWorkflowID != positionalWorkflowID {
			return nil, "", fmt.Errorf("conflicting workflow id: positional %q vs --workflow-id %q", positionalWorkflowID, *flagWorkflowID)
		}
		return ptr(positionalWorkflowID), blockID, nil
	default:
		return nil, "", fmt.Errorf("expected 1 or 2 positional arguments, got %d", len(args))
	}
}

var workflowsBlocksCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Create a workflow block from --block-file",
	Long: `Add a block to a workflow's draft graph. The block file is a
JSON object with the keys ` + "`id`" + ` (optional), ` + "`type`" + ` (required),
` + "`label`" + `, ` + "`position_x`" + `, ` + "`position_y`" + `, ` + "`width`" + `,
` + "`height`" + `, ` + "`parent_id`" + `, and ` + "`config`" + `.

Block IDs are unique per ORGANIZATION, not per workflow. Reusing a
human-friendly id like ` + "`block_extract`" + ` across two workflows in
the same org will fail with 409. Prefer the server-generated
` + "`block_<nanoid>`" + ` form (omit ` + "`id`" + ` to get one) unless you
have a reason to pin a stable name.

Review is configured inside the block's typed config as
` + "`config.review.predicate`" + `. For example, an extract block can pause
every run with ` + "`{\"review\":{\"predicate\":{\"kind\":\"always\"}}}`" + `.
Reviewable block types are ` + "`extract`" + `, ` + "`split`" + `, ` + "`classifier`" + `,
and ` + "`for_each`" + ` only when ` + "`config.map_method`" + ` is ` + "`split_by_key`" + ` and
` + "`config.key`" + ` is set.
Common review predicates are ` + "`always`" + ` and ` + "`validation_failed`" + `.
Extract also supports ` + "`any_required_field_null`" + `, ` + "`confidence_lt`" + `,
` + "`field_confidence_lt`" + `, and ` + "`json_condition`" + `;
split and split-by-key ` + "`for_each`" + ` support ` + "`split_count_neq`" + `,
` + "`any_split_pages_lt`" + `, ` + "`boundary_confidence_lt`" + `, and
` + "`json_condition`" + `; classifier supports ` + "`category_in`" + `,
` + "`confidence_lt`" + `, ` + "`top_margin_lt`" + `, and ` + "`json_condition`" + `.
Consensus criteria require ` + "`n_consensus > 1`" + ` on the reviewed block.
Use ` + "`confidence_lt`" + ` for the block's overall consensus likelihood,
` + "`field_confidence_lt`" + ` for extract field scores, ` + "`top_margin_lt`" + `
for close classifier categories, and ` + "`boundary_confidence_lt`" + ` for split
boundary scores. Numeric predicate fields are type-specific: confidence-style
predicates use ` + "`threshold`" + `, while classifier ` + "`top_margin_lt`" + ` uses
` + "`margin`" + `. ` + "`json_condition`" + ` can target the block output through
` + "`data.*`" + ` (or extract's ` + "`output-json-0.*`" + ` alias) and consensus
scores through ` + "`likelihoods.*`" + ` paths such as
` + "`likelihoods.invoice_total`" + `, ` + "`likelihoods.invoice`" + `, or
` + "`likelihoods.splits.invoice_type`" + `.
Review is not a standalone block type.`,

	Example: `  # Add one block from a JSON file
  retab workflows blocks create wf_abc123 --block-file ./extract.json

  # Minimal extract block with review (replace the id with one unique
  # to your organization, or drop the field entirely to let the server
  # generate an opaque block_<nanoid>)
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

  # Consensus review: pause when the extracted total has low agreement
  cat > extract-consensus-review.json <<'JSON'
  {
    "type": "extract",
    "label": "Extract invoice with consensus review",
    "config": {
      "model": "retab-small",
      "n_consensus": 3,
      "inputs": [{"name": "document", "type": "file", "is_primary": true}],
      "json_schema": {
        "type": "object",
        "properties": {
          "invoice_total": {"type": "number"}
        },
        "required": ["invoice_total"]
      },
      "review": {
        "predicate": {
          "kind": "json_condition",
          "condition": {
            "logical_operator": "and",
            "sub_conditions": [{
              "path_ref": {"input": "default", "path": "likelihoods.invoice_total"},
              "operator": "is_less_than",
              "value": 0.85
            }]
          }
        }
      }
    }
  }
  JSON
  retab workflows blocks create wf_abc123 --block-file ./extract-consensus-review.json

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
		req.WorkflowID = args[0]
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Blocks.Create(ctx, &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksUpdateCmd = &cobra.Command{
	Use:   "update [<workflow-id>] <block-id>",
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

For consensus-based review, patch both ` + "`n_consensus`" + ` and ` + "`review`" + `:

  printf '{"n_consensus":3,"review":{"predicate":{"kind":"confidence_lt","threshold":0.8}}}' |
    retab workflows blocks update BLK --merge-config-file -

For classifier top-margin review, the numeric field is ` + "`margin`" + `:

  printf '{"review":{"predicate":{"kind":"top_margin_lt","margin":0.2}}}' |
    retab workflows blocks update BLK --merge-config-file -

Pass ` + "`{\"review\":null}`" + ` to remove review without touching anything else.

The flags are mutually exclusive. Layout fields (` + "`position-*`" + `,
` + "`width`" + `, ` + "`height`" + `, ` + "`parent-id`" + `) only affect
the visual editor.

Block IDs are unique per organization, so the canonical form is
` + "`update <block-id>`" + `. As a convenience for users whose mental model is
"block X in workflow Y" you can also pass two positionals
(` + "`update <workflow-id> <block-id>`" + `); the workflow id is then wired
through to the same ` + "`--workflow-id`" + ` disambiguator used for legacy
duplicate block ids.`,
	Example: `  # Swap the config blob
  retab workflows blocks update block_def456 \
    --config-file ./new-config.json

  # Two-positional convenience form (same effect as --workflow-id)
  retab workflows blocks update wf_abc123 block_def456 \
    --config-file ./new-config.json

  # Add review to an existing block
  printf '{"review":{"predicate":{"kind":"always"}}}' |
    retab workflows blocks update block_def456 --merge-config-file -

  # Add consensus review to an extract or classifier block
  printf '{"n_consensus":3,"review":{"predicate":{"kind":"confidence_lt","threshold":0.8}}}' |
    retab workflows blocks update block_def456 --merge-config-file -

  # Add classifier top-margin review
  printf '{"review":{"predicate":{"kind":"top_margin_lt","margin":0.2}}}' |
    retab workflows blocks update block_def456 --merge-config-file -

  # Rename a block's label
  retab workflows blocks update block_def456 \
    --label "Extract line items"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, blockID, err := resolveBlockPositionalWorkflowID(cmd, args)
		if err != nil {
			return err
		}
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
		req := retab.WorkflowBlocksUpdateParams{}
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
			if err := rejectLegacyReviewConfig(&cfg); err != nil {
				return fmt.Errorf("--config-file: %w", err)
			}
			req.Config = &cfg
			// Tell the server this is a full replacement, not a merge.
			// Without this, the route keeps any existing keys that aren't
			// in cfg (e.g. ``review``), which silently defeats
			// ``--config-file``'s documented "replace" semantic.
			req.ConfigMode = ptr(retab.UpdateWorkflowBlockRequestConfigModeReplace)
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
			if err := rejectLegacyReviewConfig(&patch); err != nil {
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
			req.Config = &patch
			req.ConfigMode = ptr(retab.UpdateWorkflowBlockRequestConfigModeMerge)
		}
		req.WorkflowID = workflowID
		result, err := client.Workflows.Blocks.Update(ctx, blockID, &req)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsBlocksDeleteCmd = &cobra.Command{
	Use:   "delete [<workflow-id>] <block-id>",
	Short: "Delete a workflow block",
	Long: `Remove a block from the draft graph. Edges that referenced this
block are also deleted. Past runs that used this block remain intact —
deletion only affects the draft.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal.

Block IDs are unique per organization, so the canonical form is
` + "`delete <block-id>`" + `. As a convenience for users whose mental model is
"block X in workflow Y" you can also pass two positionals
(` + "`delete <workflow-id> <block-id>`" + `); the workflow id is then wired
through to the same ` + "`--workflow-id`" + ` disambiguator used for legacy
duplicate block ids.`,
	Example: `  # Remove a block (interactive, asks to confirm)
  retab workflows blocks delete block_def456

  # Two-positional convenience form (same effect as --workflow-id)
  retab workflows blocks delete wf_abc123 block_def456

  # Skip the prompt in scripts
  retab workflows blocks delete block_def456 --yes`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, blockID, err := resolveBlockPositionalWorkflowID(cmd, args)
		if err != nil {
			return err
		}
		if err := confirmDestructive(cmd, "block", blockID); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := &retab.WorkflowBlocksDeleteParams{WorkflowID: workflowID}
		if err := client.Workflows.Blocks.Delete(ctx, blockID, params); err != nil {
			return err
		}
		confirmDeleted("block", blockID)
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

	workflowsBlocksListCmd.Flags().String("workflow-id", "", "workflow id (alternative to the positional form)")
	workflowsBlocksListCmd.Flags().String("before", "", "block id: return the page before this id (mutually exclusive with --after)")
	workflowsBlocksListCmd.Flags().String("after", "", "block id: return the page after this id (mutually exclusive with --before)")
	workflowsBlocksListCmd.MarkFlagsMutuallyExclusive("before", "after")
	workflowsBlocksListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")

	// `--workflow-id` is an optional disambiguator for legacy duplicate
	// block ids. New blocks use server-generated opaque ids and never
	// need it; pre-uniqueness data does. See `workflowBlockLookupOpts`.
	workflowsBlocksGetCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksUpdateCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksDeleteCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")

	workflowsBlocksCmd.AddCommand(workflowsBlocksListCmd, workflowsBlocksGetCmd, workflowsBlocksCreateCmd, workflowsBlocksUpdateCmd, workflowsBlocksDeleteCmd)
	workflowsCmd.AddCommand(workflowsBlocksCmd)
}
