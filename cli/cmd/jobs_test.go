package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestJobsListOutputTable(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/jobs" {
			t.Fatalf("path = %s, want /jobs", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		// The backend serializes job timestamps as ISO 8601 strings on the
		// API boundary (see backend/main_server/main_server/services/v1/jobs/
		// models.py); mirror that here so the table renderer is exercised
		// against the real wire shape.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":         "job_123",
					"status":     "completed",
					"endpoint":   "/v1/parses",
					"created_at": "2026-05-13T13:32:54Z",
				},
			},
			"has_more": false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	if err := jobsListCmd.Flags().Set("limit", "1"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = jobsListCmd.Flags().Set("limit", "0") })

	stdout, stderr := captureStd(t, func() {
		if err := jobsListCmd.RunE(jobsListCmd, nil); err != nil {
			t.Fatalf("jobs list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected table output, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"ID", "STATUS", "ENDPOINT", "CREATED_AT", "job_123", "completed", "/v1/parses", "2026-05-13T13:32:54Z"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "1778679174") {
		t.Fatalf("jobs table should not contain raw epoch ints (server emits ISO 8601), got:\n%s", stdout)
	}
	if strings.Contains(stdout, "TYPE") {
		t.Fatalf("jobs table should label the status column as STATUS, got:\n%s", stdout)
	}
}

func TestJobsWaitHelpMentionsEveryTerminalStatus(t *testing.T) {
	for _, status := range []string{"completed", "failed", "cancelled", "expired"} {
		if !strings.Contains(jobsWaitCmd.Long, status) {
			t.Fatalf("jobs wait help should mention terminal status %q:\n%s", status, jobsWaitCmd.Long)
		}
	}
}

func TestJobsCommandDoesNotExposeRetrieveFull(t *testing.T) {
	for _, cmd := range jobsCmd.Commands() {
		if cmd.Name() == "retrieve-full" {
			t.Fatal("jobs command still exposes retrieve-full")
		}
	}
}

func TestJobsCreateReadsRequestFileBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test-jobs-create", RunE: jobsCreateCmd.RunE}
	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().String("request-file", "", "")
	cmd.Flags().StringArray("metadata", nil, "")

	_ = cmd.Flags().Set("endpoint", "/v1/parses")
	_ = cmd.Flags().Set("request-file", "/tmp/missing-request.json")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected request-file error")
	}
	if !strings.Contains(err.Error(), "--request-file") {
		t.Fatalf("error %q does not mention --request-file", err.Error())
	}
	if strings.Contains(err.Error(), "credentials") {
		t.Fatalf("error %q checked credentials before reading --request-file", err.Error())
	}
}

func TestJobsRejectsInvalidNumericFlagsLocally(t *testing.T) {
	cases := []struct {
		name      string
		cmd       string
		flag      string
		value     string
		wantError string
		reset     string
	}{
		{name: "negative list limit", cmd: "list", flag: "limit", value: "-1", wantError: "non-negative", reset: "0"},
		{name: "invalid list order", cmd: "list", flag: "order", value: "sideways", wantError: "asc", reset: ""},
		{name: "negative wait poll interval", cmd: "wait", flag: "poll-interval-ms", value: "-1", wantError: "non-negative", reset: "0"},
		{name: "negative wait timeout", cmd: "wait", flag: "timeout-seconds", value: "-1", wantError: "non-negative", reset: "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := jobsListCmd
			if tc.cmd == "wait" {
				cmd = jobsWaitCmd
			}
			err := cmd.Flags().Set(tc.flag, tc.value)
			if err == nil {
				t.Fatalf("expected local parse error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if resetErr := cmd.Flags().Set(tc.flag, tc.reset); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestJobsListRejectsInvalidDateFlagsLocally(t *testing.T) {
	cases := []struct {
		name string
		flag string
	}{
		{name: "invalid from date", flag: "from-date"},
		{name: "invalid to date", flag: "to-date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := jobsListCmd.Flags().Set(tc.flag, "not-a-date")
			if err == nil {
				t.Fatalf("expected local parse error for --%s=not-a-date", tc.flag)
			}
			if !strings.Contains(err.Error(), "YYYY-MM-DD") {
				t.Fatalf("error %q does not contain YYYY-MM-DD", err.Error())
			}
			if resetErr := jobsListCmd.Flags().Set(tc.flag, ""); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestJobsListRejectsInvalidFilterValuesBeforeRequest(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
	}{
		{name: "invalid status", flag: "status", value: "banana", wantError: "invalid --status"},
		{name: "invalid endpoint", flag: "endpoint", value: "/v1/nope", wantError: "invalid --endpoint"},
		{name: "invalid source", flag: "source", value: "nonsense", wantError: "invalid --source"},
		{name: "invalid document type", flag: "document-type", value: "notatype", wantError: "invalid --document-type"},
		{name: "limit too large", flag: "limit", value: "101", wantError: "invalid --limit"},
		{name: "filename regex too long", flag: "filename-regex", value: strings.Repeat("x", 257), wantError: "invalid --filename-regex"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			hits := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				t.Fatalf("server should not be reached for invalid local filter, got %s %s", r.Method, r.URL.String())
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := jobsListCmd.Flags().Set(tc.flag, tc.value); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { resetJobsListFlag(t, tc.flag) })

			var err error
			_, stderr := captureStd(t, func() {
				err = jobsListCmd.RunE(jobsListCmd, nil)
			})
			if err == nil {
				t.Fatalf("expected local validation error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if hits != 0 {
				t.Fatalf("server was hit %d time(s), want 0", hits)
			}
		})
	}
}

func TestJobsListNormalizesDocumentTypesLikeBackend(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/jobs" {
			t.Fatalf("path = %s, want /jobs", r.URL.Path)
		}
		got := r.URL.Query()["document_type"]
		want := []string{"pdf", "docx"}
		if len(got) != len(want) {
			t.Fatalf("document_type = %#v, want %#v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("document_type = %#v, want %#v", got, want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":     []map[string]any{},
			"has_more": false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for _, value := range []string{"PDF,pdf", " docx ", "pdf"} {
		if err := jobsListCmd.Flags().Set("document-type", value); err != nil {
			t.Fatalf("set --document-type=%q: %v", value, err)
		}
	}
	t.Cleanup(func() { resetJobsListFlag(t, "document-type") })

	var err error
	_, stderr := captureStd(t, func() {
		err = jobsListCmd.RunE(jobsListCmd, nil)
	})
	if err != nil {
		t.Fatalf("jobs list: %v\nstderr:\n%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if hits != 1 {
		t.Fatalf("server was hit %d time(s), want 1", hits)
	}
}

func TestJobsListRejectsInvalidFilterCombinationsBeforeRequest(t *testing.T) {
	cases := []struct {
		name      string
		flags     map[string]string
		wantError string
	}{
		{
			name: "include response limit too large",
			flags: map[string]string{
				"include-response": "true",
				"limit":            "21",
			},
			wantError: "include-response",
		},
		{
			name: "api source with project id",
			flags: map[string]string{
				"source":     "api",
				"project-id": "proj_123",
			},
			wantError: "source=api",
		},
		{
			name: "api source with workflow id",
			flags: map[string]string{
				"source":      "api",
				"workflow-id": "wf_123",
			},
			wantError: "source=api",
		},
		{
			name: "unanchored quantified filename regex",
			flags: map[string]string{
				"filename-regex": "foo.*",
			},
			wantError: "invalid --filename-regex",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			hits := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				t.Fatalf("server should not be reached for invalid local filter, got %s %s", r.Method, r.URL.String())
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			for flag, value := range tc.flags {
				if err := jobsListCmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("set --%s: %v", flag, err)
				}
				t.Cleanup(func() { resetJobsListFlag(t, flag) })
			}

			var err error
			_, stderr := captureStd(t, func() {
				err = jobsListCmd.RunE(jobsListCmd, nil)
			})
			if err == nil {
				t.Fatal("expected local validation error")
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if hits != 0 {
				t.Fatalf("server was hit %d time(s), want 0", hits)
			}
		})
	}
}

func TestJobsCreateRejectsInvalidEndpointBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		t.Fatalf("server should not be reached for invalid local endpoint, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	requestFile := t.TempDir() + "/request.json"
	if err := os.WriteFile(requestFile, []byte(`{"document":{"url":"https://example.com/doc.pdf"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := jobsCreateCmd.Flags().Set("endpoint", "/v1/nope"); err != nil {
		t.Fatal(err)
	}
	if err := jobsCreateCmd.Flags().Set("request-file", requestFile); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = jobsCreateCmd.Flags().Set("endpoint", "")
		_ = jobsCreateCmd.Flags().Set("request-file", "")
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = jobsCreateCmd.RunE(jobsCreateCmd, nil)
	})
	if err == nil {
		t.Fatal("expected local validation error for invalid endpoint")
	}
	if !strings.Contains(stderr, "invalid --endpoint") {
		t.Fatalf("stderr %q does not contain invalid --endpoint", stderr)
	}
	if hits != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits)
	}
}

func TestJobsCreateRejectsBlankEndpointBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		t.Fatalf("server should not be reached for blank local endpoint, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	requestFile := t.TempDir() + "/request.json"
	if err := os.WriteFile(requestFile, []byte(`{"document":{"url":"https://example.com/doc.pdf"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := jobsCreateCmd.Flags().Set("endpoint", "   "); err != nil {
		t.Fatal(err)
	}
	if err := jobsCreateCmd.Flags().Set("request-file", requestFile); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = jobsCreateCmd.Flags().Set("endpoint", "")
		_ = jobsCreateCmd.Flags().Set("request-file", "")
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = jobsCreateCmd.RunE(jobsCreateCmd, nil)
	})
	if err == nil {
		t.Fatal("expected local validation error for blank endpoint")
	}
	if !strings.Contains(stderr, "--endpoint must not be blank") {
		t.Fatalf("stderr %q does not contain blank endpoint error", stderr)
	}
	if hits != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits)
	}
}

func TestJobsCreateRejectsInvalidMetadataBeforeRequest(t *testing.T) {
	cases := []struct {
		name      string
		metadata  []string
		wantError string
	}{
		{name: "too many pairs", metadata: []string{
			"k00=v", "k01=v", "k02=v", "k03=v", "k04=v", "k05=v", "k06=v", "k07=v", "k08=v",
			"k09=v", "k10=v", "k11=v", "k12=v", "k13=v", "k14=v", "k15=v", "k16=v",
		}, wantError: "at most 16"},
		{name: "key too long", metadata: []string{strings.Repeat("k", 65) + "=v"}, wantError: "metadata key"},
		{name: "value too long", metadata: []string{"k=" + strings.Repeat("v", 513)}, wantError: "metadata value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			hits := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				t.Fatalf("server should not be reached for invalid local metadata, got %s %s", r.Method, r.URL.String())
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			requestFile := t.TempDir() + "/request.json"
			if err := os.WriteFile(requestFile, []byte(`{"document":{"url":"https://example.com/doc.pdf"}}`), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := jobsCreateCmd.Flags().Set("endpoint", "/v1/extractions"); err != nil {
				t.Fatal(err)
			}
			if err := jobsCreateCmd.Flags().Set("request-file", requestFile); err != nil {
				t.Fatal(err)
			}
			for _, metadata := range tc.metadata {
				if err := jobsCreateCmd.Flags().Set("metadata", metadata); err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() {
				_ = jobsCreateCmd.Flags().Set("endpoint", "")
				_ = jobsCreateCmd.Flags().Set("request-file", "")
				if slice, ok := jobsCreateCmd.Flags().Lookup("metadata").Value.(pflag.SliceValue); ok {
					_ = slice.Replace(nil)
				}
				jobsCreateCmd.Flags().Lookup("metadata").Changed = false
			})

			var err error
			_, stderr := captureStd(t, func() {
				err = jobsCreateCmd.RunE(jobsCreateCmd, nil)
			})
			if err == nil {
				t.Fatal("expected local validation error for invalid metadata")
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if hits != 0 {
				t.Fatalf("server was hit %d time(s), want 0", hits)
			}
		})
	}
}

func TestJobsWaitReturnsNonZeroForFailedJob(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/jobs/job_failed" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "job_failed",
			"status": "failed",
			"error": map[string]any{
				"message": "parse failed",
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	var err error
	stdout, stderr := captureStd(t, func() {
		err = jobsWaitCmd.RunE(jobsWaitCmd, []string{"job_failed"})
	})
	if err == nil {
		t.Fatal("expected non-zero wait result for failed job")
	}
	if !strings.Contains(stdout, `"status": "failed"`) {
		t.Fatalf("expected failed job JSON on stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "job job_failed ended with status failed") {
		t.Fatalf("expected failed status on stderr, got:\n%s", stderr)
	}
}

func TestJobWaitTerminalErrorStatuses(t *testing.T) {
	for _, tc := range []struct {
		status  string
		wantErr bool
	}{
		{status: "completed", wantErr: false},
		{status: "failed", wantErr: true},
		{status: "cancelled", wantErr: true},
		{status: "expired", wantErr: true},
		{status: "in_progress", wantErr: false},
	} {
		t.Run(tc.status, func(t *testing.T) {
			job := map[string]any{"id": "job_123", "status": tc.status}
			err := jobWaitTerminalError((*retab.Job)(&job))
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for status %q", tc.status)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for status %q, got %v", tc.status, err)
			}
		})
	}
}

func TestJobsRetrieveReturnsZeroForFailedJob(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/jobs/job_failed" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "job_failed",
			"status": "failed",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	var err error
	stdout, stderr := captureStd(t, func() {
		err = jobsRetrieveCmd.RunE(jobsRetrieveCmd, []string{"job_failed"})
	})
	if err != nil {
		t.Fatalf("retrieve should not fail for failed job records, got %v", err)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, `"status": "failed"`) {
		t.Fatalf("expected failed job JSON on stdout, got:\n%s", stdout)
	}
}

func TestJobsRetrieveHelpUsesCurrentStatusNames(t *testing.T) {
	if !strings.Contains(jobsRetrieveCmd.Example, "queued|in_progress") {
		t.Fatalf("retrieve example should poll queued|in_progress statuses:\n%s", jobsRetrieveCmd.Example)
	}
	if strings.Contains(jobsRetrieveCmd.Example, `"running"`) {
		t.Fatalf("retrieve example should not use stale running status:\n%s", jobsRetrieveCmd.Example)
	}
}

func resetJobsListFlag(t *testing.T, name string) {
	t.Helper()
	flag := jobsListCmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing jobs list flag %q", name)
	}
	if slice, ok := flag.Value.(pflag.SliceValue); ok {
		if err := slice.Replace(nil); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
		flag.Changed = false
		return
	}
	if err := jobsListCmd.Flags().Set(name, flag.DefValue); err != nil {
		t.Fatalf("reset --%s: %v", name, err)
	}
	flag.Changed = false
}
