package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

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

// TestWorkflowsTestsCreateCmd_NoCobraAutoDeprecation verifies that the real
// workflowsTestsCreateCmd does NOT trigger cobra's auto-emitted
// "Flag --workflow-id has been deprecated..." message. The custom warning
// in resolveWorkflowIDArg is the single source of truth.
func TestWorkflowsTestsCreateCmd_NoCobraAutoDeprecation(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{"workflowsTestsCreateCmd", workflowsTestsCreateCmd},
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
	// on workflowsTestsCreateCmd and capture stderr. We can't actually run
	// the RunE (it would dial the API), so we ParseFlags + invoke the
	// resolver directly the same way RunE does, then assert the captured
	// buffer has only the custom warning.
	cmd := workflowsTestsCreateCmd
	// Suppress cobra usage output during the test.
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

func TestWorkflowsTestsRunsCreateRejectsUnsupportedConsensusLocally(t *testing.T) {
	err := workflowsTestsRunsCreateCmd.Flags().Set("n-consensus", "2")
	if err == nil {
		t.Fatal("expected local parse error for --n-consensus=2")
	}
	if !strings.Contains(err.Error(), "3, 5, or 7") {
		t.Fatalf("error %q does not mention allowed consensus counts", err.Error())
	}
	if resetErr := workflowsTestsRunsCreateCmd.Flags().Set("n-consensus", "0"); resetErr != nil {
		t.Fatalf("reset --n-consensus: %v", resetErr)
	}
}

func TestWorkflowsTestsListCommandsRejectNegativeLimitLocally(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "tests list", cmd: workflowsTestsListCmd},
		{name: "test runs list", cmd: workflowsTestsRunsListCmd},
		{name: "test run results list", cmd: workflowsTestsRunsResultsListCmd},
		{name: "experiment runs list", cmd: workflowsExperimentsRunsListCmd},
		{name: "experiment run results list", cmd: workflowsExperimentsRunsResultsListCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("limit", "-1")
			if err == nil {
				t.Fatal("expected local parse error for --limit=-1")
			}
			if !strings.Contains(err.Error(), "between 0 and 100") {
				t.Fatalf("error %q does not mention backend limit range", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("limit", "0"); resetErr != nil {
				t.Fatalf("reset --limit: %v", resetErr)
			}
		})
	}
}

func TestWorkflowsTestsListCommandsRejectOverLimitLocally(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "tests list", cmd: workflowsTestsListCmd},
		{name: "test runs list", cmd: workflowsTestsRunsListCmd},
		{name: "test run results list", cmd: workflowsTestsRunsResultsListCmd},
		{name: "experiment runs list", cmd: workflowsExperimentsRunsListCmd},
		{name: "experiment run results list", cmd: workflowsExperimentsRunsResultsListCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("limit", "101")
			if err == nil {
				t.Fatal("expected local parse error for --limit=101")
			}
			if !strings.Contains(err.Error(), "between 0 and 100") {
				t.Fatalf("error %q does not mention backend limit range", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("limit", "0"); resetErr != nil {
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
			name:     "test runs list",
			cmd:      workflowsTestsRunsListCmd,
			wantPath: "/workflows/tests/runs",
		},
		{
			name:     "test run results list",
			cmd:      workflowsTestsRunsResultsListCmd,
			args:     []string{"wftestrun_123"},
			wantPath: "/workflows/tests/runs/wftestrun_123/results",
		},
		{
			name:     "experiment runs list",
			cmd:      workflowsExperimentsRunsListCmd,
			wantPath: "/workflows/experiments/runs",
		},
		{
			name:     "experiment run results list",
			cmd:      workflowsExperimentsRunsResultsListCmd,
			args:     []string{"exprun_123"},
			wantPath: "/workflows/experiments/runs/exprun_123/results",
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
				_ = tc.cmd.Flags().Set("limit", "0")
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
			name:     "test runs list",
			cmd:      workflowsTestsRunsListCmd,
			wantPath: "/workflows/tests/runs",
		},
		{
			name:     "test run results list",
			cmd:      workflowsTestsRunsResultsListCmd,
			args:     []string{"wftestrun_123"},
			wantPath: "/workflows/tests/runs/wftestrun_123/results",
		},
		{
			name:     "experiment runs list",
			cmd:      workflowsExperimentsRunsListCmd,
			wantPath: "/workflows/experiments/runs",
		},
		{
			name:     "experiment run results list",
			cmd:      workflowsExperimentsRunsResultsListCmd,
			args:     []string{"exprun_123"},
			wantPath: "/workflows/experiments/runs/exprun_123/results",
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

			if err := tc.cmd.Flags().Set("limit", "0"); err != nil {
				t.Fatal(err)
			}
			tc.cmd.Flags().Lookup("limit").Changed = false
			t.Cleanup(func() {
				_ = tc.cmd.Flags().Set("limit", "0")
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

func TestWorkflowsTestsRejectMalformedTargetAndSourceBeforeRequest(t *testing.T) {
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
			cmd:  workflowsTestsCreateCmd,
			args: []string{"wf_123"},
			flags: map[string]string{
				"target-file":    invalidTarget,
				"source-file":    validSource,
				"assertion-file": validAssertion,
			},
			wantError: "--target-file",
		},
		{
			name: "create invalid source",
			cmd:  workflowsTestsCreateCmd,
			args: []string{"wf_123"},
			flags: map[string]string{
				"target-file":    validTarget,
				"source-file":    invalidSource,
				"assertion-file": validAssertion,
			},
			wantError: "--source-file",
		},
		{
			name:      "update invalid source",
			cmd:       workflowsTestsUpdateCmd,
			args:      []string{"wf_123", "test_123"},
			flags:     map[string]string{"source-file": invalidSource},
			wantError: "--source-file",
		},
		{
			name:      "runs create invalid target",
			cmd:       workflowsTestsRunsCreateCmd,
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
		{name: "block simulate", cmd: workflowsBlocksSimulateCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("n-consensus", "2")
			if err == nil {
				t.Fatal("expected local parse error for --n-consensus=2")
			}
			if !strings.Contains(err.Error(), "3, 5, or 7") {
				t.Fatalf("error %q does not mention allowed consensus counts", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("n-consensus", "0"); resetErr != nil {
				t.Fatalf("reset --n-consensus: %v", resetErr)
			}
		})
	}
}

func TestWorkflowsExperimentsUpdateRejectsExplicitZeroConsensusBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsUpdateCmd.Flags().Set("n-consensus", "0"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsExperimentsUpdateCmd.Flags().Set("n-consensus", "0")
		workflowsExperimentsUpdateCmd.Flags().Lookup("n-consensus").Changed = false
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsExperimentsUpdateCmd.RunE(workflowsExperimentsUpdateCmd, []string{"wf_123", "exp_123"})
	})
	if err == nil {
		t.Fatal("expected explicit zero consensus error")
	}
	if !strings.Contains(stderr, "invalid --n-consensus 0") {
		t.Fatalf("stderr %q does not mention invalid zero consensus", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
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
		missingHandleInputs: `[{"provenance":{"workflow_run_id":"run_123"}}]`,
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
		{name: "missing capture run id", flags: map[string]string{"captures-file": missingRunID}, wantError: "workflow_run_id is required"},
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
				t.Cleanup(func() { _ = workflowsExperimentsCreateCmd.Flags().Set(name, "") })
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
	t.Cleanup(func() { _ = workflowsExperimentsUpdateCmd.Flags().Set("documents-file", "") })

	err := workflowsExperimentsUpdateCmd.RunE(workflowsExperimentsUpdateCmd, []string{"wf_123", "exp_123"})
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

func TestWorkflowsTestsRunsHelpSeparatesRunsAndResults(t *testing.T) {
	if strings.Contains(workflowsTestsRunsCmd.Example, "Recent result rows") {
		t.Fatalf("test runs list example should describe parent runs, got:\n%s", workflowsTestsRunsCmd.Example)
	}
	for _, stale := range []string{"tests execute", "job_id", "batch_id"} {
		if strings.Contains(workflowsTestsRunsCmd.Long+workflowsTestsRunsCmd.Example, stale) {
			t.Fatalf("test runs help should not expose %q, got:\n%s\n%s", stale, workflowsTestsRunsCmd.Long, workflowsTestsRunsCmd.Example)
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

	workflowsExperimentsRunsMetricsGetCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsExperimentsRunsMetricsGetCmd.SetContext(nil) })
	if err := workflowsExperimentsRunsMetricsGetCmd.Flags().Set("view", "banana"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsExperimentsRunsMetricsGetCmd.Flags().Set("view", "summary") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsExperimentsRunsMetricsGetCmd.RunE(workflowsExperimentsRunsMetricsGetCmd, []string{"exprun_123"})
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

func TestWorkflowsExperimentsEligibleBlocksHonorsOutputTable(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/workflows/experiments/eligible-blocks" || r.URL.Query().Get("workflow_id") != "wf_123" {
			t.Fatalf("path = %s?%s, want eligible-blocks path", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"blocks":[{"block_id":"blk_extract","block_label":"Extract","block_type":"extract","experiment_count":2}]}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	var err error
	stdout, stderr := captureStd(t, func() {
		err = workflowsExperimentsEligibleBlocksCmd.RunE(workflowsExperimentsEligibleBlocksCmd, []string{"wf_123"})
	})
	if err != nil {
		t.Fatalf("eligible-blocks: %v", err)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"ID", "NAME", "TYPE", "blk_extract", "Extract", "extract"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}

func TestWorkflowsTestsCreateRejectsAssertionMissingTargetBeforeRequest(t *testing.T) {
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
		if err := workflowsTestsCreateCmd.Flags().Set(flag, path); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = workflowsTestsCreateCmd.Flags().Set(flag, "") })
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsTestsCreateCmd.RunE(workflowsTestsCreateCmd, []string{"wf_123"})
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

func TestWorkflowsTestsCreateReadsLocalFilesBeforeCredentials(t *testing.T) {
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
		if err := workflowsTestsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = workflowsTestsCreateCmd.Flags().Set(flag, "") })
	}

	err := workflowsTestsCreateCmd.RunE(workflowsTestsCreateCmd, []string{"wf_123"})
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

func TestWorkflowsTestsUpdateReadsLocalFilesBeforeCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")

	missingPath := filepath.Join(t.TempDir(), "missing-assertion.json")
	if err := workflowsTestsUpdateCmd.Flags().Set("assertion-file", missingPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsTestsUpdateCmd.Flags().Set("assertion-file", "") })

	err := workflowsTestsUpdateCmd.RunE(workflowsTestsUpdateCmd, []string{"wf_123", "test_123"})
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

// nonEmptyLines splits s on "\n" and returns the non-empty entries. We use
// this to assert exact line counts on warning output without being tripped
// up by the trailing newline from fmt.Fprintln.
func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
