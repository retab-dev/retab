package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// fileDownloadClient is used for direct downloads from signed storage URLs.
// The timeout is generous (10 min for large files) but bounded so a wedged
// server doesn't hang the CLI in unattended scripts. Ctrl-C still works via
// the request context.
var fileDownloadClient = &http.Client{
	Timeout: 10 * time.Minute,
}

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage files",
	Long: `Manage files in your Retab workspace.

Files are document blobs (PDFs, images, emails, spreadsheets, etc.)
referenced by id across the API. Upload once via ` + "`files upload`" + `, then
pass ` + "`--file-id`" + ` everywhere a document is required (extractions,
edits, schemas, workflow runs) so the same blob isn't re-uploaded on
every call.

For files larger than the synchronous upload path allows, use the
two-phase ` + "`create-upload`" + ` -> direct PUT -> ` + "`complete-upload`" + ` flow.`,
	Example: `  # List the five most recent files
  retab files list --limit 5

  # Upload a local PDF and capture the id for reuse
  FILE_ID=$(retab files upload ./invoice.pdf | jq -r .id)

  # Reuse the id in an extraction (no re-upload)
  retab extractions create \
    --file-id $FILE_ID \
    --json-schema-file ./schema.json \
    --model gpt-4o

  # Pull the file back down to a local path
  retab files download $FILE_ID -o ./invoice.pdf`,
}

var filesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List files",
	Long: `List files in the workspace, newest first.

Output is JSON: a top-level array (or paginated wrapper, depending on the
server) with one entry per file. Use --limit, --mime-type, and --sort-by
to narrow the view. Pipe through ` + "`jq`" + ` to project just the fields you
need.`,
	Example: `  # Five most recent files
  retab files list --limit 5

  # Just PDFs, newest first
  retab files list --mime-type application/pdf

  # Extract ids only, one per line
  retab files list | jq -r '.data[].id'`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListFilesParams{ListParams: collectListParams(cmd)}
		if v, _ := cmd.Flags().GetString("mime-type"); v != "" {
			params.MIMEType = v
		}
		if v, _ := cmd.Flags().GetString("sort-by"); v != "" {
			params.SortBy = v
		}
		result, err := client.Files.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesGetCmd = &cobra.Command{
	Use:   "get <file-id>",
	Short: "Get a file by id",
	Long: `Fetch the metadata record for a single file as JSON.

Returns the file's filename, MIME type, size, sha256, created_at, and any
other workspace-level fields — but NOT the file bytes. Use
` + "`retab files download`" + ` to retrieve the content, or
` + "`retab files download-link`" + ` to get a signed URL.`,
	Example: `  # Inspect a known file
  retab files get file_abc123

  # Project just the fields you need
  retab files get file_abc123 | jq '{id, filename, size_bytes}'`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesUploadCmd = &cobra.Command{
	Use:   "upload <path>",
	Short: "Upload a local file",
	Long: `Upload a local file to Retab and receive a file id.

The id can be passed as --file-id to extractions, edits, schemas, and
workflow runs in lieu of re-uploading the same blob on every call. The
upload is synchronous; for very large files use ` + "`files create-upload`" + `
plus ` + "`files complete-upload`" + ` to upload directly to storage.`,
	Example: `  # Upload and capture the id for reuse
  FILE_ID=$(retab files upload ./invoice.pdf | jq -r .id)

  # Upload, then immediately run an extraction against the id
  retab files upload ./invoice.pdf | jq -r .id | xargs -I{} \
    retab extractions create --file-id {} \
      --json-schema-file ./schema.json --model gpt-4o`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.Upload(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesDownloadLinkCmd = &cobra.Command{
	Use:   "download-link <file-id>",
	Short: "Get a download link for a file",
	Long: `Mint a short-lived signed URL that streams the file's bytes from
storage. Useful when handing the URL to another tool (a browser, curl,
a worker in another process) instead of pulling the bytes through this
CLI. The link expires; mint a new one if it's stale.

For the common case of "save this file to disk" use ` + "`retab files download`" + `
which does the GET for you.`,
	Example: `  # Print the signed URL
  retab files download-link file_abc123 | jq -r .download_url

  # Pipe straight into curl
  curl -o invoice.pdf "$(retab files download-link file_abc123 | jq -r .download_url)"`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.GetDownloadLink(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesDownloadCmd = &cobra.Command{
	Use:   "download <file-id>",
	Short: "Download a file to a local path (- for stdout)",
	Long: `Stream a file's bytes from storage to disk (or stdout).

With no -o flag, writes to the server-recorded filename in the current
directory, falling back to the file id if the server didn't store a
name. Pass ` + "`-o <path>`" + ` for an explicit destination, or ` + "`-o -`" + ` to
write to stdout so the output can be piped into another tool.

Downloads stream chunk-by-chunk and propagate Ctrl-C, so canceling a
slow transfer leaves no half-written file open on stdout (and partial
local files can be safely deleted and retried).`,
	Example: `  # Save under the server's filename
  retab files download file_abc123

  # Save to an explicit path
  retab files download file_abc123 -o ./invoice.pdf

  # Pipe to stdout (e.g. into another tool)
  retab files download file_abc123 -o - | pdftotext - -`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		link, err := client.Files.GetDownloadLink(ctx, args[0])
		if err != nil {
			return err
		}
		out, _ := cmd.Flags().GetString("output")
		if out == "" {
			out = link.Filename
		}
		if out == "" {
			out = args[0]
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, link.DownloadURL, nil)
		if err != nil {
			return err
		}
		resp, err := fileDownloadClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("download failed: %d %s", resp.StatusCode, string(body))
		}
		var sink io.Writer
		if out == "-" {
			sink = os.Stdout
		} else {
			f, err := os.Create(out)
			if err != nil {
				return err
			}
			defer f.Close()
			sink = f
		}
		_, err = io.Copy(sink, resp.Body)
		return err
	}),
}

var filesCreateUploadCmd = &cobra.Command{
	Use:   "create-upload",
	Short: "Reserve a direct-to-storage upload session",
	Long: `Phase 1 of the two-phase upload flow for large files.

Reserves a file id on the server and returns a signed PUT URL that
accepts the bytes directly into object storage — bypassing the
synchronous upload path. Use this when ` + "`files upload`" + ` is too slow or
hits a size limit. The returned ` + "`upload_url`" + ` accepts a single PUT;
after that succeeds, call ` + "`retab files complete-upload <file-id>`" + ` to
mark the upload as finalized.

Steps: (1) ` + "`create-upload`" + ` returns ` + "`{id, upload_url, ...}`" + `;
(2) PUT the bytes to ` + "`upload_url`" + ` with the Content-Type you declared;
(3) call ` + "`complete-upload`" + ` with the file id to commit the upload.`,
	Example: `  # Three-step large-file upload
  RESP=$(retab files create-upload \
    --filename big.pdf \
    --content-type application/pdf \
    --size-bytes 1234567890)
  FILE_ID=$(echo "$RESP" | jq -r .id)
  URL=$(echo "$RESP" | jq -r .upload_url)

  curl -X PUT --upload-file ./big.pdf -H "Content-Type: application/pdf" "$URL"
  retab files complete-upload $FILE_ID`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		filename, _ := cmd.Flags().GetString("filename")
		contentType, _ := cmd.Flags().GetString("content-type")
		sizeStr, _ := cmd.Flags().GetString("size-bytes")
		sha256Hash, _ := cmd.Flags().GetString("sha256")
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid --size-bytes: %w", err)
		}
		result, err := client.Files.CreateUpload(ctx, retab.PrepareUploadRequest{
			Filename:    filename,
			ContentType: contentType,
			SizeBytes:   size,
			SHA256:      sha256Hash,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesCompleteUploadCmd = &cobra.Command{
	Use:   "complete-upload <file-id>",
	Short: "Mark a direct upload as finished",
	Long: `Phase 2 of the two-phase upload flow for large files.

After ` + "`retab files create-upload`" + ` reserved an id and you PUT the bytes
to the returned ` + "`upload_url`" + `, run this to commit the upload — the file
won't be usable in extractions or workflow runs until it's marked
complete. Optionally pass --sha256 to have the server verify integrity
against the digest you computed locally.`,
	Example: `  # Commit after a direct PUT
  retab files complete-upload file_abc123

  # Same, with server-side integrity check
  retab files complete-upload file_abc123 \
    --sha256 e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		sha256Hash, _ := cmd.Flags().GetString("sha256")
		result, err := client.Files.CompleteUpload(ctx, args[0], sha256Hash)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	addListFlags(filesListCmd, false)
	filesListCmd.Flags().String("mime-type", "", "filter by MIME type")
	filesListCmd.Flags().String("sort-by", "", "sort field (default: created_at)")

	filesDownloadCmd.Flags().StringP("output", "o", "", "output path (default: server filename, - for stdout)")

	filesCreateUploadCmd.Flags().String("filename", "", "filename (required)")
	filesCreateUploadCmd.Flags().String("content-type", "", "content type (required)")
	filesCreateUploadCmd.Flags().String("size-bytes", "0", "file size in bytes (required)")
	filesCreateUploadCmd.Flags().String("sha256", "", "sha256 hex digest (optional)")
	_ = filesCreateUploadCmd.MarkFlagRequired("filename")
	_ = filesCreateUploadCmd.MarkFlagRequired("content-type")
	_ = filesCreateUploadCmd.MarkFlagRequired("size-bytes")

	filesCompleteUploadCmd.Flags().String("sha256", "", "sha256 hex digest (optional)")

	filesCmd.AddCommand(filesListCmd, filesGetCmd, filesUploadCmd, filesDownloadLinkCmd, filesDownloadCmd, filesCreateUploadCmd, filesCompleteUploadCmd)
	rootCmd.AddCommand(filesCmd)
}
