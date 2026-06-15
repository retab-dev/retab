//go:build !retab_oagen_cli_schemas

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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
alongside your code.

Steer the result with ` + "`--instructions`" + ` (or ` + "`--instructions-file`" + ` /
` + "`-`" + ` for stdin) to constrain scope, nullability, naming, types, etc.
Lower ` + "`--image-resolution-dpi`" + ` to reduce per-page work — useful when a
long or heavily-instructed generation would otherwise exceed the server's
processing budget.

Generation is served by an internal service that can intermittently exceed
its own deadline (a transient "context deadline exceeded" failure). The CLI
resubmits automatically — tune with ` + "`--max-retries`" + ` /
` + "`--retry-delay-ms`" + `. Each attempt waits up to ` + "`--timeout-seconds`" + `,
so pair a shorter timeout with more retries to abandon a stuck job and
resubmit sooner.

` + "`--format`" + ` and ` + "`--output`" + ` look similar but control different things:

  --format chooses WHICH payload to return.
    --format schema (default) — just the JSON Schema body, the shape
      you'd paste into ` + "`--json-schema-file`" + `.
    --format json — the full server response envelope (` + "`json_schema`" + `,
      ` + "`created_at`" + `, etc.), useful when you want sidecar metadata.

  --output chooses HOW to print the selected payload. It's the global
  flag every other command honours: ` + "`--output json`" + ` always renders
  JSON, ` + "`--output table`" + ` renders a table when the payload is
  tabulable, and the default auto-detects based on stdout.`,
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
		documents := []retab.MIMEData{}
		files, _ := cmd.Flags().GetStringArray("file")
		urls, _ := cmd.Flags().GetStringArray("url")
		fileIDs, _ := cmd.Flags().GetStringArray("file-id")
		docsFile, _ := cmd.Flags().GetString("documents-file")

		if docsFile != "" {
			arr, err := readJSONArray(docsFile)
			if err != nil {
				return fmt.Errorf("--documents-file: %w", err)
			}
			for _, raw := range arr {
				document, err := mimeDataFromDocument(raw)
				if err != nil {
					return fmt.Errorf("--documents-file: %w", err)
				}
				documents = append(documents, document)
			}
		}
		for _, path := range files {
			mime, err := inferFileMIMEData(path)
			if err != nil {
				return err
			}
			documents = append(documents, retab.MIMEData{
				Filename: mime.Filename,
				URL:      mime.URL,
			})
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

		params := &retab.SchemasGenerateParams{
			Documents: documents,
		}
		if model, _ := cmd.Flags().GetString("model"); model != "" {
			params.Model = ptr(model)
		}
		// Natural-language guidance for the generator. Accept it inline
		// (--instructions) or from a file/stdin (--instructions-file), since
		// useful prompts are often long and awkward as a shell argument. The
		// two are mutually exclusive to avoid an ambiguous precedence rule.
		instr, _ := cmd.Flags().GetString("instructions")
		instrFile, _ := cmd.Flags().GetString("instructions-file")
		if instr != "" && instrFile != "" {
			return fmt.Errorf("--instructions and --instructions-file are mutually exclusive")
		}
		if instrFile != "" {
			text, err := readTextFileOrStdin(instrFile)
			if err != nil {
				return fmt.Errorf("--instructions-file: %w", err)
			}
			instr = text
		}
		if instr != "" {
			params.Instructions = ptr(instr)
		}
		// Render DPI sent to the model. Lower values reduce the work the
		// generator does per page, which is the practical lever for keeping a
		// long/instructed generation within the server's processing budget.
		if dpi, _ := cmd.Flags().GetInt("image-resolution-dpi"); dpi > 0 {
			params.ImageResolutionDpi = ptr(dpi)
		}
		background, _ := cmd.Flags().GetBool("background")
		if background {
			params.Background = ptr(true)
		}
		// Schema generation is served by an internal orchestrator that
		// intermittently exceeds its own deadline (surfaced as a "context
		// deadline exceeded" terminal failure) — especially for instructed or
		// multi-page jobs. A resubmit usually lands, so retry the whole
		// generation on transient server failures. Non-transient failures
		// (validation, auth, a genuine schema error) are returned immediately.
		maxRetries, _ := cmd.Flags().GetInt("max-retries")
		if maxRetries < 0 {
			maxRetries = 0
		}
		retryDelay := 3 * time.Second
		if ms, _ := cmd.Flags().GetInt("retry-delay-ms"); ms > 0 {
			retryDelay = time.Duration(ms) * time.Millisecond
		}
		wait, _ := cmd.Flags().GetBool("wait")

		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"schema generation: transient server failure, retrying (attempt %d/%d)...\n",
					attempt+1, maxRetries+1)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryDelay):
				}
			}

			result, err := client.Schemas.Generate(ctx, params)
			if err != nil {
				lastErr = err
				if isTransientGenFailure(err.Error()) {
					continue
				}
				return err
			}

			// Synchronous (default): the response already carries the inferred
			// schema — unless the server recorded a terminal failure inline.
			if !background {
				m, _ := primitiveMap(result)
				if termErr := primitiveTerminalError(schemaGenerationWaitSpec, m); termErr != nil {
					lastErr = termErr
					if isTransientGenFailure(genFailureDetail(m)) {
						continue
					}
					return termErr
				}
				return writeGeneratedSchema(cmd.OutOrStdout(), result, format)
			}

			// Background without --wait: nothing to retry on yet — the outcome
			// is unknown. Print the queued record (id + status) for the caller
			// to poll.
			if !wait {
				return printJSON(result)
			}

			// Background with --wait: poll to a terminal status, then resubmit
			// the whole job if it failed transiently.
			pollInterval, timeout := primitiveWaitDurations(cmd)
			final, waitErr := waitForPrimitive(ctx, cmd, schemaGenerationWaitSpec, result.ID, pollInterval, timeout)
			if waitErr != nil {
				lastErr = waitErr
				if isTransientGenFailure(waitErr.Error()) {
					continue
				}
				if final != nil {
					_ = printJSON(final)
				}
				return waitErr
			}
			if termErr := primitiveTerminalError(schemaGenerationWaitSpec, final); termErr != nil {
				lastErr = termErr
				if isTransientGenFailure(genFailureDetail(final)) {
					continue
				}
				_ = printJSON(final)
				return termErr
			}
			return writeSchemaFromMap(cmd.OutOrStdout(), final, format)
		}
		return fmt.Errorf("schema generation failed after %d attempt(s); last error: %w", maxRetries+1, lastErr)
	}),
}

// genFailureDetail pulls the server-recorded error.message out of a terminal
// generation record, so transient-failure detection can inspect the real cause
// (primitiveTerminalError only reports the status, not the detail).
func genFailureDetail(rec map[string]any) string {
	if rec == nil {
		return ""
	}
	e, ok := rec["error"].(map[string]any)
	if !ok {
		return ""
	}
	msg, _ := e["message"].(string)
	return msg
}

// isTransientGenFailure reports whether a schema-generation failure looks
// retryable. The dominant case is the backend's own "context deadline
// exceeded" when its orchestrator service is slow; the rest are the usual
// transient HTTP/network signals. These are resubmitted rather than surfaced
// as a one-off hiccup. An empty detail is treated as non-transient so genuine
// (undescribed) failures are not retried blindly.
func isTransientGenFailure(detail string) bool {
	d := strings.ToLower(detail)
	if d == "" {
		return false
	}
	for _, needle := range []string{
		"context deadline exceeded",
		"deadline exceeded",
		"timeout",
		"timed out",
		"temporarily unavailable",
		"service unavailable",
		"connection reset",
		"connection refused",
		"502", "503", "504",
		"eof",
	} {
		if strings.Contains(d, needle) {
			return true
		}
	}
	return false
}

var schemasGetCmd = &cobra.Command{
	Use:   "get <schema-generation-id>",
	Short: "Get a schema generation by id",
	Long: `Fetch a background schema generation by its id.

Created with ` + "`schemas generate --background`" + `, a generation runs
asynchronously (pending -> queued -> in_progress -> completed | failed |
cancelled). This prints the full record — including ` + "`status`" + ` and,
once complete, ` + "`json_schema`" + `. Use ` + "`schemas wait`" + ` to block
until it finishes.`,
	Example: `  retab schemas get sch_abc123`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		var result retab.SchemaGeneration
		path := schemaGenerationWaitSpec.path + "/" + url.PathEscape(args[0])
		if err := cliJSONRequestInto(cmd, http.MethodGet, path, nil, nil, &result); err != nil {
			return err
		}
		return printJSON(&result)
	}),
}

var schemasCancelCmd = &cobra.Command{
	Use:   "cancel <schema-generation-id>",
	Short: "Cancel a background schema generation",
	Long: `Cancel an in-flight background schema generation.

A non-terminal generation (pending / queued / in_progress) transitions to
` + "`cancelled`" + `; a generation that already reached a terminal status is
returned unchanged.`,
	Example: `  retab schemas cancel sch_abc123`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		var result retab.SchemaGeneration
		path := schemaGenerationWaitSpec.path + "/" + url.PathEscape(args[0]) + "/cancel"
		if err := cliJSONRequestInto(cmd, http.MethodPost, path, nil, nil, &result); err != nil {
			return err
		}
		return printJSON(&result)
	}),
}

// writeSchemaFromMap re-marshals a polled generation record into a
// SchemaGeneration so --wait output goes through the same writer (and --format
// handling) as a synchronous generate.
func writeSchemaFromMap(w io.Writer, final map[string]any, format string) error {
	raw, err := json.Marshal(final)
	if err != nil {
		return fmt.Errorf("encode schema generation: %w", err)
	}
	var result retab.SchemaGeneration
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("decode schema generation: %w", err)
	}
	return writeGeneratedSchema(w, &result, format)
}

func writeGeneratedSchema(w io.Writer, result *retab.SchemaGeneration, format string) error {
	switch format {
	case "", "schema":
		if result == nil || result.JSONSchema == nil {
			return fmt.Errorf("server response missing json_schema field")
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(result.JSONSchema)
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
	schemasGenerateCmd.Flags().String("instructions", "", "natural-language guidance for the generated schema")
	schemasGenerateCmd.Flags().String("instructions-file", "", "read schema instructions from a file (or - for stdin)")
	schemasGenerateCmd.Flags().Int("image-resolution-dpi", 0, "render DPI sent to the model (lower = less work per page; helps long/instructed generations finish in time)")
	schemasGenerateCmd.Flags().Int("max-retries", 3, "resubmit the generation this many times on transient server failures (e.g. orchestrator deadline)")
	schemasGenerateCmd.Flags().Int("retry-delay-ms", 3000, "delay between retries in milliseconds")
	schemasGenerateCmd.Flags().String("format", "schema", "output format: schema | json")
	// Async surface, matching `extractions create` / `edits create`: --background
	// queues the generation server-side; --wait blocks until it finishes.
	addPrimitiveBackgroundFlag(schemasGenerateCmd)
	addPrimitiveCreateWaitFlags(schemasGenerateCmd)

	schemasCmd.AddCommand(schemasGenerateCmd)
	schemasCmd.AddCommand(schemasGetCmd)
	schemasCmd.AddCommand(schemasCancelCmd)
	schemasCmd.AddCommand(primitiveWaitCommand(schemaGenerationWaitSpec))
	rootCmd.AddCommand(schemasCmd)
}
