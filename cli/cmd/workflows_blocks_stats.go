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

type workflowBlockStatsResponse struct {
	BlockID         string                        `json:"block_id"`
	WorkflowID      string                        `json:"workflow_id"`
	BlockType       string                        `json:"block_type"`
	QuerySource     string                        `json:"query_source"`
	QueryStatus     string                        `json:"query_status"`
	ClassifierStats *workflowBlockClassifierStats `json:"classifier_stats,omitempty"`
}

var workflowsBlocksStatsCmd = &cobra.Command{
	Use:   "stats [<workflow-id>] <block-id>",
	Short: "Get workflow block stats",
	Long: `Fetch dashboard block stats for one block.

The backend endpoint is scoped by both workflow id and block id, so pass the
workflow id positionally (` + "`stats <workflow-id> <block-id>`" + `) or with
` + "`--workflow-id`" + `. The current backend returns classifier category stats
from BigQuery when available; unsupported block types still return the block
identity and query status.`,
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

var blockStatsColumns = []TableColumn{
	{Header: "BLOCK", Extract: func(row any) string { return blockStatsCell(row, "block_id") }},
	{Header: "WORKFLOW", Extract: func(row any) string { return blockStatsCell(row, "workflow_id") }},
	{Header: "TYPE", Extract: func(row any) string { return blockStatsCell(row, "block_type") }},
	{Header: "STATUS", Extract: func(row any) string { return blockStatsCell(row, "query_status") }},
	{Header: "EXECUTIONS", Extract: func(row any) string { return blockStatsCell(row, "classifier_stats.total_executions") }},
	{Header: "CATEGORIES", Extract: blockStatsCategoryCountCell},
	{Header: "UNCATEGORIZED", Extract: func(row any) string { return blockStatsCell(row, "classifier_stats.uncategorized_executions") }},
	{Header: "LATEST_COMPLETED_AT", Extract: func(row any) string {
		return normalizeTimestampCell(blockStatsCell(row, "classifier_stats.latest_completed_at"))
	}},
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
	workflowsBlocksStatsCmd.AddCommand(workflowsBlocksStatsGetCmd)
	workflowsBlocksCmd.AddCommand(workflowsBlocksStatsCmd)
}
