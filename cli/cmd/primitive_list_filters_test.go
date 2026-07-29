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

func TestPrimitiveListCommandsReturnCursorAPIError(t *testing.T) {
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
		{name: "partitions", cmd: partitionsListCmd, path: "/v1/partitions"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "rt_test_key")
			t.Setenv("HOME", t.TempDir())

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Fatalf("path = %s, want %s", r.URL.Path, tc.path)
				}
				if got := r.URL.Query().Get("after"); got != "stale_cursor" {
					t.Fatalf("after = %q, want stale_cursor", got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"detail":"Cursor does not match a resource in this organization and environment. It may be stale, malformed, or belong to a different scope."}`))
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			tc.cmd.SetContext(context.Background())
			t.Cleanup(func() { tc.cmd.SetContext(context.Background()) })
			if err := tc.cmd.Flags().Set("after", "stale_cursor"); err != nil {
				t.Fatalf("set --after: %v", err)
			}
			t.Cleanup(func() { _ = tc.cmd.Flags().Set("after", "") })

			var runErr error
			out, stderr := captureStd(t, func() { runErr = tc.cmd.RunE(tc.cmd, nil) })
			if runErr == nil {
				t.Fatalf("%s list accepted stale cursor; output=%s", tc.name, out)
			}
			if !strings.Contains(strings.ToLower(stderr), "cursor") {
				t.Fatalf("%s list stderr did not show API cursor detail: %q", tc.name, stderr)
			}
			if strings.Contains(out, `"data"`) {
				t.Fatalf("%s list printed page data on cursor error: %s", tc.name, out)
			}
		})
	}
}

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
		{name: "partitions", cmd: partitionsListCmd, path: "/v1/partitions"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "rt_test_key")
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
