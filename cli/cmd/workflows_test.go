package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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

func TestWorkflowsDiagnoseReadsGraphFileBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := t.TempDir() + "/missing-graph.json"
	if err := workflowsDiagnoseCmd.Flags().Set("graph-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsDiagnoseCmd.Flags().Set("graph-file", "") })

	err := workflowsDiagnoseCmd.RunE(workflowsDiagnoseCmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected missing graph file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "missing-graph.json") {
		t.Fatalf("error should mention missing graph file, got %q", err.Error())
	}
}

// TestWorkflowsListRejectsUnknownFieldsLocally pins client-side
// allowlist validation for `workflows list --fields`. The server silently
// ignores unknown selectors, so a typo would otherwise return the full
// payload without complaint; reject it locally before any HTTP call.
func TestWorkflowsListRejectsUnknownFieldsLocally(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_BASE_URL", "http://127.0.0.1:1")

	if err := workflowsListCmd.Flags().Set("fields", "bogus"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsListCmd.Flags().Set("fields", "") })

	workflowsListCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsListCmd.SetContext(nil) })

	err := workflowsListCmd.RunE(workflowsListCmd, nil)
	if err == nil {
		t.Fatal("expected local validation error for --fields bogus, got nil")
	}
	if !strings.Contains(err.Error(), "not a valid field") {
		t.Fatalf("error should say not a valid field, got: %v", err)
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("error should quote the offending value, got: %v", err)
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

// Regression for CLI probing 2026-05: a clearly invalid value like
// `not@a-domain` passed to `--allowed-domain` was accepted because the
// validator only refused blank strings. The CLI now rejects shapes that
// cannot match any real inbound email — saving a misconfigured allowlist
// is silently broken otherwise (no email ever matches, no error ever
// surfaces).
func TestWorkflowsUpdateRejectsMalformedAllowlistValuesBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
	}{
		{name: "domain with @", flag: "allowed-domain", value: "not@a-domain", wantError: "--allowed-domain"},
		{name: "domain with no dot", flag: "allowed-domain", value: "localhost", wantError: "--allowed-domain"},
		{name: "domain with trailing dot", flag: "allowed-domain", value: "example.", wantError: "--allowed-domain"},
		{name: "domain with leading dot", flag: "allowed-domain", value: ".example.com", wantError: "--allowed-domain"},
		{name: "domain with consecutive dots", flag: "allowed-domain", value: "example..com", wantError: "--allowed-domain"},
		{name: "domain with bad chars", flag: "allowed-domain", value: "example!.com", wantError: "--allowed-domain"},
		{name: "sender without @", flag: "allowed-sender", value: "no-at-sign", wantError: "--allowed-sender"},
		{name: "sender empty local part", flag: "allowed-sender", value: "@example.com", wantError: "--allowed-sender"},
		{name: "sender empty domain part", flag: "allowed-sender", value: "ops@", wantError: "--allowed-sender"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for malformed allowlist value, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "test-workflow-update", RunE: workflowsUpdateCmd.RunE}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().StringArray("allowed-sender", nil, "")
			cmd.Flags().StringArray("allowed-domain", nil, "")

			if err := cmd.Flags().Set(tc.flag, tc.value); err != nil {
				t.Fatal(err)
			}
			err := cmd.RunE(cmd, []string{"wf_123"})
			if err == nil {
				t.Fatalf("expected error for --%s=%q", tc.flag, tc.value)
			}
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				err = unwrapped
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not name %s", err.Error(), tc.wantError)
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
	}
}

func TestWorkflowsUpdateRejectsBlankEmailAllowlistValuesBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "blank sender", flag: "allowed-sender", wantError: "--allowed-sender must not be blank"},
		{name: "blank domain", flag: "allowed-domain", wantError: "--allowed-domain must not be blank"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for invalid allowlist value, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "test-workflow-update", RunE: workflowsUpdateCmd.RunE}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().StringArray("allowed-sender", nil, "")
			cmd.Flags().StringArray("allowed-domain", nil, "")

			if err := cmd.Flags().Set(tc.flag, "   "); err != nil {
				t.Fatal(err)
			}

			err := cmd.RunE(cmd, []string{"wf_123"})
			if err == nil {
				t.Fatal("expected blank allowlist value error")
			}
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				err = unwrapped
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
	}
}

// Regression: passing `--allowed-sender ""` (empty string) used to silently
// wipe the persisted allowlist. pflag's StringArray drops empty values from
// the resulting slice, so “validateWorkflowEmailAllowlistValues“ looped
// zero times and the (cleared) slice flowed through to a full-replace
// “EmailTrigger“ patch on the server. The CLI now refuses to send the
// PATCH when the flag was set but resolved to an empty slice.
func TestWorkflowsUpdateRejectsEmptyAllowlistValueBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "empty sender", flag: "allowed-sender", wantError: "--allowed-sender must not be blank"},
		{name: "empty domain", flag: "allowed-domain", wantError: "--allowed-domain must not be blank"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for empty allowlist value, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "test-workflow-update", RunE: workflowsUpdateCmd.RunE}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().StringArray("allowed-sender", nil, "")
			cmd.Flags().StringArray("allowed-domain", nil, "")

			// Empty string is the bug case: pflag's StringArray will set the
			// flag as Changed but drop the value from the returned slice.
			if err := cmd.Flags().Set(tc.flag, ""); err != nil {
				t.Fatal(err)
			}

			err := cmd.RunE(cmd, []string{"wf_123"})
			if err == nil {
				t.Fatalf("expected empty allowlist value error for --%s", tc.flag)
			}
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				err = unwrapped
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
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
			if tc.name == "update" {
				cmd.Flags().StringArray("allowed-sender", nil, "")
				cmd.Flags().StringArray("allowed-domain", nil, "")
			}

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

// Regression for CLI probing 2026-05: “workflows update --allowed-domain x“
// used to wipe the existing “allowed_senders“ list (and vice versa)
// because the CLI sent only the explicitly-provided list to a
// full-replace PATCH endpoint. The user's mental model is patch
// semantics — omitting a flag should leave that list alone.
func TestWorkflowsUpdatePreservesUnspecifiedEmailAllowlist(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name            string
		setFlag         string
		setValue        string
		wantSenders     []string
		wantDomains     []string
		existingTrigger map[string]any
	}{
		{
			name:     "providing only --allowed-domain preserves senders",
			setFlag:  "allowed-domain",
			setValue: "new.com",
			existingTrigger: map[string]any{
				"allowed_senders": []string{"keep@example.com"},
				"allowed_domains": []string{"old.com"},
			},
			wantSenders: []string{"keep@example.com"},
			wantDomains: []string{"new.com"},
		},
		{
			name:     "providing only --allowed-sender preserves domains",
			setFlag:  "allowed-sender",
			setValue: "new@x.com",
			existingTrigger: map[string]any{
				"allowed_senders": []string{"old@x.com"},
				"allowed_domains": []string{"keep.com"},
			},
			wantSenders: []string{"new@x.com"},
			wantDomains: []string{"keep.com"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var patchBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/wf_abc":
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id":            "wf_abc",
						"name":          "Existing",
						"email_trigger": tc.existingTrigger,
					})
				case r.Method == http.MethodPatch && r.URL.Path == "/workflows/wf_abc":
					if err := json.NewDecoder(r.Body).Decode(&patchBody); err != nil {
						t.Fatalf("decode patch: %v", err)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id":            "wf_abc",
						"name":          "Existing",
						"email_trigger": patchBody["email_trigger"],
					})
				default:
					t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
				}
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "update", RunE: workflowsUpdateCmd.RunE}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().StringArray("allowed-sender", nil, "")
			cmd.Flags().StringArray("allowed-domain", nil, "")
			if err := cmd.Flags().Set(tc.setFlag, tc.setValue); err != nil {
				t.Fatal(err)
			}

			if err := cmd.RunE(cmd, []string{"wf_abc"}); err != nil {
				t.Fatalf("update: %v", err)
			}

			trigger, ok := patchBody["email_trigger"].(map[string]any)
			if !ok {
				t.Fatalf("PATCH body had no email_trigger: %#v", patchBody)
			}
			gotSenders := castStrings(trigger["allowed_senders"])
			gotDomains := castStrings(trigger["allowed_domains"])
			if !equalStrings(gotSenders, tc.wantSenders) {
				t.Fatalf("allowed_senders = %v, want %v", gotSenders, tc.wantSenders)
			}
			if !equalStrings(gotDomains, tc.wantDomains) {
				t.Fatalf("allowed_domains = %v, want %v", gotDomains, tc.wantDomains)
			}
		})
	}
}

func castStrings(value any) []string {
	if value == nil {
		return nil
	}
	raw, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		s, ok := v.(string)
		if !ok {
			return nil
		}
		out = append(out, s)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestWorkflowListExamplesUsePaginatedEnvelope(t *testing.T) {
	if strings.Contains(workflowsBlocksListCmd.Example, ".[].") {
		t.Fatalf("blocks list example should iterate over .data[], got:\n%s", workflowsBlocksListCmd.Example)
	}
	if !strings.Contains(workflowsBlocksListCmd.Example, ".data[].id") {
		t.Fatalf("blocks list example should read .data[].id, got:\n%s", workflowsBlocksListCmd.Example)
	}
}

func TestWorkflowsDiagnoseGraphFileRejectsMalformedGraphLocally(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cases := []struct {
		name      string
		body      string
		wantError string
	}{
		{name: "blocks not array", body: `{"blocks":{},"edges":[]}`, wantError: "blocks must be an array"},
		{name: "block item not object", body: `{"blocks":[1],"edges":[]}`, wantError: "blocks[0] must be an object"},
		{name: "edges not array", body: `{"blocks":[],"edges":{}}`, wantError: "edges must be an array"},
		{name: "edge item not object", body: `{"blocks":[],"edges":[1]}`, wantError: "edges[0] must be an object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			path := t.TempDir() + "/graph.json"
			if err := os.WriteFile(path, []byte(tc.body), 0o600); err != nil {
				t.Fatal(err)
			}
			workflowsDiagnoseCmd.SetContext(context.Background())
			t.Cleanup(func() { workflowsDiagnoseCmd.SetContext(nil) })
			if err := workflowsDiagnoseCmd.Flags().Set("graph-file", path); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = workflowsDiagnoseCmd.Flags().Set("graph-file", "") })

			var err error
			_, stderr := captureStd(t, func() {
				err = workflowsDiagnoseCmd.RunE(workflowsDiagnoseCmd, []string{"wf_graph"})
			})
			if err == nil {
				t.Fatalf("expected malformed graph error")
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want no new requests", got-before)
			}
		})
	}
}

func TestWorkflowsDiagnoseGraphFileAcceptsEntitiesShape(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/workflows/wf_graph/diagnose-graph" {
			t.Fatalf("path = %s, want diagnose-graph", r.URL.Path)
		}
		var posted map[string]any
		if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		blocks, ok := posted["blocks"].([]any)
		if !ok || len(blocks) != 1 {
			http.Error(w, "missing blocks array", http.StatusBadRequest)
			return
		}
		block, ok := blocks[0].(map[string]any)
		if !ok {
			http.Error(w, "block is not object", http.StatusBadRequest)
			return
		}
		position, ok := block["position"].(map[string]any)
		if !ok || position["x"] != float64(11) || position["y"] != float64(22) {
			http.Error(w, "block position was not normalized", http.StatusBadRequest)
			return
		}
		edges, ok := posted["edges"].([]any)
		if !ok || len(edges) != 1 {
			http.Error(w, "missing edges array", http.StatusBadRequest)
			return
		}
		edge, ok := edges[0].(map[string]any)
		if !ok || edge["source"] != "start" || edge["target"] != "extract" {
			http.Error(w, "edge endpoints were not normalized", http.StatusBadRequest)
			return
		}
		if posted["re_propagate"] != false {
			http.Error(w, "re_propagate was not preserved", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_valid":    true,
			"issues":      []any{},
			"suggestions": []any{},
			"stats": map[string]any{
				"total_blocks":          1,
				"total_edges":           1,
				"block_types":           map[string]any{"start": 1},
				"start_document_blocks": 1,
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	graphPath := t.TempDir() + "/entities.json"
	graph := `{
		"workflow": {"id": "wf_graph"},
		"blocks": [
			{
				"id": "start",
				"type": "start_document",
				"label": "Start",
				"config": {},
				"position_x": 11,
				"position_y": 22,
				"width": 300,
				"height": 200
			}
		],
		"edges": [
			{
				"id": "edge_1",
				"source_block": "start",
				"target_block": "extract",
				"source_handle": "output",
				"target_handle": "input"
			}
		],
		"re_propagate": false
	}`
	if err := os.WriteFile(graphPath, []byte(graph), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowsDiagnoseCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsDiagnoseCmd.SetContext(nil) })
	if err := workflowsDiagnoseCmd.Flags().Set("graph-file", graphPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsDiagnoseCmd.Flags().Set("graph-file", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsDiagnoseCmd.RunE(workflowsDiagnoseCmd, []string{"wf_graph"}); err != nil {
			t.Fatalf("diagnose graph-file: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, `"is_valid": true`) {
		t.Fatalf("expected diagnosis response, got:\n%s", stdout)
	}
}

// Bug 4: `workflows diagnose --graph-file <file>` with a JSON file that
// omits `blocks` and/or `edges` (the canonical case: `{}`) used to send
// `{"blocks": null, "edges": null}` to the server. The server treats nil
// arrays as "no override provided" and falls back to the persisted draft
// graph — so passing `{}` silently diagnosed the persisted draft instead
// of the user-provided empty graph (giving total_blocks: 1 for the
// default start_document instead of total_blocks: 0).
//
// The fix coerces nil to []map[string]any{} before sending, so the
// presence of `--graph-file` always means "diagnose THIS graph". This
// test pins both the wire shape (non-null arrays) and the user-facing
// semantics (the persisted draft is NOT consulted).
func TestWorkflowsDiagnoseGraphFileEmptyObjectSendsEmptyArrays(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var bodyBytes []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/workflows/wf_empty/diagnose-graph" {
			t.Fatalf("path = %s, want diagnose-graph", r.URL.Path)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		bodyBytes = b
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_valid":    true,
			"issues":      []any{},
			"suggestions": []any{},
			"stats": map[string]any{
				"total_blocks":          0,
				"total_edges":           0,
				"block_types":           map[string]any{},
				"start_document_blocks": 0,
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	graphPath := t.TempDir() + "/empty.json"
	if err := os.WriteFile(graphPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowsDiagnoseCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsDiagnoseCmd.SetContext(nil) })
	if err := workflowsDiagnoseCmd.Flags().Set("graph-file", graphPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsDiagnoseCmd.Flags().Set("graph-file", "") })

	if _, stderr := captureStd(t, func() {
		if err := workflowsDiagnoseCmd.RunE(workflowsDiagnoseCmd, []string{"wf_empty"}); err != nil {
			t.Fatalf("diagnose empty graph-file: %v", err)
		}
	}); stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	// Pin the wire shape: the request body must be a JSON object with
	// `blocks` and `edges` set to non-null empty arrays (not absent and
	// not null). The server keys off nil arrays to mean "fall back to
	// persisted draft", which is the bug we just fixed.
	var posted map[string]any
	if err := json.Unmarshal(bodyBytes, &posted); err != nil {
		t.Fatalf("unmarshal request body: %v\nbody: %s", err, string(bodyBytes))
	}
	blocksRaw, hasBlocks := posted["blocks"]
	if !hasBlocks {
		t.Fatalf("blocks key missing from request body: %s", string(bodyBytes))
	}
	if blocksRaw == nil {
		t.Fatalf("blocks is null in request body: %s", string(bodyBytes))
	}
	blocks, ok := blocksRaw.([]any)
	if !ok {
		t.Fatalf("blocks is not an array: %T (%v)", blocksRaw, blocksRaw)
	}
	if len(blocks) != 0 {
		t.Fatalf("blocks should be empty, got %d: %v", len(blocks), blocks)
	}
	edgesRaw, hasEdges := posted["edges"]
	if !hasEdges {
		t.Fatalf("edges key missing from request body: %s", string(bodyBytes))
	}
	if edgesRaw == nil {
		t.Fatalf("edges is null in request body: %s", string(bodyBytes))
	}
	edges, ok := edgesRaw.([]any)
	if !ok {
		t.Fatalf("edges is not an array: %T (%v)", edgesRaw, edgesRaw)
	}
	if len(edges) != 0 {
		t.Fatalf("edges should be empty, got %d: %v", len(edges), edges)
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
