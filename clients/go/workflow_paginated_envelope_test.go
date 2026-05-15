package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWorkflowExperimentsListUsesPaginatedEnvelope pins the canonical
// `{"data": [...], "list_metadata": {"before": null, "after": null}}` envelope
// for `GET /v1/workflows/{wf}/experiments`. The route used to ship a bare
// list — the migration to PaginatedList is a deliberate breaking change.
func TestWorkflowExperimentsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":             "exp_1",
					"workflow_id":    "wf_1",
					"block_id":       "block_1",
					"block_kind":     "extract",
					"n_consensus":    5,
					"document_count": 0,
					"name":           "Q1",
					"created_at":     "2026-04-14T00:00:00Z",
					"updated_at":     "2026-04-14T00:00:00Z",
					"status":         "draft",
					"is_stale":       false,
					"schema_drift":   "unknown",
				},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Workflows.Experiments.List(context.Background(), "wf_1")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(page.Data); got != 1 {
		t.Fatalf("page.Data length = %d, want 1", got)
	}
	if page.Data[0].ID != "exp_1" {
		t.Fatalf("page.Data[0].ID = %q, want exp_1", page.Data[0].ID)
	}
	if page.ListMetadata.After != "" || page.ListMetadata.Before != "" {
		t.Fatalf("page.ListMetadata = %+v, want zero cursors", page.ListMetadata)
	}
}

// TestWorkflowExperimentRunsListUsesPaginatedEnvelope pins the same
// canonical envelope for `GET /v1/workflows/{wf}/experiments/{id}/runs`. The
// route used to ship `{"runs": [...]}`.
func TestWorkflowExperimentRunsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":                     "exprun_1",
					"definition_fingerprint": "fp",
					"documents_fingerprint":  "fp_doc",
					"status":                 "completed",
					"block_kind":             "extract",
					"n_consensus":            5,
					"created_at":             "2026-04-14T00:00:00Z",
				},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Workflows.Experiments.Runs.List(context.Background(), "wf_1", "exp_1")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(page.Data); got != 1 {
		t.Fatalf("page.Data length = %d, want 1", got)
	}
	if page.Data[0].ID != "exprun_1" {
		t.Fatalf("page.Data[0].ID = %q, want exprun_1", page.Data[0].ID)
	}
}

// TestWorkflowTestsListUsesPaginatedEnvelope pins the canonical envelope for
// `GET /v1/workflows/{wf}/block-tests` (was `{"tests": [...]}`).
func TestWorkflowTestsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "wfnodetest_1", "workflow_id": "wf_1"},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Workflows.Tests.List(context.Background(), ListWorkflowTestsRequest{WorkflowID: "wf_1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("page.Data length = %d, want 1", len(page.Data))
	}
	if id, _ := page.Data[0]["id"].(string); id != "wfnodetest_1" {
		t.Fatalf("page.Data[0][id] = %v, want wfnodetest_1", page.Data[0]["id"])
	}
}

// TestWorkflowTestRunsListUsesPaginatedEnvelope pins the canonical envelope
// for `GET /v1/workflows/{wf}/block-tests/{id}/runs` (was `{"runs": [...]}`).
func TestWorkflowTestRunsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "wfnodetestrun_1", "test_id": "wfnodetest_1"},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Workflows.Tests.Runs.List(context.Background(), "wf_1", "wfnodetest_1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("page.Data length = %d, want 1", len(page.Data))
	}
	if id, _ := page.Data[0]["id"].(string); id != "wfnodetestrun_1" {
		t.Fatalf("page.Data[0][id] = %v, want wfnodetestrun_1", page.Data[0]["id"])
	}
}

func TestWorkflowTestsExecuteDecodesFullQueuedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/wf_1/block-tests/execute" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"batch_id":    "btbatch_1",
			"job_id":      "job_1",
			"status":      "queued",
			"workflow_id": "wf_1",
			"test_id":     "wfnodetest_1",
			"total_tests": 1,
			"target": map[string]any{
				"type":     "block",
				"block_id": "block_transform",
			},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	result, err := client.Workflows.Tests.Execute(context.Background(), ExecuteBlockTestsRequest{
		WorkflowID: "wf_1",
		TestID:     "wfnodetest_1",
		NConsensus: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "queued" || result.WorkflowID != "wf_1" || result.TestID != "wfnodetest_1" || result.TotalTests != 1 {
		t.Fatalf("execute response lost fields: %#v", result)
	}
	if result.Target == nil || (*result.Target)["block_id"] != "block_transform" {
		t.Fatalf("target = %#v, want block_transform", result.Target)
	}
}
