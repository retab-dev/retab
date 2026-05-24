package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func TestWorkflowListCommandsDoNotExposeFieldsFlag(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"workflows list", workflowsListCmd},
		{"workflows runs list", workflowsRunsListCmd},
		{"workflows experiments runs list", workflowsExperimentsRunsListCmd},
		{"workflows tests runs list", workflowsTestsRunsListCmd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if flag := tc.cmd.Flag("fields"); flag != nil {
				t.Fatalf("%s exposes unsupported --fields flag", tc.name)
			}
		})
	}
}

// The RunE-level mutex check is the testable backstop for cobra's
// `MarkFlagsMutuallyExclusive`. When both --before and --after appear,
// the command must fail before issuing any HTTP request.
func TestListCommandsRunERejectsBeforeAndAfterTogether(t *testing.T) {
	cases := []struct {
		name       string
		cmd        *cobra.Command
		positional []string
	}{
		{"workflows list", workflowsListCmd, nil},
		{"workflows runs list", workflowsRunsListCmd, nil},
		{"workflows experiments runs list", workflowsExperimentsRunsListCmd, nil},
		{"workflows tests runs list", workflowsTestsRunsListCmd, nil},
		{"jobs list", jobsListCmd, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("%s: HTTP server should not be reached when --before and --after collide, got %s %s",
					tc.name, r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := tc.cmd.Flags().Set("before", "x"); err != nil {
				t.Fatalf("%s: set --before: %v", tc.name, err)
			}
			if err := tc.cmd.Flags().Set("after", "y"); err != nil {
				t.Fatalf("%s: set --after: %v", tc.name, err)
			}
			t.Cleanup(func() {
				_ = tc.cmd.Flags().Set("before", "")
				_ = tc.cmd.Flags().Set("after", "")
			})

			var runErr error
			captureStd(t, func() {
				runErr = tc.cmd.RunE(tc.cmd, tc.positional)
			})
			if runErr == nil {
				t.Fatalf("%s: expected error for --before + --after collision", tc.name)
			}
			if !strings.Contains(runErr.Error(), "mutually exclusive") {
				t.Fatalf("%s: error %q should mention mutually exclusive", tc.name, runErr.Error())
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("%s: server hit %d time(s), want 0", tc.name, got)
			}
		})
	}
}
