//go:build !retab_oagen_cli_workflows && !retab_oagen_cli_workflows_blocks && !retab_oagen_cli_workflows_experiments && !retab_oagen_cli_workflows_evals

package cmd

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// The workflow update commands build a PATCH request out of whichever
// --flags the user changed. Invoked with zero flags the request body is
// empty, yet the CLI still round-trips to the server — a no-op PATCH
// that silently bumps updated_at and burns a request. Guard against it
// the same way `edits templates update` does: fail fast, before any
// network call, when nothing would actually change.
//
// This runs the real RunE on a fresh cobra.Command (so prior tests'
// flag state can't leak in) with no flags set, and asserts the guard
// rejects it without touching the httptest server.
func TestWorkflowsUpdateCommandsRejectNoOpBeforeRequest(t *testing.T) {
	cases := []struct {
		name          string
		runE          func(*cobra.Command, []string) error
		args          []string
		registerFlags func(*cobra.Command)
	}{
		{
			name: "workflows update",
			runE: workflowsUpdateCmd.RunE,
			args: []string{"wf_123"},
			registerFlags: func(c *cobra.Command) {
				c.Flags().String("name", "", "")
				c.Flags().String("description", "", "")
			},
		},
		{
			name: "workflows blocks update",
			runE: workflowsBlocksUpdateCmd.RunE,
			args: []string{"wf_123", "blk_123"},
			registerFlags: func(c *cobra.Command) {
				c.Flags().String("label", "", "")
				c.Flags().Float64("position-x", 0, "")
				c.Flags().Float64("position-y", 0, "")
				c.Flags().Var(&nonNegativeFloatFlagValue{}, "width", "")
				c.Flags().Var(&nonNegativeFloatFlagValue{}, "height", "")
				c.Flags().String("parent-id", "", "")
				c.Flags().String("config-file", "", "")
				c.Flags().String("merge-config-file", "", "")
			},
		},
		{
			name: "workflows experiments update",
			runE: workflowsExperimentsUpdateCmd.RunE,
			args: []string{"wf_123", "exp_123"},
			registerFlags: func(c *cobra.Command) {
				c.Flags().String("name", "", "")
				c.Flags().Var(&consensusFlagValue{}, "n-consensus", "")
				c.Flags().String("captures-file", "", "")
				c.Flags().String("documents-file", "", "")
			},
		},
		{
			name: "workflows evals update",
			runE: workflowsEvalsUpdateCmd.RunE,
			args: []string{"wf_123", "tst_123"},
			registerFlags: func(c *cobra.Command) {
				c.Flags().String("name", "", "")
				c.Flags().String("assertion-file", "", "")
				c.Flags().String("source-file", "", "")
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "rt_test_key")
			t.Setenv("HOME", t.TempDir())

			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				hits.Add(1)
				http.Error(w, "server should not be reached", http.StatusInternalServerError)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "test-update", RunE: tc.runE}
			tc.registerFlags(cmd)

			err := cmd.RunE(cmd, tc.args)
			if err == nil {
				t.Fatal("expected a no-op update error")
			}
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				err = unwrapped
			}
			if !strings.Contains(err.Error(), "nothing to update") {
				t.Fatalf("error %q does not mention \"nothing to update\"", err.Error())
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want the guard to fail before any request", got)
			}
		})
	}
}

func TestWorkflowsBlocksUpdateSendsExplicitZeroPositions(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotBody map[string]any
	var gotPath string
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "blk_123",
			"workflow_id": "wf_123",
			"type": "function",
			"label": "Function",
			"position_x": 0,
			"position_y": 0,
			"updated_at": "2026-06-18T00:00:00Z"
		}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-block-update", RunE: workflowsBlocksUpdateCmd.RunE}
	cmd.Flags().String("label", "", "")
	cmd.Flags().Float64("position-x", 0, "")
	cmd.Flags().Float64("position-y", 0, "")
	cmd.Flags().Var(&nonNegativeFloatFlagValue{}, "width", "")
	cmd.Flags().Var(&nonNegativeFloatFlagValue{}, "height", "")
	cmd.Flags().String("parent-id", "", "")
	cmd.Flags().String("config-file", "", "")
	cmd.Flags().String("merge-config-file", "", "")
	if err := cmd.Flags().Set("position-x", "0"); err != nil {
		t.Fatalf("set position-x: %v", err)
	}
	if err := cmd.Flags().Set("position-y", "0"); err != nil {
		t.Fatalf("set position-y: %v", err)
	}

	if err := cmd.RunE(cmd, []string{"wf_123", "blk_123"}); err != nil {
		t.Fatalf("update command: %v", err)
	}

	if gotPath != "/v1/workflows/blocks/blk_123" {
		t.Fatalf("path = %q, want /v1/workflows/blocks/blk_123", gotPath)
	}
	if !strings.Contains(gotQuery, "workflow_id=wf_123") {
		t.Fatalf("query = %q, want workflow_id=wf_123", gotQuery)
	}
	if gotBody["position_x"] != float64(0) {
		t.Fatalf("position_x body value = %#v, want explicit 0; body=%v", gotBody["position_x"], gotBody)
	}
	if gotBody["position_y"] != float64(0) {
		t.Fatalf("position_y body value = %#v, want explicit 0; body=%v", gotBody["position_y"], gotBody)
	}
}
