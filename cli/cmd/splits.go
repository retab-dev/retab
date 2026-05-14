package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var splitsCmd = &cobra.Command{
	Use:   "splits",
	Short: "Run and manage document splits",
}

var splitsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a split",
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
		subdocsFile, _ := cmd.Flags().GetString("subdocuments-file")
		if subdocsFile == "" {
			return fmt.Errorf("--subdocuments-file is required")
		}
		arr, err := readJSONArray(subdocsFile)
		if err != nil {
			return fmt.Errorf("--subdocuments-file: %w", err)
		}
		var subs []retab.SplitSubdocument
		for _, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("--subdocuments-file: each item must be a JSON object")
			}
			sub := retab.SplitSubdocument{}
			if v, ok := obj["name"].(string); ok {
				sub.Name = v
			}
			if v, ok := obj["description"].(string); ok {
				sub.Description = v
			}
			if v, ok := obj["allow_multiple_instances"].(bool); ok {
				sub.AllowMultipleInstances = v
			}
			subs = append(subs, sub)
		}
		model, _ := cmd.Flags().GetString("model")
		nConsensus, _ := cmd.Flags().GetInt("n-consensus")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		instructions, _ := cmd.Flags().GetString("instructions")
		result, err := client.Splits.Create(ctx, retab.SplitCreateRequest{
			Document:     doc,
			Subdocuments: subs,
			Model:        model,
			NConsensus:   nConsensus,
			BustCache:    bustCache,
			Instructions: instructions,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var splitsGetCmd = &cobra.Command{
	Use:   "get <split-id>",
	Short: "Get a split by id",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Splits.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var splitsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List splits",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := collectListParams(cmd)
		result, err := client.Splits.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var splitsDeleteCmd = &cobra.Command{
	Use:   "delete <split-id>",
	Short: "Delete a split",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Splits.Delete(ctx, args[0])
	}),
}

func init() {
	addDocumentFlags(splitsCreateCmd)
	splitsCreateCmd.Flags().String("subdocuments-file", "", "JSON array of subdocuments (name, description, allow_multiple_instances)")
	splitsCreateCmd.Flags().String("model", "", "model identifier (required)")
	splitsCreateCmd.Flags().Int("n-consensus", 0, "consensus count")
	splitsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	splitsCreateCmd.Flags().String("instructions", "", "extra instructions")
	_ = splitsCreateCmd.MarkFlagRequired("model")
	_ = splitsCreateCmd.MarkFlagRequired("subdocuments-file")

	addListFlags(splitsListCmd, false)

	splitsCmd.AddCommand(splitsCreateCmd, splitsGetCmd, splitsListCmd, splitsDeleteCmd)
	rootCmd.AddCommand(splitsCmd)
}
