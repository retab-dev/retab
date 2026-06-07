//go:build !retab_oagen_cli_files

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestFilesUploadFromStdinSendsFilenameToCreateUpload covers the `-` →
// stdin path: bytes piped into the command must be uploaded with the
// filename declared via --filename. The CLI uses the same 2-phase
// upload the SDK's Files.Upload does (create-upload + direct PUT +
// complete-upload), so the test mocks all three legs.
func TestFilesUploadFromStdinSendsFilenameToCreateUpload(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	const payload = "hello from stdin"
	var sawCreate atomic.Int32
	var sawPut atomic.Int32
	var sawComplete atomic.Int32
	var seenFilename atomic.Value
	var seenContentType atomic.Value

	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("upload server got %s, want PUT", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upload body: %v", err)
		}
		if string(body) != payload {
			t.Fatalf("upload body = %q, want %q", string(body), payload)
		}
		sawPut.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/files/upload":
			sawCreate.Add(1)
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create-upload body: %v", err)
			}
			if v, ok := body["filename"].(string); ok {
				seenFilename.Store(v)
			}
			if v, ok := body["content_type"].(string); ok {
				seenContentType.Store(v)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"fileId":        "file_stdin",
				"uploadUrl":     uploadServer.URL + "/upload",
				"uploadMethod":  "PUT",
				"uploadHeaders": map[string]string{},
				"mimeData": map[string]string{
					"filename":  "hello.txt",
					"mime_type": "text/plain",
					"url":       "https://storage.retab.com/org_1/file_stdin.txt",
				},
				"expiresAt": "2030-01-01T00:00:00Z",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/files/upload/file_stdin/complete":
			sawComplete.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"filename":  "hello.txt",
				"mime_type": "text/plain",
				"url":       "https://storage.retab.com/org_1/file_stdin.txt",
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := filesUploadCmd.Flags().Set("filename", "hello.txt"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filesUploadCmd.Flags().Set("filename", "") })

	filesUploadCmd.SetIn(strings.NewReader(payload))
	t.Cleanup(func() { filesUploadCmd.SetIn(nil) })

	stdout, stderr := captureStd(t, func() {
		if err := filesUploadCmd.RunE(filesUploadCmd, []string{"-"}); err != nil {
			t.Fatalf("upload from stdin: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	if sawCreate.Load() != 1 {
		t.Fatalf("create-upload calls = %d, want 1", sawCreate.Load())
	}
	if sawPut.Load() != 1 {
		t.Fatalf("PUT calls = %d, want 1", sawPut.Load())
	}
	if sawComplete.Load() != 1 {
		t.Fatalf("complete-upload calls = %d, want 1", sawComplete.Load())
	}

	gotFilename, _ := seenFilename.Load().(string)
	if gotFilename != "hello.txt" {
		t.Fatalf("create-upload filename = %q, want %q", gotFilename, "hello.txt")
	}
	gotContentType, _ := seenContentType.Load().(string)
	if gotContentType == "" {
		t.Fatalf("create-upload content_type was empty; want a non-empty type")
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if got["id"] != "file_stdin" {
		t.Fatalf("stdout id = %#v, want file_stdin; raw:\n%s", got["id"], stdout)
	}
	if got["filename"] != "hello.txt" {
		t.Fatalf("stdout filename = %#v, want hello.txt", got["filename"])
	}
}

// TestFilesUploadStdinRequiresFilename pins that piping bytes into the
// command without --filename is a clean error (not a confusing "file not
// found: -"). The filename is load-bearing: it's how the server names
// the blob and how content-type is inferred — defaulting it would
// silently mislabel uploads.
func TestFilesUploadStdinRequiresFilename(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	if err := filesUploadCmd.Flags().Set("filename", ""); err != nil {
		t.Fatal(err)
	}

	filesUploadCmd.SetIn(strings.NewReader("payload"))
	t.Cleanup(func() { filesUploadCmd.SetIn(nil) })

	err := filesUploadCmd.RunE(filesUploadCmd, []string{"-"})
	if err == nil {
		t.Fatal("expected error when stdin upload has no --filename")
	}
	if !strings.Contains(err.Error(), "--filename") {
		t.Fatalf("error %q does not mention --filename", err.Error())
	}
}
