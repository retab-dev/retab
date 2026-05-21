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
		`when replacing the whole typed config`,
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
	for _, help := range []string{workflowsBlocksCmd.Long, workflowsBlocksCreateCmd.Long} {
		if strings.Contains(help, "`review`") {
			t.Fatalf("help should not advertise review as a standalone block type:\n%s", help)
		}
	}
	if !strings.Contains(workflowsBlocksCreateCmd.Long, "Review is not a standalone block type") {
		t.Fatalf("blocks create help should explicitly steer users away from standalone review blocks:\n%s", workflowsBlocksCreateCmd.Long)
	}
}

func TestParseBlockCreateAcceptsOmittedID(t *testing.T) {
	// `id` is optional — the server's CreateBlockRequest carries a
	// default_factory that mints an opaque `blk_<nanoid>`. Block ids are
	// org-globally unique, so forcing a user-chosen id was a footgun: pick
	// any common name and you collide with another workflow. The server's
	// own 409 already tells users to "Create blocks with server-generated
	// opaque IDs" — that path must be reachable through the CLI.
	req, err := parseBlockCreate(map[string]any{
		"type": "extract",
	})
	if err != nil {
		t.Fatalf("omitted id should be accepted, got error: %v", err)
	}
	if req.ID != "" {
		t.Fatalf("expected ID to remain empty so server's default_factory fires, got %q", req.ID)
	}
	if req.Type != "extract" {
		t.Fatalf("expected type to round-trip, got %q", req.Type)
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

func TestParseBlockCreateRejectsMismatchedWorkflowID(t *testing.T) {
	// If the block-file body carries a ``workflow_id`` that disagrees
	// with the positional ``<workflow-id>``, the CLI must reject the
	// request rather than silently dropping the body field. Otherwise a
	// stale automation script can target the wrong workflow and the user
	// will never know.
	_, err := parseBlockCreateForWorkflow(
		"wrk_REAL",
		map[string]any{
			"id":          "x",
			"workflow_id": "wrk_FAKE",
			"type":        "extract",
		},
	)
	if err == nil {
		t.Fatal("expected mismatched workflow_id to be rejected")
	}
	msg := err.Error()
	if !strings.Contains(msg, "workflow_id") {
		t.Fatalf("error should mention workflow_id, got %q", msg)
	}
	if !strings.Contains(msg, "wrk_REAL") || !strings.Contains(msg, "wrk_FAKE") {
		t.Fatalf("error should name both ids, got %q", msg)
	}
}

func TestParseBlockCreateAcceptsMatchingOrAbsentWorkflowID(t *testing.T) {
	// Absent in body → fine (positional wins, body never said anything).
	if _, err := parseBlockCreateForWorkflow(
		"wrk_REAL",
		map[string]any{"id": "x", "type": "extract"},
	); err != nil {
		t.Fatalf("absent body workflow_id should be accepted, got %v", err)
	}
	// Matching in body → fine (echo of the positional).
	if _, err := parseBlockCreateForWorkflow(
		"wrk_REAL",
		map[string]any{
			"id":          "x",
			"workflow_id": "wrk_REAL",
			"type":        "extract",
		},
	); err != nil {
		t.Fatalf("matching body workflow_id should be accepted, got %v", err)
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

func TestWorkflowsBlocksUpdateReadsConfigFileBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-config.json")
	if err := workflowsBlocksUpdateCmd.Flags().Set("config-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksUpdateCmd.Flags().Set("config-file", "") })

	err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"blk_123"})
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

	err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"blk_123"})
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
		if r.URL.Path != "/workflows/blocks/blk_1" || r.URL.RawQuery != "" {
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
		if err := workflowsBlocksGetCmd.RunE(workflowsBlocksGetCmd, []string{"blk_1"}); err != nil {
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
