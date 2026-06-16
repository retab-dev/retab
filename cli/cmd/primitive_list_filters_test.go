package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
)

func TestPrimitiveListCommandsForwardFilterFlags(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
		path string
	}{
		{name: "extractions", cmd: extractionsListCmd, path: "/v1/extractions"},
		{name: "parses", cmd: parsesListCmd, path: "/v1/parses"},
		{name: "edits", cmd: editsListCmd, path: "/v1/edits"},
		{name: "classifications", cmd: classificationsListCmd, path: "/v1/classifications"},
		{name: "splits", cmd: splitsListCmd, path: "/v1/splits"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			var gotFilename, gotFrom, gotTo string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Fatalf("path = %s, want %s", r.URL.Path, tc.path)
				}
				gotFilename = r.URL.Query().Get("filename")
				gotFrom = r.URL.Query().Get("from_date")
				gotTo = r.URL.Query().Get("to_date")
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data":          []map[string]any{},
					"list_metadata": map[string]any{},
				})
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			tc.cmd.SetContext(context.Background())
			t.Cleanup(func() { tc.cmd.SetContext(context.Background()) })
			for flag, value := range map[string]string{
				"filename":  "cli-filter-20260614.pdf",
				"from-date": "2026-01-01T00:00:00Z",
				"to-date":   "2026-12-31T00:00:00Z",
			} {
				if err := tc.cmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("set --%s: %v", flag, err)
				}
				flag := flag
				t.Cleanup(func() { _ = tc.cmd.Flags().Set(flag, "") })
			}

			var err error
			_, _ = captureStd(t, func() { err = tc.cmd.RunE(tc.cmd, nil) })
			if err != nil {
				t.Fatalf("%s list: %v", tc.name, err)
			}
			if gotFilename != "cli-filter-20260614.pdf" {
				t.Errorf("filename query = %q, want cli-filter-20260614.pdf", gotFilename)
			}
			if gotFrom != "2026-01-01T00:00:00Z" {
				t.Errorf("from_date query = %q, want 2026-01-01T00:00:00Z", gotFrom)
			}
			if gotTo != "2026-12-31T00:00:00Z" {
				t.Errorf("to_date query = %q, want 2026-12-31T00:00:00Z", gotTo)
			}
		})
	}
}
