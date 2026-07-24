//go:build !retab_oagen_cli_files

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFilesCreateUploadOmitsEmptySha256 pins that omitting --sha256 leaves the
// field out of the request body entirely.
//
// Regression: the command always sent `Sha256: &sha256Hash`, so an unset
// --sha256 marshaled `"sha256":""`. `omitempty` on a *string only omits a nil
// pointer, never a pointer to "". The server's create-upload model constrains
// sha256 with pattern ^[a-fA-F0-9]{64}$, so the empty string fails schema
// validation with a 422 — making a documented-optional flag effectively
// required.
func TestFilesCreateUploadOmitsEmptySha256(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/files/upload" {
			t.Fatalf("path = %s, want /v1/files/upload", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"fileId":"file_123","uploadUrl":"https://s.example.com/put","mimeData":{}}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, _ = captureStd(t, func() {
		if err := runRootForTest(
			t,
			"files", "create-upload",
			"--filename", "sample.pdf",
			"--content-type", "application/pdf",
			"--size-bytes", "1024",
		); err != nil {
			t.Fatalf("files create-upload: %v", err)
		}
	})

	if _, present := body["sha256"]; present {
		t.Fatalf("sha256 present in body when --sha256 was omitted: %#v", body["sha256"])
	}
}

// TestFilesCreateUploadSendsProvidedSha256 pins the positive path: a real
// digest is forwarded verbatim.
func TestFilesCreateUploadSendsProvidedSha256(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	const digest = "0a20b374fb23d6d03dd89109d773c9bb365b0d47ea45b03b83a796e3fe3e46db"
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"fileId":"file_123","uploadUrl":"https://s.example.com/put","mimeData":{}}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, _ = captureStd(t, func() {
		if err := runRootForTest(
			t,
			"files", "create-upload",
			"--filename", "sample.pdf",
			"--content-type", "application/pdf",
			"--size-bytes", "1024",
			"--sha256", digest,
		); err != nil {
			t.Fatalf("files create-upload: %v", err)
		}
	})

	if body["sha256"] != digest {
		t.Fatalf("sha256 = %#v, want %s", body["sha256"], digest)
	}
}

// TestFilesCompleteUploadOmitsEmptySha256 pins the same fix for complete-upload.
// There the empty string is a latent integrity bug: the server compares the
// provided digest against the stored object's sha256 and would overwrite a real
// stored digest with "" when the flag is unset.
func TestFilesCompleteUploadOmitsEmptySha256(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	bodyPresent := true
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&body); err != nil {
			// An empty/absent body is acceptable; record that nothing was sent.
			bodyPresent = false
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"filename":"sample.pdf","url":"https://storage.retab.com/org_01J/file_abc123.pdf","mime_type":"application/pdf"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := runRootForTest(
			t,
			"files", "complete-upload", "file_abc123",
		); err != nil {
			t.Fatalf("files complete-upload: %v", err)
		}
	})

	if bodyPresent {
		if _, present := body["sha256"]; present {
			t.Fatalf("sha256 present in body when --sha256 was omitted: %#v", body["sha256"])
		}
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if got["id"] != "file_abc123" {
		t.Fatalf("stdout id = %#v, want file_abc123; raw:\n%s", got["id"], stdout)
	}
	if got["filename"] != "sample.pdf" {
		t.Fatalf("stdout filename = %#v, want sample.pdf", got["filename"])
	}
}
