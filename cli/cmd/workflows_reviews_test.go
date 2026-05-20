package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestReviewsApproveAcceptsInlineEditedOutput(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if err := cmd.Flags().Set("edited-output-json", `{"splits":[{"name":"invoice","pages":[1]}]}`); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
		t.Fatalf("reviews approve: %v", err)
	}
	edited, ok := body["edited_output"].(map[string]any)
	if !ok {
		t.Fatalf("edited_output missing: %#v", body)
	}
	splits, ok := edited["splits"].([]any)
	if !ok || len(splits) != 1 {
		t.Fatalf("edited output = %#v", edited)
	}
}

func TestReviewsSplitApproveSendsReviewableValue(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newSplitApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "booking confirmation=1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "legal-mentions=2-3"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
		t.Fatalf("reviews split approve: %v", err)
	}

	reviewable, ok := body["reviewable_value"].(map[string]any)
	if !ok {
		t.Fatalf("reviewable_value missing: %#v", body)
	}
	splits, ok := reviewable["splits"].([]any)
	if !ok || len(splits) != 2 {
		t.Fatalf("splits = %#v", reviewable["splits"])
	}
	second := splits[1].(map[string]any)
	pages := second["pages"].([]any)
	if pages[0] != float64(2) || pages[1] != float64(3) {
		t.Fatalf("pages = %#v", pages)
	}
	if _, ok := body["edited_output"]; ok {
		t.Fatalf("typed command must not send edited_output: %#v", body)
	}
}

func TestReviewsSplitApproveRejectsDuplicatePageInSet(t *testing.T) {
	cmd := newSplitApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "invoice=1,1"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected duplicate page to fail")
	}
	if !strings.Contains(err.Error(), "duplicate page 1") {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsSplitApproveRejectsOverlappingPagesAcrossSets(t *testing.T) {
	cmd := newSplitApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "booking confirmation=1-2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "legal-mentions=2-3"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected overlapping pages to fail")
	}
	if !strings.Contains(err.Error(), `page 2 appears in both "booking confirmation" and "legal-mentions"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsSplitApproveRejectsDuplicateLabels(t *testing.T) {
	cmd := newSplitApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "invoice=1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "invoice=2"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected duplicate split labels to fail")
	}
	if !strings.Contains(err.Error(), `duplicate split label "invoice"`) {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(err.Error(), "case-sensitive") || !strings.Contains(err.Error(), "server") {
		t.Fatalf("error should explain exact-label handling, got %v", err)
	}
}

func TestReviewsSplitApproveRejectsDuplicateLabelsFromSplitsJSON(t *testing.T) {
	cmd := newSplitApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("splits-json", `{"splits":[{"name":"invoice","pages":[1]},{"name":"invoice","pages":[2]}]}`); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected duplicate split labels to fail")
	}
	if !strings.Contains(err.Error(), `duplicate split label "invoice"`) {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(err.Error(), "case-sensitive") || !strings.Contains(err.Error(), "server") {
		t.Fatalf("error should explain exact-label handling, got %v", err)
	}
}

func TestReviewsSplitApproveKeepsSetLabelsExact(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newSplitApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", " Invoice Total =1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
		t.Fatalf("reviews split approve: %v", err)
	}

	reviewable, ok := body["reviewable_value"].(map[string]any)
	if !ok {
		t.Fatalf("reviewable_value missing: %#v", body)
	}
	splits, ok := reviewable["splits"].([]any)
	if !ok || len(splits) != 1 {
		t.Fatalf("splits = %#v", reviewable["splits"])
	}
	first := splits[0].(map[string]any)
	if first["name"] != " Invoice Total " {
		t.Fatalf("split label was normalized: %#v", first["name"])
	}
}

func TestReviewsSplitApproveRejectsInvalidSplitsJSONShape(t *testing.T) {
	cmd := newSplitApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("splits-json", `{"splits":[{"name":"invoice","pages":[1,1]}]}`); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected invalid splits JSON to fail")
	}
	if !strings.Contains(err.Error(), "splits[0].pages: duplicate page 1") {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsClassifierApproveSendsCategoryReviewableValue(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newClassifierApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("category", "Invoice"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
		t.Fatalf("reviews classifier approve: %v", err)
	}
	reviewable := body["reviewable_value"].(map[string]any)
	if reviewable["category"] != "Invoice" {
		t.Fatalf("reviewable_value = %#v", reviewable)
	}
}

func TestReviewsForEachApproveSendsPartitionReviewableValue(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "legal-mentions=1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "booking confirmation=2-3"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"}); err != nil {
		t.Fatalf("reviews for-each approve: %v", err)
	}

	reviewable, ok := body["reviewable_value"].(map[string]any)
	if !ok {
		t.Fatalf("reviewable_value missing: %#v", body)
	}
	chunks, ok := reviewable["chunks"].([]any)
	if !ok || len(chunks) != 2 {
		t.Fatalf("chunks = %#v", reviewable["chunks"])
	}
	first := chunks[0].(map[string]any)
	if first["key"] != "legal-mentions" {
		t.Fatalf("chunk key = %#v", first["key"])
	}
	second := chunks[1].(map[string]any)
	pages := second["pages"].([]any)
	if pages[0] != float64(2) || pages[1] != float64(3) {
		t.Fatalf("pages = %#v", pages)
	}
	if _, ok := body["edited_output"]; ok {
		t.Fatalf("typed command must not send edited_output: %#v", body)
	}
}

func TestReviewsForEachApproveAllowsOverlappingPagesAcrossKeys(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "10*310365=1-2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", "10*315944=2-3"); err != nil {
		t.Fatal(err)
	}

	if err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"}); err != nil {
		t.Fatalf("reviews for-each approve: %v", err)
	}

	reviewable, ok := body["reviewable_value"].(map[string]any)
	if !ok {
		t.Fatalf("reviewable_value missing: %#v", body)
	}
	chunks, ok := reviewable["chunks"].([]any)
	if !ok || len(chunks) != 2 {
		t.Fatalf("chunks = %#v", reviewable["chunks"])
	}
	second := chunks[1].(map[string]any)
	pages := second["pages"].([]any)
	if pages[0] != float64(2) || pages[1] != float64(3) {
		t.Fatalf("pages = %#v", pages)
	}
}

func TestReviewsForEachApproveRejectsDuplicateKeysFromChunksJSON(t *testing.T) {
	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("chunks-json", `{"chunks":[{"key":"invoice","pages":[1]},{"key":"invoice","pages":[2]}]}`); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"})
	if err == nil {
		t.Fatal("expected duplicate partition keys to fail")
	}
	if !strings.Contains(err.Error(), `duplicate partition key "invoice"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsForEachApproveRejectsBlankChunksJSON(t *testing.T) {
	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("chunks-json", ""); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"})
	if err == nil {
		t.Fatal("expected blank chunks-json to fail")
	}
	if !strings.Contains(err.Error(), "--chunks-json must not be blank") {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsForEachApproveRejectsBlankValueFile(t *testing.T) {
	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("value-file", ""); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"})
	if err == nil {
		t.Fatal("expected blank value-file to fail")
	}
	if !strings.Contains(err.Error(), "--value-file must not be blank") {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsForEachApproveRejectsWhitespaceNormalizedDuplicateKeys(t *testing.T) {
	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("chunks-json", `{"chunks":[{"key":"invoice","pages":[1]},{"key":" invoice ","pages":[2]}]}`); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"})
	if err == nil {
		t.Fatal("expected normalized duplicate partition keys to fail")
	}
	if !strings.Contains(err.Error(), `duplicate partition key "invoice"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsForEachApproveTrimsSetKeys(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("set", " invoice =1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"}); err != nil {
		t.Fatalf("reviews for-each approve: %v", err)
	}

	reviewable := body["reviewable_value"].(map[string]any)
	chunks := reviewable["chunks"].([]any)
	first := chunks[0].(map[string]any)
	if first["key"] != "invoice" {
		t.Fatalf("chunk key was not trimmed: %#v", first["key"])
	}
}

func TestReviewsForEachApproveAcceptsValueFile(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	tempFile := t.TempDir() + "/chunks.json"
	if err := os.WriteFile(tempFile, []byte(`{"chunks":[{"key":"a=b","pages":[1]}]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("value-file", tempFile); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"}); err != nil {
		t.Fatalf("reviews for-each approve: %v", err)
	}

	reviewable := body["reviewable_value"].(map[string]any)
	chunks := reviewable["chunks"].([]any)
	first := chunks[0].(map[string]any)
	if first["key"] != "a=b" {
		t.Fatalf("chunk key = %#v", first["key"])
	}
}

func TestReviewsForEachApproveAllowsEmptyChunks(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cmd := newForEachApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("chunks-json", `{"chunks":[]}`); err != nil {
		t.Fatal(err)
	}

	if err := cmd.RunE(cmd, []string{"run_1", "blk_for_each"}); err != nil {
		t.Fatalf("reviews for-each approve: %v", err)
	}

	reviewable, ok := body["reviewable_value"].(map[string]any)
	if !ok {
		t.Fatalf("reviewable_value missing: %#v", body)
	}
	chunks, ok := reviewable["chunks"].([]any)
	if !ok || len(chunks) != 0 {
		t.Fatalf("chunks = %#v", reviewable["chunks"])
	}
}

func TestReviewsApproveRejectsFileAndInlineEditedOutputTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cmd := newApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("edited-output-file", "fixed.json"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("edited-output-json", `{"splits":[]}`); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected mutually exclusive edited output flags to fail")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsApproveRejectsBlankEditedOutputFlags(t *testing.T) {
	for _, tc := range []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "file", flag: "edited-output-file", wantError: "--edited-output-file must not be blank"},
		{name: "json", flag: "edited-output-json", wantError: "--edited-output-json must not be blank"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			cmd := newApproveTestCmd()
			if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set(tc.flag, ""); err != nil {
				t.Fatal(err)
			}
			err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
			if err == nil {
				t.Fatal("expected blank edited output flag to fail")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error = %v", err)
			}
		})
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

func TestReviewsExtractApproveRejectsBlankValueFlags(t *testing.T) {
	for _, tc := range []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "file", flag: "value-file", wantError: "--value-file must not be blank"},
		{name: "json", flag: "value-json", wantError: "--value-json must not be blank"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newExtractApproveTestCmd()
			if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set(tc.flag, ""); err != nil {
				t.Fatal(err)
			}
			err := cmd.RunE(cmd, []string{"run_1", "blk_extract"})
			if err == nil {
				t.Fatal("expected blank value flag to fail")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestReviewsSplitApproveRejectsBlankValueFlags(t *testing.T) {
	for _, tc := range []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "file", flag: "value-file", wantError: "--value-file must not be blank"},
		{name: "json", flag: "splits-json", wantError: "--splits-json must not be blank"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newSplitApproveTestCmd()
			if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set(tc.flag, ""); err != nil {
				t.Fatal(err)
			}
			err := cmd.RunE(cmd, []string{"run_1", "blk_split"})
			if err == nil {
				t.Fatal("expected blank value flag to fail")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func newApproveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "approve", RunE: workflowsReviewsApproveCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("edited-output-file", "", "")
	cmd.Flags().String("edited-output-json", "", "")
	cmd.Flags().Int("on-seq", 0, "")
	cmd.Flags().Int("effective-seq", 0, "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func newExtractApproveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "approve", RunE: workflowsReviewsExtractApproveCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("value-file", "", "")
	cmd.Flags().String("value-json", "", "")
	cmd.Flags().StringArray("set", nil, "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func newSplitApproveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "approve", RunE: workflowsReviewsSplitApproveCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("value-file", "", "")
	cmd.Flags().String("splits-json", "", "")
	cmd.Flags().StringArray("set", nil, "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func newClassifierApproveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "approve", RunE: workflowsReviewsClassifierApproveCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("category", "", "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func newForEachApproveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "approve", RunE: workflowsReviewsForEachApproveCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("value-file", "", "")
	cmd.Flags().String("chunks-json", "", "")
	cmd.Flags().StringArray("set", nil, "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func TestReviewsEditAcceptsInlineSnapshot(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(2, "awaiting_review"))
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newEditTestCmd()
	if err := cmd.Flags().Set("version-stamp", "1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("snapshot-json", `{"splits":[{"name":"invoice","pages":[1]}]}`); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
		t.Fatalf("reviews edit: %v", err)
	}
	if body["version_stamp"] != float64(1) {
		t.Fatalf("version_stamp = %#v", body["version_stamp"])
	}
	snapshot, ok := body["snapshot"].(map[string]any)
	if !ok {
		t.Fatalf("snapshot missing: %#v", body)
	}
	if _, ok := snapshot["splits"].([]any); !ok {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestReviewsEditRequiresSnapshotFileOrInlineJSON(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cmd := newEditTestCmd()
	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected missing snapshot to fail")
	}
	if !strings.Contains(err.Error(), "--snapshot-file") || !strings.Contains(err.Error(), "--snapshot-json") {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsEditRejectsFileAndInlineSnapshotTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cmd := newEditTestCmd()
	if err := cmd.Flags().Set("snapshot-file", "corrected.json"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("snapshot-json", `{"splits":[]}`); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected mutually exclusive snapshot flags to fail")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %v", err)
	}
}

func TestReviewsEditRejectsBlankSnapshotFlags(t *testing.T) {
	for _, tc := range []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "file", flag: "snapshot-file", wantError: "--snapshot-file must not be blank"},
		{name: "json", flag: "snapshot-json", wantError: "--snapshot-json must not be blank"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			cmd := newEditTestCmd()
			if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set(tc.flag, ""); err != nil {
				t.Fatal(err)
			}
			err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
			if err == nil {
				t.Fatal("expected blank snapshot flag to fail")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func newEditTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "edit", RunE: workflowsReviewsEditCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("snapshot-file", "", "")
	cmd.Flags().String("snapshot-json", "", "")
	cmd.Flags().Var(newEnumStringFlagValue("--origin", "human_edit", "agent_edit"), "origin", "")
	cmd.Flags().String("note", "", "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}
