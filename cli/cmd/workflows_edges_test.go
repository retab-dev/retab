package cmd

import "testing"

func TestParseEdgeCreateGeneratesIDWhenMissing(t *testing.T) {
	req := parseEdgeCreate(map[string]any{
		"source_block": "start_1",
		"target_block": "extract_1",
	})

	if req.ID == "" {
		t.Fatal("expected parseEdgeCreate to generate an id when none is provided")
	}
}

func TestParseEdgeCreatePreservesExplicitID(t *testing.T) {
	req := parseEdgeCreate(map[string]any{
		"id":           "edge_custom",
		"source_block": "start_1",
		"target_block": "extract_1",
	})

	if req.ID != "edge_custom" {
		t.Fatalf("id = %q, want edge_custom", req.ID)
	}
}
