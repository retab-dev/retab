package cmd

import (
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
				c.Flags().StringArray("allowed-sender", nil, "")
				c.Flags().StringArray("allowed-domain", nil, "")
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
			name: "workflows tests update",
			runE: workflowsTestsUpdateCmd.RunE,
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
			t.Setenv("RETAB_API_KEY", "test-key")
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
