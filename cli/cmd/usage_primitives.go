package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// usagePrimitiveRecord mirrors the GET /v1/usage/primitives row
// (UsagePrimitiveRecord on the server). Only usage + operational metadata plus
// the caller's own user metadata is present by design. The origin identifiers,
// resource kind, created_at, and metadata are nullable on the wire (explicit
// JSON null when absent); decoding a null into these Go fields leaves the zero
// value, which the table renderer prints as an empty cell.
type usagePrimitiveRecord struct {
	PrimitiveExecutionID string                     `json:"primitive_execution_id"`
	Operation            string                     `json:"operation"`
	WorkflowID           string                     `json:"workflow_id,omitempty"`
	RunID                string                     `json:"run_id,omitempty"`
	ProjectID            string                     `json:"project_id,omitempty"`
	BlockID              string                     `json:"block_id,omitempty"`
	Status               string                     `json:"status"`
	ResourceKind         string                     `json:"resource_kind,omitempty"`
	Model                string                     `json:"model,omitempty"`
	CreatedAt            *string                    `json:"created_at,omitempty"`
	CompletedAt          *string                    `json:"completed_at,omitempty"`
	DurationMs           *int64                     `json:"duration_ms,omitempty"`
	PageCount            int64                      `json:"page_count"`
	Credits              float64                    `json:"credits"`
	Documents            []usagePrimitiveDocumentEl `json:"documents,omitempty"`
	Metadata             map[string]string          `json:"metadata,omitempty"`
}

// usagePrimitiveDocumentEl is one source document of a primitive execution row.
type usagePrimitiveDocumentEl struct {
	FileID   string `json:"file_id,omitempty"`
	Filename string `json:"filename,omitempty"`
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
edit, partition, schema_generation …): the operation, its origin identifiers (workflow, run,
project, block), lifecycle status, the Retab model tier, source document filenames/ids,
created/completed timestamps, deduplicated page count, credit spend, and your own metadata.

This is the per-operation grain of the usage export — the list form of the usage
dashboard's per-operation credits graph. Scoped to the authenticated organization
and environment. Model is reported only as the normalized Retab tier
(retab-micro / retab-small / retab-large) — the underlying provider model id is
never surfaced — and token counts, provider/API dollar costs, and raw error text
remain excluded.

Filter by workflow, project, run, block, operation, lifecycle status, metadata,
and created_at date range. Page by execution id with ` + "`--before`" + ` / ` + "`--after`" + `,
cap the page size with ` + "`--limit`" + ` (1-100).

By default the export is scoped to the environment of the authenticated
credential; use the global ` + "`--environment-id`" + ` flag (or RETAB_ENVIRONMENT_ID,
or the stored config default) to report on another environment within your
organization.`,
	Example: `  # Most recent 50 operations' usage
  retab usage primitives --limit 50

  # Every extraction for one project, in a date window, as CSV
  retab usage primitives --project-id proj_abc123 --operation extraction \
    --from-date 2026-06-01 --to-date 2026-06-30 --output csv > operations.csv

  # One workflow's classify operations
  retab usage primitives --workflow-id wf_abc123 --operation classify

  # Filter by user-defined metadata (repeat --metadata to AND pairs)
  retab usage primitives --metadata tenant=acme --metadata tier=gold

  # Walk pages from a known execution id
  retab usage primitives --after pexec_xyz789 --limit 100

  # Report on a specific environment in your organization
  retab --environment-id env_abc123 usage primitives --limit 50`,
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
	// Forward the CLI's selected environment (global --environment-id flag,
	// RETAB_ENVIRONMENT_ID, or the stored config default) as the environment_id
	// scope argument. Empty → the server falls back to the credential's environment.
	cfg, _ := loadConfig()
	if envID := selectedEnvironmentID(cmd, cfg); strings.TrimSpace(envID) != "" {
		query.Set("environment_id", strings.TrimSpace(envID))
	}
	addOptionalUsageQuery(cmd, query, "workflow-id", "workflow_id")
	addOptionalUsageQuery(cmd, query, "project-id", "project_id")
	addOptionalUsageQuery(cmd, query, "run-id", "run_id")
	addOptionalUsageQuery(cmd, query, "block-id", "block_id")
	addOptionalUsageQuery(cmd, query, "operation", "operation")
	addOptionalUsageQuery(cmd, query, "status", "status")
	addOptionalUsageQuery(cmd, query, "before", "before")
	addOptionalUsageQuery(cmd, query, "after", "after")
	addOptionalUsageQuery(cmd, query, "order", "order")
	if metaPairs, _ := cmd.Flags().GetStringArray("metadata"); len(metaPairs) > 0 {
		md, err := parseKVStringList(metaPairs)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(md)
		if err != nil {
			return err
		}
		query.Set("metadata", string(raw))
	}
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
	{Header: "MODEL", Extract: func(row any) string { return usagePrimitiveCell(row, "model") }},
	{Header: "WORKFLOW", Extract: func(row any) string { return usagePrimitiveCell(row, "workflow_id") }},
	{Header: "BLOCK", Extract: func(row any) string { return usagePrimitiveCell(row, "block_id") }},
	{Header: "PROJECT", Extract: func(row any) string { return usagePrimitiveCell(row, "project_id") }},
	{Header: "STATUS", Extract: func(row any) string { return usagePrimitiveCell(row, "status") }},
	{Header: "FILENAME", Extract: usagePrimitiveFilenameCell},
	{Header: "CREATED_AT", Extract: func(row any) string { return usagePrimitiveCell(row, "created_at") }},
	{Header: "COMPLETED_AT", Extract: func(row any) string { return usagePrimitiveCell(row, "completed_at") }},
	{Header: "DURATION_MS", Extract: func(row any) string { return usagePrimitiveCell(row, "duration_ms") }},
	{Header: "PAGES", Extract: func(row any) string { return usagePrimitiveCell(row, "page_count") }},
	{Header: "CREDITS", Extract: func(row any) string { return usagePrimitiveCell(row, "credits") }},
}

// usagePrimitiveFilenameCell renders the first source document's filename (with a
// "+N" suffix when the execution has more), keeping the table to one line while
// the full document list stays available in --output json.
func usagePrimitiveFilenameCell(row any) string {
	rec, ok := row.(usagePrimitiveRecord)
	if !ok || len(rec.Documents) == 0 {
		return ""
	}
	name := rec.Documents[0].Filename
	if name == "" {
		name = rec.Documents[0].FileID
	}
	if extra := len(rec.Documents) - 1; extra > 0 {
		name = fmt.Sprintf("%s +%d", name, extra)
	}
	return name
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
	usagePrimitivesCmd.Flags().String("block-id", "", "filter to a single workflow block id (origin block)")
	usagePrimitivesCmd.Flags().String("operation", "", "filter by operation (extraction, classify, split, parse, edit, partition, schema_generation)")
	usagePrimitivesCmd.Flags().String("status", "", "filter by execution lifecycle status")
	usagePrimitivesCmd.Flags().StringArray("metadata", nil, "filter by metadata key=value (repeatable; pairs AND together)")
	usagePrimitivesCmd.Flags().String("from-date", "", "inclusive created_at lower bound (YYYY-MM-DD, UTC)")
	usagePrimitivesCmd.Flags().String("to-date", "", "inclusive created_at upper bound (YYYY-MM-DD, UTC)")
	usagePrimitivesCmd.Flags().String("before", "", "execution id: return items before this id (mutually exclusive with --after)")
	usagePrimitivesCmd.Flags().String("after", "", "execution id: return items after this id (mutually exclusive with --before)")
	usagePrimitivesCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	usagePrimitivesCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")

	usageCmd.AddCommand(usagePrimitivesCmd)
}
