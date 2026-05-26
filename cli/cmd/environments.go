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

  # Select by environment id or exact name
  retab env switch Staging

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
	Use:   "add",
	Short: "Create an environment",
	Args:  cobra.NoArgs,
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
	Use:   "switch <environment-id-or-name>",
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
			return fmt.Errorf("no environment selected. Run `retab env switch <environment-id-or-name>`")
		}

		environment, err := getCLIEnvironment(cmd, environmentID)
		if err != nil {
			return err
		}
		return printSelectedEnvironment(cmd, environment, source)
	}),
}

var envClaimCmd = &cobra.Command{
	Use:   "claim <environment-id-or-name>",
	Short: "Claim an environment for this local CLI config",
	Args:  envSwitchCmd.Args,
	RunE:  envSwitchCmd.RunE,
}

var envGetCmd = &cobra.Command{
	Use:    "get <environment-id>",
	Short:  "Get an environment",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
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
	if format == OutputTable {
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
	if format == OutputTable {
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

func resolveEnvironmentSelection(raw string, list *cliPaginatedList[cliEnvironment]) (*cliEnvironment, error) {
	needle := strings.TrimSpace(raw)
	if needle == "" {
		return nil, fmt.Errorf("environment id or name is required")
	}
	var nameMatches []*cliEnvironment
	for i := range list.Data {
		environment := &list.Data[i]
		if environment.ID == needle {
			return environment, nil
		}
		if environment.Name == needle {
			nameMatches = append(nameMatches, environment)
		}
	}
	if len(nameMatches) == 1 {
		return nameMatches[0], nil
	}
	if len(nameMatches) > 1 {
		return nil, fmt.Errorf("environment name %q is ambiguous; use an environment id", needle)
	}
	return nil, fmt.Errorf("environment %q not found", needle)
}

func init() {
	envAddCmd.Flags().String("name", "", "environment name")
	envAddCmd.Flags().String("type", "non_production", "environment type: production | non_production")

	envCreateAliasCmd := &cobra.Command{
		Use:    "create",
		Short:  envAddCmd.Short,
		Hidden: true,
		Args:   envAddCmd.Args,
		RunE:   envAddCmd.RunE,
	}
	envCreateAliasCmd.Flags().String("name", "", "environment name")
	envCreateAliasCmd.Flags().String("type", "non_production", "environment type: production | non_production")

	envCmd.AddCommand(envListCmd, envAddCmd, envSwitchCmd, envWhichCmd, envClaimCmd, envGetCmd, envCreateAliasCmd)
	rootCmd.AddCommand(envCmd)
}
