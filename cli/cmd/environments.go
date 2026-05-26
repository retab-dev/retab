package cmd

import (
	"fmt"
	"os"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

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

  # Create a non-production environment
  retab env add --name Staging --type non_production

  # Select by environment id or exact name
  retab env switch Staging

  # WorkOS-style synonym for selecting a local environment
  retab env claim environment_staging

  # Archive an environment
  retab env remove environment_abc123 --yes`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organization environments",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Environments.List(ctx, nil)
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
		params := &retab.EnvironmentsCreateParams{Name: name}
		if rawType != "" {
			environmentType := retab.EnvironmentCreateRequestType(rawType)
			params.Type = &environmentType
		}

		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Environments.Create(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var envRemoveCmd = &cobra.Command{
	Use:   "remove <environment-id>",
	Short: "Archive an environment",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "environment", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Environments.Delete(ctx, args[0]); err != nil {
			return err
		}
		cfg, _ := loadConfig()
		if cfg.EnvironmentID == args[0] {
			cfg.EnvironmentID = ""
			if err := saveConfig(cfg); err != nil {
				return err
			}
		}
		return nil
	}),
}

var envSwitchCmd = &cobra.Command{
	Use:   "switch <environment-id-or-name>",
	Short: "Select the environment used by CLI requests",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		list, err := client.Environments.List(ctx, nil)
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
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Environments.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var envUpdateCmd = &cobra.Command{
	Use:    "update <environment-id>",
	Short:  "Update an environment",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		name, err := requireNonBlankFlag(cmd, "name")
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Environments.Update(ctx, args[0], &retab.EnvironmentsUpdateParams{Name: &name})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func printEnvironmentList(cmd *cobra.Command, result *retab.PaginatedList[retab.Environment]) error {
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
	if environment, ok := row.(retab.Environment); ok {
		return environmentStructCell(environment, field)
	}
	if environment, ok := row.(*retab.Environment); ok && environment != nil {
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

func environmentStructCell(environment retab.Environment, field string) string {
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

func resolveEnvironmentSelection(raw string, list *retab.PaginatedList[retab.Environment]) (*retab.Environment, error) {
	needle := strings.TrimSpace(raw)
	if needle == "" {
		return nil, fmt.Errorf("environment id or name is required")
	}
	var nameMatches []*retab.Environment
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
	envRemoveCmd.Flags().Bool("yes", false, "confirm deletion without prompting")
	envUpdateCmd.Flags().String("name", "", "environment name")

	envCreateAliasCmd := &cobra.Command{
		Use:    "create",
		Short:  envAddCmd.Short,
		Hidden: true,
		Args:   envAddCmd.Args,
		RunE:   envAddCmd.RunE,
	}
	envCreateAliasCmd.Flags().String("name", "", "environment name")
	envCreateAliasCmd.Flags().String("type", "non_production", "environment type: production | non_production")

	envDeleteAliasCmd := &cobra.Command{
		Use:    "delete <environment-id>",
		Short:  envRemoveCmd.Short,
		Hidden: true,
		Args:   envRemoveCmd.Args,
		RunE:   envRemoveCmd.RunE,
	}
	envDeleteAliasCmd.Flags().Bool("yes", false, "confirm deletion without prompting")

	envCmd.AddCommand(envListCmd, envAddCmd, envRemoveCmd, envSwitchCmd, envClaimCmd, envGetCmd, envUpdateCmd, envCreateAliasCmd, envDeleteAliasCmd)
	rootCmd.AddCommand(envCmd)
}
