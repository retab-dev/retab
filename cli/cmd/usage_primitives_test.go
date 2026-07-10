package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func usagePrimitivesFixture() usagePrimitiveListResponse {
	created := "2026-07-01T12:00:00Z"
	after := "pexec_older"
	return usagePrimitiveListResponse{
		Data: []usagePrimitiveRecord{
			{
				PrimitiveExecutionID: "pexec_abc123",
				Operation:            "extraction",
				WorkflowID:           "wf_123",
				RunID:                "run_123",
				ProjectID:            "proj_123",
				BlockID:              "block_123",
				Status:               "completed",
				ResourceKind:         "schema",
				CreatedAt:            &created,
				PageCount:            7,
				Credits:              12.5,
			},
		},
		ListMetadata: usageRunListMetadata{After: &after},
	}
}

func TestUsagePrimitivesUsesHiddenEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/usage/primitives" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usagePrimitivesFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := usagePrimitivesCmd.RunE(usagePrimitivesCmd, nil); err != nil {
			t.Fatalf("usage primitives: %v", err)
		}
	})
	if !sawRequest {
		t.Fatal("expected usage primitives request")
	}
	for _, want := range []string{`"primitive_execution_id": "pexec_abc123"`, `"operation": "extraction"`, `"page_count": 7`, `"credits": 12.5`} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestUsagePrimitivesForwardsFilterFlags(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usagePrimitivesFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	set := map[string]string{
		"workflow-id": "wf_123",
		"project-id":  "proj_123",
		"run-id":      "run_123",
		"block-id":    "block_123",
		"operation":   "extraction",
		"status":      "completed",
		"from-date":   "2026-06-01",
		"to-date":     "2026-06-30",
		"limit":       "50",
		"order":       "asc",
	}
	for k, v := range set {
		if err := usagePrimitivesCmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set --%s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for k := range set {
			_ = usagePrimitivesCmd.Flags().Set(k, "")
		}
	})

	captureStd(t, func() {
		if err := usagePrimitivesCmd.RunE(usagePrimitivesCmd, nil); err != nil {
			t.Fatalf("usage primitives: %v", err)
		}
	})
	for _, want := range []string{
		"workflow_id=wf_123",
		"project_id=proj_123",
		"run_id=run_123",
		"block_id=block_123",
		"operation=extraction",
		"status=completed",
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

func TestUsagePrimitivesForwardsMetadataFilter(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usagePrimitivesFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := usagePrimitivesCmd.Flags().Set("metadata", "tenant=acme"); err != nil {
		t.Fatalf("set --metadata tenant: %v", err)
	}
	if err := usagePrimitivesCmd.Flags().Set("metadata", "tier=gold"); err != nil {
		t.Fatalf("set --metadata tier: %v", err)
	}
	t.Cleanup(func() {
		f := usagePrimitivesCmd.Flags().Lookup("metadata")
		if sv, ok := f.Value.(interface{ Replace([]string) error }); ok {
			_ = sv.Replace(nil)
		}
		f.Changed = false
	})

	captureStd(t, func() {
		if err := usagePrimitivesCmd.RunE(usagePrimitivesCmd, nil); err != nil {
			t.Fatalf("usage primitives: %v", err)
		}
	})

	// The repeatable key=value pairs are forwarded as a single JSON object under
	// the `metadata` query param, matching the server's parsePrimitiveUsageMetadata
	// contract (a JSON object of string key/value pairs, ANDed together).
	raw := gotQuery.Get("metadata")
	if raw == "" {
		t.Fatalf("metadata query param missing; query = %v", gotQuery)
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("metadata=%q is not a JSON object: %v", raw, err)
	}
	if parsed["tenant"] != "acme" || parsed["tier"] != "gold" {
		t.Fatalf("metadata object = %v, want {tenant:acme, tier:gold}", parsed)
	}
}

func TestUsagePrimitivesRejectsMalformedMetadataPair(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP should not be reached on a malformed --metadata pair, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := usagePrimitivesCmd.Flags().Set("metadata", "novalue"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		f := usagePrimitivesCmd.Flags().Lookup("metadata")
		if sv, ok := f.Value.(interface{ Replace([]string) error }); ok {
			_ = sv.Replace(nil)
		}
		f.Changed = false
	})

	var runErr error
	captureStd(t, func() {
		runErr = usagePrimitivesCmd.RunE(usagePrimitivesCmd, nil)
	})
	if runErr == nil || !strings.Contains(runErr.Error(), "key=value") {
		t.Fatalf("expected key=value error, got %v", runErr)
	}
}

func TestUsagePrimitivesTableExposesUsageColumnsOnly(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usagePrimitivesFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, _ := captureStd(t, func() {
		if err := usagePrimitivesCmd.RunE(usagePrimitivesCmd, nil); err != nil {
			t.Fatalf("usage primitives: %v", err)
		}
	})
	for _, want := range []string{"EXECUTION_ID", "OPERATION", "WORKFLOW", "BLOCK", "PROJECT", "STATUS", "PAGES", "CREDITS", "pexec_abc123", "extraction", "block_123", "12.5"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	// Confidential-safe: no model / cost / token / filename columns leak.
	for _, unwanted := range []string{"MODEL", "PROVIDER", "COST", "TOKENS", "FILENAME", "METADATA"} {
		if strings.Contains(strings.ToUpper(stdout), unwanted) {
			t.Fatalf("stdout should not expose %s:\n%s", unwanted, stdout)
		}
	}
}

func TestUsagePrimitivesRejectsBeforeAndAfterTogether(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP should not be reached on a --before/--after collision, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := usagePrimitivesCmd.Flags().Set("before", "pexec_x"); err != nil {
		t.Fatal(err)
	}
	if err := usagePrimitivesCmd.Flags().Set("after", "pexec_y"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = usagePrimitivesCmd.Flags().Set("before", "")
		_ = usagePrimitivesCmd.Flags().Set("after", "")
	})

	var runErr error
	captureStd(t, func() {
		runErr = usagePrimitivesCmd.RunE(usagePrimitivesCmd, nil)
	})
	if runErr == nil || !strings.Contains(runErr.Error(), "mutually exclusive") {
		t.Fatalf("expected mutually-exclusive error, got %v", runErr)
	}
}
