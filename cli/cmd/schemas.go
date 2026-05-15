package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var schemasCmd = &cobra.Command{
	Use:   "schemas",
	Short: "Generate JSON schemas",
	Long: `Generate JSON Schemas from sample documents.

Point ` + "`schemas generate`" + ` at one or more representative documents and
the API returns a JSON Schema describing the fields a model should
extract — a useful starting point when you don't already have a schema
written by hand. Save the result to a file and pass it to
` + "`retab extractions create --json-schema-file`" + ` to actually run the
extraction.`,
	Example: `  # Generate a schema from a single sample document
  retab schemas generate --file ./invoice.pdf > schema.json

  # Then use it in an extraction
  retab extractions create \
    --file ./invoice.pdf \
    --json-schema-file ./schema.json \
    --model gpt-4o`,
}

var schemasGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a JSON schema from one or more documents",
	Long: `Infer a JSON Schema from one or more sample documents.

Provide documents in any combination of:
  --file <path>        local file (repeatable)
  --url <url>          publicly fetchable URL (repeatable)
  --file-id <id>       already-uploaded Retab file id (repeatable)
  --documents-file     JSON array of document descriptors (or - for stdin)

At least one document is required. The more representative samples you
pass, the more general the resulting schema. The output is a JSON Schema
suitable for ` + "`retab extractions create --json-schema-file`" + ` — save it,
review it, edit it by hand if you want tighter typing, and commit it
alongside your code. Pass ` + "`--format json`" + ` to preserve the full server
response envelope (` + "`json_schema`" + `, ` + "`created_at`" + `, etc.).`,
	Example: `  # Single sample -> schema on stdout
  retab schemas generate --file ./invoice.pdf > schema.json

  # Multiple samples for a more general schema
  retab schemas generate \
    --file ./invoices/inv1.pdf \
    --file ./invoices/inv2.pdf \
    --file ./invoices/inv3.pdf \
    --model gpt-4o > schema.json

  # Mix uploaded ids and local files
  retab schemas generate --file-id file_abc123 --file ./extra.pdf`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		if format != "" && format != "schema" && format != "json" {
			return fmt.Errorf("invalid --format value %q (want: schema | json)", format)
		}
		documents := []any{}
		files, _ := cmd.Flags().GetStringArray("file")
		urls, _ := cmd.Flags().GetStringArray("url")
		fileIDs, _ := cmd.Flags().GetStringArray("file-id")
		docsFile, _ := cmd.Flags().GetString("documents-file")

		if docsFile != "" {
			arr, err := readJSONArray(docsFile)
			if err != nil {
				return fmt.Errorf("--documents-file: %w", err)
			}
			documents = append(documents, arr...)
		}
		for _, path := range files {
			mime, err := inferFileMIMEData(path)
			if err != nil {
				return err
			}
			documents = append(documents, mime)
		}
		for _, u := range urls {
			if strings.TrimSpace(u) == "" {
				return fmt.Errorf("--url must not be blank")
			}
			// Server requires `filename` on every doc descriptor — derive
			// from the URL path, same shape as resolveDocument does in
			// common.go for the single-document commands.
			documents = append(documents, retab.MIMEData{
				Filename: filenameFromURL(u),
				URL:      u,
			})
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		for _, id := range fileIDs {
			// `FileRef{ID: ...}` alone returns 422; the server demands
			// filename + url. Look the file up via Files.GetDownloadLink
			// to fill both. One extra GET per id — fine for the UX win.
			link, err := client.Files.GetDownloadLink(ctx, id)
			if err != nil {
				return fmt.Errorf("--file-id %s: %w", id, err)
			}
			if link.DownloadURL == "" {
				return fmt.Errorf("--file-id %s: server returned no download URL", id)
			}
			filename := link.Filename
			if filename == "" {
				filename = "document"
			}
			documents = append(documents, retab.MIMEData{
				Filename: filename,
				URL:      link.DownloadURL,
			})
		}
		if len(documents) == 0 {
			return fmt.Errorf("at least one document is required (--file, --url, --file-id, or --documents-file)")
		}

		model, _ := cmd.Flags().GetString("model")
		result, err := client.Schemas.Generate(ctx, retab.GenerateSchemaRequest{
			Documents: documents,
			Model:     model,
		})
		if err != nil {
			return err
		}
		return writeGeneratedSchema(cmd.OutOrStdout(), result, format)
	}),
}

func writeGeneratedSchema(w io.Writer, result *retab.Resource, format string) error {
	switch format {
	case "", "schema":
		if result == nil {
			return fmt.Errorf("server response missing json_schema field")
		}
		schema, ok := (*result)["json_schema"]
		if !ok || schema == nil {
			return fmt.Errorf("server response missing json_schema field")
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(schema)
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(result)
	default:
		return fmt.Errorf("invalid --format value %q (want: schema | json)", format)
	}
}

func init() {
	schemasGenerateCmd.Flags().StringArray("file", nil, "path to a document (repeatable)")
	schemasGenerateCmd.Flags().StringArray("url", nil, "document URL (repeatable)")
	schemasGenerateCmd.Flags().StringArray("file-id", nil, "Retab file id (repeatable)")
	schemasGenerateCmd.Flags().String("documents-file", "", "JSON array of documents (or - for stdin)")
	schemasGenerateCmd.Flags().String("model", "", "model identifier")
	schemasGenerateCmd.Flags().String("format", "schema", "output format: schema | json")

	schemasCmd.AddCommand(schemasGenerateCmd)
	rootCmd.AddCommand(schemasCmd)
}
