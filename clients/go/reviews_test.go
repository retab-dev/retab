package retab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const reviewID = "rev_1"
const reviewVersionID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const reviewChildVersionID = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func reviewVersionJSON(versionID string, parentID any) map[string]any {
	return map[string]any{
		"id":         versionID,
		"review_id":  reviewID,
		"parent_id":  parentID,
		"author":     map[string]any{"kind": "model", "id": "m", "display_name": "Model"},
		"snapshot":   map[string]any{"output": map[string]any{"total": 100, "currency": "USD"}},
		"note":       nil,
		"created_at": "2026-05-18T09:00:00Z",
	}
}

func reviewJSON(decided bool) map[string]any {
	decision := any(nil)
	if decided {
		decision = map[string]any{
			"verdict":    "approved",
			"version_id": reviewVersionID,
			"author":     map[string]any{"kind": "human", "id": "user_1", "display_name": "Ada"},
			"decided_at": "2026-05-18T09:05:00Z",
			"reason":     nil,
		}
	}
	return map[string]any{
		"id":                  reviewID,
		"workflow_id":         "wf_1",
		"workflow_version_id": "wfv_1",
		"workflow_run_id":     "run_1",
		"block_id":            "blk_1",
		"step_id":             "step_1",
		"parent_step_id":      nil,
		"iteration_key":       nil,
		"block_type":          "extract",
		"triggered_by":        map[string]any{"kind": "any_required_field_null"},
		"created_at":          "2026-05-18T09:00:00Z",
		"decision":            decision,
	}
}

func reviewQueueRowJSON() map[string]any {
	return map[string]any{
		"id":                  reviewID,
		"workflow_id":         "wf_1",
		"workflow_version_id": "wfv_1",
		"workflow_run_id":     "run_1",
		"block_id":            "blk_1",
		"step_id":             "step_1",
		"parent_step_id":      nil,
		"iteration_key":       nil,
		"block_type":          "extract",
		"triggered_by":        map[string]any{"kind": "any_required_field_null"},
		"created_at":          "2026-05-18T09:00:00Z",
		"decision":            nil,
	}
}

func TestWorkflowReviewsListUsesHardCutoverFilters(t *testing.T) {
	var seenMethod, seenPath, seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod, seenPath, seenQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []any{reviewQueueRowJSON()},
			"list_metadata": map[string]any{"before": nil, "after": reviewID},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.List(context.Background(), &ListReviewsParams{
		WorkflowID: "wf_1", RunID: "run_1", BlockID: "blk_1", StepID: "step_1", IterationKey: "0",
		Limit: 25, DecisionStatus: "decided",
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenMethod != http.MethodGet || seenPath != "/workflows/reviews" {
		t.Fatalf("request = %s %s", seenMethod, seenPath)
	}
	for _, want := range []string{
		"workflow_id=wf_1",
		"run_id=run_1",
		"block_id=blk_1", "step_id=step_1",
		"iteration_key=0", "limit=25", "decision_status=decided",
	} {
		if !strings.Contains(seenQuery, want) {
			t.Fatalf("query %q missing %q", seenQuery, want)
		}
	}
	if strings.Contains(seenQuery, "workflow_run_id=run_1") {
		t.Fatalf("query carries deprecated workflow_run_id param: %q", seenQuery)
	}
	if strings.Contains(seenQuery, "decision=") {
		t.Fatalf("query includes removed decision filter: %q", seenQuery)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != reviewID {
		t.Fatalf("resp = %#v", resp)
	}
	if resp.ListMetadata.After != reviewID {
		t.Fatalf("after = %q", resp.ListMetadata.After)
	}
}

func TestWorkflowReviewsListDefaultsToPending(t *testing.T) {
	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "list_metadata": map[string]any{"before": nil, "after": nil}})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(seenQuery, "decision_status=pending") {
		t.Fatalf("default query = %q", seenQuery)
	}
}

func TestWorkflowReviewsListRejectsBeforeAndAfterTogether(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://example.invalid"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Workflows.Reviews.List(context.Background(), &ListReviewsParams{Before: "a", After: "b"})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutual-exclusion error, got %v", err)
	}
}

func TestWorkflowReviewsGet(t *testing.T) {
	var seenMethod, seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod, seenPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewJSON(true))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	review, err := client.Workflows.Reviews.Get(context.Background(), reviewID)
	if err != nil {
		t.Fatal(err)
	}
	if seenMethod != http.MethodGet || seenPath != "/workflows/reviews/"+reviewID {
		t.Fatalf("request = %s %s", seenMethod, seenPath)
	}
	if review.Decision == nil || review.Decision.VersionID != reviewVersionID {
		t.Fatalf("decision = %#v", review.Decision)
	}
}

func TestWorkflowReviewsGetRequiresReviewID(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://example.invalid"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Get(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty reviewID")
	}
}

func TestWorkflowReviewsApproveSendsVersionIDToApproveEndpoint(t *testing.T) {
	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"review":            reviewJSON(true),
			"resume_status":     "resumed",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.Approve(context.Background(), reviewID, ApproveReviewRequest{VersionID: reviewVersionID})
	if err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/"+reviewID+"/approve" {
		t.Fatalf("path = %s", seenPath)
	}
	if body["version_id"] != reviewVersionID || body["verdict"] != nil {
		t.Fatalf("body = %#v", body)
	}
	if resp.ResumeStatus != "resumed" || resp.Review.Decision == nil {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestWorkflowReviewsApproveSurfacesPendingResume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"review":            reviewJSON(true),
			"resume_status":     "pending",
			"resume_error":      "Temporal signal queued for retry",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.Approve(context.Background(), reviewID, ApproveReviewRequest{VersionID: reviewVersionID})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ResumeStatus != "pending" {
		t.Fatalf("ResumeStatus = %q", resp.ResumeStatus)
	}
	if resp.ResumeError == nil || *resp.ResumeError == "" {
		t.Fatalf("ResumeError = %#v", resp.ResumeError)
	}
}

func TestWorkflowReviewsRejectSendsVersionIDAndReasonToRejectEndpoint(t *testing.T) {
	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"submission_status": "accepted", "review": reviewJSON(true)})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Reject(context.Background(), reviewID, RejectReviewRequest{
		VersionID: reviewVersionID, Reason: "wrong document",
	}); err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/"+reviewID+"/reject" {
		t.Fatalf("path = %s", seenPath)
	}
	if body["version_id"] != reviewVersionID || body["reason"] != "wrong document" || body["verdict"] != nil {
		t.Fatalf("body = %#v", body)
	}
}

func TestWorkflowReviewsRejectRequiresInputs(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Reject(context.Background(), reviewID, RejectReviewRequest{Reason: "wrong"}); err == nil || !strings.Contains(err.Error(), "VersionID is required") {
		t.Fatalf("expected VersionID required error, got %v", err)
	}
	if _, err := client.Workflows.Reviews.Reject(context.Background(), reviewID, RejectReviewRequest{VersionID: reviewVersionID}); err == nil || !strings.Contains(err.Error(), "Reason is required") {
		t.Fatalf("expected Reason required error, got %v", err)
	}
}

func TestWorkflowReviewVersionsListGetAndCreate(t *testing.T) {
	var seenPath string
	var seenQuery string
	var seenMethod string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod, seenPath, seenQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/reviews/versions":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []any{reviewVersionJSON(reviewVersionID, nil), reviewVersionJSON(reviewChildVersionID, reviewVersionID)},
				"list_metadata": map[string]any{"before": nil, "after": reviewChildVersionID},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/reviews/versions/"+reviewChildVersionID:
			_ = json.NewEncoder(w).Encode(reviewVersionJSON(reviewChildVersionID, reviewVersionID))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/reviews/versions":
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			_ = json.NewEncoder(w).Encode(reviewVersionJSON(reviewChildVersionID, reviewVersionID))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	list, err := client.Workflows.Reviews.Versions.List(context.Background(), &ListReviewVersionsParams{
		ReviewID: reviewID, Limit: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenMethod != http.MethodGet || seenPath != "/workflows/reviews/versions" {
		t.Fatalf("list request = %s %s", seenMethod, seenPath)
	}
	for _, want := range []string{"review_id=" + reviewID, "limit=25"} {
		if !strings.Contains(seenQuery, want) {
			t.Fatalf("list query %q missing %q", seenQuery, want)
		}
	}
	for _, stale := range []string{"parent_id="} {
		if strings.Contains(seenQuery, stale) {
			t.Fatalf("list query %q should not contain %q", seenQuery, stale)
		}
	}
	if len(list.Data) != 2 || list.Data[1].ParentID == nil || *list.Data[1].ParentID != reviewVersionID {
		t.Fatalf("list = %#v", list)
	}

	version, err := client.Workflows.Reviews.Versions.Get(context.Background(), reviewChildVersionID)
	if err != nil {
		t.Fatal(err)
	}
	if seenMethod != http.MethodGet || seenPath != "/workflows/reviews/versions/"+reviewChildVersionID {
		t.Fatalf("get request = %s %s", seenMethod, seenPath)
	}
	if version.ID != reviewChildVersionID || version.ReviewID != reviewID {
		t.Fatalf("version = %#v", version)
	}

	created, err := client.Workflows.Reviews.Versions.Create(context.Background(), CreateReviewVersionRequest{
		ReviewID: reviewID, ParentID: reviewVersionID, Snapshot: map[string]any{"category": "Invoice"}, Note: "fixed category",
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenMethod != http.MethodPost || seenPath != "/workflows/reviews/versions" {
		t.Fatalf("create request = %s %s", seenMethod, seenPath)
	}
	if body["review_id"] != reviewID || body["parent_id"] != reviewVersionID || body["snapshot"] == nil {
		t.Fatalf("create body = %#v", body)
	}
	if body["parent_version_id"] != nil || body["origin"] != nil {
		t.Fatalf("create body includes removed fields: %#v", body)
	}
	if created.ID != reviewChildVersionID {
		t.Fatalf("created = %#v", created)
	}
}

func TestWorkflowReviewVersionsRequiresInputs(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Workflows.Reviews.Versions.List(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "ReviewID is required") {
		t.Fatalf("expected ReviewID required error, got %v", err)
	}
	_, err = client.Workflows.Reviews.Versions.Create(context.Background(), CreateReviewVersionRequest{
		Snapshot: map[string]any{"category": "Invoice"},
	})
	if err == nil || !strings.Contains(err.Error(), "ReviewID is required") {
		t.Fatalf("expected ReviewID required error, got %v", err)
	}
	_, err = client.Workflows.Reviews.Versions.Create(context.Background(), CreateReviewVersionRequest{
		ReviewID: reviewID, Snapshot: map[string]any{"category": "Invoice"},
	})
	if err == nil || !strings.Contains(err.Error(), "ParentID is required") {
		t.Fatalf("expected ParentID required error, got %v", err)
	}
}
