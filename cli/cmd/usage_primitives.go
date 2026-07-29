package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
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
	EnvironmentID        string                     `json:"environment_id,omitempty"`
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
	TriggeredBy          *usagePrimitiveTriggeredBy `json:"triggered_by,omitempty"`
}

// usagePrimitiveTriggeredBy is the triggering credential's provenance: the auth
// method plus the credential's non-secret identifiers (api key id / access
// token id / user id, and the key's display prefix + name). Null for rows
// recorded before provenance capture.
type usagePrimitiveTriggeredBy struct {
	AuthMethod    string `json:"auth_method,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	UserEmail     string `json:"user_email,omitempty"`
	APIKeyID      string `json:"api_key_id,omitempty"`
	AccessTokenID string `json:"access_token_id,omitempty"`
	KeyPrefix     string `json:"key_prefix,omitempty"`
	KeyName       string `json:"key_name,omitempty"`
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
cap the page size with ` + "`--limit`" + ` (1-10000).

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
	addOptionalUsageQuery(cmd, query, "api-key-id", "api_key_id")
	addOptionalUsageQuery(cmd, query, "access-token-id", "access_token_id")
	addOptionalUsageQuery(cmd, query, "user-id", "user_id")
	addOptionalUsageQuery(cmd, query, "run-id", "run_id")
	blockFilter, err := resolveUsageBlockIDFilter(cmd)
	if err != nil {
		return err
	}
	if blockFilter.blockID != "" {
		query.Set("block_id", blockFilter.blockID)
	}
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
	warnEmptyUsageBlockFilter(cmd, blockFilter, len(result.Data))
	return printUsagePrimitiveListResult(cmd, result)
}

// usageBlockIDFilter is the outcome of mapping the user's --block-id onto the
// block id the usage ledger actually records.
type usageBlockIDFilter struct {
	// requested is the raw --block-id value, empty when the flag is unset.
	requested string
	// blockID is what gets sent as the block_id query argument: the resolved
	// runtime id when an alias was recognized, else `requested` unchanged.
	blockID string
	// loopParentID is set when the resolved block sits inside a for_each /
	// while_loop container. Such a block never appears in the ledger under its
	// own id — see warnEmptyUsageBlockFilter.
	loopParentID string
	// known reports whether the workflow's block list contained the id at all.
	known bool
}

// resolveUsageBlockIDFilter maps a --block-id the user can legitimately see onto
// the runtime block id the usage ledger records.
//
// A workflow block is reachable by three different spellings, and only the last
// one is what `usage primitives` stores:
//
//   - the declarative id authored in the spec ("block_ex"), which is what
//     `workflows spec get` prints and what `runs create` accepts as an alias;
//   - the declarative path ("ex", "loop.item_extract"), the other alias
//     `runs create` accepts;
//   - the generated runtime id ("block_lte_nNvJ6RGq86aH3LEXD"), which only
//     `workflows blocks list` and the usage export show.
//
// Filtering by either alias used to return an empty 200, and on a usage export
// an empty page reads as "this block cost nothing" rather than "that is not the
// id I index by" — the same trap the --operation / --status enums were tightened
// for. Resolution needs the workflow's block list, so it only runs when
// --workflow-id is also set; without it the value is forwarded untouched.
//
// An unrecognized value is also forwarded untouched rather than rejected: the
// ledger legitimately carries ids that are in no block list, most importantly
// the per-iteration ids a for_each expands to
// ("<loop>_iter_<n>_<child>") and its "<loop>_sentinel_start" row.
func resolveUsageBlockIDFilter(cmd *cobra.Command) (usageBlockIDFilter, error) {
	blockID, _ := cmd.Flags().GetString("block-id")
	blockID = strings.TrimSpace(blockID)
	filter := usageBlockIDFilter{requested: blockID, blockID: blockID}
	workflowID, _ := cmd.Flags().GetString("workflow-id")
	workflowID = strings.TrimSpace(workflowID)
	if blockID == "" || workflowID == "" {
		return filter, nil
	}
	// Resolution is a convenience, never a gate. The usage ledger deliberately
	// outlives the workflows it bills for — rows survive workflow deletion — so a
	// failed, forbidden, or slow block lookup must not take down a billing read
	// that would otherwise have returned data. Every failure degrades to the raw
	// value.
	client, err := newClient(cmd)
	if err != nil {
		return filter, nil
	}
	parent := cmd.Context()
	if parent == nil {
		parent = context.Background()
	}
	// Hard deadline, because listAllWorkflowBlocks walks every page: a blocks
	// endpoint that keeps handing back a cursor would otherwise spin forever
	// inside a read the user only asked to filter.
	ctx, cancel := context.WithTimeout(parent, usageBlockLookupTimeout)
	defer cancel()
	blocks, err := listAllWorkflowBlocks(ctx, client, workflowID)
	if err != nil {
		return filter, nil
	}
	return matchUsageBlockIDFilter(blockID, blocks), nil
}

// usageBlockLookupTimeout bounds the --block-id alias lookup. Generous enough for
// a paginated block list on a slow link, short enough that a wedged or
// misbehaving endpoint costs the user a few seconds rather than the command. A
// var only so the endless-pagination regression test can shrink it.
var usageBlockLookupTimeout = 15 * time.Second

// matchUsageBlockIDFilter is resolveUsageBlockIDFilter's pure half: the
// runtime-id / declarative-id / declarative-path lookup over an already-fetched
// block list.
func matchUsageBlockIDFilter(blockID string, blocks []retab.WorkflowBlock) usageBlockIDFilter {
	filter := usageBlockIDFilter{requested: blockID, blockID: blockID}
	var matched *retab.WorkflowBlock
	for i := range blocks {
		if blocks[i].ID == blockID {
			matched = &blocks[i]
			break
		}
	}
	if matched == nil {
		// Exactly one alias must match. Declarative ids and paths are unique
		// within a workflow, so an ambiguous hit means the caller is better
		// served by the raw value than by an arbitrary pick.
		var aliasHits []*retab.WorkflowBlock
		for i := range blocks {
			decl := blocks[i].DeclarativeSourceBlockID
			path := blocks[i].DeclarativePath
			if (decl != nil && *decl == blockID) || (path != nil && *path == blockID) {
				aliasHits = append(aliasHits, &blocks[i])
			}
		}
		if len(aliasHits) == 1 {
			matched = aliasHits[0]
			filter.blockID = matched.ID
		}
	}
	if matched != nil {
		filter.known = true
		if matched.ParentID != nil {
			filter.loopParentID = strings.TrimSpace(*matched.ParentID)
		}
	}
	return filter
}

// warnEmptyUsageBlockFilter explains a zero-row page for a block id that really
// does exist. A block nested in a for_each / while_loop container is never
// recorded under its own id: the engine expands it per iteration and the ledger
// stores "<container>_iter_<n>_<block>". Filtering by the design-time id is the
// obvious thing to try and silently reports no spend, so say why instead.
func warnEmptyUsageBlockFilter(cmd *cobra.Command, filter usageBlockIDFilter, rows int) {
	if rows > 0 || filter.requested == "" || filter.loopParentID == "" {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(),
		"warning: block %s runs inside container %s, so usage is recorded once per iteration "+
			"as %s_iter_<n>_%s, not under the block id itself. "+
			"Run `retab usage blocks --workflow-id %s` to list the ledger's block ids.\n",
		filter.requested, filter.loopParentID, filter.loopParentID, filter.blockID,
		mustFlagString(cmd, "workflow-id"))
}

// mustFlagString reads a string flag, returning "" when it is absent.
func mustFlagString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return strings.TrimSpace(v)
}

var usagePrimitiveColumns = []TableColumn{
	{Header: "EXECUTION_ID", Extract: func(row any) string { return usagePrimitiveCell(row, "primitive_execution_id") }},
	{Header: "OPERATION", Extract: func(row any) string { return usagePrimitiveCell(row, "operation") }},
	{Header: "MODEL", Extract: func(row any) string { return usagePrimitiveCell(row, "model") }},
	{Header: "WORKFLOW", Extract: func(row any) string { return usagePrimitiveCell(row, "workflow_id") }},
	{Header: "BLOCK", Extract: func(row any) string { return usagePrimitiveCell(row, "block_id") }},
	{Header: "PROJECT", Extract: func(row any) string { return usagePrimitiveCell(row, "project_id") }},
	{Header: "STATUS", Extract: func(row any) string { return usagePrimitiveCell(row, "status") }},
	{Header: "TRIGGERED_BY", Extract: usagePrimitiveTriggeredByCell},
	{Header: "FILENAME", Extract: usagePrimitiveFilenameCell},
	{Header: "CREATED_AT", Extract: func(row any) string { return normalizeTimestampCell(usagePrimitiveCell(row, "created_at")) }},
	{Header: "COMPLETED_AT", Extract: func(row any) string { return normalizeTimestampCell(usagePrimitiveCell(row, "completed_at")) }},
	{Header: "DURATION_MS", Extract: func(row any) string { return usagePrimitiveCell(row, "duration_ms") }},
	{Header: "PAGES", Extract: func(row any) string { return usagePrimitiveCell(row, "page_count") }},
	{Header: "CREDITS", Extract: func(row any) string { return usagePrimitiveCell(row, "credits") }},
}

// usagePrimitiveTriggeredByCell renders the triggering credential as one short
// cell: the key's display name (or prefix) for api-key/token callers, the acting
// person's email (else their user id) for session callers, falling back to the
// bare auth method. The full provenance object stays available in --output json.
func usagePrimitiveTriggeredByCell(row any) string {
	rec, ok := row.(usagePrimitiveRecord)
	if !ok || rec.TriggeredBy == nil {
		return ""
	}
	t := rec.TriggeredBy
	// Prefer the most human-readable handle available: a key's display name for
	// application callers, the acting person's email when the API resolved one,
	// then the raw ids.
	label := t.KeyName
	if label == "" {
		label = t.KeyPrefix
	}
	if label == "" {
		label = t.UserEmail
	}
	if label == "" {
		label = t.UserID
	}
	if label == "" {
		return t.AuthMethod
	}
	if t.AuthMethod == "" {
		return label
	}
	return fmt.Sprintf("%s:%s", t.AuthMethod, label)
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
	usagePrimitivesCmd.Flags().String("api-key-id", "", "filter to executions triggered by a single API key id (the api_key_id returned under triggered_by)")
	usagePrimitivesCmd.Flags().String("access-token-id", "", "filter to executions triggered by a single access token id (the access_token_id returned under triggered_by)")
	usagePrimitivesCmd.Flags().String("user-id", "", "filter to executions triggered by a single user id (the user_id returned under triggered_by)")
	usagePrimitivesCmd.Flags().String("run-id", "", "filter to a single workflow run id (origin run)")
	usagePrimitivesCmd.Flags().String("block-id", "", "filter to a single workflow block id (origin block); with --workflow-id, a declarative spec block id or path is resolved too")
	// Validated client-side like `usage runs`' --status/--trigger-type: an
	// unknown value can only match zero rows, and a silent empty page on a usage
	// export reads as "no spend" rather than "you typed it wrong". The server
	// rejects the same set with 422; this just fails before the round trip.
	usagePrimitivesCmd.Flags().Var(
		newEnumStringFlagValue("--operation",
			"extraction", "extract", "classification", "classify",
			"split", "parse", "edit", "partition", "schema_generation"),
		"operation", "filter by operation (extraction, classify, split, parse, edit, partition, schema_generation)")
	usagePrimitivesCmd.Flags().Var(
		// "canceled" carries ONE l here: that is how a primitive execution's
		// terminal status is stored. A workflow run's equivalent status is
		// spelled "cancelled", so `usage runs --status` is not interchangeable
		// with this flag.
		newEnumStringFlagValue("--status", "created", "running", "completed", "failed", "canceled"),
		"status", "filter by execution lifecycle status (created, running, completed, failed, canceled)")
	usagePrimitivesCmd.Flags().StringArray("metadata", nil, "filter by metadata key=value (repeatable; pairs AND together)")
	usagePrimitivesCmd.Flags().String("from-date", "", "inclusive created_at lower bound (YYYY-MM-DD, UTC)")
	usagePrimitivesCmd.Flags().String("to-date", "", "inclusive created_at upper bound (YYYY-MM-DD, UTC)")
	usagePrimitivesCmd.Flags().String("before", "", "execution id: return items before this id (mutually exclusive with --after)")
	usagePrimitivesCmd.Flags().String("after", "", "execution id: return items after this id (mutually exclusive with --before)")
	usagePrimitivesCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 10000}, "limit", "max items to return (1-10000)")
	usagePrimitivesCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")

	usageCmd.AddCommand(usagePrimitivesCmd)
}
