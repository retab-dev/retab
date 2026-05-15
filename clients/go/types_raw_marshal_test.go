package retab

import (
	"encoding/json"
	"strings"
	"testing"
)

// WorkflowRun and Workflow capture the verbatim server payload in their
// Raw field on decode. Marshalling must round-trip that payload so a
// server-side field projection (`?fields=id`) survives display: decoding
// into the typed struct and re-encoding must NOT re-inflate fields the
// server never sent (empty `workflow:{}`, `trigger:{}`, ... objects).

func TestWorkflowRunMarshalPreservesServerProjection(t *testing.T) {
	// Mimics `GET /v1/workflows/runs?fields=id` — the server returns only
	// the requested field plus the pagination cursor key.
	projected := `{"id":"run_1","timing":{"created_at":"2026-05-15T13:10:49.74Z"}}`

	var run WorkflowRun
	if err := json.Unmarshal([]byte(projected), &run); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, leaked := range []string{`"workflow"`, `"trigger"`, `"lifecycle"`, `"inputs"`, `"workflow_id"`} {
		if strings.Contains(string(out), leaked) {
			t.Fatalf("re-inflated a field the server never sent (%s):\n%s", leaked, out)
		}
	}
	if !strings.Contains(string(out), `"id":"run_1"`) {
		t.Fatalf("dropped a field the server did send:\n%s", out)
	}
}

func TestWorkflowRunMarshalFallsBackWhenConstructedDirectly(t *testing.T) {
	// A run built in code (not decoded from a response) has no Raw, so
	// marshalling must fall through to normal struct encoding.
	run := WorkflowRun{ID: "run_local"}
	out, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"id":"run_local"`) {
		t.Fatalf("fallback marshal lost the id field:\n%s", out)
	}
}

func TestWorkflowMarshalPreservesServerProjection(t *testing.T) {
	projected := `{"id":"wrk_1"}`

	var wf Workflow
	if err := json.Unmarshal([]byte(projected), &wf); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, leaked := range []string{`"name"`, `"description"`, `"email_trigger"`} {
		if strings.Contains(string(out), leaked) {
			t.Fatalf("re-inflated a field the server never sent (%s):\n%s", leaked, out)
		}
	}
	if string(out) != projected {
		t.Fatalf("projection not preserved: got %s, want %s", out, projected)
	}
}

func TestWorkflowMarshalFallsBackWhenConstructedDirectly(t *testing.T) {
	wf := Workflow{ID: "wrk_local", Name: "local"}
	out, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"id":"wrk_local"`) || !strings.Contains(string(out), `"name":"local"`) {
		t.Fatalf("fallback marshal lost fields:\n%s", out)
	}
}
