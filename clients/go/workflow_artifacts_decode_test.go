package retab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Pins the live decode contract for `WorkflowArtifacts.List`. The server
// emits the canonical {data, list_metadata} envelope; PaginatedList[T] is
// the typed wrapper callers consume.

func TestArtifactsList_AcceptsEnvelopeShape(t *testing.T) {
	body := `{
		"data": [{"operation": "edit", "id": "art_3"}],
		"list_metadata": {"before": null, "after": null}
	}`
	srv := newArtifactsListServer(t, body)
	defer srv.Close()

	client := newArtifactsTestClient(t, srv)
	runID := "run_xyz"
	got, err := client.Workflows.Artifacts.List(context.Background(),
		&WorkflowArtifactsListParams{RunID: &runID})
	if err != nil {
		t.Fatalf("List against envelope-shape response: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].ID != "art_3" {
		t.Errorf("got %#v, want one artifact with id art_3", got)
	}
}

func TestArtifactsList_PreservesOperationSpecificFields(t *testing.T) {
	body := `{
		"data": [{
			"operation": "function_invocation",
			"id": "fninv_123",
			"run_id": "run_xyz",
			"step_id": "step_function",
			"block_id": "block_function",
			"output": {"ok": true}
		}],
		"list_metadata": {"before": null, "after": null}
	}`
	srv := newArtifactsListServer(t, body)
	defer srv.Close()

	client := newArtifactsTestClient(t, srv)
	runID := "run_xyz"
	got, err := client.Workflows.Artifacts.List(context.Background(),
		&WorkflowArtifactsListParams{RunID: &runID})
	if err != nil {
		t.Fatalf("List against artifact response: %v", err)
	}
	if len(got.Data) != 1 {
		t.Fatalf("data len = %d, want 1", len(got.Data))
	}
	extras := got.Data[0].AdditionalProperties
	if extras["step_id"] != "step_function" || extras["block_id"] != "block_function" {
		t.Fatalf("artifact extras = %#v, want provenance fields preserved", extras)
	}
	output, ok := extras["output"].(map[string]interface{})
	if !ok || output["ok"] != true {
		t.Fatalf("artifact output = %#v, want output.ok=true", extras["output"])
	}
}

// ---- helpers ----

func newArtifactsListServer(t *testing.T, jsonBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/workflows/artifacts" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("run_id") != "run_xyz" {
			t.Errorf("missing or wrong run_id: %q", r.URL.Query().Get("run_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(jsonBody))
	}))
}

// newArtifactsTestClient builds a Client pointed at the test server. The
// helper is name-scoped to this file (vs. a generic `newTestClient`) so
// it doesn't collide with the same-named helper in resources_test.go.
func newArtifactsTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	client, err := NewClient("api_key_for_test", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}
