//go:build !retab_oagen_cli_classifications

package cmd

import (
	"fmt"
	"strings"

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
		model, err := requireNonBlankFlag(cmd, "model")
		if err != nil {
			return err
		}
		categoriesFile, _ := cmd.Flags().GetString("categories-file")
		categoryFlags, _ := cmd.Flags().GetStringArray("category")
		var categories []*retab.Category
		if categoriesFile != "" {
			arr, err := readJSONArray(categoriesFile)
			if err != nil {
				return fmt.Errorf("--categories-file: %w", err)
			}
			for i, item := range arr {
				obj, ok := item.(map[string]any)
				if !ok {
					return fmt.Errorf("--categories-file[%d] must be a JSON object", i)
				}
				cat := &retab.Category{}
				nameValue, ok := obj["name"]
				if !ok {
					return fmt.Errorf("--categories-file[%d].name is required", i)
				}
				name, ok := nameValue.(string)
				if !ok {
					return fmt.Errorf("--categories-file[%d].name must be a string", i)
				}
				cat.Name = name
				if v, ok := obj["description"]; ok {
					description, ok := v.(string)
					if !ok {
						return fmt.Errorf("--categories-file[%d].description must be a string", i)
					}
					cat.Description = ptr(description)
				}
				if err := validateClassificationCategory(i, cat, "--categories-file"); err != nil {
					return err
				}
				categories = append(categories, cat)
			}
		}
		for i, raw := range categoryFlags {
			name, desc, _ := splitKV(raw)
			cat := &retab.Category{Name: name, Description: ptr(desc)}
			if err := validateClassificationCategory(i, cat, "--category"); err != nil {
				return err
			}
			categories = append(categories, cat)
		}
		if len(categories) == 0 {
			return fmt.Errorf("at least one category is required (--category or --categories-file)")
		}
		doc, err := resolveDocument(cmd)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()

		nConsensus, _ := cmd.Flags().GetInt("n-consensus")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		firstN, _ := cmd.Flags().GetInt("first-n-pages")
		instructions, _ := cmd.Flags().GetString("instructions")
		result, err := client.Classifications.Create(ctx, &retab.ClassificationsCreateParams{
			Document:     doc,
			Categories:   categories,
			Model:        ptr(model),
			NConsensus:   ptr(nConsensus),
			BustCache:    ptr(bustCache),
			FirstNPages:  ptr(firstN),
			Instructions: ptr(instructions),
		})
		if err != nil {
			return err
		}
		return maybeWaitForPrimitiveCreate(cmd, classificationWaitSpec, result)
	}),
}

func validateClassificationCategory(index int, category *retab.Category, source string) error {
	if strings.TrimSpace(category.Name) == "" {
		return fmt.Errorf("%s[%d].name is required", source, index)
	}
	return nil
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
		result, err := client.Classifications.Get(ctx, args[0], nil)
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
		params := retab.ClassificationsListParams{PaginationParams: collectListParams(cmd)}
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
backup with ` + "`retab classifications get`" + ` first if you may need it.

Pass ` + "`--yes`" + ` to skip the confirmation prompt in scripts and CI —
otherwise the command refuses to delete when stdin is not a terminal.`,
	Example: `  # Back up, then delete (interactive)
  retab classifications get clas_xyz789 > backup.json
  retab classifications delete clas_xyz789

  # Delete in a script
  retab classifications delete clas_xyz789 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "classification", args[0]); err != nil {
			return err
		}
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

var classificationsCancelCmd = &cobra.Command{
	Use:   "cancel <classification-id>",
	Short: "Cancel a classification",
	Long: `Cancel a pending or in-flight classification. Completed
classifications cannot be cancelled and the API returns an error.`,
	Example: `  # Cancel a running classification
  retab classifications cancel clas_xyz789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Classifications.CreateCancel(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	addDocumentFlags(classificationsCreateCmd)
	classificationsCreateCmd.Flags().String("model", "", "model identifier (required)")
	classificationsCreateCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 16}, "n-consensus", "consensus count (1-16)")
	classificationsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	classificationsCreateCmd.Flags().Var(&positiveIntFlagValue{}, "first-n-pages", "only classify the first N pages")
	classificationsCreateCmd.Flags().String("instructions", "", "extra instructions")
	classificationsCreateCmd.Flags().StringArray("category", nil, "category as name=description (repeatable)")
	classificationsCreateCmd.Flags().String("categories-file", "", "JSON array of {name, description} (or - for stdin)")
	addPrimitiveCreateWaitFlags(classificationsCreateCmd)
	_ = classificationsCreateCmd.MarkFlagRequired("model")

	addListFlags(classificationsListCmd, false)

	classificationsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	classificationsWaitCmd := primitiveWaitCommand(classificationWaitSpec)
	addPrimitiveWaitTuningFlags(classificationsWaitCmd, false)

	classificationsCmd.AddCommand(classificationsCreateCmd, classificationsGetCmd, classificationsListCmd, classificationsCancelCmd, classificationsDeleteCmd, classificationsWaitCmd)
	rootCmd.AddCommand(classificationsCmd)
}
