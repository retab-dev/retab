package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWorkflowGraphListsReturnPaginatedEnvelope is the regression guard for the
// shape conversion of the four workflow GRAPH list endpoints from bare
// arrays to PaginatedList[T] envelopes.
//
//	GET /v1/workflows/blocks?workflow_id={wf}                                    -> PaginatedList[WorkflowBlock]
//	GET /v1/workflows/edges?workflow_id={wf}                                     -> PaginatedList[WorkflowEdgeDoc]
//	GET /v1/workflows/artifacts                                      -> PaginatedList[WorkflowArtifact]
//	GET /v1/workflows/steps?run_id={run}                              -> PaginatedList[WorkflowRunStep]
func TestWorkflowGraphListsReturnPaginatedEnvelope(t *testing.T) {
	envelope := func(items ...map[string]any) map[string]any {
		return map[string]any{
			"data":          items,
			"list_metadata": map[string]any{"before": nil, "after": nil},
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/blocks" && r.URL.Query().Get("workflow_id") == "wf_aaa":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"id":          "extract-1",
				"workflow_id": "wf_aaa",
				"type":        "extract",
				"label":       "Extract",
				"position_x":  0,
				"position_y":  0,
			}))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/edges" && r.URL.Query().Get("workflow_id") == "wf_aaa":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"id":              "edge-1",
				"workflow_id":     "wf_aaa",
				"organization_id": "org_1",
				"source_block":    "start-1",
				"target_block":    "extract-1",
				"source_handle":   "output-file-0",
				"target_handle":   "input-file-0",
			}))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/artifacts":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"operation": "extraction",
				"id":        "ext_123",
			}))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/steps" && r.URL.Query().Get("run_id") == "run_aaa":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"run_id":          "run_aaa",
				"organization_id": "org_1",
				"block_id":        "extract-1",
				"step_id":         "extract-1",
				"block_type":      "extract",
				"block_label":     "Extract",
				"lifecycle":       map[string]any{"status": "completed"},
			}))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	blocks, err := client.Workflows.Blocks.List(ctx, "wf_aaa", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks.Data) != 1 || blocks.Data[0].ID != "extract-1" {
		t.Fatalf("blocks = %#v", blocks)
	}
	if blocks.ListMetadata.Before != "" || blocks.ListMetadata.After != "" {
		t.Fatalf("blocks.ListMetadata = %#v", blocks.ListMetadata)
	}

	edges, err := client.Workflows.Edges.List(ctx, "wf_aaa", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges.Data) != 1 || edges.Data[0].ID != "edge-1" {
		t.Fatalf("edges = %#v", edges)
	}

	artifacts, err := client.Workflows.Artifacts.List(ctx, ListWorkflowArtifactsParams{RunID: "run_aaa"})
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts.Data) != 1 || artifacts.Data[0]["id"] != "ext_123" {
		t.Fatalf("artifacts = %#v", artifacts)
	}

	steps, err := client.Workflows.Steps.List(ctx, "run_aaa", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps.Data) != 1 || steps.Data[0].BlockID != "extract-1" {
		t.Fatalf("steps = %#v", steps)
	}
}
