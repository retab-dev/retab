package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Publishing a draft that already matches the live version is a server-side
// no-op: no version is cut and --description is discarded. The command still
// exits 0 and prints a `published` block, so without an explicit note a CI step
// running `publish --description "release $SHA"` reads as a successful release
// while the previous version's id and release note are what came back.
func newPublishNoopServer(t *testing.T, publishedVersion string, liveDescription string) *httptest.Server {
	t.Helper()
	workflow := func() map[string]any {
		return map[string]any{
			"id":   "wf_123",
			"name": "Unchanged Workflow",
			"published": map[string]any{
				"version_id":  publishedVersion,
				"description": liveDescription,
			},
		}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/wf_123":
			_ = json.NewEncoder(w).Encode(workflow())
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/wf_123/publish":
			_ = json.NewEncoder(w).Encode(workflow())
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func runPublishCommand(t *testing.T, serverURL string, description string) (string, string) {
	t.Helper()
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_BASE_URL", serverURL)

	cmd := &cobra.Command{Use: "publish", RunE: workflowsPublishCmd.RunE}
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("force", false, "")
	if err := cmd.Flags().Set("force", "true"); err != nil {
		t.Fatal(err)
	}
	if description != "" {
		if err := cmd.Flags().Set("description", description); err != nil {
			t.Fatal(err)
		}
	}
	var err error
	stdout, stderr := captureStd(t, func() {
		err = cmd.RunE(cmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("publish should still succeed, got %v", err)
	}
	return stdout, stderr
}

func TestWorkflowsPublishNotesNoopRepublish(t *testing.T) {
	server := newPublishNoopServer(t, "ver_live", "v2: the note already on that version")
	defer server.Close()

	stdout, stderr := runPublishCommand(t, server.URL, "v3: release abc123")

	if !strings.Contains(stderr, "no new version was published") {
		t.Fatalf("stderr should name the no-op publish, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "ver_live") {
		t.Fatalf("stderr should name the version that stayed live, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "--description was not stored") {
		t.Fatalf("stderr should say the release note was dropped, got:\n%s", stderr)
	}
	if !strings.Contains(stdout, "ver_live") {
		t.Fatalf("stdout should still print the publish response, got:\n%s", stdout)
	}
}

func TestWorkflowsPublishNoopWithoutDescriptionOmitsReleaseNoteLine(t *testing.T) {
	server := newPublishNoopServer(t, "ver_live", "v2: earlier note")
	defer server.Close()

	_, stderr := runPublishCommand(t, server.URL, "")

	if !strings.Contains(stderr, "no new version was published") {
		t.Fatalf("stderr should name the no-op publish, got:\n%s", stderr)
	}
	if strings.Contains(stderr, "--description") {
		t.Fatalf("stderr should not mention --description when it was not passed, got:\n%s", stderr)
	}
}

func TestWorkflowsPublishStaysQuietOnRealRelease(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		version := "ver_new"
		if r.Method == http.MethodGet {
			version = "ver_old"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "wf_123",
			"name":      "Changed Workflow",
			"published": map[string]any{"version_id": version},
		})
	}))
	defer server.Close()

	_, stderr := runPublishCommand(t, server.URL, "v3: real release")

	if strings.Contains(stderr, "no new version was published") {
		t.Fatalf("a genuine release should not be reported as a no-op, got:\n%s", stderr)
	}
}

// An unpublished workflow has no live version to compare against; the first
// publish must not be mistaken for a no-op.
func TestWorkflowsPublishFirstReleaseIsNotANoop(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body := map[string]any{"id": "wf_123", "name": "Fresh Workflow"}
		if r.Method == http.MethodPost {
			body["published"] = map[string]any{"version_id": "ver_first"}
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	_, stderr := runPublishCommand(t, server.URL, "v1: first release")

	if strings.Contains(stderr, "no new version was published") {
		t.Fatalf("first publish should not be reported as a no-op, got:\n%s", stderr)
	}
}
