//go:build !retab_oagen_cli_workflows_experiments

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// resolveExperimentIDArg is the shared positional resolver that gives
// `get` / `update` / `delete` the same `<workflow-id> <experiment-id>`
// convenience form `runs create` already accepted. These unit cases pin the
// contract directly, independent of any command wiring.
func TestResolveExperimentIDArg(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "single arg is the experiment id", args: []string{"exp_abc"}, want: "exp_abc"},
		{name: "two args take the LAST as the experiment id", args: []string{"wf_abc", "exp_xyz"}, want: "exp_xyz"},
		{name: "surrounding whitespace is trimmed", args: []string{"  exp_pad  "}, want: "exp_pad"},
		{name: "blank single arg errors", args: []string{"   "}, wantErr: true},
		{name: "blank trailing arg errors", args: []string{"wf_abc", "  "}, wantErr: true},
		{name: "no args errors", args: nil, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveExperimentIDArg(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for args=%#v, got %q", tc.args, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for args=%#v: %v", tc.args, err)
			}
			if got != tc.want {
				t.Fatalf("resolveExperimentIDArg(%#v) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}

// TestWorkflowsExperimentsGetTwoArgForm pins the uniformity fix: `get` now
// accepts the same `<workflow-id> <experiment-id>` shape as `runs create`,
// and routes to the experiment id (the LAST positional) — the leading
// workflow id is a scoping hint and must not change which resource is fetched.
func TestWorkflowsExperimentsGetTwoArgForm(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "exp_xyz", "workflow_id": "wf_abc"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsGetCmd.RunE(workflowsExperimentsGetCmd, []string{"wf_abc", "exp_xyz"}); err != nil {
			t.Fatalf("experiments get (2-arg): %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if want := "GET /v1/workflows/experiments/exp_xyz"; strings.Join(requests, ",") != want {
		t.Fatalf("requests = %v, want %q (leading workflow id must not misroute)", requests, want)
	}
	if !strings.Contains(stdout, `"id": "exp_xyz"`) {
		t.Fatalf("expected experiment on stdout, got:\n%s", stdout)
	}
}

// TestWorkflowsExperimentsGetSingleArgStillWorks guards the original
// one-positional form against regression.
func TestWorkflowsExperimentsGetSingleArgStillWorks(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "exp_xyz"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, stderr := captureStd(t, func() {
		if err := workflowsExperimentsGetCmd.RunE(workflowsExperimentsGetCmd, []string{"exp_xyz"}); err != nil {
			t.Fatalf("experiments get (1-arg): %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if want := "GET /v1/workflows/experiments/exp_xyz"; strings.Join(requests, ",") != want {
		t.Fatalf("requests = %v, want %q", requests, want)
	}
}

// TestWorkflowsExperimentsUpdateTwoArgForm pins that `update` routes the
// PATCH to the experiment id (last positional) when given the two-positional
// convenience form. The leading workflow id is first validated with a GET (it
// matches here), then the PATCH is issued to the experiment id.
func TestWorkflowsExperimentsUpdateTwoArgForm(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "exp_xyz", "workflow_id": "wf_abc", "name": "renamed"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsUpdateCmd.Flags().Set("name", "renamed"); err != nil {
		t.Fatalf("set --name: %v", err)
	}
	t.Cleanup(func() { _ = workflowsExperimentsUpdateCmd.Flags().Set("name", "") })

	_, stderr := captureStd(t, func() {
		if err := workflowsExperimentsUpdateCmd.RunE(workflowsExperimentsUpdateCmd, []string{"wf_abc", "exp_xyz"}); err != nil {
			t.Fatalf("experiments update (2-arg): %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if want := "GET /v1/workflows/experiments/exp_xyz,PATCH /v1/workflows/experiments/exp_xyz"; strings.Join(requests, ",") != want {
		t.Fatalf("requests = %v, want %q", requests, want)
	}
}

// TestWorkflowsExperimentsUpdateRejectsMismatchedWorkflow pins that a leading
// workflow id that names a DIFFERENT workflow than the experiment's is rejected
// before any PATCH is issued — matching the server-side guard `runs create`
// enforces, so `update wf_wrong exp_x` cannot silently mutate exp_x.
func TestWorkflowsExperimentsUpdateRejectsMismatchedWorkflow(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "exp_xyz", "workflow_id": "wf_owner"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsUpdateCmd.Flags().Set("name", "renamed"); err != nil {
		t.Fatalf("set --name: %v", err)
	}
	t.Cleanup(func() { _ = workflowsExperimentsUpdateCmd.Flags().Set("name", "") })

	_, _ = captureStd(t, func() {
		err := workflowsExperimentsUpdateCmd.RunE(workflowsExperimentsUpdateCmd, []string{"wf_wrong", "exp_xyz"})
		if err == nil {
			t.Fatalf("expected mismatched-workflow error, got nil")
		}
		if !strings.Contains(err.Error(), "does not match the experiment's workflow") {
			t.Fatalf("error = %v, want workflow-mismatch message", err)
		}
	})
	if want := "GET /v1/workflows/experiments/exp_xyz"; strings.Join(requests, ",") != want {
		t.Fatalf("requests = %v, want %q (no PATCH may be issued on mismatch)", requests, want)
	}
}

// TestWorkflowsExperimentsDeleteTwoArgForm pins that `delete` routes the
// DELETE to the experiment id (last positional) in the two-positional form,
// after validating the leading workflow id with a GET (it matches here).
func TestWorkflowsExperimentsDeleteTwoArgForm(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "exp_xyz", "workflow_id": "wf_abc"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set --yes: %v", err)
	}
	t.Cleanup(func() { _ = workflowsExperimentsDeleteCmd.Flags().Set("yes", "false") })

	_, _ = captureStd(t, func() {
		if err := workflowsExperimentsDeleteCmd.RunE(workflowsExperimentsDeleteCmd, []string{"wf_abc", "exp_xyz"}); err != nil {
			t.Fatalf("experiments delete (2-arg): %v", err)
		}
	})
	if want := "GET /v1/workflows/experiments/exp_xyz,DELETE /v1/workflows/experiments/exp_xyz"; strings.Join(requests, ",") != want {
		t.Fatalf("requests = %v, want %q", requests, want)
	}
}

// TestWorkflowsExperimentsDeleteRejectsMismatchedWorkflow pins the destructive
// footgun fix: `delete wf_wrong exp_x --yes` must NOT delete exp_x when the
// leading workflow id names a different workflow — no DELETE may be issued.
func TestWorkflowsExperimentsDeleteRejectsMismatchedWorkflow(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "exp_xyz", "workflow_id": "wf_owner"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set --yes: %v", err)
	}
	t.Cleanup(func() { _ = workflowsExperimentsDeleteCmd.Flags().Set("yes", "false") })

	_, _ = captureStd(t, func() {
		err := workflowsExperimentsDeleteCmd.RunE(workflowsExperimentsDeleteCmd, []string{"wf_wrong", "exp_xyz"})
		if err == nil {
			t.Fatalf("expected mismatched-workflow error, got nil")
		}
		if !strings.Contains(err.Error(), "does not match the experiment's workflow") {
			t.Fatalf("error = %v, want workflow-mismatch message", err)
		}
	})
	if want := "GET /v1/workflows/experiments/exp_xyz"; strings.Join(requests, ",") != want {
		t.Fatalf("requests = %v, want %q (no DELETE may be issued on mismatch)", requests, want)
	}
}

// TestWorkflowsExperimentsGetRejectsMismatchedWorkflow pins that `get` surfaces
// a mismatched leading workflow id (validated against the fetched experiment)
// rather than silently returning an experiment from another workflow.
func TestWorkflowsExperimentsGetRejectsMismatchedWorkflow(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "exp_xyz", "workflow_id": "wf_owner"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, _ = captureStd(t, func() {
		err := workflowsExperimentsGetCmd.RunE(workflowsExperimentsGetCmd, []string{"wf_wrong", "exp_xyz"})
		if err == nil {
			t.Fatalf("expected mismatched-workflow error, got nil")
		}
		if !strings.Contains(err.Error(), "does not match the experiment's workflow") {
			t.Fatalf("error = %v, want workflow-mismatch message", err)
		}
	})
}
