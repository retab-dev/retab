package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFilesAnalyzePostsPublicAnalyzeEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() {
		_ = filesAnalyzeCmd.Flags().Set("mode", "")
		_ = filesAnalyzeCmd.Flags().Set("intent", "")
	})

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/files/analyze" {
			t.Fatalf("path = %s, want /v1/files/analyze", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"job_123","object":"job","status":"queued","endpoint":"/v1/files/analyze"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(
			t,
			"files",
			"analyze",
			"file_1",
			"--mode",
			"instant",
			"--intent",
			"Find invoice fields",
			"--metadata",
			"source=cli",
		); err != nil {
			t.Fatalf("files analyze: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, `"job_123"`) {
		t.Fatalf("stdout %q does not contain job id", stdout)
	}
	if body["file_id"] != "file_1" {
		t.Fatalf("file_id = %#v, want file_1", body["file_id"])
	}
	if body["mode"] != "instant" {
		t.Fatalf("mode = %#v, want instant", body["mode"])
	}
	if body["intent"] != "Find invoice fields" {
		t.Fatalf("intent = %#v, want Find invoice fields", body["intent"])
	}
	metadata, ok := body["metadata"].(map[string]any)
	if !ok || metadata["source"] != "cli" {
		t.Fatalf("metadata = %#v, want source=cli", body["metadata"])
	}
}
