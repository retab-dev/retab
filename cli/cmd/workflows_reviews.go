package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// The `workflows reviews` command group drives the review overlay —
// the versioned sidecar attached to a gated block run. The surface is
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
	Short: "Review gated block runs (review-based overlay)",
	Long: `Drive the review overlay: list the review queue, inspect a
gated block run's output history, post corrections, and submit a verdict
(approve / reject).

Setup happens on the block config: add ` + "`config.hil`" + ` to an
` + "`extract`" + `, ` + "`split`" + `, ` + "`classifier`" + `, or split-by-key
` + "`for_each`" + ` block. When
the gate predicate fires, the run status is
` + "`waiting_for_human`" + ` and this command group is the review surface.

Each gated block run has a review overlay — an append-only log of every
output version, every actor who touched it, and every decision. The
overlay's ` + "`rev`" + ` is a compare-and-swap token: mutating commands
take ` + "`--version-stamp`" + ` and fail with HTTP 409 if another reviewer
moved the overlay first.`,
	Example: `  # See what's waiting for review
  retab workflows reviews list

  # Inspect one paused block run
  retab workflows reviews get run_xyz789 blk_extract_1

  # Approve it as-is
  retab workflows reviews approve run_xyz789 blk_extract_1 --version-stamp 0

  # Add a gate before running: review whenever extraction misses a required field
  printf '%s\n' '{"hil":{"predicate":{"kind":"any_required_field_null"}}}' \
    > extract-hil.json
  retab workflows blocks update wf_abc123 blk_extract_1 \
    --merge-config-file ./extract-hil.json`,
}

var workflowsReviewsExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Review extract values",
}

var workflowsReviewsExtractApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve an extract reviewable value",
	Long: `Approve an extract block with a schema-shaped JSON value.

The value is the extracted JSON object itself, not the full output envelope.`,
	Example: `  retab workflows reviews extract approve run_123 blk_extract \
    --value-file ./fields.json

  retab workflows reviews extract approve run_123 blk_extract \
    --set invoice.total=123.45 --set invoice.vendor='"Acme"'`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		value, err := extractReviewableValueFromFlags(cmd)
		if err != nil {
			return err
		}
		return approveReviewableValue(cmd, args[0], args[1], value)
	}),
}

var workflowsReviewsSplitCmd = &cobra.Command{
	Use:   "split",
	Short: "Review split page manifests",
}

var workflowsReviewsSplitApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve a split reviewable value",
	Long: `Approve a split block with a split manifest.

Use --set "name=1-3,5" for inline edits or --splits-json / --value-file
for the JSON shape {"splits":[{"name":"invoice","pages":[1]}]}.

Split labels are case-sensitive and sent exactly as entered. The CLI checks
only local shape errors such as duplicate labels or pages; configured-label
validation is enforced by the backend.`,
	Example: `  retab workflows reviews split approve run_123 blk_split \
    --set "booking confirmation=1" \
    --set "legal-mentions=2"

  retab workflows reviews split approve run_123 blk_split \
    --splits-json '{"splits":[{"name":"invoice","pages":[1,2]}]}'`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		value, err := splitReviewableValueFromFlags(cmd)
		if err != nil {
			return err
		}
		return approveReviewableValue(cmd, args[0], args[1], value)
	}),
}

var workflowsReviewsClassifierCmd = &cobra.Command{
	Use:   "classifier",
	Short: "Review classifier categories",
}

var workflowsReviewsClassifierApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve a classifier category",
	Example: `  retab workflows reviews classifier approve run_123 blk_classifier \
    --category "Invoice"`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		category, err := requireNonBlankFlag(cmd, "category")
		if err != nil {
			return err
		}
		return approveReviewableValue(cmd, args[0], args[1], map[string]any{
			"category": category,
		})
	}),
}

var workflowsReviewsForEachCmd = &cobra.Command{
	Use:   "for-each",
	Short: "Review for_each partition chunks",
}

var workflowsReviewsForEachApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve a split_by_key for_each reviewable value",
	Long: `Approve a split_by_key for_each map-phase review with partition chunks.

Use --set "key=1-3,5" for inline edits or --chunks-json / --value-file
for the JSON shape {"chunks":[{"key":"booking confirmation","pages":[1]}]}.
Partition keys passed through --set cannot contain "="; use --chunks-json or
--value-file for those keys.

This command is only for for_each blocks configured with map_method=split_by_key.
The CLI checks local shape errors such as duplicate keys or duplicate pages
inside one chunk. Different keys may share a page when several items appear on
the same source page; the backend validates the block config and document page
bounds.`,
	Example: `  retab workflows reviews for-each approve run_123 blk_for_each \
    --set "legal-mentions=1" \
    --set "booking confirmation=2-3"

  retab workflows reviews for-each approve run_123 blk_for_each \
    --chunks-json '{"chunks":[{"key":"legal-mentions","pages":[1]}]}'`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		value, err := partitionReviewableValueFromFlags(cmd)
		if err != nil {
			return err
		}
		return approveReviewableValue(cmd, args[0], args[1], value)
	}),
}

var workflowsReviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List block runs awaiting review",
	Long: `List the review queue — gated block runs and their lifecycle,
hottest first. The heavy version/decision/audit history is omitted; pull
one item with ` + "`reviews get`" + ` to see it.`,
	Example: `  # The whole org's awaiting-review queue
  retab workflows reviews list

  # Find runs that are paused because a review is required
  retab workflows runs list --status waiting_for_human

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
		return printResult(cmd, result)
	}),
}

var workflowsReviewsGetCmd = &cobra.Command{
	Use:   "get <run-id> <block-id>",
	Short: "Get the full review overlay for a gated block run",
	Long: `Return the full review overlay: every output version (the
model's original is seq 0), every decision, the audit trail, and the
` + "`rev`" + ` CAS token you pass back as ` + "`--version-stamp`" + `.
Use the overlay's ` + "`triggered_by`" + ` field to understand why the
gate fired.`,
	Example: `  # Inspect a paused block run
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
		return printJSON(result)
	}),
}

var workflowsReviewsApproveCmd = &cobra.Command{
	Use:   "approve <run-id> <block-id>",
	Short: "Approve a gated block output",
	Long: `Approve the gated output so the run resumes.

With ` + "`--edited-output-file`" + ` or ` + "`--edited-output-json`" + `,
the corrected output is appended as a new version first, then approved —
the model's original is preserved as seq 0 for audit.`,
	Example: `  # Approve as-is
  retab workflows reviews approve run_xyz789 blk_extract_1 --version-stamp 2

  # Approve with a correction from a file
  retab workflows reviews approve run_xyz789 blk_extract_1 \
    --version-stamp 2 --edited-output-file ./fixed.json

  # Approve with a small inline correction
  retab workflows reviews approve run_xyz789 blk_extract_1 \
    --version-stamp 2 --edited-output-json '{"splits":[{"name":"invoice","pages":[1]}]}'`,
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
		if cmd.Flags().Changed("edited-output-file") && cmd.Flags().Changed("edited-output-json") {
			return fmt.Errorf("--edited-output-file and --edited-output-json are mutually exclusive")
		}
		if cmd.Flags().Changed("edited-output-file") {
			path, _ := cmd.Flags().GetString("edited-output-file")
			if strings.TrimSpace(path) == "" {
				return fmt.Errorf("--edited-output-file must not be blank")
			}
			data, err := readJSONMap(path)
			if err != nil {
				return fmt.Errorf("--edited-output-file: %w", err)
			}
			req.EditedOutput = data
		}
		if cmd.Flags().Changed("edited-output-json") {
			raw, _ := cmd.Flags().GetString("edited-output-json")
			if strings.TrimSpace(raw) == "" {
				return fmt.Errorf("--edited-output-json must not be blank")
			}
			data, err := parseJSONMap(raw)
			if err != nil {
				return fmt.Errorf("--edited-output-json: %w", err)
			}
			req.EditedOutput = data
		}
		req.OnSeq = optionalIntFlag(cmd, "on-seq")
		req.EffectiveSeq = optionalIntFlag(cmd, "effective-seq")
		result, err := client.Workflows.Reviews.Approve(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsReviewsRejectCmd = &cobra.Command{
	Use:   "reject <run-id> <block-id>",
	Short: "Reject a gated block output (cancels the run)",
	Long: `Reject the gated output. Rejecting cancels the whole workflow
run — downstream blocks never execute. A ` + "`--reason`" + ` is required
so the cancellation is auditable.`,
	Example: `  retab workflows reviews reject run_xyz789 blk_extract_1 \
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
		result, err := client.Workflows.Reviews.Reject(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsReviewsEditCmd = &cobra.Command{
	Use:   "edit <run-id> <block-id>",
	Short: "Append a corrective output version without deciding",
	Long: `Append a new output version to the overlay's history, leaving
it awaiting review. Use this to record a correction for another reviewer
to approve. The snapshot must be the FULL output object, not a patch.

Pass it with ` + "`--snapshot-file`" + ` for larger payloads or
` + "`--snapshot-json`" + ` for small inline corrections.`,
	Example: `  retab workflows reviews edit run_xyz789 blk_extract_1 \
    --version-stamp 1 --snapshot-file ./corrected.json --note "fixed currency"

  retab workflows reviews edit run_xyz789 blk_extract_1 \
    --version-stamp 1 --snapshot-json '{"splits":[{"name":"invoice","pages":[1]}]}'`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if cmd.Flags().Changed("snapshot-file") && cmd.Flags().Changed("snapshot-json") {
			return fmt.Errorf("--snapshot-file and --snapshot-json are mutually exclusive")
		}
		var snapshot map[string]any
		if cmd.Flags().Changed("snapshot-file") {
			path, _ := cmd.Flags().GetString("snapshot-file")
			if strings.TrimSpace(path) == "" {
				return fmt.Errorf("--snapshot-file must not be blank")
			}
			data, err := readJSONMap(path)
			if err != nil {
				return fmt.Errorf("--snapshot-file: %w", err)
			}
			snapshot = data
		}
		if cmd.Flags().Changed("snapshot-json") {
			raw, _ := cmd.Flags().GetString("snapshot-json")
			if strings.TrimSpace(raw) == "" {
				return fmt.Errorf("--snapshot-json must not be blank")
			}
			data, err := parseJSONMap(raw)
			if err != nil {
				return fmt.Errorf("--snapshot-json: %w", err)
			}
			snapshot = data
		}
		if snapshot == nil {
			return fmt.Errorf("one of --snapshot-file or --snapshot-json is required")
		}
		stamp, err := resolveVersionStamp(cmd, client, ctx, args[0], args[1])
		if err != nil {
			return err
		}
		req := retab.EditReviewRequest{Snapshot: snapshot, VersionStamp: stamp}
		req.Origin, _ = cmd.Flags().GetString("origin")
		req.Note, _ = cmd.Flags().GetString("note")
		req.CommandID, _ = cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Reviews.Edit(ctx, args[0], args[1], req)
		if err != nil {
			return err
		}
		return printJSON(result)
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
		return printJSON(result)
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
		return printJSON(result)
	}),
}

var workflowsReviewsWaitCmd = &cobra.Command{
	Use:   "wait <run-id> <block-id>",
	Short: "Poll until a block run is awaiting review",
	Long: `Poll the overlay until the block run is gated and awaiting
review, then print it. A 404 (the run has not reached the gate yet) is
not an error — polling continues until ` + "`--timeout`" + `.`,
	Example: `  # Block a script until the gate fires, then review
  retab workflows reviews wait run_xyz789 blk_extract_1 --timeout 300

  # Then approve the same gated block output
  retab workflows reviews approve run_xyz789 blk_extract_1 --version-stamp 0`,
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
	overlay, err := client.Workflows.Reviews.Get(ctx, runID, blockID)
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

func approveReviewableValue(cmd *cobra.Command, runID, blockID string, value map[string]any) error {
	client, err := newClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	stamp, err := resolveVersionStamp(cmd, client, ctx, runID, blockID)
	if err != nil {
		return err
	}
	req := retab.ApproveReviewRequest{
		VersionStamp:    stamp,
		ReviewableValue: value,
	}
	req.CommandID, _ = cmd.Flags().GetString("command-id")
	result, err := client.Workflows.Reviews.Approve(ctx, runID, blockID, req)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func extractReviewableValueFromFlags(cmd *cobra.Command) (map[string]any, error) {
	changed := 0
	if cmd.Flags().Changed("value-file") {
		changed++
	}
	if cmd.Flags().Changed("value-json") {
		changed++
	}
	if cmd.Flags().Changed("set") {
		changed++
	}
	if changed != 1 {
		return nil, fmt.Errorf("exactly one of --value-file, --value-json, or --set is required")
	}
	if cmd.Flags().Changed("value-file") {
		path, _ := cmd.Flags().GetString("value-file")
		if strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("--value-file must not be blank")
		}
		return readJSONMap(path)
	}
	if cmd.Flags().Changed("value-json") {
		raw, _ := cmd.Flags().GetString("value-json")
		if strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("--value-json must not be blank")
		}
		return parseJSONMap(raw)
	}
	sets, _ := cmd.Flags().GetStringArray("set")
	value := map[string]any{}
	for _, item := range sets {
		key, rawValue, ok := strings.Cut(item, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("--set must be path=value, got %q", item)
		}
		parsed, err := parseReviewLiteral(rawValue)
		if err != nil {
			return nil, fmt.Errorf("--set %q: %w", item, err)
		}
		if err := setDottedReviewValue(value, strings.TrimSpace(key), parsed); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func splitReviewableValueFromFlags(cmd *cobra.Command) (map[string]any, error) {
	changed := 0
	if cmd.Flags().Changed("value-file") {
		changed++
	}
	if cmd.Flags().Changed("splits-json") {
		changed++
	}
	if cmd.Flags().Changed("set") {
		changed++
	}
	if changed != 1 {
		return nil, fmt.Errorf("exactly one of --value-file, --splits-json, or --set is required")
	}
	if cmd.Flags().Changed("value-file") {
		path, _ := cmd.Flags().GetString("value-file")
		if strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("--value-file must not be blank")
		}
		value, err := readJSONMap(path)
		if err != nil {
			return nil, err
		}
		return value, validateSplitReviewableValue(value)
	}
	if cmd.Flags().Changed("splits-json") {
		raw, _ := cmd.Flags().GetString("splits-json")
		if strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("--splits-json must not be blank")
		}
		value, err := parseJSONMap(raw)
		if err != nil {
			return nil, err
		}
		return value, validateSplitReviewableValue(value)
	}
	sets, _ := cmd.Flags().GetStringArray("set")
	splits := make([]any, 0, len(sets))
	for _, item := range sets {
		name, rawPages, ok := strings.Cut(item, "=")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("--set must be name=pages, got %q", item)
		}
		pages, err := parsePageList(rawPages)
		if err != nil {
			return nil, fmt.Errorf("--set %q: %w", item, err)
		}
		splits = append(splits, map[string]any{
			"name":  name,
			"pages": pages,
		})
	}
	value := map[string]any{"splits": splits}
	return value, validateSplitReviewableValue(value)
}

func partitionReviewableValueFromFlags(cmd *cobra.Command) (map[string]any, error) {
	changed := 0
	if cmd.Flags().Changed("value-file") {
		changed++
	}
	if cmd.Flags().Changed("chunks-json") {
		changed++
	}
	if cmd.Flags().Changed("set") {
		changed++
	}
	if changed != 1 {
		return nil, fmt.Errorf("exactly one of --value-file, --chunks-json, or --set is required")
	}
	if cmd.Flags().Changed("value-file") {
		path, _ := cmd.Flags().GetString("value-file")
		if strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("--value-file must not be blank")
		}
		value, err := readJSONMap(path)
		if err != nil {
			return nil, err
		}
		return value, validatePartitionReviewableValue(value)
	}
	if cmd.Flags().Changed("chunks-json") {
		raw, _ := cmd.Flags().GetString("chunks-json")
		if strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("--chunks-json must not be blank")
		}
		value, err := parseJSONMap(raw)
		if err != nil {
			return nil, err
		}
		return value, validatePartitionReviewableValue(value)
	}
	sets, _ := cmd.Flags().GetStringArray("set")
	chunks := make([]any, 0, len(sets))
	for _, item := range sets {
		key, rawPages, ok := strings.Cut(item, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("--set must be key=pages, got %q", item)
		}
		pages, err := parsePageList(rawPages)
		if err != nil {
			return nil, fmt.Errorf("--set %q: %w", item, err)
		}
		chunks = append(chunks, map[string]any{
			"key":   key,
			"pages": pages,
		})
	}
	value := map[string]any{"chunks": chunks}
	return value, validatePartitionReviewableValue(value)
}

func parseReviewLiteral(raw string) (any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	var value any
	if err := json.Unmarshal([]byte(trimmed), &value); err == nil {
		return value, nil
	}
	return raw, nil
}

func setDottedReviewValue(root map[string]any, path string, value any) error {
	parts := strings.Split(path, ".")
	current := root
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return fmt.Errorf("invalid empty path segment in %q", path)
		}
		if i == len(parts)-1 {
			current[part] = value
			return nil
		}
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	return nil
}

func parsePageList(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	pages := []int{}
	seen := map[int]bool{}
	addPage := func(page int) error {
		if seen[page] {
			return fmt.Errorf("duplicate page %d", page)
		}
		pages = append(pages, page)
		seen[page] = true
		return nil
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if startRaw, endRaw, ok := strings.Cut(part, "-"); ok {
			start, err := strconv.Atoi(strings.TrimSpace(startRaw))
			if err != nil || start < 1 {
				return nil, fmt.Errorf("invalid page range start %q", startRaw)
			}
			end, err := strconv.Atoi(strings.TrimSpace(endRaw))
			if err != nil || end < start {
				return nil, fmt.Errorf("invalid page range end %q", endRaw)
			}
			for page := start; page <= end; page++ {
				if err := addPage(page); err != nil {
					return nil, err
				}
			}
			continue
		}
		page, err := strconv.Atoi(part)
		if err != nil || page < 1 {
			return nil, fmt.Errorf("invalid page %q", part)
		}
		if err := addPage(page); err != nil {
			return nil, err
		}
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("at least one page is required")
	}
	return pages, nil
}

func validateSplitReviewableValue(value map[string]any) error {
	rawSplits, ok := value["splits"].([]any)
	if !ok || len(rawSplits) == 0 {
		return fmt.Errorf("splits must be a non-empty array")
	}
	seenNames := map[string]bool{}
	pageOwners := map[int]string{}
	for index, rawSplit := range rawSplits {
		split, ok := rawSplit.(map[string]any)
		if !ok {
			return fmt.Errorf("splits[%d] must be an object", index)
		}
		nameRaw, ok := split["name"].(string)
		if !ok || strings.TrimSpace(nameRaw) == "" {
			return fmt.Errorf("splits[%d].name must be a non-empty string", index)
		}
		// Preserve the operator's exact label casing and spacing. The backend
		// owns configured-label validation; the CLI only rejects local shape
		// errors that would make the request ambiguous before it is sent.
		name := nameRaw
		if seenNames[name] {
			return fmt.Errorf("duplicate split label %q (labels are case-sensitive and sent exactly as provided; configured-label validation happens on the server)", name)
		}
		seenNames[name] = true
		pages, err := reviewPagesFromValue(split["pages"])
		if err != nil {
			return fmt.Errorf("splits[%d].pages: %w", index, err)
		}
		for _, page := range pages {
			if previousName, ok := pageOwners[page]; ok {
				return fmt.Errorf("page %d appears in both %q and %q", page, previousName, name)
			}
			pageOwners[page] = name
		}
	}
	return nil
}

func validatePartitionReviewableValue(value map[string]any) error {
	rawChunks, ok := value["chunks"].([]any)
	if !ok {
		return fmt.Errorf("chunks must be an array")
	}
	seenKeys := map[string]bool{}
	for index, rawChunk := range rawChunks {
		chunk, ok := rawChunk.(map[string]any)
		if !ok {
			return fmt.Errorf("chunks[%d] must be an object", index)
		}
		keyRaw, ok := chunk["key"].(string)
		if !ok || strings.TrimSpace(keyRaw) == "" {
			return fmt.Errorf("chunks[%d].key must be a non-empty string", index)
		}
		key := strings.TrimSpace(keyRaw)
		chunk["key"] = key
		if seenKeys[key] {
			return fmt.Errorf("duplicate partition key %q (keys are trimmed and case-sensitive)", key)
		}
		seenKeys[key] = true
		if _, err := reviewPagesFromValue(chunk["pages"]); err != nil {
			return fmt.Errorf("chunks[%d].pages: %w", index, err)
		}
	}
	return nil
}

func reviewPagesFromValue(raw any) ([]int, error) {
	if intPages, ok := raw.([]int); ok {
		if len(intPages) == 0 {
			return nil, fmt.Errorf("must be a non-empty array")
		}
		pages := make([]int, 0, len(intPages))
		seen := map[int]bool{}
		for _, page := range intPages {
			if page < 1 {
				return nil, fmt.Errorf("must contain only positive integer pages")
			}
			if seen[page] {
				return nil, fmt.Errorf("duplicate page %d", page)
			}
			seen[page] = true
			pages = append(pages, page)
		}
		return pages, nil
	}
	rawPages, ok := raw.([]any)
	if !ok || len(rawPages) == 0 {
		return nil, fmt.Errorf("must be a non-empty array")
	}
	pages := make([]int, 0, len(rawPages))
	seen := map[int]bool{}
	for _, rawPage := range rawPages {
		page, ok := intFromReviewNumber(rawPage)
		if !ok || page < 1 {
			return nil, fmt.Errorf("must contain only positive integer pages")
		}
		if seen[page] {
			return nil, fmt.Errorf("duplicate page %d", page)
		}
		seen[page] = true
		pages = append(pages, page)
	}
	return pages, nil
}

func intFromReviewNumber(raw any) (int, bool) {
	switch value := raw.(type) {
	case int:
		return value, true
	case float64:
		asInt := int(value)
		return asInt, value == float64(asInt)
	default:
		return 0, false
	}
}

func init() {
	workflowsReviewsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsReviewsListCmd.Flags().Var(
		newEnumStringFlagValue("--status", "awaiting_review", "approved", "rejected"),
		"status", "lifecycle filter: awaiting_review | approved | rejected")
	workflowsReviewsListCmd.Flags().Bool("mine", false, "only reviews you have claimed")
	workflowsReviewsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")

	workflowsReviewsApproveCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token); read by --version-stamp omitted")
	workflowsReviewsApproveCmd.Flags().String("edited-output-file", "", "JSON file with the full corrected output (or - for stdin)")
	workflowsReviewsApproveCmd.Flags().String("edited-output-json", "", "inline JSON object with the full corrected output")
	workflowsReviewsApproveCmd.Flags().Int("on-seq", 0, "version seq the decision is made against")
	workflowsReviewsApproveCmd.Flags().Int("effective-seq", 0, "version seq to ship downstream")
	workflowsReviewsApproveCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsReviewsRejectCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsReviewsRejectCmd.Flags().String("reason", "", "why the output was rejected (required)")
	workflowsReviewsRejectCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsReviewsRejectCmd.MarkFlagRequired("reason")

	workflowsReviewsEditCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsReviewsEditCmd.Flags().String("snapshot-file", "", "JSON file with the full corrected output (or - for stdin)")
	workflowsReviewsEditCmd.Flags().String("snapshot-json", "", "inline JSON object with the full corrected output")
	workflowsReviewsEditCmd.Flags().Var(
		newEnumStringFlagValue("--origin", "human_edit", "agent_edit"),
		"origin", "edit provenance: human_edit | agent_edit")
	workflowsReviewsEditCmd.Flags().String("note", "", "free-text rationale for the edit")
	workflowsReviewsEditCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsReviewsClaimCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsReviewsClaimCmd.Flags().Int("ttl-seconds", 900, "how long the advisory claim holds")

	workflowsReviewsReleaseCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")

	workflowsReviewsWaitCmd.Flags().Int("timeout", 120, "max seconds to wait for the gate to fire")
	workflowsReviewsWaitCmd.Flags().Int("poll-interval", 2, "seconds between polls")

	workflowsReviewsExtractApproveCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsReviewsExtractApproveCmd.Flags().String("value-file", "", "JSON file with the schema-shaped extracted value (or - for stdin)")
	workflowsReviewsExtractApproveCmd.Flags().String("value-json", "", "inline schema-shaped extracted JSON object")
	workflowsReviewsExtractApproveCmd.Flags().StringArray("set", nil, "set an extracted field as path=json_value; repeatable")
	workflowsReviewsExtractApproveCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsReviewsSplitApproveCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsReviewsSplitApproveCmd.Flags().String("value-file", "", "JSON file with the split reviewable value (or - for stdin)")
	workflowsReviewsSplitApproveCmd.Flags().String("splits-json", "", "inline split reviewable value JSON")
	workflowsReviewsSplitApproveCmd.Flags().StringArray("set", nil, "set split pages as name=1-3,5; repeatable")
	workflowsReviewsSplitApproveCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsReviewsClassifierApproveCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsReviewsClassifierApproveCmd.Flags().String("category", "", "approved classifier category (required)")
	workflowsReviewsClassifierApproveCmd.Flags().String("command-id", "", "idempotency command id")
	_ = workflowsReviewsClassifierApproveCmd.MarkFlagRequired("category")

	workflowsReviewsForEachApproveCmd.Flags().Int("version-stamp", 0, "overlay rev (CAS token)")
	workflowsReviewsForEachApproveCmd.Flags().String("value-file", "", "JSON file with the for_each partition reviewable value (or - for stdin)")
	workflowsReviewsForEachApproveCmd.Flags().String("chunks-json", "", "inline for_each partition reviewable value JSON")
	workflowsReviewsForEachApproveCmd.Flags().StringArray("set", nil, "set partition pages as key=1-3,5; repeatable")
	workflowsReviewsForEachApproveCmd.Flags().String("command-id", "", "idempotency command id")

	workflowsReviewsExtractCmd.AddCommand(workflowsReviewsExtractApproveCmd)
	workflowsReviewsSplitCmd.AddCommand(workflowsReviewsSplitApproveCmd)
	workflowsReviewsClassifierCmd.AddCommand(workflowsReviewsClassifierApproveCmd)
	workflowsReviewsForEachCmd.AddCommand(workflowsReviewsForEachApproveCmd)

	workflowsReviewsCmd.AddCommand(
		workflowsReviewsListCmd,
		workflowsReviewsGetCmd,
		workflowsReviewsApproveCmd,
		workflowsReviewsExtractCmd,
		workflowsReviewsSplitCmd,
		workflowsReviewsClassifierCmd,
		workflowsReviewsForEachCmd,
		workflowsReviewsRejectCmd,
		workflowsReviewsEditCmd,
		workflowsReviewsClaimCmd,
		workflowsReviewsReleaseCmd,
		workflowsReviewsWaitCmd,
	)
	workflowsCmd.AddCommand(workflowsReviewsCmd)
}
