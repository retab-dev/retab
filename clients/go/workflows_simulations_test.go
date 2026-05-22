package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkflowSimulationsCreateAndListUseCanonicalRoutes(t *testing.T) {
	var createBody map[string]any
	var sawCreate bool
	var sawList bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/simulations" && r.URL.RawQuery == "":
			sawCreate = true
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(blockSimulationResponse("sim_123"))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/simulations":
			sawList = true
			query := r.URL.Query()
			if query.Get("run_id") != "run_123" || query.Get("block_id") != "blk_extract" || query.Get("limit") != "10" {
				t.Fatalf("unexpected list query: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []map[string]any{blockSimulationResponse("sim_123")},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		default:
			t.Fatalf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	checkEligibility := false
	created, err := client.Workflows.Simulations.Create(context.Background(), CreateWorkflowSimulationRequest{
		RunID:            "run_123",
		BlockID:          "blk_extract",
		StepID:           "step_iter_0_blk_extract",
		NConsensus:       5,
		CheckEligibility: &checkEligibility,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "sim_123" || created.BlockID != "blk_extract" {
		t.Fatalf("created = %#v", created)
	}
	if createBody["run_id"] != "run_123" || createBody["block_id"] != "blk_extract" {
		t.Fatalf("create body = %#v", createBody)
	}
	if createBody["step_id"] != "step_iter_0_blk_extract" || createBody["n_consensus"] != float64(5) || createBody["check_eligibility"] != false {
		t.Fatalf("create body = %#v", createBody)
	}
	if _, ok := createBody["source_step_id"]; ok {
		t.Fatalf("create body = %#v", createBody)
	}

	list, err := client.Workflows.Simulations.List(context.Background(), ListWorkflowSimulationsParams{
		RunID:   "run_123",
		BlockID: "blk_extract",
		Limit:   10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 || list.Data[0].ID != "sim_123" {
		t.Fatalf("list = %#v", list)
	}
	if !sawCreate || !sawList {
		t.Fatalf("sawCreate=%v sawList=%v", sawCreate, sawList)
	}
}

func TestWorkflowSimulationsRequireRunAndBlockIDs(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Workflows.Simulations.Create(context.Background(), CreateWorkflowSimulationRequest{BlockID: "blk_1"}); err == nil {
		t.Fatal("expected create without run id to fail")
	}
	if _, err := client.Workflows.Simulations.Create(context.Background(), CreateWorkflowSimulationRequest{RunID: "run_1"}); err == nil {
		t.Fatal("expected create without block id to fail")
	}
	if _, err := client.Workflows.Simulations.List(context.Background(), ListWorkflowSimulationsParams{BlockID: "blk_1"}); err == nil {
		t.Fatal("expected list without run id to fail")
	}
	if _, err := client.Workflows.Simulations.List(context.Background(), ListWorkflowSimulationsParams{RunID: "run_1"}); err == nil {
		t.Fatal("expected list without block id to fail")
	}
}

func blockSimulationResponse(id string) map[string]any {
	return map[string]any{
		"id":             id,
		"workflow_id":    "wf_123",
		"run_id":         "run_123",
		"block_id":       "blk_extract",
		"block_type":     "extract",
		"lifecycle":      map[string]any{"status": "completed"},
		"handle_inputs":  map[string]any{},
		"handle_outputs": map[string]any{"output-json-0": map[string]any{"type": "json", "data": map[string]any{"vendor": "Acme"}}},
		"created_at":     "2026-05-21T10:00:00Z",
	}
}
