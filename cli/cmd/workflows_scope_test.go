//go:build !retab_oagen_cli_workflows_blocks && !retab_oagen_cli_workflows_edges && !retab_oagen_cli_workflows_experiments && !retab_oagen_cli_workflows_reviews && !retab_oagen_cli_workflows_evals

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestResolveWorkflowScope pins the shared resolver that harmonizes the two
// co-equal ways a workflow-scoped `list` command names its workflow: a
// positional `[workflow-id]` and a `--workflow-id` flag. Unlike
// resolveWorkflowIDArg (used by create-style commands, which keeps a
// deprecation warning on the flag), list commands treat both forms as equals.
//
// The `required` axis is the only remaining difference between scoped lists:
// blocks/edges/tests/experiments have no org-wide view (required=true), while
// runs/reviews default to a workspace-wide listing (required=false).
func TestResolveWorkflowScope(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		flag     string // "" means unset; use sentinel "<blank>" for explicit empty
		flagSet  bool
		required bool
		want     string
		wantErr  bool
	}{
		{name: "positional only", args: []string{"wf_pos"}, want: "wf_pos"},
		{name: "flag only", flag: "wf_flag", flagSet: true, want: "wf_flag"},
		{name: "both agree", args: []string{"wf_x"}, flag: "wf_x", flagSet: true, want: "wf_x"},
		{name: "both disagree errors", args: []string{"wf_a"}, flag: "wf_b", flagSet: true, wantErr: true},
		{name: "positional whitespace trimmed", args: []string{"  wf_pad  "}, want: "wf_pad"},
		{name: "explicit blank flag errors", flag: "", flagSet: true, wantErr: true},
		{name: "neither, optional → empty", required: false, want: ""},
		{name: "neither, required → error", required: true, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "list"}
			cmd.Flags().String("workflow-id", "", "")
			if tc.flagSet {
				if err := cmd.Flags().Set("workflow-id", tc.flag); err != nil {
					t.Fatalf("set --workflow-id: %v", err)
				}
			}
			got, err := resolveWorkflowScope(cmd, tc.args, tc.required)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %#v, got %q", tc, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %#v: %v", tc, err)
			}
			if got != tc.want {
				t.Fatalf("resolveWorkflowScope = %q, want %q", got, tc.want)
			}
		})
	}
}

// scopeListServer spins up a test backend that records every request's
// `workflow_id` query parameter and returns an empty paginated envelope.
func scopeListServer(t *testing.T, gotWorkflowID *string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotWorkflowID = r.URL.Query().Get("workflow_id")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "list_metadata": map[string]any{}})
	}))
	t.Cleanup(server.Close)
	return server
}

// withFlag sets a command flag for the duration of one eval, restoring it
// afterward so the package-level cobra singletons don't leak between cases.
func withFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set --%s: %v", name, err)
	}
	t.Cleanup(func() { _ = cmd.Flags().Set(name, "") })
}

// TestWorkflowsReviewsListAcceptsPositional pins the harmonization for the
// optional-scope side: `reviews list` (workspace-wide by default) now accepts
// the workflow id positionally, matching `runs list`.
func TestWorkflowsReviewsListAcceptsPositional(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	var gotWorkflowID string
	server := scopeListServer(t, &gotWorkflowID)
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// Positional must be allowed by the Args validator (not NoArgs anymore).
	if err := workflowsReviewsListCmd.Args(workflowsReviewsListCmd, []string{"wf_abc"}); err != nil {
		t.Fatalf("reviews list should accept one positional arg: %v", err)
	}
	_, stderr := captureStd(t, func() {
		if err := workflowsReviewsListCmd.RunE(workflowsReviewsListCmd, []string{"wf_abc"}); err != nil {
			t.Fatalf("reviews list (positional): %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if gotWorkflowID != "wf_abc" {
		t.Fatalf("workflow_id query = %q, want wf_abc", gotWorkflowID)
	}
}

// TestWorkflowsScopedListsAcceptFlag pins the harmonization for the
// required-scope side: blocks/edges/tests/experiments `list` — which used to
// require a positional `<workflow-id>` — now also accept `--workflow-id`,
// matching `runs list`.
func TestWorkflowsScopedListsAcceptFlag(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "blocks", cmd: workflowsBlocksListCmd},
		{name: "edges", cmd: workflowsEdgesListCmd},
		{name: "tests", cmd: workflowsEvalsListCmd},
		{name: "experiments", cmd: workflowsExperimentsListCmd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())
			var gotWorkflowID string
			server := scopeListServer(t, &gotWorkflowID)
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			// Flag form: no positional, id supplied via --workflow-id.
			withFlag(t, tc.cmd, "workflow-id", "wf_flagform")
			_, _ = captureStd(t, func() {
				if err := tc.cmd.RunE(tc.cmd, []string{}); err != nil {
					t.Fatalf("%s list (flag form): %v", tc.name, err)
				}
			})
			if gotWorkflowID != "wf_flagform" {
				t.Fatalf("%s: workflow_id query = %q, want wf_flagform", tc.name, gotWorkflowID)
			}
		})
	}
}

// TestWorkflowsBlocksListScopeDisagreementErrors pins that a required-scope
// list refuses to guess when the positional and flag forms disagree.
func TestWorkflowsBlocksListScopeDisagreementErrors(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	withFlag(t, workflowsBlocksListCmd, "workflow-id", "wf_b")
	err := workflowsBlocksListCmd.RunE(workflowsBlocksListCmd, []string{"wf_a"})
	if err == nil {
		t.Fatalf("expected disagreement error, got nil")
	}
	if !strings.Contains(err.Error(), "twice") {
		t.Fatalf("error = %q, want it to mention the id was specified twice", err.Error())
	}
	if hits != 0 {
		t.Fatalf("no request should be made on disagreement, got %d", hits)
	}
}

// TestWorkflowsBlocksListRequiresWorkflow pins that a required-scope list
// still errors (rather than panicking or listing org-wide) when neither form
// supplies the workflow id.
func TestWorkflowsBlocksListRequiresWorkflow(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	err := workflowsBlocksListCmd.RunE(workflowsBlocksListCmd, []string{})
	if err == nil {
		t.Fatalf("expected required-workflow error, got nil")
	}
}

// TestWorkflowsEvalsRunsListAcceptsPositional pins the last command that was
// missing from the harmonization: `evals runs list` is workspace-wide
// (optional scope, like `runs list` / `reviews list`) and must accept the
// workflow id positionally, not flag-only. Before the fix it was
// `Use: "list [flags]"` with `cobra.NoArgs`, so a positional was rejected
// outright.
func TestWorkflowsEvalsRunsListAcceptsPositional(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	var gotWorkflowID string
	server := scopeListServer(t, &gotWorkflowID)
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// Positional must be allowed by the Args validator (not NoArgs anymore).
	if err := workflowsEvalsRunsListCmd.Args(workflowsEvalsRunsListCmd, []string{"wf_abc"}); err != nil {
		t.Fatalf("evals runs list should accept one positional arg: %v", err)
	}
	_, stderr := captureStd(t, func() {
		if err := workflowsEvalsRunsListCmd.RunE(workflowsEvalsRunsListCmd, []string{"wf_abc"}); err != nil {
			t.Fatalf("evals runs list (positional): %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if gotWorkflowID != "wf_abc" {
		t.Fatalf("workflow_id query = %q, want wf_abc", gotWorkflowID)
	}
}

// TestWorkflowScopedListCommandsAcceptPositional is the drift guard. The
// "identity/parent is positional, filters are flags" convention drifted once
// (evals runs list shipped flag-only). This table walks every workflow-scoped
// `list` command and asserts, structurally, that each one:
//
//   - advertises a `workflow-id` positional in its Use string, and
//   - registers the `--workflow-id` back-compat flag, and
//   - actually accepts one positional through its Args validator
//     (i.e. is not cobra.NoArgs).
//
// A new workflow list command that forgets the positional form fails here
// instead of silently re-introducing the inconsistency.
func TestWorkflowScopedListCommandsAcceptPositional(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "runs list", cmd: workflowsRunsListCmd},
		{name: "reviews list", cmd: workflowsReviewsListCmd},
		{name: "blocks list", cmd: workflowsBlocksListCmd},
		{name: "edges list", cmd: workflowsEdgesListCmd},
		{name: "evals list", cmd: workflowsEvalsListCmd},
		{name: "evals runs list", cmd: workflowsEvalsRunsListCmd},
		{name: "experiments list", cmd: workflowsExperimentsListCmd},
		{name: "experiments runs list", cmd: workflowsExperimentsRunsListCmd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.cmd.Use, "workflow-id") {
				t.Fatalf("%s: Use %q must advertise a workflow-id positional", tc.name, tc.cmd.Use)
			}
			if tc.cmd.Flags().Lookup("workflow-id") == nil {
				t.Fatalf("%s: must register the --workflow-id back-compat flag", tc.name)
			}
			if tc.cmd.Args == nil {
				t.Fatalf("%s: Args is nil (arbitrary) — declare an explicit MaximumNArgs", tc.name)
			}
			if err := tc.cmd.Args(tc.cmd, []string{"wf_positional"}); err != nil {
				t.Fatalf("%s: must accept one positional workflow id, got: %v", tc.name, err)
			}
		})
	}
}
