package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// Mirrors review_overlay_models.VersionId — base32 of the first 16 bytes of
// sha256(canonical_json(snapshot)). 16 bytes is 26 base32 chars after the
// stripped padding. The `rvr_` prefix disambiguates review version ids from
// workflow version ids, which also use `ver_` but a different shape (hex,
// lowercase, 32+ chars). Reject anything else at the CLI layer so we surface
// a clear flag-shape error before the request leaves the client.
var reviewVersionIDPattern = regexp.MustCompile(`^rvr_[A-Z2-7]{26}$`)

// The `workflows reviews` command group drives reviews for block runs awaiting
// review. The surface is actor-neutral: a correction proposed by a model, an
// agent, or a human all flow through the same create-version / approve pair.
//
// Decisions name the exact content-addressed version id being approved or
// rejected. Created versions add a complete immutable snapshot.

var workflowsReviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Review block runs awaiting review",
	Long: `Drive reviews: list the review queue, inspect a
block run's output history, post corrections, and submit a verdict
(approve / reject).

Each block run awaiting review has review metadata, immutable output versions,
and one terminal ` + "`decision`" + `. Decisions approve or reject one exact
` + "`version_id`" + `. Versions are listed separately with
` + "`reviews versions list <review-id>`" + `.`,
	Example: `  # See what's waiting for review
  retab workflows reviews list

  # Inspect one review
  retab workflows reviews get rev_123

  # Approve one exact version
  retab workflows reviews approve rev_123 --version-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA`,
}

var workflowsReviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List block runs awaiting review",
	Long: `List the review queue — block runs awaiting review and their lifecycle,
oldest-created first. Version history and the terminal decision payload are omitted; pull
one item with ` + "`reviews get`" + ` to see it.

Use ` + "`--decision-status`" + ` to control which slice of the queue is returned.
` + "`--decision-status pending`" + ` (the default) returns the open queue.
Use approved, rejected, decided, or all to inspect past decisions.

Paginate by passing the cursor from a previous response's
` + "`list_metadata`" + `: ` + "`--after`" + ` for the next page,
` + "`--before`" + ` for the previous one. The two are mutually
exclusive.`,
	Example: `  # The whole org's awaiting-review queue
  retab workflows reviews list

  # Only one workflow
  retab workflows reviews list --workflow-id wf_abc123

  # Include every review in the listing
  retab workflows reviews list --decision-status all

  # Paginate: take the cursor from the first page and feed it back as --after
  retab workflows reviews list --after $(retab workflows reviews list --limit 50 --output json | jq -r '.list_metadata.after')`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := &retab.ListReviewsParams{}
		params.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		params.RunID, _ = cmd.Flags().GetString("run-id")
		params.BlockID, _ = cmd.Flags().GetString("block-id")
		params.StepID, _ = cmd.Flags().GetString("step-id")
		params.IterationKey, _ = cmd.Flags().GetString("iteration-key")
		params.DecisionStatus, _ = cmd.Flags().GetString("decision-status")
		params.Before, _ = cmd.Flags().GetString("before")
		params.After, _ = cmd.Flags().GetString("after")
		if params.Before != "" && params.After != "" {
			return fmt.Errorf("--before and --after are mutually exclusive")
		}
		if cmd.Flags().Changed("limit") {
			params.Limit, _ = cmd.Flags().GetInt("limit")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Reviews.List(ctx, params)
		if err != nil {
			return err
		}
		return printReviewQueueResult(cmd, result)
	}),
}

var workflowsReviewsGetCmd = &cobra.Command{
	Use:   "get <review-id>",
	Short: "Get review metadata and decision",
	Long: `Return review metadata plus the terminal ` + "`decision`" + ` when one exists.
Use ` + "`reviews schema`" + ` when you need the snapshot shape for a correction.`,
	Example: `  # Inspect a review
  retab workflows reviews get rev_123

  # Pick a version id to approve
  retab workflows reviews versions list rev_123

  # See the JSON shape expected by reviews versions create
  retab workflows reviews schema rev_123`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Reviews.Get(ctx, args[0])
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
	Use:   "schema <review-id>",
	Short: "Show the snapshot shape accepted by reviews versions create",
	Long: `Show the exact JSON snapshot shape accepted by ` + "`reviews versions create`" + ` for
this review. The command fetches the review to read the stored
` + "`block_type`" + `, then prints the block-specific snapshot contract.

This is guidance for authoring ` + "`--snapshot-file`" + `. It does not add fields
to the review and it does not change the stored review object.`,
	Example: `  # See what corrected-output.json must contain
  retab workflows reviews schema rev_123

  # Machine-readable schema guidance
  retab workflows reviews schema rev_123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		review, err := client.Workflows.Reviews.Get(ctx, args[0])
		if err != nil {
			return err
		}
		schema, err := reviewSchemaForBlockType(review.BlockType)
		if err != nil {
			return err
		}
		schema.CreateUsage = fmt.Sprintf(
			"retab workflows reviews versions create %s --parent-id <rvr_...> --snapshot-file <snapshot.json>",
			args[0],
		)
		return printReviewSchemaResult(cmd, schema)
	}),
}

var workflowsReviewsApproveCmd = &cobra.Command{
	Use:   "approve <review-id>",
	Short: "Approve a reviewed block output",
	Long: `Approve one exact output version so the run resumes. To approve a
correction, first create it with ` + "`reviews versions create`" + `, then approve the
returned content-addressed version id.
Any version returned by ` + "`reviews versions list`" + ` is a valid approval target.`,
	Example: `  retab workflows reviews approve rev_123 --version-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA`,
	Args:    cobra.ExactArgs(1),
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
		result, err := client.Workflows.Reviews.Approve(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printReviewDecisionResult(cmd, result)
	}),
}

var workflowsReviewsRejectCmd = &cobra.Command{
	Use:   "reject <review-id>",
	Short: "Reject a reviewed block output",
	Long: `Reject the reviewed output. The runtime records the review as
rejected and downstream blocks do not continue. A ` + "`--reason`" + ` is required so the review decision is
auditable.`,
	Example: `  retab workflows reviews reject rev_123 \
    --version-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA --reason "wrong document type — packing slip, not invoice"`,
	Args: cobra.ExactArgs(1),
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
		result, err := client.Workflows.Reviews.Reject(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printReviewDecisionResult(cmd, result)
	}),
}

var workflowsReviewsEscalateCmd = &cobra.Command{
	Use:    "escalate <review-id>",
	Short:  "Unsupported legacy review escalation command",
	Hidden: true,
	Long: `Review escalation is not supported by the review API.
Use ` + "`reviews approve`" + `, ` + "`reviews reject`" + `, or ` + "`reviews versions create`" + `.`,
	Example: "",
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("review escalation is not supported by the review API; use reviews approve, reviews reject, or reviews versions create")
	}),
}

var workflowsReviewsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage immutable review output versions",
	Long: `Manage immutable output versions for a review.

Use ` + "`reviews versions create`" + ` to create a complete corrected snapshot.
Existing versions are never mutated.`,
}

var workflowsReviewsVersionsListCmd = &cobra.Command{
	Use:   "list <review-id>",
	Short: "List immutable review output versions",
	Long: `List immutable output versions for one review. The review id is required
because versions are queried per review.`,
	Example: `  retab workflows reviews versions list rev_123
  retab workflows reviews versions list rev_123 --after rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := &retab.ListReviewVersionsParams{ReviewID: args[0]}
		params.Before, _ = cmd.Flags().GetString("before")
		params.After, _ = cmd.Flags().GetString("after")
		if params.Before != "" && params.After != "" {
			return fmt.Errorf("--before and --after are mutually exclusive")
		}
		if cmd.Flags().Changed("limit") {
			params.Limit, _ = cmd.Flags().GetInt("limit")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Reviews.Versions.List(ctx, params)
		if err != nil {
			return err
		}
		return printReviewVersionsResult(cmd, result)
	}),
}

var workflowsReviewsVersionsGetCmd = &cobra.Command{
	Use:     "get <version-id>",
	Short:   "Get one immutable review output version",
	Example: `  retab workflows reviews versions get rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if !reviewVersionIDPattern.MatchString(args[0]) {
			return fmt.Errorf("version-id must be a rvr_<26-char base32> version id")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Reviews.Versions.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printReviewVersionResult(cmd, result)
	}),
}

var workflowsReviewsVersionsCreateCmd = &cobra.Command{
	Use:   "create <review-id>",
	Short: "Create an immutable review output version",
	Long: `Create a new immutable output version in the review's version graph,
leaving it awaiting review. ` + "`--snapshot-file`" + ` must contain a full block
output snapshot: the entire JSON object, not a patch. ` + "`--parent-id`" + `
sets the ` + "`parent_id`" + ` for the version this snapshot derives from.

Snapshot shapes are block-specific:
  extract:    ` + "`{\"invoice_number\":\"INV-1\",\"total\":1200}`" + `
  classifier: ` + "`{\"category\":\"invoice\"}`" + `
  split:      ` + "`{\"documents\":[{\"name\":\"booking confirmation\",\"pages\":[1,2]}]}`" + `
  for_each:   ` + "`{\"partitions\":[{\"key\":\"invoice\",\"pages\":[1,2]}]}`" + `

Run ` + "`reviews schema <review-id>`" + ` to print the snapshot contract.`,
	Example: `  # Create a corrected output version from a full snapshot
  retab workflows reviews versions create rev_123 \
    --parent-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA --snapshot-file ./corrected-output.json --note "fixed currency"

  # Pipe a classifier correction
  printf '{"category":"booking_confirmation"}' |
    retab workflows reviews versions create rev_123 \
      --parent-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA --snapshot-file -`,
	Args: cobra.ExactArgs(1),
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
		req := retab.CreateReviewVersionRequest{
			ReviewID: args[0],
			ParentID: parentID,
			Snapshot: snapshot,
		}
		req.Note, _ = cmd.Flags().GetString("note")
		result, err := client.Workflows.Reviews.Versions.Create(ctx, req)
		if err != nil {
			return err
		}
		return printReviewVersionResult(cmd, result)
	}),
}

var workflowsReviewsWaitCmd = &cobra.Command{
	Use:   "wait <review-id>",
	Short: "Poll until a review is pending",
	Long: `Poll the review until it exists and has no terminal decision, then
print it. A 404 is not an error — polling continues until ` + "`--timeout`" + `.`,
	Example: `  retab workflows reviews wait rev_123 --timeout 300`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		timeout, _ := cmd.Flags().GetInt("timeout")
		interval, _ := cmd.Flags().GetInt("poll-interval")
		deadline := time.Now().Add(time.Duration(timeout) * time.Second)
		// Track the last terminal state we observed so the timeout error can
		// distinguish "the review never appeared" (only 404s) from "the
		// review is already decided" (got a body, but with a terminal
		// decision). Generic "was not pending" is unhelpful — once a review
		// is decided it never becomes pending again, so the caller needs to
		// know which case they're in.
		var lastDecided *retab.Review
		for {
			ctx, cancel := ctxFor(cmd)
			review, err := client.Workflows.Reviews.Get(ctx, args[0])
			cancel()
			if err != nil {
				if apiErr, ok := err.(*retab.APIError); !ok || apiErr.StatusCode != 404 {
					return err
				}
			} else if review.Decision == nil {
				return printReviewOverlayResult(cmd, review)
			} else {
				lastDecided = review
			}
			if time.Now().After(deadline) {
				if lastDecided != nil && lastDecided.Decision != nil {
					return fmt.Errorf(
						"review %s is already %s (decided at %s); wait will never succeed on a decided review",
						args[0],
						lastDecided.Decision.Verdict,
						lastDecided.Decision.DecidedAt.Format(time.RFC3339),
					)
				}
				return fmt.Errorf("review %s did not appear within %ds", args[0], timeout)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}),
}

func printReviewQueueResult(cmd *cobra.Command, result *retab.PaginatedList[retab.ReviewSummary]) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		if err := RenderList(os.Stdout, OutputTable, result, reviewQueueColumns); err != nil {
			return err
		}
		// Cursor-aware footer: the table goes to stdout so it stays
		// pipeable; the truncation hint goes to stderr so jq/awk don't
		// see it. JSON output is unchanged — the cursor is already in
		// the list_metadata field of the printed envelope.
		if result != nil && result.ListMetadata.After != "" {
			fmt.Fprintf(os.Stderr, "… more results available; pass --after %s to continue.\n", result.ListMetadata.After)
		}
		return nil
	}
	return printJSON(result)
}

func printReviewOverlayResult(cmd *cobra.Command, result *retab.Review) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		return RenderList(os.Stdout, OutputTable, struct {
			Data []*retab.Review `json:"data"`
		}{Data: []*retab.Review{result}}, reviewOverlayColumns)
	}
	return printJSON(result)
}

func printReviewVersionsResult(cmd *cobra.Command, result *retab.PaginatedList[retab.ReviewVersion]) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		return RenderList(os.Stdout, OutputTable, result, reviewVersionColumns)
	}
	return printJSON(result)
}

func printReviewVersionResult(cmd *cobra.Command, result *retab.ReviewVersion) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable {
		return RenderList(os.Stdout, OutputTable, struct {
			Data []*retab.ReviewVersion `json:"data"`
		}{Data: []*retab.ReviewVersion{result}}, reviewVersionColumns)
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
			"id":                result.Review.ID,
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
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "id") }},
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "workflow_run_id") }},
	{Header: "BLOCK_ID", Extract: func(row any) string { return reviewQueueCell(row, "block_id") }},
	{Header: "STEP_ID", Extract: func(row any) string { return reviewQueueCell(row, "step_id") }},
	{Header: "ITERATION", Extract: func(row any) string { return reviewQueueCell(row, "iteration_key") }},
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "VERSIONS", Extract: func(row any) string { return reviewQueueCell(row, "version_count") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return reviewQueueCell(row, "created_at") }},
	{Header: "TRIGGERED_BY", Extract: func(row any) string { return reviewQueueCell(row, "triggered_by.kind") }},
}

// reviewOverlayColumns extends reviewQueueColumns with the terminal decision
// snapshot. `reviews get` can return either an open overlay (decision is nil)
// or a decided overlay; the VERDICT and DECIDED_VERSION_ID columns stay empty
// for open overlays so the table still renders cleanly.
var reviewOverlayColumns = []TableColumn{
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "id") }},
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "workflow_run_id") }},
	{Header: "BLOCK_ID", Extract: func(row any) string { return reviewQueueCell(row, "block_id") }},
	{Header: "STEP_ID", Extract: func(row any) string { return reviewQueueCell(row, "step_id") }},
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return reviewQueueCell(row, "created_at") }},
	{Header: "TRIGGERED_BY", Extract: func(row any) string { return reviewQueueCell(row, "triggered_by.kind") }},
	{Header: "VERDICT", Extract: func(row any) string { return reviewQueueCell(row, "decision.verdict") }},
	{Header: "DECIDED_VERSION_ID", Extract: func(row any) string {
		return truncateReviewCell(reviewQueueCell(row, "decision.version_id"), 12)
	}},
}

var reviewDecisionColumns = []TableColumn{
	{Header: "SUBMISSION", Extract: func(row any) string { return reviewQueueCell(row, "submission_status") }},
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "id") }},
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

var reviewVersionColumns = []TableColumn{
	{Header: "VERSION_ID", Extract: func(row any) string { return reviewQueueCell(row, "id") }},
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "review_id") }},
	{Header: "PARENT_ID", Extract: func(row any) string { return reviewQueueCell(row, "parent_id") }},
	{Header: "AUTHOR", Extract: func(row any) string { return reviewQueueCell(row, "author.display_name") }},
	{Header: "AUTHOR_KIND", Extract: func(row any) string { return reviewQueueCell(row, "author.kind") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return reviewQueueCell(row, "created_at") }},
}

var reviewSchemaColumns = []TableColumn{
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "SCHEMA", Extract: func(row any) string { return reviewSchemaJSONCell(row, "schema") }},
	{Header: "EXAMPLE", Extract: func(row any) string { return reviewSchemaJSONCell(row, "example") }},
	{Header: "NOTES", Extract: func(row any) string { return reviewSchemaJSONCell(row, "notes") }},
	{Header: "CREATE_USAGE", Extract: func(row any) string { return reviewQueueCell(row, "create_usage") }},
}

// reviewSchemaJSONCell renders structured values (maps, slices) as compact JSON
// rather than Go's `map[k:v]` debug format, which is illegible in a table cell.
func reviewSchemaJSONCell(row any, key string) string {
	v, ok := rowField(row, key)
	if !ok || cellIsEmpty(v) {
		return ""
	}
	encoded, err := json.Marshal(v)
	if err != nil {
		return reviewQueueCell(row, key)
	}
	return string(encoded)
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
	if !reviewVersionIDPattern.MatchString(value) {
		return "", fmt.Errorf("--%s must be a rvr_<26-char base32> version id", name)
	}
	return value, nil
}

func init() {
	workflowsReviewsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsReviewsListCmd.Flags().String("run-id", "", "filter by workflow run id")
	workflowsReviewsListCmd.Flags().String("block-id", "", "filter by workflow block id")
	workflowsReviewsListCmd.Flags().String("step-id", "", "filter by execution step id")
	workflowsReviewsListCmd.Flags().String("iteration-key", "", "filter by for_each iteration key")
	workflowsReviewsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")
	decisionFlag := newEnumStringFlagValue("--decision-status", "pending", "approved", "rejected", "decided", "all")
	_ = decisionFlag.Set("pending")
	workflowsReviewsListCmd.Flags().Var(decisionFlag, "decision-status", "review slice to list: pending | approved | rejected | decided | all")
	workflowsReviewsListCmd.Flags().String("before", "", "cursor for the previous page (from list_metadata.before; mutually exclusive with --after)")
	workflowsReviewsListCmd.Flags().String("after", "", "cursor for the next page (from list_metadata.after; mutually exclusive with --before)")
	workflowsReviewsListCmd.MarkFlagsMutuallyExclusive("before", "after")

	workflowsReviewsApproveCmd.Flags().String("version-id", "", "rvr_<26-char base32> version id to approve (required)")
	_ = workflowsReviewsApproveCmd.MarkFlagRequired("version-id")

	workflowsReviewsRejectCmd.Flags().String("version-id", "", "rvr_<26-char base32> version id to reject (required)")
	workflowsReviewsRejectCmd.Flags().String("reason", "", "why the output was rejected (required)")
	_ = workflowsReviewsRejectCmd.MarkFlagRequired("version-id")
	_ = workflowsReviewsRejectCmd.MarkFlagRequired("reason")

	workflowsReviewsVersionsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")
	workflowsReviewsVersionsListCmd.Flags().String("before", "", "cursor for the previous page (from list_metadata.before; mutually exclusive with --after)")
	workflowsReviewsVersionsListCmd.Flags().String("after", "", "cursor for the next page (from list_metadata.after; mutually exclusive with --before)")
	workflowsReviewsVersionsListCmd.MarkFlagsMutuallyExclusive("before", "after")

	workflowsReviewsVersionsCreateCmd.Flags().String("parent-id", "", "rvr_<26-char base32> parent version id for the new version (required)")
	workflowsReviewsVersionsCreateCmd.Flags().String("snapshot-file", "", "JSON file with the corrected block snapshot — or - for stdin (required)")
	workflowsReviewsVersionsCreateCmd.Flags().String("note", "", "free-text rationale for the version")
	_ = workflowsReviewsVersionsCreateCmd.MarkFlagRequired("parent-id")
	_ = workflowsReviewsVersionsCreateCmd.MarkFlagRequired("snapshot-file")

	workflowsReviewsWaitCmd.Flags().Var(&positiveIntFlagValue{value: "120"}, "timeout", "max seconds to wait until review is required")
	workflowsReviewsWaitCmd.Flags().Var(&positiveIntFlagValue{value: "2"}, "poll-interval", "seconds between polls")

	workflowsReviewsVersionsCmd.AddCommand(
		workflowsReviewsVersionsListCmd,
		workflowsReviewsVersionsGetCmd,
		workflowsReviewsVersionsCreateCmd,
	)
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
