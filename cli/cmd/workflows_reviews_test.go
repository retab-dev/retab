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

const reviewTestVersionID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func reviewVersionBody(parentID any, snapshot map[string]any) map[string]any {
	return map[string]any{
		"parent_id":  parentID,
		"author":     map[string]any{"kind": "model", "id": "m", "display_name": "Model"},
		"snapshot":   snapshot,
		"created_at": "2026-05-18T09:00:00Z",
	}
}

func reviewDecisionBody(verdict string, versionID string) map[string]any {
	return map[string]any{
		"verdict":    verdict,
		"version_id": versionID,
		"author":     map[string]any{"kind": "human", "id": "user_1", "display_name": "Reviewer"},
		"decided_at": "2026-05-18T09:10:00Z",
	}
}

func reviewOverlayBody(decision map[string]any) map[string]any {
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
		"awaiting_since":      "2026-05-18T09:00:00Z",
		"priority":            0,
		"versions_by_id": map[string]any{
			reviewTestVersionID: reviewVersionBody(nil, map[string]any{"total": 100}),
		},
		"decision": decision,
	}
}

func reviewQueueItemBody(reviewID string, runID string, blockID string, blockType string) map[string]any {
	return map[string]any{
		"_id":                 reviewID,
		"organization_id":     "org_1",
		"workflow_id":         "wf_1",
		"workflow_version_id": "wfv_1",
		"workflow_run_id":     runID,
		"block_id":            blockID,
		"block_run_id":        reviewID,
		"block_type":          blockType,
		"triggered_by":        map[string]any{"kind": "always"},
		"awaiting_since":      "2026-05-18T09:00:00Z",
		"priority":            0,
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
			"data":     []any{reviewQueueItemBody("blockrun_1", "run_1", "blk_1", "extract")},
			"has_more": false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "list", RunE: workflowsReviewsListCmd.RunE}
	cmd.Flags().String("workflow-id", "", "")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 200}, "limit", "")
	if err := cmd.Flags().Set("workflow-id", "wf_1"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("limit", "50"); err != nil {
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
	if !strings.Contains(seenQuery, "workflow_id=wf_1") || !strings.Contains(seenQuery, "limit=50") {
		t.Fatalf("query = %s", seenQuery)
	}
	if strings.Contains(seenQuery, "mine=") || strings.Contains(seenQuery, "status=") {
		t.Fatalf("query should not contain legacy filters: %s", seenQuery)
	}
	if !strings.Contains(stdout, "blockrun_1") {
		t.Fatalf("stdout = %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
}

func TestReviewsListTableUsesPureQueueColumns(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		second := reviewQueueItemBody("review_2", "run_2", "blk_classify", "classify")
		second["triggered_by"] = map[string]any{"kind": "category_in"}
		second["awaiting_since"] = "2026-05-18T09:05:00Z"
		second["priority"] = 7
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{
				reviewQueueItemBody("review_1", "run_1", "blk_extract", "extract"),
				second,
			},
			"has_more": false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

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
		"REVIEW_ID", "RUN_ID", "BLOCK_ID", "BLOCK_TYPE", "AWAITING_SINCE", "PRIORITY", "TRIGGERED_BY",
		"review_1", "run_1", "blk_extract", "extract",
		"review_2", "run_2", "blk_classify", "classify",
		"2026-05-18T09:00:00Z", "always",
		"2026-05-18T09:05:00Z", "7", "category_in",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
	header := strings.Fields(strings.SplitN(stdout, "\n", 2)[0])
	for _, stale := range []string{"STATUS", "REV", "HEAD", "EFFECTIVE", "VERSION_ID", "DECISION"} {
		if containsString(header, stale) {
			t.Fatalf("table output contains stale %q column:\n%s", stale, stdout)
		}
	}
}

func TestReviewsHelpUsesVersionIDsAndNoReviewableProjections(t *testing.T) {
	for _, cmd := range []*cobra.Command{workflowsReviewsListCmd, workflowsReviewsVersionsCreateCmd, workflowsReviewsApproveCmd, workflowsReviewsRejectCmd} {
		text := strings.Join([]string{cmd.Short, cmd.Long, cmd.Example}, "\n")
		for _, stale := range []string{"reviewable value", "reviewable-value", "version_stamp", "version-stamp", "head_seq", "effective_seq", "audit history", "audit log"} {
			if strings.Contains(text, stale) {
				t.Fatalf("%s help contains stale %q:\n%s", cmd.Use, stale, text)
			}
		}
	}
	if workflowsReviewsApproveCmd.Flags().Lookup("version-id") == nil {
		t.Fatalf("%s should define --version-id", workflowsReviewsApproveCmd.Use)
	}
	if workflowsReviewsRejectCmd.Flags().Lookup("version-id") == nil {
		t.Fatalf("%s should define --version-id", workflowsReviewsRejectCmd.Use)
	}
	if workflowsReviewsVersionsCreateCmd.Flags().Lookup("parent-id") == nil {
		t.Fatalf("%s should define --parent-id", workflowsReviewsVersionsCreateCmd.Use)
	}
	if workflowsReviewsVersionsCreateCmd.Flags().Lookup("from-version") != nil {
		t.Fatalf("%s should not define legacy --from-version", workflowsReviewsVersionsCreateCmd.Use)
	}
	if workflowsReviewsVersionsCreateCmd.Flags().Lookup("snapshot-file") == nil {
		t.Fatalf("%s should define --snapshot-file", workflowsReviewsVersionsCreateCmd.Use)
	}
	for _, child := range workflowsReviewsCmd.Commands() {
		if child.Name() == "edit" {
			t.Fatalf("reviews edit command should not exist after create-version cutover")
		}
	}
	if workflowsReviewsVersionsCmd.Commands()[0].Name() != "create" {
		t.Fatalf("reviews versions should expose create, got %s", workflowsReviewsVersionsCmd.Commands()[0].Name())
	}
}

func TestReviewsVersionsCreateHelpExplainsSnapshotShapes(t *testing.T) {
	text := strings.Join([]string{
		workflowsReviewsVersionsCreateCmd.Long,
		workflowsReviewsVersionsCreateCmd.Example,
		workflowsReviewsVersionsCreateCmd.Flags().Lookup("snapshot-file").Usage,
	}, "\n")
	for _, want := range []string{
		"Snapshot shapes are block-specific",
		"extract:",
		"classifier:",
		"split:",
		"for_each:",
		`"category"`,
		`"documents"`,
		`"partitions"`,
		"--snapshot-file -",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("create version help should explain snapshot shape %q, got:\n%s", want, text)
		}
	}
	if strings.Contains(text, "reviewable value") || strings.Contains(text, "reviewable-value") {
		t.Fatalf("create version help should not reintroduce reviewable value wording:\n%s", text)
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
	for _, want := range []string{"reject", "review", "downstream"} {
		if !strings.Contains(text, want) {
			t.Fatalf("reject help should mention %q:\n%s", want, text)
		}
	}
}

func TestReviewsEscalateReturnsLegacyUnsupportedMessage(t *testing.T) {
	err := workflowsReviewsEscalateCmd.RunE(workflowsReviewsEscalateCmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected unsupported escalation error")
	}
	if !strings.Contains(err.Error(), "review escalation is not supported") {
		t.Fatalf("expected unsupported escalation guidance, got %q", err.Error())
	}
	for _, stale := range []string{"required flag", "escalate-to"} {
		if strings.Contains(err.Error(), stale) {
			t.Fatalf("escalate should not surface stale %q wording, got %q", stale, err.Error())
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
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(nil))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "get", RunE: workflowsReviewsGetCmd.RunE}
	stdout, _ := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews get: %v", err)
		}
	})
	if seenPath != "/workflows/reviews/run_1/blk_1" {
		t.Fatalf("path = %s", seenPath)
	}
	if !strings.Contains(stdout, `"versions_by_id"`) || !strings.Contains(stdout, `"`+reviewTestVersionID+`"`) {
		t.Fatalf("stdout = %s", stdout)
	}
	for _, want := range []string{`"decision": null`, `"parent_id": null`, `"note": null`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout should preserve %s:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, `"rev"`) || strings.Contains(stdout, `"head_seq"`) {
		t.Fatalf("stdout contains stale overlay fields: %s", stdout)
	}
}

func TestReviewsGetCommandHonorsOutputTable(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(nil))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

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
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") || strings.Contains(stdout, `"versions_by_id"`) {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"REVIEW_ID", "RUN_ID", "BLOCK_ID", "BLOCK_TYPE", "blockrun_1", "run_1", "blk_1", "extract"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}

func TestReviewsSchemaCommandPrintsBlockSpecificSnapshotContract(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenPath string
	var seenMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		body := reviewOverlayBody(nil)
		body["block_type"] = "classifier"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "schema", RunE: workflowsReviewsSchemaCmd.RunE}
	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews schema: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	if seenMethod != http.MethodGet {
		t.Fatalf("method = %s", seenMethod)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1" {
		t.Fatalf("path = %s", seenPath)
	}
	for _, want := range []string{
		`"block_type": "classifier"`,
		`"snapshot_schema"`,
		`"category"`,
		`"additionalProperties": false`,
		`"Do not include confidence fields or other metadata."`,
		`"create_usage"`,
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in schema output:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, `"reasoning"`) {
		t.Fatalf("classifier schema should not advertise reasoning:\n%s", stdout)
	}
}

func TestReviewsSchemaCommandCoversEveryReviewableBlockType(t *testing.T) {
	for _, tc := range []struct {
		blockType string
		want      []string
		stale     []string
	}{
		{
			blockType: "extract",
			want: []string{
				`"block_type": "extract"`,
				`"additionalProperties": true`,
				"Submit the extract output object itself.",
				"Do not wrap it in an output field",
			},
			stale: []string{`"output":`, "reviewable_value", "head_seq", "effective_seq"},
		},
		{
			blockType: "classifier",
			want: []string{
				`"block_type": "classifier"`,
				`"required": [`,
				`"category"`,
				"Submit only the selected category.",
			},
			stale: []string{`"reasoning"`, `"confidence"`, "reviewable_value", "head_seq", "effective_seq"},
		},
		{
			blockType: "split",
			want: []string{
				`"block_type": "split"`,
				`"documents"`,
				`"name"`,
				`"pages"`,
				"Submit the complete split list",
			},
			stale: []string{`"splits"`, `"chunks"`, `"output"`, "reviewable_value", "head_seq", "effective_seq"},
		},
		{
			blockType: "for_each",
			want: []string{
				`"block_type": "for_each"`,
				`"partitions"`,
				`"key"`,
				`"pages"`,
				"split-by-key",
				"Submit the complete partition list",
			},
			stale: []string{`"chunks"`, `"documents"`, `"output"`, "reviewable_value", "head_seq", "effective_seq"},
		},
	} {
		t.Run(tc.blockType, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			var requests int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodGet || r.URL.Path != "/workflows/reviews/run_1/blk_1" {
					t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				body := reviewOverlayBody(nil)
				body["block_type"] = tc.blockType
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(body)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "schema", RunE: workflowsReviewsSchemaCmd.RunE}
			stdout, stderr := captureStd(t, func() {
				if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
					t.Fatalf("reviews schema: %v", err)
				}
			})
			if stderr != "" {
				t.Fatalf("stderr = %s", stderr)
			}
			if requests != 1 {
				t.Fatalf("requests = %d", requests)
			}
			for _, want := range tc.want {
				if !strings.Contains(stdout, want) {
					t.Fatalf("expected %q in schema output:\n%s", want, stdout)
				}
			}
			for _, stale := range tc.stale {
				if strings.Contains(stdout, stale) {
					t.Fatalf("schema output should not contain stale %q:\n%s", stale, stdout)
				}
			}
		})
	}
}

func TestReviewsSchemaCommandHonorsOutputTable(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := reviewOverlayBody(nil)
		body["block_type"] = "split"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsReviewsSchemaCmd.RunE(workflowsReviewsSchemaCmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews schema: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	for _, want := range []string{"BLOCK_TYPE", "SCHEMA", "EXAMPLE", "NOTES", "CREATE_USAGE", "split", "documents", "retab workflows reviews versions create"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}

func TestReviewsSchemaCommandRejectsNonReviewableBlockType(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		body := reviewOverlayBody(nil)
		body["block_type"] = "parse"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, stderr := captureStd(t, func() {
		err := workflowsReviewsSchemaCmd.RunE(workflowsReviewsSchemaCmd, []string{"run_1", "blk_1"})
		if err == nil {
			t.Fatal("expected non-reviewable block type to fail")
		}
		if !strings.Contains(err.Error(), "parse") || !strings.Contains(err.Error(), "not reviewable") {
			t.Fatalf("error = %v", err)
		}
	})
	if !strings.Contains(stderr, "block type \"parse\" is not reviewable") {
		t.Fatalf("stderr = %s", stderr)
	}
	if requests != 1 {
		t.Fatalf("requests = %d", requests)
	}
}

func TestReviewsSchemaCommandPropagatesMissingOverlay(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"detail": "review overlay not found"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, stderr := captureStd(t, func() {
		err := workflowsReviewsSchemaCmd.RunE(workflowsReviewsSchemaCmd, []string{"run_missing", "blk_1"})
		if err == nil {
			t.Fatal("expected missing overlay to fail")
		}
	})
	if !strings.Contains(stderr, "404") || !strings.Contains(stderr, "review overlay not found") {
		t.Fatalf("stderr = %s", stderr)
	}
}

func TestReviewsWaitStopsWhenOverlayIsAwaitingReview(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var reviewGets int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reviewGets++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(nil))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newWaitTestCmd()
	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews wait: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	if reviewGets != 1 {
		t.Fatalf("reviewGets=%d", reviewGets)
	}
	if !strings.Contains(stdout, `"versions_by_id"`) || !strings.Contains(stdout, `"`+reviewTestVersionID+`"`) {
		t.Fatalf("stdout = %s", stdout)
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
	t.Setenv("RETAB_API_BASE_URL", server.URL)

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
	t.Setenv("RETAB_API_BASE_URL", server.URL)

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

func TestReviewsApproveSendsVersionID(t *testing.T) {
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
			"review":            reviewOverlayBody(reviewDecisionBody("approved", reviewTestVersionID)),
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newApproveTestCmd()
	if err := cmd.Flags().Set("version-id", reviewTestVersionID); err != nil {
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
	if body["verdict"] != "approved" || body["version_id"] != reviewTestVersionID {
		t.Fatalf("body = %#v", body)
	}
	for _, stale := range []string{"version_stamp", "reviewable_value", "snapshot"} {
		if _, ok := body[stale]; ok {
			t.Fatalf("body contains stale %q: %#v", stale, body)
		}
	}
	if stderr != "" {
		t.Fatalf("stderr = %s", stderr)
	}
	if !strings.Contains(stdout, "accepted") {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestReviewsApproveTableRendersConciseDecisionResponse(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	body := reviewOverlayBody(reviewDecisionBody("approved", reviewTestVersionID))
	body["versions_by_id"] = map[string]any{
		reviewTestVersionID: reviewVersionBody(nil, map[string]any{"huge_nested_payload": strings.Repeat("x", 4096)}),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"review":            body,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

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
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") || strings.Contains(stdout, `"versions_by_id"`) || strings.Contains(stdout, "huge_nested_payload") {
		t.Fatalf("expected concise table output, got:\n%s", stdout)
	}
	for _, want := range []string{"SUBMISSION", "REVIEW_ID", "RUN_ID", "BLOCK_ID", "BLOCK_TYPE", "VERDICT", "VERSION_ID", "accepted", "blockrun_1", "run_1", "blk_1", "extract", "approved", reviewTestVersionID} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}

func TestReviewsVersionsCreateSendsSnapshotAndParentID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	snapshotPath := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(snapshotPath, []byte(`{"output":{"total":110,"currency":"USD"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayBody(nil))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newVersionsCreateTestCmd()
	if err := cmd.Flags().Set("snapshot-file", snapshotPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("parent-id", reviewTestVersionID); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("note", "fixed total"); err != nil {
		t.Fatal(err)
	}
	captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews versions create: %v", err)
		}
	})
	if seenPath != "/workflows/reviews/run_1/blk_1/versions" {
		t.Fatalf("path = %s", seenPath)
	}
	snapshot, ok := body["snapshot"].(map[string]any)
	if !ok {
		t.Fatalf("snapshot missing from body: %#v", body)
	}
	output, ok := snapshot["output"].(map[string]any)
	if !ok || output["total"] != float64(110) || output["currency"] != "USD" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if body["parent_id"] != reviewTestVersionID || body["note"] != "fixed total" {
		t.Fatalf("body = %#v", body)
	}
	for _, stale := range []string{"origin", "version_stamp", "reviewable_value", "command_id"} {
		if _, ok := body[stale]; ok {
			t.Fatalf("body contains stale %q: %#v", stale, body)
		}
	}
}

func TestReviewsVersionsCreateReadsSnapshotBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	cmd := newVersionsCreateTestCmd()
	if err := cmd.Flags().Set("snapshot-file", filepath.Join(t.TempDir(), "missing.json")); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("parent-id", reviewTestVersionID); err != nil {
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

func TestReviewsRejectRequiresReasonBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	cmd := newRejectTestCmd()
	if err := cmd.Flags().Set("version-id", reviewTestVersionID); err != nil {
		t.Fatal(err)
	}
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

func TestReviewsDecisionCommandsValidateContentIDsBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "approve", cmd: newApproveTestCmd(), args: []string{"run_1", "blk_1"}},
		{name: "reject", cmd: newRejectTestCmd(), args: []string{"run_1", "blk_1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cmd.Flags().Set("version-id", "ver_abc"); err != nil {
				t.Fatal(err)
			}
			if tc.name == "reject" {
				if err := tc.cmd.Flags().Set("reason", "not an invoice"); err != nil {
					t.Fatal(err)
				}
			}
			err := tc.cmd.RunE(tc.cmd, tc.args)
			if err == nil {
				t.Fatal("expected malformed version id error")
			}
			if !strings.Contains(err.Error(), "--version-id") || !strings.Contains(err.Error(), "64-character hex") {
				t.Fatalf("error = %v", err)
			}
			if strings.Contains(err.Error(), "credentials") {
				t.Fatalf("error %q checked credentials before validating --version-id", err.Error())
			}
		})
	}
}

func TestReviewsVersionsCreateValidatesParentIDBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	snapshotPath := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(snapshotPath, []byte(`{"output":{"total":110}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newVersionsCreateTestCmd()
	if err := cmd.Flags().Set("snapshot-file", snapshotPath); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("parent-id", "ver_abc"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"run_1", "blk_1"})
	if err == nil {
		t.Fatal("expected malformed parent id error")
	}
	if !strings.Contains(err.Error(), "--parent-id") || !strings.Contains(err.Error(), "64-character hex") {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "credentials") {
		t.Fatalf("error %q checked credentials before validating --parent-id", err.Error())
	}
}

func TestReviewsRejectSendsVersionIDAndReason(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"review":            reviewOverlayBody(reviewDecisionBody("rejected", reviewTestVersionID)),
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newRejectTestCmd()
	if err := cmd.Flags().Set("version-id", reviewTestVersionID); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("reason", "not an invoice"); err != nil {
		t.Fatal(err)
	}
	captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"run_1", "blk_1"}); err != nil {
			t.Fatalf("reviews reject: %v", err)
		}
	})
	if body["verdict"] != "rejected" || body["version_id"] != reviewTestVersionID || body["reason"] != "not an invoice" {
		t.Fatalf("body = %#v", body)
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
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if !workflowsReviewsEscalateCmd.Hidden {
		t.Fatal("reviews escalate should be hidden from help")
	}
	if strings.Contains(workflowsReviewsCmd.Long, "escalate") {
		t.Fatalf("reviews help should not advertise escalation:\n%s", workflowsReviewsCmd.Long)
	}

	cmd := newEscalateTestCmd()
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
	for _, want := range []string{"not supported", "approve", "reject", "versions create"} {
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
	cmd.Flags().String("version-id", "", "")
	return cmd
}

func newRejectTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "reject", RunE: workflowsReviewsRejectCmd.RunE}
	cmd.Flags().String("version-id", "", "")
	cmd.Flags().String("reason", "", "")
	return cmd
}

func newVersionsCreateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "create", RunE: workflowsReviewsVersionsCreateCmd.RunE}
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().String("snapshot-file", "", "")
	cmd.Flags().String("note", "", "")
	return cmd
}

func TestReviewEnumFlagsShowAllowedValues(t *testing.T) {
	decisionFlag := newEnumStringFlagValue("--decision", "none", "any")
	err := decisionFlag.Set("typo")
	if err == nil {
		t.Fatal("expected invalid decision to fail")
	}
	if !strings.Contains(err.Error(), "none | any") {
		t.Fatalf("decision error should show allowed values, got: %v", err)
	}
}

func newEscalateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "escalate", RunE: workflowsReviewsEscalateCmd.RunE}
	cmd.Flags().String("reason", "", "")
	cmd.Flags().String("escalate-to", "", "")
	return cmd
}

func newWaitTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "wait", RunE: workflowsReviewsWaitCmd.RunE}
	cmd.Flags().Var(&positiveIntFlagValue{value: "120"}, "timeout", "")
	cmd.Flags().Var(&positiveIntFlagValue{value: "2"}, "poll-interval", "")
	return cmd
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
