package cmd

import (
	"context"
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
// flow through the same `edit` / `approve` pair.
//
// Every mutating command carries a `--version-stamp` (the overlay `rev`,
// a compare-and-swap token). When omitted, the CLI reads the current `rev`
// first and warns on stderr — convenient for scripts, but it gives up
// protection against a concurrent reviewer. Pass `--version-stamp`
// explicitly to make the write fail loudly (HTTP 409) on a stale read.

var workflowsReviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Review block runs awaiting review",
	Long: `Drive the review overlay: list the review queue, inspect a
block run's output history, post corrections, and submit a verdict
(approve / reject).

Each block run awaiting review has a review overlay — an append-only log of every
output version, every actor who touched it, and every decision. The
overlay's ` + "`rev`" + ` is a compare-and-swap token: mutating commands
take ` + "`--version-stamp`" + ` and fail with HTTP 409 if another reviewer
moved the overlay first.`,
	Example: `  # See what's waiting for review
  retab workflows reviews list

  # Inspect one block run awaiting review
  retab workflows reviews get run_xyz789 blk_extract_1

  # Approve it as-is
  retab workflows reviews approve run_xyz789 blk_extract_1 --version-stamp 0`,
}

var workflowsReviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List block runs awaiting review",
	Long: `List the review queue — block runs awaiting review and their lifecycle,
hottest first. The heavy version/decision/audit history is omitted; pull
one item with ` + "`reviews get`" + ` to see it.`,
	Example: `  # The whole org's awaiting-review queue
  retab workflows reviews list

  # Only one workflow, only reviews you have claimed
  retab workflows reviews list --workflow-id wf_abc123 --mine`,
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
		params.Status, _ = cmd.Flags().GetString("status")
		params.Mine, _ = cmd.Flags().GetBool("mine")
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
	Long: `Return the full review overlay: every output version (the
model's original is seq 0), every decision, the audit trail, and the
` + "`rev`" + ` CAS token you pass back as ` + "`--version-stamp`" + `.`,
	Example: `  # Inspect a block run awaiting review
  retab workflows reviews get run_xyz789 blk_extract_1

  # Read the current version_stamp for a follow-up decision
  retab workflows reviews get run_xyz789 blk_extract_1 | jq .rev`,
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

var workflowsReviewsApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve a reviewed block output",
	Long: `Approve the reviewed output so the run resumes. With
` + "`--snapshot-file`" + `, pass a full block output snapshot: the
entire JSON output object that should be stored and shipped downstream.
With ` + "`--reviewable-value-file`" + `, pass only the primitive-specific
reviewable value, such as an extract JSON object or classifier category
object; the server compiles it into the full snapshot. Either form is
appended as a corrective version first, then approved. The model's
original is preserved as seq 0 for audit.`,
	Example: `  # Approve as-is
  retab workflows reviews approve run_xyz789 blk_extract_1 --version-stamp 2

  # Approve with a full block output snapshot correction
  retab workflows reviews approve run_xyz789 blk_extract_1 \
    --version-stamp 2 --snapshot-file ./fixed-output.json

  # Approve with a primitive-specific reviewable value
  retab workflows reviews approve run_xyz789 blk_extract_1 \
    --version-stamp 2 --reviewable-value-file ./fixed-value.json`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		snapshotPath, _ := cmd.Flags().GetString("snapshot-file")
		editedOutputPath, _ := cmd.Flags().GetString("edited-output-file")
		reviewableValuePath, _ := cmd.Flags().GetString("reviewable-value-file")
		if snapshotPath != "" && editedOutputPath != "" {
			return fmt.Errorf("--snapshot-file and --edited-output-file are aliases; pass only one")
		}
		snapshotFlagName := "--snapshot-file"
		if snapshotPath == "" && editedOutputPath != "" {
			snapshotPath = editedOutputPath
			snapshotFlagName = "--edited-output-file"
		}
		if snapshotPath != "" && reviewableValuePath != "" {
			return fmt.Errorf("%s and --reviewable-value-file are mutually exclusive", snapshotFlagName)
		}
		var editedOutput map[string]any
		if snapshotPath != "" {
			data, err := readJSONMap(snapshotPath)
			if err != nil {
				return fmt.Errorf("%s: %w", snapshotFlagName, err)
			}
			editedOutput = data
		}
		var reviewableValue map[string]any
		if reviewableValuePath != "" {
			data, err := readJSONMap(reviewableValuePath)
			if err != nil {
				return fmt.Errorf("--reviewable-value-file: %w", err)
			}
			reviewableValue = data
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		req := retab.ApproveReviewRequest{VersionStamp: stamp}
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		req.EditedOutput = editedOutput
		req.ReviewableValue = reviewableValue
		req.OnSeq = optionalIntFlag(cmd, "on-seq")
		req.EffectiveSeq = optionalIntFlag(cmd, "effective-seq")
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
rejected and marks the workflow run error/rejected; downstream blocks do
not continue. A ` + "`--reason`" + ` is required so the review decision is
auditable.`,
	Example: `  retab workflows reviews reject run_xyz789 blk_extract_1 \
    --version-stamp 2 --reason "wrong document type — packing slip, not invoice"`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		reason, err := requireNonBlankFlag(cmd, "reason")
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		req := retab.RejectReviewRequest{VersionStamp: stamp, Reason: reason}
		req.CommandID, _ = cmd.Flags().GetString("command-id")
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
Use ` + "`reviews approve`" + `, ` + "`reviews reject`" + `, or ` + "`reviews edit`" + `.`,
	Example: "",
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("review escalation is not supported by the review overlay API; use reviews approve, reviews reject, or reviews edit")
	}),
}

var workflowsReviewsEditCmd = &cobra.Command{
	Use:   "edit <run-id> <block-id>",
	Short: "Append a corrective output version without deciding",
	Long: `Append a new output version to the overlay's history, leaving
it awaiting review. Use this to record a correction for another reviewer
to approve. ` + "`--snapshot-file`" + ` must contain a full block output snapshot:
the entire JSON output object, not a patch. ` + "`--reviewable-value-file`" + `
contains only the primitive-specific reviewable value, such as an extract
JSON object or classifier category object; the server compiles it into
the full snapshot. Pass exactly one of those files.`,
	Example: `  # Edit with a full block output snapshot
  retab workflows reviews edit run_xyz789 blk_extract_1 \
    --version-stamp 1 --snapshot-file ./corrected-output.json --note "fixed currency"

  # Edit with a primitive-specific reviewable value
  retab workflows reviews edit run_xyz789 blk_extract_1 \
    --version-stamp 1 --reviewable-value-file ./corrected-value.json`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		snapshotPath, _ := cmd.Flags().GetString("snapshot-file")
		reviewableValuePath, _ := cmd.Flags().GetString("reviewable-value-file")
		if snapshotPath == "" && reviewableValuePath == "" {
			return fmt.Errorf("one of --snapshot-file or --reviewable-value-file is required")
		}
		if snapshotPath != "" && reviewableValuePath != "" {
			return fmt.Errorf("--snapshot-file and --reviewable-value-file are mutually exclusive")
		}
		var snapshot map[string]any
		if snapshotPath != "" {
			data, err := readJSONMap(snapshotPath)
			if err != nil {
				return fmt.Errorf("--snapshot-file: %w", err)
			}
			snapshot = data
		}
		var reviewableValue map[string]any
		if reviewableValuePath != "" {
			data, err := readJSONMap(reviewableValuePath)
			if err != nil {
				return fmt.Errorf("--reviewable-value-file: %w", err)
			}
			reviewableValue = data
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		req := retab.EditReviewRequest{VersionStamp: stamp}
		req.Snapshot = snapshot
		req.ReviewableValue = reviewableValue
		req.Origin, _ = cmd.Flags().GetString("origin")
		req.Note, _ = cmd.Flags().GetString("note")
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Reviews.Edit(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printReviewOverlayResult(cmd, result)
	}),
}

var workflowsReviewsClaimCmd = &cobra.Command{
	Use:   "claim <run-id> <block-id>",
	Short: "Take the advisory review claim",
	Long: `Take the advisory claim on a review ("Dana is reviewing this").
A claim is never a lock — it only powers the UI; correctness still rests
on the ` + "`--version-stamp`" + ` CAS. Claims expire after ` + "`--ttl-seconds`" + `.`,
	Example: `  retab workflows reviews claim run_xyz789 blk_extract_1 --version-stamp 0`,
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		ttl, _ := cmd.Flags().GetInt("ttl-seconds")
		result, err := client.Workflows.Reviews.Claim(ctx, args[0], args[1], stamp, ttl)
		if err != nil {
			return err
		}
		return printReviewOverlayResult(cmd, result)
	}),
}

var workflowsReviewsReleaseCmd = &cobra.Command{
	Use:     "release <run-id> <block-id>",
	Short:   "Release the advisory review claim",
	Long:    `Clear the advisory review claim so another reviewer can take it.`,
	Example: `  retab workflows reviews release run_xyz789 blk_extract_1 --version-stamp 1`,
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		result, err := client.Workflows.Reviews.Release(ctx, args[0], args[1], stamp)
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
			} else if overlay.Status == "awaiting_review" {
				return printReviewOverlayResult(cmd, overlay)
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("run %s block %s was not awaiting_review within %ds", args[0], args[1], timeout)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}),
}

// resolveVersionStamp returns the explicit --version-stamp when the user set
// it, or reads the overlay's current rev and warns. Passing the flag
// explicitly is the clearest path for scripts because the read/review/write
// sequence is visible to the operator.
func resolveVersionStamp(cmd *cobra.Command, client *retab.Client, ctx context.Context, runID, blockID string) (int, error) {
	if cmd.Flags().Changed("version-stamp") {
		return cmd.Flags().GetInt("version-stamp")
	}
	overlay, err := client.Workflows.Reviews.Get(ctx, runID, blockID)
	if err != nil {
		return 0, err
	}
	fmt.Fprintf(os.Stderr,
		"warning: --version-stamp not set; using current rev %d "+
			"(pass --version-stamp explicitly in scripts after `reviews get`)\n", overlay.Rev)
	return overlay.Rev, nil
}

// optionalIntFlag returns a pointer to the flag value only when the user set
// it — so an unset --on-seq / --effective-seq is omitted from the request.
func optionalIntFlag(cmd *cobra.Command, name string) *int {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	v, err := cmd.Flags().GetInt(name)
	if err != nil {
		return nil
	}
	return &v
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
		}{Data: []*retab.ReviewOverlay{result}}, reviewQueueColumns)
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
			"_id":               result.Overlay.ID,
			"workflow_run_id":   result.Overlay.WorkflowRunID,
			"block_id":          result.Overlay.BlockID,
			"block_type":        result.Overlay.BlockType,
			"status":            result.Overlay.Status,
			"rev":               result.Overlay.Rev,
			"head_seq":          result.Overlay.HeadSeq,
			"effective_seq":     result.Overlay.EffectiveSeq,
		}
		return RenderList(os.Stdout, OutputTable, struct {
			Data []map[string]any `json:"data"`
		}{Data: []map[string]any{row}}, reviewDecisionColumns)
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
	{Header: "STATUS", Extract: func(row any) string { return reviewQueueCell(row, "status") }},
	{Header: "REV", Extract: func(row any) string { return reviewQueueCell(row, "rev") }},
	{Header: "HEAD", Extract: func(row any) string { return reviewQueueCell(row, "head_seq") }},
	{Header: "EFFECTIVE", Extract: func(row any) string { return reviewQueueCell(row, "effective_seq") }},
}

var reviewDecisionColumns = []TableColumn{
	{Header: "SUBMISSION", Extract: func(row any) string { return reviewQueueCell(row, "submission_status") }},
	{Header: "REVIEW_ID", Extract: func(row any) string { return reviewQueueCell(row, "_id") }},
	{Header: "RUN_ID", Extract: func(row any) string { return reviewQueueCell(row, "workflow_run_id") }},
	{Header: "BLOCK_ID", Extract: func(row any) string { return reviewQueueCell(row, "block_id") }},
	{Header: "BLOCK_TYPE", Extract: func(row any) string { return reviewQueueCell(row, "block_type") }},
	{Header: "STATUS", Extract: func(row any) string { return reviewQueueCell(row, "status") }},
	{Header: "REV", Extract: func(row any) string { return reviewQueueCell(row, "rev") }},
	{Header: "HEAD", Extract: func(row any) string { return reviewQueueCell(row, "head_seq") }},
	{Header: "EFFECTIVE", Extract: func(row any) string { return reviewQueueCell(row, "effective_seq") }},
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

func init() {
	workflowsReviewsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsReviewsListCmd.Flags().Var(
		newEnumStringFlagValue("--status", "awaiting_review", "approved", "rejected"),
		"status", "lifecycle filter: awaiting_review | approved | rejected")
	workflowsReviewsListCmd.Flags().Bool("mine", false, "only reviews you have claimed")
	workflowsReviewsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")

	workflowsReviewsApproveCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read current rev when omitted")
	workflowsReviewsApproveCmd.Flags().String("snapshot-file", "", "JSON file with the full corrected output (or - for stdin)")
	workflowsReviewsApproveCmd.Flags().String("edited-output-file", "", "legacy alias for --snapshot-file")
	_ = workflowsReviewsApproveCmd.Flags().MarkHidden("edited-output-file")
	workflowsReviewsApproveCmd.Flags().String("reviewable-value-file", "", "JSON file with the primitive-specific reviewable value (or - for stdin)")
	workflowsReviewsApproveCmd.Flags().Int("on-seq", 0, "version seq the decision is made against")
	workflowsReviewsApproveCmd.Flags().Int("effective-seq", 0, "version seq to ship downstream")
	workflowsReviewsApproveCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsReviewsRejectCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read current rev when omitted")
	workflowsReviewsRejectCmd.Flags().String("reason", "", "why the output was rejected (required)")
	workflowsReviewsRejectCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsReviewsRejectCmd.MarkFlagRequired("reason")

	workflowsReviewsEscalateCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read current rev when omitted")
	workflowsReviewsEscalateCmd.Flags().String("reason", "", "why the review is escalated (required)")
	workflowsReviewsEscalateCmd.Flags().String("escalate-to", "", "target queue/team id (required)")
	workflowsReviewsEscalateCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsReviewsEscalateCmd.MarkFlagRequired("reason")
	_ = workflowsReviewsEscalateCmd.MarkFlagRequired("escalate-to")

	workflowsReviewsEditCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read current rev when omitted")
	workflowsReviewsEditCmd.Flags().String("snapshot-file", "", "JSON file with the full corrected output — or - for stdin")
	workflowsReviewsEditCmd.Flags().String("reviewable-value-file", "", "JSON file with the primitive-specific reviewable value — or - for stdin")
	workflowsReviewsEditCmd.Flags().Var(
		newEnumStringFlagValue("--origin", "human_edit", "agent_edit"),
		"origin", "edit provenance: human_edit | agent_edit")
	workflowsReviewsEditCmd.Flags().String("note", "", "free-text rationale for the edit")
	workflowsReviewsEditCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsReviewsClaimCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read current rev when omitted")
	workflowsReviewsClaimCmd.Flags().Var(&minIntFlagValue{value: "900", min: 30}, "ttl-seconds", "how long the advisory claim holds (minimum 30)")

	workflowsReviewsReleaseCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read current rev when omitted")

	workflowsReviewsWaitCmd.Flags().Var(&positiveIntFlagValue{value: "120"}, "timeout", "max seconds to wait until review is required")
	workflowsReviewsWaitCmd.Flags().Var(&positiveIntFlagValue{value: "2"}, "poll-interval", "seconds between polls")

	workflowsReviewsCmd.AddCommand(
		workflowsReviewsListCmd,
		workflowsReviewsGetCmd,
		workflowsReviewsApproveCmd,
		workflowsReviewsRejectCmd,
		workflowsReviewsEscalateCmd,
		workflowsReviewsEditCmd,
		workflowsReviewsClaimCmd,
		workflowsReviewsReleaseCmd,
		workflowsReviewsWaitCmd,
	)
	workflowsCmd.AddCommand(workflowsReviewsCmd)
}
