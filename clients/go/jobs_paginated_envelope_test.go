package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestJobsListUsesListMetadataEnvelope pins the canonical
// `{"data": [...], "list_metadata": {"before": ..., "after": ...}}` envelope
// for `GET /v1/jobs`. The Go SDK previously declared a bespoke
// `JobListResponse{Data, FirstID, LastID, HasMore}` shape that drifted from
// the wire format every other list endpoint uses. Pin the migration so the
// jobs envelope stays aligned with the rest of the SDK and the backend.
func TestJobsListUsesListMetadataEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "job_1", "object": "job", "status": "completed"},
				{"id": "job_2", "object": "job", "status": "queued"},
			},
			"list_metadata": map[string]any{"before": "job_1", "after": "job_2"},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Jobs.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("Jobs.List: %v", err)
	}
	if got := len(page.Data); got != 2 {
		t.Fatalf("page.Data length = %d, want 2", got)
	}
	if page.Data[0].ID != "job_1" || page.Data[1].ID != "job_2" {
		t.Fatalf("page.Data ids = [%v, %v]", page.Data[0].ID, page.Data[1].ID)
	}
	if page.ListMetadata.Before != "job_1" || page.ListMetadata.After != "job_2" {
		t.Fatalf("page.ListMetadata = %+v, want {Before:\"job_1\", After:\"job_2\"}", page.ListMetadata)
	}
}

// Empty-page response must yield a non-nil, length-0 Data slice (ranging over
// the result without a nil check is a common pattern in the CLI).
func TestJobsListEmptyPageYieldsNonNilSlice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []map[string]any{},
			"list_metadata": map[string]any{"before": nil, "after": nil},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	page, err := client.Jobs.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("Jobs.List: %v", err)
	}
	if page.Data == nil {
		t.Fatalf("page.Data must be non-nil empty slice")
	}
	if len(page.Data) != 0 {
		t.Fatalf("page.Data = %#v, want empty slice", page.Data)
	}
	if page.ListMetadata.Before != "" || page.ListMetadata.After != "" {
		t.Fatalf("page.ListMetadata = %+v, want zero cursors", page.ListMetadata)
	}
}
