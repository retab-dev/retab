package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type cliPaginatedList[T any] struct {
	Data         []T `json:"data"`
	ListMetadata any `json:"list_metadata,omitempty"`
}

type cliEnvironmentType string

const (
	cliEnvironmentTypeProduction    cliEnvironmentType = "production"
	cliEnvironmentTypeNonProduction cliEnvironmentType = "non_production"
)

type cliEnvironment struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Type      cliEnvironmentType `json:"type"`
	IsDefault *bool              `json:"is_default,omitempty"`
}

type cliCreateEnvironmentRequest struct {
	Name string              `json:"name"`
	Type *cliEnvironmentType `json:"type,omitempty"`
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage Retab environments",
	Long: `Manage the Retab environment selected for CLI calls.

Environment names are user-facing labels and can change. The CLI stores the
durable environment id in ~/.retab/config.json. OAuth-backed API calls resolve
that selection by minting a short-lived Retab dashboard context token; the
environment is never sent as a public request header. Per-environment API keys
remain scoped by the server-side key record.`,
	Example: `  # List environments and see the active local selection
  retab env list

  # Show the selected local environment
  retab env which

  # Create a non-production environment
  retab env add --name Staging --type non_production

  # Select by environment id, name, or slug (name/slug are case-insensitive)
  retab env switch Staging
  retab env switch staging

  # WorkOS-style synonym for selecting a local environment
  retab env claim env_staging`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organization environments",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		result, err := listCLIEnvironments(cmd)
		if err != nil {
			return err
		}
		return printEnvironmentList(cmd, result)
	}),
}

var envAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"create"},
	Short:   "Create an environment",
	Args:    cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		name, err := requireNonBlankFlag(cmd, "name")
		if err != nil {
			return err
		}
		rawType, _ := cmd.Flags().GetString("type")
		if rawType != "" && rawType != "production" && rawType != "non_production" {
			return fmt.Errorf("invalid --type %q (want: production | non_production)", rawType)
		}
		params := &cliCreateEnvironmentRequest{Name: name}
		if rawType != "" {
			environmentType := cliEnvironmentType(rawType)
			params.Type = &environmentType
		}

		result, err := createCLIEnvironment(cmd, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var envSwitchCmd = &cobra.Command{
	Use:   "switch <environment-id-name-or-slug>",
	Short: "Select the environment used by CLI requests",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		list, err := listCLIEnvironments(cmd)
		if err != nil {
			return err
		}
		environment, err := resolveEnvironmentSelection(args[0], list)
		if err != nil {
			return err
		}
		cfg, _ := loadConfig()
		cfg.EnvironmentID = environment.ID
		// Persist the type so the offline production-confirmation gate can
		// tell whether this OAuth session targets production.
		cfg.EnvironmentType = string(environment.Type)
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Selected environment %s (%s)\n", environment.Name, environment.ID)
		return nil
	}),
}

type selectedEnvironment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"is_default"`
	Source    string `json:"source"`
}

var envWhichCmd = &cobra.Command{
	Use:   "which",
	Short: "Show the selected environment",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		cfg, _ := loadConfig()
		environmentID, source := selectedEnvironmentIDWithSource(cmd, cfg)
		if environmentID == "" {
			return fmt.Errorf("no environment selected. Run `retab env switch <environment-id-name-or-slug>`")
		}

		environment, err := getCLIEnvironment(cmd, environmentID)
		if err != nil {
			return err
		}
		return printSelectedEnvironment(cmd, environment, source)
	}),
}

var envClaimCmd = &cobra.Command{
	Use:   "claim <environment-id-name-or-slug>",
	Short: "Claim an environment for this local CLI config",
	Args:  envSwitchCmd.Args,
	RunE:  envSwitchCmd.RunE,
}

var envGetCmd = &cobra.Command{
	Use:   "get <environment-id>",
	Short: "Get an environment",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		result, err := getCLIEnvironment(cmd, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func listCLIEnvironments(cmd *cobra.Command) (*cliPaginatedList[cliEnvironment], error) {
	var result cliPaginatedList[cliEnvironment]
	err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/environments", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func createCLIEnvironment(cmd *cobra.Command, request *cliCreateEnvironmentRequest) (*cliEnvironment, error) {
	var result cliEnvironment
	err := cliJSONRequestInto(cmd, http.MethodPost, "/v1/environments", nil, request, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func getCLIEnvironment(cmd *cobra.Command, environmentID string) (*cliEnvironment, error) {
	var result cliEnvironment
	err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/environments/"+url.PathEscape(environmentID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func printEnvironmentList(cmd *cobra.Command, result *cliPaginatedList[cliEnvironment]) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		cfg, _ := loadConfig()
		return RenderList(os.Stdout, format, map[string]any{"data": result.Data}, environmentTableColumns(selectedEnvironmentID(cmd, cfg)))
	}
	return printResult(cmd, result)
}

func environmentTableColumns(activeEnvironmentID string) []TableColumn {
	return []TableColumn{
		{Header: "ID", Extract: func(row any) string { return environmentCell(row, "id") }},
		{Header: "NAME", Extract: func(row any) string { return environmentCell(row, "name") }},
		{Header: "TYPE", Extract: func(row any) string { return environmentCell(row, "type") }},
		{Header: "DEFAULT", Extract: func(row any) string { return environmentCell(row, "is_default") }},
		{Header: "ACTIVE", Extract: func(row any) string {
			if environmentCell(row, "id") == activeEnvironmentID {
				return "true"
			}
			return ""
		}},
	}
}

func environmentCell(row any, field string) string {
	if environment, ok := row.(cliEnvironment); ok {
		return environmentStructCell(environment, field)
	}
	if environment, ok := row.(*cliEnvironment); ok && environment != nil {
		return environmentStructCell(*environment, field)
	}
	if value, ok := rowField(row, field); ok {
		if field == "is_default" {
			if isDefault, ok := value.(bool); ok && isDefault {
				return "true"
			}
			return ""
		}
		return stringifyCell(value)
	}
	return ""
}

func environmentStructCell(environment cliEnvironment, field string) string {
	switch field {
	case "id":
		return environment.ID
	case "name":
		return environment.Name
	case "type":
		return string(environment.Type)
	case "is_default":
		if environment.IsDefault != nil && *environment.IsDefault {
			return "true"
		}
		return ""
	default:
		return ""
	}
}

func printSelectedEnvironment(cmd *cobra.Command, environment *cliEnvironment, source string) error {
	format, err := ResolveOutputFormat(cmd, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	isDefault := false
	if environment.IsDefault != nil {
		isDefault = *environment.IsDefault
	}
	selection := selectedEnvironment{
		ID:        environment.ID,
		Name:      environment.Name,
		Type:      string(environment.Type),
		IsDefault: isDefault,
		Source:    source,
	}
	if format == OutputTable || format == OutputCSV {
		return RenderList(
			cmd.OutOrStdout(),
			format,
			map[string]any{"data": []selectedEnvironment{selection}},
			selectedEnvironmentTableColumns(),
		)
	}
	return renderJSON(cmd.OutOrStdout(), selection)
}

func selectedEnvironmentTableColumns() []TableColumn {
	return []TableColumn{
		{Header: "ID", Extract: func(row any) string { return selectedEnvironmentCell(row, "id") }},
		{Header: "NAME", Extract: func(row any) string { return selectedEnvironmentCell(row, "name") }},
		{Header: "TYPE", Extract: func(row any) string { return selectedEnvironmentCell(row, "type") }},
		{Header: "DEFAULT", Extract: func(row any) string { return selectedEnvironmentCell(row, "is_default") }},
		{Header: "SOURCE", Extract: func(row any) string { return selectedEnvironmentCell(row, "source") }},
	}
}

func selectedEnvironmentCell(row any, field string) string {
	if selection, ok := row.(selectedEnvironment); ok {
		return selectedEnvironmentStructCell(selection, field)
	}
	if selection, ok := row.(*selectedEnvironment); ok && selection != nil {
		return selectedEnvironmentStructCell(*selection, field)
	}
	if value, ok := rowField(row, field); ok {
		if field == "is_default" {
			if isDefault, ok := value.(bool); ok && isDefault {
				return "true"
			}
			return ""
		}
		return stringifyCell(value)
	}
	return ""
}

func selectedEnvironmentStructCell(selection selectedEnvironment, field string) string {
	switch field {
	case "id":
		return selection.ID
	case "name":
		return selection.Name
	case "type":
		return selection.Type
	case "is_default":
		if selection.IsDefault {
			return "true"
		}
		return ""
	case "source":
		return selection.Source
	default:
		return ""
	}
}

// resolveEnvironmentSelection maps user input (an environment id, name, or
// slug) to a single environment. All three are unique keys, so any of them is
// accepted, and name/slug matching is case-insensitive ("staging" selects
// "Staging"). Precedence, highest first:
//
//  1. exact id match
//  2. exact (case-sensitive) name match
//  3. case-insensitive name/slug match
//
// Each tier is resolved against the whole list before falling through to the
// next, so an exact id always wins over a name collision and an exact-case
// name always wins over a fold collision. Ambiguity within a tier, or no match
// at all, fails loudly — the CLI never silently falls back to the active
// environment, which was the original source of "the set flipped on its own"
// confusion.
func resolveEnvironmentSelection(raw string, list *cliPaginatedList[cliEnvironment]) (*cliEnvironment, error) {
	needle := strings.TrimSpace(raw)
	if needle == "" {
		return nil, fmt.Errorf("environment id, name, or slug is required")
	}

	var exactNameMatches, foldNameMatches []*cliEnvironment
	for i := range list.Data {
		environment := &list.Data[i]
		// 1. Exact id wins outright — ids are opaque durable keys.
		if environment.ID == needle {
			return environment, nil
		}
		switch {
		case environment.Name == needle:
			exactNameMatches = append(exactNameMatches, environment)
		case strings.EqualFold(environment.Name, needle):
			foldNameMatches = append(foldNameMatches, environment)
		}
	}

	// 2. Prefer an exact-case name before considering case-folded matches, so
	//    a distinct exact name never loses to an unrelated fold collision.
	if match, err := pickSingleEnvironment(needle, exactNameMatches); match != nil || err != nil {
		return match, err
	}
	// 3. Case-insensitive name/slug ("staging" -> "Staging").
	if match, err := pickSingleEnvironment(needle, foldNameMatches); match != nil || err != nil {
		return match, err
	}

	return nil, fmt.Errorf("environment %q not found.%s\nRun `retab env list` to see available environments", needle, availableEnvironmentsHint(list))
}

// pickSingleEnvironment returns the lone match, a loud ambiguity error when a
// name resolves to more than one environment, or (nil, nil) when the tier is
// empty so the caller can fall through to the next tier.
func pickSingleEnvironment(needle string, matches []*cliEnvironment) (*cliEnvironment, error) {
	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("environment name %q is ambiguous (%d matches); select it by id instead", needle, len(matches))
	}
}

// availableEnvironmentsHint renders a short "id (name)" list for not-found
// errors so the failure is actionable rather than a dead end.
func availableEnvironmentsHint(list *cliPaginatedList[cliEnvironment]) string {
	if list == nil || len(list.Data) == 0 {
		return ""
	}
	entries := make([]string, 0, len(list.Data))
	for i := range list.Data {
		environment := &list.Data[i]
		entries = append(entries, fmt.Sprintf("%s (%s)", environment.ID, environment.Name))
	}
	return " Available: " + strings.Join(entries, ", ")
}

func init() {
	envAddCmd.Flags().String("name", "", "environment name")
	envAddCmd.Flags().String("type", "non_production", "environment type: production | non_production")

	envCmd.AddCommand(envListCmd, envAddCmd, envSwitchCmd, envWhichCmd, envClaimCmd, envGetCmd)
	rootCmd.AddCommand(envCmd)
}
