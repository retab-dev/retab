package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestWorkflowsExperimentsRunsCreateSingleArg pins the new single-argument
// form: `retab workflows experiments runs create exp_abc` should POST to
// /workflows/experiments/runs with `experiment_id=exp_abc` in the body and
// no `workflow_id` (the server derives it from the experiment record).
func TestWorkflowsExperimentsRunsCreateSingleArg(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/experiments/runs" {
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

func TestWorkflowsExperimentsRunsResultsGetUsesFlatResultIDRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != "/workflows/experiments/results/expresult_123" {
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
		if err := workflowsExperimentsRunsResultsGetCmd.RunE(workflowsExperimentsRunsResultsGetCmd, []string{"expresult_123"}); err != nil {
			t.Fatalf("experiment result get: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.Join(requests, ",") != "GET /workflows/experiments/results/expresult_123" {
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
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/experiments/runs/exprun_123/cancel" {
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
	if strings.Join(requests, ",") != "POST /workflows/experiments/runs/exprun_123/cancel" {
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
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/experiments/runs" {
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
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/experiments/runs" {
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

// TestWorkflowsExperimentsRunsMetricsGetRequiresDependentFlags asserts that
// the CLI catches missing required flags client-side before issuing an HTTP
// request that the server would reject with a 400.
func TestWorkflowsExperimentsRunsMetricsGetRequiresDependentFlags(t *testing.T) {
	cases := []struct {
		name      string
		view      string
		wantError string
	}{
		{name: "by_document requires --document-id", view: "by_document", wantError: "--document-id"},
		{name: "votes requires --document-id", view: "votes", wantError: "--document-id"},
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

			if err := workflowsExperimentsRunsMetricsGetCmd.Flags().Set("view", tc.view); err != nil {
				t.Fatalf("set --view: %v", err)
			}
			t.Cleanup(func() {
				_ = workflowsExperimentsRunsMetricsGetCmd.Flags().Set("view", "summary")
				_ = workflowsExperimentsRunsMetricsGetCmd.Flags().Set("document-id", "")
				_ = workflowsExperimentsRunsMetricsGetCmd.Flags().Set("target-path", "")
			})

			err := workflowsExperimentsRunsMetricsGetCmd.RunE(workflowsExperimentsRunsMetricsGetCmd, []string{"exprun_abc"})
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
