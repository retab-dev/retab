package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestUsageBlockFailedCell unit-tests the FAILED column extractor across the
// shapes it must tolerate: a populated breakdown, no-failures, a nil breakdown,
// and a non-usageBlockRecord row.
func TestUsageBlockFailedCell(t *testing.T) {
	cases := []struct {
		name string
		row  any
		want string
	}{
		{"has_failures", usageBlockRecord{StatusCounts: map[string]int64{"completed": 8, "failed": 3}}, "3"},
		{"only_completed", usageBlockRecord{StatusCounts: map[string]int64{"completed": 8}}, ""},
		{"zero_failures", usageBlockRecord{StatusCounts: map[string]int64{"completed": 8, "failed": 0}}, ""},
		{"nil_breakdown", usageBlockRecord{StatusCounts: nil}, ""},
		{"empty_breakdown", usageBlockRecord{StatusCounts: map[string]int64{}}, ""},
		{"wrong_row_type", struct{ X int }{X: 1}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := usageBlockFailedCell(tc.row); got != tc.want {
				t.Fatalf("usageBlockFailedCell(%#v) = %q, want %q", tc.row, got, tc.want)
			}
		})
	}
}

// TestUsageBlocksJSONOutputIncludesStatusCounts proves the per-status breakdown
// survives into the CLI's default JSON output (the table drops it to a single
// FAILED column, but json/csv carry the full map).
func TestUsageBlocksJSONOutputIncludesStatusCounts(t *testing.T) {
	isolateUsageBlocksFlags(t)
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	response := usageBlockListResponse{
		Data: []usageBlockRecord{{
			BlockID:        "block_sc",
			WorkflowID:     "wf_sc",
			BlockType:      "extract",
			RunCount:       5,
			ExecutionCount: 9,
			PageCount:      12,
			Credits:        3.5,
			StatusCounts:   map[string]int64{"completed": 7, "failed": 2},
		}},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := usageBlocksCmd.RunE(usageBlocksCmd, nil); err != nil {
			t.Fatalf("usage blocks: %v", err)
		}
	})
	// The JSON round-trips the whole breakdown (not just the FAILED column).
	for _, want := range []string{`"status_counts"`, `"completed": 7`, `"failed": 2`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("json output missing %q:\n%s", want, stdout)
		}
	}
	// re-decode to prove it's a valid map, not a stringified blob.
	var decoded usageBlockListResponse
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("re-decode CLI json: %v\n%s", err, stdout)
	}
	if got := decoded.Data[0].StatusCounts["failed"]; got != 2 {
		t.Fatalf("round-tripped status_counts[failed] = %d, want 2", got)
	}
}
