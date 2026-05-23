package retab

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkflowRunMarshalUsesGeneratedStructShape(t *testing.T) {
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

	for _, want := range []string{`"id":"run_1"`, `"workflow"`, `"trigger":null`} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("generated marshal output missing %s:\n%s", want, out)
		}
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

func TestWorkflowMarshalUsesGeneratedDefaults(t *testing.T) {
	projected := `{"id":"wrk_1"}`

	var wf Workflow
	if err := json.Unmarshal([]byte(projected), &wf); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, want := range []string{`"id":"wrk_1"`, `"name":"Untitled Workflow"`} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("generated marshal output missing %s:\n%s", want, out)
		}
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
