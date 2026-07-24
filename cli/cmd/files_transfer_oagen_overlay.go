//go:build retab_oagen_cli_files

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
		result, detectedContentType, err := uploadFile(ctx, client, uploadPath)
		if err != nil {
			return err
		}
		out, err := shapeUploadResponse(result, uploadPath, detectedContentType)
		if err != nil {
			return err
		}
		return printJSON(out)
	}),
}

// uploadFile runs the two-phase upload and also returns the content type it
// detected from the file bytes (http.DetectContentType). That detected type
// is the last-resort hint shapeUploadResponse uses for `mime_type` when the
// server response carries an empty MIMEType and extension inference fails —
// so the caller never has to re-read the bytes to recover what we already
// sniffed here.
func uploadFile(ctx context.Context, client *retab.Client, uploadPath string) (*retab.MIMEData, string, error) {
	data, err := os.ReadFile(uploadPath)
	if err != nil {
		return nil, "", err
	}
	filename := filepath.Base(uploadPath)
	contentType := detectUploadContentType(uploadPath, data)
	sum := sha256.Sum256(data)
	sha256Hash := hex.EncodeToString(sum[:])
	prepared, err := client.Files.CreateUpload(ctx, &retab.FilesCreateUploadParams{
		Filename:    filename,
		ContentType: &contentType,
		SizeBytes:   len(data),
		Sha256:      &sha256Hash,
	})
	if err != nil {
		return nil, "", err
	}
	method := prepared.UploadMethod
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequestWithContext(ctx, method, prepared.UploadURL, bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	for key, value := range prepared.UploadHeaders {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", contentType)
	// Use the bounded transfer client, not http.DefaultClient: the latter has
	// no timeout, so a wedged storage endpoint would hang an upload forever in
	// an unattended script (only Ctrl-C via ctx would break it).
	resp, err := fileDownloadClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, "", fmt.Errorf("direct upload failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	result, err := client.Files.CompleteUpload(ctx, prepared.FileID, &retab.FilesCompleteUploadParams{Sha256: &sha256Hash})
	if err != nil {
		return nil, "", err
	}
	return result, contentType, nil
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
// mime_type resolution — the server's CompleteUpload response frequently
// comes back with an EMPTY MIMEType, but the CLI already knows the file's
// type locally, so we never leave the field blank for a normal upload.
// Preference order (see resolveUploadMIMEType):
//  1. server result.MIMEType, when non-empty;
//  2. extension-based mime.TypeByExtension(uploadPath) — the same
//     extension→mime mapping the server itself applies at run time, so a
//     .pdf reads as application/pdf and a .jpg as image/jpeg;
//  3. the http.DetectContentType value uploadFile already computed from the
//     bytes, as a last resort for extension-less paths.
func shapeUploadResponse(result *retab.MIMEData, uploadPath, detectedContentType string) (uploadResponse, error) {
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
	if mimeType := resolveUploadMIMEType(result.MIMEType, uploadPath, detectedContentType); mimeType != "" {
		out.pairs = append(out.pairs, uploadResponseField{"mime_type", mimeType})
	}
	return out, nil
}

// resolveUploadMIMEType picks the mime_type displayed by `files upload`,
// never returning empty for a normal upload. It prefers the server-resolved
// type, then the extension-based type (matching the server's run-time
// resolution), then the content-sniffed type uploadFile detected from the
// bytes. mime.TypeByExtension can append parameters (e.g. "; charset=utf-8");
// we strip those so the field is a bare media type like "application/pdf".
func resolveUploadMIMEType(serverMIMEType, uploadPath, detectedContentType string) string {
	if serverMIMEType != "" {
		return serverMIMEType
	}
	if byExt := mimeTypeFromExtension(uploadPath); byExt != "" {
		return byExt
	}
	return detectedContentType
}

func detectUploadContentType(uploadPath string, data []byte) string {
	if byExt := mimeTypeFromExtension(uploadPath); byExt != "" {
		return byExt
	}
	return http.DetectContentType(data)
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
		oFlag, _ := cmd.Flags().GetString("out")
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
			dest = safeDownloadName(link.Filename)
			if dest == "" {
				dest = args[0]
			}
		}
		if !toStdout {
			dest = resolveDirDest(dest, link.Filename, args[0])
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

func init() {
	filesDownloadCmd.Flags().StringP("out", "o", "", "output path, - for stdout (alternative to the [dest] positional; default: server filename)")

	filesUploadCmd.Flags().String("filename", "", "filename to record on the server (required when reading from stdin)")

	filesCmd.AddCommand(filesUploadCmd, filesDownloadCmd)
}
