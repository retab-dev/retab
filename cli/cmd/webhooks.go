package cmd

import (
	"fmt"
	"io"
	"os"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Webhook utilities",
}

var webhooksVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a Retab webhook signature and print the parsed event",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		body, err := readBody(cmd)
		if err != nil {
			return err
		}
		signature, _ := cmd.Flags().GetString("signature")
		secret, _ := cmd.Flags().GetString("secret")
		if signature == "" {
			signature = os.Getenv("RETAB_WEBHOOK_SIGNATURE")
		}
		if secret == "" {
			secret = os.Getenv("RETAB_WEBHOOK_SECRET")
		}
		if signature == "" {
			return fmt.Errorf("--signature or RETAB_WEBHOOK_SIGNATURE is required")
		}
		if secret == "" {
			return fmt.Errorf("--secret or RETAB_WEBHOOK_SECRET is required")
		}
		event, err := retab.VerifyEvent[map[string]any](body, signature, secret)
		if err != nil {
			return err
		}
		return printJSON(event)
	}),
}

func readBody(cmd *cobra.Command) ([]byte, error) {
	path, _ := cmd.Flags().GetString("body-file")
	literal, _ := cmd.Flags().GetString("body")
	if literal != "" && path != "" {
		return nil, fmt.Errorf("--body and --body-file are mutually exclusive")
	}
	if literal != "" {
		return []byte(literal), nil
	}
	if path == "" || path == "-" {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return raw, nil
	}
	return os.ReadFile(path)
}

func init() {
	webhooksVerifyCmd.Flags().String("body", "", "raw webhook body as a string")
	webhooksVerifyCmd.Flags().String("body-file", "", "path to webhook body (or - for stdin)")
	webhooksVerifyCmd.Flags().String("signature", "", "webhook signature (env: RETAB_WEBHOOK_SIGNATURE)")
	webhooksVerifyCmd.Flags().String("secret", "", "webhook secret (env: RETAB_WEBHOOK_SECRET)")

	webhooksCmd.AddCommand(webhooksVerifyCmd)
	rootCmd.AddCommand(webhooksCmd)
}
