//go:build !retab_oagen_cli_secrets

package cmd

import (
	"fmt"
	"io"
	"os"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage environment-scoped secrets",
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Secrets.List(ctx)
		if err != nil {
			return err
		}
		return printSecretsListResult(cmd, result)
	}),
}

var secretsGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get secret metadata",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Secrets.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var secretsSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set a secret value",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		value, err := secretValueFromInput(cmd)
		if err != nil {
			return err
		}
		if value == "" {
			return fmt.Errorf("secret value must not be empty")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Secrets.Update(ctx, args[0], &retab.SecretsUpdateParams{Value: value})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var secretsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "secret", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Secrets.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("secret", args[0])
		return nil
	}),
}

func secretValueFromInput(cmd *cobra.Command) (string, error) {
	fromFile, _ := cmd.Flags().GetString("from-file")
	fromStdin, _ := cmd.Flags().GetBool("from-stdin")
	if fromFile != "" && fromStdin {
		return "", fmt.Errorf("--from-file and --from-stdin are mutually exclusive")
	}
	if fromFile != "" {
		raw, err := os.ReadFile(fromFile)
		if err != nil {
			return "", fmt.Errorf("read --from-file: %w", err)
		}
		return string(raw), nil
	}
	if fromStdin {
		raw, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("read --from-stdin: %w", err)
		}
		return string(raw), nil
	}
	return promptSecret("Secret value: ")
}

func printSecretsListResult(cmd *cobra.Command, result *retab.SecretListResponse) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputJSON {
		return printJSON(result)
	}
	rows := make([]any, 0)
	if result != nil {
		rows = make([]any, 0, len(result.Secrets))
		for _, secret := range result.Secrets {
			rows = append(rows, secret)
		}
	}
	return renderAutoTable(os.Stdout, rows, secretListColumns)
}

var secretListColumns = []TableColumn{
	{Header: "NAME", Extract: func(row any) string { return secretCell(row, "name") }},
	{Header: "UPDATED_AT", Extract: func(row any) string { return secretCell(row, "updated_at") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return secretCell(row, "created_at") }},
}

func secretCell(row any, key string) string {
	switch secret := row.(type) {
	case *retab.Secret:
		if secret == nil {
			return ""
		}
		switch key {
		case "name":
			return secret.Name
		case "created_at":
			return stringifyCell(secret.CreatedAt)
		case "updated_at":
			return stringifyCell(secret.UpdatedAt)
		default:
			return ""
		}
	case retab.Secret:
		return secretCell(&secret, key)
	case map[string]any:
		return stringifyCell(secret[key])
	default:
		return ""
	}
}

func init() {
	secretsSetCmd.Flags().String("from-file", "", "read secret value from a file")
	secretsSetCmd.Flags().Bool("from-stdin", false, "read secret value from stdin")
	secretsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	secretsCmd.AddCommand(secretsListCmd, secretsGetCmd, secretsSetCmd, secretsDeleteCmd)
	rootCmd.AddCommand(secretsCmd)
}
