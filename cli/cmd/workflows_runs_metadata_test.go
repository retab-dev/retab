//go:build !retab_oagen_cli_workflows_runs

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

// TestWorkflowsRunsRestartCarriesSourceRunMetadata pins that a restart inherits
// the source run's user-defined metadata. Restart used to compose the create
// body from the source run's inputs only, so the retry landed with no metadata
// at all — which silently hid it from `runs list --metadata k=v`, the exact
// query used to trace a failing run together with its retries.
func TestWorkflowsRunsRestartCarriesSourceRunMetadata(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs/run_src":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                  "run_src",
				"workflow_id":         "wf_1",
				"workflow_version_id": "ver_old",
				"trigger":             map[string]any{"type": "manual"},
				"lifecycle":           map[string]any{"status": "error"},
				"metadata":            map[string]any{"suite": "nightly", "tenant": "acme"},
				"inputs": map[string]any{
					"json_data": map[string]any{"start_json": map[string]any{"n": 1}},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                  "run_new",
				"workflow_id":         "wf_1",
				"workflow_version_id": "ver_new",
				"trigger":             map[string]any{"type": "restart"},
				"lifecycle":           map[string]any{"status": "running"},
				"timing":              map[string]any{"created_at": "2026-07-25T00:00:00Z"},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsRestartCmd, "metadata") })

	_, _ = captureStd(t, func() {
		if err := workflowsRunsRestartCmd.RunE(workflowsRunsRestartCmd, []string{"run_src"}); err != nil {
			t.Fatalf("runs restart: %v", err)
		}
	})

	metadata, ok := body["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("restart dropped the source run's metadata: body = %#v", body)
	}
	if metadata["suite"] != "nightly" || metadata["tenant"] != "acme" {
		t.Fatalf("metadata = %#v, want the source run's pairs carried over", metadata)
	}
}

// TestWorkflowsRunsRestartMetadataFlagLayersOverInherited pins that
// `--metadata k=v` on restart overrides one inherited key without erasing the
// rest — a retry can be retagged (attempt=2) while staying findable by the tags
// the original run carried.
func TestWorkflowsRunsRestartMetadataFlagLayersOverInherited(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs/run_src":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                  "run_src",
				"workflow_id":         "wf_1",
				"workflow_version_id": "ver_old",
				"trigger":             map[string]any{"type": "manual"},
				"lifecycle":           map[string]any{"status": "error"},
				"metadata":            map[string]any{"suite": "nightly", "attempt": "1"},
				"inputs": map[string]any{
					"json_data": map[string]any{"start_json": map[string]any{"n": 1}},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                  "run_new",
				"workflow_id":         "wf_1",
				"workflow_version_id": "ver_new",
				"trigger":             map[string]any{"type": "restart"},
				"lifecycle":           map[string]any{"status": "running"},
				"timing":              map[string]any{"created_at": "2026-07-25T00:00:00Z"},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsRunsRestartCmd.Flags().Set("metadata", "attempt=2"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resetWorkflowRunsFlag(t, workflowsRunsRestartCmd, "metadata") })

	_, _ = captureStd(t, func() {
		if err := workflowsRunsRestartCmd.RunE(workflowsRunsRestartCmd, []string{"run_src"}); err != nil {
			t.Fatalf("runs restart: %v", err)
		}
	})

	metadata, ok := body["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata missing from create body: %#v", body)
	}
	if metadata["attempt"] != "2" {
		t.Fatalf("--metadata attempt=2 did not override the inherited value: %#v", metadata)
	}
	if metadata["suite"] != "nightly" {
		t.Fatalf("--metadata erased an unnamed inherited key: %#v", metadata)
	}
}
