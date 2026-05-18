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

// reviewOverlayJSON is a minimal awaiting-review overlay the stub server
// returns for Get / decision / version / claim responses.
func reviewOverlayJSON(rev int, status string) map[string]any {
	return map[string]any{
		"_id":                 "blockrun_1",
		"organization_id":     "org_1",
		"workflow_id":         "wf_1",
		"workflow_version_id": "wfv_1",
		"workflow_run_id":     "run_1",
		"block_id":            "blk_1",
		"block_run_id":        "blockrun_1",
		"block_type":          "extract",
		"triggered_by":        map[string]any{"kind": "any_required_field_null"},
		"status":              status,
		"awaiting_since":      "2026-05-18T09:00:00Z",
		"priority":            0,
		"rev":                 rev,
		"versions": []any{
			map[string]any{
				"seq":            0,
				"parent_seq":     nil,
				"author":         map[string]any{"kind": "model", "id": "m", "display_name": "Model"},
				"origin":         "model_output",
				"snapshot":       map[string]any{"total": 100},
				"content_sha256": "abc",
				"created_at":     "2026-05-18T09:00:00Z",
			},
		},
		"decisions": []any{},
		"audit":     []any{},
		"head_seq":  0,
	}
}

func TestWorkflowReviewsList(t *testing.T) {
	var seenMethod, seenPath, seenQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod, seenPath, seenQuery = r.Method, r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":     []any{reviewOverlayJSON(0, "awaiting_review")},
			"has_more": true,
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.List(context.Background(), &ListReviewsParams{
		WorkflowID: "wf_1", Status: "awaiting_review", Mine: true, Limit: 25,
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
	for _, want := range []string{"workflow_id=wf_1", "status=awaiting_review", "mine=true", "limit=25"} {
		if !strings.Contains(seenQuery, want) {
			t.Fatalf("query %q missing %q", seenQuery, want)
		}
	}
	if !resp.HasMore || len(resp.Data) != 1 {
		t.Fatalf("resp = %#v", resp)
	}
	if resp.Data[0].ID != "blockrun_1" || resp.Data[0].BlockType != "extract" {
		t.Fatalf("queue item = %#v", resp.Data[0])
	}
}

func TestWorkflowReviewsGet(t *testing.T) {
	var seenMethod, seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod, seenPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayJSON(3, "awaiting_review"))
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
	if overlay.Rev != 3 || overlay.Status != "awaiting_review" {
		t.Fatalf("overlay = %#v", overlay)
	}
	if len(overlay.Versions) != 1 || overlay.Versions[0].Origin != "model_output" {
		t.Fatalf("versions = %#v", overlay.Versions)
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

func TestWorkflowReviewsApproveSendsVersionStamp(t *testing.T) {
	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"overlay":           reviewOverlayJSON(1, "approved"),
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.Approve(context.Background(), "run_1", "blk_1", ApproveReviewRequest{
		VersionStamp: 0,
		EditedOutput: map[string]any{"total": 110},
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1/decision" {
		t.Fatalf("path = %s", seenPath)
	}
	if body["verdict"] != "approved" {
		t.Fatalf("verdict = %#v", body["verdict"])
	}
	// version_stamp must always be present — even when 0 (a valid CAS token).
	if _, ok := body["version_stamp"]; !ok {
		t.Fatalf("version_stamp missing from body: %#v", body)
	}
	if body["edited_output"] == nil {
		t.Fatalf("edited_output missing: %#v", body)
	}
	if resp.SubmissionStatus != "accepted" || resp.Overlay.Status != "approved" {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestWorkflowReviewsRejectAndEscalate(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted",
			"overlay":           reviewOverlayJSON(1, "rejected"),
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Reject(context.Background(), "run_1", "blk_1", RejectReviewRequest{
		VersionStamp: 2, Reason: "wrong document",
	}); err != nil {
		t.Fatal(err)
	}
	if body["verdict"] != "rejected" || body["reason"] != "wrong document" {
		t.Fatalf("reject body = %#v", body)
	}

	if _, err := client.Workflows.Reviews.Escalate(context.Background(), "run_1", "blk_1", EscalateReviewRequest{
		VersionStamp: 2, Reason: "needs senior", EscalateTo: "queue_senior",
	}); err != nil {
		t.Fatal(err)
	}
	if body["verdict"] != "escalated" || body["escalate_to"] != "queue_senior" {
		t.Fatalf("escalate body = %#v", body)
	}
}

func TestWorkflowReviewsEditClaimRelease(t *testing.T) {
	var seenPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reviewOverlayJSON(1, "awaiting_review"))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Reviews.Edit(context.Background(), "run_1", "blk_1", EditReviewRequest{
		Snapshot: map[string]any{"total": 110}, VersionStamp: 0, Origin: "human_edit",
	}); err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1/versions" {
		t.Fatalf("edit path = %s", seenPath)
	}
	if body["origin"] != "human_edit" || body["snapshot"] == nil {
		t.Fatalf("edit body = %#v", body)
	}

	if _, err := client.Workflows.Reviews.Claim(context.Background(), "run_1", "blk_1", 1, 600); err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1/claim" {
		t.Fatalf("claim path = %s", seenPath)
	}
	if body["ttl_seconds"] != float64(600) {
		t.Fatalf("claim body = %#v", body)
	}

	if _, err := client.Workflows.Reviews.Release(context.Background(), "run_1", "blk_1", 2); err != nil {
		t.Fatal(err)
	}
	if seenPath != "/workflows/reviews/run_1/blk_1/release" {
		t.Fatalf("release path = %s", seenPath)
	}
}
