//go:build !retab_oagen_cli_splits

package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSplitsReconstructPostsRequestFile(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Method != http.MethodPost || r.URL.Path != "/v1/splits/reconstruct" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tables": []any{
				map[string]any{
					"label":  "orders",
					"header": []string{"order_id"},
					"rows":   []any{map[string]any{"cells": []string{"A-1"}}},
					"csv":    "order_id\nA-1\n",
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	requestPath := filepath.Join(t.TempDir(), "reconstruct.json")
	if err := writeJSONFile(requestPath, map[string]any{
		"document": map[string]any{
			"id":        "file_abc123",
			"filename":  "orders.xlsx",
			"mime_type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
		"subdocuments": []any{
			map[string]any{
				"name":          "orders",
				"partition_key": "order_id",
				"regions": []any{
					map[string]any{
						"sheet_name":  "Orders",
						"sheet_index": 0,
						"row_start":   1,
						"row_end":     25,
						"header_rows": []int{1},
					},
				},
			},
		},
	}); err != nil {
		t.Fatalf("write request: %v", err)
	}

	splitsReconstructCmd.SetContext(context.Background())
	t.Cleanup(func() {
		splitsReconstructCmd.SetContext(context.Background())
		_ = splitsReconstructCmd.Flags().Set("request-file", "")
		if f := splitsReconstructCmd.Flags().Lookup("request-file"); f != nil {
			f.Changed = false
		}
	})
	if err := splitsReconstructCmd.Flags().Set("request-file", requestPath); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := captureStd(t, func() {
		if err := splitsReconstructCmd.RunE(splitsReconstructCmd, nil); err != nil {
			t.Fatalf("splits reconstruct: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, `"label": "orders"`) {
		t.Fatalf("expected reconstruct response on stdout, got:\n%s", stdout)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("server was hit %d time(s), want 1", got)
	}
	document, ok := body["document"].(map[string]any)
	if !ok || document["id"] != "file_abc123" || document["filename"] != "orders.xlsx" {
		t.Fatalf("document = %#v", body["document"])
	}
	subdocuments, ok := body["subdocuments"].([]any)
	if !ok || len(subdocuments) != 1 {
		t.Fatalf("subdocuments = %#v, want one", body["subdocuments"])
	}
}
