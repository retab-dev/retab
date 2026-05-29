package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// writeJSONTo encodes v as indented JSON to w with the same settings as
// printJSON (two-space indent, HTML escaping off, trailing newline), so
// output is identical whether it lands on stdout or a file.
func writeJSONTo(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

var filesParseCmd = &cobra.Command{
	Use:   "parse <path>",
	Short: "Parse a local document to text or structured JSON",
	Long: `Parse a local document into layout-preserved text or structured JSON,
entirely locally — no upload, no API call.

PDFs and images are parsed with LiteParse (the ` + "`lit`" + ` binary: PDFium
text extraction plus selective Tesseract OCR). text/csv/xlsx/docx are parsed
natively in Go, preserving exact line and cell structure.

With --format text (default) it prints the document's text. With
--format json it emits a normalized parse result: filename, mime_type,
document_type, source (pdf_text_layer | ocr | native), total_pages, and
per-page text. Add --bbox to include per-item bounding boxes for pdf/image.`,
	Example: `  # Print the text of a PDF
  retab files parse invoice.pdf

  # Structured JSON with positioned items
  retab files parse invoice.pdf --format json --bbox

  # Only the first five pages, no OCR
  retab files parse scan.pdf --pages 1-5 --no-ocr

  # A spreadsheet, parsed natively
  retab files parse data.xlsx --format json -o data.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		path := args[0]
		kind := detectKind(path)
		if kind == kindUnknown {
			return fmt.Errorf("unsupported file type for %s (supported: pdf, images, txt/md/json, csv/tsv, xlsx, docx)", path)
		}

		format, _ := cmd.Flags().GetString("format")
		switch format {
		case "", "text", "json":
		default:
			return fmt.Errorf("--format %q must be text or json", format)
		}
		withBbox, _ := cmd.Flags().GetBool("bbox")

		ctx, cancel := ctxFor(cmd)
		defer cancel()

		result, err := loadParse(ctx, path, kind, liteBinFromCmd(cmd), parseOptionsFromCmd(cmd), useCacheFromCmd(cmd))
		if err != nil {
			return err
		}

		out := os.Stdout
		outPath, _ := cmd.Flags().GetString("out")
		if outPath != "" {
			f, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("create %s: %w", outPath, err)
			}
			defer func() { _ = f.Close() }()
			out = f
		}

		if format == "json" {
			view := result
			if !withBbox {
				view = stripItems(result)
			}
			return writeJSONTo(out, view)
		}
		_, err = out.WriteString(parseResultText(result))
		return err
	}),
}

// stripItems returns a shallow copy of result with per-page Items dropped, so
// the default JSON view stays compact (the positioned items are large and only
// useful with --bbox).
func stripItems(result *ParseResult) *ParseResult {
	clone := *result
	clone.Pages = make([]ParsedPage, len(result.Pages))
	for i, p := range result.Pages {
		p.Items = nil
		clone.Pages[i] = p
	}
	return &clone
}

// parseResultText renders the text view: page texts joined by a blank line.
// A single page (text/csv/docx) prints just its text.
func parseResultText(result *ParseResult) string {
	if len(result.Pages) == 1 {
		return ensureTrailingNewline(result.Pages[0].Text)
	}
	var b strings.Builder
	for i, p := range result.Pages {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(p.Text)
		if !strings.HasSuffix(p.Text, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func ensureTrailingNewline(s string) string {
	if s == "" || strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}
