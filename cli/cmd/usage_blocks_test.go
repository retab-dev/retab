package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func usageBlocksFixture() usageBlockListResponse {
	first := "2026-07-01T09:00:00Z"
	last := "2026-07-01T18:00:00Z"
	after := "block_older"
	return usageBlockListResponse{
		Data: []usageBlockRecord{
			{
				BlockID:         "block_abc123",
				WorkflowID:      "wf_123",
				BlockType:       "extract",
				RunCount:        4,
				ExecutionCount:  9,
				PageCount:       21,
				Credits:         12.5,
				FirstActivityAt: &first,
				LastActivityAt:  &last,
			},
		},
		ListMetadata: usageRunListMetadata{After: &after},
	}
}

func TestUsageBlocksUsesHiddenEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/usage/blocks" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usageBlocksFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := usageBlocksCmd.RunE(usageBlocksCmd, nil); err != nil {
			t.Fatalf("usage blocks: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected usage blocks request")
	}
	for _, want := range []string{`"block_id": "block_abc123"`, `"page_count": 21`, `"credits": 12.5`, `"run_count": 4`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestUsageBlocksForwardsFilterFlags(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usageBlocksFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	set := map[string]string{
		"workflow-id": "wf_123",
		"block-type":  "extract",
		"from-date":   "2026-06-01",
		"to-date":     "2026-06-30",
		"limit":       "50",
		"order":       "asc",
	}
	for k, v := range set {
		if err := usageBlocksCmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set --%s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for k := range set {
			_ = usageBlocksCmd.Flags().Set(k, "")
		}
	})

	captureStd(t, func() {
		if err := usageBlocksCmd.RunE(usageBlocksCmd, nil); err != nil {
			t.Fatalf("usage blocks: %v", err)
		}
	})
	for _, want := range []string{
		"workflow_id=wf_123",
		"block_type=extract",
		"from_date=2026-06-01",
		"to_date=2026-06-30",
		"limit=50",
		"order=asc",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %s, want %s", gotQuery, want)
		}
	}
}

// TestUsageBlocksTableExposesUsageColumnsOnly guards that the table renderer
// surfaces only the safe usage columns and never a confidential field.
func TestUsageBlocksTableExposesUsageColumnsOnly(t *testing.T) {
	headers := map[string]bool{}
	for _, col := range usageBlockColumns {
		headers[col.Header] = true
	}
	for _, want := range []string{"BLOCK_ID", "WORKFLOW", "TYPE", "RUNS", "EXECS", "PAGES", "CREDITS", "LAST_ACTIVITY"} {
		if !headers[want] {
			t.Fatalf("usage blocks table missing column %q", want)
		}
	}
	for _, forbidden := range []string{"FILENAME", "COST", "MODEL", "API_COST", "MARGIN", "RESULT"} {
		if headers[forbidden] {
			t.Fatalf("usage blocks table exposes confidential column %q", forbidden)
		}
	}
}

func TestUsageBlocksRejectsBeforeAndAfterTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_BASE_URL", "http://127.0.0.1:0")

	if err := usageBlocksCmd.Flags().Set("before", "block_a"); err != nil {
		t.Fatalf("set before: %v", err)
	}
	if err := usageBlocksCmd.Flags().Set("after", "block_b"); err != nil {
		t.Fatalf("set after: %v", err)
	}
	t.Cleanup(func() {
		_ = usageBlocksCmd.Flags().Set("before", "")
		_ = usageBlocksCmd.Flags().Set("after", "")
	})

	err := usageBlocksCmd.RunE(usageBlocksCmd, nil)
	if err == nil {
		t.Fatal("expected --before/--after mutual-exclusion error")
	}
}
