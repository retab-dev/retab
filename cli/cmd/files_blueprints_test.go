//go:build !retab_oagen_cli_files

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFilesBlueprintsCreatePostsPublicBlueprintEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() {
		_ = filesBlueprintsCreateCmd.Flags().Set("intent", "")
		_ = filesBlueprintsCreateCmd.Flags().Set("background", "true")
	})

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/files/blueprints" {
			t.Fatalf("path = %s, want /v1/files/blueprints", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"fbp_123","object":"file_blueprint","status":"queued"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(
			t,
			"files",
			"blueprints",
			"create",
			"file_1",
			"--intent",
			"Find invoice fields",
		); err != nil {
			t.Fatalf("files blueprints create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, `"fbp_123"`) {
		t.Fatalf("stdout %q does not contain blueprint id", stdout)
	}
	if body["file_id"] != "file_1" {
		t.Fatalf("file_id = %#v, want file_1", body["file_id"])
	}
	if body["intent"] != "Find invoice fields" {
		t.Fatalf("intent = %#v, want Find invoice fields", body["intent"])
	}
	if body["background"] != true {
		t.Fatalf("background = %#v, want true", body["background"])
	}
}

func TestFilesBlueprintsCreateCanOptOutOfBackground(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() {
		_ = filesBlueprintsCreateCmd.Flags().Set("background", "true")
	})

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"fbp_123","object":"file_blueprint","status":"completed"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := runRootForTest(
		t,
		"files",
		"blueprints",
		"create",
		"file_1",
		"--background=false",
	); err != nil {
		t.Fatalf("files blueprints create: %v", err)
	}

	if body["background"] != false {
		t.Fatalf("background = %#v, want false", body["background"])
	}
}

func TestFilesBlueprintsGetAndCancelUseBlueprintIDEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	seen := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"fbp_123","object":"file_blueprint","status":"cancelled"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := runRootForTest(t, "files", "blueprints", "get", "fbp_123", "--include-output=false"); err != nil {
		t.Fatalf("files blueprints get: %v", err)
	}
	if err := runRootForTest(t, "files", "blueprints", "cancel", "fbp_123"); err != nil {
		t.Fatalf("files blueprints cancel: %v", err)
	}

	want := []string{
		"GET /v1/files/blueprints/fbp_123?include_output=false",
		"POST /v1/files/blueprints/fbp_123/cancel",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("requests:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
}
