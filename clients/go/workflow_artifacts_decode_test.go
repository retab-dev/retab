package retab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Pins the dual-shape decode contract for `Workflows.Artifacts.List`:
//
//   - Production currently returns a bare JSON array `[...]` for this
//     endpoint, unlike every other list endpoint which returns the
//     paginated envelope `{"data": [...], "list_metadata": {...}}`.
//   - The SDK transparently handles both. Callers always receive a
//     `*PaginatedList[WorkflowArtifact]`.
//
// If a future refactor makes the SDK strict about the envelope shape,
// these tests fail immediately — a far better signal than a CLI user
// hitting "cannot unmarshal array into Go value of type retab.alias[...]"
// at runtime against a real workspace.

func TestArtifactsList_AcceptsBareArrayShape(t *testing.T) {
	body := `[
		{"operation": "extraction", "id": "art_1"},
		{"operation": "parse", "id": "art_2"}
	]`
	srv := newArtifactsListServer(t, body)
	defer srv.Close()

	client := newArtifactsTestClient(t, srv)
	got, err := client.Workflows.Artifacts.List(context.Background(),
		ListWorkflowArtifactsParams{RunID: "run_xyz"})
	if err != nil {
		t.Fatalf("List against bare-array response: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 artifacts, got %d: %#v", len(got.Data), got)
	}
	if op, _ := got.Data[0]["operation"].(string); op != "extraction" {
		t.Errorf("first artifact operation = %q, want %q", op, "extraction")
	}
	if op, _ := got.Data[1]["operation"].(string); op != "parse" {
		t.Errorf("second artifact operation = %q, want %q", op, "parse")
	}
}

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

// Empty array is a legitimate "no artifacts" response — must not error.
func TestArtifactsList_EmptyBareArray(t *testing.T) {
	srv := newArtifactsListServer(t, `[]`)
	defer srv.Close()
	client := newArtifactsTestClient(t, srv)
	got, err := client.Workflows.Artifacts.List(context.Background(),
		ListWorkflowArtifactsParams{RunID: "run_xyz"})
	if err != nil {
		t.Fatalf("empty array should decode cleanly: %v", err)
	}
	if len(got.Data) != 0 {
		t.Errorf("expected 0 items, got %d", len(got.Data))
	}
}

// Anything other than `[` or `{` is a protocol violation we surface
// clearly rather than letting it silently produce empty results.
func TestArtifactsList_RejectsUnexpectedShape(t *testing.T) {
	srv := newArtifactsListServer(t, `"unexpected string body"`)
	defer srv.Close()
	client := newArtifactsTestClient(t, srv)
	_, err := client.Workflows.Artifacts.List(context.Background(),
		ListWorkflowArtifactsParams{RunID: "run_xyz"})
	if err == nil {
		t.Fatal("expected error for non-array, non-envelope response")
	}
	if !strings.Contains(err.Error(), "unexpected artifact list response shape") {
		t.Errorf("error message %q doesn't describe the actual problem", err.Error())
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
