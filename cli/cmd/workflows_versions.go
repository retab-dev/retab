//go:build !retab_oagen_cli_workflows

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// resolveWorkflowVersionRef maps the friendly selectors "production" /
// "published" / "live" / "latest" to the workflow's currently published version
// id, so the version inspect / compare / restore commands accept the same tokens
// as `runs create --version`. A literal "ver_..." id (or any other value) passes
// through untouched — the server returns 404 for a genuinely unknown id. "draft"
// is rejected with guidance because the draft is not a published version and the
// version endpoints cannot address it.
func resolveWorkflowVersionRef(ctx context.Context, client *retab.Client, workflowID, ref string) (string, error) {
	switch ref {
	case "production", "published", "live", "latest":
		wf, err := client.Workflows.Get(ctx, workflowID)
		if err != nil {
			return "", err
		}
		if wf == nil || wf.Published == nil || wf.Published.VersionID == nil || *wf.Published.VersionID == "" {
			return "", fmt.Errorf("workflow %s has no published version to resolve %q; publish it first", workflowID, ref)
		}
		return *wf.Published.VersionID, nil
	case "draft":
		return "", fmt.Errorf("%q is not a published version; the version commands operate on published versions only — inspect the draft with `retab workflows spec get %s` or `retab workflows blocks list %s`, or publish it first", ref, workflowID, workflowID)
	default:
		return ref, nil
	}
}

var workflowsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Inspect and restore workflow versions",
	Long:  `Inspect published workflow versions and restore immutable graph versions into the current draft.`,
}

type cliPublishedWorkflowVersion struct {
	ID                string  `json:"id"`
	WorkflowID        string  `json:"workflow_id"`
	Version           int     `json:"version"`
	WorkflowVersionID string  `json:"workflow_version_id"`
	Description       *string `json:"description,omitempty"`
	BlockCount        int     `json:"block_count"`
	EdgeCount         int     `json:"edge_count"`
	PublishedBy       *string `json:"published_by,omitempty"`
	PublishedByEmail  *string `json:"published_by_email,omitempty"`
	PublishedByName   *string `json:"published_by_name,omitempty"`
	PublishedAt       string  `json:"published_at"`
	IsCurrent         bool    `json:"is_current"`
}

// publishedVersionColumns is the dedicated TableColumn spec for
// `workflows versions list --output table/csv`.
//
// The generic auto-renderer matched only `id` out of preferredColumnOrder
// (these rows carry no name/type/created_at), collapsing the table to a single
// column of `wph_...` publication-record ids. That id is a dead end: nothing
// accepts it — `versions get/diff/restore` all take the `ver_...`
// workflow_version_id — so the sole identifier on screen 404s when pasted into
// the next command, while the one it needs was hidden. Lead with the version
// ordinal and the actionable id, and say which version is live.
var publishedVersionColumns = []TableColumn{
	{Header: "VERSION", Extract: func(row any) string { return workflowVersionCell(row, "version") }},
	{Header: "WORKFLOW_VERSION_ID", Extract: func(row any) string { return workflowVersionCell(row, "workflow_version_id") }},
	{Header: "CURRENT", Extract: publishedVersionCurrentCell},
	{Header: "DESCRIPTION", Extract: func(row any) string { return workflowVersionCell(row, "description") }},
	{Header: "PUBLISHED_AT", Extract: func(row any) string {
		return normalizeTimestampCell(workflowVersionCell(row, "published_at"))
	}},
}

func workflowVersionCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

// publishedVersionCurrentCell renders is_current. It is spelled out rather than
// routed through workflowVersionCell because cellIsEmpty treats the zero value
// (false) as absent, which would blank the column for every non-live version and
// leave the reader unable to tell "not current" from "field missing".
func publishedVersionCurrentCell(row any) string {
	value, ok := rowField(row, "is_current")
	if !ok {
		return ""
	}
	if current, isBool := value.(bool); isBool && current {
		return "yes"
	}
	return ""
}

var workflowsVersionsListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List published workflow versions",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		query := url.Values{}
		if limit > 0 {
			query.Set("limit", fmt.Sprint(limit))
		}
		var result cliPaginatedList[cliPublishedWorkflowVersion]
		path := fmt.Sprintf("/v1/workflows/%s/published-versions", url.PathEscape(args[0]))
		if err := cliJSONRequestInto(cmd, http.MethodGet, path, query, nil, &result); err != nil {
			return err
		}
		format := OutputAuto
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			if raw := f.Value.String(); raw != "" {
				format = OutputFormat(raw)
			}
		}
		if format == OutputTable || format == OutputCSV {
			return RenderList(os.Stdout, format, result, publishedVersionColumns)
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionsGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <workflow-version-id>",
	Short: "Get a workflow version",
	Long: `Get a single immutable workflow graph version.

The version selector accepts a literal ` + "`ver_...`" + ` id or the friendly token
` + "`production`" + ` (aliases ` + "`published`" + `, ` + "`live`" + `, ` + "`latest`" + `), which resolves to the
workflow's currently published version.`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		versionID, err := resolveWorkflowVersionRef(ctx, client, args[0], args[1])
		if err != nil {
			return err
		}
		result, err := client.Workflows.GetVersion(ctx, versionID, &retab.WorkflowsGetVersionParams{WorkflowID: args[0]})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionsDiffCmd = &cobra.Command{
	Use:   "diff <workflow-id> <from-version-id> <to-version-id>",
	Short: "Diff two workflow versions",
	Long: `Diff two immutable workflow graph versions.

Each version selector accepts a literal ` + "`ver_...`" + ` id or the friendly token
` + "`production`" + ` (aliases ` + "`published`" + `, ` + "`live`" + `, ` + "`latest`" + `), which resolves to the
workflow's currently published version — so ` + "`diff production ver_old`" + ` compares
the live version against an earlier one.`,
	Args: cobra.ExactArgs(3),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		fromID, err := resolveWorkflowVersionRef(ctx, client, args[0], args[1])
		if err != nil {
			return err
		}
		toID, err := resolveWorkflowVersionRef(ctx, client, args[0], args[2])
		if err != nil {
			return err
		}
		result, err := client.Workflows.ListDiff(ctx, &retab.WorkflowsListDiffParams{
			WorkflowID:            args[0],
			FromWorkflowVersionID: fromID,
			ToWorkflowVersionID:   toID,
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionsRestoreCmd = &cobra.Command{
	Use:   "restore <workflow-id> <workflow-version-id>",
	Short: "Restore a workflow version into the draft",
	Long: `Restore an immutable workflow graph version into the workflow's current draft.
This overwrites the editable draft graph with a fresh draft created from the
selected version. Pass ` + "`--yes`" + ` to confirm.

Restores into the DRAFT only — run ` + "`retab workflows publish <workflow-id>`" + ` to
make the restored graph the live version.`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "workflow draft", fmt.Sprintf("%s from %s", args[0], args[1])); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		versionID, err := resolveWorkflowVersionRef(ctx, client, args[0], args[1])
		if err != nil {
			return err
		}
		result, err := client.Workflows.CreateVersionRestore(ctx, versionID, &retab.WorkflowsCreateVersionRestoreParams{WorkflowID: args[0]})
		if err != nil {
			return err
		}
		printDraftPublishHint(cmd, args[0])
		return printResult(cmd, result)
	}),
}

func init() {
	workflowsVersionsListCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")
	workflowsVersionsRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	workflowsVersionsCmd.AddCommand(workflowsVersionsListCmd, workflowsVersionsGetCmd, workflowsVersionsDiffCmd, workflowsVersionsRestoreCmd)
	workflowsCmd.AddCommand(workflowsVersionsCmd)
}
