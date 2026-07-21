//go:build !retab_oagen_cli_workflows

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// isEffectivelyEmptyDraft is the pure shape predicate that decides whether
// a draft warrants the empty-workflow warning. Pin every interesting
// shape — fully empty, single start_document, single non-start, multi-block —
// so a future refactor can't silently change behaviour.
func TestIsEffectivelyEmptyDraft(t *testing.T) {
	cases := []struct {
		name   string
		blocks []retab.WorkflowBlock
		want   bool
	}{
		{name: "no blocks", blocks: nil, want: true},
		{name: "empty slice", blocks: []retab.WorkflowBlock{}, want: true},
		{name: "single start_document block — freshly-created shape", blocks: []retab.WorkflowBlock{{Type: "start_document"}}, want: true},
		{name: "single non-start_document block", blocks: []retab.WorkflowBlock{{Type: "extract"}}, want: false},
		{name: "two blocks", blocks: []retab.WorkflowBlock{{Type: "start_document"}, {Type: "extract"}}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isEffectivelyEmptyDraft(tc.blocks); got != tc.want {
				t.Errorf("isEffectivelyEmptyDraft(%+v) = %v, want %v", tc.blocks, got, tc.want)
			}
		})
	}
}

func TestWorkflowsListRejectsInvalidListFlagsLocally(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
		reset     string
	}{
		{name: "negative limit", flag: "limit", value: "-1", wantError: "between 1 and 100", reset: "1"},
		{name: "invalid order", flag: "order", value: "sideways", wantError: "asc", reset: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := workflowsListCmd.Flags().Set(tc.flag, tc.value)
			if err == nil {
				t.Fatalf("expected local parse error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if resetErr := workflowsListCmd.Flags().Set(tc.flag, tc.reset); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestWorkflowsListRejectsOverLimitLocally(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "workflows list", cmd: workflowsListCmd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("limit", "101")
			if err == nil {
				t.Fatal("expected local parse error for --limit=101")
			}
			if !strings.Contains(err.Error(), "between 1 and 100") {
				t.Fatalf("error %q does not mention backend limit range", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("limit", "1"); resetErr != nil {
				t.Fatalf("reset --limit: %v", resetErr)
			}
		})
	}
}

func TestWorkflowExamplesUseCurrentDocumentFlag(t *testing.T) {
	for _, tc := range []struct {
		name    string
		example string
	}{
		{name: "workflows", example: workflowsCmd.Example},
		{name: "publish", example: workflowsPublishCmd.Example},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Contains(tc.example, "--document-file") {
				t.Fatalf("%s example should use --document, got:\n%s", tc.name, tc.example)
			}
			if !strings.Contains(tc.example, "--document start=") {
				t.Fatalf("%s example should include --document start=..., got:\n%s", tc.name, tc.example)
			}
		})
	}
}

func TestWorkflowHelpGuidesReviewConfig(t *testing.T) {
	surfaces := []struct {
		name string
		text string
	}{
		{name: "workflows long", text: workflowsCmd.Long},
		{name: "blocks long", text: workflowsBlocksCmd.Long},
		{name: "blocks create long", text: workflowsBlocksCreateCmd.Long},
		{name: "blocks update long", text: workflowsBlocksUpdateCmd.Long},
		{name: "runs long", text: workflowsRunsCmd.Long},
		{name: "reviews long", text: workflowsReviewsCmd.Long},
	}
	for _, surface := range surfaces {
		t.Run(surface.name, func(t *testing.T) {
			staleBacktick := "`hil` " + "block"
			staleTitle := "HIL " + "block"
			if strings.Contains(surface.text, staleBacktick) || strings.Contains(surface.text, staleTitle) {
				t.Fatalf("%s should not describe HIL as a standalone block:\n%s", surface.name, surface.text)
			}
			for _, stale := range []string{"human-review gate", "review gate", "review gates"} {
				if strings.Contains(surface.text, stale) {
					t.Fatalf("%s should use review-centric wording without %q:\n%s", surface.name, stale, surface.text)
				}
			}
			if !strings.Contains(surface.text, "config.review") && !strings.Contains(surface.text, "awaiting_review") && !strings.Contains(surface.text, "review") {
				t.Fatalf("%s should guide users toward review config/reviews, got:\n%s", surface.name, surface.text)
			}
		})
	}
}

func TestWorkflowHelpMentionsConsensusReviewCriteria(t *testing.T) {
	for _, want := range []string{
		"n_consensus > 1",
		"confidence_lt",
		"top_margin_lt",
		"boundary_confidence_lt",
		"json_condition",
		"likelihoods.*",
	} {
		if !strings.Contains(workflowsCmd.Long, want) {
			t.Fatalf("workflows help should mention consensus review criteria %q, got:\n%s", want, workflowsCmd.Long)
		}
	}
}

func TestWorkflowRenderedHelpMentionsConsensusReviewCriteria(t *testing.T) {
	var buf bytes.Buffer
	workflowsCmd.SetOut(&buf)
	workflowsCmd.SetErr(&buf)
	t.Cleanup(func() {
		workflowsCmd.SetOut(nil)
		workflowsCmd.SetErr(nil)
	})
	if err := workflowsCmd.Help(); err != nil {
		t.Fatalf("render workflows help: %v", err)
	}
	help := buf.String()

	for _, want := range []string{
		"Review is configured on the block",
		"awaiting_review",
		"n_consensus > 1",
		"confidence_lt",
		"top_margin_lt",
		"boundary_confidence_lt",
		"json_condition",
		"likelihoods.*",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("rendered workflows help should mention consensus review criteria %q, got:\n%s", want, help)
		}
	}
	for _, stale := range []string{"review gate", "`hil`", "human-review gate"} {
		if strings.Contains(help, stale) {
			t.Fatalf("rendered workflows help should not contain stale %q, got:\n%s", stale, help)
		}
	}
}

func TestParseBlockCreateRejectsStandaloneHILBlockType(t *testing.T) {
	_, err := parseBlockCreate(map[string]any{
		"id":   "review",
		"type": "hil",
	})
	if err == nil {
		t.Fatal("expected type=hil to be rejected locally")
	}
	for _, want := range []string{"config.review", "reviewable block"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q should contain %q", err.Error(), want)
		}
	}
}

func TestMergeWorkflowBlockConfigPreservesExistingConfig(t *testing.T) {
	existing := map[string]any{
		"model": "retab-small",
		"json_schema": map[string]any{
			"type":     "object",
			"required": []any{"total"},
		},
		"nested": map[string]any{
			"keep": "yes",
			"old":  "value",
		},
	}
	patch := map[string]any{
		"review": map[string]any{
			"predicate": map[string]any{"kind": "always"},
		},
		"nested": map[string]any{
			"old": "new",
		},
	}

	merged := mergeWorkflowBlockConfig(existing, patch)
	if merged["model"] != "retab-small" {
		t.Fatalf("model was not preserved: %#v", merged)
	}
	if _, ok := merged["json_schema"].(map[string]any); !ok {
		t.Fatalf("json_schema was not preserved: %#v", merged)
	}
	nested := merged["nested"].(map[string]any)
	if nested["keep"] != "yes" || nested["old"] != "new" {
		t.Fatalf("nested merge = %#v", nested)
	}
	if _, ok := merged["review"].(map[string]any); !ok {
		t.Fatalf("review patch was not applied: %#v", merged)
	}
}

func TestMergeWorkflowBlockConfigNullDeletesKey(t *testing.T) {
	existing := map[string]any{
		"model":  "retab-small",
		"review": map[string]any{"predicate": map[string]any{"kind": "always"}},
		"nested": map[string]any{"keep": "yes", "drop": 1},
	}
	patch := map[string]any{
		"review": nil,                         // RFC 7396: delete the key
		"nested": map[string]any{"drop": nil}, // delete a nested key
	}

	merged := mergeWorkflowBlockConfig(existing, patch)
	if _, ok := merged["review"]; ok {
		t.Fatalf("review should have been deleted: %#v", merged)
	}
	if merged["model"] != "retab-small" {
		t.Fatalf("model should be preserved: %#v", merged)
	}
	nested, ok := merged["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested should be preserved: %#v", merged)
	}
	if _, ok := nested["drop"]; ok {
		t.Fatalf("nested.drop should have been deleted: %#v", nested)
	}
	if nested["keep"] != "yes" {
		t.Fatalf("nested.keep should be preserved: %#v", nested)
	}
}

func TestWorkflowsBlocksGetUsesBlockEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawGet bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/blk_extract" || r.URL.RawQuery != "" {
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		sawGet = true
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "blk_extract",
			"type":   "extract",
			"label":  "Extract",
			"config": map[string]any{"model": "retab-small"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "get", RunE: workflowsBlocksGetCmd.RunE}
	if err := cmd.RunE(cmd, []string{"blk_extract"}); err != nil {
		t.Fatalf("blocks get: %v", err)
	}
	if !sawGet {
		t.Fatal("expected blocks get to fetch the block")
	}
}

func TestWorkflowsBlocksUpdateMergeConfigSendsPatchMode(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	patchPath := t.TempDir() + "/review.json"
	if err := os.WriteFile(
		patchPath,
		[]byte(`{"review":{"predicate":{"kind":"always"}}}`),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	var sawPatch bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/v1/workflows/blocks/blk_extract" && r.URL.RawQuery == "":
			sawPatch = true
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode patch: %v", err)
			}
			config := body["config"].(map[string]any)
			if config["review"] == nil {
				t.Fatalf("review patch missing from body: %#v", config)
			}
			if body["config_mode"] != "merge" {
				t.Fatalf("config_mode = %#v, want merge", body["config_mode"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "blk_extract",
				"type":   "extract",
				"config": config,
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "update", RunE: workflowsBlocksUpdateCmd.RunE}
	cmd.Flags().String("label", "", "")
	cmd.Flags().Float64("position-x", 0, "")
	cmd.Flags().Float64("position-y", 0, "")
	cmd.Flags().Var(&nonNegativeFloatFlagValue{}, "width", "")
	cmd.Flags().Var(&nonNegativeFloatFlagValue{}, "height", "")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().String("config-file", "", "")
	cmd.Flags().String("merge-config-file", "", "")
	if err := cmd.Flags().Set("merge-config-file", patchPath); err != nil {
		t.Fatal(err)
	}

	if err := cmd.RunE(cmd, []string{"blk_extract"}); err != nil {
		t.Fatalf("blocks update: %v", err)
	}
	if !sawPatch {
		t.Fatal("expected PATCH")
	}
}

func TestWorkflowsBlocksUpdateRejectsReplaceAndMergeConfigTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "update", RunE: workflowsBlocksUpdateCmd.RunE}
	cmd.Flags().String("label", "", "")
	cmd.Flags().Float64("position-x", 0, "")
	cmd.Flags().Float64("position-y", 0, "")
	cmd.Flags().Var(&nonNegativeFloatFlagValue{}, "width", "")
	cmd.Flags().Var(&nonNegativeFloatFlagValue{}, "height", "")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().String("config-file", "", "")
	cmd.Flags().String("merge-config-file", "", "")
	if err := cmd.Flags().Set("config-file", "full.json"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("merge-config-file", "patch.json"); err != nil {
		t.Fatal(err)
	}

	err := cmd.RunE(cmd, []string{"blk_extract"})
	if err == nil {
		t.Fatal("expected mutually exclusive config flags to fail")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %v", err)
	}
}

func TestWorkflowsRejectBlankNamesBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "create", cmd: workflowsCreateCmd},
		{name: "update", cmd: workflowsUpdateCmd, args: []string{"wf_123"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for blank workflow name, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "test-workflow-" + tc.name, RunE: tc.cmd.RunE}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")

			if err := cmd.Flags().Set("name", "   "); err != nil {
				t.Fatal(err)
			}

			err := cmd.RunE(cmd, tc.args)
			if err == nil {
				t.Fatal("expected blank workflow name error")
			}
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				err = unwrapped
			}
			if !strings.Contains(err.Error(), "workflow name must not be blank") {
				t.Fatalf("error %q does not mention blank workflow name", err.Error())
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
	}
}

func TestWorkflowListExamplesUsePaginatedEnvelope(t *testing.T) {
	if strings.Contains(workflowsBlocksListCmd.Example, ".[].") {
		t.Fatalf("blocks list example should iterate over .data[], got:\n%s", workflowsBlocksListCmd.Example)
	}
	if !strings.Contains(workflowsBlocksListCmd.Example, ".data[].id") {
		t.Fatalf("blocks list example should read .data[].id, got:\n%s", workflowsBlocksListCmd.Example)
	}
}

func TestWorkflowsGetExampleUsesPublishedObjectShape(t *testing.T) {
	if strings.Contains(workflowsGetCmd.Example, ".published_version") {
		t.Fatalf("workflows get example references removed field:\n%s", workflowsGetCmd.Example)
	}
	if !strings.Contains(workflowsGetCmd.Example, ".published.version_id") {
		t.Fatalf("workflows get example should show current published version path:\n%s", workflowsGetCmd.Example)
	}
}

func TestWorkflowsPublishRejectsMalformedSuccessResponse(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/wf_123/publish" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workflow": map[string]any{
				"id":   "wf_123",
				"name": "Wrapped Workflow",
			},
			"baseline": map[string]any{"workflow_id": "wf_123"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "publish", RunE: workflowsPublishCmd.RunE}
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("force", false, "")
	if err := cmd.Flags().Set("force", "true"); err != nil {
		t.Fatal(err)
	}

	var err error
	stdout, stderr := captureStd(t, func() {
		err = cmd.RunE(cmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected malformed publish response to fail")
	}
	if stdout != "" {
		t.Fatalf("malformed publish response should not print stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "publish response did not include workflow id") {
		t.Fatalf("stderr = %q, want missing workflow id message", stderr)
	}
}

// TestWarnIfEmptyWorkflowOnPublish_StartOnly mocks Blocks.List against a
// fake HTTP server returning a single `start_document` block, then asserts the
// warning text — and only that text — lands on the provided writer.
// This is the canonical "user fat-fingered `workflows publish` on a
// fresh draft" scenario.
func TestWarnIfEmptyWorkflowOnPublish_StartOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workflows/blocks" || r.URL.Query().Get("workflow_id") != "wf_abc" {
			t.Errorf("unexpected path: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "start_1", "workflow_id": "wf_abc", "type": "start_document", "label": "Start"},
			},
			"list_metadata": map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := retab.NewClient("fake-key", retab.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var buf bytes.Buffer
	warnIfEmptyWorkflowOnPublish(context.Background(), client, "wf_abc", &buf)

	out := buf.String()
	if !strings.Contains(out, "warning: workflow has only a start_document block") {
		t.Errorf("missing primary warning. got:\n%s", out)
	}
	if !strings.Contains(out, "add blocks with `retab workflows blocks create`") {
		t.Errorf("missing follow-up warning. got:\n%s", out)
	}
}

// TestWarnIfEmptyWorkflowOnPublish_NonEmpty pins the silent path: when
// the workflow has real blocks (start + extract), no warning is emitted.
// Catches future regressions where the predicate widens accidentally.
func TestWarnIfEmptyWorkflowOnPublish_NonEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "start_1", "type": "start_document"},
				{"id": "blk_extract_1", "type": "extract"},
			},
			"list_metadata": map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := retab.NewClient("fake-key", retab.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var buf bytes.Buffer
	warnIfEmptyWorkflowOnPublish(context.Background(), client, "wf_abc", &buf)

	if buf.Len() != 0 {
		t.Errorf("expected silent on non-empty workflow, got:\n%s", buf.String())
	}
}

// TestWarnIfEmptyWorkflowOnPublish_NetworkError pins the best-effort
// contract: a server error MUST NOT block publishing. The helper swallows
// the error and emits nothing — the publish call itself will surface any
// real auth/server failure separately.
func TestWarnIfEmptyWorkflowOnPublish_NetworkError(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := retab.NewClient("fake-key", retab.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var buf bytes.Buffer
	warnIfEmptyWorkflowOnPublish(context.Background(), client, "wf_abc", &buf)

	if calls.Load() == 0 {
		t.Fatal("expected helper to hit the server at least once")
	}
	if buf.Len() != 0 {
		t.Errorf("expected silent on server error (best-effort), got:\n%s", buf.String())
	}
}

func TestWorkflowsDeleteWithYesFlagProceedsWithoutPrompt(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawDelete atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/v1/workflows/wf_to_delete" {
			sawDelete.Add(1)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsDeleteCmd.Flags().Set("yes", "false")
	})

	if err := workflowsDeleteCmd.RunE(workflowsDeleteCmd, []string{"wf_to_delete"}); err != nil {
		t.Fatalf("delete with --yes: %v", err)
	}
	if sawDelete.Load() != 1 {
		t.Fatalf("expected one DELETE call, got %d", sawDelete.Load())
	}
}

func TestWorkflowsDeleteWithoutYesAndNonTTYStdinRefuses(t *testing.T) {
	// Without --yes and stdin not a TTY (any test environment, CI, pipe),
	// the command must refuse before hitting the server. A stray newline
	// or empty pipe could otherwise auto-confirm and nuke a workflow.
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// Ensure --yes is not set, in case a previous test left it on.
	if err := workflowsDeleteCmd.Flags().Set("yes", "false"); err != nil {
		t.Fatal(err)
	}

	err := workflowsDeleteCmd.RunE(workflowsDeleteCmd, []string{"wf_keep"})
	if err == nil {
		t.Fatal("expected refusal when stdin is not a TTY")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("error %q does not mention --yes", err.Error())
	}
	if hits.Load() != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits.Load())
	}
}

// TestWorkflowsDiscardDraftWithYesFlagProceedsWithoutPrompt pins that the
// command POSTs to the discard-draft action route when --yes skips the
// confirmation prompt.
func TestWorkflowsDiscardDraftWithYesFlagProceedsWithoutPrompt(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawDiscard atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/wf_revert/discard-draft" {
			sawDiscard.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"wf_revert"}`))
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsDiscardDraftCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsDiscardDraftCmd.Flags().Set("yes", "false")
	})

	if err := workflowsDiscardDraftCmd.RunE(workflowsDiscardDraftCmd, []string{"wf_revert"}); err != nil {
		t.Fatalf("discard-draft with --yes: %v", err)
	}
	if sawDiscard.Load() != 1 {
		t.Fatalf("expected one POST to discard-draft, got %d", sawDiscard.Load())
	}
}

// TestWorkflowsDiscardDraftWithoutYesAndNonTTYStdinRefuses pins the
// destructive-confirmation guard: discarding draft edits without --yes and
// without a TTY must refuse before hitting the server.
func TestWorkflowsDiscardDraftWithoutYesAndNonTTYStdinRefuses(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsDiscardDraftCmd.Flags().Set("yes", "false"); err != nil {
		t.Fatal(err)
	}

	err := workflowsDiscardDraftCmd.RunE(workflowsDiscardDraftCmd, []string{"wf_keep"})
	if err == nil {
		t.Fatal("expected refusal when stdin is not a TTY")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("error %q does not mention --yes", err.Error())
	}
	if hits.Load() != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits.Load())
	}
}
