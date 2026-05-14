package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var classificationsCmd = &cobra.Command{
	Use:   "classifications",
	Short: "Run and manage classifications",
}

var classificationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a classification",
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

		categoriesFile, _ := cmd.Flags().GetString("categories-file")
		categoryFlags, _ := cmd.Flags().GetStringArray("category")
		var categories []retab.ClassificationCategory
		if categoriesFile != "" {
			arr, err := readJSONArray(categoriesFile)
			if err != nil {
				return fmt.Errorf("--categories-file: %w", err)
			}
			for _, item := range arr {
				obj, ok := item.(map[string]any)
				if !ok {
					return fmt.Errorf("--categories-file: each item must be a JSON object")
				}
				cat := retab.ClassificationCategory{}
				if v, ok := obj["name"].(string); ok {
					cat.Name = v
				}
				if v, ok := obj["description"].(string); ok {
					cat.Description = v
				}
				categories = append(categories, cat)
			}
		}
		for _, raw := range categoryFlags {
			name, desc, _ := splitKV(raw)
			categories = append(categories, retab.ClassificationCategory{Name: name, Description: desc})
		}
		if len(categories) == 0 {
			return fmt.Errorf("at least one category is required (--category or --categories-file)")
		}

		model, _ := cmd.Flags().GetString("model")
		nConsensus, _ := cmd.Flags().GetInt("n-consensus")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		firstN, _ := cmd.Flags().GetInt("first-n-pages")
		instructions, _ := cmd.Flags().GetString("instructions")
		result, err := client.Classifications.Create(ctx, retab.ClassificationCreateRequest{
			Document:     doc,
			Categories:   categories,
			Model:        model,
			NConsensus:   nConsensus,
			BustCache:    bustCache,
			FirstNPages:  firstN,
			Instructions: instructions,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var classificationsGetCmd = &cobra.Command{
	Use:   "get <classification-id>",
	Short: "Get a classification by id",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Classifications.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var classificationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List classifications",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := collectListParams(cmd)
		result, err := client.Classifications.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var classificationsDeleteCmd = &cobra.Command{
	Use:   "delete <classification-id>",
	Short: "Delete a classification",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Classifications.Delete(ctx, args[0])
	}),
}

func init() {
	addDocumentFlags(classificationsCreateCmd)
	classificationsCreateCmd.Flags().String("model", "", "model identifier (required)")
	classificationsCreateCmd.Flags().Int("n-consensus", 0, "consensus count")
	classificationsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	classificationsCreateCmd.Flags().Int("first-n-pages", 0, "only classify the first N pages")
	classificationsCreateCmd.Flags().String("instructions", "", "extra instructions")
	classificationsCreateCmd.Flags().StringArray("category", nil, "category as name=description (repeatable)")
	classificationsCreateCmd.Flags().String("categories-file", "", "JSON array of {name, description} (or - for stdin)")
	_ = classificationsCreateCmd.MarkFlagRequired("model")

	addListFlags(classificationsListCmd, false)

	classificationsCmd.AddCommand(classificationsCreateCmd, classificationsGetCmd, classificationsListCmd, classificationsDeleteCmd)
	rootCmd.AddCommand(classificationsCmd)
}
