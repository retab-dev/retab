package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// Bug 8: --fields used `StringArray` on `workflows runs list` but plain
// `String` on every sibling list command (`workflows list`, `workflows
// experiments runs list`, `workflows tests runs list`). The two shapes
// have different invocation patterns:
//
//   - StringArray: --fields a --fields b   (repeatable)
//   - String:      --fields a,b           (comma-separated)
//
// The wire format is comma-separated, so the `String` shape mirrors the
// API and is the majority. Normalize to `String` on `workflows runs
// list` so users can run the same `--fields a,b` syntax everywhere.
func TestWorkflowRunsListFieldsFlagIsCommaSeparatedString(t *testing.T) {
	flag := workflowsRunsListCmd.Flag("fields")
	if flag == nil {
		t.Fatal("workflows runs list missing --fields flag")
	}
	if flag.Value.Type() != "string" {
		t.Fatalf("workflows runs list --fields type = %q, want \"string\" (comma-separated). "+
			"Sibling commands use String; keep the wire shape consistent.",
			flag.Value.Type())
	}
}

// Pin that the comma-separated --fields value reaches the outbound HTTP
// query string verbatim. The CLI must split locally into a []string only
// to satisfy the SDK's Fields type, then the SDK joins them back with
// commas — the round-trip is invisible on the wire.
func TestWorkflowRunsListFieldsCommaSeparatedReachesQueryString(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var seenQuery string
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"list_metadata":{"after":null}}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsListCmd.Flags().Set("fields", "id,workflow.workflow_id"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsListCmd, "fields") })

	if _, stderr := captureStd(t, func() {
		if err := workflowsRunsListCmd.RunE(workflowsRunsListCmd, nil); err != nil {
			t.Fatalf("runs list: %v", err)
		}
	}); stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("expected exactly 1 HTTP call, got %d", got)
	}
	if !strings.Contains(seenQuery, "fields=id%2Cworkflow.workflow_id") &&
		!strings.Contains(seenQuery, "fields=id,workflow.workflow_id") {
		t.Fatalf("expected fields= to carry the comma-separated value through, got %q", seenQuery)
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
