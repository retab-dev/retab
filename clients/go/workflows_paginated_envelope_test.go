package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWorkflowGraphListsReturnPaginatedEnvelope is the regression guard for the
// shape conversion of the seven workflow GRAPH list endpoints from bare
// arrays to PaginatedList[T] envelopes.
//
//   GET /v1/workflows/{wf}/blocks                                    -> PaginatedList[WorkflowBlock]
//   GET /v1/workflows/{wf}/blocks/{id}/config-history                -> PaginatedList[BlockConfigVersion]
//   GET /v1/workflows/{wf}/edges                                     -> PaginatedList[WorkflowEdgeDoc]
//   GET /v1/workflows/artifacts                                      -> PaginatedList[WorkflowArtifact]
//   GET /v1/workflows/{wf}/snapshots                                 -> PaginatedList[WorkflowSnapshot]
//   GET /v1/workflows/runs/{run}/steps                               -> PaginatedList[WorkflowRunStep]
//   GET /v1/workflows/runs/{run}/steps/{block}/simulations           -> PaginatedList[BlockSimulation]
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
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_aaa/blocks":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"id":          "extract-1",
				"workflow_id": "wf_aaa",
				"type":        "extract",
				"label":       "Extract",
				"position_x":  0,
				"position_y":  0,
			}))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_aaa/blocks/extract-1/config-history":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"config_fingerprint": "fp_1",
				"block_type":         "extract",
				"block_label":        "Extract",
				"first_seen_at":      "2026-01-01T00:00:00Z",
				"last_seen_at":       "2026-01-02T00:00:00Z",
				"snapshot_versions":  []int{1, 2},
				"run_count":          7,
				"is_current":         true,
			}))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_aaa/edges":
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
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_aaa/snapshots":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"id":           "wfsn_1",
				"snapshot_id":  "snap_1",
				"workflow_id":  "wf_aaa",
				"version":      3,
				"description":  "third release",
				"block_count":  4,
				"edge_count":   3,
				"published_at": "2026-04-01T12:00:00Z",
			}))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_aaa/steps":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"run_id":          "run_aaa",
				"organization_id": "org_1",
				"block_id":        "extract-1",
				"step_id":         "extract-1",
				"block_type":      "extract",
				"block_label":     "Extract",
				"lifecycle":       map[string]any{"status": "completed"},
			}))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_aaa/steps/extract-1/simulations":
			_ = json.NewEncoder(w).Encode(envelope(map[string]any{
				"id":          "sim_1",
				"workflow_id": "wf_aaa",
				"run_id":      "run_aaa",
				"block_id":    "extract-1",
				"block_type":  "extract",
				"success":     true,
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

	blocks, err := client.Workflows.Blocks.List(ctx, "wf_aaa")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks.Data) != 1 || blocks.Data[0].ID != "extract-1" {
		t.Fatalf("blocks = %#v", blocks)
	}
	if blocks.ListMetadata.Before != "" || blocks.ListMetadata.After != "" {
		t.Fatalf("blocks.ListMetadata = %#v", blocks.ListMetadata)
	}

	history, err := client.Workflows.Blocks.ConfigHistory(ctx, "wf_aaa", "extract-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Data) != 1 || history.Data[0].ConfigFingerprint != "fp_1" || history.Data[0].RunCount != 7 {
		t.Fatalf("config history = %#v", history)
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

	snaps, err := client.Workflows.ListSnapshots(ctx, "wf_aaa", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps.Data) != 1 || snaps.Data[0].SnapshotID != "snap_1" || snaps.Data[0].Version != 3 {
		t.Fatalf("snapshots = %#v", snaps)
	}

	steps, err := client.Workflows.Runs.Steps.List(ctx, "run_aaa")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps.Data) != 1 || steps.Data[0].BlockID != "extract-1" {
		t.Fatalf("steps = %#v", steps)
	}

	sims, err := client.Workflows.Blocks.ListSimulations(ctx, "run_aaa", "extract-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sims.Data) != 1 || sims.Data[0].ID != "sim_1" || !sims.Data[0].Success {
		t.Fatalf("simulations = %#v", sims)
	}
}
