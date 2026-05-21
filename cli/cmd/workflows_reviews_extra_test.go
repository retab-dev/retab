package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// setReviewsBaseURL sets both env vars covering the in-flight rename of
// RETAB_BASE_URL → RETAB_API_BASE_URL so this test file stays correct
// regardless of which side of the rename has landed in auth.go.
func setReviewsBaseURL(t *testing.T, url string) {
	t.Helper()
	t.Setenv("RETAB_BASE_URL", url)
	t.Setenv("RETAB_API_BASE_URL", url)
}

func newReviewsListTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list", RunE: workflowsReviewsListCmd.RunE}
	cmd.Flags().String("workflow-id", "", "")
	cmd.Flags().String("run-id", "", "")
	cmd.Flags().String("block-id", "", "")
	cmd.Flags().String("step-id", "", "")
	cmd.Flags().String("iteration-key", "", "")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "")
	decisionFlag := newEnumStringFlagValue("--decision-status", "pending", "approved", "rejected", "decided", "all")
	_ = decisionFlag.Set("pending")
	cmd.Flags().Var(decisionFlag, "decision-status", "")
	cmd.Flags().String("before", "", "")
	cmd.Flags().String("after", "", "")
	cmd.MarkFlagsMutuallyExclusive("before", "after")
	return cmd
}

func TestReviewsListPassesDecisionFlag(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []any{},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	cmd := newReviewsListTestCmd()
	if err := cmd.Flags().Set("decision-status", "all"); err != nil {
		t.Fatal(err)
	}

	captureStd(t, func() {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("reviews list: %v", err)
		}
	})
	if !strings.Contains(seenQuery, "decision_status=all") {
		t.Fatalf("expected decision_status=all in query, got %q", seenQuery)
	}
}

func TestReviewsListDefaultsDecisionStatusToPending(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []any{},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	cmd := newReviewsListTestCmd()
	captureStd(t, func() {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("reviews list: %v", err)
		}
	})
	if !strings.Contains(seenQuery, "decision_status=pending") {
		t.Fatalf("expected decision_status=pending by default, got %q", seenQuery)
	}
}

func TestReviewsApproveTableShowsResumeStatusAndError(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	resumeErr := "Workflow run not found for run_id=run_x" +
		" — temporal returned NotFound — reconcile loop will retry — please retry shortly"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"resume_status":     "pending",
			"resume_error":      resumeErr,
			"review":            reviewOverlayBody(reviewDecisionBody("approved", reviewTestVersionID)),
		})
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	cmd := newApproveTestCmd()
	cmd.PersistentFlags().String("output", "table", "")
	if err := cmd.Flags().Set("version-id", reviewTestVersionID); err != nil {
		t.Fatal(err)
	}
	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"rev_1"}); err != nil {
			t.Fatalf("reviews approve: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	for _, want := range []string{
		"SUBMISSION", "RESUME_STATUS", "RESUME_ERROR",
		"accepted", "pending",
		"Workflow run not found for run_id=run_x",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in approve table output:\n%s", want, stdout)
		}
	}
	// The raw resume_error string is >60 chars; cells must be truncated.
	if strings.Contains(stdout, resumeErr) {
		t.Fatalf("expected long resume_error to be truncated in table view, got full string:\n%s", stdout)
	}
	if !strings.Contains(stdout, "...") {
		t.Fatalf("expected truncation ellipsis in table output:\n%s", stdout)
	}
}

func TestReviewsGetTableShowsDecisionColumnsForDecidedOverlay(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(reviewDecisionBody("approved", reviewTestVersionID)))
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsReviewsGetCmd.RunE(workflowsReviewsGetCmd, []string{"rev_1"}); err != nil {
			t.Fatalf("reviews get: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"VERDICT", "DECIDED_VERSION_ID", "approved", reviewTestVersionID[:12]} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in get table output:\n%s", want, stdout)
		}
	}
}

func TestReviewsGetTableShowsEmptyDecisionForOpenOverlay(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(nil))
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsReviewsGetCmd.RunE(workflowsReviewsGetCmd, []string{"rev_1"}); err != nil {
			t.Fatalf("reviews get: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"VERDICT", "DECIDED_VERSION_ID"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q header in get table output:\n%s", want, stdout)
		}
	}
	// Open overlay: VERDICT / DECIDED_VERSION_ID cells must be empty —
	// no nil / null / verdict strings leaking through. We scan the data
	// row's last two columns specifically so the substring check isn't
	// fooled by legitimate values elsewhere in the table (e.g. the
	// triggered_by kind "any_required_field_null").
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header + data row, got:\n%s", stdout)
	}
	dataRow := lines[len(lines)-1]
	dataCols := strings.Fields(dataRow)
	// reviewOverlayColumns has 10 columns; an open overlay should render
	// 8 populated cells (queue projection only), so Fields() returns 8.
	if len(dataCols) != 8 {
		t.Fatalf("expected 8 populated cells for open overlay (last 2 empty), got %d:\n%s", len(dataCols), stdout)
	}
	for _, bad := range []string{"<nil>", "approved", "rejected"} {
		if strings.Contains(stdout, bad) {
			t.Fatalf("open-overlay table output should not contain %q:\n%s", bad, stdout)
		}
	}
}

// reviewQueueRowJSON returns one well-formed queue row so the new
// pagination tests can drive the table renderer without each test having
// to hand-roll the same payload.
func reviewQueueRowJSON(blockRunID string) map[string]any {
	return map[string]any{
		"id":              blockRunID,
		"workflow_id":     "wf_1",
		"workflow_run_id": "run_1",
		"block_id":        "blk_1",
		"step_id":         "step_1",
		"parent_step_id":  nil,
		"iteration_key":   nil,
		"block_type":      "extract",
		"triggered_by":    map[string]any{"kind": "any_required_field_null"},
		"awaiting_since":  "2026-05-21T09:00:00Z",
		"priority":        0,
		"seed_version_id": reviewTestVersionID,
		"version_count":   1,
		"decision":        nil,
	}
}

// TestReviewsListPassesBeforeAndAfterCursors verifies the CLI threads
// --before / --after into the outbound HTTP query.
func TestReviewsListPassesBeforeAndAfterCursors(t *testing.T) {
	cases := []struct {
		name string
		flag string
		val  string
		want string
	}{
		{"before", "before", "brun_x", "before=brun_x"},
		{"after", "after", "brun_y", "after=brun_y"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			var seenQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seenQuery = r.URL.RawQuery
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data":          []any{},
					"list_metadata": map[string]any{"before": nil, "after": nil},
				})
			}))
			defer server.Close()
			setReviewsBaseURL(t, server.URL)

			cmd := newReviewsListTestCmd()
			if err := cmd.Flags().Set(tc.flag, tc.val); err != nil {
				t.Fatal(err)
			}
			captureStd(t, func() {
				if err := cmd.RunE(cmd, nil); err != nil {
					t.Fatalf("reviews list: %v", err)
				}
			})
			if !strings.Contains(seenQuery, tc.want) {
				t.Fatalf("expected %q in query, got %q", tc.want, seenQuery)
			}
		})
	}
}

// TestReviewsListRejectsBeforeAndAfterTogether verifies the CLI rejects the
// mutex violation at the command layer — no HTTP call is made.
func TestReviewsListRejectsBeforeAndAfterTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	cmd := newReviewsListTestCmd()
	if err := cmd.Flags().Set("before", "brun_x"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("after", "brun_y"); err != nil {
		t.Fatal(err)
	}
	var runErr error
	captureStd(t, func() {
		runErr = cmd.RunE(cmd, nil)
	})
	if runErr == nil || !strings.Contains(runErr.Error(), "mutually exclusive") {
		t.Fatalf("expected mutual-exclusion error, got %v", runErr)
	}
	if called {
		t.Fatalf("HTTP server was called despite mutex violation")
	}
}

// TestReviewsListTableShowsMoreResultsFooterWhenAfterCursorPresent
// verifies that the table renderer prints the cursor footer on stderr (so
// stdout stays pipeable) when list_metadata.after is set, and that the
// stdout payload contains the row itself.
func TestReviewsListTableShowsMoreResultsFooterWhenAfterCursorPresent(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []any{reviewQueueRowJSON("blockrun_1")},
			"list_metadata": map[string]any{"before": nil, "after": "brun_next"},
		})
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	cmd := newReviewsListTestCmd()
	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("reviews list: %v", err)
		}
	})
	if !strings.Contains(stdout, "blockrun_1") {
		t.Fatalf("expected row in stdout table, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "more results available; pass --after brun_next") {
		t.Fatalf("expected cursor footer on stderr, got %q", stderr)
	}
	// The footer must NOT leak into stdout — that would break piping the
	// table into jq / awk / column.
	if strings.Contains(stdout, "more results available") {
		t.Fatalf("footer leaked into stdout:\n%s", stdout)
	}
}

// TestReviewsListTableOmitsFooterWhenNoMorePages pins the symmetric case:
// when the backend returns list_metadata.after=null the CLI must not
// emit a stray footer.
func TestReviewsListTableOmitsFooterWhenNoMorePages(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []any{reviewQueueRowJSON("blockrun_1")},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	cmd := newReviewsListTestCmd()
	_, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("reviews list: %v", err)
		}
	})
	if strings.Contains(stderr, "more results available") {
		t.Fatalf("expected no footer when there are no more pages, got %q", stderr)
	}
}

// reviews schema --output table used to render nested maps via Go's default
// fmt.Sprintf %v, producing `map[k:v]` debug noise that's illegible in a table
// cell. Pin the new compact-JSON rendering so this can't regress.
func TestReviewsSchemaTableRendersStructuredCellsAsJSON(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := reviewOverlayBody(nil)
		body["block_type"] = "split"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, _ := captureStd(t, func() {
		if err := workflowsReviewsSchemaCmd.RunE(workflowsReviewsSchemaCmd, []string{"rev_1"}); err != nil {
			t.Fatalf("reviews schema: %v", err)
		}
	})
	if strings.Contains(stdout, "map[") {
		t.Fatalf("schema table cell contains Go-debug map output (`map[...]`):\n%s", stdout)
	}
	for _, want := range []string{`{"`, `"required":["documents"]`, `[1]`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected JSON-style cell containing %q:\n%s", want, stdout)
		}
	}
}
