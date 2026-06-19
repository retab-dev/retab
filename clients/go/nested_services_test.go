package retab

import "testing"

func TestNewClientWiresNestedServices(t *testing.T) {
	client, err := NewClient("sk_test")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if client.Workflows.Runs == nil {
		t.Fatalf("expected client.Workflows.Runs to be wired")
	}
	if client.Workflows.Blocks == nil {
		t.Fatalf("expected client.Workflows.Blocks to be wired")
	}
	if client.Workflows.Blocks.Executions == nil {
		t.Fatalf("expected client.Workflows.Blocks.Executions to be wired")
	}
	if client.Workflows.Spec == nil {
		t.Fatalf("expected client.Workflows.Spec to be wired")
	}
	if client.Workflows.Evals == nil {
		t.Fatalf("expected client.Workflows.Evals to be wired")
	}
	if client.Workflows.Evals.Runs == nil {
		t.Fatalf("expected client.Workflows.Evals.Runs to be wired")
	}
	if client.Edits.Templates == nil {
		t.Fatalf("expected client.Edits.Templates to be wired")
	}
}
