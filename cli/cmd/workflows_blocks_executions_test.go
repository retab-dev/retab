package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestWorkflowsBlockExecutionsCreateUsesCanonicalEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/blocks/executions" || r.URL.RawQuery != "" {
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "sim_123",
			"workflow_id": "wf_123",
			"run_id":      "run_123",
			"block_id":    "blk_extract",
			"block_type":  "extract",
			"lifecycle":   map[string]any{"status": "completed"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{
		"block-id":    "blk_extract",
		"step-id":     "step_iter_0_blk_extract",
		"n-consensus": "5",
	} {
		if err := workflowsBlockExecutionsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
		t.Cleanup(func() { resetWorkflowBlockExecutionsFlag(t, workflowsBlockExecutionsCreateCmd, flag) })
	}
	if err := workflowsBlockExecutionsCreateCmd.Flags().Set("no-check-eligibility", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowBlockExecutionsFlag(t, workflowsBlockExecutionsCreateCmd, "no-check-eligibility") })

	var err error
	captureStd(t, func() {
		err = workflowsBlockExecutionsCreateCmd.RunE(workflowsBlockExecutionsCreateCmd, []string{"run_123"})
	})
	if err != nil {
		t.Fatalf("block executions create: %v", err)
	}
	if body["run_id"] != "run_123" || body["block_id"] != "blk_extract" {
		t.Fatalf("body = %#v", body)
	}
	if body["step_id"] != "step_iter_0_blk_extract" || body["n_consensus"] != float64(5) || body["check_eligibility"] != false {
		t.Fatalf("body = %#v", body)
	}
	if _, ok := body["source_step_id"]; ok {
		t.Fatalf("body = %#v", body)
	}
}

func TestWorkflowsBlockExecutionsListUsesCanonicalEndpoint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawRequest bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/blocks/executions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("run_id") != "run_123" || query.Get("block_id") != "blk_extract" || query.Get("limit") != "10" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		sawRequest = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"id":          "sim_123",
				"workflow_id": "wf_123",
				"run_id":      "run_123",
				"block_id":    "blk_extract",
				"block_type":  "extract",
				"lifecycle":   map[string]any{"status": "completed"},
			}},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlockExecutionsListCmd.Flags().Set("block-id", "blk_extract"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlockExecutionsListCmd.Flags().Set("limit", "10"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		resetWorkflowBlockExecutionsFlag(t, workflowsBlockExecutionsListCmd, "block-id")
		resetWorkflowBlockExecutionsFlag(t, workflowsBlockExecutionsListCmd, "limit")
	})

	var err error
	captureStd(t, func() {
		err = workflowsBlockExecutionsListCmd.RunE(workflowsBlockExecutionsListCmd, []string{"run_123"})
	})
	if err != nil {
		t.Fatalf("block executions list: %v", err)
	}
	if !sawRequest {
		t.Fatal("expected block executions list request")
	}
}

func TestWorkflowsBlockExecutionsRejectsInvalidNConsensusBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be reached, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlockExecutionsCreateCmd.Flags().Set("block-id", "blk_extract"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlockExecutionsCreateCmd.Flags().Set("n-consensus", "4"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		resetWorkflowBlockExecutionsFlag(t, workflowsBlockExecutionsCreateCmd, "block-id")
		resetWorkflowBlockExecutionsFlag(t, workflowsBlockExecutionsCreateCmd, "n-consensus")
	})

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsBlockExecutionsCreateCmd.RunE(workflowsBlockExecutionsCreateCmd, []string{"run_123"})
	})
	if err == nil {
		t.Fatal("expected invalid n-consensus error")
	}
	if !strings.Contains(stderr, "--n-consensus") {
		t.Fatalf("stderr should mention --n-consensus, got:\n%s", stderr)
	}
}

func resetWorkflowBlockExecutionsFlag(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing flag %s", name)
	}
	if value, ok := flag.Value.(*boundedIntFlagValue); ok {
		value.value = ""
		flag.Changed = false
		return
	}
	if err := flag.Value.Set(flag.DefValue); err != nil {
		if v, ok := flag.Value.(*consensusFlagValue); ok {
			v.value = ""
		} else {
			t.Fatalf("reset --%s: %v", name, err)
		}
	}
	flag.Changed = false
}
