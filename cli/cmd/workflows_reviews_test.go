package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func reviewQueueItemBody(reviewID string, runID string, blockID string, blockType string, status string, rev int, headSeq int, effectiveSeq *int) map[string]any {
	item := reviewOverlayBody(rev, status)
	item["_id"] = reviewID
	item["workflow_run_id"] = runID
	item["block_id"] = blockID
	item["block_run_id"] = reviewID
	item["block_type"] = blockType
	item["head_seq"] = headSeq
	if effectiveSeq != nil {
		item["effective_seq"] = *effectiveSeq
	}
	return item
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

func TestReviewsListTableUsesReviewSpecificColumns(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	effectiveSeq := 3
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{
				reviewQueueItemBody("review_1", "run_1", "blk_extract", "extract", "awaiting_review", 4, 2, nil),
				reviewQueueItemBody("review_2", "run_2", "blk_classify", "classify", "approved", 8, 3, &effectiveSeq),
			},
			"has_more": false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })
	workflowsReviewsListCmd.SetContext(nil)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsReviewsListCmd.RunE(workflowsReviewsListCmd, nil); err != nil {
			t.Fatalf("reviews list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{
		"REVIEW_ID", "RUN_ID", "BLOCK_ID", "BLOCK_TYPE", "STATUS", "REV", "HEAD", "EFFECTIVE",
		"review_1", "run_1", "blk_extract", "extract", "awaiting_review", "4", "2",
		"review_2", "run_2", "blk_classify", "classify", "approved", "8", "3",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
	if strings.Contains(strings.SplitN(stdout, "\n", 2)[0], "TYPE") && !strings.Contains(strings.SplitN(stdout, "\n", 2)[0], "BLOCK_TYPE") {
		t.Fatalf("table header should not use ambiguous TYPE column:\n%s", stdout)
	}
	if strings.Contains(stdout, "0x") {
		t.Fatalf("table should render effective_seq values, not pointer addresses:\n%s", stdout)
	}
	for _, line := range strings.Split(stdout, "\n") {
		if !strings.Contains(line, "review_2") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[len(fields)-1] != "3" {
			t.Fatalf("effective_seq should be the final cell rendered as 3, got row %q", line)
		}
		return
	}
	t.Fatalf("missing review_2 row:\n%s", stdout)
}

func TestReviewsListRejectsBadStatusEnum(t *testing.T) {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Var(newEnumStringFlagValue("--status", "awaiting_review", "approved", "rejected"), "status", "")
	if err := cmd.Flags().Set("status", "bogus"); err == nil {
		t.Fatal("expected --status to reject an unknown value")
	}
}

func TestReviewsEditAndApproveHelpExplainSnapshotVsReviewableValue(t *testing.T) {
	for _, cmd := range []*cobra.Command{workflowsReviewsEditCmd, workflowsReviewsApproveCmd} {
		if !strings.Contains(cmd.Long, "reviewable value") {
			t.Fatalf("%s help should explain reviewable values:\n%s", cmd.Use, cmd.Long)
		}
		if !strings.Contains(cmd.Long, "full block output snapshot") {
			t.Fatalf("%s help should distinguish full snapshots:\n%s", cmd.Use, cmd.Long)
		}
		if cmd.Flags().Lookup("reviewable-value-file") == nil {
			t.Fatalf("%s should define --reviewable-value-file", cmd.Use)
		}
	}
	if workflowsReviewsApproveCmd.Flags().Lookup("snapshot-file") == nil {
		t.Fatalf("%s should define --snapshot-file", workflowsReviewsApproveCmd.Use)
	}
	if strings.Contains(workflowsReviewsApproveCmd.Long, "edited-output-file") {
		t.Fatalf("approve help should use --snapshot-file, got:\n%s", workflowsReviewsApproveCmd.Long)
	}
}

func TestReviewsRejectHelpDoesNotPromiseRunCancellation(t *testing.T) {
	text := strings.ToLower(strings.Join([]string{
		workflowsReviewsRejectCmd.Short,
		workflowsReviewsRejectCmd.Long,
		workflowsReviewsRejectCmd.Example,
	}, "\n"))
	for _, stale := range []string{"cancel", "cancellation"} {
		if strings.Contains(text, stale) {
			t.Fatalf("reject help should use review/error wording, not %q:\n%s", stale, text)
		}
	}
	for _, want := range []string{"reject", "review", "error"} {
		if !strings.Contains(text, want) {
			t.Fatalf("reject help should mention %q:\n%s", want, text)
		}
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

func TestReviewsGetCommandHonorsOutputTable(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	effectiveSeq := 2
	body := reviewOverlayBody(3, "awaiting_review")
	body["effective_seq"] = effectiveSeq
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

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
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") || strings.Contains(stdout, `"versions"`) {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"REVIEW_ID", "RUN_ID", "BLOCK_ID", "BLOCK_TYPE", "STATUS", "REV", "HEAD", "EFFECTIVE", "blockrun_1", "run_1", "blk_1", "extract", "awaiting_review", "3", "0", "2"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}

func TestReviewsWaitStopsWhenRunTerminalBeforeOverlayExists(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var reviewGets int
	var runGets int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/workflows/reviews/run_1/blk_1":
			reviewGets++
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "review overlay not found"})
		case "/workflows/runs/run_1":
			runGets++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "run_1",
				"lifecycle": map[string]any{
					"status":  "error",
					"message": "block failed before review",
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newWaitTestCmd()
	if err := cmd.Flags().Set("timeout", "30"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("poll-interval", "30"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected wait to fail when the run is already terminal")
	}
	if !strings.Contains(err.Error(), "error") || !strings.Contains(err.Error(), "block failed before review") {
		t.Fatalf("error = %v", err)
	}
	if reviewGets != 1 || runGets != 1 {
		t.Fatalf("reviewGets=%d runGets=%d", reviewGets, runGets)
	}
}

func TestReviewsWaitFailsFastWhenRunIsMissing(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var reviewGets int
	var runGets int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/workflows/reviews/run_missing/blk_1":
			reviewGets++
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "review overlay not found"})
		case "/workflows/runs/run_missing":
			runGets++
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "run not found"})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newWaitTestCmd()
	err := cmd.RunE(cmd, []string{"run_missing", "blk_1"})
	if err == nil {
		t.Fatal("expected missing run to fail")
	}
	if !strings.Contains(err.Error(), "run_missing") || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v", err)
	}
	if reviewGets != 1 || runGets != 1 {
		t.Fatalf("reviewGets=%d runGets=%d", reviewGets, runGets)
	}
}

func TestReviewsWaitRejectsNonPositiveTimingFlags(t *testing.T) {
	cmd := newWaitTestCmd()

	for _, tc := range []struct {
		flag  string
		value string
	}{
		{flag: "timeout", value: "0"},
		{flag: "timeout", value: "-1"},
		{flag: "poll-interval", value: "0"},
		{flag: "poll-interval", value: "-1"},
	} {
		if err := cmd.Flags().Set(tc.flag, tc.value); err == nil {
			t.Fatalf("expected --%s=%s to fail", tc.flag, tc.value)
		}
	}
}

func TestReviewsClaimRejectsNonPositiveTTL(t *testing.T) {
	t.Cleanup(func() { _ = workflowsReviewsClaimCmd.Flags().Set("ttl-seconds", "900") })

	for _, value := range []string{"0", "-1", "29"} {
		if err := workflowsReviewsClaimCmd.Flags().Set("ttl-seconds", value); err == nil {
			t.Fatalf("expected --ttl-seconds=%s to fail", value)
		}
	}
}

func TestReviewsApproveSendsReviewableValueFile(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	valuePath := filepath.Join(t.TempDir(), "reviewable.json")
	if err := os.WriteFile(valuePath, []byte(`{"total":110,"currency":"USD"}`), 0o644); err != nil {
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

	cmd := newApproveTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("reviewable-value-file", valuePath); err != nil {
		t.Fatal(err)
	}
	captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews approve: %v", err)
		}
	})
	reviewableValue, ok := body["reviewable_value"].(map[string]any)
	if !ok {
		t.Fatalf("reviewable_value missing from body: %#v", body)
	}
	if reviewableValue["total"] != float64(110) || reviewableValue["currency"] != "USD" {
		t.Fatalf("reviewable_value = %#v", reviewableValue)
	}
	if _, ok := body["edited_output"]; ok {
		t.Fatalf("reviewable approval should not also send edited_output: %#v", body)
	}
}

func TestReviewsEditSendsReviewableValueFile(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	valuePath := filepath.Join(t.TempDir(), "reviewable.json")
	if err := os.WriteFile(valuePath, []byte(`{"category":"Invoice"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(1, "awaiting_review"))
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newEditTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("reviewable-value-file", valuePath); err != nil {
		t.Fatal(err)
	}
	captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews edit: %v", err)
		}
	})
	reviewableValue, ok := body["reviewable_value"].(map[string]any)
	if !ok {
		t.Fatalf("reviewable_value missing from body: %#v", body)
	}
	if reviewableValue["category"] != "Invoice" {
		t.Fatalf("reviewable_value = %#v", reviewableValue)
	}
	if _, ok := body["snapshot"]; ok {
		t.Fatalf("reviewable edit should not also send snapshot: %#v", body)
	}
}

func TestReviewsEditRejectsMutuallyExclusiveFilesBeforeVersionStampFetch(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	snapshotPath := filepath.Join(dir, "snapshot.json")
	reviewablePath := filepath.Join(dir, "reviewable.json")
	if err := os.WriteFile(snapshotPath, []byte(`{"output":{"total":110}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reviewablePath, []byte(`{"total":120}`), 0o644); err != nil {
		t.Fatal(err)
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(7, "awaiting_review"))
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newEditTestCmd()
	if err := cmd.Flags().Set("snapshot-file", snapshotPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("reviewable-value-file", reviewablePath); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected mutually exclusive file flags to fail")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no network request before local validation, got %d", requests)
	}
}

func TestReviewsEditReadsFileBeforeVersionStampFetch(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(7, "awaiting_review"))
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newEditTestCmd()
	if err := cmd.Flags().Set("reviewable-value-file", filepath.Join(t.TempDir(), "missing.json")); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected missing file to fail")
	}
	if !strings.Contains(err.Error(), "--reviewable-value-file") {
		t.Fatalf("error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no network request before local file validation, got %d", requests)
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

func TestReviewsApproveTableRendersConciseDecisionResponse(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	body := reviewOverlayBody(9, "approved")
	body["effective_seq"] = 0
	body["versions"] = []any{map[string]any{
		"seq":      0,
		"snapshot": map[string]any{"huge_nested_payload": strings.Repeat("x", 4096)},
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"overlay":           body,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newApproveTestCmd()
	cmd.PersistentFlags().String("output", "table", "")
	if err := cmd.Flags().Set("version-stamp", "8"); err != nil {
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
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") || strings.Contains(stdout, `"versions"`) || strings.Contains(stdout, "huge_nested_payload") {
		t.Fatalf("expected concise table output, got:\n%s", stdout)
	}
	for _, want := range []string{"SUBMISSION", "REVIEW_ID", "RUN_ID", "BLOCK_ID", "STATUS", "REV", "accepted", "blockrun_1", "run_1", "blk_1", "approved", "9"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
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
	if strings.Contains(stderr, "no protection") {
		t.Fatalf("auto-fetch warning should not incorrectly claim CAS is disabled: %q", stderr)
	}
	if !strings.Contains(stderr, "reviews get") {
		t.Fatalf("auto-fetch warning should point scripts to reviews get, got: %q", stderr)
	}
}

func TestReviewsApproveRejectsMutuallyExclusiveFilesBeforeVersionStampFetch(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	editedPath := filepath.Join(dir, "edited.json")
	reviewablePath := filepath.Join(dir, "reviewable.json")
	if err := os.WriteFile(editedPath, []byte(`{"total":110}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reviewablePath, []byte(`{"total":120}`), 0o644); err != nil {
		t.Fatal(err)
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(7, "awaiting_review"))
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := newApproveTestCmd()
	if err := cmd.Flags().Set("snapshot-file", editedPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("reviewable-value-file", reviewablePath); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected mutually exclusive file flags to fail")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no network request before local validation, got %d", requests)
	}
}

func TestReviewsApproveReadsSnapshotBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	cmd := newApproveTestCmd()
	if err := cmd.Flags().Set("snapshot-file", filepath.Join(t.TempDir(), "missing.json")); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("version-stamp", "1"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected missing snapshot error")
	}
	if !strings.Contains(err.Error(), "--snapshot-file") {
		t.Fatalf("error %q does not mention --snapshot-file", err.Error())
	}
	if strings.Contains(err.Error(), "credentials") {
		t.Fatalf("error %q checked credentials before reading --snapshot-file", err.Error())
	}
}

func TestReviewsRejectRequiresReason(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_BASE_URL", "")
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
	if strings.Contains(err.Error(), "credentials") {
		t.Fatalf("error %q checked credentials before validating --reason", err.Error())
	}
}

func TestReviewsEscalateIsHiddenAndDisabled(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		t.Fatalf("reviews escalate should not reach backend, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if !workflowsReviewsEscalateCmd.Hidden {
		t.Fatal("reviews escalate should be hidden from help")
	}
	if strings.Contains(workflowsReviewsCmd.Long, "escalate") {
		t.Fatalf("reviews help should not advertise escalation:\n%s", workflowsReviewsCmd.Long)
	}

	cmd := newEscalateTestCmd()
	if err := cmd.Flags().Set("version-stamp", "2"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("reason", "needs senior sign-off"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("escalate-to", "queue_senior"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected disabled escalation command to fail locally")
	}
	for _, want := range []string{"not supported", "approve", "reject", "edit"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q should contain %q", err.Error(), want)
		}
	}
	if hits != 0 {
		t.Fatalf("backend hits = %d", hits)
	}
}

func newApproveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "approve", RunE: workflowsReviewsApproveCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("snapshot-file", "", "")
	cmd.Flags().String("edited-output-file", "", "")
	cmd.Flags().String("reviewable-value-file", "", "")
	cmd.Flags().Int("on-seq", 0, "")
	cmd.Flags().Int("effective-seq", 0, "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func newEditTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "edit", RunE: workflowsReviewsEditCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("snapshot-file", "", "")
	cmd.Flags().String("reviewable-value-file", "", "")
	cmd.Flags().Var(newEnumStringFlagValue("--origin", "human_edit", "agent_edit"), "origin", "")
	cmd.Flags().String("note", "", "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func TestReviewEnumFlagsShowAllowedValues(t *testing.T) {
	statusFlag := newEnumStringFlagValue("--status", "awaiting_review", "approved", "rejected")
	err := statusFlag.Set("bogus")
	if err == nil {
		t.Fatal("expected invalid status to fail")
	}
	if !strings.Contains(err.Error(), "awaiting_review | approved | rejected") {
		t.Fatalf("status error should show allowed values, got: %v", err)
	}

	originFlag := newEnumStringFlagValue("--origin", "human_edit", "agent_edit")
	err = originFlag.Set("typo")
	if err == nil {
		t.Fatal("expected invalid origin to fail")
	}
	if !strings.Contains(err.Error(), "human_edit | agent_edit") {
		t.Fatalf("origin error should show allowed values, got: %v", err)
	}
}

func newEscalateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "escalate", RunE: workflowsReviewsEscalateCmd.RunE}
	cmd.Flags().Int("version-stamp", 0, "")
	cmd.Flags().String("reason", "", "")
	cmd.Flags().String("escalate-to", "", "")
	cmd.Flags().String("command-id", "", "")
	return cmd
}

func newWaitTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "wait", RunE: workflowsReviewsWaitCmd.RunE}
	cmd.Flags().Var(&positiveIntFlagValue{value: "120"}, "timeout", "")
	cmd.Flags().Var(&positiveIntFlagValue{value: "2"}, "poll-interval", "")
	return cmd
}
