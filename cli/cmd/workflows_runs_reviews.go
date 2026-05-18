package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// The `workflows runs reviews` command group drives the HIL review overlay —
// the versioned sidecar attached to a gated block run. The surface is
// actor-neutral: a correction proposed by a model, an agent, or a human all
// flow through the same `edit` / `approve` pair.
//
// Every mutating command carries a `--version-stamp` (the overlay `rev`,
// a compare-and-swap token). When omitted, the CLI reads the current `rev`
// first and warns on stderr — convenient for scripts, but it gives up
// protection against a concurrent reviewer. Pass `--version-stamp`
// explicitly to make the write fail loudly (HTTP 409) on a stale read.

var workflowsRunsReviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Review gated block runs (human-in-the-loop overlay)",
	Long: `Drive the HIL review overlay: list the review queue, inspect a
gated block run's output history, post corrections, and submit a verdict
(approve / reject / escalate).

Each gated block run has a review overlay — an append-only log of every
output version, every actor who touched it, and every decision. The
overlay's ` + "`rev`" + ` is a compare-and-swap token: mutating commands
take ` + "`--version-stamp`" + ` and fail with HTTP 409 if another reviewer
moved the overlay first.`,
	Example: `  # See what's waiting for review
  retab workflows runs reviews list

  # Inspect one paused block run
  retab workflows runs reviews get run_xyz789 blk_extract_1

  # Approve it as-is
  retab workflows runs reviews approve run_xyz789 blk_extract_1 --version-stamp 0`,
}

var workflowsRunsReviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List block runs awaiting review",
	Long: `List the review queue — gated block runs and their lifecycle,
hottest first. The heavy version/decision/audit history is omitted; pull
one item with ` + "`reviews get`" + ` to see it.`,
	Example: `  # The whole org's awaiting-review queue
  retab workflows runs reviews list

  # Only one workflow, only reviews you have claimed
  retab workflows runs reviews list --workflow-id wf_abc123 --mine`,
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
		result, err := client.Workflows.Runs.Reviews.List(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsRunsReviewsGetCmd = &cobra.Command{
	Use:   "get <run-id> <block-id>",
	Short: "Get the full review overlay for a gated block run",
	Long: `Return the full review overlay: every output version (the
model's original is seq 0), every decision, the audit trail, and the
` + "`rev`" + ` CAS token you pass back as ` + "`--version-stamp`" + `.`,
	Example: `  # Inspect a paused block run
  retab workflows runs reviews get run_xyz789 blk_extract_1

  # Read the current version_stamp for a follow-up decision
  retab workflows runs reviews get run_xyz789 blk_extract_1 | jq .rev`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.Reviews.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsReviewsApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve a gated block output",
	Long: `Approve the gated output so the run resumes. With
` + "`--edited-output-file`" + ` the file is appended as a corrective
version first, then approved — the model's original is preserved as
seq 0 for audit.`,
	Example: `  # Approve as-is
  retab workflows runs reviews approve run_xyz789 blk_extract_1 --version-stamp 2

  # Approve with a correction
  retab workflows runs reviews approve run_xyz789 blk_extract_1 \
    --version-stamp 2 --edited-output-file ./fixed.json`,
	Args: cobra.ExactArgs(2),
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
		req := retab.ApproveReviewRequest{VersionStamp: stamp}
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		if path, _ := cmd.Flags().GetString("edited-output-file"); path != "" {
			data, err := readJSONMap(path)
			if err != nil {
				return fmt.Errorf("--edited-output-file: %w", err)
			}
			req.EditedOutput = data
		}
		req.OnSeq = optionalIntFlag(cmd, "on-seq")
		req.EffectiveSeq = optionalIntFlag(cmd, "effective-seq")
		result, err := client.Workflows.Runs.Reviews.Approve(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsReviewsRejectCmd = &cobra.Command{
	Use:   "reject <run-id> <block-id>",
	Short: "Reject a gated block output (cancels the run)",
	Long: `Reject the gated output. Rejecting cancels the whole workflow
run — downstream blocks never execute. A ` + "`--reason`" + ` is required
so the cancellation is auditable.`,
	Example: `  retab workflows runs reviews reject run_xyz789 blk_extract_1 \
    --version-stamp 2 --reason "wrong document type — packing slip, not invoice"`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		reason, err := requireNonBlankFlag(cmd, "reason")
		if err != nil {
			return err
		}
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		req := retab.RejectReviewRequest{VersionStamp: stamp, Reason: reason}
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Runs.Reviews.Reject(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsReviewsEscalateCmd = &cobra.Command{
	Use:   "escalate <run-id> <block-id>",
	Short: "Escalate a review to another queue",
	Long: `Escalate the review instead of deciding it. Escalation is
non-terminal — the overlay stays awaiting review, re-routed to the
` + "`--escalate-to`" + ` queue. Both ` + "`--reason`" + ` and
` + "`--escalate-to`" + ` are required.`,
	Example: `  retab workflows runs reviews escalate run_xyz789 blk_extract_1 \
    --version-stamp 2 --reason "needs senior sign-off" --escalate-to queue_senior`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		reason, err := requireNonBlankFlag(cmd, "reason")
		if err != nil {
			return err
		}
		escalateTo, err := requireNonBlankFlag(cmd, "escalate-to")
		if err != nil {
			return err
		}
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		req := retab.EscalateReviewRequest{
			VersionStamp: stamp, Reason: reason, EscalateTo: escalateTo,
		}
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Runs.Reviews.Escalate(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsReviewsEditCmd = &cobra.Command{
	Use:   "edit <run-id> <block-id>",
	Short: "Append a corrective output version without deciding",
	Long: `Append a new output version to the overlay's history, leaving
it awaiting review. Use this to record a correction for another reviewer
to approve. The snapshot must be the FULL output object, not a patch.`,
	Example: `  retab workflows runs reviews edit run_xyz789 blk_extract_1 \
    --version-stamp 1 --snapshot-file ./corrected.json --note "fixed currency"`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		path, err := requireNonBlankFlag(cmd, "snapshot-file")
		if err != nil {
			return err
		}
		snapshot, err := readJSONMap(path)
		if err != nil {
			return fmt.Errorf("--snapshot-file: %w", err)
		}
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		req := retab.EditReviewRequest{Snapshot: snapshot, VersionStamp: stamp}
		req.Origin, _ = cmd.Flags().GetString("origin")
		req.Note, _ = cmd.Flags().GetString("note")
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Runs.Reviews.Edit(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsReviewsClaimCmd = &cobra.Command{
	Use:   "claim <run-id> <block-id>",
	Short: "Take the advisory review claim",
	Long: `Take the advisory claim on a review ("Dana is reviewing this").
A claim is never a lock — it only powers the UI; correctness still rests
on the ` + "`--version-stamp`" + ` CAS. Claims expire after ` + "`--ttl-seconds`" + `.`,
	Example: `  retab workflows runs reviews claim run_xyz789 blk_extract_1 --version-stamp 0`,
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
		result, err := client.Workflows.Runs.Reviews.Claim(ctx, args[0], args[1], stamp, ttl)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsReviewsReleaseCmd = &cobra.Command{
	Use:   "release <run-id> <block-id>",
	Short: "Release the advisory review claim",
	Long:  `Clear the advisory review claim so another reviewer can take it.`,
	Example: `  retab workflows runs reviews release run_xyz789 blk_extract_1 --version-stamp 1`,
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
		result, err := client.Workflows.Runs.Reviews.Release(ctx, args[0], args[1], stamp)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsRunsReviewsWaitCmd = &cobra.Command{
	Use:   "wait <run-id> <block-id>",
	Short: "Poll until a block run is awaiting review",
	Long: `Poll the overlay until the block run is gated and awaiting
review, then print it. A 404 (the run has not reached the gate yet) is
not an error — polling continues until ` + "`--timeout`" + `.`,
	Example: `  # Block a script until the gate fires, then review
  retab workflows runs reviews wait run_xyz789 blk_extract_1 --timeout 300`,
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
			overlay, err := client.Workflows.Runs.Reviews.Get(ctx, args[0], args[1])
			cancel()
			if err != nil {
				if apiErr, ok := err.(*retab.APIError); !ok || apiErr.StatusCode != 404 {
					return err
				}
			} else if overlay.Status == "awaiting_review" {
				return printJSON(overlay)
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
// explicitly is the safe path — the write then fails loudly on a stale read.
func resolveVersionStamp(cmd *cobra.Command, client *retab.Client, ctx context.Context, runID, blockID string) (int, error) {
	if cmd.Flags().Changed("version-stamp") {
		return cmd.Flags().GetInt("version-stamp")
	}
	overlay, err := client.Workflows.Runs.Reviews.Get(ctx, runID, blockID)
	if err != nil {
		return 0, err
	}
	fmt.Fprintf(os.Stderr,
		"warning: --version-stamp not set; using current rev %d "+
			"(no protection against a concurrent reviewer)\n", overlay.Rev)
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

func init() {
	workflowsRunsReviewsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsRunsReviewsListCmd.Flags().Var(
		newEnumStringFlagValue("--status", "awaiting_review", "approved", "rejected"),
		"status", "lifecycle filter: awaiting_review | approved | rejected")
	workflowsRunsReviewsListCmd.Flags().Bool("mine", false, "only reviews you have claimed")
	workflowsRunsReviewsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")

	workflowsRunsReviewsApproveCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read by --version-stamp omitted")
	workflowsRunsReviewsApproveCmd.Flags().String("edited-output-file", "", "JSON file with the full corrected output (or - for stdin)")
	workflowsRunsReviewsApproveCmd.Flags().Int("on-seq", 0, "version seq the decision is made against")
	workflowsRunsReviewsApproveCmd.Flags().Int("effective-seq", 0, "version seq to ship downstream")
	workflowsRunsReviewsApproveCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsRunsReviewsRejectCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsRunsReviewsRejectCmd.Flags().String("reason", "", "why the output was rejected (required)")
	workflowsRunsReviewsRejectCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsRunsReviewsRejectCmd.MarkFlagRequired("reason")

	workflowsRunsReviewsEscalateCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsRunsReviewsEscalateCmd.Flags().String("reason", "", "why the review is escalated (required)")
	workflowsRunsReviewsEscalateCmd.Flags().String("escalate-to", "", "target queue/team id (required)")
	workflowsRunsReviewsEscalateCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsRunsReviewsEscalateCmd.MarkFlagRequired("reason")
	_ = workflowsRunsReviewsEscalateCmd.MarkFlagRequired("escalate-to")

	workflowsRunsReviewsEditCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsRunsReviewsEditCmd.Flags().String("snapshot-file", "", "JSON file with the full corrected output (required) — or - for stdin")
	workflowsRunsReviewsEditCmd.Flags().Var(
		newEnumStringFlagValue("--origin", "human_edit", "agent_edit"),
		"origin", "edit provenance: human_edit | agent_edit")
	workflowsRunsReviewsEditCmd.Flags().String("note", "", "free-text rationale for the edit")
	workflowsRunsReviewsEditCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsRunsReviewsEditCmd.MarkFlagRequired("snapshot-file")

	workflowsRunsReviewsClaimCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsRunsReviewsClaimCmd.Flags().Int("ttl-seconds", 900, "how long the advisory claim holds")

	workflowsRunsReviewsReleaseCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")

	workflowsRunsReviewsWaitCmd.Flags().Int("timeout", 120, "max seconds to wait for the gate to fire")
	workflowsRunsReviewsWaitCmd.Flags().Int("poll-interval", 2, "seconds between polls")

	workflowsRunsReviewsCmd.AddCommand(
		workflowsRunsReviewsListCmd,
		workflowsRunsReviewsGetCmd,
		workflowsRunsReviewsApproveCmd,
		workflowsRunsReviewsRejectCmd,
		workflowsRunsReviewsEscalateCmd,
		workflowsRunsReviewsEditCmd,
		workflowsRunsReviewsClaimCmd,
		workflowsRunsReviewsReleaseCmd,
		workflowsRunsReviewsWaitCmd,
	)
	workflowsRunsCmd.AddCommand(workflowsRunsReviewsCmd)
}
