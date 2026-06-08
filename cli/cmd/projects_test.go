package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestProjectsListHitsProjectsEndpoint pins that `retab projects list`
// fetches the env-scoped /v1/projects surface and renders the pagination
// envelope. This is the discovery path for the project ids that
// `workflows create` / `workflows spec apply` now require.
func TestProjectsListHitsProjectsEndpoint(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"proj_1","name":"Logistics","slug":"logistics"}],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, err := captureStdAndRun(t, func() error {
		return projectsListCmd.RunE(projectsListCmd, nil)
	})
	if err != nil {
		t.Fatalf("projects list: %v", err)
	}

	if seenPath != "/v1/projects" {
		t.Fatalf("path = %q, want /v1/projects", seenPath)
	}
	if !strings.Contains(stdout, "proj_1") || !strings.Contains(stdout, "logistics") {
		t.Fatalf("projects list output missing project, got:\n%s", stdout)
	}
}

// TestProjectsListForwardsPaginationQuery pins that --after / --limit /
// --include-archived become query params on the request.
func TestProjectsListForwardsPaginationQuery(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"list_metadata":{"before":null,"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := projectsListCmd.Flags().Set("after", "proj_1"); err != nil {
		t.Fatal(err)
	}
	if err := projectsListCmd.Flags().Set("limit", "20"); err != nil {
		t.Fatal(err)
	}
	if err := projectsListCmd.Flags().Set("include-archived", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = projectsListCmd.Flags().Set("after", "")
		_ = projectsListCmd.Flags().Set("limit", "0")
		_ = projectsListCmd.Flags().Set("include-archived", "false")
	})

	if _, err := captureStdAndRun(t, func() error {
		return projectsListCmd.RunE(projectsListCmd, nil)
	}); err != nil {
		t.Fatalf("projects list: %v", err)
	}

	for _, want := range []string{"after=proj_1", "limit=20", "include_archived=true"} {
		if !strings.Contains(seenQuery, want) {
			t.Fatalf("query %q missing %q", seenQuery, want)
		}
	}
}
