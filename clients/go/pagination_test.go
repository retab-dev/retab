package retab

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Two-page server: page 1 has after="cursor-2"; page 2 has after="".
// Each list method takes the `after` query param.
func newTwoPageWorkflowsServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var seenAfter []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/workflows" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		after := r.URL.Query().Get("after")
		seenAfter = append(seenAfter, after)
		switch after {
		case "":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "wf_1", "name": "first"},
					{"id": "wf_2", "name": "second"},
				},
				"list_metadata": map[string]any{"before": "", "after": "cursor-2"},
			})
		case "cursor-2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "wf_3", "name": "third"},
					{"id": "wf_4", "name": "fourth"},
				},
				"list_metadata": map[string]any{"before": "", "after": ""},
			})
		default:
			t.Fatalf("unexpected after cursor: %q", after)
		}
	}))
	return server, &seenAfter
}

func TestAutoPagingWalksEveryItemAcrossPages(t *testing.T) {
	server, seenAfter := newTwoPageWorkflowsServer(t)
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	page, err := client.Workflows.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if len(page.Data) != 2 {
		t.Fatalf("expected 2 items on first page, got %d", len(page.Data))
	}
	if !page.HasNextPage() {
		t.Fatalf("expected HasNextPage true after first page; got false (after=%q)", page.ListMetadata.After)
	}

	var ids []string
	if err := page.AutoPaging(context.Background(), func(w Workflow) error {
		ids = append(ids, w.ID)
		return nil
	}); err != nil {
		t.Fatalf("AutoPaging: %v", err)
	}
	if len(ids) != 4 {
		t.Fatalf("expected 4 items across both pages, got %d: %v", len(ids), ids)
	}
	want := []string{"wf_1", "wf_2", "wf_3", "wf_4"}
	for i, w := range want {
		if ids[i] != w {
			t.Fatalf("ids[%d] = %q, want %q", i, ids[i], w)
		}
	}
	// Two requests: initial (after=""), then auto-paging (after="cursor-2").
	if len(*seenAfter) != 2 {
		t.Fatalf("expected 2 GETs to /v1/workflows, got %d (%v)", len(*seenAfter), *seenAfter)
	}
	if (*seenAfter)[0] != "" || (*seenAfter)[1] != "cursor-2" {
		t.Fatalf("server saw after cursors %v, want [\"\", \"cursor-2\"]", *seenAfter)
	}
}

func TestAutoPagingShortCircuitsOnYieldError(t *testing.T) {
	server, seenAfter := newTwoPageWorkflowsServer(t)
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	page, err := client.Workflows.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	sentinel := errors.New("stop iteration")
	count := 0
	err = page.AutoPaging(context.Background(), func(w Workflow) error {
		count++
		if w.ID == "wf_2" {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("AutoPaging error = %v, want sentinel", err)
	}
	if count != 2 {
		t.Fatalf("expected to short-circuit after 2 items, got %d", count)
	}
	// Should NOT have fetched the second page.
	if len(*seenAfter) != 1 {
		t.Fatalf("expected exactly 1 GET (no second page fetch), got %d (%v)", len(*seenAfter), *seenAfter)
	}
}

func TestAutoPagingStopsWhenNoMorePages(t *testing.T) {
	// Single-page server: after="" in response means no more pages.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "wf_only", "name": "only one"},
			},
			"list_metadata": map[string]any{"before": "", "after": ""},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.Workflows.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if page.HasNextPage() {
		t.Fatalf("HasNextPage should be false when after is empty; got true")
	}
	next, err := page.NextPage(context.Background())
	if err != nil {
		t.Fatalf("NextPage error on terminal page = %v, want nil", err)
	}
	if next != nil {
		t.Fatalf("NextPage on terminal page should return nil, got %#v", next)
	}

	var ids []string
	if err := page.AutoPaging(context.Background(), func(w Workflow) error {
		ids = append(ids, w.ID)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "wf_only" {
		t.Fatalf("ids = %v, want [wf_only]", ids)
	}
}

// AutoPaging when fetchNext is nil (e.g. a PaginatedList built by hand or
// unmarshalled directly) should still iterate the current page items.
func TestAutoPagingWithoutFetchNextIteratesCurrentPageOnly(t *testing.T) {
	page := &PaginatedList[Workflow]{
		Data: []Workflow{{ID: "wf_a"}, {ID: "wf_b"}},
		// ListMetadata.After is "next", but fetchNext is nil so we can't fetch.
		ListMetadata: PaginationCursor{After: "next"},
	}
	var ids []string
	if err := page.AutoPaging(context.Background(), func(w Workflow) error {
		ids = append(ids, w.ID)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "wf_a" || ids[1] != "wf_b" {
		t.Fatalf("ids = %v, want [wf_a wf_b]", ids)
	}
}

// NextPage on a paginated list that *does* have a next-page cursor should
// fetch the second page using fetchNext.
func TestNextPageFetchesUsingCursor(t *testing.T) {
	server, _ := newTwoPageWorkflowsServer(t)
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.Workflows.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	next, err := page.NextPage(context.Background())
	if err != nil {
		t.Fatalf("NextPage: %v", err)
	}
	if next == nil {
		t.Fatal("NextPage returned nil for a page with HasNextPage=true")
	}
	if len(next.Data) != 2 || next.Data[0].ID != "wf_3" || next.Data[1].ID != "wf_4" {
		t.Fatalf("next page data = %#v", next.Data)
	}
	if next.HasNextPage() {
		t.Fatalf("last page should not have a next; after = %q", next.ListMetadata.After)
	}
}
