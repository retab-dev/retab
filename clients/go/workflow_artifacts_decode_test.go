package retab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Pins the live decode contract for `Workflows.Artifacts.List`. The server
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
	got, err := client.Workflows.Artifacts.List(context.Background(),
		ListWorkflowArtifactsParams{RunID: "run_xyz"})
	if err != nil {
		t.Fatalf("List against envelope-shape response: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0]["id"] != "art_3" {
		t.Errorf("got %#v, want one artifact with id art_3", got)
	}
}

// RunID is required — guards against an accidental unfiltered call.
// The check returns before any HTTP request, so the base URL is never
// dialled; a properly-constructed client (services wired up) is still
// needed because List() is a method on the Artifacts service.
func TestArtifactsList_RequiresRunID(t *testing.T) {
	client, err := NewClient("api_key_for_test", WithBaseURL("http://127.0.0.1:0/v1"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = client.Workflows.Artifacts.List(context.Background(),
		ListWorkflowArtifactsParams{})
	if err == nil || !strings.Contains(err.Error(), "runID is required") {
		t.Errorf("expected runID-required error, got %v", err)
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
	client, err := NewClient("api_key_for_test", WithBaseURL(srv.URL+"/v1"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}
