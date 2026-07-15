//go:build !retab_oagen_cli_workflows_runs

package cmd

import "testing"

// TestWorkflowRunCreateRequestBodyIncludesMetadata proves the --metadata
// key=value pairs reach the request body as a metadata object (parity with the
// extraction/primitive `metadata` field).
func TestWorkflowRunCreateRequestBodyIncludesMetadata(t *testing.T) {
	body, err := workflowRunCreateRequestBody(workflowRunCreateParams{
		WorkflowID: "wf_1",
		Metadata:   map[string]string{"customer": "acme", "env": "prod"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	metadata, ok := body["metadata"].(map[string]string)
	if !ok {
		t.Fatalf("body metadata type = %T (%v)", body["metadata"], body["metadata"])
	}
	if metadata["customer"] != "acme" || metadata["env"] != "prod" {
		t.Fatalf("body metadata = %+v", metadata)
	}
}

// TestWorkflowRunCreateRequestBodyOmitsEmptyMetadata keeps the body clean when
// no --metadata flags are supplied.
func TestWorkflowRunCreateRequestBodyOmitsEmptyMetadata(t *testing.T) {
	body, err := workflowRunCreateRequestBody(workflowRunCreateParams{WorkflowID: "wf_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := body["metadata"]; present {
		t.Fatalf("body should omit metadata when absent, got %v", body["metadata"])
	}
}
