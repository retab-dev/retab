package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowsBlocksCreateHelpShowsExtractReviewConfig(t *testing.T) {
	help := workflowsBlocksCreateCmd.Long + "\n" + workflowsBlocksCreateCmd.Example

	for _, want := range []string{
		`"type": "extract"`,
		`"review"`,
		`"predicate"`,
		`"kind": "always"`,
		`"inputs": [{"name": "document", "type": "file", "is_primary": true}]`,
		`any_required_field_null`,
		`split_count_neq`,
		`category_in`,
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("blocks create help should show review config fragment %q, got:\n%s", want, help)
		}
	}
	if strings.Contains(help, "review gate") {
		t.Fatalf("blocks create help should use review-centric wording, got:\n%s", help)
	}
}

func TestWorkflowsBlocksHelpUsesReviewCentricBlockNames(t *testing.T) {
	help := workflowsBlocksCmd.Long + "\n" + workflowsBlocksCmd.Example

	for _, want := range []string{"`classifier`", "review config", "--merge-config-file"} {
		if !strings.Contains(help, want) {
			t.Fatalf("blocks help should mention %q, got:\n%s", want, help)
		}
	}
	for _, stale := range []string{"`classify`", "review gate", "review gates"} {
		if strings.Contains(help, stale) {
			t.Fatalf("blocks help should not use stale %q, got:\n%s", stale, help)
		}
	}
}

func TestWorkflowsBlocksUpdateHelpShowsReviewConfig(t *testing.T) {
	help := workflowsBlocksUpdateCmd.Long + "\n" + workflowsBlocksUpdateCmd.Example

	for _, want := range []string{
		`"review"`,
		`"predicate"`,
		`"kind":"always"`,
		`--merge-config-file -`,
		`only when replacing the whole typed config`,
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("blocks update help should show review config guidance %q, got:\n%s", want, help)
		}
	}
	for _, stale := range []string{"review gate", "hil"} {
		if strings.Contains(help, stale) {
			t.Fatalf("blocks update help should use review-centric wording without %q, got:\n%s", stale, help)
		}
	}
}

func TestWorkflowsBlocksHelpDoesNotAdvertiseStandaloneReviewBlock(t *testing.T) {
	for _, help := range []string{workflowsBlocksCmd.Long, workflowsBlocksCreateCmd.Long, workflowsBlocksCreateBatchCmd.Long} {
		if strings.Contains(help, "`review`") {
			t.Fatalf("help should not advertise review as a standalone block type:\n%s", help)
		}
	}
	if !strings.Contains(workflowsBlocksCreateCmd.Long, "Review is not a standalone block type") {
		t.Fatalf("blocks create help should explicitly steer users away from standalone review blocks:\n%s", workflowsBlocksCreateCmd.Long)
	}
	if !strings.Contains(workflowsBlocksCreateBatchCmd.Long, "Review is not a standalone block type") {
		t.Fatalf("blocks create-batch help should explicitly steer users away from standalone review blocks:\n%s", workflowsBlocksCreateBatchCmd.Long)
	}
}

func TestParseBlockCreateRejectsStandaloneReviewBlock(t *testing.T) {
	_, err := parseBlockCreate(map[string]any{
		"id":   "legacy_review",
		"type": "review",
	})
	if err == nil {
		t.Fatal("expected standalone review block to be rejected")
	}
	if !strings.Contains(err.Error(), "config.review") || !strings.Contains(err.Error(), "reviewable block") {
		t.Fatalf("error should guide toward embedded review config, got %q", err.Error())
	}
}

func TestParseBlockCreateRejectsLegacyHilConfig(t *testing.T) {
	_, err := parseBlockCreate(map[string]any{
		"id":   "extract_hil",
		"type": "extract",
		"config": map[string]any{
			"hil": map[string]any{"predicate": map[string]any{"kind": "always"}},
		},
	})
	if err == nil {
		t.Fatal("expected legacy config.hil to be rejected")
	}
	if !strings.Contains(err.Error(), "config.hil") || !strings.Contains(err.Error(), "config.review") {
		t.Fatalf("error should guide toward config.review, got %q", err.Error())
	}
}

func TestWorkflowsBlocksCreateReadsLocalFileBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-block.json")
	if err := workflowsBlocksCreateCmd.Flags().Set("block-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksCreateCmd.Flags().Set("block-file", "") })

	err := workflowsBlocksCreateCmd.RunE(workflowsBlocksCreateCmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected missing block file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "missing-block.json") {
		t.Fatalf("error should mention missing block file, got %q", err.Error())
	}
}

func TestWorkflowsBlocksCreateBatchReadsLocalFileBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-blocks.json")
	if err := workflowsBlocksCreateBatchCmd.Flags().Set("blocks-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksCreateBatchCmd.Flags().Set("blocks-file", "") })

	err := workflowsBlocksCreateBatchCmd.RunE(workflowsBlocksCreateBatchCmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected missing blocks file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "missing-blocks.json") {
		t.Fatalf("error should mention missing blocks file, got %q", err.Error())
	}
}

func TestWorkflowsBlocksUpdateReadsConfigFileBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-config.json")
	if err := workflowsBlocksUpdateCmd.Flags().Set("config-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksUpdateCmd.Flags().Set("config-file", "") })

	err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"wf_123", "blk_123"})
	if err == nil {
		t.Fatal("expected missing config file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "--config-file") || !strings.Contains(err.Error(), "missing-config.json") {
		t.Fatalf("error should mention missing config file, got %q", err.Error())
	}
}

func TestWorkflowsBlocksUpdateRejectsLegacyHilConfigBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	configPath := filepath.Join(t.TempDir(), "hil-config.json")
	if err := os.WriteFile(configPath, []byte(`{"hil":{"predicate":{"kind":"always"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksUpdateCmd.Flags().Set("config-file", configPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksUpdateCmd.Flags().Set("config-file", "") })

	err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"wf_123", "blk_123"})
	if err == nil {
		t.Fatal("expected legacy config.hil error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("config.hil validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "config.hil") || !strings.Contains(err.Error(), "config.review") {
		t.Fatalf("error should guide toward config.review, got %q", err.Error())
	}
}

func TestWorkflowsBlocksGetHonorsTableOutputFallback(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/workflows/blocks/blk_1" || r.URL.Query().Get("workflow_id") != "wf_blocks" {
			t.Fatalf("path = %s?%s, want block get", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_1",
			"workflow_id": "wf_blocks",
			"type":        "split",
			"config":      map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsBlocksGetCmd.RunE(workflowsBlocksGetCmd, []string{"wf_blocks", "blk_1"}); err != nil {
			t.Fatalf("blocks get: %v", err)
		}
	})
	if !strings.Contains(stderr, "falling back to json") {
		t.Fatalf("expected table fallback warning, got stderr %q", stderr)
	}
	if !strings.Contains(stdout, `"id": "blk_1"`) {
		t.Fatalf("expected JSON fallback payload, got:\n%s", stdout)
	}
}
