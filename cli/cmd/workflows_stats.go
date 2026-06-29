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

type workflowStatsAnalyticsTimeRange struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Granularity string `json:"granularity"`
}

type workflowStatsRunVolumePoint struct {
	BucketStart string `json:"bucket_start"`
	Runs        int    `json:"runs"`
}

type workflowStatsRunVolumeMetric struct {
	TotalRuns int                           `json:"total_runs"`
	Series    []workflowStatsRunVolumePoint `json:"series"`
}

type workflowStatsBucket struct {
	Bucket     string  `json:"bucket"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type workflowStatsDocumentShape struct {
	DocumentFormatDistribution   []workflowStatsBucket `json:"document_format_distribution"`
	PagesPerDocumentDistribution []workflowStatsBucket `json:"pages_per_document_distribution"`
}

type workflowStatsAnalytics struct {
	GeneratedAt   string                          `json:"generated_at"`
	TimeRange     workflowStatsAnalyticsTimeRange `json:"time_range"`
	RunVolume     workflowStatsRunVolumeMetric    `json:"run_volume"`
	DocumentShape workflowStatsDocumentShape      `json:"document_shape"`
}

type workflowStatsResponse struct {
	WorkflowID  string                  `json:"workflow_id"`
	QuerySource string                  `json:"query_source"`
	QueryStatus string                  `json:"query_status"`
	Analytics   *workflowStatsAnalytics `json:"analytics,omitempty"`
}

var workflowsStatsCmd = &cobra.Command{
	Use:   "stats <workflow-id>",
	Short: "Get workflow stats",
	Long: `Fetch dashboard workflow stats.

Workflow stats focus on metrics that describe the workflow as a whole: run
volume and input document shape. Block-specific shape metrics live under
` + "`workflows stats blocks`" + `.`,
	Example: `  # Get workflow-level stats
  retab workflows stats wf_abc123

  # Explicit get subcommand with a time window
  retab workflows stats get wf_abc123 --from 2026-06-01 --to 2026-06-29

  # Get one block's stats through the stats namespace
  retab workflows stats blocks get wf_abc123 blk_extract`,
	Args: cobra.ExactArgs(1),
	RunE: runE(runWorkflowsStatsGet),
}

var workflowsStatsGetCmd = &cobra.Command{
	Use:     "get <workflow-id>",
	Aliases: []string{"retrieve"},
	Short:   "Get workflow stats",
	Long:    workflowsStatsCmd.Long,
	Example: `  retab workflows stats get wf_abc123
  retab workflows stats get wf_abc123 --granularity week`,
	Args: cobra.ExactArgs(1),
	RunE: runE(runWorkflowsStatsGet),
}

var workflowsStatsBlocksCmd = &cobra.Command{
	Use:   "blocks [<workflow-id>] <block-id>",
	Short: "Get workflow block stats",
	Long:  workflowsBlocksStatsCmd.Long,
	Example: `  retab workflows stats blocks wf_abc123 blk_extract
  retab workflows stats blocks get --workflow-id wf_abc123 blk_extract`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(runWorkflowsBlocksStatsGet),
}

var workflowsStatsBlocksGetCmd = &cobra.Command{
	Use:   "get [<workflow-id>] <block-id>",
	Short: "Get workflow block stats",
	Long:  workflowsBlocksStatsCmd.Long,
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runE(runWorkflowsBlocksStatsGet),
}

func runWorkflowsStatsGet(cmd *cobra.Command, args []string) error {
	workflowID := strings.TrimSpace(args[0])
	if workflowID == "" {
		return fmt.Errorf("workflow-id positional argument is empty")
	}
	query := url.Values{}
	addOptionalStatsQuery(cmd, query, "from")
	addOptionalStatsQuery(cmd, query, "to")
	addOptionalStatsQuery(cmd, query, "granularity")
	var result workflowStatsResponse
	requestPath := "/v1/workflows/" + url.PathEscape(workflowID) + "/stats"
	if err := cliJSONRequestInto(cmd, http.MethodGet, requestPath, query, nil, &result); err != nil {
		return err
	}
	return printWorkflowStatsResult(cmd, result)
}

var workflowStatsColumns = []TableColumn{
	{Header: "WORKFLOW", Extract: func(row any) string { return workflowStatsCell(row, "workflow_id") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowStatsCell(row, "query_status") }},
	{Header: "RUNS", Extract: func(row any) string { return workflowStatsCell(row, "analytics.run_volume.total_runs") }},
	{Header: "TOP_FORMAT", Extract: func(row any) string {
		return workflowStatsTopBucketCell(row, "analytics.document_shape.document_format_distribution")
	}},
	{Header: "PAGES", Extract: func(row any) string {
		return workflowStatsTopBucketCell(row, "analytics.document_shape.pages_per_document_distribution")
	}},
}

func printWorkflowStatsResult(cmd *cobra.Command, result workflowStatsResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputJSON {
		return printJSON(result)
	}
	rows := []any{result}
	if format == OutputCSV {
		return renderAutoCSV(os.Stdout, rows, workflowStatsColumns)
	}
	return renderAutoTable(os.Stdout, rows, workflowStatsColumns)
}

func workflowStatsCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func workflowStatsTopBucketCell(row any, key string) string {
	value, ok := rowField(row, key)
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
	if rv.Len() == 0 {
		return ""
	}
	first := rv.Index(0).Interface()
	bucket := workflowStatsCell(first, "bucket")
	count := workflowStatsCell(first, "count")
	if bucket == "" {
		return ""
	}
	if count == "" {
		return bucket
	}
	return bucket + " (" + count + ")"
}

func workflowStatsBucketCountCell(value any) string {
	if value == nil {
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
	workflowsStatsCmd.PersistentFlags().String("from", "", "analytics window start (YYYY-MM-DD or RFC3339)")
	workflowsStatsCmd.PersistentFlags().String("to", "", "analytics window end (YYYY-MM-DD or RFC3339)")
	workflowsStatsCmd.PersistentFlags().String("granularity", "", "analytics bucket granularity: hour | day | week")
	workflowsStatsBlocksCmd.PersistentFlags().String("workflow-id", "", "workflow id (required unless passed positionally)")
	workflowsStatsBlocksCmd.AddCommand(workflowsStatsBlocksGetCmd)
	workflowsStatsCmd.AddCommand(workflowsStatsGetCmd, workflowsStatsBlocksCmd)
	workflowsCmd.AddCommand(workflowsStatsCmd)
}
