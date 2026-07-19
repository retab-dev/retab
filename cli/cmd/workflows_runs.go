//go:build !retab_oagen_cli_workflows_runs

package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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
// passed.
//
// A given block-id must appear at most once across BOTH flags. Repeating a
// block id inside the same flag, or having the same block id in both flags,
// is a hard error — silently letting the last entry win would mask user
// typos and produce surprising "which one ran?" behavior at runtime.
//
// Each entry must be of the form `block-id=path`. Empty keys, empty values,
// and entries without an `=` produce an error.
// documentPathHint returns guidance when a `--document <block>=<value>` value
// looks like a file id or a misplaced flag rather than a local path. `--document`
// resolves a LOCAL FILE PATH; an already-uploaded file id belongs on
// `--document-id`, a URL on `--document-url`. The common stumble is
// `--document start=--file-id file_xxx` (or `--document start=file_xxx`), which
// reaches the file resolver as a literal path and fails with "file not found".
func documentPathHint(blockID, path string) string {
	trimmed := strings.TrimSpace(path)
	switch {
	case strings.HasPrefix(trimmed, "file_"):
		return fmt.Sprintf("hint: %q looks like a file id; use --document-id %s=%s for an already-uploaded file", trimmed, blockID, trimmed)
	case strings.HasPrefix(trimmed, "--file-id"), strings.HasPrefix(trimmed, "-"):
		return fmt.Sprintf("hint: --document takes a local file path; for an uploaded file id use --document-id %s=<file-id>, for a URL use --document-url %s=<url>", blockID, blockID)
	case strings.HasPrefix(trimmed, "http://"), strings.HasPrefix(trimmed, "https://"):
		return fmt.Sprintf("hint: %q looks like a URL; use --document-url %s=%s", trimmed, blockID, trimmed)
	default:
		return ""
	}
}

// isExternalFetchDocumentURL reports whether rawURL is an external http(s) URL
// that the workflow runs API refuses to fetch and that the CLI must therefore
// download and inline itself. Retab storage URLs (which the server resolves)
// and non-http schemes such as `data:` are left untouched for the server.
func isExternalFetchDocumentURL(rawURL string) bool {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return !strings.EqualFold(parsed.Hostname(), "storage.retab.com")
}

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
		// Trim the key like the --document-id/--document-url parsers do: an
		// untrimmed key (e.g. --document " start"=a.pdf) would dodge the
		// cross-flag duplicate/overlap checks (which compare trimmed keys
		// against this map) and dodge start-block alias resolution, reaching
		// the server as a bogus block id.
		key = strings.TrimSpace(key)
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

// These allowlists must stay in sync with the SDK enums the flags map into
// (WorkflowExportPayloadRequestExcludeStatus / ...TriggerType in
// clients/go/enums.go). Omitting a valid value makes the CLI's client-side
// validator stricter than the API, rejecting a legitimate filter before the
// request is ever sent. (Kept identical to the oagen overlay variant.)
var allowedWorkflowRunStatuses = map[string]bool{
	"pending":         true,
	"queued":          true,
	"running":         true,
	"completed":       true,
	"error":           true,
	"failed":          true,
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

const workflowRunStatusValues = "pending, queued, running, completed, error, failed, awaiting_review, cancelled"
const workflowRunTriggerTypeValues = "manual, api, schedule, webhook, email, restart"
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

var workflowsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Manage workflow runs",
	Long: `Execute, inspect, cancel, and replay workflow runs.

A run is one execution of a workflow against a set of inputs. Use this
subgroup to start runs (` + "`create`" + `), watch their lifecycle
(` + "`get`" + `, ` + "`workflows steps list`" + `), or replay prior inputs
(` + "`restart`" + `).

Review-based: when a block pauses for review, the run enters status
` + "`awaiting_review`" + `. Decide reviewed block runs with the sibling
` + "`retab workflows reviews`" + ` command group —
` + "`reviews list`" + ` for the queue, ` + "`reviews get`" + ` to inspect
a block run awaiting review, then ` + "`reviews approve`" + ` or ` + "`reviews reject`" + `
to decide it. Use ` + "`reviews versions create`" + ` to create a corrected output before approving.

For declarative regression testing of workflow outputs, see
` + "`retab workflows evals --help`" + `.`,
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
four ways to supply them, mix and match as needed:

  ` + "`--document BLOCK=PATH`" + ` — upload a local file. The MIME
  type is inferred from the file extension. Repeatable.

  ` + "`--document-url BLOCK=URL`" + ` — reference a document by URL. A
  Retab storage URL (` + "`https://storage.retab.com/...`" + `) is passed
  through for the server to resolve; any other external http(s) URL is
  downloaded and inlined by the CLI before upload, since the server does
  not fetch arbitrary external URLs. Repeatable.

  ` + "`--document-id BLOCK=FILE_ID`" + ` — reference a file you already
  uploaded (e.g. via ` + "`retab files upload`" + `) by its file id. The
  filename and MIME type are resolved from the stored file. Repeatable.

  ` + "`--documents-file PATH`" + ` — a JSON object mapping block ids to
  pre-built document payloads. Use this for advanced cases or when the
  documents come from upstream pipelines. Each value may be a MIME-data
  payload (` + "`filename`/`url`/`content`/`mime_type`" + `) or a file
  reference (` + "`id`/`filename`/`mime_type`" + `).

Add ` + "`--json-inputs-file PATH`" + ` for blocks that accept structured
JSON instead of (or alongside) documents.

By default the workflow's latest published version runs; pin a specific
version with ` + "`--version`" + `, or pass ` + "`--draft`" + ` to run the
editable draft without publishing it first (handy for iterating on a
workflow before it has a published version). Inspect the resulting run with
` + "`workflows runs get`" + ` or ` + "`workflows steps list`" + `.

Note: ` + "`runs create`" + ` executes the WHOLE graph, including side-effecting
tail blocks (e.g. an ` + "`api_call`" + ` that POSTs to an external system). To
test one block's draft config in isolation against a prior run's inputs — without
running downstream blocks — use ` + "`workflows blocks executions create <run-id> --block-id <id>`" + ` instead.

The legacy ` + "`--document-file BLOCK=PATH`" + ` spelling is still
accepted as a deprecated alias for ` + "`--document`" + ` and will be
removed in a future release.`,
	Example: `  # Upload a local file into the start_document block
  retab workflows runs create wf_abc123 \
    --document start=./invoice.pdf

  # Reference a previously-uploaded file by its id
  retab workflows runs create wf_abc123 \
    --document-id start=file_LPjuee2tTZgfM_Km5yh_G

  # Multiple inputs across different blocks
  retab workflows runs create wf_abc123 \
    --document start=./invoice.pdf \
    --document-url reference=https://acme.com/po.pdf

  # JSON-only inputs (e.g. an api_call block)
  retab workflows runs create wf_abc123 \
    --json-inputs-file ./inputs.json

  # Pin to a specific published version (production, draft, or a ver_... id)
  retab workflows runs create wf_abc123 \
    --version ver_xxx --document start=./invoice.pdf

  # Run the editable draft without publishing (alias for --version draft)
  retab workflows runs create wf_abc123 \
    --draft --document start=./invoice.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		req := workflowRunCreateParams{WorkflowID: args[0]}
		req.Version, _ = cmd.Flags().GetString("version")
		// --draft is a convenience alias for --version draft (run the editable
		// draft without publishing). It conflicts with an explicit --version that
		// names something other than the draft.
		if draft, _ := cmd.Flags().GetBool("draft"); draft {
			if req.Version != "" && req.Version != "draft" {
				return fmt.Errorf("--draft conflicts with --version %q (use one)", req.Version)
			}
			req.Version = "draft"
		}
		docsFile, _ := cmd.Flags().GetString("documents-file")
		if docsFile != "" {
			docs, err := readJSONMap(docsFile)
			if err != nil {
				return fmt.Errorf("--documents-file: %w", err)
			}
			req.Documents = docs
		}
		// Block ids already claimed by --documents-file. The per-flag conflict
		// checks below also test against these so a --document / --document-id
		// / --document-url for the same block errors out instead of silently
		// overwriting the documents-file entry.
		docsFileKeys := map[string]bool{}
		for key := range req.Documents {
			docsFileKeys[key] = true
		}
		jsonInputsFile, _ := cmd.Flags().GetString("json-inputs-file")
		if jsonInputsFile != "" {
			inputs, err := readJSONMap(jsonInputsFile)
			if err != nil {
				return fmt.Errorf("--json-inputs-file: %w", err)
			}
			req.JSONInputs = inputs
		}
		metaPairs, _ := cmd.Flags().GetStringArray("metadata")
		metadata, err := parseKVStringList(metaPairs)
		if err != nil {
			return fmt.Errorf("--metadata: %w", err)
		}
		if len(metadata) > 0 {
			req.Metadata = metadata
		}
		docFlags, _ := cmd.Flags().GetStringArray("document")
		legacyFileFlags, _ := cmd.Flags().GetStringArray("document-file")
		fileEntries, err := parseDocumentArgs(docFlags, legacyFileFlags, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		urlFlags, _ := cmd.Flags().GetStringArray("document-url")
		idFlags, _ := cmd.Flags().GetStringArray("document-id")
		// Parse --document-id BLOCK=FILE_ID into a block-id → file-id map,
		// rejecting duplicates and cross-flag collisions the same way
		// parseDocumentArgs does for --document / --document-file. A block
		// id may be claimed by exactly one source.
		docIDs := map[string]string{}
		for _, raw := range idFlags {
			key, fileID, ok := splitKV(raw)
			if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(fileID) == "" {
				return fmt.Errorf("--document-id expects block-id=file-id, got %q", raw)
			}
			key = strings.TrimSpace(key)
			if _, dup := docIDs[key]; dup {
				return fmt.Errorf("block %q passed twice via --document-id; each block id must appear at most once", key)
			}
			if _, conflict := fileEntries[key]; conflict {
				return fmt.Errorf("block %q has both --document and --document-id; pass exactly one source per block", key)
			}
			if docsFileKeys[key] {
				return fmt.Errorf("block %q has both --documents-file and --document-id; pass exactly one source per block", key)
			}
			docIDs[key] = strings.TrimSpace(fileID)
		}
		// Reject overlap between --document/--document-id and --document-url
		// for the same block id. Without this guard a later flag silently
		// overwrites the earlier one in req.Documents, producing surprising
		// "which one won?" behavior and (when the URL was the loser) a
		// 500 from the document-fetch path. Flag the conflict up-front.
		urlKeys := map[string]bool{}
		for _, raw := range urlFlags {
			key, _, ok := splitKV(raw)
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			if urlKeys[key] {
				return fmt.Errorf("block %q passed twice via --document-url; each block id must appear at most once", key)
			}
			urlKeys[key] = true
			if _, conflict := fileEntries[key]; conflict {
				return fmt.Errorf(
					"block %q has both --document and --document-url; pass exactly one source per block",
					key,
				)
			}
			if _, conflict := docIDs[key]; conflict {
				return fmt.Errorf(
					"block %q has both --document-id and --document-url; pass exactly one source per block",
					key,
				)
			}
			if docsFileKeys[key] {
				return fmt.Errorf(
					"block %q has both --documents-file and --document-url; pass exactly one source per block",
					key,
				)
			}
		}
		// Created up front (before the document loops) so the --document-url
		// loop can download + inline external URLs the server refuses to fetch.
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if len(fileEntries) > 0 || len(urlFlags) > 0 {
			if req.Documents == nil {
				req.Documents = map[string]any{}
			}
			for key, path := range fileEntries {
				if docsFileKeys[key] {
					return fmt.Errorf("block %q has both --documents-file and --document; pass exactly one source per block", key)
				}
				mime, err := inferFileMIMEData(path)
				if err != nil {
					if hint := documentPathHint(key, path); hint != "" {
						return fmt.Errorf("--document %s=%s: %w\n%s", key, path, err, hint)
					}
					return fmt.Errorf("--document %s=%s: %w", key, path, err)
				}
				req.Documents[key] = mime
			}
			for _, raw := range urlFlags {
				key, rawURL, ok := splitKV(raw)
				if !ok || strings.TrimSpace(key) == "" || rawURL == "" {
					return fmt.Errorf("--document-url expects block-id=url, got %q", raw)
				}
				// Trim the key like the conflict pre-pass does: an untrimmed
				// key would dodge the duplicate/overlap checks above and then
				// fail (or misroute) alias resolution downstream.
				key = strings.TrimSpace(key)
				// Server requires `filename` on every document descriptor;
				// derive from URL path's last segment (same rule applied
				// across single-document commands in common.go).
				doc := retab.MIMEData{
					Filename: filenameFromURL(rawURL),
					URL:      rawURL,
				}
				// The workflow runs API refuses to fetch arbitrary external
				// URLs (SSRF policy): it accepts only Retab storage URLs and
				// inline data: content, otherwise it 422s. So an external URL
				// like the `--document-url` help advertises must be downloaded
				// and re-inlined client-side, mirroring the edit-template
				// `--url` path. Retab storage and data: URLs pass through for
				// the server to resolve.
				if isExternalFetchDocumentURL(rawURL) {
					inlined, err := materializeInlineMIMEData(ctx, doc)
					if err != nil {
						return fmt.Errorf("--document-url %s=%s: %w", key, rawURL, err)
					}
					doc = inlined
				}
				req.Documents[key] = doc
			}
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		// Record each --document-id as a by-id FileRef; the reference
		// resolution pass below turns it into the durable storage.retab.com
		// URL the canonical create route accepts (alias resolution first
		// handles the `start` shorthand).
		if len(docIDs) > 0 {
			if req.Documents == nil {
				req.Documents = map[string]any{}
			}
			for key, fileID := range docIDs {
				req.Documents[key] = retab.FileRef{ID: fileID}
			}
		}
		if req.Documents != nil {
			req.Documents, err = resolveWorkflowRunDocumentAliases(ctx, client, args[0], req.Documents)
			if err != nil {
				return err
			}
			// Convert by-reference inputs (a stored file id, or inline base64
			// content) into url-backed MIMEData. The canonical
			// POST /v1/workflows/runs route accepts only url-backed documents;
			// file ids resolve to the durable storage.retab.com URL (the server
			// short-circuits it back to the stored object — no re-download), so
			// runs stay on the public route instead of the legacy per-workflow
			// run endpoint.
			if err := resolveWorkflowRunDocumentReferences(cmd, req.Documents); err != nil {
				return err
			}
		}
		// Resolve any remaining file-id document inputs (id references supplied
		// via --documents-file) into URL-backed MIMEData. --document-id is
		// resolved above; this covers the descriptor-map form. The workflow-runs
		// route accepts MIMEData only — a bare {id,...} body is rejected (422).
		if req.Documents != nil {
			for key, doc := range req.Documents {
				fileID, ok := workflowRunDocumentFileID(doc)
				if !ok {
					continue
				}
				resolved, err := resolveFileIDToMIMEDataWithClient(ctx, client, fileID)
				if err != nil {
					return fmt.Errorf("documents[%q]: resolving file id %s: %w", key, fileID, err)
				}
				if name := workflowRunDocumentFilename(doc); name != "" {
					resolved.Filename = name
				}
				req.Documents[key] = resolved
			}
		}
		if req.JSONInputs != nil {
			req.JSONInputs, err = resolveWorkflowRunJSONInputAliases(ctx, client, args[0], req.JSONInputs)
			if err != nil {
				return err
			}
		}
		body, err := workflowRunCreateRequestBody(req)
		if err != nil {
			return err
		}
		result, err := cliJSONRequest(cmd, http.MethodPost, "/v1/workflows/runs", nil, body)
		if err != nil {
			return err
		}
		return maybeWaitForWorkflowRun(cmd, result)
	}),
}

var workflowRunWaitTerminalStatuses = map[string]bool{
	"completed":       true,
	"error":           true,
	"cancelled":       true,
	"awaiting_review": true,
}

// maybeWaitForWorkflowRun prints the create response immediately unless
// --wait was passed; with --wait it polls the run until it reaches a
// terminal lifecycle status (completed / error / cancelled) or pauses for
// human review (awaiting_review), then prints the final run. This mirrors
// the --wait contract already on `extractions create` and
// `experiments runs create`, closing the gap where `workflows runs create`
// forced callers to hand-roll a poll loop around `runs get`.
func maybeWaitForWorkflowRun(cmd *cobra.Command, result any) error {
	if wait, _ := cmd.Flags().GetBool("wait"); !wait {
		return printResult(cmd, result)
	}
	resource, err := primitiveMap(result)
	if err != nil {
		return err
	}
	id, _ := resource["id"].(string)
	if id == "" {
		return fmt.Errorf("workflow run create response did not include an id")
	}
	return waitForWorkflowRunByID(cmd, id, resource)
}

// waitForWorkflowRunByID polls GET /v1/workflows/runs/<id> until the run
// reaches a terminal lifecycle status (completed/error/cancelled) or pauses
// for human review (awaiting_review), then prints the final record. `initial`
// is the most recently-seen run state — the create response for
// `runs create --wait`, or the first GET for the standalone `runs wait`
// command — so an already-terminal run short-circuits without an extra fetch.
// Shared by both entry points so the two stay in lockstep.
func waitForWorkflowRunByID(cmd *cobra.Command, id string, initial map[string]any) error {
	pollInterval, timeout := primitiveWaitDurations(cmd)
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()
	last := initial
	for {
		if status := primitiveStatus(last); workflowRunWaitTerminalStatuses[status] {
			if err := printResult(cmd, last); err != nil {
				return err
			}
			// error/cancelled are failures (non-zero exit), matching the
			// contract on every other run family (experiments, tests) and the
			// primitives. completed and awaiting_review — a pause for human
			// review, not a failure — exit 0.
			if status == "error" || status == "cancelled" {
				return fmt.Errorf("workflow run %s ended with status %s", id, status)
			}
			return nil
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = printResult(cmd, last)
			return fmt.Errorf("timed out waiting for workflow run %s: %w", id, ctx.Err())
		case <-timer.C:
		}
		// Bound the poll GET with the deadline ctx, not just the sleep between
		// polls: cliJSONRequest would derive a fresh, deadline-less context via
		// ctxFor, so a server that accepts the connection but never responds
		// would hang the wait past --timeout-seconds forever. Mirrors
		// waitForPrimitive, which uses cliJSONRequestIntoCtx for the same reason.
		var current any
		if err := cliJSONRequestIntoCtx(ctx, cmd, http.MethodGet, "/v1/workflows/runs/"+url.PathEscape(id), nil, nil, &current); err != nil {
			return err
		}
		var err error
		if last, err = primitiveMap(current); err != nil {
			return err
		}
	}
}

// workflowsRunsWaitCmd is the standalone poller for an already-created run,
// mirroring `experiments runs wait` and the primitive `wait` commands so the
// `create --wait` / standalone-`wait` pair is consistent across every
// pollable resource.
var workflowsRunsWaitCmd = &cobra.Command{
	Use:   "wait <run-id>",
	Short: "Poll until a workflow run reaches a terminal status",
	Long: `Block until a workflow run settles (` + "`completed`" + `/` + "`error`" + `/
` + "`cancelled`" + `) or pauses for human review (` + "`awaiting_review`" + `),
polling on a configurable interval. Defaults: 2-second polls, 10-minute
timeout.

Cleaner than scripting a poll loop around ` + "`runs get`" + ` — the CLI
handles the interval and timeout, prints the final run, and exits non-zero
if the run ends in ` + "`error`" + `/` + "`cancelled`" + ` or the timeout elapses. Pair with
` + "`runs create --wait`" + ` to create and block in a single step.`,
	Example: `  # Wait with defaults (2s polls, 600s timeout)
  retab workflows runs wait run_abc123

  # Faster polls, longer ceiling
  retab workflows runs wait run_abc123 \
    --poll-interval-ms 1000 --timeout-seconds 1800`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		id := args[0]
		current, err := cliJSONRequest(cmd, http.MethodGet, "/v1/workflows/runs/"+url.PathEscape(id), nil, nil)
		if err != nil {
			return err
		}
		initial, err := primitiveMap(current)
		if err != nil {
			return err
		}
		return waitForWorkflowRunByID(cmd, id, initial)
	}),
}

type workflowRunCreateParams struct {
	WorkflowID  string
	Documents   map[string]any
	JSONInputs  map[string]any
	Version     string
	Metadata    map[string]string
	TriggerType string
	ParentRunID string
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
	if len(request.Metadata) > 0 {
		body["metadata"] = request.Metadata
	}
	if request.TriggerType != "" {
		body["trigger_type"] = request.TriggerType
	}
	if request.ParentRunID != "" {
		body["parent_run_id"] = request.ParentRunID
	}
	return body, nil
}

// resolveWorkflowRunDocumentReferences rewrites by-reference document inputs
// into the url-backed MIMEData the canonical POST /v1/workflows/runs route
// accepts. A stored file id resolves to the durable storage.retab.com URL via
// the Files download-link endpoint (the server short-circuits that URL back to
// the stored object, so it remains a reference — no re-download). Inline base64
// content folds into a data: URL. Inputs that already carry a url are left
// untouched.
func resolveWorkflowRunDocumentReferences(cmd *cobra.Command, documents map[string]any) error {
	for blockID, document := range documents {
		resolved, changed, err := resolveWorkflowRunDocumentReference(cmd, document)
		if err != nil {
			return fmt.Errorf("document %q: %w", blockID, err)
		}
		if changed {
			documents[blockID] = resolved
		}
	}
	return nil
}

// workflowRunDocumentFileID reports the stored file id of a by-reference
// document descriptor, and false when the descriptor is not a bare file-id
// reference. A descriptor that already carries a url or inline content is
// url/data-backed and must be sent as-is, so those short-circuit to false —
// only a {id}-shaped map (no url, no content) needs resolving into MIMEData.
func workflowRunDocumentFileID(document any) (string, bool) {
	switch value := document.(type) {
	case map[string]any:
		if s, _ := value["url"].(string); s != "" {
			return "", false
		}
		if s, _ := value["content"].(string); s != "" {
			return "", false
		}
		id, _ := value["id"].(string)
		return id, id != ""
	case map[string]string:
		if value["url"] != "" || value["content"] != "" {
			return "", false
		}
		return value["id"], value["id"] != ""
	default:
		return "", false
	}
}

// workflowRunDocumentFilename returns the caller-supplied filename on a document
// descriptor, or "" when absent. It is used to preserve the original filename
// after a file id is resolved into a fresh signed URL (which carries no name).
func workflowRunDocumentFilename(document any) string {
	switch value := document.(type) {
	case map[string]any:
		name, _ := value["filename"].(string)
		return name
	case map[string]string:
		return value["filename"]
	default:
		return ""
	}
}

func resolveWorkflowRunDocumentReference(cmd *cobra.Command, document any) (any, bool, error) {
	id, content, mimeType, filename, hasURL := workflowRunDocumentReferenceParts(document)
	if hasURL {
		return document, false, nil
	}
	switch {
	case id != "":
		mime, err := resolveFileIDToMIMEData(cmd, id)
		if err != nil {
			return nil, false, err
		}
		return mime, true, nil
	case content != "" && mimeType != "":
		if filename == "" {
			filename = "document"
		}
		return retab.MIMEData{Filename: filename, URL: "data:" + mimeType + ";base64," + content}, true, nil
	default:
		return document, false, nil
	}
}

// workflowRunDocumentReferenceParts extracts the by-reference fields from any
// supported document input shape (MIMEData, FileRef, or a raw map descriptor).
func workflowRunDocumentReferenceParts(document any) (id, content, mimeType, filename string, hasURL bool) {
	switch value := document.(type) {
	case retab.MIMEData:
		return "", value.Content, value.MIMEType, value.Filename, value.URL != ""
	case *retab.MIMEData:
		if value == nil {
			return "", "", "", "", false
		}
		return "", value.Content, value.MIMEType, value.Filename, value.URL != ""
	case retab.FileRef:
		return value.ID, value.Content, value.MIMEType, value.Filename, false
	case *retab.FileRef:
		if value == nil {
			return "", "", "", "", false
		}
		return value.ID, value.Content, value.MIMEType, value.Filename, false
	case map[string]any:
		return stringFromAnyMap(value, "id"), stringFromAnyMap(value, "content"), stringFromAnyMap(value, "mime_type"), stringFromAnyMap(value, "filename"), stringFromAnyMap(value, "url") != ""
	case map[string]string:
		return value["id"], value["content"], value["mime_type"], value["filename"], value["url"] != ""
	default:
		return "", "", "", "", false
	}
}

func stringFromAnyMap(m map[string]any, key string) string {
	if raw, ok := m[key]; ok {
		if s, ok := raw.(string); ok {
			return s
		}
	}
	return ""
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
		// The workflow-runs route accepts MIMEData only; file ids are resolved
		// to a download URL before reaching here (see resolveFileIDToMIMEData*),
		// so "id" is intentionally not copied into the request body.
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
	if len(documents) == 0 {
		return documents, nil
	}
	// Always list the workflow's blocks before resolving keys. An earlier
	// optimization skipped this when every key already looked like a canonical
	// "block_..." id, but that is unsound now that a "block_" key can be a
	// declarative_source_block_id alias (e.g. "block_document" -> the generated
	// start_document block): the client cannot tell an alias from a real id
	// without the block list, so it must always resolve.
	blocks, err := listAllWorkflowBlocks(ctx, client, workflowID)
	if err != nil {
		return nil, fmt.Errorf("resolve workflow run document aliases: %w", err)
	}
	blockIDs := map[string]bool{}
	var startDocumentBlocks []retab.WorkflowBlock
	for _, block := range blocks {
		blockIDs[block.ID] = true
		if isStartDocumentBlock(block) {
			startDocumentBlocks = append(startDocumentBlocks, block)
		}
	}
	resolved := make(map[string]any, len(documents))
	for key, value := range documents {
		resolvedKey, err := resolveWorkflowRunDocumentKey(key, blockIDs, startDocumentBlocks)
		if err != nil {
			return nil, err
		}
		if _, exists := resolved[resolvedKey]; exists {
			return nil, fmt.Errorf("document inputs supply more than one input for start_document block %q", resolvedKey)
		}
		resolved[resolvedKey] = value
	}
	return resolved, nil
}

func resolveWorkflowRunDocumentKey(key string, blockIDs map[string]bool, startDocumentBlocks []retab.WorkflowBlock) (string, error) {
	if blockIDs[key] {
		return key, nil
	}
	var matches []retab.WorkflowBlock
	for _, block := range startDocumentBlocks {
		if block.DeclarativePath != nil && *block.DeclarativePath == key {
			matches = append(matches, block)
			continue
		}
		if block.DeclarativeSourceBlockID != nil && *block.DeclarativeSourceBlockID == key {
			matches = append(matches, block)
		}
	}
	if len(matches) == 0 && key == "start" {
		matches = startDocumentBlocks
	}
	switch len(matches) {
	case 0:
		return key, nil
	case 1:
		return matches[0].ID, nil
	default:
		return "", fmt.Errorf("document input key %q is ambiguous: workflow has %d matching start_document blocks; use the concrete block id", key, len(matches))
	}
}

func resolveWorkflowRunJSONInputAliases(
	ctx context.Context,
	client *retab.Client,
	workflowID string,
	inputs map[string]any,
) (map[string]any, error) {
	if len(inputs) == 0 {
		return inputs, nil
	}
	blocks, err := listAllWorkflowBlocks(ctx, client, workflowID)
	if err != nil {
		return nil, fmt.Errorf("resolve --json-inputs-file aliases: %w", err)
	}
	blockIDs := map[string]bool{}
	var startJSONBlocks []retab.WorkflowBlock
	for _, block := range blocks {
		blockIDs[block.ID] = true
		if block.Type == "start_json" {
			startJSONBlocks = append(startJSONBlocks, block)
		}
	}
	resolved := make(map[string]any, len(inputs))
	for key, value := range inputs {
		resolvedKey, err := resolveWorkflowRunJSONInputKey(key, blockIDs, startJSONBlocks)
		if err != nil {
			return nil, err
		}
		if _, exists := resolved[resolvedKey]; exists {
			return nil, fmt.Errorf("--json-inputs-file supplies more than one input for start_json block %q", resolvedKey)
		}
		resolved[resolvedKey] = value
	}
	return resolved, nil
}

func resolveWorkflowRunJSONInputKey(key string, blockIDs map[string]bool, startJSONBlocks []retab.WorkflowBlock) (string, error) {
	if blockIDs[key] {
		return key, nil
	}
	var matches []retab.WorkflowBlock
	for _, block := range startJSONBlocks {
		if block.DeclarativePath != nil && *block.DeclarativePath == key {
			matches = append(matches, block)
			continue
		}
		if block.DeclarativeSourceBlockID != nil && *block.DeclarativeSourceBlockID == key {
			matches = append(matches, block)
		}
	}
	if len(matches) == 0 && key == "start" {
		matches = startJSONBlocks
	}
	switch len(matches) {
	case 0:
		return key, nil
	case 1:
		return matches[0].ID, nil
	default:
		return "", fmt.Errorf("--json-inputs-file key %q is ambiguous: workflow has %d matching start_json blocks; use the concrete block id", key, len(matches))
	}
}

var workflowsRunsGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get a workflow run",
	Long: `Fetch a run's metadata: status, trigger type, timestamps,
duration, cost, error info. The run payload does not include per-block
steps by default; pass ` + "`--steps`" + ` to fetch and embed them under
` + "`steps`" + `, or use ` + "`workflows steps list`" + ` for the full
per-block detail and outputs.

Run ids are globally unique, so read and poll commands take only the
` + "`<run-id>`" + ` — the workflow id is never in the path. Only ` + "`runs create`" + `,
which addresses a parent collection, takes a workflow id. The same holds for
` + "`workflows steps list <run-id>`" + ` and the tests/experiments run getters.`,
	Example: `  # Inspect a run
  retab workflows runs get run_xyz789

  # Poll until done
  while [ "$(retab workflows runs get run_xyz789 | jq -r '.lifecycle.status')" = "running" ]; do
    sleep 2
  done`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		runID := strings.TrimSpace(args[0])
		if runID == "" {
			return fmt.Errorf("expected the run id")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Runs.Get(ctx, runID)
		if err != nil {
			return err
		}
		// The run GET endpoint returns lifecycle/timing/inputs but NOT the
		// per-block steps. With --steps, fetch them and embed under "steps"
		// so callers get the run + its execution records in one command
		// instead of a second `workflows steps list` round trip.
		if includeSteps, _ := cmd.Flags().GetBool("steps"); includeSteps {
			page, err := client.Workflows.Steps.List(ctx, &retab.WorkflowStepsListParams{RunID: ptr(runID)})
			if err != nil {
				return err
			}
			// Walk every page: a run with more steps than the server's default
			// page size (easy with loop/for_each blocks) would otherwise embed
			// a silently truncated subset. Same rationale as
			// listAllWorkflowBlocks/listAllWorkflowEdges.
			steps := []retab.WorkflowRunStep{}
			if err := page.AutoPaging(ctx, func(s retab.WorkflowRunStep) error {
				steps = append(steps, s)
				return nil
			}); err != nil {
				return err
			}
			merged, err := primitiveMap(result)
			if err != nil {
				return err
			}
			merged["steps"] = steps
			return printResult(cmd, merged)
		}
		return printResult(cmd, result)
	}),
}

var workflowsRunsListCmd = &cobra.Command{
	Use:   "list [workflow-id]",
	Short: "List workflow runs",
	Long: `List workflow runs. Without a workflow id the result spans the
whole workspace; with one (passed positionally OR via ` + "`--workflow-id`" + `)
the result is scoped to that workflow. Passing both forms is accepted
when they reference the same workflow id; an error is only raised when
they disagree, so a real typo isn't silently masked. Other filters
available: status, trigger type, date range, cost, and duration. Page by
run id (` + "`--after`" + ` / ` + "`--before`" + ` / ` + "`--limit`" + `).`,
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
		// Honour <workflow-id> positionally OR via --workflow-id (the two
		// forms are co-equal across every workflow-scoped list). Scope is
		// optional here: with no id the listing spans the whole workspace.
		effectiveID, err := resolveWorkflowScope(cmd, args, false)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.WorkflowRunsListParams{}
		if effectiveID != "" {
			params.WorkflowID = ptr(effectiveID)
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			status := retab.WorkflowRunsStatus(v)
			params.Status = &status
		}
		if v, _ := cmd.Flags().GetString("exclude-status"); v != "" {
			excludeStatus := retab.WorkflowRunsExcludeStatus(v)
			params.ExcludeStatus = &excludeStatus
		}
		if v, _ := cmd.Flags().GetString("trigger-type"); v != "" {
			triggerType := retab.WorkflowRunsTriggerType(v)
			params.TriggerType = &triggerType
		}
		fromDate, _ := cmd.Flags().GetString("from-date")
		if fromDate != "" {
			params.FromDate = ptr(fromDate)
		}
		toDate, _ := cmd.Flags().GetString("to-date")
		if toDate != "" {
			params.ToDate = ptr(toDate)
		}
		if err := validateDateRange("from-date", "to-date", fromDate, toDate); err != nil {
			return err
		}
		if v, _ := cmd.Flags().GetString("search"); v != "" {
			params.Search = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("sort-by"); v != "" {
			params.SortBy = ptr(v)
		}
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		before, _ := cmd.Flags().GetString("before")
		after, _ := cmd.Flags().GetString("after")
		if before != "" {
			params.Before = ptr(before)
		}
		if after != "" {
			params.After = ptr(after)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			params.Limit = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("order"); v != "" {
			params.Order = ptr(v)
		}
		// `--min-cost`, `--max-cost`, `--min-duration`, `--max-duration` were
		// available on the legacy SDK list params but are no longer projected
		// onto the regenerated `WorkflowRunsListParams`. Keep the flag wiring
		// so existing scripts don't fail, but only honour the duration filters
		// against the typed shape — the cost filters drop silently until the
		// SDK exposes them again.
		var minDuration, maxDuration *int
		if cmd.Flags().Changed("min-duration") {
			v, _ := cmd.Flags().GetInt("min-duration")
			minDuration = &v
			params.MinDurationMs = &v
		}
		if cmd.Flags().Changed("max-duration") {
			v, _ := cmd.Flags().GetInt("max-duration")
			maxDuration = &v
			params.MaxDurationMs = &v
		}
		if minDuration != nil && maxDuration != nil && *minDuration > *maxDuration {
			return fmt.Errorf("--min-duration (%d) cannot be greater than --max-duration (%d)", *minDuration, *maxDuration)
		}
		var minCost, maxCost *float64
		if cmd.Flags().Changed("min-cost") {
			v, _ := cmd.Flags().GetFloat64("min-cost")
			minCost = &v
		}
		if cmd.Flags().Changed("max-cost") {
			v, _ := cmd.Flags().GetFloat64("max-cost")
			maxCost = &v
		}
		if minCost != nil && maxCost != nil && *minCost > *maxCost {
			return fmt.Errorf("--min-cost (%g) cannot be greater than --max-cost (%g)", *minCost, *maxCost)
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
	// `runs list` is always scoped to one workflow id, so a per-run workflow
	// name column would be identical on every row — and the run response
	// doesn't project workflow.name_at_run_time anyway (it flattens to
	// workflow_id/workflow_version_id), so that column was structurally empty.
	// Surface the trigger type instead: it's populated and varies per run.
	{Header: "TRIGGER", Extract: func(row any) string { return workflowRunCell(row, "trigger.type") }},
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
` + "`workflows artifacts`" + `).

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal. Mirrors the contract of ` + "`workflows delete`" + `,
` + "`workflows blocks delete`" + `, etc.`,
	Example: `  # Delete a run (interactive, asks to confirm)
  retab workflows runs delete run_xyz789

  # Skip the prompt in scripts
  retab workflows runs delete run_xyz789 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "run", args[0]); err != nil {
			return err
		}
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
		cancelParams := &retab.WorkflowRunsCancelParams{}
		if commandID != "" {
			cancelParams.Body = retab.CancelWorkflowRequest{CommandID: &commandID}
		}
		result, err := client.Workflows.Runs.Cancel(ctx, args[0], cancelParams)
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
		//
		// But a cancel can lose a race: the run reaches a terminal state
		// (typically ``completed``) between the engine accepting the
		// cancel and serializing the response, which still reports
		// cancellation_status="cancellation_failed". In that case the
		// generic "poll until terminal" advice is wrong — the embedded run
		// is already terminal — so key the note on the run's actual
		// lifecycle status, not on cancellation_status alone.
		if result != nil && result.CancellationStatus != nil && *result.CancellationStatus != retab.CancelWorkflowResponseCancellationStatusCancelled {
			runStatus := result.Run.Lifecycle.Status()
			// Only completed/error/cancelled are settled here. awaiting_review
			// is a pause that is still cancellable, so it does not defuse the
			// "poll until terminal" advice.
			if runStatus == "completed" || runStatus == "error" || runStatus == "cancelled" {
				fmt.Fprintf(
					os.Stderr,
					"note: cancellation_status=%q — the run reached a terminal state (lifecycle.status=%q) before the cancel took effect, so it was not cancelled.\n",
					string(*result.CancellationStatus),
					runStatus,
				)
			} else {
				fmt.Fprintf(
					os.Stderr,
					"note: cancellation_status=%q — the cancel request was accepted but the run has not yet reached a terminal state. Poll `retab workflows runs get %s` until lifecycle.status is one of cancelled / completed / error.\n",
					string(*result.CancellationStatus),
					args[0],
				)
			}
		}
		return printResult(cmd, result)
	}),
}

var workflowsRunsRestartCmd = &cobra.Command{
	Use:   "restart <run-id>",
	Short: "Restart a workflow run",
	Long: `Re-execute a previous run, reusing the original inputs.
By default the restarted run uses the latest published workflow config.
Use ` + "`--config-source draft`" + ` after tweaking draft block config.`,
	Example: `  # Restart a previous run with the same inputs
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
		if sourceRun.WorkflowID == "" {
			return fmt.Errorf("source run %s does not include workflow_id", args[0])
		}
		version := "production"
		if configSource == "draft" {
			version = "draft"
		}
		params := workflowRunCreateParams{
			WorkflowID:  sourceRun.WorkflowID,
			Version:     version,
			TriggerType: "restart",
			ParentRunID: args[0],
		}
		// Reuse the source run's stored inputs verbatim. Document inputs are
		// stored as FileRefs ({id, filename, mime_type}); Runs.Create resolves
		// each one into a durable storage.retab.com MIMEData URL (via the Files
		// download-link endpoint) before posting to the canonical create route,
		// which only accepts url-backed MIMEData. This keeps restart on the
		// public route — no legacy /v1/workflows/{id}/run fallback.
		if sourceRun.Inputs != nil {
			if sourceRun.Inputs.Documents != nil {
				documents := map[string]interface{}{}
				for key, value := range sourceRun.Inputs.Documents {
					documents[key] = value
				}
				params.Documents = documents
			}
			if sourceRun.Inputs.JSONData != nil {
				params.JSONInputs = sourceRun.Inputs.JSONData
			}
		}
		if params.Documents != nil {
			if err := resolveWorkflowRunDocumentReferences(cmd, params.Documents); err != nil {
				return err
			}
		}
		body, err := workflowRunCreateRequestBody(params)
		if err != nil {
			return err
		}
		result, err := cliJSONRequest(cmd, http.MethodPost, "/v1/workflows/runs", nil, body)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsRunsExportCmd = &cobra.Command{
	Use:   "export <workflow-id> [flags]",
	Short: "Export workflow runs as CSV",
	Long: `Bulk-export the outputs of many runs of a workflow as CSV,
focused on one block's output. Pass the workflow id as the first
positional argument (matching the rest of the ` + "`workflows runs`" + `
group). Filter by status, date range, trigger type, or an explicit set
of run ids. Use ` + "`--preferred-column`" + ` to pin column order.

CSV shape: the default delimiter is ` + "`;`" + ` (convenient for Excel
in EU locales where ` + "`,`" + ` is the decimal separator) but hostile
to pandas / RFC-4180 consumers. Pass ` + "`--delimiter ,`" + ` for the
portable shape. ` + "`--quote`" + ` and ` + "`--line-delimiter`" + ` are
also configurable.`,
	Example: `  # Export every successful run of a block to CSV
  retab workflows runs export wf_abc123 \
    --block-id block_extract_1 \
    --status completed \
    --from-date 2026-05-01

  # Pandas-friendly CSV (RFC-4180 delimiter)
  retab workflows runs export wf_abc123 \
    --block-id block_extract_1 --delimiter ,

  # Export a specific set of runs with custom columns
  retab workflows runs export wf_abc123 \
    --block-id block_extract_1 \
    --run-id run_xyz789 --run-id run_aaa000 \
    --preferred-column invoice_id --preferred-column total`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateWorkflowRunsExportFilters(cmd); err != nil {
			return err
		}
		workflowID, err := resolveWorkflowIDArg(cmd, args)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.WorkflowRunsExportParams{}
		req.WorkflowID = workflowID
		req.BlockID, _ = cmd.Flags().GetString("block-id")
		if v, _ := cmd.Flags().GetString("export-source"); v != "" {
			source := retab.WorkflowExportPayloadRequestExportSource(v)
			req.ExportSource = &source
		}
		selectedRunIDs, err := nonBlankStringArrayFlag(cmd, "run-id")
		if err != nil {
			return err
		}
		req.SelectedRunIDs = selectedRunIDs
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			status := retab.WorkflowExportPayloadRequestStatus(v)
			req.Status = &status
		}
		if v, _ := cmd.Flags().GetString("exclude-status"); v != "" {
			excludeStatus := retab.WorkflowExportPayloadRequestExcludeStatus(v)
			req.ExcludeStatus = &excludeStatus
		}
		fromDate, _ := cmd.Flags().GetString("from-date")
		toDate, _ := cmd.Flags().GetString("to-date")
		if fromDate != "" {
			req.FromDate = ptr(fromDate)
		}
		if toDate != "" {
			req.ToDate = ptr(toDate)
		}
		if err := validateDateRange("from-date", "to-date", fromDate, toDate); err != nil {
			return err
		}
		if v, _ := cmd.Flags().GetString("trigger-type"); v != "" {
			triggerType := retab.WorkflowExportPayloadRequestTriggerType(v)
			req.TriggerType = &triggerType
		}
		preferredColumns, err := nonBlankStringArrayFlag(cmd, "preferred-column")
		if err != nil {
			return err
		}
		req.PreferredColumns = preferredColumns
		// CSV shape flags. Validated client-side so the user sees a clean
		// "must be a single character" error instead of a 400 from the
		// server with the same message.
		if cmd.Flags().Changed("delimiter") {
			d, _ := cmd.Flags().GetString("delimiter")
			if len([]rune(d)) != 1 {
				return fmt.Errorf("--delimiter must be a single character")
			}
			req.Delimiter = &d
		}
		if cmd.Flags().Changed("line-delimiter") {
			ld, _ := cmd.Flags().GetString("line-delimiter")
			if ld == "" {
				return fmt.Errorf("--line-delimiter must not be empty")
			}
			req.LineDelimiter = &ld
		}
		if cmd.Flags().Changed("quote") {
			q, _ := cmd.Flags().GetString("quote")
			if len([]rune(q)) != 1 {
				return fmt.Errorf("--quote must be a single character")
			}
			req.Quote = &q
		}
		result, err := client.Workflows.Runs.Export(ctx, &req)
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
		if shouldDumpRawExportCSV(raw, outFormat, term.IsTerminal(int(os.Stdout.Fd()))) {
			if result != nil {
				_, err := os.Stdout.WriteString(result.CsvData)
				if err == nil && !strings.HasSuffix(result.CsvData, "\n") {
					_, err = os.Stdout.WriteString("\n")
				}
				return err
			}
		}
		return printResult(cmd, result)
	}),
}

// shouldDumpRawExportCSV reports whether `runs export` should emit the raw CSV
// bytes instead of the JSON envelope. Raw output is chosen for --raw, an
// explicit table/csv format, or the auto-detect default on a TTY. Both "" and
// "auto" mean auto-detect (--output accepts "auto" as an explicit synonym for
// the default), so both must consult the TTY — checking only "" let an explicit
// `--output auto` fall through to the JSON envelope.
func shouldDumpRawExportCSV(raw bool, outFormat string, isTTY bool) bool {
	switch outFormat {
	case "table", string(OutputCSV):
		return true
	}
	if raw {
		return true
	}
	return (outFormat == "" || outFormat == "auto") && isTTY
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
	workflowsRunsCreateCmd.Flags().String("version", "", "workflow version (defaults to production)")
	workflowsRunsCreateCmd.Flags().Bool("draft", false, "run the editable draft without publishing (alias for --version draft)")
	workflowsRunsCreateCmd.Flags().String("documents-file", "", "JSON object {block-id: document} (or - for stdin)")
	workflowsRunsCreateCmd.Flags().StringArray("document", nil, "document file as block-id=path (repeatable)")
	// `--document-file` collided with the `--document-file <json-path>` flag
	// every other primitive command exposes via addDocumentFlags. Kept here
	// as a hidden deprecated alias for one release cycle; parseDocumentArgs
	// emits a single stderr warning when it's used.
	workflowsRunsCreateCmd.Flags().StringArray("document-file", nil, "deprecated alias for --document")
	_ = workflowsRunsCreateCmd.Flags().MarkHidden("document-file")
	workflowsRunsCreateCmd.Flags().StringArray("document-url", nil, "document url as block-id=url (repeatable)")
	workflowsRunsCreateCmd.Flags().StringArray("document-id", nil, "previously-uploaded file as block-id=file-id (repeatable)")
	workflowsRunsCreateCmd.Flags().String("json-inputs-file", "", "JSON inputs object (or - for stdin)")
	workflowsRunsCreateCmd.Flags().StringArray("metadata", nil, "user-defined metadata as key=value (repeatable)")
	// --wait blocks until the run settles (completed/error/cancelled) or
	// pauses for review (awaiting_review); --poll-interval-ms / --timeout-seconds
	// tune the poll loop, matching `extractions create` and `experiments runs create`.
	workflowsRunsCreateCmd.Flags().Bool("wait", false, "block until the run reaches a terminal status (completed/error/cancelled/awaiting_review), then print the final run")
	addPrimitiveWaitTuningFlags(workflowsRunsCreateCmd, true)

	workflowsRunsGetCmd.Flags().Bool("steps", false, "also fetch the run's per-block step records and embed them under \"steps\"")

	workflowsRunsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	workflowsRunsListCmd.Flags().String("status", "", "filter by status")
	workflowsRunsListCmd.Flags().String("exclude-status", "", "exclude status")
	workflowsRunsListCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsRunsListCmd.Flags().Var(&rfc3339FlagValue{}, "from-date", "filter from this date (YYYY-MM-DD or RFC3339)")
	workflowsRunsListCmd.Flags().Var(&rfc3339FlagValue{}, "to-date", "filter to this date (YYYY-MM-DD or RFC3339)")
	workflowsRunsListCmd.Flags().String("search", "", "search by run id (partial match) — does NOT match document filenames")
	workflowsRunsListCmd.Flags().Var(newEnumStringFlagValue("--sort-by", "timing.created_at", "timing.started_at"), "sort-by", "sort field: timing.created_at | timing.started_at")
	workflowsRunsListCmd.Flags().String("before", "", "run id: return items before this id (mutually exclusive with --after)")
	workflowsRunsListCmd.Flags().String("after", "", "run id: return items after this id (mutually exclusive with --before)")
	workflowsRunsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	workflowsRunsListCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
	workflowsRunsListCmd.Flags().Var(&nonNegativeFloatFlagValue{}, "min-cost", "min cost")
	workflowsRunsListCmd.Flags().Var(&nonNegativeFloatFlagValue{}, "max-cost", "max cost")
	workflowsRunsListCmd.Flags().Var(&nonNegativeIntFlagValue{}, "min-duration", "min duration (ms)")
	workflowsRunsListCmd.Flags().Var(&nonNegativeIntFlagValue{}, "max-duration", "max duration (ms)")

	workflowsRunsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	workflowsRunsCancelCmd.Flags().String("command-id", "", "idempotency command id")
	workflowsRunsRestartCmd.Flags().String("config-source", "published", "published | draft")

	workflowsRunsExportCmd.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	workflowsRunsExportCmd.Flags().String("block-id", "", "block id (required)")
	workflowsRunsExportCmd.Flags().String("export-source", "", "export source (default outputs)")
	workflowsRunsExportCmd.Flags().StringArray("run-id", nil, "filter to selected run ids (repeatable)")
	workflowsRunsExportCmd.Flags().String("status", "", "status filter")
	workflowsRunsExportCmd.Flags().String("exclude-status", "", "exclude status")
	workflowsRunsExportCmd.Flags().Var(&rfc3339FlagValue{}, "from-date", "filter from this date (YYYY-MM-DD or RFC3339)")
	workflowsRunsExportCmd.Flags().Var(&rfc3339FlagValue{}, "to-date", "filter to this date (YYYY-MM-DD or RFC3339)")
	workflowsRunsExportCmd.Flags().String("trigger-type", "", "filter by trigger type")
	workflowsRunsExportCmd.Flags().StringArray("preferred-column", nil, "preferred CSV column (repeatable)")
	workflowsRunsExportCmd.Flags().Bool("raw", false, "write raw CSV to stdout (default for TTY); JSON envelope is used for non-TTY unless --raw or --output table")
	// CSV shape flags — defaults come from the server side. The default
	// delimiter is ``;`` (Excel-friendly in EU locales) which is hostile
	// to pandas / RFC-4180 consumers; pass ``--delimiter ,`` for the
	// portable shape.
	workflowsRunsExportCmd.Flags().String("delimiter", "", "CSV field delimiter (single character; server default ';')")
	workflowsRunsExportCmd.Flags().String("line-delimiter", "", "CSV line terminator (server default '\\n')")
	workflowsRunsExportCmd.Flags().String("quote", "", "CSV quote character (single character; server default '\"')")
	// Keep the flag hidden but DO NOT use MarkDeprecated — cobra's auto warning
	// duplicates the more-specific message emitted by resolveWorkflowIDArg.
	_ = workflowsRunsExportCmd.Flags().MarkHidden("workflow-id")
	_ = workflowsRunsExportCmd.MarkFlagRequired("block-id")

	// Standalone poller for an already-running run; tuning flags match the
	// `runs create --wait` knobs and the primitive/experiment wait commands.
	addPrimitiveWaitTuningFlags(workflowsRunsWaitCmd, false)

	workflowsRunsCmd.AddCommand(workflowsRunsCreateCmd, workflowsRunsGetCmd, workflowsRunsListCmd, workflowsRunsDeleteCmd, workflowsRunsCancelCmd, workflowsRunsRestartCmd, workflowsRunsExportCmd, workflowsRunsWaitCmd)
	workflowsCmd.AddCommand(workflowsRunsCmd)
}
