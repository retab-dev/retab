package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var parsesCmd = &cobra.Command{
	Use:   "parses",
	Short: "Run and manage document parses",
}

var parsesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a parse",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		doc, err := resolveDocument(cmd)
		if err != nil {
			return err
		}
		model, _ := cmd.Flags().GetString("model")
		tableFormat, _ := cmd.Flags().GetString("table-parsing-format")
		dpi, _ := cmd.Flags().GetInt("image-resolution-dpi")
		instructions, _ := cmd.Flags().GetString("instructions")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		result, err := client.Parses.Create(ctx, retab.ParseCreateRequest{
			Document:           doc,
			Model:              model,
			TableParsingFormat: tableFormat,
			ImageResolutionDPI: dpi,
			Instructions:       instructions,
			BustCache:          bustCache,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var parsesGetCmd = &cobra.Command{
	Use:   "get <parse-id>",
	Short: "Get a parse by id",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Parses.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var parsesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List parses",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := collectListParams(cmd)
		result, err := client.Parses.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var parsesDeleteCmd = &cobra.Command{
	Use:   "delete <parse-id>",
	Short: "Delete a parse",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Parses.Delete(ctx, args[0])
	}),
}

func init() {
	addDocumentFlags(parsesCreateCmd)
	parsesCreateCmd.Flags().String("model", "", "model identifier (required)")
	parsesCreateCmd.Flags().String("table-parsing-format", "", "table parsing format")
	parsesCreateCmd.Flags().Int("image-resolution-dpi", 0, "image resolution DPI")
	parsesCreateCmd.Flags().String("instructions", "", "extra instructions")
	parsesCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	_ = parsesCreateCmd.MarkFlagRequired("model")

	addListFlags(parsesListCmd, false)

	parsesCmd.AddCommand(parsesCreateCmd, parsesGetCmd, parsesListCmd, parsesDeleteCmd)
	rootCmd.AddCommand(parsesCmd)
}
