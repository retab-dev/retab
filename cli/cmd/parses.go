//go:build !retab_oagen_cli_parses

package cmd

import (
	"fmt"

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

var allowedTableParsingFormats = map[string]bool{
	"markdown": true,
	"yaml":     true,
	"html":     true,
	"json":     true,
}

func validateTableParsingFormat(value string) error {
	if value == "" || allowedTableParsingFormats[value] {
		return nil
	}
	return fmt.Errorf("invalid --table-parsing-format %q (want: markdown | yaml | html | json)", value)
}

var parsesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a parse",
	Long: `Parse a document into LLM-ready markdown.

Accepts the standard document sources (` + "`--file`" + `, ` + "`--url`" + `,
` + "`--file-id`" + `, ` + "`--document-file`" + `). Tune table rendering with
` + "`--table-parsing-format`" + ` (e.g. ` + "`markdown`" + `, ` + "`html`" + `).`,
	Example: `  # Parse a PDF to markdown
  retab parses create --file ./report.pdf --model gpt-4o

  # Parse an Excel file with HTML tables for downstream rendering
  retab parses create \
    --file ./book.xlsx --model gpt-4o \
    --table-parsing-format html

  # Parse a scanned image
  retab parses create --file ./scan.png --model gpt-4o`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		tableFormat, _ := cmd.Flags().GetString("table-parsing-format")
		if err := validateTableParsingFormat(tableFormat); err != nil {
			return err
		}
		model, err := requireNonBlankFlag(cmd, "model")
		if err != nil {
			return err
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
		instructions, _ := cmd.Flags().GetString("instructions")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		params := &retab.ParsesCreateParams{
			Document:     doc,
			Model:        ptr(model),
			Instructions: ptr(instructions),
			BustCache:    ptr(bustCache),
			Background:   primitiveBackgroundParam(cmd),
		}
		if tableFormat != "" {
			tpf := retab.ParseRequestTableParsingFormat(tableFormat)
			params.TableParsingFormat = &tpf
		}
		result, err := client.Parses.Create(ctx, params)
		if err != nil {
			return err
		}
		return maybeWaitForPrimitiveCreate(cmd, parseWaitSpec, result)
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
		result, err := client.Parses.Get(ctx, args[0], &retab.ParsesGetParams{})
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

Page by parse id with ` + "`--before`" + ` / ` + "`--after`" + `, cap page size with
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
		params := retab.ParsesListParams{PaginationParams: collectListParams(cmd)}
		params.Filename, params.FromDate, params.ToDate = collectFileDateListFilters(cmd)
		result, err := client.Parses.List(ctx, &params)
		if err != nil {
			return err
		}
		return printPrimitiveListResult(cmd, result)
	}),
}

var parsesCancelCmd = &cobra.Command{
	Use:   "cancel <parse-id>",
	Short: "Cancel a parse",
	Long: `Cancel a pending or in-flight parse. Completed parses cannot be
cancelled and the API returns an error.`,
	Example: `  # Cancel a running parse
  retab parses cancel parse_xyz789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Parses.Cancel(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	addDocumentFlags(parsesCreateCmd)
	parsesCreateCmd.Flags().String("model", "", "model identifier (required)")
	parsesCreateCmd.Flags().String("table-parsing-format", "", "table parsing format")
	parsesCreateCmd.Flags().String("image-resolution-dpi", "", "ignored legacy image resolution DPI")
	_ = parsesCreateCmd.Flags().MarkHidden("image-resolution-dpi")
	parsesCreateCmd.Flags().String("instructions", "", "extra instructions")
	parsesCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	addPrimitiveBackgroundFlag(parsesCreateCmd)
	addPrimitiveCreateWaitFlags(parsesCreateCmd)
	_ = parsesCreateCmd.MarkFlagRequired("model")

	addListFlags(parsesListCmd, false)

	parsesWaitCmd := primitiveWaitCommand(parseWaitSpec)
	addPrimitiveWaitTuningFlags(parsesWaitCmd, false)

	parsesCmd.AddCommand(parsesCreateCmd, parsesGetCmd, parsesListCmd, parsesCancelCmd, parsesWaitCmd)
	rootCmd.AddCommand(parsesCmd)
}
