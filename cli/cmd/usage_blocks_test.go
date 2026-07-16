package cmd

import (
	"encoding/csv"
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

func isolateUsageBlocksFlags(t *testing.T) {
	t.Helper()
	clearUsageBlocksFlags(t)
	t.Cleanup(func() { clearUsageBlocksFlags(t) })
}

func clearUsageBlocksFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"workflow-id", "block-type", "from-date", "to-date", "before", "after", "order"} {
		flag := usageBlocksCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing usage blocks flag --%s", name)
		}
		if err := flag.Value.Set(""); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
		flag.Changed = false
	}
	flag := usageBlocksCmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("missing usage blocks flag --limit")
	}
	value, ok := flag.Value.(*boundedIntFlagValue)
	if !ok {
		t.Fatalf("--limit value type = %T, want *boundedIntFlagValue", flag.Value)
	}
	value.value = ""
	flag.Changed = false
}

func TestUsageBlocksUsesHiddenEndpoint(t *testing.T) {
	isolateUsageBlocksFlags(t)
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
	isolateUsageBlocksFlags(t)
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

func TestUsageBlocksRootCommandDispatchesAndPreservesOpaqueCursor(t *testing.T) {
	isolateUsageBlocksFlags(t)
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cursor := "opaque+/= cursor"
	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/usage/blocks" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("after"); got != cursor {
			t.Fatalf("after cursor = %q, want %q (raw query: %s)", got, cursor, r.URL.RawQuery)
		}
		if got := query.Get("workflow_id"); got != "wf_trimmed" {
			t.Fatalf("workflow_id = %q, want trimmed wf_trimmed", got)
		}
		if got := query.Get("block_type"); got != "extract" {
			t.Fatalf("block_type = %q", got)
		}
		if got := query.Get("limit"); got != "2" {
			t.Fatalf("limit = %q", got)
		}
		if got := query.Get("order"); got != "desc" {
			t.Fatalf("order = %q", got)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usageBlocksFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := runRootForTest(t,
			"--output", "json",
			"usage", "blocks",
			"--workflow-id", "  wf_trimmed  ",
			"--block-type", "extract",
			"--after", cursor,
			"--limit", "2",
			"--order", "desc",
		); err != nil {
			t.Fatalf("root usage blocks: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected root-dispatched usage blocks request")
	}
	if !strings.Contains(stdout, `"block_id": "block_abc123"`) {
		t.Fatalf("root-dispatched stdout missing block payload:\n%s", stdout)
	}
}

func TestUsageBlocksCSVOutputUsesSafeColumnsAndEscapesFormulaCells(t *testing.T) {
	isolateUsageBlocksFlags(t)
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	last := "2026-07-01T18:00:00Z"
	response := usageBlockListResponse{
		Data: []usageBlockRecord{{
			BlockID:        "=HYPERLINK(\"https://example.invalid\")",
			WorkflowID:     "wf_csv",
			BlockType:      "extract",
			RunCount:       2,
			ExecutionCount: 3,
			PageCount:      5,
			Credits:        7.25,
			StatusCounts:   map[string]int64{"completed": 1, "failed": 2},
			LastActivityAt: &last,
		}},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := runRootForTest(t, "--output", "csv", "usage", "blocks"); err != nil {
			t.Fatalf("usage blocks csv: %v", err)
		}
	})
	rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv output: %v\n%s", err, stdout)
	}
	if len(rows) != 2 {
		t.Fatalf("csv rows = %d, want header + one data row:\n%s", len(rows), stdout)
	}
	wantHeader := []string{"BLOCK_ID", "WORKFLOW", "TYPE", "RUNS", "EXECS", "PAGES", "FAILED", "CREDITS", "LAST_ACTIVITY"}
	for i, want := range wantHeader {
		if rows[0][i] != want {
			t.Fatalf("header[%d] = %q, want %q; header=%v", i, rows[0][i], want, rows[0])
		}
	}
	if rows[1][0] != "'=HYPERLINK(\"https://example.invalid\")" {
		t.Fatalf("formula-like block id was not neutralized: %q", rows[1][0])
	}
	// FAILED column (index 6) surfaces the failed-execution count from status_counts.
	if rows[1][6] != "2" {
		t.Fatalf("FAILED cell = %q, want 2; row=%v", rows[1][6], rows[1])
	}
	for _, forbidden := range []string{"FILENAME", "API_COST", "PROVIDER", "MARGIN", "MODEL", "RESULT"} {
		if strings.Contains(strings.ToUpper(stdout), forbidden) {
			t.Fatalf("csv output exposed forbidden field %s:\n%s", forbidden, stdout)
		}
	}
}

func TestUsageBlocksRejectsInvalidDateRangeBeforeHTTP(t *testing.T) {
	isolateUsageBlocksFlags(t)
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP should not be reached for reversed date range, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := usageBlocksCmd.Flags().Set("from-date", "2026-07-31"); err != nil {
		t.Fatalf("set from-date: %v", err)
	}
	if err := usageBlocksCmd.Flags().Set("to-date", "2026-07-01"); err != nil {
		t.Fatalf("set to-date: %v", err)
	}
	var runErr error
	captureStd(t, func() {
		runErr = usageBlocksCmd.RunE(usageBlocksCmd, nil)
	})
	if runErr == nil || !strings.Contains(runErr.Error(), "date range is reversed") {
		t.Fatalf("expected reversed date range error, got %v", runErr)
	}
}

func TestUsageBlocksLimitRejectsOutOfRangeValuesLocally(t *testing.T) {
	isolateUsageBlocksFlags(t)
	for _, value := range []string{"0", "10001", "-1"} {
		err := usageBlocksCmd.Flags().Set("limit", value)
		if err == nil {
			t.Fatalf("expected --limit=%s to be rejected", value)
		}
		if !strings.Contains(err.Error(), "between 1 and 10000") {
			t.Fatalf("--limit=%s error %q does not mention range", value, err.Error())
		}
	}
}

// TestUsageBlocksTableExposesUsageColumnsOnly guards that the table renderer
// surfaces only the safe usage columns and never a confidential field.
func TestUsageBlocksTableExposesUsageColumnsOnly(t *testing.T) {
	isolateUsageBlocksFlags(t)
	headers := map[string]bool{}
	for _, col := range usageBlockColumns {
		headers[col.Header] = true
	}
	for _, want := range []string{"BLOCK_ID", "WORKFLOW", "TYPE", "RUNS", "EXECS", "PAGES", "FAILED", "CREDITS", "LAST_ACTIVITY"} {
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

func TestUsageBlocksHelpDescribesOpaqueCursors(t *testing.T) {
	usage := usageBlocksCmd.UsageString()
	for _, want := range []string{"list_metadata.after", "list_metadata.before", "cursor from list_metadata.after"} {
		if !strings.Contains(usage, want) {
			t.Fatalf("usage blocks help missing %q:\n%s", want, usage)
		}
	}
	for _, forbidden := range []string{"block id: return items after", "block id: return items before", "Walk pages from a known block id"} {
		if strings.Contains(usage, forbidden) {
			t.Fatalf("usage blocks help still suggests raw block-id cursor %q:\n%s", forbidden, usage)
		}
	}
}

func TestUsageBlocksRejectsBeforeAndAfterTogether(t *testing.T) {
	isolateUsageBlocksFlags(t)
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
