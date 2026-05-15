package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

// isEffectivelyEmptyDraft is the pure shape predicate that decides whether
// a draft warrants the empty-workflow warning. Pin every interesting
// shape — fully empty, single start, single non-start, multi-block —
// so a future refactor can't silently change behaviour.
func TestIsEffectivelyEmptyDraft(t *testing.T) {
	cases := []struct {
		name   string
		blocks []retab.WorkflowBlock
		want   bool
	}{
		{name: "no blocks", blocks: nil, want: true},
		{name: "empty slice", blocks: []retab.WorkflowBlock{}, want: true},
		{name: "single start block — freshly-created shape", blocks: []retab.WorkflowBlock{{Type: "start"}}, want: true},
		{name: "single non-start block", blocks: []retab.WorkflowBlock{{Type: "extract"}}, want: false},
		{name: "two blocks", blocks: []retab.WorkflowBlock{{Type: "start"}, {Type: "extract"}}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isEffectivelyEmptyDraft(tc.blocks); got != tc.want {
				t.Errorf("isEffectivelyEmptyDraft(%+v) = %v, want %v", tc.blocks, got, tc.want)
			}
		})
	}
}

// TestWarnIfEmptyWorkflowOnPublish_StartOnly mocks Blocks.List against a
// fake HTTP server returning a single `start` block, then asserts the
// warning text — and only that text — lands on the provided writer.
// This is the canonical "user fat-fingered `workflows publish` on a
// fresh draft" scenario.
func TestWarnIfEmptyWorkflowOnPublish_StartOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workflows/wf_abc/blocks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "start_1", "workflow_id": "wf_abc", "type": "start", "label": "Start"},
			},
			"list_metadata": map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := retab.NewClient("fake-key", retab.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var buf bytes.Buffer
	warnIfEmptyWorkflowOnPublish(context.Background(), client, "wf_abc", &buf)

	out := buf.String()
	if !strings.Contains(out, "warning: workflow has only a start block") {
		t.Errorf("missing primary warning. got:\n%s", out)
	}
	if !strings.Contains(out, "add blocks with `retab workflows blocks create`") {
		t.Errorf("missing follow-up warning. got:\n%s", out)
	}
}

// TestWarnIfEmptyWorkflowOnPublish_NonEmpty pins the silent path: when
// the workflow has real blocks (start + extract), no warning is emitted.
// Catches future regressions where the predicate widens accidentally.
func TestWarnIfEmptyWorkflowOnPublish_NonEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "start_1", "type": "start"},
				{"id": "blk_extract_1", "type": "extract"},
			},
			"list_metadata": map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := retab.NewClient("fake-key", retab.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var buf bytes.Buffer
	warnIfEmptyWorkflowOnPublish(context.Background(), client, "wf_abc", &buf)

	if buf.Len() != 0 {
		t.Errorf("expected silent on non-empty workflow, got:\n%s", buf.String())
	}
}

// TestWarnIfEmptyWorkflowOnPublish_NetworkError pins the best-effort
// contract: a server error MUST NOT block publishing. The helper swallows
// the error and emits nothing — the publish call itself will surface any
// real auth/server failure separately.
func TestWarnIfEmptyWorkflowOnPublish_NetworkError(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := retab.NewClient("fake-key", retab.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var buf bytes.Buffer
	warnIfEmptyWorkflowOnPublish(context.Background(), client, "wf_abc", &buf)

	if calls.Load() == 0 {
		t.Fatal("expected helper to hit the server at least once")
	}
	if buf.Len() != 0 {
		t.Errorf("expected silent on server error (best-effort), got:\n%s", buf.String())
	}
}
