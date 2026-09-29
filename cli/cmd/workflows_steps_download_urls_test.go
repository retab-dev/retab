package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// stepsDownloadURLServer answers both step routes with a split step whose file
// output carries a download_url, recording the include_download_urls query
// value (and whether it was sent at all) for each request.
func stepsDownloadURLServer(t *testing.T, gotQuery *[]string) *httptest.Server {
	t.Helper()
	step := map[string]any{
		"step_id":     "run_1_block_split",
		"run_id":      "run_1",
		"block_id":    "split",
		"block_type":  "split",
		"block_label": "Split",
		"lifecycle":   map[string]any{"status": "completed"},
		"handle_outputs": map[string]any{
			"output-file-invoice": map[string]any{
				"type":         "file",
				"document":     map[string]any{"id": "file_inv", "filename": "invoice.pdf", "mime_type": "application/pdf"},
				"download_url": "https://storage.retab.test/file_inv?sig=1",
				"expires_at":   "2026-09-29T12:00:00Z",
			},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, sent := r.URL.Query()["include_download_urls"]
		if sent {
			*gotQuery = append(*gotQuery, strings.Join(values, ","))
		} else {
			*gotQuery = append(*gotQuery, "<unset>")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/steps":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []any{step},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case "/v1/workflows/steps/run_1_block_split":
			_ = json.NewEncoder(w).Encode(step)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
}

func runStepsCommand(t *testing.T, cmd *cobra.Command, args []string, includeDownloadURLs string) string {
	t.Helper()
	t.Cleanup(func() { setFlagClean(cmd, "include-download-urls", "false") })
	if includeDownloadURLs != "" {
		if err := cmd.Flags().Set("include-download-urls", includeDownloadURLs); err != nil {
			t.Fatalf("set --include-download-urls: %v", err)
		}
	}
	stdout, stderr := captureStd(t, func() {
		if err := cmd.RunE(cmd, args); err != nil {
			t.Fatalf("%s: %v", cmd.Name(), err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	return stdout
}

func TestWorkflowsStepsIncludeDownloadURLsFlag(t *testing.T) {
	cases := []struct {
		name      string
		cmd       *cobra.Command
		args      []string
		flag      string
		wantQuery string
	}{
		{"list sends the flag", workflowsStepsListCmd, []string{"run_1"}, "true", "true"},
		{"list leaves the server default when omitted", workflowsStepsListCmd, []string{"run_1"}, "", "<unset>"},
		{"get sends the flag", workflowsStepsGetCmd, []string{"run_1_block_split"}, "true", "true"},
		{"get leaves the server default when omitted", workflowsStepsGetCmd, []string{"run_1_block_split"}, "", "<unset>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "rt_test_key")
			t.Setenv("HOME", t.TempDir())
			var gotQuery []string
			server := stepsDownloadURLServer(t, &gotQuery)
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			stdout := runStepsCommand(t, tc.cmd, tc.args, tc.flag)

			if len(gotQuery) != 1 || gotQuery[0] != tc.wantQuery {
				t.Fatalf("include_download_urls query = %v, want [%s]", gotQuery, tc.wantQuery)
			}
			if !strings.Contains(stdout, "https://storage.retab.test/file_inv?sig=1") {
				t.Fatalf("download_url missing from output:\n%s", stdout)
			}
		})
	}
}

func TestWorkflowsStepsHelpDocumentsIncludeDownloadURLs(t *testing.T) {
	for _, cmd := range []*cobra.Command{workflowsStepsListCmd, workflowsStepsGetCmd} {
		if cmd.Flags().Lookup("include-download-urls") == nil {
			t.Fatalf("%s has no --include-download-urls flag", cmd.Name())
		}
		if !strings.Contains(cmd.Long, "--include-download-urls") || !strings.Contains(cmd.Example, "--include-download-urls") {
			t.Fatalf("%s help should explain and show --include-download-urls", cmd.Name())
		}
	}
}
