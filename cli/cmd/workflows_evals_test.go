//go:build !retab_oagen_cli_workflows_evals

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// makeCmdWithWorkflowIDFlag mimics the flag wiring used by the workflows
// subcommands that have migrated workflow-id from a required flag to a
// positional argument with a hidden --workflow-id fallback. The flag is
// MarkHidden (not MarkDeprecated) so cobra does NOT emit its own
// auto-deprecation warning that would duplicate the one in
// resolveWorkflowIDArgTo.
func makeCmdWithWorkflowIDFlag() *cobra.Command {
	c := &cobra.Command{Use: "fake"}
	c.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	_ = c.Flags().MarkHidden("workflow-id")
	return c
}

func TestResolveWorkflowIDArg_PositionalAlone(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags(nil); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, []string{"wf_abc"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_abc" {
		t.Fatalf("got %q want wf_abc", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

// evals get/update/delete are eval-scoped (the wfnodeeval_… id is globally
// unique), but users coming from `evals create <workflow-id>` habitually prefix
// the workflow id. The commands accept an optional leading workflow id and use
// the LAST positional as the eval id, so both `get <eval-id>` and
// `get <workflow-id> <eval-id>` hit the same flat /v1/workflows/evals/<eval-id>
// route instead of erroring "accepts 1 arg(s), received 2".
func TestWorkflowsEvalsGetAcceptsOptionalWorkflowIDPrefix(t *testing.T) {
	for _, args := range [][]string{
		{"wfnodeeval_123"},
		{"wrk_456", "wfnodeeval_123"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			var requests []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"id": "wfnodeeval_123"})
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := workflowsEvalsGetCmd.Args(workflowsEvalsGetCmd, args); err != nil {
				t.Fatalf("arg validation rejected %v: %v", args, err)
			}
			_, _ = captureStd(t, func() {
				if err := workflowsEvalsGetCmd.RunE(workflowsEvalsGetCmd, args); err != nil {
					t.Fatalf("evals get %v: %v", args, err)
				}
			})
			if strings.Join(requests, ",") != "GET /v1/workflows/evals/wfnodeeval_123" {
				t.Fatalf("args %v → requests %v, want the flat eval-id route", args, requests)
			}
		})
	}
}

// `experiments create --run --wait` reuses runs-create's wait machinery, so it
// must expose the same cadence/timeout knobs — otherwise the flags error
// "unknown flag" on the create command.
func TestWorkflowsExperimentsCreateExposesWaitFlags(t *testing.T) {
	for _, name := range []string{"poll-interval-ms", "timeout-seconds"} {
		if workflowsExperimentsCreateCmd.Flags().Lookup(name) == nil {
			t.Fatalf("experiments create is missing --%s (needed for --run --wait)", name)
		}
	}
}

func TestWorkflowsEvalsResultsGetUsesFlatResultIDRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/evals/results/wfresult_123" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":      "wfresult_123",
			"run_id":  "wfevalrun_123",
			"eval_id": "wfnodeeval_123",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsEvalsResultsGetCmd.RunE(workflowsEvalsResultsGetCmd, []string{"wfresult_123"}); err != nil {
			t.Fatalf("eval result get: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.Join(requests, ",") != "GET /v1/workflows/evals/results/wfresult_123" {
		t.Fatalf("requests = %v", requests)
	}
	if !strings.Contains(stdout, `"id": "wfresult_123"`) {
		t.Fatalf("expected stdout to contain result id, got:\n%s", stdout)
	}
}

func TestWorkflowsEvalsRunsCreateSendsScopedBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	if err := os.WriteFile(targetPath, []byte(`{"type":"block","block_id":"extract_1"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name     string
		flags    map[string]string
		wantBody map[string]any
	}{
		{
			name:  "workflow scope",
			flags: map[string]string{},
			wantBody: map[string]any{
				"workflow_id": "wf_123",
			},
		},
		{
			name: "single scope",
			flags: map[string]string{
				"eval-id": "wfnodeeval_123",
			},
			wantBody: map[string]any{
				"workflow_id": "wf_123",
				"scope": map[string]any{
					"type":    "single",
					"eval_id": "wfnodeeval_123",
				},
			},
		},
		{
			name: "block scope",
			flags: map[string]string{
				"target-file": targetPath,
			},
			wantBody: map[string]any{
				"workflow_id": "wf_123",
				"scope": map[string]any{
					"type":     "block",
					"block_id": "extract_1",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/evals/runs" {
					t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"wfevalrun_123"}`))
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			for name, value := range tc.flags {
				if err := workflowsEvalsRunsCreateCmd.Flags().Set(name, value); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = workflowsEvalsRunsCreateCmd.Flags().Set(name, "") })
			}

			var err error
			captureStd(t, func() {
				err = workflowsEvalsRunsCreateCmd.RunE(workflowsEvalsRunsCreateCmd, []string{"wf_123"})
			})
			if err != nil {
				t.Fatalf("runs create: %v", err)
			}
			if !reflect.DeepEqual(gotBody, tc.wantBody) {
				t.Fatalf("body = %#v, want %#v", gotBody, tc.wantBody)
			}
		})
	}
}

func TestResolveWorkflowIDArg_FlagAloneEmitsWarning(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags([]string{"--workflow-id", "wf_flag"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_flag" {
		t.Fatalf("got %q want wf_flag", got)
	}
	// Exactly one line, with the "pass as first positional" wording.
	lines := nonEmptyLines(warn.String())
	if len(lines) != 1 {
		t.Fatalf("expected exactly one warning line, got %d: %q", len(lines), warn.String())
	}
	want := "warning: --workflow-id is deprecated; pass the workflow id as the first positional argument"
	if lines[0] != want {
		t.Fatalf("warning mismatch:\n  got:  %q\n  want: %q", lines[0], want)
	}
}

func TestResolveWorkflowIDArg_PositionalWinsOverFlag(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags([]string{"--workflow-id", "wf_flag"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, []string{"wf_pos"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_pos" {
		t.Fatalf("got %q want wf_pos (positional must win)", got)
	}
	lines := nonEmptyLines(warn.String())
	if len(lines) != 1 {
		t.Fatalf("expected exactly one warning line, got %d: %q", len(lines), warn.String())
	}
	want := "warning: --workflow-id is deprecated; positional argument takes precedence"
	if lines[0] != want {
		t.Fatalf("warning mismatch:\n  got:  %q\n  want: %q", lines[0], want)
	}
}

func TestResolveWorkflowIDArg_NeitherErrors(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags(nil); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	_, err := resolveWorkflowIDArgTo(c, nil, &warn)
	if err == nil {
		t.Fatalf("expected error when neither positional nor flag is set")
	}
	if !strings.Contains(err.Error(), "workflow id required") {
		t.Fatalf("expected 'workflow id required' message, got %q", err.Error())
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning when neither is set, got %q", warn.String())
	}
}

func TestResolveWorkflowIDArg_BlankValuesError(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		flag string
	}{
		{name: "blank positional", args: []string{"   "}},
		{name: "blank deprecated flag", flag: "   "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := makeCmdWithWorkflowIDFlag()
			if tc.flag != "" {
				if err := c.ParseFlags([]string{"--workflow-id", tc.flag}); err != nil {
					t.Fatalf("parse: %v", err)
				}
			} else if err := c.ParseFlags(nil); err != nil {
				t.Fatalf("parse: %v", err)
			}
			var warn bytes.Buffer
			_, err := resolveWorkflowIDArgTo(c, tc.args, &warn)
			if err == nil {
				t.Fatal("expected blank workflow id error")
			}
			if !strings.Contains(err.Error(), "workflow id required") {
				t.Fatalf("expected workflow id required message, got %q", err.Error())
			}
			if warn.Len() != 0 {
				t.Fatalf("expected no warning for blank workflow id, got %q", warn.String())
			}
		})
	}
}

// TestWorkflowsEvalsCreateCmd_NoCobraAutoDeprecation verifies that the real
// workflowsEvalsCreateCmd does NOT trigger cobra's auto-emitted
// "Flag --workflow-id has been deprecated..." message. The custom warning
// in resolveWorkflowIDArg is the single source of truth.
func TestWorkflowsEvalsCreateCmd_NoCobraAutoDeprecation(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{"workflowsEvalsCreateCmd", workflowsEvalsCreateCmd},
		{"workflowsExperimentsCreateCmd", workflowsExperimentsCreateCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flag := tc.cmd.Flags().Lookup("workflow-id")
			if flag == nil {
				t.Fatalf("%s: --workflow-id flag is missing — it must still be registered as a hidden fallback", tc.name)
			}
			if flag.Deprecated != "" {
				t.Fatalf("%s: --workflow-id has flag.Deprecated=%q — cobra will auto-emit a duplicate warning. Use MarkHidden, not MarkDeprecated.", tc.name, flag.Deprecated)
			}
			if !flag.Hidden {
				t.Fatalf("%s: --workflow-id should be hidden from --help", tc.name)
			}
		})
	}

	// End-to-end: parse `--workflow-id wf_xyz` through real cobra plumbing
	// on workflowsEvalsCreateCmd and capture stderr. We can't actually run
	// the RunE (it would dial the API), so we ParseFlags + invoke the
	// resolver directly the same way RunE does, then assert the captured
	// buffer has only the custom warning.
	cmd := workflowsEvalsCreateCmd
	// Suppress cobra usage output during the eval.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	// Reset the flag value so a prior test doesn't leak in.
	if err := cmd.Flags().Set("workflow-id", ""); err != nil {
		t.Fatalf("reset workflow-id: %v", err)
	}
	cmd.Flags().Lookup("workflow-id").Changed = false

	if err := cmd.ParseFlags([]string{"--workflow-id", "wf_xyz"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	// Cobra emits its auto-deprecation message during execution
	// (HasFlags + deprecation check). To pin behavior precisely, set the
	// command's stderr sink to a buffer and call the production helper
	// resolveWorkflowIDArgTo with the same buffer. If MarkDeprecated were
	// still in effect, the cobra auto-message would surface on subsequent
	// command execution. Here we verify the static config instead — the
	// table-driven checks above are sufficient to guarantee no auto-message.
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	got, err := resolveWorkflowIDArgTo(cmd, nil, &buf)
	if err != nil {
		t.Fatalf("resolveWorkflowIDArgTo: %v", err)
	}
	if got != "wf_xyz" {
		t.Fatalf("got %q want wf_xyz", got)
	}
	out := buf.String()
	if strings.Contains(out, "Flag --workflow-id has been deprecated") {
		t.Fatalf("cobra auto-deprecation message leaked through:\n%s", out)
	}
	if !strings.Contains(out, "pass the workflow id as the first positional argument") {
		t.Fatalf("expected the custom deprecation warning, got %q", out)
	}
	// Cleanup so other tests in the package don't inherit the flag value.
	if err := cmd.Flags().Set("workflow-id", ""); err != nil {
		t.Fatalf("cleanup reset workflow-id: %v", err)
	}
	cmd.Flags().Lookup("workflow-id").Changed = false
}

func TestWorkflowsEvalsRunsCreateDoesNotExposeConsensusOverride(t *testing.T) {
	if flag := workflowsEvalsRunsCreateCmd.Flags().Lookup("n-consensus"); flag != nil {
		t.Fatalf("runs create should not expose unsupported --n-consensus flag")
	}
}

// Regression for CLI probing 2026-05: `workflows evals runs list`
// forwarded `--status`, `--from-date`, `--order` straight to the server
// without local validation, surfacing raw HTTP 400/422 envelopes that the
// sibling `workflows runs list` traps client-side. Match the behaviour
// so users get the same clean error shape regardless of which list
// command they call.
func TestWorkflowsEvalsRunsListRejectsInvalidFiltersBeforeRequest(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
		reset     string
	}{
		{name: "invalid status", flag: "status", value: "banana", wantError: "invalid --status", reset: ""},
		{name: "invalid from-date", flag: "from-date", value: "not-a-date", wantError: "YYYY-MM-DD", reset: ""},
		{name: "invalid to-date", flag: "to-date", value: "not-a-date", wantError: "YYYY-MM-DD", reset: ""},
		{name: "invalid order", flag: "order", value: "sideways", wantError: "asc", reset: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for invalid local filter, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			setErr := workflowsEvalsRunsListCmd.Flags().Set(tc.flag, tc.value)
			t.Cleanup(func() {
				_ = workflowsEvalsRunsListCmd.Flags().Set(tc.flag, tc.reset)
			})

			var err error
			if setErr != nil {
				err = setErr
			} else {
				err = workflowsEvalsRunsListCmd.RunE(workflowsEvalsRunsListCmd, nil)
			}
			if err == nil {
				t.Fatalf("expected local validation error for --%s=%s", tc.flag, tc.value)
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

func TestWorkflowsEvalsListCommandsRejectNegativeLimitLocally(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "evals list", cmd: workflowsEvalsListCmd},
		{name: "eval runs list", cmd: workflowsEvalsRunsListCmd},
		{name: "eval run results list", cmd: workflowsEvalsResultsListCmd},
		{name: "experiment runs list", cmd: workflowsExperimentsRunsListCmd},
		{name: "experiment run results list", cmd: workflowsExperimentsResultsListCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("limit", "-1")
			if err == nil {
				t.Fatal("expected local parse error for --limit=-1")
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

func TestWorkflowsEvalsListTableRendersFreshnessColumn(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":   "wfnodeeval_smoke",
				"name": "smoke-test",
				"target": map[string]any{
					"block_id": "blk_extract",
				},
				"freshness": map[string]any{
					"status": "fresh",
				},
				"schema_drift": "none",
				"created_at":   "2026-05-21T12:00:00Z",
			},
		},
	}

	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, workflowEvalColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines (header + 1 row), got %d:\n%s", len(lines), out)
	}
	header := lines[0]
	row := lines[1]
	for _, want := range []string{"TARGET_BLOCK_ID", "FRESHNESS", "SCHEMA_DRIFT"} {
		if !strings.Contains(header, want) {
			t.Fatalf("header missing %s column:\n%s", want, header)
		}
	}
	for _, want := range []string{"blk_extract", "fresh", "none"} {
		if !strings.Contains(row, want) {
			t.Fatalf("row missing %s value:\n%s", want, row)
		}
	}
}

func TestWorkflowsEvalsListCSVUsesDedicatedColumns(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workflows/evals" {
			t.Fatalf("path = %q, want /v1/workflows/evals", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [{
				"id": "wfnodeeval_csv",
				"name": "csv smoke",
				"target": {"type": "block", "block_id": "blk_fn"},
				"freshness": {"status": "fresh"},
				"schema_drift": "none",
				"created_at": "2026-06-19T08:13:28Z"
			}],
			"list_metadata": {}
		}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "csv"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, _ := captureStd(t, func() {
		if err := workflowsEvalsListCmd.RunE(workflowsEvalsListCmd, []string{"wrk_123"}); err != nil {
			t.Fatalf("evals list: %v", err)
		}
	})
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected CSV, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"ID,NAME,TARGET_BLOCK_ID,FRESHNESS,SCHEMA_DRIFT,CREATED_AT", "wfnodeeval_csv,csv smoke,blk_fn,fresh,none"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("CSV missing %q:\n%s", want, stdout)
		}
	}
}

func TestWorkflowsEvalsRunsListFormatsDateFiltersAsRFC3339(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotFrom, gotTo string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workflows/evals/runs" {
			t.Fatalf("path = %q, want /v1/workflows/evals/runs", r.URL.Path)
		}
		gotFrom = r.URL.Query().Get("from_date")
		gotTo = r.URL.Query().Get("to_date")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"list_metadata":{}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{
		"from-date": "2026-06-19",
		"to-date":   "2026-06-19",
	} {
		if err := workflowsEvalsRunsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		flag := flag
		t.Cleanup(func() {
			_ = workflowsEvalsRunsListCmd.Flags().Set(flag, "")
			workflowsEvalsRunsListCmd.Flags().Lookup(flag).Changed = false
		})
	}

	_, _ = captureStd(t, func() {
		if err := workflowsEvalsRunsListCmd.RunE(workflowsEvalsRunsListCmd, []string{"wrk_123"}); err != nil {
			t.Fatalf("evals runs list: %v", err)
		}
	})
	if gotFrom != "2026-06-19T00:00:00Z" {
		t.Fatalf("from_date = %q, want 2026-06-19T00:00:00Z", gotFrom)
	}
	if gotTo != "2026-06-19T23:59:59Z" {
		t.Fatalf("to_date = %q, want 2026-06-19T23:59:59Z", gotTo)
	}
}

func TestWorkflowsEvalsListCommandsRejectOverLimitLocally(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "evals list", cmd: workflowsEvalsListCmd},
		{name: "eval runs list", cmd: workflowsEvalsRunsListCmd},
		{name: "eval run results list", cmd: workflowsEvalsResultsListCmd},
		{name: "experiment runs list", cmd: workflowsExperimentsRunsListCmd},
		{name: "experiment run results list", cmd: workflowsExperimentsResultsListCmd},
	} {
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

func TestWorkflowRunListCommandsHonorExplicitLimit(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	for _, tc := range []struct {
		name     string
		cmd      *cobra.Command
		args     []string
		wantPath string
	}{
		{
			name:     "eval runs list",
			cmd:      workflowsEvalsRunsListCmd,
			wantPath: "/v1/workflows/evals/runs",
		},
		{
			name:     "eval run results list",
			cmd:      workflowsEvalsResultsListCmd,
			args:     []string{"wfevalrun_123"},
			wantPath: "/v1/workflows/evals/results",
		},
		{
			name:     "experiment runs list",
			cmd:      workflowsExperimentsRunsListCmd,
			wantPath: "/v1/workflows/experiments/runs",
		},
		{
			name:     "experiment run results list",
			cmd:      workflowsExperimentsResultsListCmd,
			args:     []string{"exprun_123"},
			wantPath: "/v1/workflows/experiments/results",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				if r.URL.Path != tc.wantPath {
					t.Errorf("path = %q, want %q", r.URL.Path, tc.wantPath)
				}
				if got := r.URL.Query().Get("limit"); got != "7" {
					t.Errorf("limit = %q, want 7", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[],"list_metadata":{"after":null,"before":null}}`))
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := tc.cmd.Flags().Set("limit", "7"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = tc.cmd.Flags().Set("limit", "20")
				tc.cmd.Flags().Lookup("limit").Changed = false
			})

			var err error
			captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, tc.args)
			})
			if err != nil {
				t.Fatalf("RunE returned error: %v", err)
			}
			if got := hits.Load(); got != 1 {
				t.Fatalf("server was hit %d time(s), want 1", got)
			}
		})
	}
}

func TestWorkflowsEvalsResultsListSupportsCursorPagination(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workflows/evals/results" {
			t.Fatalf("path = %q, want /v1/workflows/evals/results", r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"list_metadata":{"after":null,"before":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{
		"limit": "3",
		"after": "wfnodeevalrun_cursor",
	} {
		if err := workflowsEvalsResultsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		flag := flag
		t.Cleanup(func() {
			reset := ""
			if flag == "limit" {
				reset = "20"
			}
			_ = workflowsEvalsResultsListCmd.Flags().Set(flag, reset)
			workflowsEvalsResultsListCmd.Flags().Lookup(flag).Changed = false
		})
	}

	var err error
	captureStd(t, func() {
		err = workflowsEvalsResultsListCmd.RunE(workflowsEvalsResultsListCmd, []string{"wfevalrun_123"})
	})
	if err != nil {
		t.Fatalf("results list: %v", err)
	}
	values, parseErr := url.ParseQuery(gotQuery)
	if parseErr != nil {
		t.Fatalf("parse query %q: %v", gotQuery, parseErr)
	}
	if got := values.Get("run_id"); got != "wfevalrun_123" {
		t.Fatalf("run_id = %q, want wfevalrun_123", got)
	}
	if got := values.Get("limit"); got != "3" {
		t.Fatalf("limit = %q, want 3", got)
	}
	if got := values.Get("after"); got != "wfnodeevalrun_cursor" {
		t.Fatalf("after = %q, want wfnodeevalrun_cursor", got)
	}
}

func TestWorkflowsEvalsResultsListRejectsBeforeAfterTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	for flag, value := range map[string]string{
		"before": "wfnodeevalrun_before",
		"after":  "wfnodeevalrun_after",
	} {
		if err := workflowsEvalsResultsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		flag := flag
		t.Cleanup(func() {
			_ = workflowsEvalsResultsListCmd.Flags().Set(flag, "")
			workflowsEvalsResultsListCmd.Flags().Lookup(flag).Changed = false
		})
	}

	err := workflowsEvalsResultsListCmd.RunE(workflowsEvalsResultsListCmd, []string{"wfevalrun_123"})
	if err == nil {
		t.Fatal("expected before/after mutual exclusion error")
	}
	if !strings.Contains(err.Error(), "--before and --after are mutually exclusive") {
		t.Fatalf("error %q does not contain before/after message", err.Error())
	}
}

func TestWorkflowRunListCommandsUseDocumentedDefaultLimit(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	for _, tc := range []struct {
		name     string
		cmd      *cobra.Command
		args     []string
		wantPath string
	}{
		{
			name:     "eval runs list",
			cmd:      workflowsEvalsRunsListCmd,
			wantPath: "/v1/workflows/evals/runs",
		},
		{
			name:     "eval run results list",
			cmd:      workflowsEvalsResultsListCmd,
			args:     []string{"wfevalrun_123"},
			wantPath: "/v1/workflows/evals/results",
		},
		{
			name:     "experiment runs list",
			cmd:      workflowsExperimentsRunsListCmd,
			wantPath: "/v1/workflows/experiments/runs",
		},
		{
			name:     "experiment run results list",
			cmd:      workflowsExperimentsResultsListCmd,
			args:     []string{"exprun_123"},
			wantPath: "/v1/workflows/experiments/results",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				if r.URL.Path != tc.wantPath {
					t.Errorf("path = %q, want %q", r.URL.Path, tc.wantPath)
				}
				if got := r.URL.Query().Get("limit"); got != "20" {
					t.Errorf("limit = %q, want 20", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[],"list_metadata":{"after":null,"before":null}}`))
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			// Mark the flag as unset so the command falls back to its
			// documented default. The flag's value retains whatever the
			// previous test left behind — that's fine because
			// getIntFlagOrDefault now consults Changed before reading.
			tc.cmd.Flags().Lookup("limit").Changed = false
			t.Cleanup(func() {
				tc.cmd.Flags().Lookup("limit").Changed = false
			})

			var err error
			captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, tc.args)
			})
			if err != nil {
				t.Fatalf("RunE returned error: %v", err)
			}
			if got := hits.Load(); got != 1 {
				t.Fatalf("server was hit %d time(s), want 1", got)
			}
		})
	}
}

func TestWorkflowsEvalsRejectMalformedTargetAndSourceBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	validTarget := filepath.Join(dir, "target.json")
	validSource := filepath.Join(dir, "source.json")
	validAssertion := filepath.Join(dir, "assertion.json")
	invalidTarget := filepath.Join(dir, "invalid-target.json")
	invalidSource := filepath.Join(dir, "invalid-source.json")
	for path, body := range map[string]string{
		validTarget:    `{"type":"block","block_id":"blk_1"}`,
		validSource:    `{"type":"manual","handle_inputs":{}}`,
		validAssertion: `{"target":{"output_handle_id":"output-json-0"},"condition":{"kind":"exists"}}`,
		invalidTarget:  `{}`,
		invalidSource:  `{}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		name      string
		cmd       *cobra.Command
		args      []string
		flags     map[string]string
		wantError string
	}{
		{
			name: "create invalid target",
			cmd:  workflowsEvalsCreateCmd,
			args: []string{"wf_123"},
			flags: map[string]string{
				"name":           "baseline",
				"target-file":    invalidTarget,
				"source-file":    validSource,
				"assertion-file": validAssertion,
			},
			wantError: "--target-file",
		},
		{
			name: "create invalid source",
			cmd:  workflowsEvalsCreateCmd,
			args: []string{"wf_123"},
			flags: map[string]string{
				"name":           "baseline",
				"target-file":    validTarget,
				"source-file":    invalidSource,
				"assertion-file": validAssertion,
			},
			wantError: "--source-file",
		},
		{
			name:      "update invalid source",
			cmd:       workflowsEvalsUpdateCmd,
			args:      []string{"wf_123", "test_123"},
			flags:     map[string]string{"source-file": invalidSource},
			wantError: "--source-file",
		},
		{
			name:      "runs create invalid target",
			cmd:       workflowsEvalsRunsCreateCmd,
			args:      []string{"wf_123"},
			flags:     map[string]string{"target-file": invalidTarget},
			wantError: "--target-file",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			for name, value := range tc.flags {
				if err := tc.cmd.Flags().Set(name, value); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = tc.cmd.Flags().Set(name, "") })
			}

			var err error
			_, stderr := captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, tc.args)
			})
			if err == nil {
				t.Fatal("expected malformed local file error")
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want 0", got-before)
			}
		})
	}
}

func TestWorkflowsExperimentsConsensusFlagsMatchBackendContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "create", cmd: workflowsExperimentsCreateCmd},
		{name: "update", cmd: workflowsExperimentsUpdateCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("n-consensus", "2")
			if err == nil {
				t.Fatal("expected local parse error for --n-consensus=2")
			}
			if !strings.Contains(err.Error(), "3, 5, or 7") {
				t.Fatalf("error %q does not mention allowed consensus counts", err.Error())
			}
			resetConsensusFlag(t, tc.cmd)
		})
	}
}

func TestWorkflowsExperimentsUpdateRejectsExplicitZeroConsensus(t *testing.T) {
	// Explicit --n-consensus=0 must fail at flag-parsing time with the same
	// error users see in --help (no "0" listed as a valid value), so scripts
	// that pass 0 don't reach the server with an ambiguous payload.
	err := workflowsExperimentsUpdateCmd.Flags().Set("n-consensus", "0")
	if err == nil {
		t.Fatal("expected --n-consensus=0 to be rejected by Set")
	}
	if !strings.Contains(err.Error(), "3, 5, or 7") {
		t.Fatalf("error %q does not match --help text", err.Error())
	}
	if strings.Contains(err.Error(), "0, 3, 5, or 7") {
		t.Fatalf("error %q still lists 0 as valid", err.Error())
	}
}

// resetConsensusFlag clears the --n-consensus flag back to its unset state.
// The flag rejects "0" at parse time, so a Set("0") reset is no longer valid;
// reach into the flag's value directly.
func resetConsensusFlag(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	flag := cmd.Flags().Lookup("n-consensus")
	if flag == nil {
		t.Fatalf("cmd %q has no --n-consensus flag", cmd.Name())
	}
	if v, ok := flag.Value.(*consensusFlagValue); ok {
		v.value = ""
	}
	flag.Changed = false
}

func TestWorkflowsExperimentsCreateRejectsInvalidDocumentInputsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	emptyCaptures := filepath.Join(dir, "empty-captures.json")
	missingRunID := filepath.Join(dir, "missing-run-id.json")
	missingHandleInputs := filepath.Join(dir, "missing-handle-inputs.json")
	for path, body := range map[string]string{
		emptyCaptures:       `[]`,
		missingRunID:        `[{"step_id":"step_123"}]`,
		missingHandleInputs: `[{"provenance":{"run_id":"run_123"}}]`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		name      string
		flags     map[string]string
		wantError string
	}{
		{name: "no documents", flags: nil, wantError: "at least one document"},
		{name: "empty captures", flags: map[string]string{"captures-file": emptyCaptures}, wantError: "at least one document"},
		{name: "missing capture run id", flags: map[string]string{"captures-file": missingRunID}, wantError: "run_id is required"},
		{name: "missing explicit handle inputs", flags: map[string]string{"documents-file": missingHandleInputs}, wantError: "handle_inputs is required"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			if err := workflowsExperimentsCreateCmd.Flags().Set("block-id", "blk_123"); err != nil {
				t.Fatal(err)
			}
			if err := workflowsExperimentsCreateCmd.Flags().Set("name", "experiment"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = workflowsExperimentsCreateCmd.Flags().Set("block-id", "")
				_ = workflowsExperimentsCreateCmd.Flags().Set("name", "")
			})
			for name, value := range tc.flags {
				if err := workflowsExperimentsCreateCmd.Flags().Set(name, value); err != nil {
					t.Fatal(err)
				}
				// Reset the value AND the Changed bit. Cobra's Set("")
				// keeps Changed=true, which would otherwise leak into the
				// next subtest and trip the captures-vs-documents mutual
				// exclusion check.
				flagName := name
				t.Cleanup(func() {
					_ = workflowsExperimentsCreateCmd.Flags().Set(flagName, "")
					if f := workflowsExperimentsCreateCmd.Flags().Lookup(flagName); f != nil {
						f.Changed = false
					}
				})
			}

			var err error
			_, stderr := captureStd(t, func() {
				err = workflowsExperimentsCreateCmd.RunE(workflowsExperimentsCreateCmd, []string{"wf_123"})
			})
			if err == nil {
				t.Fatal("expected invalid experiment document input error")
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want 0", got-before)
			}
		})
	}
}

// TestWorkflowsExperimentsCreateRejectsBothSourceFlagsTogether pins the
// mutual-exclusion contract between --captures-file and --documents-file.
// The help text describes them as alternatives ("Provide the input
// documents in one of two ways"), so passing both must fail client-side
// before any file I/O or network call.
func TestWorkflowsExperimentsCreateRejectsBothSourceFlagsTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	capturesFile := filepath.Join(dir, "captures.json")
	documentsFile := filepath.Join(dir, "documents.json")
	if err := os.WriteFile(capturesFile, []byte(`[{"run_id":"run_123","step_id":"step_123"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(documentsFile, []byte(`[{"handle_inputs":{"foo":"bar"}}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := workflowsExperimentsCreateCmd.Flags().Set("block-id", "blk_123"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsExperimentsCreateCmd.Flags().Set("name", "experiment"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsExperimentsCreateCmd.Flags().Set("captures-file", capturesFile); err != nil {
		t.Fatal(err)
	}
	if err := workflowsExperimentsCreateCmd.Flags().Set("documents-file", documentsFile); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsExperimentsCreateCmd.Flags().Set("block-id", "")
		_ = workflowsExperimentsCreateCmd.Flags().Set("name", "")
		_ = workflowsExperimentsCreateCmd.Flags().Set("captures-file", "")
		_ = workflowsExperimentsCreateCmd.Flags().Set("documents-file", "")
		// Cobra's Set("") keeps Changed=true; clear it so neighbour tests
		// don't see stale "changed" state.
		for _, name := range []string{"captures-file", "documents-file"} {
			if f := workflowsExperimentsCreateCmd.Flags().Lookup(name); f != nil {
				f.Changed = false
			}
		}
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsExperimentsCreateCmd.RunE(workflowsExperimentsCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected mutual-exclusion error for --captures-file and --documents-file")
	}
	if !strings.Contains(stderr, "--captures-file") || !strings.Contains(stderr, "--documents-file") {
		t.Fatalf("stderr %q does not mention both flag names", stderr)
	}
	if !strings.Contains(stderr, "cannot be used together") {
		t.Fatalf("stderr %q does not match expected mutual-exclusion wording", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestWorkflowsExperimentsCreateReadsDocumentFilesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-captures.json")
	if err := workflowsExperimentsCreateCmd.Flags().Set("block-id", "blk_123"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsExperimentsCreateCmd.Flags().Set("name", "experiment"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsExperimentsCreateCmd.Flags().Set("captures-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsExperimentsCreateCmd.Flags().Set("block-id", "")
		_ = workflowsExperimentsCreateCmd.Flags().Set("name", "")
		_ = workflowsExperimentsCreateCmd.Flags().Set("captures-file", "")
		// Cobra's Set("") keeps Changed=true; clear it so the captures
		// vs. documents mutual-exclusion check in neighbour tests does
		// not see stale "changed" state.
		if f := workflowsExperimentsCreateCmd.Flags().Lookup("captures-file"); f != nil {
			f.Changed = false
		}
	})

	err := workflowsExperimentsCreateCmd.RunE(workflowsExperimentsCreateCmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected missing captures file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "--captures-file") || !strings.Contains(err.Error(), "missing-captures.json") {
		t.Fatalf("error should mention missing captures file, got %q", err.Error())
	}
}

func TestWorkflowsExperimentsUpdateReadsDocumentFilesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-documents.json")
	if err := workflowsExperimentsUpdateCmd.Flags().Set("documents-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsExperimentsUpdateCmd.Flags().Set("documents-file", "")
		if f := workflowsExperimentsUpdateCmd.Flags().Lookup("documents-file"); f != nil {
			f.Changed = false
		}
	})

	err := workflowsExperimentsUpdateCmd.RunE(workflowsExperimentsUpdateCmd, []string{"exp_123"})
	if err == nil {
		t.Fatal("expected missing documents file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "--documents-file") || !strings.Contains(err.Error(), "missing-documents.json") {
		t.Fatalf("error should mention missing documents file, got %q", err.Error())
	}
}

func TestWorkflowsExperimentsRejectBlankNameBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	docsFile := filepath.Join(t.TempDir(), "documents.json")
	if err := os.WriteFile(docsFile, []byte(`[{"handle_inputs":{}}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		cmd   *cobra.Command
		args  []string
		setup func(t *testing.T)
	}{
		{
			name: "create",
			cmd:  workflowsExperimentsCreateCmd,
			args: []string{"wf_123"},
			setup: func(t *testing.T) {
				t.Helper()
				if err := workflowsExperimentsCreateCmd.Flags().Set("block-id", "blk_123"); err != nil {
					t.Fatal(err)
				}
				if err := workflowsExperimentsCreateCmd.Flags().Set("documents-file", docsFile); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = workflowsExperimentsCreateCmd.Flags().Set("block-id", "")
					_ = workflowsExperimentsCreateCmd.Flags().Set("documents-file", "")
				})
			},
		},
		{
			name:  "update",
			cmd:   workflowsExperimentsUpdateCmd,
			args:  []string{"wf_123", "exp_123"},
			setup: func(t *testing.T) { t.Helper() },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			tc.setup(t)
			if err := tc.cmd.Flags().Set("name", "   "); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = tc.cmd.Flags().Set("name", "") })

			var err error
			_, stderr := captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, tc.args)
			})
			if err == nil {
				t.Fatal("expected blank name error")
			}
			if !strings.Contains(stderr, "experiment name is required") {
				t.Fatalf("stderr %q does not mention blank experiment name", stderr)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want 0", got-before)
			}
		})
	}
}

// TestWorkflowsEvalsRejectBlankNameBeforeRequest mirrors the equivalent
// experiments-side guard. `workflows evals create` previously accepted a
// whitespace-only --name and stored it as-is; `workflows evals update`
// did the same. Both should fail locally with the same wording the
// sibling workflow / experiment commands use, and must NOT reach the
// server.
func TestWorkflowsEvalsRejectBlankNameBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	sourcePath := filepath.Join(dir, "source.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	for path, body := range map[string]string{
		targetPath:    `{"type":"block","block_id":"blk_1"}`,
		sourcePath:    `{"type":"manual","handle_inputs":{}}`,
		assertionPath: `{"target":{"output_handle_id":"output-json-0"},"condition":{"kind":"exists"}}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		name  string
		cmd   *cobra.Command
		args  []string
		setup func(t *testing.T)
	}{
		{
			name: "create",
			cmd:  workflowsEvalsCreateCmd,
			args: []string{"wf_123"},
			setup: func(t *testing.T) {
				t.Helper()
				if err := workflowsEvalsCreateCmd.Flags().Set("target-file", targetPath); err != nil {
					t.Fatal(err)
				}
				if err := workflowsEvalsCreateCmd.Flags().Set("source-file", sourcePath); err != nil {
					t.Fatal(err)
				}
				if err := workflowsEvalsCreateCmd.Flags().Set("assertion-file", assertionPath); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = workflowsEvalsCreateCmd.Flags().Set("target-file", "")
					_ = workflowsEvalsCreateCmd.Flags().Set("source-file", "")
					_ = workflowsEvalsCreateCmd.Flags().Set("assertion-file", "")
				})
			},
		},
		{
			name:  "update",
			cmd:   workflowsEvalsUpdateCmd,
			args:  []string{"tst_123"},
			setup: func(t *testing.T) { t.Helper() },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			tc.setup(t)
			if err := tc.cmd.Flags().Set("name", "   "); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = tc.cmd.Flags().Set("name", "")
				tc.cmd.Flags().Lookup("name").Changed = false
			})

			var err error
			_, stderr := captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, tc.args)
			})
			if err == nil {
				t.Fatal("expected blank eval name error")
			}
			if !strings.Contains(stderr, "eval name") || !strings.Contains(stderr, "must not be blank") {
				t.Fatalf("stderr %q does not mention blank eval name", stderr)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want 0", got-before)
			}
		})
	}
}

// TestWorkflowsEvalsUpdateAllowsAbsentNameFlag confirms that an update
// invocation that doesn't touch --name still works (the blank-name guard
// must only run when --name is explicitly passed). Without this nuance,
// `update --assertion-file ...` would erroneously error.
func TestWorkflowsEvalsUpdateAllowsAbsentNameFlag(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"tst_123","name":"unchanged"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	assertionPath := filepath.Join(dir, "assertion.json")
	if err := os.WriteFile(assertionPath, []byte(`{"target":{"output_handle_id":"output-json-0"},"condition":{"kind":"exists"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := workflowsEvalsUpdateCmd.Flags().Set("assertion-file", assertionPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsEvalsUpdateCmd.Flags().Set("assertion-file", "")
		workflowsEvalsUpdateCmd.Flags().Lookup("name").Changed = false
	})

	var err error
	captureStd(t, func() {
		err = workflowsEvalsUpdateCmd.RunE(workflowsEvalsUpdateCmd, []string{"tst_123"})
	})
	if err != nil {
		t.Fatalf("update without --name should succeed, got %v", err)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("server was hit %d time(s), want 1", got)
	}
}

// TestWorkflowsEvalsUpdateInlineAssertion pins that `evals update` accepts the
// same inline assertion flags as `evals create` (--path/--equals), assembling the
// equals-assertion body without a JSON file — closing the create/update ergonomics
// gap where update previously required hand-authored --assertion-file JSON.
func TestWorkflowsEvalsUpdateInlineAssertion(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Inline update fetches the existing test first to merge overrides.
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"id":"wfnodeeval_x","workflow_id":"wrk_1","target":{"type":"block","block_id":"blk_1"},"source":{"type":"run_step","run_id":"run_1"},"assertion":{"target":{"output_handle_id":"output-json-0","path":"old_path"},"condition":{"kind":"equals","expected":"Old"}}}`))
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"id":"wfnodeeval_x"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{"path": "vendor_name", "equals": "Acme"} {
		if err := workflowsEvalsUpdateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for _, f := range []string{"path", "equals", "output-handle-id", "run-id", "step-id", "assertion-file", "source-file", "name"} {
			_ = workflowsEvalsUpdateCmd.Flags().Set(f, "")
			workflowsEvalsUpdateCmd.Flags().Lookup(f).Changed = false
		}
	})

	var err error
	captureStd(t, func() {
		err = workflowsEvalsUpdateCmd.RunE(workflowsEvalsUpdateCmd, []string{"wfnodeeval_x"})
	})
	if err != nil {
		t.Fatalf("update (inline assertion): %v", err)
	}
	assertion, ok := gotBody["assertion"].(map[string]any)
	if !ok {
		t.Fatalf("assertion missing from PATCH body: %#v", gotBody)
	}
	at, ok := assertion["target"].(map[string]any)
	if !ok || at["output_handle_id"] != "output-json-0" || at["path"] != "vendor_name" {
		t.Fatalf("assertion.target = %#v, want {output_handle_id:output-json-0, path:vendor_name}", assertion["target"])
	}
	condition, ok := assertion["condition"].(map[string]any)
	if !ok || condition["kind"] != "equals" || condition["expected"] != "Acme" {
		t.Fatalf("assertion.condition = %#v, want {kind:equals, expected:Acme}", assertion["condition"])
	}
}

func TestWorkflowsExperimentsCreateHelpDoesNotMentionRunBatch(t *testing.T) {
	if strings.Contains(workflowsExperimentsCreateCmd.Long, "run-batch") {
		t.Fatalf("experiments create help should not mention run-batch, got:\n%s", workflowsExperimentsCreateCmd.Long)
	}
}

func TestWorkflowsExperimentsRunsHelpUsesCreateRunTerminology(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{name: "experiments long", text: workflowsExperimentsCmd.Long},
		{name: "experiments create long", text: workflowsExperimentsCreateCmd.Long},
		{name: "runs long", text: workflowsExperimentsRunsCmd.Long},
		{name: "runs example", text: workflowsExperimentsRunsCmd.Example},
		{name: "runs create short", text: workflowsExperimentsRunsCreateCmd.Short},
		{name: "runs create long", text: workflowsExperimentsRunsCreateCmd.Long},
		{name: "runs create example", text: workflowsExperimentsRunsCreateCmd.Example},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Contains(strings.ToLower(tc.text), "trigger") {
				t.Fatalf("experiment run help should use create-run terminology, got:\n%s", tc.text)
			}
		})
	}
}

func TestWorkflowsEvalsRunsHelpSeparatesRunsAndResults(t *testing.T) {
	if strings.Contains(workflowsEvalsRunsCmd.Example, "Recent result rows") {
		t.Fatalf("eval runs list example should describe parent runs, got:\n%s", workflowsEvalsRunsCmd.Example)
	}
	for _, stale := range []string{"evals execute", "job_id", "batch_id"} {
		if strings.Contains(workflowsEvalsRunsCmd.Long+workflowsEvalsRunsCmd.Example, stale) {
			t.Fatalf("eval runs help should not expose %q, got:\n%s\n%s", stale, workflowsEvalsRunsCmd.Long, workflowsEvalsRunsCmd.Example)
		}
	}
}

func TestWorkflowsExperimentsUnsupportedRunOverrideFlagsAreNotRegistered(t *testing.T) {
	if flag := workflowsExperimentsRunsCreateCmd.Flags().Lookup("n-consensus"); flag != nil {
		t.Fatalf("runs create should not expose unsupported --n-consensus flag")
	}
	if flag := workflowsExperimentsRunsCreateCmd.Flags().Lookup("retry-failed-only"); flag != nil {
		t.Fatalf("runs create should not expose unsupported --retry-failed-only flag")
	}
}

func TestWorkflowsExperimentsMetricsRejectsInvalidViewBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	workflowsExperimentsMetricsGetCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsExperimentsMetricsGetCmd.SetContext(context.Background()) })
	if err := workflowsExperimentsMetricsGetCmd.Flags().Set("view", "banana"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsExperimentsMetricsGetCmd.Flags().Set("view", "summary") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsExperimentsMetricsGetCmd.RunE(workflowsExperimentsMetricsGetCmd, []string{"exprun_123"})
	})
	if err == nil {
		t.Fatal("expected invalid view error")
	}
	if !strings.Contains(stderr, "invalid --view") {
		t.Fatalf("stderr %q does not mention invalid --view", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

func TestWorkflowsExperimentsMetricsViewsMatchBackendContract(t *testing.T) {
	for _, view := range []string{"", "summary", "by_document", "by_target", "votes"} {
		if err := validateExperimentMetricsView(view); err != nil {
			t.Fatalf("view %q should be valid, got %v", view, err)
		}
	}
	for _, view := range []string{"per_run", "per_document", "per_field"} {
		err := validateExperimentMetricsView(view)
		if err == nil {
			t.Fatalf("view %q should be rejected", view)
		}
		if !strings.Contains(err.Error(), "by_document") || !strings.Contains(err.Error(), "votes") {
			t.Fatalf("error %q does not mention backend view names", err.Error())
		}
	}
}

func TestWorkflowsEvalsCreateRejectsAssertionMissingTargetBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	sourcePath := filepath.Join(dir, "source.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	if err := os.WriteFile(targetPath, []byte(`{"type":"block","block_id":"block_123"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte(`{"type":"manual","handle_inputs":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assertionPath, []byte(`{"condition":{"kind":"exists"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	for flag, path := range map[string]string{
		"name":           "missing target",
		"target-file":    targetPath,
		"source-file":    sourcePath,
		"assertion-file": assertionPath,
	} {
		if err := workflowsEvalsCreateCmd.Flags().Set(flag, path); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = workflowsEvalsCreateCmd.Flags().Set(flag, "") })
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsEvalsCreateCmd.RunE(workflowsEvalsCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected assertion validation error")
	}
	if !strings.Contains(stderr, "--assertion-file: assertion.target is required") {
		t.Fatalf("stderr %q does not mention assertion target", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

func TestWorkflowsEvalsCreateReadsLocalFilesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	missingPath := filepath.Join(dir, "missing-source.json")
	if err := os.WriteFile(targetPath, []byte(`{"type":"block","block_id":"block_123"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assertionPath, []byte(`{"target":{"output_handle_id":"output-json-0"},"condition":{"kind":"exists"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for flag, value := range map[string]string{
		"name":           "baseline",
		"target-file":    targetPath,
		"source-file":    missingPath,
		"assertion-file": assertionPath,
	} {
		if err := workflowsEvalsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = workflowsEvalsCreateCmd.Flags().Set(flag, "") })
	}

	err := workflowsEvalsCreateCmd.RunE(workflowsEvalsCreateCmd, []string{"wf_123"})
	if err == nil {
		t.Fatal("expected missing source file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "--source-file") || !strings.Contains(err.Error(), "missing-source.json") {
		t.Fatalf("error should mention missing source file, got %q", err.Error())
	}
}

func TestWorkflowsEvalsUpdateReadsLocalFilesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-assertion.json")
	if err := workflowsEvalsUpdateCmd.Flags().Set("assertion-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsEvalsUpdateCmd.Flags().Set("assertion-file", "") })

	err := workflowsEvalsUpdateCmd.RunE(workflowsEvalsUpdateCmd, []string{"test_123"})
	if err == nil {
		t.Fatal("expected missing assertion file error")
	}
	if strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("local file validation should run before credentials, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "--assertion-file") || !strings.Contains(err.Error(), "missing-assertion.json") {
		t.Fatalf("error should mention missing assertion file, got %q", err.Error())
	}
}

// Issue 7: “workflows evals create“ help previously described
// --source-file as "the input. Usually a reference to a run/step that
// supplied known-good input." with no schema — the same level of
// detail --target-file already gets. The fix gives --source-file the
// same concrete-example treatment so users don't have to discover the
// shape by trial-and-error 400s.
//
// The server-side discriminated union (services/v1/workflows/evals/
// models.py:343-354) is:
//
//   - {"type": "manual", "handle_inputs": {"input-json-0": {"type": "json", "data": {...}}}}
//   - {"type": "run_step", "run_id": "run_..."}
//
// so the help must show both shapes.
func TestWorkflowsEvalsCreateHelpDocumentsSourceFileShape(t *testing.T) {
	long := workflowsEvalsCreateCmd.Long
	if !strings.Contains(long, `"type": "manual"`) {
		t.Fatalf(`workflows evals create help must show the manual source shape ({"type": "manual", ...}), got:\n%s`, long)
	}
	if !strings.Contains(long, `"input-json-0"`) || !strings.Contains(long, `"type": "json"`) || !strings.Contains(long, `"data"`) {
		t.Fatalf("workflows evals create help must show typed manual handle inputs, got:\n%s", long)
	}
	if !strings.Contains(long, `"type": "run_step"`) {
		t.Fatalf(`workflows evals create help must show the run_step source shape ({"type": "run_step", "run_id": "run_..."}), got:\n%s`, long)
	}
	if !strings.Contains(long, `"run_id"`) {
		t.Fatalf("workflows evals create help must document run_id as the run_step disambiguator, got:\n%s", long)
	}
}

func TestWorkflowsEvalsCreateRejectsUntypedManualHandleInputsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	sourcePath := filepath.Join(dir, "source.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	for path, body := range map[string]string{
		targetPath:    `{"type":"block","block_id":"block_123"}`,
		sourcePath:    `{"type":"manual","handle_inputs":{"input-json-0":{"customer":"ManualCo","amount":40}}}`,
		assertionPath: `{"target":{"output_handle_id":"output-json-0","path":"message"},"condition":{"kind":"equals","expected":"ManualCo"}}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	for flag, value := range map[string]string{
		"name":           "manual source",
		"target-file":    targetPath,
		"source-file":    sourcePath,
		"assertion-file": assertionPath,
	} {
		if err := workflowsEvalsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { resetEvalsCreateInlineFlags(t) })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsEvalsCreateCmd.RunE(workflowsEvalsCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected manual handle input validation error")
	}
	if !strings.Contains(stderr, "--source-file: source.handle_inputs.input-json-0.type is required") {
		t.Fatalf("stderr %q does not mention typed manual handle input requirement", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

// TestWorkflowsEvalsCreatePreservesRunStepAndExpectedFields is the
// regression guard for the silent field-dropping bug: `workflows evals
// create` decoded the source into the manual-only SDK struct (dropping
// run_step's run_id/step_id) and the assertion into a Condition type that
// collapses to the kind-only ExistCondition (dropping `expected`). The
// server then 422'd on `source.run_step.run_id` and
// `assertion.condition.equals.expected`. The fix splices the validated
// raw source/assertion into the request body, so the wire payload must
// carry both fields.
func TestWorkflowsEvalsCreatePreservesRunStepAndExpectedFields(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.json")
	sourcePath := filepath.Join(dir, "source.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	for path, body := range map[string]string{
		targetPath:    `{"type":"block","block_id":"block_jT"}`,
		sourcePath:    `{"type":"run_step","run_id":"run_f_nrFkuN2Uvh","step_id":"step_abc"}`,
		assertionPath: `{"target":{"output_handle_id":"output-json-0","path":"bank_name"},"condition":{"kind":"equals","expected":"Commerce Bank"}}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/evals" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"wfnodeeval_123"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{
		"name":           "Invoice baseline",
		"target-file":    targetPath,
		"source-file":    sourcePath,
		"assertion-file": assertionPath,
	} {
		if err := workflowsEvalsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
		flagName := flag
		t.Cleanup(func() {
			_ = workflowsEvalsCreateCmd.Flags().Set(flagName, "")
			if f := workflowsEvalsCreateCmd.Flags().Lookup(flagName); f != nil {
				f.Changed = false
			}
		})
	}

	var err error
	captureStd(t, func() {
		err = workflowsEvalsCreateCmd.RunE(workflowsEvalsCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("evals create: %v", err)
	}

	source, ok := gotBody["source"].(map[string]any)
	if !ok {
		t.Fatalf("source missing or wrong type in body: %#v", gotBody["source"])
	}
	if source["type"] != "run_step" {
		t.Fatalf("source.type = %v, want run_step", source["type"])
	}
	if source["run_id"] != "run_f_nrFkuN2Uvh" {
		t.Fatalf("source.run_id = %v, want run_f_nrFkuN2Uvh (was dropped before fix)", source["run_id"])
	}
	if source["step_id"] != "step_abc" {
		t.Fatalf("source.step_id = %v, want step_abc (was dropped before fix)", source["step_id"])
	}

	assertion, ok := gotBody["assertion"].(map[string]any)
	if !ok {
		t.Fatalf("assertion missing or wrong type in body: %#v", gotBody["assertion"])
	}
	condition, ok := assertion["condition"].(map[string]any)
	if !ok {
		t.Fatalf("assertion.condition missing or wrong type: %#v", assertion["condition"])
	}
	if condition["kind"] != "equals" {
		t.Fatalf("condition.kind = %v, want equals", condition["kind"])
	}
	if condition["expected"] != "Commerce Bank" {
		t.Fatalf("condition.expected = %v, want \"Commerce Bank\" (was dropped before fix)", condition["expected"])
	}
}

// TestWorkflowsEvalsUpdatePreservesRunStepAndExpectedFields mirrors the
// create regression for the update path, which shared the same lossy
// decode into manual-only / kind-only SDK structs.
func TestWorkflowsEvalsUpdatePreservesRunStepAndExpectedFields(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.json")
	assertionPath := filepath.Join(dir, "assertion.json")
	if err := os.WriteFile(sourcePath, []byte(`{"type":"run_step","run_id":"run_xyz"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assertionPath, []byte(`{"target":{"output_handle_id":"output-json-0"},"condition":{"kind":"equals","expected":42}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"wfnodeeval_123"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsEvalsUpdateCmd.Flags().Set("source-file", sourcePath); err != nil {
		t.Fatal(err)
	}
	if err := workflowsEvalsUpdateCmd.Flags().Set("assertion-file", assertionPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsEvalsUpdateCmd.Flags().Set("source-file", "")
		_ = workflowsEvalsUpdateCmd.Flags().Set("assertion-file", "")
		for _, name := range []string{"source-file", "assertion-file", "name"} {
			if f := workflowsEvalsUpdateCmd.Flags().Lookup(name); f != nil {
				f.Changed = false
			}
		}
	})

	var err error
	captureStd(t, func() {
		err = workflowsEvalsUpdateCmd.RunE(workflowsEvalsUpdateCmd, []string{"wfnodeeval_123"})
	})
	if err != nil {
		t.Fatalf("evals update: %v", err)
	}

	source, ok := gotBody["source"].(map[string]any)
	if !ok {
		t.Fatalf("source missing or wrong type: %#v", gotBody["source"])
	}
	if source["run_id"] != "run_xyz" {
		t.Fatalf("source.run_id = %v, want run_xyz (was dropped before fix)", source["run_id"])
	}
	assertion, ok := gotBody["assertion"].(map[string]any)
	if !ok {
		t.Fatalf("assertion missing or wrong type: %#v", gotBody["assertion"])
	}
	condition, ok := assertion["condition"].(map[string]any)
	if !ok {
		t.Fatalf("assertion.condition missing: %#v", assertion["condition"])
	}
	if condition["expected"] != float64(42) {
		t.Fatalf("condition.expected = %v, want 42 (was dropped before fix)", condition["expected"])
	}
}

// resetEvalsCreateInlineFlags clears the inline create flags (value AND the
// Changed bit) so a mutating subtest doesn't leak state into its neighbours —
// the inline assertion builder keys off Changed, so a stale Changed=true would
// trip the next test.
func resetEvalsCreateInlineFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"name", "target-file", "source-file", "assertion-file",
		"block-id", "run-id", "step-id", "output-handle-id", "path", "equals",
	} {
		_ = workflowsEvalsCreateCmd.Flags().Set(name, "")
		if f := workflowsEvalsCreateCmd.Flags().Lookup(name); f != nil {
			f.Changed = false
		}
	}
}

// TestWorkflowsEvalsCreateInlineBuildsThreeComponents pins the inline form:
// `--block-id / --run-id [--step-id] / --path --equals` must assemble the same
// target / source / assertion request body the three JSON files would, without
// any file on disk. `--equals 300000` must serialize as the number 300000 and
// --output-handle-id must default to output-json-0.
func TestWorkflowsEvalsCreateInlineBuildsThreeComponents(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/evals" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"wfnodeeval_inline"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{
		"name":     "Invoice 17 baseline",
		"block-id": "extract_1",
		"run-id":   "run_xxx",
		"path":     "net_amount_payable_usd",
		"equals":   "300000",
	} {
		if err := workflowsEvalsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { resetEvalsCreateInlineFlags(t) })

	var err error
	captureStd(t, func() {
		err = workflowsEvalsCreateCmd.RunE(workflowsEvalsCreateCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("evals create (inline): %v", err)
	}

	target, ok := gotBody["target"].(map[string]any)
	if !ok || target["type"] != "block" || target["block_id"] != "extract_1" {
		t.Fatalf("target = %#v, want {type:block, block_id:extract_1}", gotBody["target"])
	}
	source, ok := gotBody["source"].(map[string]any)
	if !ok || source["type"] != "run_step" || source["run_id"] != "run_xxx" {
		t.Fatalf("source = %#v, want {type:run_step, run_id:run_xxx}", gotBody["source"])
	}
	if _, present := source["step_id"]; present {
		t.Fatalf("source.step_id should be absent when --step-id is unset, got %#v", source["step_id"])
	}
	assertion, ok := gotBody["assertion"].(map[string]any)
	if !ok {
		t.Fatalf("assertion missing: %#v", gotBody["assertion"])
	}
	at, ok := assertion["target"].(map[string]any)
	if !ok || at["output_handle_id"] != "output-json-0" || at["path"] != "net_amount_payable_usd" {
		t.Fatalf("assertion.target = %#v, want {output_handle_id:output-json-0, path:net_amount_payable_usd}", assertion["target"])
	}
	condition, ok := assertion["condition"].(map[string]any)
	if !ok || condition["kind"] != "equals" {
		t.Fatalf("assertion.condition = %#v, want kind=equals", assertion["condition"])
	}
	if condition["expected"] != float64(300000) {
		t.Fatalf("assertion.condition.expected = %#v (%T), want numeric 300000", condition["expected"], condition["expected"])
	}
}

// TestWorkflowsEvalsCreateInlineAndFileAreMutuallyExclusive pins that mixing a
// file form and the inline form for the SAME component fails locally, before
// any network call.
func TestWorkflowsEvalsCreateInlineAndFileAreMutuallyExclusive(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	targetPath := filepath.Join(t.TempDir(), "target.json")
	if err := os.WriteFile(targetPath, []byte(`{"type":"block","block_id":"extract_1"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	for flag, value := range map[string]string{
		"name":        "baseline",
		"target-file": targetPath,
		"block-id":    "extract_1",
	} {
		if err := workflowsEvalsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { resetEvalsCreateInlineFlags(t) })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsEvalsCreateCmd.RunE(workflowsEvalsCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected mutual-exclusion error for --target-file and --block-id")
	}
	if !strings.Contains(stderr, "--target-file") || !strings.Contains(stderr, "--block-id") || !strings.Contains(stderr, "mutually exclusive") {
		t.Fatalf("stderr %q should name both forms and say mutually exclusive", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

// TestWorkflowsEvalsCreateInlineAssertionRequiresEquals pins that an inline
// assertion triggered by --path (or --output-handle-id) but missing --equals
// fails locally with a message pointing at --equals.
func TestWorkflowsEvalsCreateInlineAssertionRequiresEquals(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{
		"name":     "baseline",
		"block-id": "extract_1",
		"run-id":   "run_xxx",
		"path":     "net_amount_payable_usd",
	} {
		if err := workflowsEvalsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { resetEvalsCreateInlineFlags(t) })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsEvalsCreateCmd.RunE(workflowsEvalsCreateCmd, []string{"wf_123"})
	})
	if err == nil {
		t.Fatal("expected error when inline assertion omits --equals")
	}
	if !strings.Contains(stderr, "--equals") {
		t.Fatalf("stderr %q should point at --equals", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

// resetTestsRunsCreateWaitFlags restores the --wait/--poll/--timeout flags on
// the singleton runs-create command back to defaults.
func resetTestsRunsCreateWaitFlags(t *testing.T) {
	t.Helper()
	_ = workflowsEvalsRunsCreateCmd.Flags().Set("wait", "false")
	_ = workflowsEvalsRunsCreateCmd.Flags().Set("poll-interval-ms", "2000")
	_ = workflowsEvalsRunsCreateCmd.Flags().Set("timeout-seconds", "600")
}

// TestWorkflowsEvalsRunsCreateWaitPollsUntilTerminal pins the new `--wait`
// behavior on eval runs: after POSTing the run the CLI polls
// GET .../evals/runs/<id> until the lifecycle reaches a terminal status and
// prints the FINAL run record, not the freshly-queued one. Mirrors the
// equivalent contract on `experiments runs create --wait`.
func TestWorkflowsEvalsRunsCreateWaitPollsUntilTerminal(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posts, gets atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/evals/runs":
			posts.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":        "wfevalrun_wait",
				"lifecycle": map[string]any{"status": "pending"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/evals/runs/wfevalrun_wait":
			status := "running"
			if gets.Add(1) >= 2 {
				status = "completed"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wfevalrun_wait",
				"total_evals": 3,
				"lifecycle":   map[string]any{"status": status},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsEvalsRunsCreateCmd.Flags().Set("wait", "true"); err != nil {
		t.Fatalf("set --wait: %v", err)
	}
	if err := workflowsEvalsRunsCreateCmd.Flags().Set("poll-interval-ms", "1"); err != nil {
		t.Fatalf("set --poll-interval-ms: %v", err)
	}
	t.Cleanup(func() { resetTestsRunsCreateWaitFlags(t) })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsEvalsRunsCreateCmd.RunE(workflowsEvalsRunsCreateCmd, []string{"wf_123"}); err != nil {
			t.Fatalf("evals runs create --wait: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if got := posts.Load(); got != 1 {
		t.Fatalf("expected exactly 1 POST, got %d", got)
	}
	if got := gets.Load(); got < 2 {
		t.Fatalf("expected the CLI to poll at least twice before terminal, got %d GET(s)", got)
	}
	if !strings.Contains(stdout, `"status": "completed"`) {
		t.Fatalf("expected final completed run on stdout, got:\n%s", stdout)
	}
	// The printed record must be the polled terminal one (carries total_evals),
	// not the freshly-queued POST response.
	if !strings.Contains(stdout, `"total_evals": 3`) {
		t.Fatalf("expected the final polled run on stdout, got:\n%s", stdout)
	}
}

// TestWorkflowsEvalsRunsWaitStandalone pins the standalone `evals runs wait
// <run-id>`: it polls GET .../evals/runs/<id> until terminal and prints the
// final run. Mirrors `experiments runs wait`, closing the gap where eval runs
// had `create --wait` but no standalone poller.
func TestWorkflowsEvalsRunsWaitStandalone(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gets atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/evals/runs/wfevalrun_wait" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		status := "running"
		if gets.Add(1) >= 2 {
			status = "completed"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "wfevalrun_wait",
			"lifecycle": map[string]any{"status": status},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsEvalsRunsWaitCmd.Flags().Set("poll-interval-ms", "1"); err != nil {
		t.Fatalf("set --poll-interval-ms: %v", err)
	}
	t.Cleanup(func() {
		_ = workflowsEvalsRunsWaitCmd.Flags().Set("poll-interval-ms", "2000")
		_ = workflowsEvalsRunsWaitCmd.Flags().Set("timeout-seconds", "600")
	})

	stdout, stderr := captureStd(t, func() {
		if err := workflowsEvalsRunsWaitCmd.RunE(workflowsEvalsRunsWaitCmd, []string{"wfevalrun_wait"}); err != nil {
			t.Fatalf("evals runs wait: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if got := gets.Load(); got < 2 {
		t.Fatalf("expected at least 2 polls, got %d", got)
	}
	if !strings.Contains(stdout, `"status": "completed"`) {
		t.Fatalf("expected final completed run on stdout, got:\n%s", stdout)
	}
}

// TestWorkflowsEvalsRunsWaitErrorStatusExitsNonZero pins that a eval run that
// settles in error/cancelled surfaces a non-nil error (non-zero shell exit)
// while still printing the run record for context.
func TestWorkflowsEvalsRunsWaitErrorStatusExitsNonZero(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/evals/runs/wfevalrun_bad" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "wfevalrun_bad",
			"lifecycle": map[string]any{"status": "error"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsEvalsRunsWaitCmd.Flags().Set("poll-interval-ms", "1"); err != nil {
		t.Fatalf("set --poll-interval-ms: %v", err)
	}
	t.Cleanup(func() {
		_ = workflowsEvalsRunsWaitCmd.Flags().Set("poll-interval-ms", "2000")
		_ = workflowsEvalsRunsWaitCmd.Flags().Set("timeout-seconds", "600")
	})

	var runErr error
	stdout, _ := captureStd(t, func() {
		runErr = workflowsEvalsRunsWaitCmd.RunE(workflowsEvalsRunsWaitCmd, []string{"wfevalrun_bad"})
	})
	if runErr == nil {
		t.Fatal("expected non-nil error for a run that ended in error status")
	}
	if !strings.Contains(runErr.Error(), "wfevalrun_bad") || !strings.Contains(runErr.Error(), "error") {
		t.Fatalf("error %q should name the run and its terminal status", runErr.Error())
	}
	if !strings.Contains(stdout, `"status": "error"`) {
		t.Fatalf("expected the failed run record on stdout for context, got:\n%s", stdout)
	}
}

// nonEmptyLines splits s on "\n" and returns the non-empty entries. We use
// this to assert exact line counts on warning output without being tripped
// up by the trailing newline from fmt.Fprintln.
func nonEmptyLines(s string) []string {
	var out []string
	for line := range strings.SplitSeq(s, "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// Regression: `evals update --equals X` (without --path) must NOT drop the
// existing assertion path. A partial inline update should merge onto the
// existing assertion (fetched first), changing only the expected value and
// preserving target.path / target.output_handle_id. Previously the CLI rebuilt
// a fresh assertion from inline flags alone and silently replaced the stored
// one, turning "vendor_name equals X" into "whole handle equals X".
func TestWorkflowsEvalsUpdateEqualsPreservesExistingPath(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var patchAssertion map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/evals/wfnodeeval_x":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wfnodeeval_x",
				"workflow_id": "wrk_1",
				"target":      map[string]any{"type": "block", "block_id": "blk_1"},
				"source":      map[string]any{"type": "run_step", "run_id": "run_1", "step_id": "step_1"},
				"assertion": map[string]any{
					"target":    map[string]any{"output_handle_id": "output-json-0", "path": "vendor_name"},
					"condition": map[string]any{"kind": "equals", "expected": "Old Vendor"},
				},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/v1/workflows/evals/wfnodeeval_x":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if a, ok := body["assertion"].(map[string]any); ok {
				patchAssertion = a
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "wfnodeeval_x", "workflow_id": "wrk_1",
				"target": map[string]any{"type": "block", "block_id": "blk_1"},
				"source": map[string]any{"type": "run_step", "run_id": "run_1"}})
		default:
			http.Error(w, "unexpected "+r.Method+" "+r.URL.Path, http.StatusInternalServerError)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := runRootForTest(t, "workflows", "evals", "update", "wfnodeeval_x", "--equals", "New Vendor"); err != nil {
		t.Fatalf("evals update: %v", err)
	}
	if patchAssertion == nil {
		t.Fatalf("no PATCH assertion captured")
	}
	target, _ := patchAssertion["target"].(map[string]any)
	if target == nil || target["path"] != "vendor_name" {
		t.Fatalf("update dropped the assertion path; target=%v", target)
	}
	cond, _ := patchAssertion["condition"].(map[string]any)
	if cond == nil || cond["expected"] != "New Vendor" {
		t.Fatalf("update did not apply new expected value; condition=%v", cond)
	}
}

// TestValidateWorkflowEvalSourceRejectsUnknownKeys guards the documented
// extra="forbid" contract on the eval source/target value objects. The CLI
// decodes source/target into typed structs, and the server's nested source
// variants are deliberately loose, so a typo'd field (e.g. "step" for
// "step_id") would otherwise be silently dropped end-to-end — the eval would
// pin to an auto-resolved step instead of the one the user named, with no
// error. The create help promises these are "rejected with 'Extra inputs are
// not permitted'", so the local validators must enforce it.
func TestValidateWorkflowEvalSourceRejectsUnknownKeys(t *testing.T) {
	cases := []struct {
		name      string
		source    map[string]any
		wantError string // substring; "" means the source must validate cleanly
	}{
		{
			name:   "run_step valid with optional step_id",
			source: map[string]any{"type": "run_step", "run_id": "run_1", "step_id": "step_1"},
		},
		{
			name:   "run_step valid without step_id",
			source: map[string]any{"type": "run_step", "run_id": "run_1"},
		},
		{
			name:      "run_step typo step instead of step_id",
			source:    map[string]any{"type": "run_step", "run_id": "run_1", "step": "step_1"},
			wantError: "step",
		},
		{
			name:      "run_step unknown key",
			source:    map[string]any{"type": "run_step", "run_id": "run_1", "bogus_extra": 123},
			wantError: "bogus_extra",
		},
		{
			name:   "manual valid",
			source: map[string]any{"type": "manual", "handle_inputs": map[string]any{}},
		},
		{
			name:      "manual unknown key",
			source:    map[string]any{"type": "manual", "handle_inputs": map[string]any{}, "extra": true},
			wantError: "extra",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWorkflowEvalSource(tc.source)
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("expected valid source, got error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error mentioning %q, got nil", tc.wantError)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not mention %q", err.Error(), tc.wantError)
			}
		})
	}
}

func TestValidateWorkflowEvalTargetRejectsUnknownKeys(t *testing.T) {
	if err := validateWorkflowEvalTarget(map[string]any{"type": "block", "block_id": "blk_1"}); err != nil {
		t.Fatalf("expected valid target, got error: %v", err)
	}
	err := validateWorkflowEvalTarget(map[string]any{"type": "block", "block_id": "blk_1", "junk": 1})
	if err == nil {
		t.Fatal("expected error for unknown target key, got nil")
	}
	if !strings.Contains(err.Error(), "junk") {
		t.Fatalf("error %q does not mention the offending key", err.Error())
	}
}

// TestWorkflowEvalRunTerminalErrorFailsOnNonPassingAssertions guards that
// `workflows evals runs create --wait` exits non-zero when a completed run has
// failed or blocked assertions, so CI can gate a build on a detected regression
// (not only on lifecycle error/cancelled/timeout).
func TestWorkflowEvalRunTerminalErrorFailsOnNonPassingAssertions(t *testing.T) {
	completed := func(passed, failed, blocked int) *retab.WorkflowEvalRun {
		return &retab.WorkflowEvalRun{
			ID:        "wfevalrun_x",
			Lifecycle: retab.WorkflowEvalRunStatusFromCompletedWorkflowEvalRun(retab.CompletedWorkflowEvalRun{}),
			Counts: &retab.BlockEvalBatchExecutionCounts{
				Outcome: &retab.BlockEvalOutcomeCounts{Passed: &passed, Failed: &failed, Blocked: &blocked},
			},
		}
	}
	completedWithChildLifecycle := func(errored, cancelled int) *retab.WorkflowEvalRun {
		return &retab.WorkflowEvalRun{
			ID:        "wfevalrun_child_lifecycle",
			Lifecycle: retab.WorkflowEvalRunStatusFromCompletedWorkflowEvalRun(retab.CompletedWorkflowEvalRun{}),
			Counts: &retab.BlockEvalBatchExecutionCounts{
				LifecycleCounts: &retab.BlockEvalLifecycleCounts{Error: &errored, Cancelled: &cancelled},
			},
		}
	}
	cases := []struct {
		name    string
		run     *retab.WorkflowEvalRun
		wantErr bool
	}{
		{name: "all passed", run: completed(8, 0, 0), wantErr: false},
		{name: "some failed", run: completed(7, 1, 0), wantErr: true},
		{name: "some blocked", run: completed(7, 0, 1), wantErr: true},
		{name: "failed and blocked", run: completed(5, 2, 1), wantErr: true},
		{name: "child eval error", run: completedWithChildLifecycle(1, 0), wantErr: true},
		{name: "child eval cancelled", run: completedWithChildLifecycle(0, 1), wantErr: true},
		{
			name: "completed with no counts is success",
			run: &retab.WorkflowEvalRun{
				ID:        "wfevalrun_y",
				Lifecycle: retab.WorkflowEvalRunStatusFromCompletedWorkflowEvalRun(retab.CompletedWorkflowEvalRun{}),
			},
			wantErr: false,
		},
		{
			name: "error lifecycle still fails",
			run: &retab.WorkflowEvalRun{
				ID:        "wfevalrun_z",
				Lifecycle: retab.WorkflowEvalRunStatusFromErrorWorkflowEvalRun(retab.ErrorWorkflowEvalRun{}),
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := workflowEvalRunTerminalError(tc.run)
			if tc.wantErr && err == nil {
				t.Fatalf("expected non-zero exit error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
		})
	}
}
