//go:build retab_oagen_cli_workflows_runs

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// parseDocumentArgs merges the new `--document` flag with its deprecated
// `--document-file` alias into a single block-id → path map.
//
// `--document-file` was renamed to `--document` because it collided with the
// `--document-file <json-path>` flag every other primitive command exposes
// via addDocumentFlags. The two surfaces had identical names but completely
// different protocols (single JSON descriptor vs repeatable block-id=path
// pair), which produced confusing errors when users carried muscle memory
// from one surface to the other.
//
// The deprecated alias is preserved for one release cycle so existing
// scripts keep working. When the alias is used at least once, a single
// warning line is written to warnTo regardless of how many values were
// passed.
//
// A given block-id must appear at most once across BOTH flags. Repeating a
// block id inside the same flag, or having the same block id in both flags,
// is a hard error — silently letting the last entry win would mask user
// typos and produce surprising "which one ran?" behavior at runtime.
//
// Each entry must be of the form `block-id=path`. Empty keys, empty values,
// and entries without an `=` produce an error.
func parseDocumentArgs(docs []string, docFiles []string, warnTo io.Writer) (map[string]string, error) {
	out := map[string]string{}
	// `source` tracks which flag claimed each block id so cross-flag
	// collisions can name both flags in the error message.
	source := map[string]string{}
	if err := appendKVPairs(out, source, docs, "--document"); err != nil {
		return nil, err
	}
	if len(docFiles) > 0 {
		if warnTo != nil {
			if _, err := fmt.Fprintln(warnTo, "warning: --document-file is deprecated for workflows runs create; use --document <block-id>=<path>"); err != nil {
				return nil, err
			}
		}
		if err := appendKVPairs(out, source, docFiles, "--document-file"); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func appendKVPairs(into map[string]string, source map[string]string, raws []string, flagName string) error {
	for _, raw := range raws {
		key, path, ok := splitKV(raw)
		if !ok {
			return fmt.Errorf("%s expects block-id=path, got %q", flagName, raw)
		}
		if key == "" || path == "" {
			return fmt.Errorf("%s expects block-id=path, got %q", flagName, raw)
		}
		if prevFlag, exists := source[key]; exists {
			if prevFlag == flagName {
				return fmt.Errorf(
					"block %q passed twice via %s; each block id must appear at most once",
					key, flagName,
				)
			}
			return fmt.Errorf(
				"block %q has both %s and %s; pass exactly one source per block",
				key, prevFlag, flagName,
			)
		}
		into[key] = path
		source[key] = flagName
	}
	return nil
}

func parseWorkflowRunConfigSource(value string) (string, error) {
	if value == "" {
		return "published", nil
	}
	switch value {
	case "published", "draft":
		return value, nil
	default:
		return "", fmt.Errorf("--config-source must be published or draft, got %q", value)
	}
}

var allowedWorkflowRunStatuses = map[string]bool{
	"pending":         true,
	"running":         true,
	"completed":       true,
	"error":           true,
	"awaiting_review": true,
	"cancelled":       true,
}

var allowedWorkflowRunTriggerTypes = map[string]bool{
	"manual":   true,
	"api":      true,
	"schedule": true,
	"webhook":  true,
	"restart":  true,
}

var allowedWorkflowRunExportSources = map[string]bool{
	"outputs": true,
	"inputs":  true,
}

const workflowRunStatusValues = "pending, running, completed, error, awaiting_review, cancelled"
const workflowRunTriggerTypeValues = "manual, api, schedule, webhook, restart"
const workflowRunExportSourceValues = "outputs, inputs"

func validateWorkflowRunsListFilters(cmd *cobra.Command) error {
	if err := validateEnumFlag(cmd, "status", allowedWorkflowRunStatuses, workflowRunStatusValues); err != nil {
		return err
	}
	if err := validateEnumFlag(cmd, "exclude-status", allowedWorkflowRunStatuses, workflowRunStatusValues); err != nil {
		return err
	}
	return validateEnumFlag(cmd, "trigger-type", allowedWorkflowRunTriggerTypes, workflowRunTriggerTypeValues)
}

func validateWorkflowRunsExportFilters(cmd *cobra.Command) error {
	if err := validateEnumFlag(cmd, "export-source", allowedWorkflowRunExportSources, workflowRunExportSourceValues); err != nil {
		return err
	}
	if err := validateEnumFlag(cmd, "status", allowedWorkflowRunStatuses, workflowRunStatusValues); err != nil {
		return err
	}
	if err := validateEnumFlag(cmd, "exclude-status", allowedWorkflowRunStatuses, workflowRunStatusValues); err != nil {
		return err
	}
	return validateEnumFlag(cmd, "trigger-type", allowedWorkflowRunTriggerTypes, workflowRunTriggerTypeValues)
}

func validateMutuallyExclusiveChangedFlags(cmd *cobra.Command, left string, right string) error {
	if cmd.Flags().Changed(left) && cmd.Flags().Changed(right) {
		return fmt.Errorf("--%s and --%s cannot be used together", left, right)
	}
	return nil
}

func validateEnumFlag(cmd *cobra.Command, flagName string, allowed map[string]bool, allowedValues string) error {
	value, _ := cmd.Flags().GetString(flagName)
	if value == "" {
		return nil
	}
	if !allowed[value] {
		return fmt.Errorf("invalid --%s %q (want: %s)", flagName, value, allowedValues)
	}
	return nil
}

type workflowRunCreateParams struct {
	WorkflowID string
	Documents  map[string]any
	JSONInputs map[string]any
	Version    string
}

func workflowRunCreateRequestBody(request workflowRunCreateParams) (map[string]any, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}
	body := map[string]any{"workflow_id": request.WorkflowID}
	if request.Documents != nil {
		documents, err := workflowRunDocumentsRequestBody(request.Documents)
		if err != nil {
			return nil, err
		}
		body["documents"] = documents
	}
	if request.JSONInputs != nil {
		body["json_inputs"] = request.JSONInputs
	}
	if request.Version == "" {
		body["version"] = "production"
	} else {
		body["version"] = request.Version
	}
	return body, nil
}

func workflowRunDocumentsRequestBody(documents map[string]any) (map[string]map[string]string, error) {
	body := map[string]map[string]string{}
	for blockID, document := range documents {
		descriptor, err := workflowRunDocumentRequestBody(blockID, document)
		if err != nil {
			return nil, err
		}
		body[blockID] = descriptor
	}
	return body, nil
}

func workflowRunDocumentRequestBody(blockID string, document any) (map[string]string, error) {
	switch value := document.(type) {
	case retab.MIMEData:
		return workflowRunMIMEDataRequestBody(blockID, value), nil
	case *retab.MIMEData:
		if value == nil {
			return nil, fmt.Errorf("workflow run document %s must not be nil", blockID)
		}
		return workflowRunMIMEDataRequestBody(blockID, *value), nil
	case map[string]string:
		return normalizeWorkflowRunDocumentRequestBody(blockID, value), nil
	case map[string]any:
		descriptor := map[string]string{}
		for _, key := range []string{"filename", "url", "content", "mime_type"} {
			raw, ok := value[key]
			if !ok {
				continue
			}
			text, ok := raw.(string)
			if !ok {
				return nil, fmt.Errorf("workflow run document %s field %q must be a string", blockID, key)
			}
			if text != "" {
				descriptor[key] = text
			}
		}
		return normalizeWorkflowRunDocumentRequestBody(blockID, descriptor), nil
	default:
		mimeData, err := retab.InferMIMEData(document)
		if err != nil {
			return nil, fmt.Errorf("workflow run document %s must be a document descriptor or supported MIME input: %w", blockID, err)
		}
		return workflowRunMIMEDataRequestBody(blockID, mimeData), nil
	}
}

func workflowRunMIMEDataRequestBody(blockID string, mimeData retab.MIMEData) map[string]string {
	descriptor := map[string]string{}
	if mimeData.Filename != "" {
		descriptor["filename"] = mimeData.Filename
	}
	if mimeData.URL != "" {
		descriptor["url"] = mimeData.URL
	}
	if mimeData.Content != "" {
		descriptor["content"] = mimeData.Content
	}
	if mimeData.MIMEType != "" {
		descriptor["mime_type"] = mimeData.MIMEType
	}
	return normalizeWorkflowRunDocumentRequestBody(blockID, descriptor)
}

func normalizeWorkflowRunDocumentRequestBody(blockID string, descriptor map[string]string) map[string]string {
	normalized := map[string]string{}
	for key, value := range descriptor {
		if value != "" {
			normalized[key] = value
		}
	}
	if normalized["filename"] == "" {
		normalized["filename"] = blockID
	}
	return normalized
}

func resolveWorkflowRunDocumentAliases(
	ctx context.Context,
	client *retab.Client,
	workflowID string,
	documents map[string]any,
) (map[string]any, error) {
	if _, ok := documents["start"]; !ok {
		return documents, nil
	}
	blocks, err := client.Workflows.Blocks.List(ctx, &retab.WorkflowBlocksListParams{WorkflowID: workflowID})
	if err != nil {
		return nil, fmt.Errorf("resolve --document start alias: %w", err)
	}
	for _, block := range blocks.Data {
		if block.ID == "start" {
			return documents, nil
		}
	}
	var startDocumentBlocks []retab.WorkflowBlock
	for _, block := range blocks.Data {
		if isStartDocumentBlock(block) {
			startDocumentBlocks = append(startDocumentBlocks, block)
		}
	}
	if len(startDocumentBlocks) == 0 {
		return documents, nil
	}
	if len(startDocumentBlocks) > 1 {
		return nil, fmt.Errorf("--document start=... is ambiguous: workflow has %d start_document blocks; use the concrete block id", len(startDocumentBlocks))
	}
	resolved := make(map[string]any, len(documents))
	for key, value := range documents {
		if key == "start" {
			resolved[startDocumentBlocks[0].ID] = value
			continue
		}
		resolved[key] = value
	}
	return resolved, nil
}

func printWorkflowRunListResult(cmd *cobra.Command, result any) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	return RenderList(os.Stdout, format, result, workflowRunListColumns)
}

var workflowRunListColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return workflowRunCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return workflowRunCell(row, "workflow.name_at_run_time") }},
	{Header: "STATUS", Extract: func(row any) string { return workflowRunCell(row, "lifecycle.status") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return workflowRunCell(row, "timing.created_at") }},
}

func workflowRunCell(row any, key string) string {
	v, ok := rowField(row, key)
	if !ok || cellIsEmpty(v) {
		return ""
	}
	return stringifyCell(v)
}

var workflowsRunsRestartCmd = &cobra.Command{
	Use:   "restart <run-id>",
	Short: "Restart a workflow run",
	Long: `Re-execute a failed or cancelled run, reusing the original
inputs. By default the restarted run uses the latest published workflow
config. Use ` + "`--config-source draft`" + ` after tweaking draft block config.`,
	Example: `  # Restart a failed run
  retab workflows runs restart run_xyz789

  # Restart against the current draft config
  retab workflows runs restart run_xyz789 --config-source draft`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		configSourceValue, _ := cmd.Flags().GetString("config-source")
		configSource, err := parseWorkflowRunConfigSource(configSourceValue)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		sourceRun, err := client.Workflows.Runs.Get(ctx, args[0])
		if err != nil {
			return err
		}
		version := "production"
		if configSource == "draft" {
			version = "draft"
		}
		params := &retab.WorkflowRunsCreateParams{
			WorkflowID: sourceRun.WorkflowID,
			Version:    ptr(version),
		}
		if params.WorkflowID == "" {
			return fmt.Errorf("source run %s does not include workflow_id", args[0])
		}
		if sourceRun.Inputs != nil {
			if sourceRun.Inputs.Documents != nil {
				documents := map[string]interface{}{}
				for key, value := range sourceRun.Inputs.Documents {
					documents[key] = value
				}
				params.Documents = &documents
			}
			if sourceRun.Inputs.JSONData != nil {
				params.JSONInputs = &sourceRun.Inputs.JSONData
			}
		}
		result, err := client.Workflows.Runs.Create(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func nonBlankStringArrayFlag(cmd *cobra.Command, flagName string) ([]string, error) {
	values, _ := cmd.Flags().GetStringArray(flagName)
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("--%s must not be blank", flagName)
		}
	}
	return values, nil
}

func init() {
	workflowsRunsRestartCmd.Flags().String("config-source", "published", "published | draft")

	workflowsRunsCmd.AddCommand(workflowsRunsRestartCmd)
}
