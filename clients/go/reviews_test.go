package retab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const reviewVersionID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const reviewChildVersionID = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func reviewOverlayJSON(decided bool) map[string]any {
	decision := any(nil)
	if decided {
		decision = map[string]any{
			"verdict":    "approved",
			"version_id": reviewVersionID,
			"decided_by": map[string]any{"kind": "human", "id": "user_1", "display_name": "Ada"},
			"decided_at": "2026-05-18T09:05:00Z",
			"reason":     nil,
		}
	}
	return map[string]any{
		"_id":                 "blockrun_1",
		"organization_id":     "org_1",
		"workflow_id":         "wf_1",
		"workflow_version_id": "wfv_1",
		"workflow_run_id":     "run_1",
		"block_id":            "blk_1",
		"block_run_id":        "blockrun_1",
		"runtime_block_id":    nil,
		"block_type":          "extract",
		"triggered_by":        map[string]any{"kind": "any_required_field_null"},
		"awaiting_since":      "2026-05-18T09:00:00Z",
		"priority":            0,
		"versions_by_id": map[string]any{
			reviewVersionID: map[string]any{
				"parent_id":  nil,
				"author":     map[string]any{"kind": "model", "id": "m", "display_name": "Model"},
				"origin":     "model_output",
				"snapshot":   map[string]any{"output": map[string]any{"total": 100, "currency": "USD"}},
				"created_at": "2026-05-18T09:00:00Z",
			},
			reviewChildVersionID: map[string]any{
				"parent_id":  reviewVersionID,
				"author":     map[string]any{"kind": "agent", "id": "agent_1", "display_name": "Agent"},
				"origin":     "agent_created",
				"snapshot":   map[string]any{"output": map[string]any{"total": 110, "currency": "USD"}},
				"created_at": "2026-05-18T09:01:00Z",
			},
		},
		"decision": decision,
	}
}

func TestWorkflowReviewsList(t *testing.T) {
	var seenMethod, seenPath, seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod, seenPath, seenQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{
				map[string]any{
					"_id":                 "blockrun_1",
					"organization_id":     "org_1",
					"workflow_id":         "wf_1",
					"workflow_version_id": "wfv_1",
					"workflow_run_id":     "run_1",
					"block_id":            "blk_1",
					"block_run_id":        "blockrun_1",
					"block_type":          "extract",
					"triggered_by":        map[string]any{"kind": "any_required_field_null"},
					"awaiting_since":      "2026-05-18T09:00:00Z",
					"priority":            0,
				},
			},
			"has_more": true,
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.List(context.Background(), &ListReviewsParams{
		WorkflowID: "wf_1", Limit: 25, Decision: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenMethod != http.MethodGet {
		t.Fatalf("method = %s", seenMethod)
	}
	if seenPath != "/workflows/reviews" {
		t.Fatalf("path = %s", seenPath)
	}
	for _, want := range []string{"workflow_id=wf_1", "limit=25", "decision=none"} {
		if !strings.Contains(seenQuery, want) {
			t.Fatalf("query %q missing %q", seenQuery, want)
		}
	}
	if strings.Contains(seenQuery, "status=") || strings.Contains(seenQuery, "mine=") {
		t.Fatalf("query includes removed filters: %q", seenQuery)
	}
	if !resp.HasMore || len(resp.Data) != 1 {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestWorkflowReviewsListOmitsDecisionWhenEmpty(t *testing.T) {
	var seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "has_more": false})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.List(context.Background(), &ListReviewsParams{Limit: 5}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(seenQuery, "decision=") {
		t.Fatalf("empty Decision should omit the query param: %q", seenQuery)
	}
}

func TestWorkflowReviewsGet(t *testing.T) {
	var seenMethod, seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod, seenPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayJSON(true))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := client.Workflows.Reviews.Get(context.Background(), "run_1", "blk_1")
	if err != nil {
		t.Fatal(err)
	}
	if seenMethod != http.MethodGet {
		t.Fatalf("method = %s", seenMethod)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1" {
		t.Fatalf("path = %s", seenPath)
	}
	if len(overlay.VersionsByID) != 2 {
		t.Fatalf("versions_by_id = %#v", overlay.VersionsByID)
	}
	if overlay.VersionsByID[reviewChildVersionID].ParentID == nil || *overlay.VersionsByID[reviewChildVersionID].ParentID != reviewVersionID {
		t.Fatalf("child version = %#v", overlay.VersionsByID[reviewChildVersionID])
	}
	if overlay.Decision == nil || overlay.Decision.VersionID != reviewVersionID {
		t.Fatalf("decision = %#v", overlay.Decision)
	}
	encoded, err := json.Marshal(overlay)
	if err != nil {
		t.Fatal(err)
	}
	encodedText := string(encoded)
	for _, want := range []string{`"parent_id":null`, `"note":null`, `"reason":null`} {
		if !strings.Contains(encodedText, want) {
			t.Fatalf("marshaled overlay should preserve %s: %s", want, encodedText)
		}
	}
}

func TestWorkflowReviewsGetRequiresIDs(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://example.invalid"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Get(context.Background(), "", "blk_1"); err == nil {
		t.Fatal("expected error for empty runID")
	}
	if _, err := client.Workflows.Reviews.Get(context.Background(), "run_1", ""); err == nil {
		t.Fatal("expected error for empty blockID")
	}
}

func TestWorkflowReviewsWaitForPollsUntilAwaitingReview(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "no overlay"})
			return
		}
		_ = json.NewEncoder(w).Encode(reviewOverlayJSON(false))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := client.Workflows.Reviews.WaitFor(context.Background(), "run_1", "blk_1", &ReviewWaitForParams{
		PollInterval: time.Millisecond,
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
	if overlay.Decision != nil {
		t.Fatalf("overlay = %#v", overlay)
	}
	encoded, err := json.Marshal(overlay)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"decision":null`) {
		t.Fatalf("marshaled awaiting overlay should preserve decision:null: %s", string(encoded))
	}
}

func TestWorkflowReviewsApproveSendsVersionID(t *testing.T) {
	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"review":            reviewOverlayJSON(true),
			"resume_status":     "resumed",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.Approve(context.Background(), "run_1", "blk_1", ApproveReviewRequest{
		VersionID: reviewVersionID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1/decision" {
		t.Fatalf("path = %s", seenPath)
	}
	if body["verdict"] != "approved" || body["version_id"] != reviewVersionID {
		t.Fatalf("body = %#v", body)
	}
	if _, ok := body["version_stamp"]; ok {
		t.Fatalf("body includes removed version_stamp: %#v", body)
	}
	if resp.SubmissionStatus != "accepted" || resp.Review.Decision == nil {
		t.Fatalf("resp = %#v", resp)
	}
	if resp.ResumeStatus != "resumed" {
		t.Fatalf("ResumeStatus = %q (want \"resumed\")", resp.ResumeStatus)
	}
	if resp.ResumeError != nil {
		t.Fatalf("ResumeError = %#v (want nil)", resp.ResumeError)
	}
}

func TestWorkflowReviewsApproveSurfaceResumeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"review":            reviewOverlayJSON(true),
			"resume_status":     "failed",
			"resume_error":      "Workflow run not found for run_id=run_1",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.Approve(context.Background(), "run_1", "blk_1", ApproveReviewRequest{
		VersionID: reviewVersionID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ResumeStatus != "failed" {
		t.Fatalf("ResumeStatus = %q (want \"failed\")", resp.ResumeStatus)
	}
	if resp.ResumeError == nil || *resp.ResumeError == "" {
		t.Fatalf("ResumeError = %#v (want non-empty)", resp.ResumeError)
	}
}

func TestWorkflowReviewsApproveRequiresVersionID(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Approve(context.Background(), "run_1", "blk_1", ApproveReviewRequest{}); err == nil || !strings.Contains(err.Error(), "VersionID is required") {
		t.Fatalf("expected VersionID required error, got %v", err)
	}
}

func TestWorkflowReviewsRejectSendsVersionIDAndReason(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"review":            reviewOverlayJSON(true),
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Reject(context.Background(), "run_1", "blk_1", RejectReviewRequest{
		VersionID: reviewVersionID, Reason: "wrong document",
	}); err != nil {
		t.Fatal(err)
	}
	if body["verdict"] != "rejected" || body["version_id"] != reviewVersionID || body["reason"] != "wrong document" {
		t.Fatalf("reject body = %#v", body)
	}
}

func TestWorkflowReviewsRejectRequiresReason(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Reject(context.Background(), "run_1", "blk_1", RejectReviewRequest{
		VersionID: reviewVersionID,
	}); err == nil || !strings.Contains(err.Error(), "Reason is required") {
		t.Fatalf("expected Reason required error, got %v", err)
	}
}

func TestWorkflowReviewsRejectRequiresVersionID(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Reject(context.Background(), "run_1", "blk_1", RejectReviewRequest{
		Reason: "wrong document",
	}); err == nil || !strings.Contains(err.Error(), "VersionID is required") {
		t.Fatalf("expected VersionID required error, got %v", err)
	}
}

func TestWorkflowReviewsCreateVersionPostsSnapshotVersion(t *testing.T) {
	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		body = map[string]any{}
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayJSON(false))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.CreateVersion(context.Background(), "run_1", "blk_1", CreateReviewVersionRequest{
		Snapshot: map[string]any{"category": "Invoice"}, ParentID: reviewVersionID, Origin: "human_created", Note: "fixed category",
	}); err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1/versions" {
		t.Fatalf("create-version path = %s", seenPath)
	}
	if body["origin"] != "human_created" || body["parent_id"] != reviewVersionID || body["snapshot"] == nil {
		t.Fatalf("create-version body = %#v", body)
	}
	if _, ok := body["version_stamp"]; ok {
		t.Fatalf("create-version body includes removed version_stamp: %#v", body)
	}
	if _, ok := body["reviewable_value"]; ok {
		t.Fatalf("create-version body includes removed reviewable_value: %#v", body)
	}
}

func TestWorkflowReviewsCreateVersionRequiresParentID(t *testing.T) {
	client, err := NewClient("test-key", WithBaseURL("http://127.0.0.1"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Workflows.Reviews.CreateVersion(context.Background(), "run_1", "blk_1", CreateReviewVersionRequest{
		Snapshot: map[string]any{"category": "Invoice"},
	})
	if err == nil || !strings.Contains(err.Error(), "ParentID is required") {
		t.Fatalf("expected ParentID required error, got %v", err)
	}
}
