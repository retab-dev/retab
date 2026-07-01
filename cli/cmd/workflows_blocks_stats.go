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

type workflowBlockAnalyticsTimeRange struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Granularity string `json:"granularity"`
}

type workflowBlockRunVolumePoint struct {
	BucketStart string `json:"bucket_start"`
	Runs        int    `json:"runs"`
}

type workflowBlockRunVolumeMetric struct {
	TotalRuns int                           `json:"total_runs"`
	Series    []workflowBlockRunVolumePoint `json:"series"`
}

type workflowBlockAnalytics struct {
	GeneratedAt string                          `json:"generated_at"`
	TimeRange   workflowBlockAnalyticsTimeRange `json:"time_range"`
	RunVolume   workflowBlockRunVolumeMetric    `json:"run_volume"`
	Details     map[string]any                  `json:"details,omitempty"`
}

type workflowBlockStatsResponse struct {
	BlockID     string                  `json:"block_id"`
	WorkflowID  string                  `json:"workflow_id"`
	BlockType   string                  `json:"block_type"`
	QuerySource string                  `json:"query_source"`
	QueryStatus string                  `json:"query_status"`
	Analytics   *workflowBlockAnalytics `json:"analytics,omitempty"`
}

var workflowsBlocksStatsCmd = &cobra.Command{
	Use:   "stats [<workflow-id>] <block-id>",
	Short: "Get workflow block stats",
	Long: `Fetch dashboard block stats for one block.

The backend endpoint is scoped by both workflow id and block id, so pass the
workflow id positionally (` + "`stats <workflow-id> <block-id>`" + `) or with
` + "`--workflow-id`" + `. The endpoint returns dashboard analytics for run
volume plus the block-specific output shape. Use ` + "`--from`" + `, ` + "`--to`" + `, and
` + "`--granularity`" + ` to choose the analytics time window.`,
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
	addOptionalStatsQuery(cmd, query, "from")
	addOptionalStatsQuery(cmd, query, "to")
	addOptionalStatsQuery(cmd, query, "granularity")
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

func addOptionalStatsQuery(cmd *cobra.Command, query url.Values, name string) {
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
	{Header: "RUNS", Extract: func(row any) string { return blockStatsCell(row, "analytics.run_volume.total_runs") }},
	{Header: "DETAIL", Extract: func(row any) string { return blockStatsCell(row, "analytics.details.block_type") }},
	{Header: "TOTAL", Extract: blockStatsDetailTotalCell},
	{Header: "AVG", Extract: blockStatsDetailAverageCell},
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

func blockStatsDetailTotalCell(row any) string {
	if value := blockStatsFirstCell(row,
		"analytics.details.item_volume.total_items",
		"analytics.details.output_shape.total_subdocuments",
		"analytics.details.output_shape.schema_field_count",
	); value != "" {
		return value
	}
	categories, ok := rowField(row, "analytics.details.classification_categories")
	if !ok {
		return ""
	}
	return blockStatsCategoryCountCell(categories)
}

func blockStatsDetailAverageCell(row any) string {
	return blockStatsFirstCell(row,
		"analytics.details.item_volume.avg_items_per_run",
		"analytics.details.output_shape.avg_subdocuments_per_run",
		"analytics.details.output_shape.avg_filled_fields_per_run",
	)
}

func blockStatsCategoryCountCell(value any) string {
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
	workflowsBlocksStatsCmd.PersistentFlags().String("workflow-id", "", "workflow id (required unless passed positionally)")
	workflowsBlocksStatsCmd.PersistentFlags().String("from", "", "analytics window start (YYYY-MM-DD or RFC3339)")
	workflowsBlocksStatsCmd.PersistentFlags().String("to", "", "analytics window end (YYYY-MM-DD or RFC3339)")
	workflowsBlocksStatsCmd.PersistentFlags().String("granularity", "", "analytics bucket granularity: hour | day | week")
	workflowsBlocksStatsCmd.AddCommand(workflowsBlocksStatsGetCmd)
	workflowsBlocksCmd.AddCommand(workflowsBlocksStatsCmd)
}
