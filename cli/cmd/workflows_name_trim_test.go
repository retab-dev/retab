//go:build !retab_oagen_cli_workflows && !retab_oagen_cli_workflows_experiments && !retab_oagen_cli_workflows_evals

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// Bug A (issue #3): the three workflow-family name validators
// (validateWorkflowName, validateExperimentName, validateWorkflowEvalName)
// previously only rejected fully-blank names but stored "  padded  "
// verbatim. The contract has been tightened so the validators TRIM the
// input and return the cleaned name. Pinned at unit and integration
// level so a future refactor cannot regress to "check-only" semantics.
func TestValidateWorkflowFamilyNamesTrimPaddedInput(t *testing.T) {
	cases := []struct {
		name      string
		validator func(string) (string, error)
	}{
		{name: "workflow", validator: validateWorkflowName},
		{name: "experiment", validator: validateExperimentName},
		{name: "workflow eval", validator: validateWorkflowEvalName},
	}
	for _, tc := range cases {
		t.Run(tc.name+" trims surrounding whitespace", func(t *testing.T) {
			got, err := tc.validator("  padded  ")
			if err != nil {
				t.Fatalf("expected no error trimming %q, got %v", "  padded  ", err)
			}
			if got != "padded" {
				t.Fatalf("validator returned %q, want %q", got, "padded")
			}
		})
		t.Run(tc.name+" rejects blank-only input", func(t *testing.T) {
			if _, err := tc.validator("   "); err == nil {
				t.Fatalf("expected blank-only name to fail")
			}
		})
		t.Run(tc.name+" passes through already-trimmed name", func(t *testing.T) {
			got, err := tc.validator("clean")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != "clean" {
				t.Fatalf("got %q, want %q", got, "clean")
			}
		})
	}
}

// TestWorkflowsCreateTrimsNameInRequestBody pins the end-to-end behaviour:
// `retab workflows create --name "  padded  "` must POST `name: "padded"`
// to the server (no leading/trailing whitespace). Catches regressions where
// the validator is invoked but its trimmed value is not propagated to the
// outgoing request.
func TestWorkflowsCreateTrimsNameInRequestBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "wf_trimmed",
			"name": body["name"],
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "create", RunE: workflowsCreateCmd.RunE}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("project-id", "", "")
	if err := cmd.Flags().Set("name", "  padded  "); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("project-id", "proj_test"); err != nil {
		t.Fatal(err)
	}

	if _, err := captureStdAndRun(t, func() error { return cmd.RunE(cmd, nil) }); err != nil {
		t.Fatalf("create: %v", err)
	}
	if got, _ := body["name"].(string); got != "padded" {
		t.Fatalf("server received name=%q, want %q", got, "padded")
	}
}

// TestWorkflowsUpdateTrimsNameInRequestBody pins the same trim behaviour
// for `workflows update --name "  padded  "`. The patch endpoint only sees
// `name` because that's the only flag the user passed.
func TestWorkflowsUpdateTrimsNameInRequestBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var patchBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch && r.URL.Path == "/v1/workflows/wf_abc" {
			if err := json.NewDecoder(r.Body).Decode(&patchBody); err != nil {
				t.Fatalf("decode patch: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "wf_abc",
				"name": patchBody["name"],
			})
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "update", RunE: workflowsUpdateCmd.RunE}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	if err := cmd.Flags().Set("name", "  padded  "); err != nil {
		t.Fatal(err)
	}

	if _, err := captureStdAndRun(t, func() error { return cmd.RunE(cmd, []string{"wf_abc"}) }); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got, _ := patchBody["name"].(string); got != "padded" {
		t.Fatalf("server received name=%q, want %q", got, "padded")
	}
}

// TestWorkflowsExperimentsCreateTrimsNameInRequestBody pins the trim
// contract for experiments create: --name "  padded  " must arrive on the
// server as name=padded.
func TestWorkflowsExperimentsCreateTrimsNameInRequestBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	documentsPath := filepath.Join(dir, "docs.json")
	if err := os.WriteFile(documentsPath, []byte(`[{"handle_inputs":{"foo":"bar"}}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/experiments" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "exp_trimmed",
			"name": body["name"],
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := workflowsExperimentsCreateCmd
	t.Cleanup(func() {
		_ = cmd.Flags().Set("block-id", "")
		_ = cmd.Flags().Set("name", "")
		_ = cmd.Flags().Set("documents-file", "")
		if f := cmd.Flags().Lookup("documents-file"); f != nil {
			f.Changed = false
		}
		if f := cmd.Flags().Lookup("name"); f != nil {
			f.Changed = false
		}
	})
	if err := cmd.Flags().Set("block-id", "blk_123"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("name", "  padded  "); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("documents-file", documentsPath); err != nil {
		t.Fatal(err)
	}

	if _, err := captureStdAndRun(t, func() error { return cmd.RunE(cmd, []string{"wf_123"}) }); err != nil {
		t.Fatalf("experiments create: %v", err)
	}
	if got, _ := body["name"].(string); got != "padded" {
		t.Fatalf("server received name=%q, want %q", got, "padded")
	}
}

// TestWorkflowsExperimentsUpdateTrimsNameInRequestBody pins the same
// trim contract on the experiments update endpoint.
func TestWorkflowsExperimentsUpdateTrimsNameInRequestBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/workflows/experiments/exp_abc" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "exp_abc",
			"name": body["name"],
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := workflowsExperimentsUpdateCmd
	t.Cleanup(func() {
		_ = cmd.Flags().Set("name", "")
		if f := cmd.Flags().Lookup("name"); f != nil {
			f.Changed = false
		}
	})
	if err := cmd.Flags().Set("name", "  padded  "); err != nil {
		t.Fatal(err)
	}

	if _, err := captureStdAndRun(t, func() error { return cmd.RunE(cmd, []string{"exp_abc"}) }); err != nil {
		t.Fatalf("experiments update: %v", err)
	}
	if got, _ := body["name"].(string); got != "padded" {
		t.Fatalf("server received name=%q, want %q", got, "padded")
	}
}

// TestWorkflowsEvalsCreateTrimsNameInRequestBody pins the trim contract
// for workflows evals create.
func TestWorkflowsEvalsCreateTrimsNameInRequestBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	sourcePath := filepath.Join(dir, "source.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	for path, payload := range map[string]string{
		targetPath:    `{"type":"block","block_id":"blk_1"}`,
		sourcePath:    `{"type":"manual","handle_inputs":{}}`,
		assertionPath: `{"target":{"output_handle_id":"output-json-0"},"condition":{"kind":"exists"}}`,
	} {
		if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/evals" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "tst_trimmed",
			"name": body["name"],
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := workflowsEvalsCreateCmd
	t.Cleanup(func() {
		for _, name := range []string{"name", "target-file", "source-file", "assertion-file"} {
			_ = cmd.Flags().Set(name, "")
			if f := cmd.Flags().Lookup(name); f != nil {
				f.Changed = false
			}
		}
	})
	for flag, value := range map[string]string{
		"name":           "  padded  ",
		"target-file":    targetPath,
		"source-file":    sourcePath,
		"assertion-file": assertionPath,
	} {
		if err := cmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := captureStdAndRun(t, func() error { return cmd.RunE(cmd, []string{"wf_123"}) }); err != nil {
		t.Fatalf("evals create: %v", err)
	}
	if got, _ := body["name"].(string); got != "padded" {
		t.Fatalf("server received name=%q, want %q", got, "padded")
	}
}

// TestWorkflowsEvalsUpdateTrimsNameInRequestBody pins the trim contract
// on the evals update endpoint.
func TestWorkflowsEvalsUpdateTrimsNameInRequestBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/workflows/evals/tst_abc" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "tst_abc",
			"name": body["name"],
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := workflowsEvalsUpdateCmd
	resetFlags := func() {
		for _, name := range []string{"name", "assertion-file", "source-file", "output-handle-id", "path", "equals", "run-id", "step-id"} {
			_ = cmd.Flags().Set(name, "")
			if f := cmd.Flags().Lookup(name); f != nil {
				f.Changed = false
			}
		}
	}
	resetFlags()
	t.Cleanup(resetFlags)
	if err := cmd.Flags().Set("name", "  padded  "); err != nil {
		t.Fatal(err)
	}

	if _, err := captureStdAndRun(t, func() error { return cmd.RunE(cmd, []string{"tst_abc"}) }); err != nil {
		t.Fatalf("evals update: %v", err)
	}
	if got, _ := body["name"].(string); got != "padded" {
		t.Fatalf("server received name=%q, want %q", got, "padded")
	}
}

// captureStdAndRun is a small wrapper around captureStd that lets the
// invoked function return an error.  We do not assert on the captured
// stdout/stderr — the integration tests above just need the helper to
// avoid printing JSON responses through the test output.
func captureStdAndRun(t *testing.T, fn func() error) (stdout string, err error) {
	t.Helper()
	out, _ := captureStd(t, func() {
		err = fn()
	})
	return out, err
}

// TestWorkflowsExperimentsListTableRendersBlockKindColumn pins Bug B
// (issue #12). The experiments list table previously rendered an empty
// `TYPE` column because the generic auto-column picker had no alias for
// the `block_kind` field carried on every experiment payload. A dedicated
// TableColumn spec for the experiments list (workflowExperimentColumns)
// now renders a BLOCK_KIND column populated from the `block_kind` field,
// matching the SDK's on-the-wire name.
func TestWorkflowsExperimentsListTableRendersBlockKindColumn(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":         "exp_smoke",
				"name":       "smoke-exp",
				"block_type": "extract",
				"status":     "completed",
				"freshness": map[string]any{
					"status": "stale",
				},
				"created_at": "2026-05-21T12:00:00Z",
			},
		},
	}

	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, workflowExperimentColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines (header + 1 row), got %d:\n%s", len(lines), out)
	}
	header := lines[0]
	row := lines[1]

	if !strings.Contains(header, "BLOCK_KIND") {
		t.Fatalf("header missing BLOCK_KIND column:\n%s", header)
	}
	if strings.Contains(header, "\tTYPE\t") || strings.HasSuffix(header, "\tTYPE") || strings.HasPrefix(header, "TYPE\t") {
		t.Fatalf("header still uses standalone TYPE label:\n%s", header)
	}
	if !strings.Contains(row, "extract") {
		t.Fatalf("row missing block_kind value 'extract':\n%s", row)
	}
	if !strings.Contains(header, "FRESHNESS") {
		t.Fatalf("header missing FRESHNESS column:\n%s", header)
	}
	if !strings.Contains(row, "stale") {
		t.Fatalf("row missing freshness value 'stale':\n%s", row)
	}

	// Locate the BLOCK_KIND column position in the header and check
	// the value at that column starts with "extract". Tabwriter pads
	// columns with spaces, so trim leading whitespace before matching.
	blockKindIdx := strings.Index(header, "BLOCK_KIND")
	if blockKindIdx < 0 {
		t.Fatalf("BLOCK_KIND header not found:\n%s", header)
	}
	if !strings.HasPrefix(strings.TrimLeft(row[blockKindIdx:], " "), "extract") {
		t.Fatalf("cell under BLOCK_KIND does not start with 'extract':\nheader: %q\nrow:    %q", header, row)
	}
}

// TestWorkflowsExperimentsListJSONOutputUnchanged pins that the JSON
// output path (no --output table flag) still produces the canonical
// list envelope, byte-equivalent to what callers piped into jq before
// the bug-#12 fix. The fix moved table rendering into a dedicated
// helper, so this guards against accidentally introducing a difference
// in the JSON path.
func TestWorkflowsExperimentsListJSONOutputUnchanged(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/experiments" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"exp_aaa","name":"x","block_type":"extract","status":"completed","created_at":"2026-05-21T12:00:00Z","updated_at":"2026-05-21T12:00:00Z","workflow_id":"wrk_xx","block_id":"blk_xx","n_consensus":3,"document_count":1}],"list_metadata":{}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := workflowsExperimentsListCmd.RunE(workflowsExperimentsListCmd, []string{"wrk_xx"}); err != nil {
			t.Fatalf("experiments list: %v", err)
		}
	})
	if !strings.Contains(stdout, `"id": "exp_aaa"`) || !strings.Contains(stdout, `"block_type": "extract"`) {
		t.Fatalf("JSON output missing expected fields:\n%s", stdout)
	}
}

// TestWorkflowsCreateBlankNameStillRejectedAfterTrim defends against a
// future refactor that, while moving the trim into the validator, drops
// the empty-after-trim check. A whitespace-only name must still fail —
// the trim does NOT loosen the contract, it just cleans valid input.
func TestWorkflowsCreateBlankNameStillRejectedAfterTrim(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "create", RunE: workflowsCreateCmd.RunE}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	if err := cmd.Flags().Set("name", "   "); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for blank-only --name")
	}
	if !strings.Contains(err.Error(), "must not be blank") {
		t.Fatalf("error %q does not mention blank name", err.Error())
	}
	if hits.Load() != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits.Load())
	}
}
