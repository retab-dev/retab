package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// usageBlockRecord mirrors the GET /v1/usage/blocks row (UsageBlockRecord on the
// server). Only usage + operational metadata is present by design.
type usageBlockRecord struct {
	BlockID        string  `json:"block_id"`
	WorkflowID     string  `json:"workflow_id"`
	BlockType      string  `json:"block_type"`
	RunCount       int64   `json:"run_count"`
	ExecutionCount int64   `json:"execution_count"`
	PageCount      int64   `json:"page_count"`
	Credits        float64 `json:"credits"`
	// StatusCounts is the block's execution count broken down by lifecycle status
	// (e.g. {"completed": 190, "failed": 11}); its values sum to ExecutionCount.
	StatusCounts    map[string]int64 `json:"status_counts,omitempty"`
	FirstActivityAt *string          `json:"first_activity_at,omitempty"`
	LastActivityAt  *string          `json:"last_activity_at,omitempty"`
}

// usageBlockListResponse is the GET /v1/usage/blocks envelope. The `data` field
// name lets RenderList extract the row slice for table / CSV output.
type usageBlockListResponse struct {
	Data         []usageBlockRecord   `json:"data"`
	ListMetadata usageRunListMetadata `json:"list_metadata"`
}

var usageBlocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Per-block usage export (credits and pages per workflow block)",
	Long: `List one usage row per workflow block: block type, distinct-run and execution
counts, deduplicated page count, credit spend, and first/last activity.

Scoped to the authenticated organization and environment. Like ` + "`usage runs`" + `
it is confidential-safe — it never exposes document content, filenames, artifact
URIs, or provider/API dollar costs. Only workflow blocks appear; standalone
(API-origin) extractions have no block and are excluded.

Filter by workflow, block type, and activity date range. Page with the opaque
cursors returned in ` + "`list_metadata.before`" + ` / ` + "`list_metadata.after`" + `; block ids are
not unique enough to use as cursors by themselves. Cap the page size with
` + "`--limit`" + ` (1-10000).`,
	Example: `  # Every block's usage for one workflow
  retab usage blocks --workflow-id wf_abc123

  # Only extract blocks active in a date window, as CSV
  retab usage blocks --block-type extract \
    --from-date 2026-06-01 --to-date 2026-06-30 --output csv > blocks.csv

  # Walk pages using a cursor returned by a previous response
  retab usage blocks --after <list_metadata.after> --limit 100`,
	Args: cobra.NoArgs,
	RunE: runE(runUsageBlocksList),
}

func runUsageBlocksList(cmd *cobra.Command, _ []string) error {
	if err := validateBeforeAfterMutex(cmd); err != nil {
		return err
	}
	if err := validateOrderFlag(cmd, "order"); err != nil {
		return err
	}
	if err := validateDateFlag(cmd, "from-date"); err != nil {
		return err
	}
	if err := validateDateFlag(cmd, "to-date"); err != nil {
		return err
	}
	fromDate, _ := cmd.Flags().GetString("from-date")
	toDate, _ := cmd.Flags().GetString("to-date")
	if err := validateDateRange("from-date", "to-date", fromDate, toDate); err != nil {
		return err
	}

	query := url.Values{}
	addOptionalUsageQuery(cmd, query, "workflow-id", "workflow_id")
	addOptionalUsageQuery(cmd, query, "block-type", "block_type")
	addOptionalUsageQuery(cmd, query, "before", "before")
	addOptionalUsageQuery(cmd, query, "after", "after")
	addOptionalUsageQuery(cmd, query, "order", "order")
	if fromDate != "" {
		query.Set("from_date", fromDate)
	}
	if toDate != "" {
		query.Set("to_date", toDate)
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		query.Set("limit", fmt.Sprintf("%d", v))
	}

	var result usageBlockListResponse
	if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/usage/blocks", query, nil, &result); err != nil {
		return err
	}
	return printUsageBlockListResult(cmd, result)
}

var usageBlockColumns = []TableColumn{
	{Header: "BLOCK_ID", Extract: func(row any) string { return usageBlockCell(row, "block_id") }},
	{Header: "WORKFLOW", Extract: func(row any) string { return usageBlockCell(row, "workflow_id") }},
	{Header: "TYPE", Extract: func(row any) string { return usageBlockCell(row, "block_type") }},
	{Header: "RUNS", Extract: func(row any) string { return usageBlockCell(row, "run_count") }},
	{Header: "EXECS", Extract: func(row any) string { return usageBlockCell(row, "execution_count") }},
	{Header: "PAGES", Extract: func(row any) string { return usageBlockCell(row, "page_count") }},
	{Header: "FAILED", Extract: usageBlockFailedCell},
	{Header: "CREDITS", Extract: func(row any) string { return usageBlockCell(row, "credits") }},
	{Header: "LAST_ACTIVITY", Extract: func(row any) string { return usageBlockCell(row, "last_activity_at") }},
}

// usageBlockFailedCell surfaces the block's failed-execution count from the
// status_counts breakdown — the at-a-glance signal that distinguishes an
// all-failing block from an all-succeeding one at equal credits. Blank when the
// block has no failures (or no breakdown). The full per-status map is available
// in --output json / csv.
func usageBlockFailedCell(row any) string {
	r, ok := row.(usageBlockRecord)
	if !ok {
		return ""
	}
	if n := r.StatusCounts["failed"]; n > 0 {
		return strconv.FormatInt(n, 10)
	}
	return ""
}

func printUsageBlockListResult(cmd *cobra.Command, result usageBlockListResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	return RenderList(os.Stdout, format, result, usageBlockColumns)
}

func usageBlockCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func init() {
	usageBlocksCmd.Flags().String("workflow-id", "", "filter to a single workflow id")
	usageBlocksCmd.Flags().String("block-type", "", "filter by block type (e.g. extract, classify, split, parse, edit, partition)")
	usageBlocksCmd.Flags().String("from-date", "", "inclusive activity lower bound (YYYY-MM-DD, UTC)")
	usageBlocksCmd.Flags().String("to-date", "", "inclusive activity upper bound (YYYY-MM-DD, UTC)")
	usageBlocksCmd.Flags().String("before", "", "cursor from list_metadata.before (mutually exclusive with --after)")
	usageBlocksCmd.Flags().String("after", "", "cursor from list_metadata.after (mutually exclusive with --before)")
	usageBlocksCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 10000}, "limit", "max items to return (1-10000)")
	usageBlocksCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")

	usageCmd.AddCommand(usageBlocksCmd)
}
