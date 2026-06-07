//go:build !retab_oagen_cli_files

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

// failingReader yields data once, then errors — executing a transfer that
// drops mid-stream (network failure, Ctrl-C).
type failingReader struct {
	data []byte
	done bool
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, errors.New("executed transfer failure")
	}
	r.done = true
	return copy(p, r.data), nil
}

// TestStreamDownloadToFileSuccess pins that a successful download lands the
// full content at dest and leaves no temp file behind.
func TestStreamDownloadToFileSuccess(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")

	if err := streamDownloadToFile(dest, strings.NewReader("hello world")); err != nil {
		t.Fatalf("streamDownloadToFile: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != "hello world" {
		t.Fatalf("dest content = %q, want %q", got, "hello world")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected only dest in dir, got %d entries: %v", len(entries), entries)
	}
}

// TestStreamDownloadToFileFailureKeepsExistingFile is the regression guard:
// a download that fails mid-stream must NOT truncate a pre-existing file at
// dest and must NOT leave a partial file behind. The old os.Create(dest)
// implementation truncated dest to zero bytes before the first byte arrived.
func TestStreamDownloadToFileFailureKeepsExistingFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "precious.pdf")
	if err := os.WriteFile(dest, []byte("ORIGINAL IMPORTANT CONTENT"), 0o644); err != nil {
		t.Fatalf("seed dest: %v", err)
	}

	err := streamDownloadToFile(dest, &failingReader{data: []byte("junk")})
	if err == nil {
		t.Fatal("expected error from failing transfer")
	}

	got, readErr := os.ReadFile(dest)
	if readErr != nil {
		t.Fatalf("pre-existing dest disappeared: %v", readErr)
	}
	if string(got) != "ORIGINAL IMPORTANT CONTENT" {
		t.Fatalf("failed download corrupted existing file: dest = %q", got)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("failed download left a partial file behind: %v", entries)
	}
}

// TestResolveDownloadDest pins the destination-resolution rules for
// `retab files download`: positional vs. -o flag, "-" → stdout, and the
// mutex error when both forms are passed.
//
// The CLI surfaces two equivalent destination inputs because users coming
// from the help text (which used to say "- for stdout") expect the
// positional form to work, while older docs and pipelines lean on the
// -o flag. This test guards both shapes from regressing.
func TestResolveDownloadDest(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		oFlag       string
		wantPath    string
		wantStdout  bool
		wantErr     bool
		wantErrSubs string
	}{
		{
			name:       "one arg, no flag — defer to server filename",
			args:       []string{"file_abc"},
			oFlag:      "",
			wantPath:   "",
			wantStdout: false,
		},
		{
			name:       "one arg, explicit -o path",
			args:       []string{"file_abc"},
			oFlag:      "./out.pdf",
			wantPath:   "./out.pdf",
			wantStdout: false,
		},
		{
			name:       "one arg, -o - means stdout",
			args:       []string{"file_abc"},
			oFlag:      "-",
			wantPath:   "",
			wantStdout: true,
		},
		{
			name:       "two args, positional - means stdout",
			args:       []string{"file_abc", "-"},
			oFlag:      "",
			wantPath:   "",
			wantStdout: true,
		},
		{
			name:       "two args, positional path",
			args:       []string{"file_abc", "./out.pdf"},
			oFlag:      "",
			wantPath:   "./out.pdf",
			wantStdout: false,
		},
		{
			name:        "both positional and flag — reject",
			args:        []string{"file_abc", "./out.pdf"},
			oFlag:       "./other",
			wantErr:     true,
			wantErrSubs: "cannot use positional ./out.pdf and -o flag together",
		},
		{
			name:        "both positional - and flag — also reject",
			args:        []string{"file_abc", "-"},
			oFlag:       "./other",
			wantErr:     true,
			wantErrSubs: "cannot use positional - and -o flag together",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotStdout, err := resolveDownloadDest(tc.args, tc.oFlag)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got path=%q stdout=%v", gotPath, gotStdout)
				}
				if tc.wantErrSubs != "" && !strings.Contains(err.Error(), tc.wantErrSubs) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrSubs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotPath != tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tc.wantPath)
			}
			if gotStdout != tc.wantStdout {
				t.Errorf("toStdout = %v, want %v", gotStdout, tc.wantStdout)
			}
		})
	}
}

func TestFilesUploadReadsLocalFileBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	err := filesUploadCmd.RunE(filesUploadCmd, []string{"/tmp/missing.pdf"})
	if err == nil {
		t.Fatal("expected local file error")
	}
	if !strings.Contains(err.Error(), "file not found: /tmp/missing.pdf") {
		t.Fatalf("error %q does not mention missing file", err.Error())
	}
	if strings.Contains(err.Error(), "credentials") {
		t.Fatalf("error %q checked credentials before reading upload file", err.Error())
	}
}

// TestFileIDFromURL pins the URL-parsing fallback used when the SDK's
// typed MIMEData.ID() returns "" (non-retab storage host, etc.). The
// rules are deliberately conservative: tolerate query strings and
// trailing slashes, but reject anything we can't confidently turn into
// a file id (no path, no extension, empty basename) so we never invent
// a bogus id and silently hand it to downstream commands.
//
// This is the regression guard for the bug where `retab files upload`
// returned `{"filename": ..., "url": ...}` with no top-level `id` —
// users have to either parse the URL or list files afterwards if the
// id is missing, both of which are ugly. The fix relies on `ID()` from
// the SDK (preferred, validates the retab host) plus this fallback for
// any future server that hands us a non-retab URL.
func TestFileIDFromURL(t *testing.T) {
	cases := []struct {
		name   string
		url    string
		wantID string
	}{
		{
			name:   "retab storage URL",
			url:    "https://storage.retab.com/org_x/file_abc123.pdf",
			wantID: "file_abc123",
		},
		{
			name:   "non-retab CDN URL",
			url:    "https://cdn.example.com/v1/file_zzz.txt",
			wantID: "file_zzz",
		},
		{
			name:   "URL with query string",
			url:    "https://cdn.example.com/file_a.pdf?token=abc&exp=12345",
			wantID: "file_a",
		},
		{
			name:   "URL with trailing slash after extension",
			url:    "https://cdn.example.com/file_a.pdf/",
			wantID: "file_a",
		},
		{
			name:   "URL with fragment",
			url:    "https://cdn.example.com/file_q.pdf#section",
			wantID: "file_q",
		},
		{
			name:   "deeply nested path",
			url:    "https://storage.retab.com/a/b/c/d/file_deep.json",
			wantID: "file_deep",
		},
		{
			name:   "compound extension takes only the last segment",
			url:    "https://storage.retab.com/org_x/file_abc.tar.gz",
			wantID: "file_abc.tar",
		},
		{
			name:   "no extension — reject (can't tell id from filename)",
			url:    "https://cdn.example.com/file_no_ext",
			wantID: "",
		},
		{
			name:   "path is empty",
			url:    "https://cdn.example.com/",
			wantID: "",
		},
		{
			name:   "host only, no trailing slash",
			url:    "https://cdn.example.com",
			wantID: "",
		},
		{
			name:   "basename is just a dot — reject",
			url:    "https://cdn.example.com/.",
			wantID: "",
		},
		{
			name:   "basename starts with a dot (dotfile) — reject, no real extension",
			url:    "https://cdn.example.com/.hidden",
			wantID: "",
		},
		{
			name:   "basename ends with a dot — reject, no extension after it",
			url:    "https://cdn.example.com/trailing.",
			wantID: "",
		},
		{
			name:   "empty string",
			url:    "",
			wantID: "",
		},
		{
			name:   "malformed URL",
			url:    "://not a url",
			wantID: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotID := fileIDFromURL(tc.url)
			if gotID != tc.wantID {
				t.Errorf("fileIDFromURL(%q) = %q, want %q", tc.url, gotID, tc.wantID)
			}
		})
	}
}

func TestFilesCreateUploadShapesDocumentedOutput(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/files/upload" {
			t.Fatalf("path = %s, want /v1/files/upload", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fileId":       "file_123",
			"uploadUrl":    "https://uploads.example.com/file_123",
			"uploadMethod": "PUT",
			"uploadHeaders": map[string]string{
				"Content-Type": "text/plain",
			},
			"mimeData": map[string]string{
				"filename":  "direct.txt",
				"mime_type": "text/plain",
				"url":       "https://storage.retab.com/org_1/file_123.txt",
			},
			"expiresAt": "2026-05-15T12:32:26Z",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := filesCreateUploadCmd.Flags().Set("filename", "direct.txt"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filesCreateUploadCmd.Flags().Set("filename", "") })
	if err := filesCreateUploadCmd.Flags().Set("content-type", "text/plain"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filesCreateUploadCmd.Flags().Set("content-type", "") })
	if err := filesCreateUploadCmd.Flags().Set("size-bytes", "37"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filesCreateUploadCmd.Flags().Set("size-bytes", "0") })

	stdout, stderr := captureStd(t, func() {
		if err := filesCreateUploadCmd.RunE(filesCreateUploadCmd, nil); err != nil {
			t.Fatalf("create-upload: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if got["id"] != "file_123" {
		t.Fatalf("id = %#v, want file_123; raw:\n%s", got["id"], stdout)
	}
	if got["upload_url"] != "https://uploads.example.com/file_123" {
		t.Fatalf("upload_url = %#v, want documented snake_case key; raw:\n%s", got["upload_url"], stdout)
	}
	if _, ok := got["fileId"]; ok {
		t.Fatalf("raw SDK key fileId leaked into CLI output:\n%s", stdout)
	}
}

func TestFilesListUpdatedAtSortTableShowsUpdatedAt(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/files" {
			t.Fatalf("path = %s, want /v1/files", r.URL.Path)
		}
		if got := r.URL.Query().Get("sort_by"); got != "updated_at" {
			t.Fatalf("sort_by = %q, want updated_at", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":         "file_123",
					"filename":   "invoice.pdf",
					"created_at": "2026-05-15T10:00:00Z",
					"updated_at": "2026-05-15T13:24:08Z",
				},
			},
			"list_metadata": map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	filesListCmd.SetContext(context.Background())
	t.Cleanup(func() { filesListCmd.SetContext(context.Background()) })
	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })
	if err := filesListCmd.Flags().Set("sort-by", "updated_at"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filesListCmd.Flags().Set("sort-by", "") })

	var err error
	stdout, stderr := captureStd(t, func() {
		err = filesListCmd.RunE(filesListCmd, nil)
	})
	if err != nil {
		t.Fatalf("files list: %v", err)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "UPDATED_AT") {
		t.Fatalf("expected UPDATED_AT column, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "CREATED_AT") {
		t.Fatalf("did not expect CREATED_AT column for updated_at sort, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "2026-05-15T13:24:08Z") {
		t.Fatalf("expected updated_at value in table, got:\n%s", stdout)
	}
}

func TestFilesUploadSHA256FlagsRejectInvalidValuesLocally(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
	}{
		{name: "create-upload", cmd: "create-upload"},
		{name: "complete-upload", cmd: "complete-upload"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, _, err := filesCmd.Find([]string{tc.cmd})
			if err != nil {
				t.Fatalf("find %s: %v", tc.cmd, err)
			}
			err = cmd.Flags().Set("sha256", "banana")
			if err == nil {
				t.Fatal("expected local parse error for invalid --sha256")
			}
			if !strings.Contains(err.Error(), "64-character hex") {
				t.Fatalf("error %q does not describe the SHA-256 format", err.Error())
			}
			if resetErr := cmd.Flags().Set("sha256", strings.Repeat("a", 64)); resetErr != nil {
				t.Fatalf("set valid --sha256: %v", resetErr)
			}
			if resetErr := cmd.Flags().Set("sha256", ""); resetErr != nil {
				t.Fatalf("clear --sha256: %v", resetErr)
			}
		})
	}
}

// TestShapeUploadResponse covers both branches of id resolution:
//   - The SDK's typed MIMEData.ID() returns a non-empty id (the common
//     case for retab storage URLs), and we pass it straight through.
//   - MIMEData.ID() returns "" (non-retab host or any URL the SDK doesn't
//     recognize), and we fall back to URL parsing.
//
// We also pin the error path: if neither route yields an id, the CLI
// MUST error out instead of emitting a partial response that silently
// drops the id — every downstream command keys off `.id`, so an empty
// id is worse than a hard failure.
func TestShapeUploadResponse(t *testing.T) {
	cases := []struct {
		name        string
		input       retab.MIMEData
		wantID      string
		wantErr     bool
		wantErrSubs string
	}{
		{
			name: "SDK ID() succeeds (retab storage URL)",
			input: retab.MIMEData{
				Filename: "invoice.pdf",
				URL:      "https://storage.retab.com/org_01J/file_abc123.pdf",
			},
			wantID: "file_abc123",
		},
		{
			name: "SDK ID() fails, fallback URL parse succeeds",
			input: retab.MIMEData{
				Filename: "report.txt",
				URL:      "https://cdn.example.com/path/file_xyz.txt",
			},
			wantID: "file_xyz",
		},
		{
			name: "neither route yields an id — hard error",
			input: retab.MIMEData{
				Filename: "weird.bin",
				URL:      "https://cdn.example.com/no_extension",
			},
			wantErr:     true,
			wantErrSubs: "missing a file id",
		},
		{
			name: "empty URL — hard error",
			input: retab.MIMEData{
				Filename: "nothing.bin",
				URL:      "",
			},
			wantErr:     true,
			wantErrSubs: "missing a file id",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := shapeUploadResponse(&tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got out=%+v", out)
				}
				if tc.wantErrSubs != "" && !strings.Contains(err.Error(), tc.wantErrSubs) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrSubs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// id must be the first pair so users see it in pretty-printed
			// output and so `| jq -r .id` works deterministically.
			if len(out.pairs) < 3 {
				t.Fatalf("expected at least 3 pairs, got %d", len(out.pairs))
			}
			if out.pairs[0].Key != "id" || out.pairs[0].Value != tc.wantID {
				t.Errorf("first pair = %+v, want {id, %q}", out.pairs[0], tc.wantID)
			}
			if out.pairs[1].Key != "filename" || out.pairs[1].Value != tc.input.Filename {
				t.Errorf("second pair = %+v, want {filename, %q}", out.pairs[1], tc.input.Filename)
			}
			if out.pairs[2].Key != "url" || out.pairs[2].Value != tc.input.URL {
				t.Errorf("third pair = %+v, want {url, %q}", out.pairs[2], tc.input.URL)
			}
		})
	}
}

// TestShapeUploadResponse_PassesThroughMIMEType pins the optional
// mime_type pass-through. The SDK's MIMEData carries a MIMEType field
// that's only populated on the complete-upload response; when present
// we surface it after the three core fields. Other server-side fields
// (size_bytes, sha256, created_at, etc.) aren't currently on MIMEData
// so we can't pass them through — when the SDK shape grows, extend
// this test.
func TestShapeUploadResponse_PassesThroughMIMEType(t *testing.T) {
	out, err := shapeUploadResponse(&retab.MIMEData{
		Filename: "invoice.pdf",
		URL:      "https://storage.retab.com/org_01J/file_with_mime.pdf",
		MIMEType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.pairs) != 4 {
		t.Fatalf("expected 4 pairs (id, filename, url, mime_type), got %d: %+v", len(out.pairs), out.pairs)
	}
	if out.pairs[3].Key != "mime_type" || out.pairs[3].Value != "application/pdf" {
		t.Errorf("fourth pair = %+v, want {mime_type, application/pdf}", out.pairs[3])
	}
}

// TestUploadResponseMarshalJSON pins both the field order and the
// shape of the rendered JSON. End-to-end the only thing that matters
// is that `jq -r .id` works on the rendered output — so the test
// asserts (a) `id` is the first key in the encoded bytes, (b) the
// rendered JSON is valid and round-trips to a map with the same keys,
// and (c) the `id` value matches what the SDK + fallback resolved to.
//
// Acceptance criterion: after upload, the JSON has at minimum
// {id, filename, url}, and `id` is the first / most prominent field.
func TestUploadResponseMarshalJSON(t *testing.T) {
	out, err := shapeUploadResponse(&retab.MIMEData{
		Filename: "invoice.pdf",
		URL:      "https://storage.retab.com/org_01J/file_render_test.pdf",
		MIMEType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	rendered := string(encoded)

	// `id` must appear first so pretty-printers don't bury it.
	idIdx := strings.Index(rendered, `"id":`)
	filenameIdx := strings.Index(rendered, `"filename":`)
	urlIdx := strings.Index(rendered, `"url":`)
	mimeIdx := strings.Index(rendered, `"mime_type":`)
	if idIdx == -1 || filenameIdx == -1 || urlIdx == -1 || mimeIdx == -1 {
		t.Fatalf("expected all 4 keys present, got: %s", rendered)
	}
	if idIdx >= filenameIdx || filenameIdx >= urlIdx || urlIdx >= mimeIdx {
		t.Errorf("expected key order id, filename, url, mime_type — got: %s", rendered)
	}

	// JSON must round-trip.
	var roundTrip map[string]string
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatalf("round-trip unmarshal failed: %v\nrendered: %s", err, rendered)
	}
	if roundTrip["id"] != "file_render_test" {
		t.Errorf("id = %q, want %q", roundTrip["id"], "file_render_test")
	}
	if roundTrip["filename"] != "invoice.pdf" {
		t.Errorf("filename = %q, want %q", roundTrip["filename"], "invoice.pdf")
	}
	if roundTrip["url"] != "https://storage.retab.com/org_01J/file_render_test.pdf" {
		t.Errorf("url = %q, want %q", roundTrip["url"], "https://storage.retab.com/org_01J/file_render_test.pdf")
	}
	if roundTrip["mime_type"] != "application/pdf" {
		t.Errorf("mime_type = %q, want %q", roundTrip["mime_type"], "application/pdf")
	}
}

// TestUploadResponseMarshalJSON_FallbackEndToEnd is the end-to-end
// case the bug report asks for: given a fake SDK response (a non-retab
// host where the SDK's typed ID() returns ""), the rendered JSON still
// contains a top-level `"id":"file_..."`. This is what `jq -r .id`
// will read on every user's shell — the test guards the contract.
func TestUploadResponseMarshalJSON_FallbackEndToEnd(t *testing.T) {
	// Non-retab host → SDK ID() returns "" → fallback parses URL.
	fakeSDKResponse := &retab.MIMEData{
		Filename: "report.txt",
		URL:      "https://cdn.example.com/v1/files/file_e2e_fallback.txt",
	}
	out, err := shapeUploadResponse(fakeSDKResponse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	rendered := string(encoded)
	if !strings.Contains(rendered, `"id":"file_e2e_fallback"`) {
		t.Errorf("rendered output should contain top-level id from URL fallback, got: %s", rendered)
	}
}

// TestFilesDownloadCmdArgsRange pins that the cobra Args validator accepts
// 1-2 positional args and rejects 0 or 3+. This catches accidental
// regressions to cobra.ExactArgs(1) or RangeArgs(1, 3).
func TestFilesDownloadCmdArgsRange(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "zero args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"file_abc"}, wantErr: false},
		{name: "two args", args: []string{"file_abc", "-"}, wantErr: false},
		{name: "two args path", args: []string{"file_abc", "./out.pdf"}, wantErr: false},
		{name: "three args", args: []string{"file_abc", "-", "extra"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := filesDownloadCmd.Args(filesDownloadCmd, tc.args)
			if tc.wantErr && err == nil {
				t.Fatalf("want error for args=%v, got nil", tc.args)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for args=%v: %v", tc.args, err)
			}
		})
	}
}
