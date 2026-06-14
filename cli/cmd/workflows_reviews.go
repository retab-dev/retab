//go:build !retab_oagen_cli_workflows_reviews

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

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
	Use:   "list [workflow-id]",
	Short: "List block runs awaiting review",
	Long: `List the review queue — block runs awaiting review and their lifecycle,
oldest-created first. Version history and the terminal decision payload are omitted; pull
one item with ` + "`reviews get`" + ` to see it.

Use ` + "`--decision-status`" + ` to control which slice of the queue is returned.
` + "`--decision-status pending`" + ` (the default) returns the open queue.
Use approved, rejected, decided, or all to inspect past decisions.

Without a workflow id the queue spans the whole workspace; scope it to one
workflow either positionally (` + "`list <workflow-id>`" + `) or with the
` + "`--workflow-id`" + ` flag — the two forms are equivalent and may be
combined when they agree.

Paginate by passing the cursor from a previous response's
` + "`list_metadata`" + `: ` + "`--after`" + ` for the next page,
` + "`--before`" + ` for the previous one. The two are mutually
exclusive.`,
	Example: `  # The whole org's awaiting-review queue
  retab workflows reviews list

  # Only one workflow (positional, matches the rest of workflows)
  retab workflows reviews list wf_abc123

  # Same, with the flag form
  retab workflows reviews list --workflow-id wf_abc123

  # Include every review in the listing
  retab workflows reviews list --decision-status all

  # Paginate: take the cursor from the first page and feed it back as --after
  retab workflows reviews list --after $(retab workflows reviews list --limit 50 --output json | jq -r '.list_metadata.after')`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := &retab.WorkflowReviewsListParams{PaginationParams: collectListParams(cmd)}
		// Scope by workflow positionally OR via --workflow-id (co-equal forms);
		// optional here — no id lists the whole org's review queue.
		workflowID, err := resolveWorkflowScope(cmd, args, false)
		if err != nil {
			return err
		}
		if workflowID != "" {
			params.WorkflowID = ptr(workflowID)
		}
		if runID, _ := cmd.Flags().GetString("run-id"); runID != "" {
			params.RunID = ptr(runID)
		}
		if blockID, _ := cmd.Flags().GetString("block-id"); blockID != "" {
			params.BlockID = ptr(blockID)
		}
		if stepID, _ := cmd.Flags().GetString("step-id"); stepID != "" {
			params.StepID = ptr(stepID)
		}
		if iterationKey, _ := cmd.Flags().GetString("iteration-key"); iterationKey != "" {
			params.IterationKey = ptr(iterationKey)
		}
		if decisionStatus, _ := cmd.Flags().GetString("decision-status"); decisionStatus != "" {
			status := retab.ReviewDecisionStatus(decisionStatus)
			params.DecisionStatus = &status
		}
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
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
		schema, err := reviewSchemaForBlockType(string(review.BlockType))
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
	Long: `Approve one output version so the run resumes. By default this
approves the review's LATEST version, so ` + "`reviews approve <review-id>`" + ` is
enough for the common case. To pin a specific version (e.g. a correction you
created with ` + "`reviews versions create`" + `), pass ` + "`--version-id`" + `.
Any version returned by ` + "`reviews versions list`" + ` is a valid approval target.`,
	Example: `  # Approve the latest version
  retab workflows reviews approve rev_123

  # Approve a specific version
  retab workflows reviews approve rev_123 --version-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateReviewVersionIDFlagIfSet(cmd); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		versionID, err := resolveReviewDecisionVersionID(ctx, client, cmd, args[0])
		if err != nil {
			return err
		}
		req := retab.WorkflowReviewsApproveParams{VersionID: versionID}
		result, err := client.Workflows.Reviews.Approve(ctx, args[0], &req)
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
auditable. By default the review's LATEST version is rejected; pass
` + "`--version-id`" + ` to pin a specific one.`,
	Example: `  retab workflows reviews reject rev_123 --reason "wrong document type — packing slip, not invoice"
  retab workflows reviews reject rev_123 \
    --version-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA --reason "wrong document type"`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		reason, err := requireNonBlankFlag(cmd, "reason")
		if err != nil {
			return err
		}
		if err := validateReviewVersionIDFlagIfSet(cmd); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		versionID, err := resolveReviewDecisionVersionID(ctx, client, cmd, args[0])
		if err != nil {
			return err
		}
		req := retab.WorkflowReviewsRejectParams{VersionID: versionID, Reason: reason}
		result, err := client.Workflows.Reviews.Reject(ctx, args[0], &req)
		if err != nil {
			return err
		}
		return printReviewDecisionResult(cmd, result)
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
		params := &retab.WorkflowReviewVersionsListParams{
			PaginationParams: collectListParams(cmd),
			ReviewID:         args[0],
		}
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
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
		req := retab.WorkflowReviewVersionsCreateParams{
			ReviewID: args[0],
			ParentID: parentID,
			Snapshot: snapshot,
		}
		if note, _ := cmd.Flags().GetString("note"); note != "" {
			req.Note = ptr(note)
		}
		result, err := client.Workflows.Reviews.Versions.Create(ctx, &req)
		if err != nil {
			return err
		}
		return printReviewVersionResult(cmd, result)
	}),
}

func printReviewQueueResult(cmd *cobra.Command, result *retab.PaginatedList[retab.Review]) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		if err := RenderList(os.Stdout, format, result, reviewQueueColumns); err != nil {
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
	if format == OutputTable || format == OutputCSV {
		return RenderList(os.Stdout, format, struct {
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
	if format == OutputTable || format == OutputCSV {
		return RenderList(os.Stdout, format, result, reviewVersionColumns)
	}
	return printJSON(result)
}

func printReviewVersionResult(cmd *cobra.Command, result *retab.ReviewVersion) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		return RenderList(os.Stdout, format, struct {
			Data []*retab.ReviewVersion `json:"data"`
		}{Data: []*retab.ReviewVersion{result}}, reviewVersionColumns)
	}
	return printJSON(result)
}

func printReviewDecisionResult(cmd *cobra.Command, result *retab.SubmitDecisionResponse) error {
	// Conflict means the submission did NOT take effect — the review
	// already had a different decision. Surface this as a non-zero exit
	// so scripts gating on exit code don't mistake "I tried to reject"
	// for "the reject succeeded." The JSON is still printed on stdout
	// (so programmatic consumers see the full payload); the diagnostic
	// goes to stderr and the command exits non-zero via the returned
	// error.
	//
	// SubmissionStatusAccepted and SubmissionStatusAlreadyApplied are
	// both treated as success (already-applied is idempotent: the user
	// asked for X, X is what's recorded).
	conflict := result != nil && result.SubmissionStatus != nil && *result.SubmissionStatus == retab.SubmissionStatusConflict
	format, err := resolveReviewOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}

	// Surface non-resumed engine state to stderr so the user notices
	// when approve/reject succeeded at the API level but the Temporal
	// resume signal didn't land yet ("pending") or was a no-op
	// ("skipped" — e.g. run already terminal). The JSON on stdout still
	// carries the full payload for programmatic consumers; the stderr
	// note is a one-line nudge that a poll on `runs get` is needed.
	if format != OutputTable && result != nil && result.ResumeStatus != nil && *result.ResumeStatus != retab.ResumeStatusResumed {
		switch *result.ResumeStatus {
		case retab.ResumeStatusPending:
			fmt.Fprintf(
				os.Stderr,
				"note: resume_status=%q — decision was accepted but the workflow has not resumed yet. Poll `retab workflows runs get %s` until lifecycle.status changes from awaiting_review.\n",
				*result.ResumeStatus,
				result.Review.RunID,
			)
		case retab.ResumeStatusSkipped:
			fmt.Fprintf(
				os.Stderr,
				"note: resume_status=%q — decision was recorded but no resume signal was sent (the run is likely already terminal).\n",
				*result.ResumeStatus,
			)
		default:
			fmt.Fprintf(
				os.Stderr,
				"note: resume_status=%q — see the JSON response for details.\n",
				*result.ResumeStatus,
			)
		}
		if result.ResumeError != nil && *result.ResumeError != "" {
			fmt.Fprintf(os.Stderr, "note: resume_error: %s\n", *result.ResumeError)
		}
	}
	if format == OutputTable || format == OutputCSV {
		row := map[string]any{
			"submission_status": result.SubmissionStatus,
			"id":                result.Review.ID,
			"run_id":            result.Review.RunID,
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
		if err := RenderList(os.Stdout, format, struct {
			Data []map[string]any `json:"data"`
		}{Data: []map[string]any{row}}, reviewDecisionColumns); err != nil {
			return err
		}
	} else {
		if err := printJSON(result); err != nil {
			return err
		}
	}
	if conflict {
		// Print the diagnostic AFTER the JSON/table so users see both
		// the payload and the reason their scripted approve/reject
		// produced exit 1.
		fmt.Fprintf(
			os.Stderr,
			"error: submission_status=%q — the review already has a different decision; this call did NOT change it. Inspect the current decision with `retab workflows reviews get %s`.\n",
			*result.SubmissionStatus,
			result.Review.ID,
		)
		return fmt.Errorf("review decision conflict: %s already has a different decision", result.Review.ID)
	}
	return nil
}

func printReviewSchemaResult(cmd *cobra.Command, result reviewSnapshotSchema) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		rows := []map[string]any{
			{
				"block_type":   result.BlockType,
				"schema":       result.Snapshot,
				"example":      result.Example,
				"notes":        result.Notes,
				"create_usage": result.CreateUsage,
			},
		}
		return RenderList(os.Stdout, format, struct {
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
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "run_id") }},
	{Header: "BLOCK_ID", Extract: func(row any) string { return reviewQueueCell(row, "block_id") }},
	{Header: "STEP_ID", Extract: func(row any) string { return reviewQueueCell(row, "step_id") }},
	{Header: "ITERATION", Extract: func(row any) string { return reviewQueueCell(row, "iteration_key") }},
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return reviewQueueCell(row, "created_at") }},
	{Header: "TRIGGERED_BY", Extract: func(row any) string { return reviewQueueCell(row, "triggered_by.kind") }},
}

// reviewOverlayColumns extends reviewQueueColumns with the terminal decision
// snapshot. `reviews get` can return either an open overlay (decision is nil)
// or a decided overlay; the VERDICT and DECIDED_VERSION_ID columns stay empty
// for open overlays so the table still renders cleanly.
var reviewOverlayColumns = []TableColumn{
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "id") }},
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "run_id") }},
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
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "run_id") }},
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

// validateReviewVersionIDFlagIfSet rejects a malformed --version-id up-front
// (before any credential/client work) so a bad id fails fast, while leaving the
// flag optional. A blank flag is fine — the decision then defaults to the latest
// version.
func validateReviewVersionIDFlagIfSet(cmd *cobra.Command) error {
	v, _ := cmd.Flags().GetString("version-id")
	if strings.TrimSpace(v) != "" && !reviewVersionIDPattern.MatchString(v) {
		return fmt.Errorf("--version-id must be a rvr_<26-char base32> version id")
	}
	return nil
}

// resolveReviewDecisionVersionID returns the version id an approve/reject acts
// on. An explicit --version-id (rvr_…) wins — pass it to pin a specific version.
// Otherwise it defaults to the review's LATEST version (most recently created),
// so the common case (approve the version you just reviewed) needs no extra
// lookup. Previously --version-id was required, which forced a `reviews versions
// list` round-trip just to discover an id `reviews get` never surfaces. The flag
// format is assumed pre-validated by validateReviewVersionIDFlagIfSet.
func resolveReviewDecisionVersionID(ctx context.Context, client *retab.Client, cmd *cobra.Command, reviewID string) (string, error) {
	if v, _ := cmd.Flags().GetString("version-id"); strings.TrimSpace(v) != "" {
		return v, nil
	}
	versions, err := client.Workflows.Reviews.Versions.List(ctx, &retab.WorkflowReviewVersionsListParams{ReviewID: reviewID})
	if err != nil {
		return "", err
	}
	if len(versions.Data) == 0 {
		return "", fmt.Errorf("review %s has no versions to act on", reviewID)
	}
	latest := versions.Data[0]
	for _, v := range versions.Data[1:] {
		if v.CreatedAt.After(latest.CreatedAt) {
			latest = v
		}
	}
	return latest.ID, nil
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
	// Mutex enforced inside RunE via validateBeforeAfterMutex (concise
	// handwritten message; see workflowsListCmd for the rationale).

	workflowsReviewsApproveCmd.Flags().String("version-id", "", "rvr_<26-char base32> version id to approve (defaults to the review's latest version)")

	workflowsReviewsRejectCmd.Flags().String("version-id", "", "rvr_<26-char base32> version id to reject (defaults to the review's latest version)")
	workflowsReviewsRejectCmd.Flags().String("reason", "", "why the output was rejected (required)")
	_ = workflowsReviewsRejectCmd.MarkFlagRequired("reason")

	workflowsReviewsVersionsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")
	workflowsReviewsVersionsListCmd.Flags().String("before", "", "cursor for the previous page (from list_metadata.before; mutually exclusive with --after)")
	workflowsReviewsVersionsListCmd.Flags().String("after", "", "cursor for the next page (from list_metadata.after; mutually exclusive with --before)")
	// Mutex enforced inside RunE via validateBeforeAfterMutex (concise
	// handwritten message; see workflowsListCmd for the rationale).

	workflowsReviewsVersionsCreateCmd.Flags().String("parent-id", "", "rvr_<26-char base32> parent version id for the new version (required)")
	workflowsReviewsVersionsCreateCmd.Flags().String("snapshot-file", "", "JSON file with the corrected block snapshot — or - for stdin (required)")
	workflowsReviewsVersionsCreateCmd.Flags().String("note", "", "free-text rationale for the version")
	_ = workflowsReviewsVersionsCreateCmd.MarkFlagRequired("parent-id")
	_ = workflowsReviewsVersionsCreateCmd.MarkFlagRequired("snapshot-file")

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
		workflowsReviewsVersionsCmd,
	)
	workflowsCmd.AddCommand(workflowsReviewsCmd)
}
