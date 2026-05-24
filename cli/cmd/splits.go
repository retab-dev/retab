//go:build !retab_oagen_cli_splits

package cmd

import (
	"fmt"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var splitsCmd = &cobra.Command{
	Use:   "splits",
	Short: "Intelligently split documents into logical sections",
	Long: `Split a multi-section document into its constituent subdocuments.

Useful when a single PDF actually contains several logical documents — for
example invoice + packing slip + terms-and-conditions concatenated into
one file. You describe the expected subdocument types up front and the
service identifies their boundaries. Downstream, you can run a separate
` + "`retab extractions create`" + ` per section with a section-specific schema.`,
}

var splitsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a split",
	Long: `Split a document into logical subdocuments.

` + "`--subdocuments-file`" + ` is a JSON array (or ` + "`-`" + ` for stdin) of
expected sections, each with ` + "`name`" + ` and ` + "`description`" + ` (plus
optional ` + "`allow_multiple_instances`" + ` for sections that may appear more
than once, e.g. multiple invoices in a single file).

Boost accuracy on ambiguous boundaries with ` + "`--n-consensus`" + `. The
returned subdocument references can be fed back into per-section
` + "`retab extractions create`" + ` runs.`,
	Example: `  # Split a multi-section PDF into named pieces
  retab splits create \
    --file ./bundle.pdf --model gpt-4o \
    --subdocuments-file ./sections.json

  # Allow multiple invoices in one file (set in JSON)
  cat <<'JSON' > sections.json
  [
    {"name": "invoice",      "description": "vendor bill", "allow_multiple_instances": true},
    {"name": "packing_slip", "description": "what shipped"}
  ]
  JSON
  retab splits create \
    --file-id file_abc123 --model gpt-4o \
    --subdocuments-file ./sections.json --n-consensus 3`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		model, err := requireNonBlankFlag(cmd, "model")
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
		if len(arr) == 0 {
			return fmt.Errorf("--subdocuments-file: at least one subdocument is required")
		}
		var subs []*retab.Subdocument
		for i, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("--subdocuments-file[%d] must be a JSON object", i)
			}
			sub := &retab.Subdocument{}
			nameValue, ok := obj["name"]
			if !ok {
				return fmt.Errorf("--subdocuments-file[%d].name is required", i)
			}
			name, ok := nameValue.(string)
			if !ok {
				return fmt.Errorf("--subdocuments-file[%d].name must be a string", i)
			}
			sub.Name = name
			if v, ok := obj["description"]; ok {
				description, ok := v.(string)
				if !ok {
					return fmt.Errorf("--subdocuments-file[%d].description must be a string", i)
				}
				sub.Description = ptr(description)
			}
			if v, ok := obj["allow_multiple_instances"]; ok {
				allowMultipleInstances, ok := v.(bool)
				if !ok {
					return fmt.Errorf("--subdocuments-file[%d].allow_multiple_instances must be a boolean", i)
				}
				sub.AllowMultipleInstances = ptr(allowMultipleInstances)
			}
			if err := validateSplitSubdocument(i, sub); err != nil {
				return err
			}
			subs = append(subs, sub)
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
		instructions, _ := cmd.Flags().GetString("instructions")
		result, err := client.Splits.Create(ctx, &retab.SplitsCreateParams{
			Document:     doc,
			Subdocuments: subs,
			Model:        ptr(model),
			NConsensus:   ptr(nConsensus),
			BustCache:    ptr(bustCache),
			Instructions: ptr(instructions),
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func validateSplitSubdocument(index int, subdocument *retab.Subdocument) error {
	if strings.TrimSpace(subdocument.Name) == "" {
		return fmt.Errorf("--subdocuments-file[%d].name is required", index)
	}
	return nil
}

var splitsGetCmd = &cobra.Command{
	Use:   "get <split-id>",
	Short: "Get a split by id",
	Long: `Fetch a single split by id.

Returns the source document reference, the requested subdocument
definitions, and the resolved subdocument page ranges.`,
	Example: `  # Fetch a known split
  retab splits get split_xyz789

  # List the page ranges of each resolved subdocument
  retab splits get split_xyz789 | jq '.subdocuments[] | {name, pages}'`,
	Args: cobra.ExactArgs(1),
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
	Long: `List splits, newest first by default.

Page by split id with ` + "`--before`" + ` / ` + "`--after`" + `, cap page size with
` + "`--limit`" + `.`,
	Example: `  # Most recent 25 splits
  retab splits list --limit 25

  # Walk pages from a known id
  retab splits list --after split_xyz789 --limit 50`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.SplitsListParams{PaginationParams: collectListParams(cmd)}
		result, err := client.Splits.List(ctx, &params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var splitsDeleteCmd = &cobra.Command{
	Use:   "delete <split-id>",
	Short: "Delete a split",
	Long: `Permanently delete a split.

Destructive and irreversible. The source document is not affected. Take a
backup with ` + "`retab splits get`" + ` first if you may need the subdocument
boundaries again.

Pass ` + "`--yes`" + ` to skip the confirmation prompt in scripts and CI —
otherwise the command refuses to delete when stdin is not a terminal.`,
	Example: `  # Back up, then delete
  retab splits get split_xyz789 > backup.json
  retab splits delete split_xyz789

  # Skip the prompt in scripts
  retab splits delete split_xyz789 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "split", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Splits.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("split", args[0])
		return nil
	}),
}

func init() {
	addDocumentFlags(splitsCreateCmd)
	splitsCreateCmd.Flags().String("subdocuments-file", "", "JSON array of subdocuments (name, description, allow_multiple_instances) (required)")
	splitsCreateCmd.Flags().String("model", "", "model identifier (required)")
	splitsCreateCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 8}, "n-consensus", "consensus count (1-8)")
	splitsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	splitsCreateCmd.Flags().String("instructions", "", "extra instructions")
	_ = splitsCreateCmd.MarkFlagRequired("model")
	_ = splitsCreateCmd.MarkFlagRequired("subdocuments-file")

	addListFlags(splitsListCmd, false)

	splitsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	splitsCmd.AddCommand(splitsCreateCmd, splitsGetCmd, splitsListCmd, splitsDeleteCmd)
	rootCmd.AddCommand(splitsCmd)
}
