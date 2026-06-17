//go:build !retab_oagen_cli_workflows_experiments

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// resetExperimentRunWaitFlags restores the shared --wait/--poll/--timeout
// flags on the create + wait commands to their defaults. The commands are
// package-level singletons, so a test that mutates a flag must hand it back.
func resetExperimentRunWaitFlags(t *testing.T) {
	t.Helper()
	_ = workflowsExperimentsRunsCreateCmd.Flags().Set("wait", "false")
	_ = workflowsExperimentsRunsCreateCmd.Flags().Set("poll-interval-ms", "2000")
	_ = workflowsExperimentsRunsCreateCmd.Flags().Set("timeout-seconds", "600")
	_ = workflowsExperimentsRunsWaitCmd.Flags().Set("poll-interval-ms", "2000")
	_ = workflowsExperimentsRunsWaitCmd.Flags().Set("timeout-seconds", "600")
}

// TestWorkflowsExperimentsRunsCreateWaitPollsUntilTerminal pins the new
// `--wait` behavior: after POSTing the run, the CLI polls GET .../runs/<id>
// on the configured interval until the lifecycle reaches a terminal status
// (here: queued → running → completed) and prints the FINAL run record, not
// the freshly-queued one. This removes the hand-rolled `until` poll loop.
func TestWorkflowsExperimentsRunsCreateWaitPollsUntilTerminal(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var posts, gets atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/experiments/runs":
			posts.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":        "exprun_wait",
				"lifecycle": map[string]any{"status": "queued"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/experiments/runs/exprun_wait":
			// First GET still running, second GET is terminal.
			status := "running"
			if gets.Add(1) >= 2 {
				status = "completed"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":        "exprun_wait",
				"score":     0.9889,
				"lifecycle": map[string]any{"status": status},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsRunsCreateCmd.Flags().Set("wait", "true"); err != nil {
		t.Fatalf("set --wait: %v", err)
	}
	if err := workflowsExperimentsRunsCreateCmd.Flags().Set("poll-interval-ms", "1"); err != nil {
		t.Fatalf("set --poll-interval-ms: %v", err)
	}
	t.Cleanup(func() { resetExperimentRunWaitFlags(t) })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsRunsCreateCmd.RunE(workflowsExperimentsRunsCreateCmd, []string{"exp_123"}); err != nil {
			t.Fatalf("experiment runs create --wait: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if got := posts.Load(); got != 1 {
		t.Fatalf("expected exactly 1 POST, got %d", got)
	}
	if got := gets.Load(); got < 2 {
		t.Fatalf("expected the CLI to poll at least twice before terminal, got %d GET(s)", got)
	}
	if !strings.Contains(stdout, `"status": "completed"`) {
		t.Fatalf("expected final completed run on stdout, got:\n%s", stdout)
	}
	// The printed record must be the polled terminal one (carries the score),
	// not the freshly-queued POST response.
	if !strings.Contains(stdout, "0.9889") {
		t.Fatalf("expected the final polled run (with score) on stdout, got:\n%s", stdout)
	}
}

// TestWorkflowsExperimentsRunsWaitStandalone pins `runs wait <run-id>`:
// it polls GET until terminal and prints the final run.
func TestWorkflowsExperimentsRunsWaitStandalone(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var gets atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/experiments/runs/exprun_wait" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		status := "running"
		if gets.Add(1) >= 2 {
			status = "completed"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "exprun_wait",
			"lifecycle": map[string]any{"status": status},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsRunsWaitCmd.Flags().Set("poll-interval-ms", "1"); err != nil {
		t.Fatalf("set --poll-interval-ms: %v", err)
	}
	t.Cleanup(func() { resetExperimentRunWaitFlags(t) })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsRunsWaitCmd.RunE(workflowsExperimentsRunsWaitCmd, []string{"exprun_wait"}); err != nil {
			t.Fatalf("experiment runs wait: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if got := gets.Load(); got < 2 {
		t.Fatalf("expected at least 2 polls, got %d", got)
	}
	if !strings.Contains(stdout, `"status": "completed"`) {
		t.Fatalf("expected final completed run on stdout, got:\n%s", stdout)
	}
}

// TestWorkflowsExperimentsRunsWaitErrorStatusExitsNonZero pins that a run
// that settles in a non-success terminal status (error/cancelled) surfaces a
// non-nil error so the shell sees a non-zero exit — the run record is still
// printed for forensic context.
func TestWorkflowsExperimentsRunsWaitErrorStatusExitsNonZero(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/experiments/runs/exprun_bad" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "exprun_bad",
			"lifecycle": map[string]any{"status": "error"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsExperimentsRunsWaitCmd.Flags().Set("poll-interval-ms", "1"); err != nil {
		t.Fatalf("set --poll-interval-ms: %v", err)
	}
	t.Cleanup(func() { resetExperimentRunWaitFlags(t) })

	var runErr error
	stdout, _ := captureStd(t, func() {
		runErr = workflowsExperimentsRunsWaitCmd.RunE(workflowsExperimentsRunsWaitCmd, []string{"exprun_bad"})
	})
	if runErr == nil {
		t.Fatalf("expected non-nil error for a run that ended in error status")
	}
	if !strings.Contains(runErr.Error(), "exprun_bad") || !strings.Contains(runErr.Error(), "error") {
		t.Fatalf("error %q should name the run and its terminal status", runErr.Error())
	}
	if !strings.Contains(stdout, `"status": "error"`) {
		t.Fatalf("expected the failed run record on stdout for context, got:\n%s", stdout)
	}
}

// TestWorkflowsExperimentsRunsCreateSingleArg pins the new single-argument
// form: `retab workflows experiments runs create exp_abc` should POST to
// /v1/workflows/experiments/runs with `experiment_id=exp_abc` in the body and
// no `workflow_id` (the server derives it from the experiment record).
func TestWorkflowsExperimentsRunsCreateSingleArg(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/experiments/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "exprun_123",
			"workflow_id":   "wf_123",
			"experiment_id": "exp_123",
			"status":        "running",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsRunsCreateCmd.RunE(workflowsExperimentsRunsCreateCmd, []string{"exp_123"}); err != nil {
			t.Fatalf("experiment runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "exprun_123") {
		t.Fatalf("expected experiment run response on stdout, got:\n%s", stdout)
	}
	if body["experiment_id"] != "exp_123" {
		t.Fatalf("expected experiment_id=exp_123, got body=%#v", body)
	}
	// The CLI must not send workflow_id; the server derives it from the
	// experiment record.
	if _, ok := body["workflow_id"]; ok {
		t.Fatalf("expected workflow_id absent from body, got body=%#v", body)
	}
}

func TestWorkflowsExperimentsRunsListTableRendersStatusAndCounts(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/experiments/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{
				map[string]any{
					"id":                       "exprun_123",
					"workflow_id":              "wf_123",
					"experiment_id":            "exp_123",
					"block_type":               "extract",
					"lifecycle":                map[string]any{"status": "completed"},
					"score":                    0.9444,
					"total_document_count":     2,
					"completed_document_count": 1,
					"error_count":              1,
					"timing":                   map[string]any{"created_at": "2026-06-18T05:38:15Z"},
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatalf("set output: %v", err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsRunsListCmd.RunE(workflowsExperimentsRunsListCmd, []string{"wf_123", "exp_123"}); err != nil {
			t.Fatalf("experiment runs list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	for _, want := range []string{"STATUS", "BLOCK_KIND", "DOCS", "DONE", "ERRORS", "SCORE", "completed", "extract", "2", "1", "0.9444"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("runs table missing %q:\n%s", want, stdout)
		}
	}
}

func TestWorkflowsExperimentsResultsGetUsesFlatResultIDRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/experiments/results/expresult_123" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "expresult_123",
			"run_id":        "exprun_123",
			"experiment_id": "exp_123",
			"document_id":   "expdoc_123",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsResultsGetCmd.RunE(workflowsExperimentsResultsGetCmd, []string{"expresult_123"}); err != nil {
			t.Fatalf("experiment result get: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.Join(requests, ",") != "GET /v1/workflows/experiments/results/expresult_123" {
		t.Fatalf("requests = %v", requests)
	}
	if !strings.Contains(stdout, `"id": "expresult_123"`) {
		t.Fatalf("expected stdout to contain result id, got:\n%s", stdout)
	}
}

func TestWorkflowsExperimentsRunsCancelPrintsCompactResponse(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/experiments/runs/exprun_123/cancel" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "exprun_123",
			"lifecycle": map[string]any{"status": "cancelled"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsExperimentsRunsCancelCmd.RunE(workflowsExperimentsRunsCancelCmd, []string{"exprun_123"}); err != nil {
			t.Fatalf("experiment run cancel: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.Join(requests, ",") != "POST /v1/workflows/experiments/runs/exprun_123/cancel" {
		t.Fatalf("requests = %v", requests)
	}
	if !strings.Contains(stdout, `"id": "exprun_123"`) || !strings.Contains(stdout, `"status": "cancelled"`) {
		t.Fatalf("expected compact cancel response, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "experiment_id") || strings.Contains(stdout, "workflow") {
		t.Fatalf("cancel output should not require a full experiment run shape, got:\n%s", stdout)
	}
}

// TestWorkflowsExperimentsRunsCreateTwoArgsBackwardCompat keeps the
// historical 2-arg invocation alive: callers still pass `<workflow-id>
// <experiment-id>` while we transition to the 1-arg shape. The CLI accepts
// it without surfacing an error and reaches the same flat endpoint.
func TestWorkflowsExperimentsRunsCreateTwoArgsBackwardCompat(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/experiments/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "exprun_123",
			"workflow_id":   "wf_123",
			"experiment_id": "exp_123",
			"status":        "running",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, _ = captureStd(t, func() {
		if err := workflowsExperimentsRunsCreateCmd.RunE(workflowsExperimentsRunsCreateCmd, []string{"wf_123", "exp_123"}); err != nil {
			t.Fatalf("experiment runs create (2-arg): %v", err)
		}
	})
	if body["experiment_id"] != "exp_123" {
		t.Fatalf("expected experiment_id=exp_123, got body=%#v", body)
	}
}

// TestWorkflowsExperimentsRunsCreateTwoArgsForwardsWorkflowID pins R2-6:
// when callers pass `<workflow-id> <experiment-id>` the CLI must forward
// the workflow id to the server in the request body so the server can
// validate the pairing. Without this forwarding, a typo in the workflow
// slot was silently dropped — the server derived `workflow_id` from the
// experiment record and the run executed against the wrong workflow.
func TestWorkflowsExperimentsRunsCreateTwoArgsForwardsWorkflowID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/experiments/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "exprun_123",
			"workflow_id":   "wf_first_pos",
			"experiment_id": "exp_second_pos",
			"status":        "running",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, _ = captureStd(t, func() {
		if err := workflowsExperimentsRunsCreateCmd.RunE(workflowsExperimentsRunsCreateCmd, []string{"wf_first_pos", "exp_second_pos"}); err != nil {
			t.Fatalf("experiment runs create (2-arg): %v", err)
		}
	})
	if got, want := body["workflow_id"], "wf_first_pos"; got != want {
		t.Fatalf("expected workflow_id=%q in body, got %#v (full body: %#v)", want, got, body)
	}
	if got, want := body["experiment_id"], "exp_second_pos"; got != want {
		t.Fatalf("expected experiment_id=%q in body, got %#v (full body: %#v)", want, got, body)
	}
}

// TestWorkflowsExperimentsMetricsGetRequiresDependentFlags asserts that
// the CLI catches missing required flags client-side before issuing an HTTP
// request that the server would reject with a 400.
func TestWorkflowsExperimentsMetricsGetRequiresDependentFlags(t *testing.T) {
	cases := []struct {
		name string
		view string
		// documentID is pre-set before invoking so we can exercise the *next*
		// missing dependent flag (e.g. votes still needs --target-path once
		// --document-id is present).
		documentID string
		wantError  string
	}{
		{name: "by_document requires --document-id", view: "by_document", wantError: "--document-id"},
		{name: "votes requires --document-id", view: "votes", wantError: "--document-id"},
		{name: "votes requires --target-path even with --document-id", view: "votes", documentID: "expdoc_abc", wantError: "--target-path"},
		{name: "by_target requires --target-path", view: "by_target", wantError: "--target-path"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached when required flag is missing, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)

			if err := workflowsExperimentsMetricsGetCmd.Flags().Set("view", tc.view); err != nil {
				t.Fatalf("set --view: %v", err)
			}
			if tc.documentID != "" {
				if err := workflowsExperimentsMetricsGetCmd.Flags().Set("document-id", tc.documentID); err != nil {
					t.Fatalf("set --document-id: %v", err)
				}
			}
			t.Cleanup(func() {
				_ = workflowsExperimentsMetricsGetCmd.Flags().Set("view", "summary")
				_ = workflowsExperimentsMetricsGetCmd.Flags().Set("document-id", "")
				_ = workflowsExperimentsMetricsGetCmd.Flags().Set("target-path", "")
			})

			err := workflowsExperimentsMetricsGetCmd.RunE(workflowsExperimentsMetricsGetCmd, []string{"exprun_abc"})
			if err == nil {
				t.Fatalf("expected error when --view=%s and dependent flag is missing", tc.view)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not mention %q", err.Error(), tc.wantError)
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
	}
}

// TestWorkflowsExperimentsRunsListValidatesFlags pins the client-side guards
// that bring `workflows experiments runs list` in line with its sibling
// `workflows runs list`: --order and --from-date/--to-date are validated at
// flag-parse time, and a reversed --from-date/--to-date range is rejected in
// RunE before any HTTP request is issued (a swapped pair otherwise returns the
// empty set, indistinguishable from "nothing matched").
func TestWorkflowsExperimentsRunsListValidatesFlags(t *testing.T) {
	cmd := workflowsExperimentsRunsListCmd

	t.Run("rejects invalid --order at parse time", func(t *testing.T) {
		t.Cleanup(func() { _ = cmd.Flags().Set("order", "") })
		if err := cmd.Flags().Set("order", "sideways"); err == nil {
			t.Fatal("expected --order=sideways to be rejected at parse time")
		}
	})

	t.Run("rejects malformed --from-date at parse time", func(t *testing.T) {
		t.Cleanup(func() { _ = cmd.Flags().Set("from-date", "") })
		if err := cmd.Flags().Set("from-date", "not-a-date"); err == nil {
			t.Fatal("expected --from-date=not-a-date to be rejected at parse time")
		}
	})

	t.Run("rejects reversed date range before any request", func(t *testing.T) {
		t.Setenv("RETAB_API_KEY", "test-key")
		t.Setenv("HOME", t.TempDir())

		var hits atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			t.Fatalf("server should not be reached for a reversed date range, got %s %s", r.Method, r.URL.Path)
		}))
		defer server.Close()
		t.Setenv("RETAB_API_BASE_URL", server.URL)

		if err := cmd.Flags().Set("from-date", "2026-06-10"); err != nil {
			t.Fatalf("set --from-date: %v", err)
		}
		if err := cmd.Flags().Set("to-date", "2026-06-01"); err != nil {
			t.Fatalf("set --to-date: %v", err)
		}
		t.Cleanup(func() {
			_ = cmd.Flags().Set("from-date", "")
			_ = cmd.Flags().Set("to-date", "")
		})

		err := cmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected an error for a reversed --from-date/--to-date range")
		}
		if !strings.Contains(err.Error(), "from-date") {
			t.Fatalf("error %q should name --from-date", err.Error())
		}
		if got := hits.Load(); got != 0 {
			t.Fatalf("server was hit %d time(s), want 0", got)
		}
	})
}
