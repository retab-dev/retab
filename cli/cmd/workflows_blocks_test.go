//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestWorkflowsBlocksCreateHelpShowsExtractReviewConfig(t *testing.T) {
	help := workflowsBlocksCreateCmd.Long + "\n" + workflowsBlocksCreateCmd.Example

	for _, want := range []string{
		"`id` (optional)",
		`"type": "extract"`,
		`"review"`,
		`"predicate"`,
		`"kind": "always"`,
		`"inputs": [{"name": "document", "type": "file", "is_primary": true}]`,
		`any_required_field_null`,
		`confidence_lt`,
		`field_confidence_lt`,
		`top_margin_lt`,
		`boundary_confidence_lt`,
		`split_count_neq`,
		`category_in`,
		`json_condition`,
		`n_consensus`,
		`n_consensus > 1`,
		`output-json-0`,
		`likelihoods.invoice_total`,
		`likelihoods.splits.invoice_type`,
		`"n_consensus": 3`,
		`"path": "likelihoods.invoice_total"`,
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("blocks create help should show review config fragment %q, got:\n%s", want, help)
		}
	}
	if strings.Contains(help, "review gate") {
		t.Fatalf("blocks create help should use review-centric wording, got:\n%s", help)
	}
}

func TestWorkflowsBlocksCreateRenderedHelpShowsConsensusReviewConfig(t *testing.T) {
	help := renderWorkflowBlocksHelpForTest(t, workflowsBlocksCreateCmd)

	for _, want := range []string{
		"Reviewable block types",
		"Consensus criteria require `n_consensus > 1`",
		"`confidence_lt`",
		"`field_confidence_lt`",
		"`top_margin_lt`",
		"`boundary_confidence_lt`",
		"`json_condition`",
		"`likelihoods.invoice_total`",
		"`likelihoods.splits.invoice_type`",
		"extract-consensus-review.json",
		`"n_consensus": 3`,
		`"path": "likelihoods.invoice_total"`,
		`"operator": "is_less_than"`,
		`"value": 0.85`,
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("rendered blocks create help should contain %q, got:\n%s", want, help)
		}
	}
	for _, stale := range []string{"review gate", "`review` block", "`hil`"} {
		if strings.Contains(help, stale) {
			t.Fatalf("rendered blocks create help should not contain stale %q, got:\n%s", stale, help)
		}
	}
}

func TestWorkflowsBlocksCreateConsensusReviewExampleParses(t *testing.T) {
	raw := extractHelpJSONHeredoc(t, workflowsBlocksCreateCmd.Example, "extract-consensus-review.json")
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		t.Fatalf("consensus review example should be valid JSON, got %v for:\n%s", err, raw)
	}

	req, err := parseBlockCreate(obj)
	if err != nil {
		t.Fatalf("consensus review example should parse as a block create request: %v", err)
	}
	if req.ID != nil {
		t.Fatalf("consensus review example should let the server generate an id, got %q", *req.ID)
	}
	if req.Type != "extract" {
		t.Fatalf("consensus review example type = %q, want extract", req.Type)
	}
	if req.Config == nil {
		t.Fatal("consensus review example should preserve config")
	}
	cfg := *req.Config
	if got := cfg["n_consensus"]; got != float64(3) {
		t.Fatalf("n_consensus = %#v, want 3", got)
	}
	if _, ok := cfg["hil"]; ok {
		t.Fatalf("consensus review example should not use legacy config.hil: %#v", cfg)
	}
	assertConsensusReviewPredicate(t, cfg, "json_condition", "likelihoods.invoice_total", "is_less_than", float64(0.85))
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
		`n_consensus`,
		`confidence_lt`,
		`"threshold":0.8`,
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

func TestWorkflowsBlocksUpdateRenderedHelpShowsConsensusPatch(t *testing.T) {
	help := renderWorkflowBlocksHelpForTest(t, workflowsBlocksUpdateCmd)

	for _, want := range []string{
		"For consensus-based review, patch both `n_consensus` and `review`",
		`{"n_consensus":3,"review":{"predicate":{"kind":"confidence_lt","threshold":0.8}}}`,
		"Add consensus review to an extract or classifier block",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("rendered blocks update help should contain %q, got:\n%s", want, help)
		}
	}
}

func TestWorkflowsBlocksUpdateConsensusPatchExampleMerges(t *testing.T) {
	raw := extractPrintfJSONContaining(t, workflowsBlocksUpdateCmd.Example, `"n_consensus":3`)
	var patch map[string]any
	if err := json.Unmarshal([]byte(raw), &patch); err != nil {
		t.Fatalf("consensus review update example should be valid JSON, got %v for:\n%s", err, raw)
	}

	merged := mergeWorkflowBlockConfig(
		map[string]any{
			"model": "retab-small",
			"json_schema": map[string]any{
				"type": "object",
			},
		},
		patch,
	)
	if got := merged["model"]; got != "retab-small" {
		t.Fatalf("merge patch should preserve existing model, got %#v", got)
	}
	if got := merged["n_consensus"]; got != float64(3) {
		t.Fatalf("merged n_consensus = %#v, want 3", got)
	}
	review, ok := merged["review"].(map[string]any)
	if !ok {
		t.Fatalf("merged review should be an object, got %#v", merged["review"])
	}
	predicate, ok := review["predicate"].(map[string]any)
	if !ok {
		t.Fatalf("merged review.predicate should be an object, got %#v", review["predicate"])
	}
	if got := predicate["kind"]; got != "confidence_lt" {
		t.Fatalf("merged predicate kind = %#v, want confidence_lt", got)
	}
	if got := predicate["threshold"]; got != float64(0.8) {
		t.Fatalf("merged predicate threshold = %#v, want 0.8", got)
	}
}

func TestWorkflowsBlocksHelpDoesNotAdvertiseStandaloneReviewBlock(t *testing.T) {
	for _, help := range []string{workflowsBlocksCmd.Long, workflowsBlocksCreateCmd.Long} {
		if strings.Contains(help, "`review`") {
			t.Fatalf("help should not advertise review as a standalone block type:\n%s", help)
		}
	}
	if !strings.Contains(workflowsBlocksCreateCmd.Long, "Review is not a standalone block type") {
		t.Fatalf("blocks create help should state that review is configured inside block configs:\n%s", workflowsBlocksCreateCmd.Long)
	}
}

func renderWorkflowBlocksHelpForTest(t *testing.T, cmd *cobra.Command) string {
	t.Helper()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	t.Cleanup(func() {
		cmd.SetOut(nil)
		cmd.SetErr(nil)
	})
	if err := cmd.Help(); err != nil {
		t.Fatalf("render help for %s: %v", cmd.Use, err)
	}
	return buf.String()
}

func extractHelpJSONHeredoc(t *testing.T, text string, filename string) string {
	t.Helper()
	marker := "cat > " + filename + " <<'JSON'"
	start := strings.Index(text, marker)
	if start < 0 {
		t.Fatalf("could not find heredoc marker %q in:\n%s", marker, text)
	}
	lines := strings.Split(text[start+len(marker):], "\n")
	collected := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "JSON" {
			if len(collected) == 0 {
				t.Fatalf("heredoc %q was empty in:\n%s", filename, text)
			}
			return strings.Join(collected, "\n")
		}
		collected = append(collected, line)
	}
	t.Fatalf("could not find closing JSON marker for %q in:\n%s", filename, text)
	return ""
}

func extractPrintfJSONContaining(t *testing.T, text string, needle string) string {
	t.Helper()
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "printf '") || !strings.Contains(line, needle) {
			continue
		}
		rest := strings.TrimPrefix(line, "printf '")
		end := strings.Index(rest, "'")
		if end < 0 {
			t.Fatalf("printf JSON line is missing closing quote: %s", line)
		}
		return rest[:end]
	}
	t.Fatalf("could not find printf JSON containing %q in:\n%s", needle, text)
	return ""
}

func assertConsensusReviewPredicate(t *testing.T, cfg map[string]any, wantKind string, wantPath string, wantOperator string, wantValue any) {
	t.Helper()
	review, ok := cfg["review"].(map[string]any)
	if !ok {
		t.Fatalf("review should be an object, got %#v", cfg["review"])
	}
	predicate, ok := review["predicate"].(map[string]any)
	if !ok {
		t.Fatalf("review.predicate should be an object, got %#v", review["predicate"])
	}
	if got := predicate["kind"]; got != wantKind {
		t.Fatalf("predicate kind = %#v, want %q", got, wantKind)
	}
	condition, ok := predicate["condition"].(map[string]any)
	if !ok {
		t.Fatalf("predicate.condition should be an object, got %#v", predicate["condition"])
	}
	subConditions, ok := condition["sub_conditions"].([]any)
	if !ok || len(subConditions) != 1 {
		t.Fatalf("sub_conditions = %#v, want one condition", condition["sub_conditions"])
	}
	subCondition, ok := subConditions[0].(map[string]any)
	if !ok {
		t.Fatalf("sub_conditions[0] should be an object, got %#v", subConditions[0])
	}
	pathRef, ok := subCondition["path_ref"].(map[string]any)
	if !ok {
		t.Fatalf("path_ref should be an object, got %#v", subCondition["path_ref"])
	}
	if got := pathRef["path"]; got != wantPath {
		t.Fatalf("path_ref.path = %#v, want %q", got, wantPath)
	}
	if got := subCondition["operator"]; got != wantOperator {
		t.Fatalf("condition operator = %#v, want %q", got, wantOperator)
	}
	if got := subCondition["value"]; got != wantValue {
		t.Fatalf("condition value = %#v, want %#v", got, wantValue)
	}
}

func TestParseBlockCreateAcceptsOmittedID(t *testing.T) {
	// `id` is optional — the server's CreateBlockRequest carries a
	// default_factory that mints an opaque `block_<nanoid>`. Block ids are
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
	if req.ID != nil {
		t.Fatalf("expected ID to remain nil so server's default_factory fires, got %q", *req.ID)
	}
	if req.Type != "extract" {
		t.Fatalf("expected type to round-trip, got %q", req.Type)
	}
}

func TestParseBlockCreatePreservesExplicitZeroPositions(t *testing.T) {
	req, err := parseBlockCreate(map[string]any{
		"type":       "function",
		"position_x": float64(0),
		"position_y": float64(0),
	})
	if err != nil {
		t.Fatalf("zero positions should be accepted, got error: %v", err)
	}
	if req.PositionX == nil || *req.PositionX != 0 {
		t.Fatalf("position_x = %v, want explicit 0 pointer", req.PositionX)
	}
	if req.PositionY == nil || *req.PositionY != 0 {
		t.Fatalf("position_y = %v, want explicit 0 pointer", req.PositionY)
	}
}

func TestParseBlockCreateDefaultsNoteDimensions(t *testing.T) {
	req, err := parseBlockCreate(map[string]any{
		"type": "note",
	})
	if err != nil {
		t.Fatalf("note should be accepted, got error: %v", err)
	}
	if req.Width == nil || *req.Width != defaultNoteBlockWidth {
		t.Fatalf("width = %v, want %v", req.Width, defaultNoteBlockWidth)
	}
	if req.Height == nil || *req.Height != defaultNoteBlockHeight {
		t.Fatalf("height = %v, want %v", req.Height, defaultNoteBlockHeight)
	}
}

func TestParseBlockCreateDefaultsZeroNoteDimensions(t *testing.T) {
	req, err := parseBlockCreate(map[string]any{
		"type":   "note",
		"width":  float64(0),
		"height": float64(0),
	})
	if err != nil {
		t.Fatalf("note should be accepted, got error: %v", err)
	}
	if req.Width == nil || *req.Width != defaultNoteBlockWidth {
		t.Fatalf("width = %v, want %v", req.Width, defaultNoteBlockWidth)
	}
	if req.Height == nil || *req.Height != defaultNoteBlockHeight {
		t.Fatalf("height = %v, want %v", req.Height, defaultNoteBlockHeight)
	}
}

func TestParseBlockCreatePreservesPositiveNoteDimensions(t *testing.T) {
	req, err := parseBlockCreate(map[string]any{
		"type":   "note",
		"width":  float64(220),
		"height": float64(80),
	})
	if err != nil {
		t.Fatalf("note should be accepted, got error: %v", err)
	}
	if req.Width == nil || *req.Width != 220 {
		t.Fatalf("width = %v, want 220", req.Width)
	}
	if req.Height == nil || *req.Height != 80 {
		t.Fatalf("height = %v, want 80", req.Height)
	}
}

func TestParseBlockCreateDefaultsContainerDimensions(t *testing.T) {
	cases := []struct {
		name       string
		blockType  string
		width      any
		height     any
		wantWidth  float64
		wantHeight float64
	}{
		{name: "for_each omitted", blockType: "for_each", wantWidth: 800, wantHeight: 800},
		{name: "while_loop omitted", blockType: "while_loop", wantWidth: 800, wantHeight: 800},
		{name: "for_each zero", blockType: "for_each", width: float64(0), height: float64(0), wantWidth: 800, wantHeight: 800},
		{name: "while_loop negative", blockType: "while_loop", width: float64(-1), height: float64(-2), wantWidth: 800, wantHeight: 800},
		{name: "for_each below minimum clamps independently", blockType: "for_each", width: float64(799), height: float64(900), wantWidth: 800, wantHeight: 900},
		{name: "while_loop above minimum is preserved", blockType: "while_loop", width: float64(950), height: float64(875), wantWidth: 950, wantHeight: 875},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{"type": tc.blockType}
			if tc.width != nil {
				body["width"] = tc.width
			}
			if tc.height != nil {
				body["height"] = tc.height
			}
			req, err := parseBlockCreate(body)
			if err != nil {
				t.Fatalf("%s should be accepted, got error: %v", tc.blockType, err)
			}
			if req.Width == nil || *req.Width != tc.wantWidth {
				t.Fatalf("width = %v, want %v", req.Width, tc.wantWidth)
			}
			if req.Height == nil || *req.Height != tc.wantHeight {
				t.Fatalf("height = %v, want %v", req.Height, tc.wantHeight)
			}
		})
	}
}

func TestParseBlockCreateLeavesRegularBlockDimensionsUnset(t *testing.T) {
	req, err := parseBlockCreate(map[string]any{
		"type": "extract",
	})
	if err != nil {
		t.Fatalf("extract should be accepted, got error: %v", err)
	}
	if req.Width != nil || req.Height != nil {
		t.Fatalf("regular block should not get default dimensions, got width=%v height=%v", req.Width, req.Height)
	}
}

func TestDeriveBlockHandlesUsesRuntimeFunctionInputHandle(t *testing.T) {
	handles := deriveBlockHandles(map[string]any{
		"type": "function",
		"config": map[string]any{
			"code": "from models import Input, Output\n",
			"inputs": []any{
				map[string]any{"name": "payload", "type": "json"},
			},
			"output_schema": map[string]any{"type": "object"},
		},
	})
	inputs, ok := handles["input"].([]string)
	if !ok {
		t.Fatalf("handles[input] = %#v, want []string", handles["input"])
	}
	if len(inputs) != 1 || inputs[0] != "input-json-0" {
		t.Fatalf("function input handles = %#v, want [input-json-0]", inputs)
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
	if !strings.Contains(msg, "workflow_id") && !strings.Contains(msg, "workflow id") {
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

func TestWorkflowsBlocksUpdateRejectsEmptyMergeConfigBeforeHTTP(t *testing.T) {
	// Regression: ``--merge-config-file`` pointing at an empty JSON object
	// used to round-trip ``{"config_mode":"merge"}`` to the server because
	// Go's ``json.Marshal`` drops empty-map ``Config`` fields via
	// ``omitempty``. The server then returned a confusing 422 saying
	// ``config_mode is only meaningful when 'config' is also provided``.
	// The CLI must catch this client-side and explain it before any HTTP
	// call is made.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	httpCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	patchPath := filepath.Join(t.TempDir(), "empty-patch.json")
	if err := os.WriteFile(patchPath, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksUpdateCmd.Flags().Set("merge-config-file", patchPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksUpdateCmd.Flags().Set("merge-config-file", "") })

	err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"blk_123"})
	if err == nil {
		t.Fatal("expected empty merge patch to be rejected")
	}
	if !strings.Contains(err.Error(), "--merge-config-file") {
		t.Fatalf("error should mention --merge-config-file, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "no keys") && !strings.Contains(err.Error(), "empty") {
		t.Fatalf("error should explain the patch is empty, got %q", err.Error())
	}
	if httpCalled {
		t.Fatal("CLI must reject empty merge patch before any HTTP call")
	}
}

func TestWorkflowsBlocksCreateHelpExampleHasUniquePlaceholderID(t *testing.T) {
	// Regression: the inline example used to hard-code
	// ``"id": "extract_review"``. Block ids are unique per organization,
	// so a user pasting the example twice always hit a 409 on the second
	// run. The example must use a placeholder that signals "replace me"
	// — never a fixed literal that obviously collides.
	help := workflowsBlocksCreateCmd.Long + "\n" + workflowsBlocksCreateCmd.Example
	if strings.Contains(help, `"id": "extract_review"`) {
		t.Fatalf("blocks create example must not hard-code a colliding id, got:\n%s", help)
	}
}

func TestWorkflowsBlocksUpdateAcceptsTwoPositionalArgs(t *testing.T) {
	// `blocks update wf_x blk_y` should resolve to the canonical block-id
	// shape on the wire (PATCH /v1/workflows/blocks/blk_y) and wire the
	// positional workflow id through to the `--workflow-id` query string,
	// matching what `--workflow-id wf_x` produces.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var gotPath, gotQuery, gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_y",
			"workflow_id": "wf_x",
			"type":        "extract",
			"config":      map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksUpdateCmd.Flags().Set("label", "renamed"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksUpdateCmd.Flags().Set("label", "")
		// Reset workflow-id so subsequent tests start clean (positional path sets it).
		_ = workflowsBlocksUpdateCmd.Flags().Set("workflow-id", "")
	})

	if err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"wf_x", "blk_y"}); err != nil {
		t.Fatalf("update with two positionals: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Fatalf("method = %s, want PATCH", gotMethod)
	}
	if gotPath != "/v1/workflows/blocks/blk_y" {
		t.Fatalf("path = %s, want /v1/workflows/blocks/blk_y", gotPath)
	}
	if !strings.Contains(gotQuery, "workflow_id=wf_x") {
		t.Fatalf("query = %s, want workflow_id=wf_x", gotQuery)
	}
}

func TestWorkflowsBlocksUpdateAcceptsSinglePositionalArg(t *testing.T) {
	// Backward-compat: `blocks update <block-id>` keeps its original shape
	// with no workflow_id query parameter (server resolves by org-unique id).
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_y",
			"workflow_id": "wf_x",
			"type":        "extract",
			"config":      map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksUpdateCmd.Flags().Set("label", "renamed"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksUpdateCmd.Flags().Set("label", "")
		_ = workflowsBlocksUpdateCmd.Flags().Set("workflow-id", "")
	})

	if err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"blk_y"}); err != nil {
		t.Fatalf("update with one positional: %v", err)
	}

	if gotPath != "/v1/workflows/blocks/blk_y" {
		t.Fatalf("path = %s, want /v1/workflows/blocks/blk_y", gotPath)
	}
	if gotQuery != "" {
		t.Fatalf("query = %s, want empty", gotQuery)
	}
}

func TestWorkflowsBlocksUpdateRejectsConflictingWorkflowID(t *testing.T) {
	// `blocks update wf_a blk_y --workflow-id wf_b` is ambiguous: the user
	// asked for two different workflow ids in the same invocation. The CLI
	// must refuse with a clear conflict message rather than silently picking
	// one and issuing a request that lands on the wrong workflow.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	httpCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksUpdateCmd.Flags().Set("workflow-id", "wf_b"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksUpdateCmd.Flags().Set("label", "renamed"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksUpdateCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksUpdateCmd.Flags().Set("label", "")
	})

	err := workflowsBlocksUpdateCmd.RunE(workflowsBlocksUpdateCmd, []string{"wf_a", "blk_y"})
	if err == nil {
		t.Fatal("expected conflicting workflow id to be rejected")
	}
	msg := err.Error()
	if !strings.Contains(msg, "conflict") {
		t.Fatalf("error should mention conflict, got %q", msg)
	}
	if !strings.Contains(msg, "wf_a") || !strings.Contains(msg, "wf_b") {
		t.Fatalf("error should name both workflow ids, got %q", msg)
	}
	if httpCalled {
		t.Fatal("CLI must reject conflict before any HTTP call")
	}
}

func TestWorkflowsBlocksGetAcceptsTwoPositionalArgs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_y",
			"workflow_id": "wf_x",
			"type":        "extract",
			"config":      map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	t.Cleanup(func() { _ = workflowsBlocksGetCmd.Flags().Set("workflow-id", "") })

	if err := workflowsBlocksGetCmd.RunE(workflowsBlocksGetCmd, []string{"wf_x", "blk_y"}); err != nil {
		t.Fatalf("get with two positionals: %v", err)
	}
	if gotPath != "/v1/workflows/blocks/blk_y" {
		t.Fatalf("path = %s, want /v1/workflows/blocks/blk_y", gotPath)
	}
	if !strings.Contains(gotQuery, "workflow_id=wf_x") {
		t.Fatalf("query = %s, want workflow_id=wf_x", gotQuery)
	}
}

func TestWorkflowsBlocksGetRejectsConflictingWorkflowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	httpCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksGetCmd.Flags().Set("workflow-id", "wf_b"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksGetCmd.Flags().Set("workflow-id", "") })

	err := workflowsBlocksGetCmd.RunE(workflowsBlocksGetCmd, []string{"wf_a", "blk_y"})
	if err == nil {
		t.Fatal("expected conflicting workflow id to be rejected")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("error should mention conflict, got %q", err.Error())
	}
	if httpCalled {
		t.Fatal("CLI must reject conflict before any HTTP call")
	}
}

func TestWorkflowsBlocksDeleteAcceptsTwoPositionalArgs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var gotPath, gotQuery, gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksDeleteCmd.Flags().Set("yes", "false")
		_ = workflowsBlocksDeleteCmd.Flags().Set("workflow-id", "")
	})

	_, _ = captureStd(t, func() {
		if err := workflowsBlocksDeleteCmd.RunE(workflowsBlocksDeleteCmd, []string{"wf_x", "blk_y"}); err != nil {
			t.Fatalf("delete with two positionals: %v", err)
		}
	})

	if gotMethod != http.MethodDelete {
		t.Fatalf("method = %s, want DELETE", gotMethod)
	}
	if gotPath != "/v1/workflows/blocks/blk_y" {
		t.Fatalf("path = %s, want /v1/workflows/blocks/blk_y", gotPath)
	}
	if !strings.Contains(gotQuery, "workflow_id=wf_x") {
		t.Fatalf("query = %s, want workflow_id=wf_x", gotQuery)
	}
}

func TestWorkflowsBlocksDeleteRejectsConflictingWorkflowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	httpCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksDeleteCmd.Flags().Set("workflow-id", "wf_b"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksDeleteCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksDeleteCmd.Flags().Set("yes", "false")
	})

	err := workflowsBlocksDeleteCmd.RunE(workflowsBlocksDeleteCmd, []string{"wf_a", "blk_y"})
	if err == nil {
		t.Fatal("expected conflicting workflow id to be rejected")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("error should mention conflict, got %q", err.Error())
	}
	if httpCalled {
		t.Fatal("CLI must reject conflict before any HTTP call")
	}
}

func TestWorkflowsBlocksGetHonorsTableOutputFallback(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/workflows/blocks/blk_1" || r.URL.RawQuery != "" {
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
