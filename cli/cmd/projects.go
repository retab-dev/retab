package cmd

import (
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// cliProject mirrors the server's `Project` response (the env-scoped
// dashboard `/v1/projects` surface). Only the fields a CLI user needs to
// pick a project id are decoded; unknown fields are ignored. Projects are
// a dashboard-context resource (env-scoped, not part of the public
// OpenAPI / generated SDK), so — like `env list` — the command talks to
// the endpoint directly via cliJSONRequestInto rather than the SDK.
type cliProject struct {
	ID            string `json:"id"`
	EnvironmentID string `json:"environment_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description,omitempty"`
	Visibility    string `json:"visibility,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long: `List the projects in the selected environment.

A workflow must belong to a project — ` + "`workflows create`" + ` and
` + "`workflows spec apply`" + ` both require ` + "`--project-id`" + `. Use
` + "`retab projects list`" + ` to discover the project ids available in the
active environment.

Projects are environment-scoped: the list reflects the environment selected
with ` + "`retab env switch`" + ` (or the ` + "`--environment-id`" + ` flag).`,
	Example: `  # List projects in the active environment
  retab projects list

  # Capture a project id for piping into workflows create
  PROJ=$(retab projects list | jq -r '.data[0].id')
  retab workflows create --name "Invoices" --project-id "$PROJ"`,
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects in the selected environment",
	Long: `List projects in the active environment, with id pagination.
Use ` + "`--after`" + ` / ` + "`--before`" + ` to walk through pages.`,
	Example: `  # First page
  retab projects list --limit 20

  # Next page (use the last project id from the previous response)
  retab projects list --after proj_abc123 --limit 20

  # Include archived projects
  retab projects list --include-archived`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		query := url.Values{}
		if before, _ := cmd.Flags().GetString("before"); before != "" {
			query.Set("before", before)
		}
		if after, _ := cmd.Flags().GetString("after"); after != "" {
			query.Set("after", after)
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			query.Set("limit", strconv.Itoa(limit))
		}
		if includeArchived, _ := cmd.Flags().GetBool("include-archived"); includeArchived {
			query.Set("include_archived", "true")
		}

		var result cliPaginatedList[cliProject]
		if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/projects", query, nil, &result); err != nil {
			return err
		}
		return printProjectsList(cmd, &result)
	}),
}

var projectsGetCmd = &cobra.Command{
	Use:   "get <project-id>",
	Short: "Get a project",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		var result cliProject
		if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/projects/"+url.PathEscape(args[0]), nil, nil, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var projectColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return projectCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return projectCell(row, "name") }},
	{Header: "SLUG", Extract: func(row any) string { return projectCell(row, "slug") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return normalizeTimestampCell(projectCell(row, "created_at")) }},
}

func projectCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func printProjectsList(cmd *cobra.Command, result *cliPaginatedList[cliProject]) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		return RenderList(os.Stdout, format, result, projectColumns)
	}
	return printJSON(result)
}

func init() {
	projectsListCmd.Flags().String("before", "", "project id: return items before this id (mutually exclusive with --after)")
	projectsListCmd.Flags().String("after", "", "project id: return items after this id (mutually exclusive with --before)")
	projectsListCmd.Flags().Int("limit", 0, "maximum number of projects to return")
	projectsListCmd.Flags().Bool("include-archived", false, "include archived projects in the list")

	projectsCmd.AddCommand(projectsListCmd, projectsGetCmd)
	rootCmd.AddCommand(projectsCmd)
}
