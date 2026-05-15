package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var classificationsCmd = &cobra.Command{
	Use:   "classifications",
	Short: "Categorize documents based on content and type",
	Long: `Assign documents to one of a fixed set of categories.

A classification takes a document plus a list of named categories (each
with a short natural-language description) and returns the chosen category
along with confidence. Useful for routing — e.g. invoice vs. receipt vs.
purchase order — before running a category-specific extraction with
` + "`retab extractions create`" + `.`,
}

var classificationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a classification",
	Long: `Classify a document into one of a fixed set of named categories.

Supply categories either inline with ` + "`--category name=description`" + ` (one
flag per category, repeatable) or as a JSON array via
` + "`--categories-file`" + ` (path, or ` + "`-`" + ` for stdin; each item must be
` + "`{\"name\":..., \"description\":...}`" + `). Descriptions should be short
natural-language guidance that disambiguates close categories.

For higher accuracy on ambiguous documents, set ` + "`--n-consensus`" + ` to
sample multiple model runs and return the majority pick. Use
` + "`--first-n-pages`" + ` to classify only the first few pages of long docs
when the type is obvious from the cover.`,
	Example: `  # Classify with inline categories
  retab classifications create \
    --file ./doc.pdf --model gpt-4o \
    --category invoice="vendor bills with line items and totals" \
    --category receipt="point-of-sale purchase confirmations" \
    --category po="buyer purchase orders before goods ship"

  # Same, but categories loaded from a JSON file
  retab classifications create \
    --file ./doc.pdf --model gpt-4o \
    --categories-file ./categories.json

  # Higher-accuracy consensus over 5 samples, only first page
  retab classifications create \
    --file-id file_abc123 --model gpt-4o \
    --categories-file ./categories.json \
    --n-consensus 5 --first-n-pages 1`,
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
	Long: `Fetch a single classification by id.

Returns the chosen category, confidence, the full list of candidate
categories supplied at create time, and the source document reference.`,
	Example: `  # Fetch a known classification
  retab classifications get clas_xyz789

  # Extract just the chosen category name
  retab classifications get clas_xyz789 | jq -r '.category'`,
	Args: cobra.ExactArgs(1),
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
	Long: `List classifications, newest first by default.

Results are paginated by classification id via ` + "`--after`" + ` / ` + "`--before`" + `. Use
` + "`--limit`" + ` to cap the page size and ` + "`--order`" + ` to flip between
ascending and descending.`,
	Example: `  # Most recent 25 classifications
  retab classifications list --limit 25

  # Page from a known classification id
  retab classifications list --after clas_xyz789 --limit 50`,
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
		return printResult(cmd, result)
	}),
}

var classificationsDeleteCmd = &cobra.Command{
	Use:   "delete <classification-id>",
	Short: "Delete a classification",
	Long: `Permanently delete a classification.

Destructive and irreversible. The source document is not affected. Take a
backup with ` + "`retab classifications get`" + ` first if you may need it.`,
	Example: `  # Back up, then delete
  retab classifications get clas_xyz789 > backup.json
  retab classifications delete clas_xyz789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Classifications.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("classification", args[0])
		return nil
	}),
}

func init() {
	addDocumentFlags(classificationsCreateCmd)
	classificationsCreateCmd.Flags().String("model", "", "model identifier (required)")
	classificationsCreateCmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "consensus count")
	classificationsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	classificationsCreateCmd.Flags().Var(&nonNegativeIntFlagValue{}, "first-n-pages", "only classify the first N pages")
	classificationsCreateCmd.Flags().String("instructions", "", "extra instructions")
	classificationsCreateCmd.Flags().StringArray("category", nil, "category as name=description (repeatable)")
	classificationsCreateCmd.Flags().String("categories-file", "", "JSON array of {name, description} (or - for stdin)")
	_ = classificationsCreateCmd.MarkFlagRequired("model")

	addListFlags(classificationsListCmd, false)

	classificationsCmd.AddCommand(classificationsCreateCmd, classificationsGetCmd, classificationsListCmd, classificationsDeleteCmd)
	rootCmd.AddCommand(classificationsCmd)
}
