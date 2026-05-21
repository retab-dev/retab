package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
// passed. Mixing both flags is allowed — the maps are unioned, with the
// new `--document` flag winning on key collision — and still produces
// exactly one warning line.
//
// Each entry must be of the form `block-id=path`. Empty keys, empty values,
// and entries without an `=` produce an error.
func parseDocumentArgs(docs []string, docFiles []string, warnTo io.Writer) (map[string]string, error) {
	out := map[string]string{}
	if err := appendKVPairs(out, docs, "--document"); err != nil {
		return nil, err
	}
	if len(docFiles) > 0 {
		if warnTo != nil {
			fmt.Fprintln(warnTo, "warning: --document-file is deprecated for workflows runs create; use --document <block-id>=<path>")
		}
		// Stage the deprecated entries first, then let --document overwrite
		// on key collision so the new flag wins when both are passed.
		staged := map[string]string{}
		if err := appendKVPairs(staged, docFiles, "--document-file"); err != nil {
			return nil, err
		}
		for k, v := range out {
			staged[k] = v
		}
		out = staged
	}
	return out, nil
}

func appendKVPairs(into map[string]string, raws []string, flagName string) error {
	for _, raw := range raws {
		key, path, ok := splitKV(raw)
		if !ok {
			return fmt.Errorf("%s expects block-id=path, got %q", flagName, raw)
		}
		if key == "" || path == "" {
			return fmt.Errorf("%s expects block-id=path, got %q", flagName, raw)
		}
		into[key] = path
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
	"email":    true,
	"restart":  true,
}

var allowedWorkflowRunExportSources = map[string]bool{
	"outputs": true,
	"inputs":  true,
}

const workflowRunStatusValues = "pending, running, completed, error, awaiting_review, cancelled"
const workflowRunTriggerTypeValues = "manual, api, schedule, webhook, email, restart"
const workflowRunExportSourceValues = "outputs, inputs"

// Step statuses are a superset of run statuses: orchestrator-level
// ``skipped`` shows up here but not on a run as a whole. Kept aligned
// with `StepStatusLiteral` in `backend/.../steps/query.py`.
var allowedWorkflowStepStatuses = map[string]bool{
	"pending":         true,
	"running":         true,
	"completed":       true,
	"skipped":         true,
	"error":           true,
	"awaiting_review": true,
	"cancelled":       true,
}

const workflowStepStatusValues = "pending, running, completed, skipped, error, awaiting_review, cancelled"

func validateWorkflowRunsListFilters(cmd *cobra.Command) error {
	if err := validateEnumFlag(cmd, "status", allowedWorkflowRunStatuses, workflowRunStatusValues); err != nil {
		return err
	}
	if err := validateEnumArrayFlag(cmd, "statuses", allowedWorkflowRunStatuses, workflowRunStatusValues); err != nil {
		return err
	}
	if err := validateEnumFlag(cmd, "exclude-status", allowedWorkflowRunStatuses, workflowRunStatusValues); err != nil {
		return err
	}
	if err := validateEnumFlag(cmd, "trigger-type", allowedWorkflowRunTriggerTypes, workflowRunTriggerTypeValues); err != nil {
		return err
	}
	return validateEnumArrayFlag(cmd, "trigger-types", allowedWorkflowRunTriggerTypes, workflowRunTriggerTypeValues)
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
	return validateEnumArrayFlag(cmd, "trigger-types", allowedWorkflowRunTriggerTypes, workflowRunTriggerTypeValues)
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

func validateEnumArrayFlag(cmd *cobra.Command, flagName string, allowed map[string]bool, allowedValues string) error {
	_, err := normalizeEnumArrayFlag(cmd, flagName, allowed, allowedValues)
	return err
}

func normalizeEnumArrayFlag(cmd *cobra.Command, flagName string, allowed map[string]bool, allowedValues string) ([]string, error) {
	rawValues, _ := cmd.Flags().GetStringArray(flagName)
	values := make([]string, 0, len(rawValues))
	for _, raw := range rawValues {
		for _, value := range strings.Split(raw, ",") {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if !allowed[value] {
				return nil, fmt.Errorf("invalid --%s %q (want: %s)", flagName, value, allowedValues)
			}
			values = append(values, value)
		}
	}
	return values, nil
}

var workflowsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Manage workflow runs",
	Long: `Execute, inspect, cancel, and replay workflow runs.

A run is one execution of a workflow against a set of inputs. Use this
subgroup to start runs (` + "`create`" + `), watch their lifecycle
(` + "`get`" + `, ` + "`workflows steps list`" + `), or restart failed runs
(` + "`restart`" + `).

Review-based: when a block pauses for review, the run enters status
` + "`awaiting_review`" + `. Decide reviewed block runs with the sibling
` + "`retab workflows reviews`" + ` command group —
` + "`reviews list`" + ` for the queue, ` + "`reviews get`" + ` to inspect
a block run awaiting review, then ` + "`reviews approve`" + ` or ` + "`reviews reject`" + `
to decide it. Use ` + "`reviews versions create`" + ` to create a corrected output before approving.

For declarative regression testing of workflow outputs, see
` + "`retab workflows tests --help`" + `.`,
	Example: `  # Start a run by uploading a document into the start_document block
  retab workflows runs create wf_abc123 \
    --document start=./invoice.pdf

  # Inspect a run's lifecycle (status, errors, timing)
  retab workflows runs get run_xyz789

  # Stream per-block execution records
  retab workflows steps list run_xyz789

  # Cancel an in-flight run
  retab workflows runs cancel run_xyz789

  # Approve a review by id
  retab workflows reviews approve rev_123 \
    --version-id rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA`,
}

var workflowsRunsCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Start a workflow run",
	Long: `Start one run of a workflow against the supplied inputs.

Inputs are keyed by the id of the input block they target. There are
three ways to supply them, mix and match as needed:

  ` + "`--document BLOCK=PATH`" + ` — upload a local file. The MIME
  type is inferred from the file extension. Repeatable.

  ` + "`--document-url BLOCK=URL`" + ` — pass a remote URL the server will
  fetch. Repeatable.

  ` + "`--documents-file PATH`" + ` — a JSON object mapping block ids to
  pre-built MIME-data payloads. Use this for advanced cases or when the
  documents come from upstream pipelines.

Add ` + "`--json-inputs-file PATH`" + ` for blocks that accept structured
JSON instead of (or alongside) documents.

By default the workflow's latest published version runs; pin a specific
version with ` + "`--version`" + `. Inspect the resulting run with
` + "`workflows runs get`" + ` or ` + "`workflows steps list`" + `.

The legacy ` + "`--document-file BLOCK=PATH`" + ` spelling is still
accepted as a deprecated alias for ` + "`--document`" + ` and will be
removed in a future release.`,
	Example: `  # Upload a local file into the start_document block
  retab workflows runs create wf_abc123 \
    --document start=./invoice.pdf

  # Multiple inputs across different blocks
  retab workflows runs create wf_abc123 \
    --document start=./invoice.pdf \
    --document-url reference=https://acme.com/po.pdf

  # JSON-only inputs (e.g. an api_call block)
  retab workflows runs create wf_abc123 \
    --json-inputs-file ./inputs.json

  # Pin to a specific published version
  retab workflows runs create wf_abc123 \
    --version v3 --document start=./invoice.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		req := retab.CreateWorkflowRunRequest{WorkflowID: args[0]}
		req.Version, _ = cmd.Flags().GetString("version")
		docsFile, _ := cmd.Flags().GetString("documents-file")
		if docsFile != "" {
			docs, err := readJSONMap(docsFile)
			if err != nil {
				return fmt.Errorf("--documents-file: %w", err)
			}
			req.Documents = docs
		}
		jsonInputsFile, _ := cmd.Flags().GetString("json-inputs-file")
		if jsonInputsFile != "" {
			inputs, err := readJSONMap(jsonInputsFile)
			if err != nil {
				return fmt.Errorf("--json-inputs-file: %w", err)
			}
			req.JSONInputs = inputs
		}
		docFlags, _ := cmd.Flags().GetStringArray("document")
		legacyFileFlags, _ := cmd.Flags().GetStringArray("document-file")
		fileEntries, err := parseDocumentArgs(docFlags, legacyFileFlags, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		urlFlags, _ := cmd.Flags().GetStringArray("document-url")
		if len(fileEntries) > 0 || len(urlFlags) > 0 {
			if req.Documents == nil {
				req.Documents = map[string]any{}
			}
			for key, path := range fileEntries {
				mime, err := inferFileMIMEData(path)
				if err != nil {
					return fmt.Errorf("--document %s=%s: %w", key, path, err)
				}
				req.Documents[key] = mime
			}
			for _, raw := range urlFlags {
				key, url, ok := splitKV(raw)
				if !ok || strings.TrimSpace(key) == "" || url == "" {
					return fmt.Errorf("--document-url expects block-id=url, got %q", raw)
				}
				// Server requires `filename` on every document descriptor;
				// derive from URL path's last segment (same rule applied
				// across single-document commands in common.go).
				req.Documents[key] = retab.MIMEData{
					Filename: filenameFromURL(url),
					URL:      url,
				}
			}
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if req.Documents != nil {
			req.Documents, err = resolveWorkflowRunDocumentAliases(ctx, client, args[0], req.Documents)
			if err != nil {
				return err
			}
		}
		body, err := workflowRunCreateRequestBody(req)
		if err != nil {
			return err
		}
		result, err := cliJSONRequest(cmd, http.MethodPost, "/workflows/runs", nil, body)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func workflowRunCreateRequestBody(request retab.CreateWorkflowRunRequest) (map[string]any, error) {
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
	blocks, err := client.Workflows.Blocks.List(ctx, workflowID)
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

var workflowsRunsGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get a workflow run",
	Long: `Fetch a run's metadata: status, trigger type, timestamps,
duration, cost, error info, final outputs. For per-block detail use
` + "`workflows steps list`" + `.`,
	Example: `  # Inspect a run
  retab workflows runs get run_xyz789

  # Poll until done
  while [ "$(retab workflows runs get run_xyz789 | jq -r '.lifecycle.status')" = "running" ]; do
    sleep 2
  done`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsRunsListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List workflow runs",
	Long: `List workflow runs. Without a workflow id the result spans the
whole workspace; with one (passed positionally OR via ` + "`--workflow-id`" + `,
not both) the result is scoped to that workflow. Other filters available:
status, trigger type, date range, cost, and duration. Page by run id
(` + "`--after`" + ` / ` + "`--before`" + ` / ` + "`--limit`" + `) and
` + "`--fields`" + ` to keep responses small on busy projects.`,
	Example: `  # Scope to a single workflow (positional, matches the rest of workflows)
  retab workflows runs list wf_abc123 --limit 50

  # Same, but with the flag form when composing many filters
  retab workflows runs list --workflow-id wf_abc123 --status error --from-date 2026-05-13

  # Workspace-wide (omit the workflow id entirely)
  retab workflows runs list --status error --limit 50`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateWorkflowRunsListFilters(cmd); err != nil {
			return err
		}
		// Honour <workflow-id> positional (matches blocks list / edges
		// create convention). The flag form stays supported. If both are
		// set and they disagree, error
		// — silently picking one would mask real user mistakes.
		flagID, _ := cmd.Flags().GetString("workflow-id")
		var posID string
		if len(args) == 1 {
			posID = args[0]
		}
		effectiveID := ""
		switch {
		case posID != "" && flagID != "" && posID != flagID:
			return fmt.Errorf("workflow id specified twice (positional %q, --workflow-id %q)", posID, flagID)
		case posID != "":
			effectiveID = posID
		case flagID != "":
			effectiveID = flagID
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowRunsParams{}
		params.WorkflowID = effectiveID
		params.Status, _ = cmd.Flags().GetString("status")
		statuses, err := normalizeEnumArrayFlag(cmd, "statuses", allowedWorkflowRunStatuses, workflowRunStatusValues)
		if err != nil {
			return err
		}
		params.Statuses = statuses
		params.ExcludeStatus, _ = cmd.Flags().GetString("exclude-status")
		params.TriggerType, _ = cmd.Flags().GetString("trigger-type")
		triggerTypes, err := normalizeEnumArrayFlag(cmd, "trigger-types", allowedWorkflowRunTriggerTypes, workflowRunTriggerTypeValues)
		if err != nil {
			return err
		}
		params.TriggerTypes = triggerTypes
		params.FromDate, _ = cmd.Flags().GetString("from-date")
		params.ToDate, _ = cmd.Flags().GetString("to-date")
		params.Search, _ = cmd.Flags().GetString("search")
		params.SortBy, _ = cmd.Flags().GetString("sort-by")
		fields, err := nonBlankStringArrayFlag(cmd, "fields")
		if err != nil {
			return err
		}
		params.Fields = fields
		params.Before, _ = cmd.Flags().GetString("before")
		params.After, _ = cmd.Flags().GetString("after")
		params.Limit, _ = cmd.Flags().GetInt("limit")
		params.Order, _ = cmd.Flags().GetString("order")
		if cmd.Flags().Changed("min-cost") {
			v, _ := cmd.Flags().GetFloat64("min-cost")
			params.MinCost = &v
		}
		if cmd.Flags().Changed("max-cost") {
			v, _ := cmd.Flags().GetFloat64("max-cost")
			params.MaxCost = &v
		}
		if cmd.Flags().Changed("min-duration") {
			v, _ := cmd.Flags().GetInt("min-duration")
			params.MinDuration = &v
		}
		if cmd.Flags().Changed("max-duration") {
			v, _ := cmd.Flags().GetInt("max-duration")
			params.MaxDuration = &v
		}
		result, err := client.Workflows.Runs.List(ctx, &params)
		if err != nil {
			return err
		}
		return printWorkflowRunListResult(cmd, result)
	}),
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

var workflowsRunsDeleteCmd = &cobra.Command{
	Use:   "delete <run-id>",
	Short: "Delete a workflow run",
	Long: `Permanently delete a run and its step records. Artifacts
produced by the run are preserved separately (see
` + "`workflows artifacts`" + `).`,
	Example: `  # Delete a run
  retab workflows runs delete run_xyz789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Workflows.Runs.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("run", args[0])
		return nil
	}),
}

var workflowsRunsCancelCmd = &cobra.Command{
	Use:   "cancel <run-id>",
	Short: "Cancel a workflow run",
	Long: `Cancel an in-flight run. Already-completed blocks keep their
outputs; in-progress blocks are terminated. Use ` + "`--command-id`" + ` to
make the cancel idempotent if you may retry the request.`,
	Example: `  # Cancel a stuck run
  retab workflows runs cancel run_xyz789

  # Idempotent cancel (safe to retry on network errors)
  retab workflows runs cancel run_xyz789 --command-id cancel-once-123`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		commandID, _ := cmd.Flags().GetString("command-id")
		result, err := client.Workflows.Runs.Cancel(ctx, args[0], retab.WorkflowRunCommandRequest{CommandID: commandID})
		if err != nil {
			return err
		}
		// The server returns 200 even when the cancel signal hasn't been
		// fully delivered yet (Temporal may be backed up; the engine flag
		// stays at ``cancel_confirmation_pending``). Surface that
		// non-finalized state to stderr so a user piping the response
		// into a script can spot it. The JSON on stdout still carries
		// the full ``cancellation_status`` field for programmatic
		// consumers.
		if result != nil && result.CancellationStatus != "" && result.CancellationStatus != "cancelled" {
			fmt.Fprintf(
				os.Stderr,
				"note: cancellation_status=%q — the cancel request was accepted but the run has not yet reached a terminal state. Poll `retab workflows runs get %s` until lifecycle.status is one of cancelled / completed / error.\n",
				result.CancellationStatus,
				args[0],
			)
		}
		return printResult(cmd, result)
	}),
}

var workflowsRunsRestartCmd = &cobra.Command{
	Use:   "restart <run-id>",
	Short: "Restart a workflow run",
	Long: `Re-execute a failed or cancelled run, reusing the original
inputs. By default the restarted run uses the latest published workflow
config. Use ` + "`--config-source draft`" + ` after tweaking draft block config,
or ` + "`--command-id`" + ` for idempotency.`,
	Example: `  # Restart a failed run
  retab workflows runs restart run_xyz789

  # Restart against the current draft config
 retab workflows runs restart run_xyz789 --config-source draft`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		commandID, _ := cmd.Flags().GetString("command-id")
		configSourceValue, _ := cmd.Flags().GetString("config-source")
		configSource, err := parseWorkflowRunConfigSource(configSourceValue)
		if err != nil {
			return err
		}
		body := map[string]any{
			"restart_of":    args[0],
			"config_source": configSource,
		}
		if commandID != "" {
			body["command_id"] = commandID
		}
		result, err := cliJSONRequest(cmd, http.MethodPost, "/workflows/runs", nil, body)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsRunsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export workflow runs as CSV",
	Long: `Bulk-export the outputs of many runs of a workflow as CSV,
focused on one block's output. Filter by status, date range, trigger
type, or an explicit set of run ids. Use ` + "`--preferred-column`" + ` to
pin column order.`,
	Example: `  # Export every successful run of a block to CSV
  retab workflows runs export \
    --workflow-id wf_abc123 --block-id blk_extract_1 \
    --status completed \
    --from-date 2026-05-01

  # Export a specific set of runs with custom columns
  retab workflows runs export \
    --workflow-id wf_abc123 --block-id blk_extract_1 \
    --run-id run_xyz789 --run-id run_aaa000 \
    --preferred-column invoice_id --preferred-column total`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateWorkflowRunsExportFilters(cmd); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.ExportWorkflowRunsRequest{}
		req.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		req.BlockID, _ = cmd.Flags().GetString("block-id")
		req.ExportSource, _ = cmd.Flags().GetString("export-source")
		selectedRunIDs, err := nonBlankStringArrayFlag(cmd, "run-id")
		if err != nil {
			return err
		}
		req.SelectedRunIDs = selectedRunIDs
		req.Status, _ = cmd.Flags().GetString("status")
		req.ExcludeStatus, _ = cmd.Flags().GetString("exclude-status")
		req.FromDate, _ = cmd.Flags().GetString("from-date")
		req.ToDate, _ = cmd.Flags().GetString("to-date")
		triggerTypes, err := normalizeEnumArrayFlag(cmd, "trigger-types", allowedWorkflowRunTriggerTypes, workflowRunTriggerTypeValues)
		if err != nil {
			return err
		}
		req.TriggerTypes = triggerTypes
		preferredColumns, err := nonBlankStringArrayFlag(cmd, "preferred-column")
		if err != nil {
			return err
		}
		req.PreferredColumns = preferredColumns
		result, err := client.Workflows.Runs.Export(ctx, req)
		if err != nil {
			return err
		}
		// The export endpoint wraps the CSV bytes in a JSON envelope
		// (``{"csv_data": "...", "rows": N, "columns": M}``). For
		// interactive / non-JSON output, dump the raw CSV directly so
		// the user can pipe straight into a file or another tool
		// (``... > out.csv``). JSON output keeps the full envelope for
		// programmatic consumers that need the row / column counts.
		raw, _ := cmd.Flags().GetBool("raw")
		var outFormat string
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			outFormat = f.Value.String()
		}
		if raw || outFormat == "table" || (outFormat == "" && term.IsTerminal(int(os.Stdout.Fd()))) {
			if result != nil {
				_, err := os.Stdout.WriteString(result.CSVData)
				if err == nil && !strings.HasSuffix(result.CSVData, "\n") {
					_, err = os.Stdout.WriteString("\n")
				}
				return err
			}
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

// ---- steps subgroup ----

var workflowsStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "Inspect workflow run steps",
	Long: `A step is the execution record of one block within a run —
input it saw, output it produced, error if any, timing, cost. Steps are
the entry point for debugging: when a run output looks wrong, list its
steps and pull the offending one with ` + "`workflows steps get`" + `
to see exactly what the block did.`,
	Example: `  # List every step in a run
  retab workflows steps list run_xyz789

  # Pull the full execution record for one step
  retab workflows steps get step_extract_1`,
}

var workflowsStepsListCmd = &cobra.Command{
	Use:   "list <run-id>",
	Short: "List run steps",
	Long: `List every step in a run — one record per block execution.
Includes status, timing, and a summary of input/output sizes. For the
full input/output payload of one step use ` + "`steps get`" + `.`,
	Example: `  # List steps
  retab workflows steps list run_xyz789

  # Find the first failed step
  retab workflows steps list run_xyz789 \
    | jq '.data[] | select(.lifecycle.status == "error") | .block_id' | head -1`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Steps.List(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsStepsGetCmd = &cobra.Command{
	Use:   "get <step-id>",
	Short: "Get the full step execution record",
	Long: `Return everything about one step execution in a run: the
exact input payload, the produced output, any error, timing, cost, and
model usage if applicable.

This is the canonical entry point when debugging a run that produced
the wrong output — find the offending step id, inspect its inputs, and
correlate against the step's block config.`,
	Example: `  # Pull the full record for a single step
  retab workflows steps get step_extract_1

  # Save the input payload for offline replay
  retab workflows steps get step_extract_1 \
    | jq '.handle_inputs' > inputs.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Steps.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsStepsQueryCmd = &cobra.Command{
	Use:   "query <workflow-id>",
	Short: "Query joined workflow step rows",
	Long: `Query joined step rows for a workflow. Use this when you need the
flat step record joined with its execution fingerprint for debugging,
comparison, or experiment analysis.`,
	Example: `  # Query recent extract steps for a workflow
  retab workflows steps query wf_abc123 --block-type extract --status completed --limit 20`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()

		if err := validateEnumArrayFlag(cmd, "status", allowedWorkflowStepStatuses, workflowStepStatusValues); err != nil {
			return err
		}

		request := retab.StepsQueryRequest{WorkflowID: args[0]}
		request.BlockID, _ = cmd.Flags().GetString("block-id")
		request.BlockType, _ = cmd.Flags().GetString("block-type")
		request.SourceKind, _ = cmd.Flags().GetString("source-kind")
		request.Status, _ = cmd.Flags().GetStringArray("status")
		if cmd.Flags().Changed("limit") {
			request.Limit, _ = cmd.Flags().GetInt("limit")
		}

		result, err := client.Workflows.Steps.Query(ctx, request)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsRunsCreateCmd.Flags().String("version", "", "workflow version (defaults to production)")
	workflowsRunsCreateCmd.Flags().String("documents-file", "", "JSON object {block-id: document} (or - for stdin)")
	workflowsRunsCreateCmd.Flags().StringArray("document", nil, "document file as block-id=path (repeatable)")
	// `--document-file` collided with the `--document-file <json-path>` flag
	// every other primitive command exposes via addDocumentFlags. Kept here
	// as a hidden deprecated alias for one release cycle; parseDocumentArgs
	// emits a single stderr warning when it's used.
	workflowsRunsCreateCmd.Flags().StringArray("document-file", nil, "deprecated alias for --document")
	_ = workflowsRunsCreateCmd.Flags().MarkHidden("document-file")
	workflowsRunsCreateCmd.Flags().StringArray("document-url", nil, "document url as block-id=url (repeatable)")
	workflowsRunsCreateCmd.Flags().String("json-inputs-file", "", "JSON inputs object (or - for stdin)")

	workflowsRunsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsRunsListCmd.Flags().String("status", "", "filter by status")
	workflowsRunsListCmd.Flags().StringArray("statuses", nil, "filter by status (repeatable)")
	workflowsRunsListCmd.Flags().String("exclude-status", "", "exclude status")
	workflowsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsRunsListCmd.Flags().StringArray("trigger-types", nil, "filter by trigger types (repeatable)")
	workflowsRunsListCmd.Flags().Var(&dateFlagValue{}, "from-date", "filter from this YYYY-MM-DD date")
	workflowsRunsListCmd.Flags().Var(&dateFlagValue{}, "to-date", "filter to this YYYY-MM-DD date")
	workflowsRunsListCmd.Flags().String("search", "", "search query")
	workflowsRunsListCmd.Flags().Var(newEnumStringFlagValue("--sort-by", "timing.created_at", "timing.started_at"), "sort-by", "sort field: timing.created_at | timing.started_at")
	workflowsRunsListCmd.Flags().StringArray("fields", nil, "include only these fields (repeatable)")
	workflowsRunsListCmd.Flags().String("before", "", "run id: return items before this id")
	workflowsRunsListCmd.Flags().String("after", "", "run id: return items after this id")
	workflowsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	workflowsRunsListCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
	workflowsRunsListCmd.Flags().Var(&nonNegativeFloatFlagValue{}, "min-cost", "min cost")
	workflowsRunsListCmd.Flags().Var(&nonNegativeFloatFlagValue{}, "max-cost", "max cost")
	workflowsRunsListCmd.Flags().Var(&nonNegativeIntFlagValue{}, "min-duration", "min duration (ms)")
	workflowsRunsListCmd.Flags().Var(&nonNegativeIntFlagValue{}, "max-duration", "max duration (ms)")

	workflowsRunsCancelCmd.Flags().String("command-id", "", "idempotency command id")
	workflowsRunsRestartCmd.Flags().String("command-id", "", "idempotency command id")
	workflowsRunsRestartCmd.Flags().String("config-source", "published", "published | draft")

	workflowsRunsExportCmd.Flags().String("workflow-id", "", "workflow id (required)")
	workflowsRunsExportCmd.Flags().String("block-id", "", "block id (required)")
	workflowsRunsExportCmd.Flags().String("export-source", "", "export source (default outputs)")
	workflowsRunsExportCmd.Flags().StringArray("run-id", nil, "filter to selected run ids (repeatable)")
	workflowsRunsExportCmd.Flags().String("status", "", "status filter")
	workflowsRunsExportCmd.Flags().String("exclude-status", "", "exclude status")
	workflowsRunsExportCmd.Flags().Var(&dateFlagValue{}, "from-date", "from YYYY-MM-DD date")
	workflowsRunsExportCmd.Flags().Var(&dateFlagValue{}, "to-date", "to YYYY-MM-DD date")
	workflowsRunsExportCmd.Flags().StringArray("trigger-types", nil, "trigger types (repeatable)")
	workflowsRunsExportCmd.Flags().StringArray("preferred-column", nil, "preferred CSV column (repeatable)")
	workflowsRunsExportCmd.Flags().Bool("raw", false, "write raw CSV to stdout (default for TTY); JSON envelope is used for non-TTY unless --raw or --output table")
	_ = workflowsRunsExportCmd.MarkFlagRequired("workflow-id")
	_ = workflowsRunsExportCmd.MarkFlagRequired("block-id")

	workflowsStepsQueryCmd.Flags().String("block-id", "", "filter by block id")
	workflowsStepsQueryCmd.Flags().String("block-type", "", "filter by block type")
	workflowsStepsQueryCmd.Flags().String("source-kind", "", "filter by source kind")
	workflowsStepsQueryCmd.Flags().StringArray("status", nil, "filter by lifecycle status (repeatable)")
	workflowsStepsQueryCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 1000}, "limit", "max rows to return (1-1000)")

	workflowsStepsCmd.AddCommand(workflowsStepsListCmd, workflowsStepsGetCmd, workflowsStepsQueryCmd)
	workflowsCmd.AddCommand(workflowsStepsCmd)
	workflowsRunsCmd.AddCommand(workflowsRunsCreateCmd, workflowsRunsGetCmd, workflowsRunsListCmd, workflowsRunsDeleteCmd, workflowsRunsCancelCmd, workflowsRunsRestartCmd, workflowsRunsExportCmd)
	workflowsCmd.AddCommand(workflowsRunsCmd)
}
