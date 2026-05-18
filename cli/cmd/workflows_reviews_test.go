package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func reviewOverlayBody(rev int, status string) map[string]any {
	return map[string]any{
		"_id":                 "blockrun_1",
		"organization_id":     "org_1",
		"workflow_id":         "wf_1",
		"workflow_version_id": "wfv_1",
		"workflow_run_id":     "run_1",
		"block_id":            "blk_1",
		"block_run_id":        "blockrun_1",
		"block_type":          "extract",
		"triggered_by":        map[string]any{"kind": "any_required_field_null"},
		"status":              status,
		"awaiting_since":      "2026-05-18T09:00:00Z",
		"priority":            0,
		"rev":                 rev,
		"versions": []any{map[string]any{
			"seq": 0, "parent_seq": nil,
			"author":         map[string]any{"kind": "model", "id": "m", "display_name": "Model"},
			"origin":         "model_output",
			"snapshot":       map[string]any{"total": 100},
			"content_sha256": "abc",
			"created_at":     "2026-05-18T09:00:00Z",
		}},
		"decisions": []any{},
		"audit":     []any{},
		"head_seq":  0,
	}
}

func TestReviewsListCommand(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenPath, seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath, seenQuery = r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":     []any{reviewOverlayBody(0, "awaiting_review")},
			"has_more": false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "list", RunE: workflowsReviewsListCmd.RunE}
	cmd.Flags().String("workflow-id", "", "")
	cmd.Flags().Var(newEnumStringFlagValue("--status", "awaiting_review", "approved", "rejected"), "status", "")
	cmd.Flags().Bool("mine", false, "")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "")
	if err := cmd.Flags().Set("workflow-id", "wf_1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("mine", "true"); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("reviews list: %v", err)
		}
	})
	if seenPath != "/workflows/reviews" {
		t.Fatalf("path = %s", seenPath)
	}
	if !strings.Contains(seenQuery, "workflow_id=wf_1") || !strings.Contains(seenQuery, "mine=true") {
		t.Fatalf("query = %s", seenQuery)
	}
	if !strings.Contains(stdout, "blockrun_1") {
		t.Fatalf("stdout = %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
}

func TestReviewsListRejectsBadStatusEnum(t *testing.T) {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Var(newEnumStringFlagValue("--status", "awaiting_review", "approved", "rejected"), "status", "")
	if err := cmd.Flags().Set("status", "bogus"); err == nil {
		t.Fatal("expected --status to reject an unknown value")
	}
}

func TestReviewsGetCommand(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(3, "awaiting_review"))
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "get", RunE: workflowsReviewsGetCmd.RunE}
	stdout, _ := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews get: %v", err)
		}
	})
	if seenPath != "/workflows/reviews/run_1/blk_1" {
		t.Fatalf("path = %s", seenPath)
	}
	if !strings.Contains(stdout, `"rev": 3`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestReviewsApproveSendsExplicitVersionStamp(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"overlay":           reviewOverlayBody(1, "approved"),
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews approve: %v", err)
		}
	})
	if seenPath != "/workflows/reviews/run_1/blk_1/decision" {
		t.Fatalf("path = %s", seenPath)
	}
	if body["verdict"] != "approved" {
		t.Fatalf("verdict = %#v", body["verdict"])
	}
	if body["version_stamp"] != float64(2) {
		t.Fatalf("version_stamp = %#v", body["version_stamp"])
	}
	// An explicit --version-stamp must NOT trigger the auto-fetch warning.
	if strings.Contains(stderr, "warning") {
		t.Fatalf("unexpected stderr warning: %s", stderr)
	}
	if !strings.Contains(stdout, "accepted") {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestReviewsApproveWithoutVersionStampFetchesAndWarns(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			// the auto-fetch read — current rev is 7
			_ = json.NewEncoder(w).Encode(reviewOverlayBody(7, "awaiting_review"))
			return
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"overlay":           reviewOverlayBody(8, "approved"),
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newApproveTestCmd() // --version-stamp left unset
	_, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews approve: %v", err)
		}
	})
	if body["version_stamp"] != float64(7) {
		t.Fatalf("expected auto-fetched version_stamp 7, got %#v", body["version_stamp"])
	}
	if !strings.Contains(stderr, "warning") || !strings.Contains(stderr, "rev 7") {
		t.Fatalf("expected an auto-fetch warning on stderr, got: %q", stderr)
	}
}

func TestReviewsRejectRequiresReason(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "reject", RunE: workflowsReviewsRejectCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("reason", "", "")
	cmd.Flags().String("command-id", "", "")
	// reason left blank — RunE must reject before any request.
	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected reject to fail without --reason")
	}
	if !strings.Contains(err.Error(), "reason") {
		t.Fatalf("error = %v", err)
	}
}

func newApproveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "approve", RunE: workflowsReviewsApproveCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("edited-output-file", "", "")
	cmd.Flags().Int("on-seq", 0, "")
	cmd.Flags().Int("effective-seq", 0, "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}
