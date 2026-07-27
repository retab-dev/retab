//go:build !retab_oagen_cli_workflows

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestWorkflowVersionsListUsesPublishedVersionsRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var seenPath string
	var seenQuery string
	var seenAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenQuery = r.URL.RawQuery
		seenAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{
				map[string]any{
					"id":                  "wph_1",
					"workflow_id":         "wrk_1",
					"version":             2,
					"workflow_version_id": "ver_1",
					"description":         "release note",
					"block_count":         3,
					"edge_count":          2,
					"published_at":        "2026-06-25T10:00:00Z",
					"is_current":          true,
				},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newWorkflowVersionsListTestCmd()
	if err := cmd.Flags().Set("limit", "7"); err != nil {
		t.Fatalf("set limit: %v", err)
	}
	stdout, _ := captureStd(t, func() {
		if err := cmd.RunE(cmd, []string{"wrk_1"}); err != nil {
			t.Fatalf("versions list: %v", err)
		}
	})

	if seenPath != "/v1/workflows/wrk_1/published-versions" {
		t.Fatalf("path = %s, want published versions route", seenPath)
	}
	if seenQuery != "limit=7" {
		t.Fatalf("query = %s, want limit=7", seenQuery)
	}
	if seenAuth != "Bearer rt_test_key" {
		t.Fatalf("authorization = %q, want bearer api key", seenAuth)
	}
	for _, want := range []string{`"workflow_version_id": "ver_1"`, `"description": "release note"`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %s:\n%s", want, stdout)
		}
	}
}

func newWorkflowVersionsListTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "list", RunE: workflowsVersionsListCmd.RunE}
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "")
	return cmd
}
