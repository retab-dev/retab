package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var parsesCmd = &cobra.Command{
	Use:   "parses",
	Short: "Convert any file (PDFs, Excel, emails, images) into LLM-ready markdown",
	Long: `Convert arbitrary file types into LLM-ready markdown.

A parse takes any supported input (PDF, Excel/CSV, .eml, image, etc.) and
returns a normalized markdown rendering with preserved structure — tables,
headings, lists, and image alt text. This is typically the first step in a
pipeline: parse → feed into prompts, or parse → extractions/classifications
when the downstream task wants normalized text rather than the raw file.`,
}

var parsesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a parse",
	Long: `Parse a document into LLM-ready markdown.

Accepts the standard document sources (` + "`--file`" + `, ` + "`--url`" + `,
` + "`--file-id`" + `, ` + "`--document-file`" + `). Tune table rendering with
` + "`--table-parsing-format`" + ` (e.g. ` + "`markdown`" + `, ` + "`html`" + `) and
raise ` + "`--image-resolution-dpi`" + ` for image-heavy or low-quality scans
where the default is too coarse.`,
	Example: `  # Parse a PDF to markdown
  retab parses create --file ./report.pdf --model gpt-4o

  # Parse an Excel file with HTML tables for downstream rendering
  retab parses create \
    --file ./book.xlsx --model gpt-4o \
    --table-parsing-format html

  # High-DPI parse for a scanned image
  retab parses create \
    --file ./scan.png --model gpt-4o \
    --image-resolution-dpi 300`,
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
	Long: `Fetch a single parse by id.

Returns the parse record including the rendered markdown, source document
reference, and per-page metadata.`,
	Example: `  # Print the rendered markdown
  retab parses get parse_xyz789 | jq -r '.markdown'

  # Save the whole parse record
  retab parses get parse_xyz789 > parse.json`,
	Args: cobra.ExactArgs(1),
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
	Long: `List parses, newest first by default.

Cursor-paginate with ` + "`--before`" + ` / ` + "`--after`" + `, cap page size with
` + "`--limit`" + `.`,
	Example: `  # Most recent 25 parses
  retab parses list --limit 25

  # Walk pages from a known id
  retab parses list --after parse_xyz789 --limit 50`,
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
	Long: `Permanently delete a parse.

Destructive and irreversible. The source document is not affected. Take a
backup with ` + "`retab parses get`" + ` first if you may need the markdown.`,
	Example: `  # Back up, then delete
  retab parses get parse_xyz789 > backup.json
  retab parses delete parse_xyz789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Parses.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("parse", args[0])
		return nil
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
