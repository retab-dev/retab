package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

// usagePrimitiveRecord mirrors the GET /v1/usage/primitives row
// (UsagePrimitiveRecord on the server). Only usage + operational metadata is
// present by design.
type usagePrimitiveRecord struct {
	PrimitiveExecutionID string  `json:"primitive_execution_id"`
	Operation            string  `json:"operation"`
	WorkflowID           string  `json:"workflow_id,omitempty"`
	RunID                string  `json:"run_id,omitempty"`
	ProjectID            string  `json:"project_id,omitempty"`
	BlockID              string  `json:"block_id,omitempty"`
	SourceType           string  `json:"source_type,omitempty"`
	Status               string  `json:"status"`
	ResourceKind         string  `json:"resource_kind,omitempty"`
	CreatedAt            *string `json:"created_at,omitempty"`
	PageCount            int64   `json:"page_count"`
	Credits              float64 `json:"credits"`
}

// usagePrimitiveListResponse is the GET /v1/usage/primitives envelope. The `data`
// field name lets RenderList extract the row slice for table / CSV output.
type usagePrimitiveListResponse struct {
	Data         []usagePrimitiveRecord `json:"data"`
	ListMetadata usageRunListMetadata   `json:"list_metadata"`
}

var usagePrimitivesCmd = &cobra.Command{
	Use:   "primitives",
	Short: "Per-operation usage export (credits and pages per primitive execution)",
	Long: `List one usage row per primitive execution (extraction, classify, split, parse,
edit, schema_generation …): the operation, its origin identifiers (workflow, run,
project, block), lifecycle status, deduplicated page count, and credit spend.

This is the per-operation grain of the usage export — the list form of the usage
dashboard's per-operation credits graph. Scoped to the authenticated organization
and environment, and confidential-safe like ` + "`usage runs`" + `: it never exposes
model names, token counts, provider/API dollar costs, request metadata, or
document content.

Filter by workflow, project, run, operation, lifecycle status, and created_at
date range. Page by execution id with ` + "`--before`" + ` / ` + "`--after`" + `, cap the
page size with ` + "`--limit`" + ` (1-100).`,
	Example: `  # Most recent 50 operations' usage
  retab usage primitives --limit 50

  # Every extraction for one project, in a date window, as CSV
  retab usage primitives --project-id proj_abc123 --operation extraction \
    --from-date 2026-06-01 --to-date 2026-06-30 --output csv > operations.csv

  # One workflow's classify operations
  retab usage primitives --workflow-id wf_abc123 --operation classify

  # Walk pages from a known execution id
  retab usage primitives --after pexec_xyz789 --limit 100`,
	Args: cobra.NoArgs,
	RunE: runE(runUsagePrimitivesList),
}

func runUsagePrimitivesList(cmd *cobra.Command, _ []string) error {
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
	addOptionalUsageQuery(cmd, query, "project-id", "project_id")
	addOptionalUsageQuery(cmd, query, "run-id", "run_id")
	addOptionalUsageQuery(cmd, query, "operation", "operation")
	addOptionalUsageQuery(cmd, query, "status", "status")
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

	var result usagePrimitiveListResponse
	if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/usage/primitives", query, nil, &result); err != nil {
		return err
	}
	return printUsagePrimitiveListResult(cmd, result)
}

var usagePrimitiveColumns = []TableColumn{
	{Header: "EXECUTION_ID", Extract: func(row any) string { return usagePrimitiveCell(row, "primitive_execution_id") }},
	{Header: "OPERATION", Extract: func(row any) string { return usagePrimitiveCell(row, "operation") }},
	{Header: "WORKFLOW", Extract: func(row any) string { return usagePrimitiveCell(row, "workflow_id") }},
	{Header: "PROJECT", Extract: func(row any) string { return usagePrimitiveCell(row, "project_id") }},
	{Header: "STATUS", Extract: func(row any) string { return usagePrimitiveCell(row, "status") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return usagePrimitiveCell(row, "created_at") }},
	{Header: "PAGES", Extract: func(row any) string { return usagePrimitiveCell(row, "page_count") }},
	{Header: "CREDITS", Extract: func(row any) string { return usagePrimitiveCell(row, "credits") }},
}

func printUsagePrimitiveListResult(cmd *cobra.Command, result usagePrimitiveListResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	return RenderList(os.Stdout, format, result, usagePrimitiveColumns)
}

func usagePrimitiveCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func init() {
	usagePrimitivesCmd.Flags().String("workflow-id", "", "filter to a single workflow id (origin workflow)")
	usagePrimitivesCmd.Flags().String("project-id", "", "filter to executions owned by a single project id")
	usagePrimitivesCmd.Flags().String("run-id", "", "filter to a single workflow run id (origin run)")
	usagePrimitivesCmd.Flags().String("operation", "", "filter by operation (extraction, classify, split, parse, edit, schema_generation)")
	usagePrimitivesCmd.Flags().String("status", "", "filter by execution lifecycle status")
	usagePrimitivesCmd.Flags().String("from-date", "", "inclusive created_at lower bound (YYYY-MM-DD, UTC)")
	usagePrimitivesCmd.Flags().String("to-date", "", "inclusive created_at upper bound (YYYY-MM-DD, UTC)")
	usagePrimitivesCmd.Flags().String("before", "", "execution id: return items before this id (mutually exclusive with --after)")
	usagePrimitivesCmd.Flags().String("after", "", "execution id: return items after this id (mutually exclusive with --before)")
	usagePrimitivesCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	usagePrimitivesCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")

	usageCmd.AddCommand(usagePrimitivesCmd)
}
