package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// The `usage` command group surfaces the organization + environment scoped
// usage ledger. It maps to the hidden /v1/usage/* routes, which are not part of
// the public OpenAPI contract (and therefore not generated into the SDK), so the
// commands are hand-written over the raw JSON request path, exactly like
// `workflows stats`.
var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Inspect organization usage (credits and pages)",
	Long: `Inspect the usage ledger for the authenticated organization and environment.

The usage routes are scoped to the caller's own organization + environment and
expose usage metadata only — credits and page counts — never document content,
filenames, artifact URIs, or provider costs.`,
}

// usageRunRecord mirrors the GET /v1/usage/runs row (UsageRunRecord on the
// server). Only usage + operational metadata is present by design.
type usageRunRecord struct {
	RunID       string  `json:"run_id"`
	WorkflowID  string  `json:"workflow_id"`
	Status      string  `json:"status"`
	TriggerType string  `json:"trigger_type"`
	CreatedAt   *string `json:"created_at,omitempty"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	DurationMs  *int64  `json:"duration_ms,omitempty"`
	PageCount   int64   `json:"page_count"`
	Credits     float64 `json:"credits"`
}

type usageRunListMetadata struct {
	Before *string `json:"before"`
	After  *string `json:"after"`
}

// usageRunListResponse is the GET /v1/usage/runs envelope. The `data` field name
// lets RenderList extract the row slice for table / CSV output.
type usageRunListResponse struct {
	Data         []usageRunRecord     `json:"data"`
	ListMetadata usageRunListMetadata `json:"list_metadata"`
}

var usageRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Per-run usage export (credits and pages per workflow run)",
	Long: `List one usage row per workflow run: status, trigger, timing, deduplicated
page count, and credit spend.

Scoped to the authenticated organization and environment. This is the
confidential-safe, productized form of the offline workflow-run dashboard
export — it never exposes document content, filenames, artifact URIs, or
provider/API dollar costs.

Filter by workflow, lifecycle status, trigger type, and created_at date range.
Page by run id with ` + "`--before`" + ` / ` + "`--after`" + `, cap the page size
with ` + "`--limit`" + ` (1-100).`,
	Example: `  # Most recent 50 runs' usage
  retab usage runs --limit 50

  # One workflow, errored runs only, in a date window
  retab usage runs --workflow-id wf_abc123 --status error \
    --from-date 2026-06-01 --to-date 2026-06-30

  # As CSV for a spreadsheet
  retab usage runs --limit 100 --output csv > usage.csv

  # Walk pages from a known run id
  retab usage runs --after run_xyz789 --limit 100`,
	Args: cobra.NoArgs,
	RunE: runE(runUsageRunsList),
}

func runUsageRunsList(cmd *cobra.Command, _ []string) error {
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
	addOptionalUsageQuery(cmd, query, "status", "status")
	addOptionalUsageQuery(cmd, query, "trigger-type", "trigger_type")
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

	var result usageRunListResponse
	if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/usage/runs", query, nil, &result); err != nil {
		return err
	}
	return printUsageRunListResult(cmd, result)
}

// addOptionalUsageQuery copies a non-empty string flag into the query under the
// snake_case server parameter name.
func addOptionalUsageQuery(cmd *cobra.Command, query url.Values, flagName, queryName string) {
	if v, _ := cmd.Flags().GetString(flagName); strings.TrimSpace(v) != "" {
		query.Set(queryName, strings.TrimSpace(v))
	}
}

var usageRunColumns = []TableColumn{
	{Header: "RUN_ID", Extract: func(row any) string { return usageRunCell(row, "run_id") }},
	{Header: "WORKFLOW", Extract: func(row any) string { return usageRunCell(row, "workflow_id") }},
	{Header: "STATUS", Extract: func(row any) string { return usageRunCell(row, "status") }},
	{Header: "TRIGGER", Extract: func(row any) string { return usageRunCell(row, "trigger_type") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return usageRunCell(row, "created_at") }},
	{Header: "DURATION_MS", Extract: func(row any) string { return usageRunCell(row, "duration_ms") }},
	{Header: "PAGES", Extract: func(row any) string { return usageRunCell(row, "page_count") }},
	{Header: "CREDITS", Extract: func(row any) string { return usageRunCell(row, "credits") }},
}

func printUsageRunListResult(cmd *cobra.Command, result usageRunListResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	return RenderList(os.Stdout, format, result, usageRunColumns)
}

func usageRunCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func init() {
	usageRunsCmd.Flags().String("workflow-id", "", "filter to a single workflow id")
	usageRunsCmd.Flags().Var(
		newEnumStringFlagValue("--status",
			"pending", "queued", "running", "completed", "error", "failed", "awaiting_review", "cancelled"),
		"status", "filter by lifecycle status")
	usageRunsCmd.Flags().Var(
		newEnumStringFlagValue("--trigger-type",
			"manual", "api", "schedule", "webhook", "email", "restart"),
		"trigger-type", "filter by trigger type")
	usageRunsCmd.Flags().String("from-date", "", "inclusive created_at lower bound (YYYY-MM-DD, UTC)")
	usageRunsCmd.Flags().String("to-date", "", "inclusive created_at upper bound (YYYY-MM-DD, UTC)")
	usageRunsCmd.Flags().String("before", "", "run id: return items before this id (mutually exclusive with --after)")
	usageRunsCmd.Flags().String("after", "", "run id: return items after this id (mutually exclusive with --before)")
	usageRunsCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	usageRunsCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")

	usageCmd.AddCommand(usageRunsCmd)
	rootCmd.AddCommand(usageCmd)
}
