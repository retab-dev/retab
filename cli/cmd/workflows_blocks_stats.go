//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type workflowBlockClassifierCategoryStat struct {
	Category          string  `json:"category"`
	HandleKey         *string `json:"handle_key,omitempty"`
	ExecutionCount    int     `json:"execution_count"`
	ExecutionPercent  float64 `json:"execution_percent"`
	LatestCompletedAt *string `json:"latest_completed_at,omitempty"`
}

type workflowBlockClassifierStats struct {
	TotalExecutions         int                                   `json:"total_executions"`
	UncategorizedExecutions int                                   `json:"uncategorized_executions"`
	LatestCompletedAt       *string                               `json:"latest_completed_at,omitempty"`
	Categories              []workflowBlockClassifierCategoryStat `json:"categories"`
}

type workflowBlockAnalyticsTimeRange struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Granularity string `json:"granularity"`
}

type workflowBlockAnalyticsSummary struct {
	TotalExecutions   int      `json:"total_executions"`
	CompletedCount    int      `json:"completed_count"`
	ErrorCount        int      `json:"error_count"`
	SkippedCount      int      `json:"skipped_count"`
	CancelledCount    int      `json:"cancelled_count"`
	RunningCount      int      `json:"running_count"`
	CompletionRate    float64  `json:"completion_rate"`
	ErrorRate         float64  `json:"error_rate"`
	P50DurationMs     *float64 `json:"p50_duration_ms,omitempty"`
	P95DurationMs     *float64 `json:"p95_duration_ms,omitempty"`
	LatestCompletedAt *string  `json:"latest_completed_at,omitempty"`
}

type workflowBlockAnalyticsTimeSeriesPoint struct {
	BucketStart    string   `json:"bucket_start"`
	Executions     int      `json:"executions"`
	CompletedCount int      `json:"completed_count"`
	ErrorCount     int      `json:"error_count"`
	SkippedCount   int      `json:"skipped_count"`
	CancelledCount int      `json:"cancelled_count"`
	P50DurationMs  *float64 `json:"p50_duration_ms,omitempty"`
	P95DurationMs  *float64 `json:"p95_duration_ms,omitempty"`
}

type workflowBlockAnalyticsStatusBreakdown struct {
	Status  string  `json:"status"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

type workflowBlockAnalyticsConfigVersion struct {
	Fingerprint       string   `json:"fingerprint"`
	Executions        int      `json:"executions"`
	CompletedCount    int      `json:"completed_count"`
	ErrorCount        int      `json:"error_count"`
	CompletionRate    float64  `json:"completion_rate"`
	ErrorRate         float64  `json:"error_rate"`
	P95DurationMs     *float64 `json:"p95_duration_ms,omitempty"`
	LatestCompletedAt *string  `json:"latest_completed_at,omitempty"`
}

type workflowBlockAnalyticsErrorGroup struct {
	MessageHash string  `json:"message_hash"`
	Message     string  `json:"message"`
	Count       int     `json:"count"`
	LatestAt    *string `json:"latest_at,omitempty"`
}

type workflowBlockExtractFieldStat struct {
	FieldPath    string  `json:"field_path"`
	PresentCount int     `json:"present_count"`
	MissingCount int     `json:"missing_count"`
	FillRate     float64 `json:"fill_rate"`
}

type workflowBlockExtractStats struct {
	Fields []workflowBlockExtractFieldStat `json:"fields"`
}

type workflowBlockAnalytics struct {
	GeneratedAt     string                                  `json:"generated_at"`
	TimeRange       workflowBlockAnalyticsTimeRange         `json:"time_range"`
	Summary         workflowBlockAnalyticsSummary           `json:"summary"`
	TimeSeries      []workflowBlockAnalyticsTimeSeriesPoint `json:"time_series"`
	StatusBreakdown []workflowBlockAnalyticsStatusBreakdown `json:"status_breakdown"`
	ConfigVersions  []workflowBlockAnalyticsConfigVersion   `json:"config_versions"`
	ErrorGroups     []workflowBlockAnalyticsErrorGroup      `json:"error_groups"`
	ExtractStats    *workflowBlockExtractStats              `json:"extract_stats,omitempty"`
}

type workflowBlockStatsResponse struct {
	BlockID         string                        `json:"block_id"`
	WorkflowID      string                        `json:"workflow_id"`
	BlockType       string                        `json:"block_type"`
	QuerySource     string                        `json:"query_source"`
	QueryStatus     string                        `json:"query_status"`
	Analytics       *workflowBlockAnalytics       `json:"analytics,omitempty"`
	ClassifierStats *workflowBlockClassifierStats `json:"classifier_stats,omitempty"`
}

var workflowsBlocksStatsCmd = &cobra.Command{
	Use:   "stats [<workflow-id>] <block-id>",
	Short: "Get workflow block stats",
	Long: `Fetch dashboard block stats for one block.

The backend endpoint is scoped by both workflow id and block id, so pass the
workflow id positionally (` + "`stats <workflow-id> <block-id>`" + `) or with
` + "`--workflow-id`" + `. The endpoint returns BigQuery-backed execution
analytics for all block types when enabled, plus classifier category stats for
classifier blocks. Use ` + "`--from`" + `, ` + "`--to`" + `, and ` + "`--granularity`" + ` to choose the
analytics time window.`,
	Example: `  # Get stats for a block
  retab workflows blocks stats wf_abc123 blk_classifier

  # Equivalent flag form
  retab workflows blocks stats --workflow-id wf_abc123 blk_classifier

  # Explicit get subcommand
  retab workflows blocks stats get wf_abc123 blk_classifier`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(runWorkflowsBlocksStatsGet),
}

var workflowsBlocksStatsGetCmd = &cobra.Command{
	Use:   "get [<workflow-id>] <block-id>",
	Short: "Get workflow block stats",
	Long:  workflowsBlocksStatsCmd.Long,
	Example: `  retab workflows blocks stats get wf_abc123 blk_classifier
  retab workflows blocks stats get --workflow-id wf_abc123 blk_classifier`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(runWorkflowsBlocksStatsGet),
}

func runWorkflowsBlocksStatsGet(cmd *cobra.Command, args []string) error {
	workflowID, blockID, err := resolveBlockStatsScope(cmd, args)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("workflow_id", workflowID)
	addOptionalBlockStatsQuery(cmd, query, "from")
	addOptionalBlockStatsQuery(cmd, query, "to")
	addOptionalBlockStatsQuery(cmd, query, "granularity")
	var result workflowBlockStatsResponse
	requestPath := "/v1/workflows/blocks/" + url.PathEscape(blockID) + "/stats"
	if err := cliJSONRequestInto(cmd, http.MethodGet, requestPath, query, nil, &result); err != nil {
		return err
	}
	return printBlockStatsResult(cmd, result)
}

func resolveBlockStatsScope(cmd *cobra.Command, args []string) (string, string, error) {
	flagWorkflowID := workflowBlockStatsWorkflowID(cmd)
	var workflowID string
	var blockID string
	switch len(args) {
	case 1:
		workflowID = flagWorkflowID
		blockID = strings.TrimSpace(args[0])
	case 2:
		workflowID = strings.TrimSpace(args[0])
		blockID = strings.TrimSpace(args[1])
		if workflowID == "" {
			return "", "", fmt.Errorf("workflow-id positional argument is empty")
		}
		if flagWorkflowID != "" && flagWorkflowID != workflowID {
			return "", "", fmt.Errorf("conflicting workflow id: positional %q vs --workflow-id %q", workflowID, flagWorkflowID)
		}
	default:
		return "", "", fmt.Errorf("expected 1 or 2 positional arguments, got %d", len(args))
	}
	if blockID == "" {
		return "", "", fmt.Errorf("block-id positional argument is empty")
	}
	if workflowID == "" {
		return "", "", fmt.Errorf("workflow id is required for block stats; pass `stats <workflow-id> <block-id>` or `--workflow-id <workflow-id>`")
	}
	return workflowID, blockID, nil
}

func workflowBlockStatsWorkflowID(cmd *cobra.Command) string {
	for current := cmd; current != nil; current = current.Parent() {
		if f := current.Flags().Lookup("workflow-id"); f != nil {
			value, _ := current.Flags().GetString("workflow-id")
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
		if f := current.PersistentFlags().Lookup("workflow-id"); f != nil {
			value, _ := current.PersistentFlags().GetString("workflow-id")
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func addOptionalBlockStatsQuery(cmd *cobra.Command, query url.Values, name string) {
	for current := cmd; current != nil; current = current.Parent() {
		if f := current.Flags().Lookup(name); f != nil {
			value, _ := current.Flags().GetString(name)
			if strings.TrimSpace(value) != "" {
				query.Set(name, strings.TrimSpace(value))
				return
			}
		}
		if f := current.PersistentFlags().Lookup(name); f != nil {
			value, _ := current.PersistentFlags().GetString(name)
			if strings.TrimSpace(value) != "" {
				query.Set(name, strings.TrimSpace(value))
				return
			}
		}
	}
}

var blockStatsColumns = []TableColumn{
	{Header: "BLOCK", Extract: func(row any) string { return blockStatsCell(row, "block_id") }},
	{Header: "WORKFLOW", Extract: func(row any) string { return blockStatsCell(row, "workflow_id") }},
	{Header: "TYPE", Extract: func(row any) string { return blockStatsCell(row, "block_type") }},
	{Header: "STATUS", Extract: func(row any) string { return blockStatsCell(row, "query_status") }},
	{Header: "EXECUTIONS", Extract: blockStatsTotalExecutionsCell},
	{Header: "COMPLETED", Extract: func(row any) string { return blockStatsCell(row, "analytics.summary.completed_count") }},
	{Header: "ERRORS", Extract: func(row any) string { return blockStatsCell(row, "analytics.summary.error_count") }},
	{Header: "COMPLETION_RATE", Extract: func(row any) string { return blockStatsCell(row, "analytics.summary.completion_rate") }},
	{Header: "ERROR_RATE", Extract: func(row any) string { return blockStatsCell(row, "analytics.summary.error_rate") }},
	{Header: "P95_MS", Extract: func(row any) string { return blockStatsCell(row, "analytics.summary.p95_duration_ms") }},
	{Header: "CATEGORIES", Extract: blockStatsCategoryCountCell},
	{Header: "UNCATEGORIZED", Extract: func(row any) string { return blockStatsCell(row, "classifier_stats.uncategorized_executions") }},
	{Header: "LATEST_COMPLETED_AT", Extract: blockStatsLatestCompletedAtCell},
}

func printBlockStatsResult(cmd *cobra.Command, result workflowBlockStatsResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputJSON {
		return printJSON(result)
	}
	rows := []any{result}
	if format == OutputCSV {
		return renderAutoCSV(os.Stdout, rows, blockStatsColumns)
	}
	return renderAutoTable(os.Stdout, rows, blockStatsColumns)
}

func blockStatsCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func blockStatsFirstCell(row any, keys ...string) string {
	for _, key := range keys {
		if value := blockStatsCell(row, key); value != "" {
			return value
		}
	}
	return ""
}

func blockStatsTotalExecutionsCell(row any) string {
	return blockStatsFirstCell(row, "classifier_stats.total_executions", "analytics.summary.total_executions")
}

func blockStatsLatestCompletedAtCell(row any) string {
	return normalizeTimestampCell(blockStatsFirstCell(row, "classifier_stats.latest_completed_at", "analytics.summary.latest_completed_at"))
}

func blockStatsCategoryCountCell(row any) string {
	value, ok := rowField(row, "classifier_stats.categories")
	if !ok || value == nil {
		return ""
	}
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return ""
	}
	return strconv.Itoa(rv.Len())
}

func init() {
	workflowsBlocksStatsCmd.PersistentFlags().String("workflow-id", "", "workflow id (required unless passed positionally)")
	workflowsBlocksStatsCmd.PersistentFlags().String("from", "", "analytics window start (YYYY-MM-DD or RFC3339)")
	workflowsBlocksStatsCmd.PersistentFlags().String("to", "", "analytics window end (YYYY-MM-DD or RFC3339)")
	workflowsBlocksStatsCmd.PersistentFlags().String("granularity", "", "analytics bucket granularity: hour | day | week")
	workflowsBlocksStatsCmd.AddCommand(workflowsBlocksStatsGetCmd)
	workflowsBlocksCmd.AddCommand(workflowsBlocksStatsCmd)
}
