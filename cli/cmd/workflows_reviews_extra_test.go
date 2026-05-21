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
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "")
	decisionFlag := newEnumStringFlagValue("--decision", "none", "any")
	_ = decisionFlag.Set("none")
	cmd.Flags().Var(decisionFlag, "decision", "")
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
			"data":     []any{},
			"has_more": false,
		})
	}))
	defer server.Close()
	setReviewsBaseURL(t, server.URL)

	cmd := newReviewsListTestCmd()
	if err := cmd.Flags().Set("decision", "any"); err != nil {
		t.Fatal(err)
	}

	captureStd(t, func() {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("reviews list: %v", err)
		}
	})
	if !strings.Contains(seenQuery, "decision=any") {
		t.Fatalf("expected decision=any in query, got %q", seenQuery)
	}
}

func TestReviewsListDefaultsDecisionToNone(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":     []any{},
			"has_more": false,
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
	if !strings.Contains(seenQuery, "decision=none") {
		t.Fatalf("expected decision=none by default, got %q", seenQuery)
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
			"submission_status": "accepted_pending_resume",
			"resume_status":     "failed",
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
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews approve: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	for _, want := range []string{
		"SUBMISSION", "RESUME_STATUS", "RESUME_ERROR",
		"accepted_pending_resume", "failed",
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
		if err := workflowsReviewsGetCmd.RunE(workflowsReviewsGetCmd, []string{"run_1", "blk_1"}); err != nil {
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
		if err := workflowsReviewsGetCmd.RunE(workflowsReviewsGetCmd, []string{"run_1", "blk_1"}); err != nil {
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
	// reviewOverlayColumns has 9 columns; an open overlay should render
	// 7 populated cells (queue projection only), so Fields() returns 7.
	if len(dataCols) != 7 {
		t.Fatalf("expected 7 populated cells for open overlay (last 2 empty), got %d:\n%s", len(dataCols), stdout)
	}
	for _, bad := range []string{"<nil>", "approved", "rejected"} {
		if strings.Contains(stdout, bad) {
			t.Fatalf("open-overlay table output should not contain %q:\n%s", bad, stdout)
		}
	}
}
