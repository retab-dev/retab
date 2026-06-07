//go:build !retab_oagen_cli_classifications && !retab_oagen_cli_edits && !retab_oagen_cli_partitions && !retab_oagen_cli_splits

package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDocumentResourceListCommandsHonorOutputTable(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		path string
		id   string
	}{
		{name: "classifications", cmd: classificationsListCmd, path: "/v1/classifications", id: "clas_123"},
		{name: "splits", cmd: splitsListCmd, path: "/v1/splits", id: "splt_123"},
		{name: "partitions", cmd: partitionsListCmd, path: "/v1/partitions", id: "part_123"},
		{name: "edits", cmd: editsListCmd, path: "/v1/edits", id: "edt_123"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != tc.path {
					t.Fatalf("path = %s, want %s", r.URL.Path, tc.path)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{
							"id":         tc.id,
							"model":      "retab-small",
							"created_at": "2026-05-15T11:30:00Z",
						},
					},
					"list_metadata": map[string]any{},
				})
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

			if err := tc.cmd.Flags().Set("limit", "1"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = tc.cmd.Flags().Set("limit", "0") })
			tc.cmd.SetContext(context.Background())
			t.Cleanup(func() { tc.cmd.SetContext(context.Background()) })

			stdout, stderr := captureStd(t, func() {
				if err := tc.cmd.RunE(tc.cmd, nil); err != nil {
					t.Fatalf("%s list: %v", tc.name, err)
				}
			})
			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
			if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
				t.Fatalf("expected table output, got JSON:\n%s", stdout)
			}
			for _, want := range []string{"ID", "MODEL", "CREATED_AT", tc.id, "retab-small"} {
				if !strings.Contains(stdout, want) {
					t.Fatalf("expected %q in table output:\n%s", want, stdout)
				}
			}
		})
	}
}

func TestEditTemplatesListHonorsOutputTable(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/edits/templates" {
			t.Fatalf("path = %s, want /v1/edits/templates", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":         "tmpl_123",
					"name":       "Invoice Template",
					"created_at": "2026-05-15T11:30:00Z",
				},
			},
			"list_metadata": map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	if err := editsTemplatesListCmd.Flags().Set("limit", "1"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsTemplatesListCmd.Flags().Set("limit", "0") })
	editsTemplatesListCmd.SetContext(context.Background())
	t.Cleanup(func() { editsTemplatesListCmd.SetContext(context.Background()) })

	stdout, stderr := captureStd(t, func() {
		if err := editsTemplatesListCmd.RunE(editsTemplatesListCmd, nil); err != nil {
			t.Fatalf("edits templates list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"ID", "NAME", "CREATED_AT", "tmpl_123", "Invoice Template"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}
