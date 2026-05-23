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
// for `GET /v1/workflows/experiments?workflow_id={wf}`. The route used to ship a bare
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

// TestWorkflowExperimentRunsListUsesPaginatedEnvelope pins the canonical
// envelope for `GET /v1/workflows/experiments/runs`. The route used to ship
// scoped run lists.
func TestWorkflowExperimentRunsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/experiments/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("workflow_id") != "wf_1" ||
			query.Get("experiment_id") != "exp_1" ||
			query.Get("block_id") != "block_1" ||
			query.Get("status") != "completed" ||
			query.Get("statuses") != "completed,error" ||
			query.Get("exclude_status") != "cancelled" ||
			query.Get("trigger_type") != "api" ||
			query.Get("trigger_types") != "api,manual_run" ||
			query.Get("from_date") != "2026-05-01" ||
			query.Get("to_date") != "2026-05-18" ||
			query.Get("sort_by") != "created_at" ||
			query.Get("fields") != "id,lifecycle" ||
			query.Get("before") != "exprun_before" ||
			query.Get("after") != "exprun_after" ||
			query.Get("limit") != "10" ||
			query.Get("order") != "asc" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":                     "exprun_1",
					"definition_fingerprint": "fp",
					"documents_fingerprint":  "fp_doc",
					"lifecycle":              map[string]any{"status": "completed"},
					"block_kind":             "extract",
					"n_consensus":            5,
				},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Workflows.Experiments.Runs.List(context.Background(), &ListExperimentRunsParams{
		WorkflowID:    "wf_1",
		ExperimentID:  "exp_1",
		BlockID:       "block_1",
		Status:        "completed",
		Statuses:      []string{"completed", "error"},
		ExcludeStatus: "cancelled",
		TriggerType:   "api",
		TriggerTypes:  []string{"api", "manual_run"},
		FromDate:      "2026-05-01",
		ToDate:        "2026-05-18",
		SortBy:        "created_at",
		Fields:        []string{"id", "lifecycle"},
		Before:        "exprun_before",
		After:         "exprun_after",
		Limit:         10,
		Order:         "asc",
	})
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
// `GET /v1/workflows/tests?workflow_id={wf}` (was `{"tests": [...]}`).
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
// for `GET /v1/workflows/tests/runs`.
func TestWorkflowTestRunsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/tests/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("workflow_id") != "wf_1" ||
			query.Get("test_id") != "wfnodetest_1" ||
			query.Get("target_block_id") != "block_1" ||
			query.Get("status") != "passed" ||
			query.Get("statuses") != "passed,failed" ||
			query.Get("exclude_status") != "cancelled" ||
			query.Get("trigger_type") != "api" ||
			query.Get("trigger_types") != "api,manual_run" ||
			query.Get("from_date") != "2026-05-01" ||
			query.Get("to_date") != "2026-05-18" ||
			query.Get("sort_by") != "created_at" ||
			query.Get("fields") != "id,lifecycle" ||
			query.Get("before") != "wftestrun_before" ||
			query.Get("after") != "wftestrun_after" ||
			query.Get("limit") != "10" ||
			query.Get("order") != "asc" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":        "wfnodetestrun_1",
					"workflow":  map[string]any{"workflow_id": "wf_1", "version_id": "draft_1", "name_at_run_time": "Workflow"},
					"trigger":   map[string]any{"type": "api"},
					"lifecycle": map[string]any{"status": "completed"},
					"timing":    map[string]any{"created_at": "2026-05-18T10:00:00Z", "started_at": "2026-05-18T10:00:01Z"},
					"test_id":   "wfnodetest_1",
				},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Workflows.Tests.Runs.List(context.Background(), ListWorkflowTestRunsParams{
		WorkflowID:    "wf_1",
		TestID:        "wfnodetest_1",
		TargetBlockID: "block_1",
		Status:        "passed",
		Statuses:      []string{"passed", "failed"},
		ExcludeStatus: "cancelled",
		TriggerType:   "api",
		TriggerTypes:  []string{"api", "manual_run"},
		FromDate:      "2026-05-01",
		ToDate:        "2026-05-18",
		SortBy:        "created_at",
		Fields:        []string{"id", "lifecycle"},
		Before:        "wftestrun_before",
		After:         "wftestrun_after",
		Limit:         10,
		Order:         "asc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("page.Data length = %d, want 1", len(page.Data))
	}
	if page.Data[0].ID != "wfnodetestrun_1" {
		t.Fatalf("page.Data[0].ID = %q, want wfnodetestrun_1", page.Data[0].ID)
	}
}

func TestWorkflowTestRunsCreateDecodesRunResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/tests/runs" || r.URL.RawQuery != "" {
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		var body Resource
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["workflow_id"] != "wf_1" || body["test_id"] != "wfnodetest_1" {
			t.Fatalf("unexpected request body %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "wftestrun_1",
			"workflow":    map[string]any{"workflow_id": "wf_1", "version_id": "draft_1", "name_at_run_time": "Workflow"},
			"trigger":     map[string]any{"type": "api"},
			"lifecycle":   map[string]any{"status": "pending"},
			"timing":      map[string]any{"created_at": "2026-05-18T10:00:00Z", "started_at": nil},
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

	result, err := client.Workflows.Tests.Runs.Create(context.Background(), CreateWorkflowTestRunRequest{
		WorkflowID: "wf_1",
		TestID:     "wfnodetest_1",
		NConsensus: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Lifecycle.Status != "pending" || result.Workflow.WorkflowID != "wf_1" || result.TestID != "wfnodetest_1" || result.TotalTests != 1 {
		t.Fatalf("create response lost fields: %#v", result)
	}
	if result.ID != "wftestrun_1" {
		t.Fatalf("run id = %q, want wftestrun_1", result.ID)
	}
	if result.Target == nil || (*result.Target)["block_id"] != "block_transform" {
		t.Fatalf("target = %#v, want block_transform", result.Target)
	}
	if result.Timing.StartedAt != nil {
		t.Fatalf("started_at = %v, want nil for pending run", result.Timing.StartedAt)
	}
}
