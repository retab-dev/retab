package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// The `workflows reviews` command group drives the review overlay —
// the versioned sidecar attached to a block run awaiting review. The surface is
// actor-neutral: a correction proposed by a model, an agent, or a human all
// flow through the same create-version / approve pair.
//
// Decisions name the exact content-addressed version id being approved or
// rejected. Created versions append a complete snapshot under versions_by_id.

var workflowsReviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Review block runs awaiting review",
	Long: `Drive the review overlay: list the review queue, inspect a
block run's output history, post corrections, and submit a verdict
(approve / reject).

Each block run awaiting review has a review overlay. The overlay stores
immutable output snapshots in ` + "`versions_by_id`" + ` and one terminal
` + "`decision`" + `. Decisions approve or reject one exact ` + "`version_id`" + `.`,
	Example: `  # See what's waiting for review
  retab workflows reviews list

  # Inspect one block run awaiting review
  retab workflows reviews get run_xyz789 blk_extract_1

  # Approve one exact version
  retab workflows reviews approve run_xyz789 blk_extract_1 --version-id 9d6d...`,
}

var workflowsReviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List block runs awaiting review",
	Long: `List the review queue — block runs awaiting review and their lifecycle,
hottest first. Version history and the terminal decision payload are omitted; pull
one item with ` + "`reviews get`" + ` to see it.

Use ` + "`--decision`" + ` to control which slice of the queue is returned.
` + "`--decision none`" + ` (the default) returns the open queue: overlays still
awaiting review. ` + "`--decision any`" + ` returns every overlay, including ones
already approved or rejected, so an operator can review past decisions.`,
	Example: `  # The whole org's awaiting-review queue
  retab workflows reviews list

  # Only one workflow
  retab workflows reviews list --workflow-id wf_abc123

  # Include decided overlays in the listing
  retab workflows reviews list --decision any`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := &retab.ListReviewsParams{}
		params.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		params.Decision, _ = cmd.Flags().GetString("decision")
		if cmd.Flags().Changed("limit") {
			params.Limit, _ = cmd.Flags().GetInt("limit")
		}
		result, err := client.Workflows.Reviews.List(ctx, params)
		if err != nil {
			return err
		}
		return printReviewQueueResult(cmd, result)
	}),
}

var workflowsReviewsGetCmd = &cobra.Command{
	Use:   "get <run-id> <block-id>",
	Short: "Get the full review overlay for a block run",
	Long: `Return the full review overlay: immutable output versions in
` + "`versions_by_id`" + ` plus the terminal ` + "`decision`" + ` when one exists.
Use ` + "`reviews schema`" + ` when you need the snapshot shape for a correction.`,
	Example: `  # Inspect a block run awaiting review
  retab workflows reviews get run_xyz789 blk_extract_1

  # Pick a version id to approve
  retab workflows reviews get run_xyz789 blk_extract_1 | jq '.versions_by_id | keys'

  # See the JSON shape expected by reviews versions create
  retab workflows reviews schema run_xyz789 blk_extract_1`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Reviews.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printReviewOverlayResult(cmd, result)
	}),
}

type reviewSnapshotSchema struct {
	BlockType   string         `json:"block_type"`
	Snapshot    map[string]any `json:"snapshot_schema"`
	Example     map[string]any `json:"example"`
	Notes       []string       `json:"notes,omitempty"`
	CreateUsage string         `json:"create_usage"`
}

var workflowsReviewsSchemaCmd = &cobra.Command{
	Use:   "schema <run-id> <block-id>",
	Short: "Show the snapshot shape accepted by reviews versions create",
	Long: `Show the exact JSON snapshot shape accepted by ` + "`reviews versions create`" + ` for
this reviewed block. The command fetches the review overlay to read the stored
` + "`block_type`" + `, then prints the block-specific snapshot contract.

This is guidance for authoring ` + "`--snapshot-file`" + `. It does not add fields
to the review overlay and it does not change the stored review object.`,
	Example: `  # See what corrected-output.json must contain
  retab workflows reviews schema run_xyz789 blk_extract_1

  # Machine-readable schema guidance
  retab workflows reviews schema run_xyz789 blk_classifier_1 --output json`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		overlay, err := client.Workflows.Reviews.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		schema, err := reviewSchemaForBlockType(overlay.BlockType)
		if err != nil {
			return err
		}
		schema.CreateUsage = fmt.Sprintf(
			"retab workflows reviews versions create %s %s --parent-id <version_id> --snapshot-file <snapshot.json>",
			args[0],
			args[1],
		)
		return printReviewSchemaResult(cmd, schema)
	}),
}

var workflowsReviewsApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve a reviewed block output",
	Long: `Approve one exact output version so the run resumes. To approve a
correction, first create it with ` + "`reviews versions create`" + `, then approve the
returned content-addressed version id.`,
	Example: `  retab workflows reviews approve run_xyz789 blk_extract_1 --version-id 9d6d...`,
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		versionID, err := requireReviewVersionIDFlag(cmd, "version-id")
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.ApproveReviewRequest{VersionID: versionID}
		result, err := client.Workflows.Reviews.Approve(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printReviewDecisionResult(cmd, result)
	}),
}

var workflowsReviewsRejectCmd = &cobra.Command{
	Use:   "reject <run-id> <block-id>",
	Short: "Reject a reviewed block output",
	Long: `Reject the reviewed output. The runtime records the review as
rejected and downstream blocks do not continue. A ` + "`--reason`" + ` is required so the review decision is
auditable.`,
	Example: `  retab workflows reviews reject run_xyz789 blk_extract_1 \
    --version-id 9d6d... --reason "wrong document type — packing slip, not invoice"`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		reason, err := requireNonBlankFlag(cmd, "reason")
		if err != nil {
			return err
		}
		versionID, err := requireReviewVersionIDFlag(cmd, "version-id")
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.RejectReviewRequest{VersionID: versionID, Reason: reason}
		result, err := client.Workflows.Reviews.Reject(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printReviewDecisionResult(cmd, result)
	}),
}

var workflowsReviewsEscalateCmd = &cobra.Command{
	Use:    "escalate <run-id> <block-id>",
	Short:  "Unsupported legacy review escalation command",
	Hidden: true,
	Long: `Review escalation is not supported by the review overlay API.
Use ` + "`reviews approve`" + `, ` + "`reviews reject`" + `, or ` + "`reviews versions create`" + `.`,
	Example: "",
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("review escalation is not supported by the review overlay API; use reviews approve, reviews reject, or reviews versions create")
	}),
}

var workflowsReviewsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage immutable review output versions",
	Long: `Manage immutable output versions inside a review overlay.

Use ` + "`reviews versions create`" + ` to append a complete corrected snapshot.
Existing versions are never mutated.`,
}

var workflowsReviewsVersionsCreateCmd = &cobra.Command{
	Use:   "create <run-id> <block-id>",
	Short: "Create an immutable review output version",
	Long: `Create a new immutable output version in the overlay's version history,
leaving it awaiting review. ` + "`--snapshot-file`" + ` must contain a full block
output snapshot: the entire JSON object, not a patch. ` + "`--parent-id`" + `
sets the ` + "`parent_id`" + ` field to the version this snapshot derives from.

Snapshot shapes are block-specific:
  extract:    ` + "`{\"invoice_number\":\"INV-1\",\"total\":1200}`" + `
  classifier: ` + "`{\"category\":\"invoice\"}`" + `
  split:      ` + "`{\"documents\":[{\"name\":\"booking confirmation\",\"pages\":[1,2]}]}`" + `
  for_each:   ` + "`{\"partitions\":[{\"key\":\"invoice\",\"pages\":[1,2]}]}`" + `

Run ` + "`reviews schema <run-id> <block-id>`" + ` to print the snapshot contract
for a paused block.`,
	Example: `  # Create a corrected output version from a full snapshot
  retab workflows reviews versions create run_xyz789 blk_extract_1 \
    --parent-id 2b8a... --snapshot-file ./corrected-output.json --note "fixed currency"

  # Pipe a classifier correction
  printf '{"category":"booking_confirmation"}' |
    retab workflows reviews versions create run_xyz789 blk_classifier_1 \
      --parent-id 2b8a... --snapshot-file - --origin human_created`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		snapshotPath, _ := cmd.Flags().GetString("snapshot-file")
		if snapshotPath == "" {
			return fmt.Errorf("--snapshot-file is required")
		}
		var snapshot map[string]any
		data, err := readJSONMap(snapshotPath)
		if err != nil {
			return fmt.Errorf("--snapshot-file: %w", err)
		}
		snapshot = data
		parentID, err := requireReviewVersionIDFlag(cmd, "parent-id")
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.CreateReviewVersionRequest{ParentID: parentID}
		req.Snapshot = snapshot
		req.Origin, _ = cmd.Flags().GetString("origin")
		req.Note, _ = cmd.Flags().GetString("note")
		result, err := client.Workflows.Reviews.CreateVersion(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printReviewOverlayResult(cmd, result)
	}),
}

var workflowsReviewsWaitCmd = &cobra.Command{
	Use:   "wait <run-id> <block-id>",
	Short: "Poll until a block run is awaiting review",
	Long: `Poll the overlay until the block run is awaiting review, then
print it. A 404 (the run has not reached the review point yet) is
not an error — polling continues until ` + "`--timeout`" + `. If the workflow
run becomes terminal before the overlay exists, wait exits immediately with
the run status and message.`,
	Example: `  # Block a script until review is required, then inspect it
  retab workflows reviews wait run_xyz789 blk_extract_1 --timeout 300`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		timeout, _ := cmd.Flags().GetInt("timeout")
		interval, _ := cmd.Flags().GetInt("poll-interval")
		deadline := time.Now().Add(time.Duration(timeout) * time.Second)
		for {
			ctx, cancel := ctxFor(cmd)
			overlay, err := client.Workflows.Reviews.Get(ctx, args[0], args[1])
			cancel()
			if err != nil {
				if apiErr, ok := err.(*retab.APIError); !ok || apiErr.StatusCode != 404 {
					return err
				}
				runCtx, runCancel := ctxFor(cmd)
				run, runErr := client.Workflows.Runs.Get(runCtx, args[0])
				runCancel()
				if runErr != nil {
					if apiErr, ok := runErr.(*retab.APIError); ok && apiErr.StatusCode == 404 {
						return fmt.Errorf("run %s was not found", args[0])
					}
					if apiErr, ok := runErr.(*retab.APIError); !ok || apiErr.StatusCode != 404 {
						return runErr
					}
				} else if run.Terminal() {
					message := run.Lifecycle.Message
					if message == "" {
						message = "no review overlay exists"
					}
					return fmt.Errorf("run %s reached terminal status %s before block %s was awaiting_review: %s", args[0], run.Lifecycle.Status, args[1], message)
				}
			} else if overlay.Decision == nil {
				return printReviewOverlayResult(cmd, overlay)
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("run %s block %s was not awaiting_review within %ds", args[0], args[1], timeout)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}),
}

func printReviewQueueResult(cmd *cobra.Command, result *retab.ReviewQueueResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		return RenderList(os.Stdout, OutputTable, result, reviewQueueColumns)
	}
	return printJSON(result)
}

func printReviewOverlayResult(cmd *cobra.Command, result *retab.ReviewOverlay) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		return RenderList(os.Stdout, OutputTable, struct {
			Data []*retab.ReviewOverlay `json:"data"`
		}{Data: []*retab.ReviewOverlay{result}}, reviewOverlayColumns)
	}
	return printJSON(result)
}

func printReviewDecisionResult(cmd *cobra.Command, result *retab.SubmitReviewDecisionResponse) error {
	format, err := resolveReviewOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		row := map[string]any{
			"submission_status": result.SubmissionStatus,
			"_id":               result.Review.ID,
			"workflow_run_id":   result.Review.WorkflowRunID,
			"block_id":          result.Review.BlockID,
			"block_type":        result.Review.BlockType,
			"resume_status":     result.ResumeStatus,
		}
		if result.ResumeError != nil {
			row["resume_error"] = *result.ResumeError
		}
		if result.Review.Decision != nil {
			row["verdict"] = result.Review.Decision.Verdict
			row["version_id"] = result.Review.Decision.VersionID
		}
		return RenderList(os.Stdout, OutputTable, struct {
			Data []map[string]any `json:"data"`
		}{Data: []map[string]any{row}}, reviewDecisionColumns)
	}
	return printJSON(result)
}

func printReviewSchemaResult(cmd *cobra.Command, result reviewSnapshotSchema) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		rows := []map[string]any{
			{
				"block_type":   result.BlockType,
				"schema":       result.Snapshot,
				"example":      result.Example,
				"notes":        result.Notes,
				"create_usage": result.CreateUsage,
			},
		}
		return RenderList(os.Stdout, OutputTable, struct {
			Data []map[string]any `json:"data"`
		}{Data: rows}, reviewSchemaColumns)
	}
	return printJSON(result)
}

func resolveReviewOutputFormat(cmd *cobra.Command, w io.Writer) (OutputFormat, error) {
	format, err := ResolveOutputFormat(cmd, w)
	if err != nil {
		return "", err
	}
	if format == OutputTable {
		return format, nil
	}
	if cmd != nil && cmd.Root().PersistentFlags().Lookup("output") == nil {
		if f := rootCmd.PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputTable) {
			return OutputTable, nil
		}
	}
	return format, nil
}

var reviewQueueColumns = []TableColumn{
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "_id") }},
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "workflow_run_id") }},
	{Header: "BLOCK_ID", Extract: func(row any) string { return reviewQueueCell(row, "block_id") }},
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "AWAITING_SINCE", Extract: func(row any) string { return reviewQueueCell(row, "awaiting_since") }},
	{Header: "PRIORITY", Extract: func(row any) string { return reviewQueueCell(row, "priority") }},
	{Header: "TRIGGERED_BY", Extract: func(row any) string { return reviewQueueCell(row, "triggered_by.kind") }},
}

// reviewOverlayColumns extends reviewQueueColumns with the terminal decision
// snapshot. `reviews get` can return either an open overlay (decision is nil)
// or a decided overlay; the VERDICT and DECIDED_VERSION_ID columns stay empty
// for open overlays so the table still renders cleanly.
var reviewOverlayColumns = []TableColumn{
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "_id") }},
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "workflow_run_id") }},
	{Header: "BLOCK_ID", Extract: func(row any) string { return reviewQueueCell(row, "block_id") }},
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "AWAITING_SINCE", Extract: func(row any) string { return reviewQueueCell(row, "awaiting_since") }},
	{Header: "PRIORITY", Extract: func(row any) string { return reviewQueueCell(row, "priority") }},
	{Header: "TRIGGERED_BY", Extract: func(row any) string { return reviewQueueCell(row, "triggered_by.kind") }},
	{Header: "VERDICT", Extract: func(row any) string { return reviewQueueCell(row, "decision.verdict") }},
	{Header: "DECIDED_VERSION_ID", Extract: func(row any) string {
		return truncateReviewCell(reviewQueueCell(row, "decision.version_id"), 12)
	}},
}

var reviewDecisionColumns = []TableColumn{
	{Header: "SUBMISSION", Extract: func(row any) string { return reviewQueueCell(row, "submission_status") }},
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "_id") }},
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "workflow_run_id") }},
	{Header: "BLOCK_ID", Extract: func(row any) string { return reviewQueueCell(row, "block_id") }},
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "VERDICT", Extract: func(row any) string { return reviewQueueCell(row, "verdict") }},
	{Header: "VERSION_ID", Extract: func(row any) string { return reviewQueueCell(row, "version_id") }},
	{Header: "RESUME_STATUS", Extract: func(row any) string { return reviewQueueCell(row, "resume_status") }},
	{Header: "RESUME_ERROR", Extract: func(row any) string {
		return truncateReviewCell(reviewQueueCell(row, "resume_error"), 60)
	}},
}

var reviewSchemaColumns = []TableColumn{
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "SCHEMA", Extract: func(row any) string { return reviewQueueCell(row, "schema") }},
	{Header: "EXAMPLE", Extract: func(row any) string { return reviewQueueCell(row, "example") }},
	{Header: "NOTES", Extract: func(row any) string { return reviewQueueCell(row, "notes") }},
	{Header: "CREATE_USAGE", Extract: func(row any) string { return reviewQueueCell(row, "create_usage") }},
}

func reviewSchemaForBlockType(blockType string) (reviewSnapshotSchema, error) {
	switch blockType {
	case "extract":
		return reviewSnapshotSchema{
			BlockType: blockType,
			Snapshot: map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			},
			Example: map[string]any{
				"invoice_number": "INV-1",
				"total":          1200,
			},
			Notes: []string{
				"Submit the extract output object itself.",
				"Do not wrap it in an output field and do not submit a patch.",
				"The allowed fields are the fields configured on the extract block.",
			},
		}, nil
	case "classifier":
		return reviewSnapshotSchema{
			BlockType: blockType,
			Snapshot: map[string]any{
				"type":                 "object",
				"required":             []string{"category"},
				"additionalProperties": false,
				"properties": map[string]any{
					"category": map[string]any{"type": "string"},
				},
			},
			Example: map[string]any{
				"category": "booking-confirmation",
			},
			Notes: []string{
				"Submit only the selected category.",
				"The category must exactly match one configured classifier category.",
				"Do not include confidence fields or other metadata.",
			},
		}, nil
	case "split":
		return reviewSnapshotSchema{
			BlockType: blockType,
			Snapshot: map[string]any{
				"type":                 "object",
				"required":             []string{"documents"},
				"additionalProperties": false,
				"properties": map[string]any{
					"documents": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":                 "object",
							"required":             []string{"name", "pages"},
							"additionalProperties": false,
							"properties": map[string]any{
								"name":  map[string]any{"type": "string"},
								"pages": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
							},
						},
					},
				},
			},
			Example: map[string]any{
				"documents": []map[string]any{
					{"name": "legal-mentions", "pages": []int{1}},
					{"name": "booking-confirmation", "pages": []int{2, 3}},
				},
			},
			Notes: []string{
				"pages are 1-based positive integers.",
				"pages must be sorted ascending with no duplicates inside one document.",
				"Submit the complete split list, not only the changed document.",
			},
		}, nil
	case "for_each":
		return reviewSnapshotSchema{
			BlockType: blockType,
			Snapshot: map[string]any{
				"type":                 "object",
				"required":             []string{"partitions"},
				"additionalProperties": false,
				"properties": map[string]any{
					"partitions": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":                 "object",
							"required":             []string{"key", "pages"},
							"additionalProperties": false,
							"properties": map[string]any{
								"key":   map[string]any{"type": "string"},
								"pages": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
							},
						},
					},
				},
			},
			Example: map[string]any{
				"partitions": []map[string]any{
					{"key": "booking-confirmation", "pages": []int{1, 2}},
					{"key": "legal-mentions", "pages": []int{3}},
				},
			},
			Notes: []string{
				"for_each review is only available for split-by-key blocks.",
				"pages are 1-based positive integers.",
				"pages must be sorted ascending with no duplicates inside one partition.",
				"If the for_each block has allow_overlap=false, one page cannot appear in more than one partition.",
				"Submit the complete partition list, not only the changed partition.",
			},
		}, nil
	default:
		return reviewSnapshotSchema{}, fmt.Errorf("block type %q is not reviewable", blockType)
	}
}

// truncateReviewCell shortens long string cells so the table view stays
// readable. An empty input stays empty; otherwise the first `limit` runes
// are kept and an ellipsis is appended when truncation actually happened.
// JSON output never goes through this — the raw fields stay intact.
func truncateReviewCell(value string, limit int) string {
	if value == "" || limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

func reviewQueueCell(row any, key string) string {
	v, ok := rowField(row, key)
	if !ok || cellIsEmpty(v) {
		return ""
	}
	if ptr, ok := v.(*int); ok {
		if ptr == nil {
			return ""
		}
		return fmt.Sprintf("%d", *ptr)
	}
	return stringifyCell(v)
}

func requireReviewVersionIDFlag(cmd *cobra.Command, name string) (string, error) {
	value, err := requireNonBlankFlag(cmd, name)
	if err != nil {
		return "", err
	}
	if !sha256HexPattern.MatchString(value) {
		return "", fmt.Errorf("--%s must be a 64-character hex content id", name)
	}
	return value, nil
}

func init() {
	workflowsReviewsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsReviewsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")
	decisionFlag := newEnumStringFlagValue("--decision", "none", "any")
	_ = decisionFlag.Set("none")
	workflowsReviewsListCmd.Flags().Var(decisionFlag, "decision", "review slice to list: none (open queue, default) | any (include decided overlays)")

	workflowsReviewsApproveCmd.Flags().String("version-id", "", "64-character content-addressed version id to approve (required)")
	_ = workflowsReviewsApproveCmd.MarkFlagRequired("version-id")

	workflowsReviewsRejectCmd.Flags().String("version-id", "", "64-character content-addressed version id to reject (required)")
	workflowsReviewsRejectCmd.Flags().String("reason", "", "why the output was rejected (required)")
	_ = workflowsReviewsRejectCmd.MarkFlagRequired("version-id")
	_ = workflowsReviewsRejectCmd.MarkFlagRequired("reason")

	workflowsReviewsVersionsCreateCmd.Flags().String("parent-id", "", "64-character content-addressed parent_id for the new version (required)")
	workflowsReviewsVersionsCreateCmd.Flags().String("snapshot-file", "", "JSON file with the corrected block snapshot — or - for stdin (required)")
	workflowsReviewsVersionsCreateCmd.Flags().Var(
		newEnumStringFlagValue("--origin", "human_created", "agent_created"),
		"origin", "version provenance: human_created | agent_created")
	workflowsReviewsVersionsCreateCmd.Flags().String("note", "", "free-text rationale for the version")
	_ = workflowsReviewsVersionsCreateCmd.MarkFlagRequired("parent-id")
	_ = workflowsReviewsVersionsCreateCmd.MarkFlagRequired("snapshot-file")

	workflowsReviewsWaitCmd.Flags().Var(&positiveIntFlagValue{value: "120"}, "timeout", "max seconds to wait until review is required")
	workflowsReviewsWaitCmd.Flags().Var(&positiveIntFlagValue{value: "2"}, "poll-interval", "seconds between polls")

	workflowsReviewsVersionsCmd.AddCommand(workflowsReviewsVersionsCreateCmd)
	workflowsReviewsCmd.AddCommand(
		workflowsReviewsListCmd,
		workflowsReviewsGetCmd,
		workflowsReviewsSchemaCmd,
		workflowsReviewsApproveCmd,
		workflowsReviewsRejectCmd,
		workflowsReviewsEscalateCmd,
		workflowsReviewsVersionsCmd,
		workflowsReviewsWaitCmd,
	)
	workflowsCmd.AddCommand(workflowsReviewsCmd)
}
