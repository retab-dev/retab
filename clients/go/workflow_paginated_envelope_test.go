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

	page, err := client.Workflows.Experiments.List(context.Background(), &WorkflowExperimentsListParams{WorkflowID: "wf_1"})
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
			query.Get("exclude_status") != "cancelled" ||
			query.Get("trigger_type") != "api" ||
			query.Get("from_date") != "2026-05-01" ||
			query.Get("to_date") != "2026-05-18" ||
			query.Get("sort_by") != "created_at" ||
			query.Has("fields") ||
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

	completedStatus := WorkflowExperimentsStatus(LatestBlockEvalRunSummaryStatusCompleted)
	cancelledStatus := WorkflowExperimentsExcludeStatus(LatestBlockEvalRunSummaryStatusCancelled)
	limit := 10
	page, err := client.Workflows.Experiments.Runs.List(context.Background(), &ExperimentRunsListParams{
		WorkflowID:    ptrTo("wf_1"),
		ExperimentID:  ptrTo("exp_1"),
		BlockID:       ptrTo("block_1"),
		Status:        &completedStatus,
		ExcludeStatus: &cancelledStatus,
		TriggerType:   ptrTo("api"),
		FromDate:      ptrTo("2026-05-01"),
		ToDate:        ptrTo("2026-05-18"),
		SortBy:        ptrTo("created_at"),
		PaginationParams: PaginationParams{
			Before: ptrTo("exprun_before"),
			After:  ptrTo("exprun_after"),
			Limit:  &limit,
			Order:  ptrTo("asc"),
		},
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

// TestWorkflowEvalsListUsesPaginatedEnvelope pins the canonical envelope for
// `GET /v1/workflows/evals?workflow_id={wf}` (was `{"tests": [...]}`).
func TestWorkflowEvalsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "wfnodeeval_1", "workflow_id": "wf_1"},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Workflows.Evals.List(context.Background(), &WorkflowEvalsListParams{WorkflowID: "wf_1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("page.Data length = %d, want 1", len(page.Data))
	}
	if page.Data[0].ID != "wfnodeeval_1" {
		t.Fatalf("page.Data[0].ID = %v, want wfnodeeval_1", page.Data[0].ID)
	}
}

// TestWorkflowEvalRunsListUsesPaginatedEnvelope pins the canonical envelope
// for `GET /v1/workflows/evals/runs`.
func TestWorkflowEvalRunsListUsesPaginatedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/evals/runs" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("workflow_id") != "wf_1" ||
			query.Get("eval_id") != "wfnodeeval_1" ||
			query.Get("target_block_id") != "block_1" ||
			query.Get("status") != "passed" ||
			query.Get("exclude_status") != "cancelled" ||
			query.Get("trigger_type") != "api" ||
			query.Get("sort_by") != "created_at" ||
			query.Has("fields") ||
			query.Get("before") != "wfevalrun_before" ||
			query.Get("after") != "wfevalrun_after" ||
			query.Get("limit") != "10" ||
			query.Get("order") != "asc" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":                  "wfnodeevalrun_1",
					"workflow_id":         "wf_1",
					"workflow_version_id": "draft_1",
					"trigger":             map[string]any{"type": "api"},
					"lifecycle":           map[string]any{"status": "completed"},
					"timing":              map[string]any{"created_at": "2026-05-18T10:00:00Z", "started_at": "2026-05-18T10:00:01Z"},
					"eval_id":             "wfnodeeval_1",
				},
			},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	limit := 10
	page, err := client.Workflows.Evals.Runs.List(context.Background(), &WorkflowEvalRunsListParams{
		WorkflowID:    ptrTo("wf_1"),
		EvalID:        ptrTo("wfnodeeval_1"),
		TargetBlockID: ptrTo("block_1"),
		Status:        ptrTo("passed"),
		ExcludeStatus: ptrTo("cancelled"),
		TriggerType:   ptrTo("api"),
		SortBy:        ptrTo("created_at"),
		PaginationParams: PaginationParams{
			Before: ptrTo("wfevalrun_before"),
			After:  ptrTo("wfevalrun_after"),
			Limit:  &limit,
			Order:  ptrTo("asc"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("page.Data length = %d, want 1", len(page.Data))
	}
	if page.Data[0].ID != "wfnodeevalrun_1" {
		t.Fatalf("page.Data[0].ID = %q, want wfnodeevalrun_1", page.Data[0].ID)
	}
}

func TestWorkflowEvalRunsCreateDecodesRunResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/evals/runs" || r.URL.RawQuery != "" {
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		var body Resource
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		scope, ok := body["scope"].(map[string]any)
		if !ok {
			t.Fatalf("scope = %#v", body["scope"])
		}
		if body["workflow_id"] != "wf_1" || scope["type"] != "single" || scope["eval_id"] != "wfnodeeval_1" {
			t.Fatalf("unexpected request body %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                  "wfevalrun_1",
			"workflow_id":         "wf_1",
			"workflow_version_id": "draft_1",
			"trigger":             map[string]any{"type": "api"},
			"lifecycle":           map[string]any{"status": "pending"},
			"timing":              map[string]any{"created_at": "2026-05-18T10:00:00Z", "started_at": nil},
			"eval_id":             "wfnodeeval_1",
			"total_evals":         1,
			"target": map[string]any{
				"type":     "block",
				"block_id": "block_transform",
			},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	evalID := "wfnodeeval_1"
	result, err := client.Workflows.Evals.Runs.Create(context.Background(), &WorkflowEvalRunsCreateParams{
		WorkflowID: "wf_1",
		Scope: &WorkflowEvalRunScope{
			Type:   WorkflowEvalRunScopeTypeSingle,
			EvalID: &evalID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Lifecycle.Status() != "pending" || result.WorkflowID != "wf_1" || result.EvalID == nil || *result.EvalID != "wfnodeeval_1" || result.TotalEvals != 1 {
		t.Fatalf("create response lost fields: %#v", result)
	}
	if result.ID != "wfevalrun_1" {
		t.Fatalf("run id = %q, want wfevalrun_1", result.ID)
	}
	if result.Target == nil || result.Target.BlockID != "block_transform" {
		t.Fatalf("target = %#v, want block_transform", result.Target)
	}
	if result.Timing.StartedAt != nil {
		t.Fatalf("started_at = %v, want nil for pending run", result.Timing.StartedAt)
	}
}
