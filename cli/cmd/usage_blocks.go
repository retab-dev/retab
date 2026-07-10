package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

// usageBlockRecord mirrors the GET /v1/usage/blocks row (UsageBlockRecord on the
// server). Only usage + operational metadata is present by design.
type usageBlockRecord struct {
	BlockID         string  `json:"block_id"`
	WorkflowID      string  `json:"workflow_id"`
	BlockType       string  `json:"block_type"`
	RunCount        int64   `json:"run_count"`
	ExecutionCount  int64   `json:"execution_count"`
	PageCount       int64   `json:"page_count"`
	Credits         float64 `json:"credits"`
	FirstActivityAt *string `json:"first_activity_at,omitempty"`
	LastActivityAt  *string `json:"last_activity_at,omitempty"`
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

Filter by workflow, block type, and activity date range. Page by block id with
` + "`--before`" + ` / ` + "`--after`" + `, cap the page size with ` + "`--limit`" + ` (1-100).`,
	Example: `  # Every block's usage for one workflow
  retab usage blocks --workflow-id wf_abc123

  # Only extract blocks active in a date window, as CSV
  retab usage blocks --block-type extract \
    --from-date 2026-06-01 --to-date 2026-06-30 --output csv > blocks.csv

  # Walk pages from a known block id
  retab usage blocks --after block_xyz789 --limit 100`,
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
	{Header: "CREDITS", Extract: func(row any) string { return usageBlockCell(row, "credits") }},
	{Header: "LAST_ACTIVITY", Extract: func(row any) string { return usageBlockCell(row, "last_activity_at") }},
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
	usageBlocksCmd.Flags().String("before", "", "block id: return items before this id (mutually exclusive with --after)")
	usageBlocksCmd.Flags().String("after", "", "block id: return items after this id (mutually exclusive with --before)")
	usageBlocksCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	usageBlocksCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")

	usageCmd.AddCommand(usageBlocksCmd)
}
