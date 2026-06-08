//go:build !retab_oagen_cli_files

package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
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
two-phase ` + "`create-upload`" + ` -> direct PUT -> ` + "`complete-upload`" + ` flow.

Local document tools do not require an API key and never upload data:

  parse              extract local text or structured JSON
  grep               search local PDFs, images, docs, and sheets
  inspect --render   render PDF/image pages to PNG files for visual review`,
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
  retab files download $FILE_ID -o ./invoice.pdf

  # Parse a local PDF without uploading it
  retab files parse ./invoice.pdf --format json --bbox

  # Render local PDF pages to PNGs that an agent or human can inspect
  retab files inspect ./statement.pdf --render 1-3 --out ./pages`,
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
		params := retab.FilesListParams{PaginationParams: collectListParams(cmd)}
		if v, _ := cmd.Flags().GetString("mime-type"); v != "" {
			params.MIMEType = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("sort-by"); v != "" {
			params.SortBy = ptr(v)
		}
		result, err := client.Files.List(ctx, &params)
		if err != nil {
			return err
		}
		sortBy := ""
		if params.SortBy != nil {
			sortBy = *params.SortBy
		}
		return printFilesListResult(cmd, result, sortBy)
	}),
}

func printFilesListResult(cmd *cobra.Command, result *retab.PaginatedList[retab.File], sortBy string) error {
	var raw string
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	if raw != string(OutputTable) && raw != string(OutputCSV) {
		return printJSON(result)
	}
	timestampHeader := "CREATED_AT"
	timestampField := "created_at"
	if sortBy == "updated_at" {
		timestampHeader = "UPDATED_AT"
		timestampField = "updated_at"
	}
	columns := []TableColumn{
		{Header: "ID", Extract: func(row any) string { return resourceCell(row, "id") }},
		{Header: "NAME", Extract: func(row any) string { return resourceCell(row, "filename") }},
		{Header: timestampHeader, Extract: func(row any) string { return normalizeTimestampCell(resourceCell(row, timestampField)) }},
	}
	rows := resourcesToRows(result.Data)
	if raw == string(OutputCSV) {
		return renderAutoCSV(os.Stdout, rows, columns)
	}
	return renderAutoTable(os.Stdout, rows, columns)
}

func resourcesToRows[T any](resources []T) []any {
	rows := make([]any, len(resources))
	for i, resource := range resources {
		rows[i] = resource
	}
	return rows
}

func resourceCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
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
	Short: "Upload a local file (or - to read from stdin)",
	Long: `Upload a local file to Retab and receive a file id.

The id can be passed as --file-id to extractions, edits, schemas, and
workflow runs in lieu of re-uploading the same blob on every call. The
upload is synchronous; for very large files use ` + "`files create-upload`" + `
plus ` + "`files complete-upload`" + ` to upload directly to storage.

Pass ` + "`-`" + ` as the positional path to read the file bytes from stdin.
` + "`--filename`" + ` is required in that mode so the server has a name to
store and a hint for content-type inference.`,
	Example: `  # Upload and capture the id for reuse
  FILE_ID=$(retab files upload ./invoice.pdf | jq -r .id)

  # Pipe bytes in from stdin (--filename required)
  cat invoice.pdf | retab files upload - --filename invoice.pdf

  # Upload, then immediately run an extraction against the id
  retab files upload ./invoice.pdf | jq -r .id | xargs -I{} \
    retab extractions create --file-id {} \
      --json-schema-file ./schema.json --model gpt-4o`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		uploadPath := args[0]
		if uploadPath == "-" {
			stdinPath, cleanup, err := stageStdinUpload(cmd)
			if err != nil {
				return err
			}
			defer cleanup()
			uploadPath = stdinPath
		} else {
			if _, err := inferFileMIMEData(uploadPath); err != nil {
				return err
			}
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := uploadFile(ctx, client, uploadPath)
		if err != nil {
			return err
		}
		out, err := shapeUploadResponse(result)
		if err != nil {
			return err
		}
		return printJSON(out)
	}),
}

func uploadFile(ctx context.Context, client *retab.Client, uploadPath string) (*retab.MIMEData, error) {
	data, err := os.ReadFile(uploadPath)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(uploadPath)
	contentType := http.DetectContentType(data)
	sum := sha256.Sum256(data)
	sha256Hash := hex.EncodeToString(sum[:])
	prepared, err := client.Files.CreateUpload(ctx, &retab.FilesCreateUploadParams{
		Filename:    filename,
		ContentType: &contentType,
		SizeBytes:   len(data),
		Sha256:      &sha256Hash,
	})
	if err != nil {
		return nil, err
	}
	method := prepared.UploadMethod
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequestWithContext(ctx, method, prepared.UploadURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	for key, value := range prepared.UploadHeaders {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("direct upload failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return client.Files.CompleteUpload(ctx, prepared.FileID, &retab.FilesCompleteUploadParams{Sha256: &sha256Hash})
}

// stageStdinUpload drains stdin into a temp file named per --filename so the
// existing client.Files.Upload path (SHA256, content-type inference, 2-phase
// upload) handles stdin transparently. Returning a cleanup closure makes the
// RunE site read like the file path — no special-case branching on bytes vs.
// path further down. --filename is required (no sensible default for a piped
// blob; we'd otherwise label it "stdin" or similar and silently mis-type the
// content).
func stageStdinUpload(cmd *cobra.Command) (string, func(), error) {
	filename, _ := cmd.Flags().GetString("filename")
	if strings.TrimSpace(filename) == "" {
		return "", func() {}, fmt.Errorf("--filename is required when uploading from stdin (-)")
	}
	// Disallow path separators in the filename — we use it as a basename
	// for the staging file and we don't want a `..` or `/` to escape the
	// temp dir or be passed to the server as a path.
	if strings.ContainsAny(filename, `/\`) {
		return "", func() {}, fmt.Errorf("--filename must be a bare filename, not a path: %q", filename)
	}
	body, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return "", func() {}, fmt.Errorf("read stdin: %w", err)
	}
	dir, err := os.MkdirTemp("", "retab-upload-stdin-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("stage stdin upload: %w", err)
	}
	stagedPath := filepath.Join(dir, filename)
	if err := os.WriteFile(stagedPath, body, 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return "", func() {}, fmt.Errorf("stage stdin upload: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	return stagedPath, cleanup, nil
}

// shapeUploadResponse builds the JSON written by `files upload`. The SDK
// returns MIMEData (filename + url + optional mime_type) — the id is
// encoded in the URL path but isn't a distinct JSON field, so we surface
// it explicitly. The Long description on `files upload` literally tells
// users to run `| jq -r .id`, so we owe them an `.id` to read.
//
// id resolution prefers the SDK's typed accessor (which validates the
// retab storage URL shape); if that returns "", we fall back to parsing
// the URL path's basename minus the extension. If neither yields an id,
// we surface a hard error instead of silently emitting a partial blob —
// downstream commands all key off `.id`, so an empty id is a footgun.
//
// Output ordering — `id` first, then `filename`, then `url`, then any
// optional fields — is enforced via an OrderedMap-style sidecar map so
// the JSON encoder doesn't reshuffle keys. We use slice-pair encoding
// (see uploadResponse.MarshalJSON) because encoding/json sorts map keys
// alphabetically; without this `filename` would jump ahead of `id`.
func shapeUploadResponse(result *retab.MIMEData) (uploadResponse, error) {
	id := result.ID()
	if id == "" {
		id = fileIDFromURL(result.URL)
	}
	if id == "" {
		return uploadResponse{}, fmt.Errorf(
			"upload succeeded but server response is missing a file id (url=%s) — please report this",
			result.URL,
		)
	}
	out := uploadResponse{
		pairs: []uploadResponseField{
			{"id", id},
			{"filename", result.Filename},
			{"url", result.URL},
		},
	}
	if result.MIMEType != "" {
		out.pairs = append(out.pairs, uploadResponseField{"mime_type", result.MIMEType})
	}
	return out, nil
}

// uploadResponse is an ordered-key JSON object so that `id` is the first
// field a user sees — both for human readability and so tools that
// pretty-print don't bury the load-bearing field below the noise.
// encoding/json sorts map keys alphabetically, hence the custom marshal.
type uploadResponse struct {
	pairs []uploadResponseField
}

type uploadResponseField struct {
	Key   string
	Value any
}

// MarshalJSON emits the pairs in insertion order. Values are always
// JSON-encodable here, so we lean on encoding/json's encoder for each
// key/value (handles escaping, unicode, nested objects, etc.) and glue them
// together — no need for json.RawMessage gymnastics.
//
// Encoding goes through encodeJSONNoHTMLEscape rather than the package-level
// json.Marshal: the latter always HTML-escapes `&`, `<` and `>` into their
// unicode-escape forms, which would mangle signed upload URLs (full of `&`
// query separators) and clash with the rest of the CLI, whose output
// (printJSON) sets SetEscapeHTML(false).
func (u uploadResponse) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, p := range u.pairs {
		if i > 0 {
			b.WriteByte(',')
		}
		keyBytes, err := encodeJSONNoHTMLEscape(p.Key)
		if err != nil {
			return nil, err
		}
		valBytes, err := encodeJSONNoHTMLEscape(p.Value)
		if err != nil {
			return nil, err
		}
		b.Write(keyBytes)
		b.WriteByte(':')
		b.Write(valBytes)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

// encodeJSONNoHTMLEscape marshals v to compact JSON with HTML escaping
// disabled, so `&`, `<`, `>` survive verbatim. json.Encoder appends a
// trailing newline that we strip to keep the result a drop-in for
// json.Marshal.
func encodeJSONNoHTMLEscape(v any) ([]byte, error) {
	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return []byte(strings.TrimRight(sb.String(), "\n")), nil
}

// fileIDFromURL is the URL-parsing fallback when the SDK's typed
// MIMEData.ID() returns "" (e.g. the server returned a non-retab storage
// host). It pulls the basename out of the path and strips the file
// extension; query strings and trailing slashes are tolerated, but
// pathless URLs, empty basenames, and basenames without an extension are
// rejected so we never invent a bogus id.
//
// Examples:
//
//	https://storage.retab.com/org_x/file_abc123.pdf  → file_abc123
//	https://cdn.example/v1/file_zzz.txt              → file_zzz
//	https://x/file_a.pdf?token=...                   → file_a
//	https://x/file_a.pdf/                            → file_a
//	https://x/file_a                                 → "" (no extension)
//	https://x/                                       → "" (no basename)
func fileIDFromURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	// Trim query/fragment by ignoring them; reject completely empty paths.
	p := strings.Trim(parsed.Path, "/")
	if p == "" {
		return ""
	}
	base := path.Base(p)
	if base == "." || base == "/" || base == "" {
		return ""
	}
	idx := strings.LastIndex(base, ".")
	// Require a real extension (dot not at start, not at end).
	if idx <= 0 || idx == len(base)-1 {
		return ""
	}
	return base[:idx]
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
	Use:   "download <file-id> [dest]",
	Short: "Download a file to a local path (- writes to stdout)",
	Long: `Stream a file's bytes from storage to disk (or stdout).

The destination can be given as a positional argument or via ` + "`-o`" + `.
With neither, writes to the server-recorded filename in the current
directory, falling back to the file id if the server didn't store a
name. Use ` + "`-`" + ` (positional or ` + "`-o -`" + `) to write to stdout so the output
can be piped into another tool.

The positional and ` + "`-o`" + ` forms are mutually exclusive — passing both
is rejected. Pick whichever reads better at the call site.

Downloads stream chunk-by-chunk and propagate Ctrl-C. A file destination
is written atomically — bytes land in a temp file and are renamed over
the destination only on success — so a canceled or failed transfer never
truncates an existing file or leaves a half-written one behind.`,
	Example: `  # Save under the server's filename
  retab files download file_abc123

  # Pipe to stdout (e.g. into another tool)
  retab files download file_abc123 -

  # Save to an explicit path (positional)
  retab files download file_abc123 ./invoice.pdf

  # Same, with the -o flag form
  retab files download file_abc123 -o ./invoice.pdf

  # Stdout via the flag form (equivalent to the positional -)
  retab files download file_abc123 -o - | pdftotext - -`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		oFlag, _ := cmd.Flags().GetString("output")
		dest, toStdout, err := resolveDownloadDest(args, oFlag)
		if err != nil {
			return err
		}
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
		if !toStdout && dest == "" {
			dest = link.Filename
			if dest == "" {
				dest = args[0]
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, link.DownloadURL, nil)
		if err != nil {
			return err
		}
		resp, err := fileDownloadClient.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("download failed: %d %s", resp.StatusCode, string(body))
		}
		if toStdout {
			_, err = io.Copy(os.Stdout, resp.Body)
			return err
		}
		return streamDownloadToFile(dest, resp.Body)
	}),
}

// streamDownloadToFile writes src to dest atomically: bytes go to a temp
// file in the same directory and are renamed over dest only after the copy
// fully succeeds. A failed or interrupted transfer (network drop, Ctrl-C)
// therefore never truncates a pre-existing dest and never leaves a
// half-written file at the dest path — the temp file is removed instead.
//
// The earlier implementation called os.Create(dest) directly, which
// truncated dest to zero bytes *before* the first byte arrived: a download
// that then failed mid-stream destroyed whatever the user already had at
// that path.
func streamDownloadToFile(dest string, src io.Reader) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".partial-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Any failure past this point: drop the temp file so the dest path is
	// never left half-written and a retry starts from a clean slate.
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	if _, err = io.Copy(tmp, src); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpName, dest); err != nil {
		return err
	}
	return nil
}

// resolveDownloadDest collapses the two destination-input forms (positional
// arg vs. -o flag) into a single (path, toStdout) result. Pulled out of
// filesDownloadCmd.RunE so the resolution rules can be exercised without
// spinning up an HTTP client or a real file.
//
// Rules:
//   - args[1] and oFlag are mutually exclusive; passing both is an error.
//   - "-" (positional or flag) maps to (path="", toStdout=true).
//   - Any other non-empty value maps to (path=value, toStdout=false).
//   - Neither given maps to (path="", toStdout=false) — the caller then
//     falls back to the server-recorded filename / file id.
func resolveDownloadDest(args []string, oFlag string) (path string, toStdout bool, err error) {
	var positional string
	if len(args) >= 2 {
		positional = args[1]
	}
	if positional != "" && oFlag != "" {
		return "", false, fmt.Errorf("cannot use positional %s and -o flag together", positional)
	}
	value := positional
	if value == "" {
		value = oFlag
	}
	if value == "-" {
		return "", true, nil
	}
	return value, false, nil
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
		filename, err := requireNonBlankFlag(cmd, "filename")
		if err != nil {
			return err
		}
		contentType, err := requireNonBlankFlag(cmd, "content-type")
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		size, _ := cmd.Flags().GetInt64("size-bytes")
		sha256Hash, _ := cmd.Flags().GetString("sha256")
		result, err := client.Files.CreateUpload(ctx, &retab.FilesCreateUploadParams{
			Filename:    filename,
			ContentType: &contentType,
			SizeBytes:   int(size),
			Sha256:      &sha256Hash,
		})
		if err != nil {
			return err
		}
		out, err := shapeCreateUploadResponse(result)
		if err != nil {
			return err
		}
		return printJSON(out)
	}),
}

func shapeCreateUploadResponse(result *retab.CreateUploadResponse) (uploadResponse, error) {
	if result.FileID == "" {
		return uploadResponse{}, fmt.Errorf("create-upload succeeded but server response is missing a file id")
	}
	if result.UploadURL == "" {
		return uploadResponse{}, fmt.Errorf("create-upload succeeded but server response is missing an upload URL")
	}
	return uploadResponse{
		pairs: []uploadResponseField{
			{"id", result.FileID},
			{"upload_url", result.UploadURL},
			{"upload_method", result.UploadMethod},
			{"upload_headers", result.UploadHeaders},
			{"mime_data", result.MIMEData},
			{"expires_at", result.ExpiresAt},
		},
	}, nil
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
		result, err := client.Files.CompleteUpload(ctx, args[0], &retab.FilesCompleteUploadParams{Sha256: &sha256Hash})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	addListFlags(filesListCmd, false)
	filesListCmd.Flags().String("mime-type", "", "filter by MIME type")
	filesListCmd.Flags().Var(newEnumStringFlagValue("--sort-by", "created_at", "updated_at"), "sort-by", "sort field: created_at | updated_at (default: created_at)")

	filesDownloadCmd.Flags().StringP("output", "o", "", "output path, - for stdout (alternative to the [dest] positional; default: server filename)")

	filesUploadCmd.Flags().String("filename", "", "filename to record on the server (required when reading from stdin)")

	filesCreateUploadCmd.Flags().String("filename", "", "filename (required)")
	filesCreateUploadCmd.Flags().String("content-type", "", "content type (required)")
	filesCreateUploadCmd.Flags().Var(&nonNegativeInt64FlagValue{}, "size-bytes", "file size in bytes (required)")
	filesCreateUploadCmd.Flags().Var(&sha256FlagValue{}, "sha256", "sha256 hex digest (optional)")
	_ = filesCreateUploadCmd.MarkFlagRequired("filename")
	_ = filesCreateUploadCmd.MarkFlagRequired("content-type")
	_ = filesCreateUploadCmd.MarkFlagRequired("size-bytes")

	filesCompleteUploadCmd.Flags().Var(&sha256FlagValue{}, "sha256", "sha256 hex digest (optional)")

	filesCmd.AddCommand(filesListCmd, filesGetCmd, filesUploadCmd, filesDownloadLinkCmd, filesDownloadCmd, filesCreateUploadCmd, filesCompleteUploadCmd)
	rootCmd.AddCommand(filesCmd)
}
