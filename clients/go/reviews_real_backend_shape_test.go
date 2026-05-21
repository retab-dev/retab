package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Probe: does the Go SDK decode what the LIVE backend actually returns?
//
// Verified by Python probing the FastAPI route stack end-to-end against dev
// mongo. The backend ReviewResponse model strips `organization_id` and
// `runtime_block_id`, and the SubmitDecisionResponse envelope uses the field
// name `review` (NOT `overlay`) for the embedded review payload.
//
// If these don't match the Go struct tags, every Approve/Reject call
// silently returns a zero-valued Overlay — a critical bug.

// realPublicOverlayJSON mirrors the actual JSON the backend emits.
// Notable: no `organization_id`, no `runtime_block_id`.
func realPublicOverlayJSON() map[string]any {
	return map[string]any{
		"_id":                 "blockrun_1",
		"workflow_id":         "wf_1",
		"workflow_version_id": "wfv_1",
		"workflow_run_id":     "run_1",
		"block_id":            "blk_1",
		"block_run_id":        "blockrun_1",
		"block_type":          "extract",
		"triggered_by":        map[string]any{"kind": "any_required_field_null"},
		"awaiting_since":      "2026-05-21T09:00:00Z",
		"priority":            0,
		"versions_by_id": map[string]any{
			reviewVersionID: map[string]any{
				"parent_id":  nil,
				"author":     map[string]any{"kind": "model", "id": "m", "display_name": "Model"},
				"origin":     "model_output",
				"snapshot":   map[string]any{"total": 100, "currency": nil},
				"note":       nil,
				"created_at": "2026-05-21T09:00:00Z",
			},
		},
		"decision": map[string]any{
			"verdict":    "approved",
			"version_id": reviewVersionID,
			"decided_by": map[string]any{"kind": "human", "id": "ada", "display_name": "ada"},
			"decided_at": "2026-05-21T09:05:00Z",
			"reason":     nil,
		},
	}
}

// TestRealBackendWireShape_SubmitDecisionUsesReviewField verifies that the
// SDK decodes the `review` field (not `overlay`). The backend route layer
// returns `{"submission_status": ..., "review": {...}, "resume_status": ..., "resume_error": ...}`.
func TestRealBackendWireShape_SubmitDecisionUsesReviewField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"submission_status": "accepted_pending_resume",
			"review":            realPublicOverlayJSON(),
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
	if resp.SubmissionStatus != SubmissionStatusAcceptedPendingResume {
		t.Fatalf("SubmissionStatus = %q (want %q)", resp.SubmissionStatus, SubmissionStatusAcceptedPendingResume)
	}
	// CRITICAL: the SDK must surface the review the backend returned.
	if resp.Review.ID == "" {
		t.Fatalf("Review was zero-valued — Go struct must decode the `review` field the backend emits")
	}
	if resp.Review.ID != "blockrun_1" {
		t.Fatalf("Review.ID = %q (want %q)", resp.Review.ID, "blockrun_1")
	}
	if resp.Review.Decision == nil || resp.Review.Decision.VersionID != reviewVersionID {
		t.Fatalf("Review.Decision = %#v", resp.Review.Decision)
	}
	if resp.ResumeStatus != ResumeStatusFailed {
		t.Fatalf("ResumeStatus = %q", resp.ResumeStatus)
	}
	if resp.ResumeError == nil || *resp.ResumeError == "" {
		t.Fatalf("ResumeError = %v", resp.ResumeError)
	}
}

// TestSubmissionStatusConstantsMatchBackendWire pins the SDK constants
// against the literal strings the backend emits. A drift here means the
// constants would silently miscompare in downstream callers.
func TestSubmissionStatusConstantsMatchBackendWire(t *testing.T) {
	cases := map[string]string{
		"accepted":                SubmissionStatusAccepted,
		"already_applied":         SubmissionStatusAlreadyApplied,
		"accepted_pending_resume": SubmissionStatusAcceptedPendingResume,
		"resumed":                 ResumeStatusResumed,
		"skipped":                 ResumeStatusSkipped,
		"failed":                  ResumeStatusFailed,
	}
	for wire, constant := range cases {
		if constant != wire {
			t.Errorf("constant %q does not match wire value %q", constant, wire)
		}
	}
}

// TestRealBackendWireShape_ReviewOverlayLacksOrganizationID verifies that
// the Go ReviewOverlay struct decodes cleanly from a payload that omits
// `organization_id` and `runtime_block_id` (which the backend strips).
func TestRealBackendWireShape_ReviewOverlayLacksOrganizationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(realPublicOverlayJSON())
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
	if overlay.ID != "blockrun_1" {
		t.Fatalf("Overlay.ID = %q", overlay.ID)
	}
	if overlay.Decision == nil {
		t.Fatalf("Decision should be present in the seeded payload")
	}
}

// TestRealBackendWireShape_QueueResponseLacksOrganizationID verifies the
// queue projection also works without `organization_id`.
func TestRealBackendWireShape_QueueResponseLacksOrganizationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []any{
				map[string]any{
					"_id":                 "blockrun_1",
					"workflow_id":         "wf_1",
					"workflow_version_id": "wfv_1",
					"workflow_run_id":     "run_1",
					"block_id":            "blk_1",
					"block_run_id":        "blockrun_1",
					"block_type":          "extract",
					"triggered_by":        map[string]any{"kind": "any_required_field_null"},
					"awaiting_since":      "2026-05-21T09:00:00Z",
					"priority":            0,
				},
			},
			"has_more": false,
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Workflows.Reviews.List(context.Background(), &ListReviewsParams{Decision: "none"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("Data = %#v", resp.Data)
	}
	if resp.Data[0].ID != "blockrun_1" {
		t.Fatalf("Data[0].ID = %q", resp.Data[0].ID)
	}
}
