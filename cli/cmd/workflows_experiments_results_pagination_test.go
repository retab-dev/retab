//go:build !retab_oagen_cli_workflows_experiments

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// resetExperimentResultsListFlags hands the package-level singleton command its
// default cursor flags back, so one test's --after does not leak into the next.
func resetExperimentResultsListFlags(t *testing.T) {
	t.Helper()
	_ = workflowsExperimentsResultsListCmd.Flags().Set("before", "")
	_ = workflowsExperimentsResultsListCmd.Flags().Set("after", "")
	_ = workflowsExperimentsResultsListCmd.Flags().Set("order", "")
	_ = workflowsExperimentsResultsListCmd.Flags().Set("limit", "20")
}

// The results route pages on a keyset cursor and its response already carried
// `list_metadata.after`, but the command only wired `--limit`. So the cursor was
// a dead end: on an experiment with more documents than one page, every result
// past the first page was unreachable from the CLI. Pin that the cursor and
// order flags reach the query string.
func TestWorkflowsExperimentsResultsListForwardsCursorFlags(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.Contains(r.URL.Path, "/results") {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []any{},
			"list_metadata": map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	t.Cleanup(func() { resetExperimentResultsListFlags(t) })
	for flag, value := range map[string]string{"after": "expjob_page1_last", "limit": "50", "order": "asc"} {
		if err := workflowsExperimentsResultsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
	}

	captureStd(t, func() {
		if err := workflowsExperimentsResultsListCmd.RunE(workflowsExperimentsResultsListCmd, []string{"exprun_aaa"}); err != nil {
			t.Fatalf("experiment results list: %v", err)
		}
	})

	for _, want := range []string{"after=expjob_page1_last", "limit=50", "order=asc"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q missing %q", gotQuery, want)
		}
	}
	if strings.Contains(gotQuery, "before=") {
		t.Fatalf("query %q sent an empty before cursor", gotQuery)
	}
}

// --before walks back, and must not be sent alongside --after.
func TestWorkflowsExperimentsResultsListBeforeAfterMutuallyExclusive(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("the CLI must reject the flag pair before issuing %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	t.Cleanup(func() { resetExperimentResultsListFlags(t) })
	if err := workflowsExperimentsResultsListCmd.Flags().Set("before", "expjob_a"); err != nil {
		t.Fatalf("set --before: %v", err)
	}
	if err := workflowsExperimentsResultsListCmd.Flags().Set("after", "expjob_b"); err != nil {
		t.Fatalf("set --after: %v", err)
	}

	err := workflowsExperimentsResultsListCmd.RunE(workflowsExperimentsResultsListCmd, []string{"exprun_aaa"})
	if err == nil {
		t.Fatal("expected --before with --after to be rejected, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error %q does not explain the flag conflict", err.Error())
	}
}

// A plain first-page call stays unchanged: no cursor keys in the query.
func TestWorkflowsExperimentsResultsListOmitsUnsetCursors(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "list_metadata": map[string]any{}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)
	t.Cleanup(func() { resetExperimentResultsListFlags(t) })

	captureStd(t, func() {
		if err := workflowsExperimentsResultsListCmd.RunE(workflowsExperimentsResultsListCmd, []string{"exprun_aaa"}); err != nil {
			t.Fatalf("experiment results list: %v", err)
		}
	})
	for _, unwanted := range []string{"before=", "after=", "order="} {
		if strings.Contains(gotQuery, unwanted) {
			t.Fatalf("query %q sent an unset %q", gotQuery, unwanted)
		}
	}
	if !strings.Contains(gotQuery, "limit=20") {
		t.Fatalf("query %q lost the default limit", gotQuery)
	}
}
