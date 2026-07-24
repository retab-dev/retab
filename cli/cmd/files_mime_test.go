//go:build !retab_oagen_cli_files

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveUploadMIMEType pins the preference order that guarantees
// `files upload` never emits an empty mime_type for a normal upload:
//  1. the server-resolved MIMEType, when non-empty;
//  2. otherwise the extension-based type (matching the server's run-time
//     resolution) — e.g. .pdf → application/pdf, .jpg → image/jpeg;
//  3. otherwise the http.DetectContentType value sniffed from the bytes.
func TestResolveUploadMIMEType(t *testing.T) {
	cases := []struct {
		name       string
		serverMIME string
		uploadPath string
		detected   string
		want       string
	}{
		{
			name:       "server MIMEType wins when present",
			serverMIME: "application/pdf",
			uploadPath: "/tmp/whatever.bin",
			detected:   "application/octet-stream",
			want:       "application/pdf",
		},
		{
			name:       "empty server type falls back to extension (.pdf)",
			serverMIME: "",
			uploadPath: "/tmp/invoice.pdf",
			detected:   "application/octet-stream",
			want:       "application/pdf",
		},
		{
			name:       "empty server type falls back to extension (.jpg)",
			serverMIME: "",
			uploadPath: "/tmp/photo.jpg",
			detected:   "application/octet-stream",
			want:       "image/jpeg",
		},
		{
			name:       "empty server type falls back to extension (.jpeg)",
			serverMIME: "",
			uploadPath: "/tmp/photo.jpeg",
			detected:   "application/octet-stream",
			want:       "image/jpeg",
		},
		{
			name:       "extension params are stripped to a bare media type",
			serverMIME: "",
			uploadPath: "/tmp/notes.txt",
			detected:   "application/octet-stream",
			// .txt resolves to "text/plain; charset=utf-8" — we keep only the
			// bare media type.
			want: "text/plain",
		},
		{
			name:       "no extension falls back to detected content type",
			serverMIME: "",
			uploadPath: "/tmp/no_extension",
			detected:   "application/pdf",
			want:       "application/pdf",
		},
		{
			name:       "staged stdin path keeps its extension",
			serverMIME: "",
			uploadPath: filepath.Join(os.TempDir(), "retab-upload-stdin-xyz", "invoice.pdf"),
			detected:   "application/octet-stream",
			want:       "application/pdf",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveUploadMIMEType(tc.serverMIME, tc.uploadPath, tc.detected)
			if got != tc.want {
				t.Errorf("resolveUploadMIMEType(%q, %q, %q) = %q, want %q",
					tc.serverMIME, tc.uploadPath, tc.detected, got, tc.want)
			}
		})
	}
}

// TestFilesUploadAlwaysEmitsMIMEType is the regression guard for the bug:
// the server's complete-upload response comes back with an EMPTY mime_type,
// yet `files upload` must still print a correct, non-empty mime_type derived
// from the local file (application/pdf for a .pdf, image/jpeg for a .jpg).
//
// The whole two-phase flow is driven against a single httptest server:
// POST /v1/files/upload (create), PUT to the signed upload URL, then
// POST /v1/files/upload/{id}/complete — whose response deliberately omits
// mime_type to reproduce the reported shape.
func TestFilesUploadAlwaysEmitsMIMEType(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		bytes    []byte
		wantMIME string
	}{
		{
			name:     "pdf",
			filename: "invoice.pdf",
			// %PDF magic so http.DetectContentType also recognizes it.
			bytes:    []byte("%PDF-1.4\n%test pdf body\n"),
			wantMIME: "application/pdf",
		},
		{
			name:     "jpeg",
			filename: "photo.jpg",
			// JPEG SOI marker.
			bytes:    []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00},
			wantMIME: "image/jpeg",
		},
		{
			name:     "tiff extension beats octet-stream sniff",
			filename: "scan.tif",
			// Deliberately not enough magic for http.DetectContentType to
			// identify TIFF; extension inference is what should reach the server.
			bytes:    []byte("not enough tiff magic"),
			wantMIME: "image/tiff",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "rt_test_key")
			t.Setenv("HOME", t.TempDir())

			dir := t.TempDir()
			localPath := filepath.Join(dir, tc.filename)
			if err := os.WriteFile(localPath, tc.bytes, 0o600); err != nil {
				t.Fatalf("seed upload file: %v", err)
			}

			var seenCreateContentType string
			var seenPutContentType string
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/v1/files/upload":
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("decode create-upload body: %v", err)
					}
					seenCreateContentType, _ = body["content_type"].(string)
					// create-upload: hand back a signed PUT URL pointing back
					// at this same server, plus the reserved file id.
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"fileId":        "file_mime_e2e",
						"uploadUrl":     server.URL + "/signed-put/file_mime_e2e",
						"uploadMethod":  "PUT",
						"uploadHeaders": map[string]string{},
						"mimeData": map[string]string{
							"filename": tc.filename,
							"url":      "https://storage.retab.com/org_1/file_mime_e2e" + filepath.Ext(tc.filename),
						},
						"expiresAt": "2026-05-15T12:32:26Z",
					})
				case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/signed-put/"):
					// direct upload PUT: accept the bytes.
					seenPutContentType = r.Header.Get("Content-Type")
					w.WriteHeader(http.StatusOK)
				case r.Method == http.MethodPost && r.URL.Path == "/v1/files/upload/file_mime_e2e/complete":
					// complete-upload: reproduce the reported bug — the server
					// response carries an EMPTY mime_type.
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"filename": tc.filename,
						"url":      "https://storage.retab.com/org_1/file_mime_e2e" + filepath.Ext(tc.filename),
						// mime_type intentionally omitted/empty.
					})
				default:
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			t.Cleanup(func() { _ = filesUploadCmd.Flags().Set("filename", "") })

			stdout, stderr := captureStd(t, func() {
				if err := filesUploadCmd.RunE(filesUploadCmd, []string{localPath}); err != nil {
					t.Fatalf("files upload: %v", err)
				}
			})
			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}

			var got map[string]any
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
			}
			if got["id"] != "file_mime_e2e" {
				t.Fatalf("id = %#v, want file_mime_e2e; raw:\n%s", got["id"], stdout)
			}
			mt, ok := got["mime_type"].(string)
			if !ok || mt == "" {
				t.Fatalf("mime_type missing/empty in output (bug not fixed); raw:\n%s", stdout)
			}
			if mt != tc.wantMIME {
				t.Fatalf("mime_type = %q, want %q; raw:\n%s", mt, tc.wantMIME, stdout)
			}
			if seenCreateContentType != tc.wantMIME {
				t.Fatalf("create-upload content_type = %q, want %q", seenCreateContentType, tc.wantMIME)
			}
			if seenPutContentType != tc.wantMIME {
				t.Fatalf("PUT Content-Type = %q, want %q", seenPutContentType, tc.wantMIME)
			}
		})
	}
}
