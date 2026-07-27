package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func usageRunsFixture() usageRunListResponse {
	created := "2026-07-01T12:00:00Z"
	completed := "2026-07-01T12:00:42Z"
	dur := int64(42000)
	after := "run_older"
	return usageRunListResponse{
		Data: []usageRunRecord{
			{
				RunID:       "run_abc123",
				WorkflowID:  "wf_123",
				Status:      "completed",
				TriggerType: "api",
				CreatedAt:   &created,
				CompletedAt: &completed,
				DurationMs:  &dur,
				PageCount:   7,
				Credits:     12.5,
			},
		},
		ListMetadata: usageRunListMetadata{After: &after},
	}
}

func TestUsageRunsUsesHiddenEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/usage/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer rt_test_key" {
			t.Fatalf("Authorization = %q, want Bearer rt_test_key", got)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usageRunsFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := usageRunsCmd.RunE(usageRunsCmd, nil); err != nil {
			t.Fatalf("usage runs: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected usage runs request")
	}
	for _, want := range []string{`"run_id": "run_abc123"`, `"page_count": 7`, `"credits": 12.5`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestUsageRunsForwardsFilterFlags(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usageRunsFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	set := map[string]string{
		"workflow-id":  "wf_123",
		"status":       "completed",
		"trigger-type": "api",
		"from-date":    "2026-06-01",
		"to-date":      "2026-06-30",
		"limit":        "50",
		"order":        "asc",
	}
	for k, v := range set {
		if err := usageRunsCmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set --%s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for k := range set {
			_ = usageRunsCmd.Flags().Set(k, "")
		}
	})

	captureStd(t, func() {
		if err := usageRunsCmd.RunE(usageRunsCmd, nil); err != nil {
			t.Fatalf("usage runs: %v", err)
		}
	})
	for _, want := range []string{
		"workflow_id=wf_123",
		"status=completed",
		"trigger_type=api",
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

func TestUsageRunsTableExposesUsageColumnsOnly(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usageRunsFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, _ := captureStd(t, func() {
		if err := usageRunsCmd.RunE(usageRunsCmd, nil); err != nil {
			t.Fatalf("usage runs: %v", err)
		}
	})
	for _, want := range []string{"RUN_ID", "WORKFLOW", "STATUS", "TRIGGER", "PAGES", "CREDITS", "run_abc123", "12.5"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	// The export is confidential-safe: no filename / cost / model columns leak.
	for _, unwanted := range []string{"FILENAME", "PROVIDER", "COST", "MODEL", "TOKENS"} {
		if strings.Contains(strings.ToUpper(stdout), unwanted) {
			t.Fatalf("stdout should not expose %s:\n%s", unwanted, stdout)
		}
	}
}

func TestUsageRunsRejectsBeforeAndAfterTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP should not be reached on a --before/--after collision, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := usageRunsCmd.Flags().Set("before", "run_x"); err != nil {
		t.Fatal(err)
	}
	if err := usageRunsCmd.Flags().Set("after", "run_y"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = usageRunsCmd.Flags().Set("before", "")
		_ = usageRunsCmd.Flags().Set("after", "")
	})

	var runErr error
	captureStd(t, func() {
		runErr = usageRunsCmd.RunE(usageRunsCmd, nil)
	})
	if runErr == nil || !strings.Contains(runErr.Error(), "mutually exclusive") {
		t.Fatalf("expected mutually-exclusive error, got %v", runErr)
	}
}
