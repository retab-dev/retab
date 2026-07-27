package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestReviewsApproveDefaultsToLatestVersion pins that omitting --version-id makes
// approve resolve the review's LATEST version (max created_at) via the versions
// list, instead of erroring "required". This removes the old 3-call dance
// (reviews list -> reviews versions list -> approve --version-id).
func TestReviewsApproveDefaultsToLatestVersion(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	const older = "rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA"
	const newer = "rvr_BBBBBBBBBBBBBBBBBBBBBBBBBB"

	var approveBody map[string]any
	var listedVersions bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "version"):
			listedVersions = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": older, "review_id": "rev_1", "created_at": "2026-06-01T00:00:00Z"},
					{"id": newer, "review_id": "rev_1", "created_at": "2026-06-02T00:00:00Z"},
				},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "approve"):
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &approveBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"submission_status": "accepted",
				"review":            reviewOverlayBody(reviewDecisionBody("approved", newer)),
			})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newApproveTestCmd()
	// no --version-id set
	if _, err := captureStdAndRun(t, func() error {
		return cmd.RunE(cmd, []string{"rev_1"})
	}); err != nil {
		t.Fatalf("reviews approve (default latest): %v", err)
	}
	if !listedVersions {
		t.Fatalf("expected a versions list lookup to resolve the latest version")
	}
	if approveBody["version_id"] != newer {
		t.Fatalf("approve used version_id %v, want latest %s", approveBody["version_id"], newer)
	}
}

// TestReviewsApproveNoVersionsIsCleanError pins that a review with zero versions
// gives a clear error rather than a confusing API failure.
func TestReviewsApproveNoVersionsIsCleanError(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []map[string]any{},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newApproveTestCmd()
	_, err := captureStdAndRun(t, func() error {
		return cmd.RunE(cmd, []string{"rev_1"})
	})
	if err == nil || !strings.Contains(err.Error(), "no versions") {
		t.Fatalf("expected 'no versions' error, got %v", err)
	}
}
