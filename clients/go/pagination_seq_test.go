//go:build go1.23

package retab

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAutoPagingSeqWalksEveryItemAcrossPages(t *testing.T) {
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

	var ids []string
	for w, err := range page.AutoPagingSeq(context.Background()) {
		if err != nil {
			t.Fatalf("unexpected err mid-iteration: %v", err)
		}
		ids = append(ids, w.ID)
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
		t.Fatalf("expected 2 server hits, got %d: %v", len(*seenAfter), *seenAfter)
	}
}

func TestAutoPagingSeqBreaksCleanlyWithoutFetchingNextPage(t *testing.T) {
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

	var ids []string
	for w, err := range page.AutoPagingSeq(context.Background()) {
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		ids = append(ids, w.ID)
		if len(ids) == 1 {
			break // caller short-circuits after first item
		}
	}

	if len(ids) != 1 {
		t.Fatalf("expected 1 item after early break, got %d", len(ids))
	}
	// Only the first page should have been fetched — break must NOT trigger
	// the second-page fetch the underlying fetchNext closure would otherwise do.
	if len(*seenAfter) != 1 {
		t.Fatalf("expected 1 server hit (no second page after break), got %d: %v",
			len(*seenAfter), *seenAfter)
	}
}

func TestAutoPagingSeqYieldsErrorOnFetchFailure(t *testing.T) {
	// First page succeeds with cursor pointing at a second page that 500s.
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"data":[{"id":"wf_1","name":"first"}],"list_metadata":{"before":"","after":"cursor-2"}}`)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"detail":"boom"}`)
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	page, err := client.Workflows.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("first page should succeed: %v", err)
	}

	var ids []string
	var seenErr error
	for w, err := range page.AutoPagingSeq(context.Background()) {
		if err != nil {
			seenErr = err
			break
		}
		ids = append(ids, w.ID)
	}

	if seenErr == nil {
		t.Fatal("expected an error to be yielded after the second page fetch failed")
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 item from the first page before the error, got %d", len(ids))
	}
	// Exactly two server calls: the initial list + the failing second-page fetch.
	if calls != 2 {
		t.Fatalf("expected 2 server calls, got %d", calls)
	}
}
