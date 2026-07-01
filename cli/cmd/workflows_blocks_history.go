//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

type workflowBlockConfigHistoryResponse struct {
	Data         []workflowBlockConfigHistoryVersion `json:"data"`
	ListMetadata retab.PaginationCursor              `json:"list_metadata"`
}

type workflowBlockConfigHistoryVersion struct {
	ID                            string                            `json:"id,omitempty"`
	BlockConfigVersionFingerprint string                            `json:"block_config_version_fingerprint"`
	BlockType                     string                            `json:"block_type,omitempty"`
	BlockLabel                    string                            `json:"block_label,omitempty"`
	ConfigSnapshot                map[string]interface{}            `json:"config_snapshot,omitempty"`
	FirstSeenAt                   string                            `json:"first_seen_at,omitempty"`
	LastSeenAt                    string                            `json:"last_seen_at,omitempty"`
	PublishEpochs                 []workflowBlockConfigPublishEpoch `json:"publish_epochs,omitempty"`
	RunCount                      *int                              `json:"run_count,omitempty"`
	MatchesCurrentDraft           bool                              `json:"matches_current_draft,omitempty"`
	IsCurrentPublished            bool                              `json:"is_current_published,omitempty"`
	IsCurrent                     bool                              `json:"is_current,omitempty"`
}

type workflowBlockConfigPublishEpoch struct {
	ID                string `json:"id,omitempty"`
	PublishVersion    int    `json:"publish_version,omitempty"`
	PublishedAt       string `json:"published_at,omitempty"`
	SnapshotID        string `json:"snapshot_id,omitempty"`
	WorkflowVersionID string `json:"workflow_version_id,omitempty"`
}

var workflowsBlocksHistoryCmd = &cobra.Command{
	Use:   "history [<workflow-id>] <block-id>",
	Short: "List workflow block config history",
	Long: `List dashboard block config-history eras for one block.

This is different from ` + "`workflows blocks versions`" + `: versions are immutable
published block snapshots, while config history groups identical draft/published
configurations into eras with first/last seen timestamps, publish epochs, run
counts, and current-draft/current-published flags.`,
	Example: `  # List config-history eras for a block
  retab workflows blocks history wf_abc123 blk_extract

  # Equivalent flag form
  retab workflows blocks history --workflow-id wf_abc123 blk_extract

  # Explicit list subcommand
  retab workflows blocks history list wf_abc123 blk_extract --limit 20`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(runWorkflowsBlocksHistoryList),
}

var workflowsBlocksHistoryListCmd = &cobra.Command{
	Use:   "list [<workflow-id>] <block-id>",
	Short: "List workflow block config history",
	Long:  workflowsBlocksHistoryCmd.Long,
	Example: `  retab workflows blocks history list wf_abc123 blk_extract
  retab workflows blocks history list --workflow-id wf_abc123 blk_extract --limit 20`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(runWorkflowsBlocksHistoryList),
}

func runWorkflowsBlocksHistoryList(cmd *cobra.Command, args []string) error {
	if err := validateBeforeAfterMutex(cmd); err != nil {
		return err
	}
	workflowID, blockID, err := resolveWorkflowBlockScope(cmd, args, true, "history")
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("workflow_id", workflowID)
	params := collectListParams(cmd)
	if params.Before != nil {
		query.Set("before", *params.Before)
	}
	if params.After != nil {
		query.Set("after", *params.After)
	}
	if params.Limit != nil {
		query.Set("limit", strconv.Itoa(*params.Limit))
	}

	var result workflowBlockConfigHistoryResponse
	requestPath := "/v1/workflows/blocks/" + url.PathEscape(blockID) + "/config-history"
	if err := cliJSONRequestInto(cmd, http.MethodGet, requestPath, query, nil, &result); err != nil {
		return err
	}
	return printBlockConfigHistoryResult(cmd, result)
}

var blockConfigHistoryColumns = []TableColumn{
	{Header: "FINGERPRINT", Extract: func(row any) string { return blockConfigHistoryCell(row, "block_config_version_fingerprint") }},
	{Header: "TYPE", Extract: func(row any) string { return blockConfigHistoryCell(row, "block_type") }},
	{Header: "LABEL", Extract: func(row any) string { return blockConfigHistoryCell(row, "block_label") }},
	{Header: "FIRST_SEEN_AT", Extract: func(row any) string {
		return normalizeTimestampCell(blockConfigHistoryCell(row, "first_seen_at"))
	}},
	{Header: "LAST_SEEN_AT", Extract: func(row any) string {
		return normalizeTimestampCell(blockConfigHistoryCell(row, "last_seen_at"))
	}},
	{Header: "RUNS", Extract: func(row any) string { return blockConfigHistoryCell(row, "run_count") }},
	{Header: "CURRENT_DRAFT", Extract: func(row any) string { return blockConfigHistoryCell(row, "matches_current_draft") }},
	{Header: "CURRENT_PUBLISHED", Extract: func(row any) string { return blockConfigHistoryCell(row, "is_current_published") }},
	{Header: "PUBLISHES", Extract: blockConfigHistoryPublishCountCell},
}

func printBlockConfigHistoryResult(cmd *cobra.Command, result workflowBlockConfigHistoryResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		return RenderList(os.Stdout, format, result, blockConfigHistoryColumns)
	}
	return printJSON(result)
}

func blockConfigHistoryCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func blockConfigHistoryPublishCountCell(row any) string {
	value, ok := rowField(row, "publish_epochs")
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

func addBlockConfigHistoryListFlags(cmd *cobra.Command) {
	cmd.Flags().String("before", "", "history item id: return items before this id (mutually exclusive with --after)")
	cmd.Flags().String("after", "", "history item id: return items after this id (mutually exclusive with --before)")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "max items to return (1-200)")
}

func init() {
	workflowsBlocksHistoryCmd.PersistentFlags().String("workflow-id", "", "workflow id (required unless passed positionally)")
	addBlockConfigHistoryListFlags(workflowsBlocksHistoryCmd)
	addBlockConfigHistoryListFlags(workflowsBlocksHistoryListCmd)
	workflowsBlocksHistoryCmd.AddCommand(workflowsBlocksHistoryListCmd)
	workflowsBlocksCmd.AddCommand(workflowsBlocksHistoryCmd)
}
