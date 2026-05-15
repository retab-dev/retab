package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
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

func TestWorkflowsListRejectsInvalidListFlagsLocally(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
		reset     string
	}{
		{name: "negative limit", flag: "limit", value: "-1", wantError: "between 0 and 100", reset: "0"},
		{name: "invalid order", flag: "order", value: "sideways", wantError: "asc", reset: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := workflowsListCmd.Flags().Set(tc.flag, tc.value)
			if err == nil {
				t.Fatalf("expected local parse error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if resetErr := workflowsListCmd.Flags().Set(tc.flag, tc.reset); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestWorkflowsListRejectsOverLimitLocally(t *testing.T) {
	cases := []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "workflows list", cmd: workflowsListCmd},
		{name: "workflows snapshots", cmd: workflowsSnapshotsCmd},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Flags().Set("limit", "101")
			if err == nil {
				t.Fatal("expected local parse error for --limit=101")
			}
			if !strings.Contains(err.Error(), "between 0 and 100") {
				t.Fatalf("error %q does not mention backend limit range", err.Error())
			}
			if resetErr := tc.cmd.Flags().Set("limit", "0"); resetErr != nil {
				t.Fatalf("reset --limit: %v", resetErr)
			}
		})
	}
}

func TestWorkflowExamplesUseCurrentDocumentFlag(t *testing.T) {
	for _, tc := range []struct {
		name    string
		example string
	}{
		{name: "workflows", example: workflowsCmd.Example},
		{name: "publish", example: workflowsPublishCmd.Example},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Contains(tc.example, "--document-file") {
				t.Fatalf("%s example should use --document, got:\n%s", tc.name, tc.example)
			}
			if !strings.Contains(tc.example, "--document start=") {
				t.Fatalf("%s example should include --document start=..., got:\n%s", tc.name, tc.example)
			}
		})
	}
}

func TestWorkflowsUpdateRejectsBlankEmailAllowlistValuesBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name      string
		flag      string
		wantError string
	}{
		{name: "blank sender", flag: "allowed-sender", wantError: "--allowed-sender must not be blank"},
		{name: "blank domain", flag: "allowed-domain", wantError: "--allowed-domain must not be blank"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for invalid allowlist value, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "test-workflow-update", RunE: workflowsUpdateCmd.RunE}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().StringArray("allowed-sender", nil, "")
			cmd.Flags().StringArray("allowed-domain", nil, "")

			if err := cmd.Flags().Set(tc.flag, "   "); err != nil {
				t.Fatal(err)
			}

			err := cmd.RunE(cmd, []string{"wf_123"})
			if err == nil {
				t.Fatal("expected blank allowlist value error")
			}
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				err = unwrapped
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
	}
}

func TestWorkflowsRejectBlankNamesBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	cases := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{name: "create", cmd: workflowsCreateCmd},
		{name: "update", cmd: workflowsUpdateCmd, args: []string{"wf_123"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				t.Fatalf("server should not be reached for blank workflow name, got %s %s", r.Method, r.URL.Path)
			}))
			defer server.Close()
			t.Setenv("RETAB_BASE_URL", server.URL)

			cmd := &cobra.Command{Use: "test-workflow-" + tc.name, RunE: tc.cmd.RunE}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			if tc.name == "update" {
				cmd.Flags().StringArray("allowed-sender", nil, "")
				cmd.Flags().StringArray("allowed-domain", nil, "")
			}

			if err := cmd.Flags().Set("name", "   "); err != nil {
				t.Fatal(err)
			}

			err := cmd.RunE(cmd, tc.args)
			if err == nil {
				t.Fatal("expected blank workflow name error")
			}
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				err = unwrapped
			}
			if !strings.Contains(err.Error(), "workflow name must not be blank") {
				t.Fatalf("error %q does not mention blank workflow name", err.Error())
			}
			if got := hits.Load(); got != 0 {
				t.Fatalf("server was hit %d time(s), want 0", got)
			}
		})
	}
}

func TestWorkflowListExamplesUsePaginatedEnvelope(t *testing.T) {
	if strings.Contains(workflowsBlocksListCmd.Example, ".[].") {
		t.Fatalf("blocks list example should iterate over .data[], got:\n%s", workflowsBlocksListCmd.Example)
	}
	if !strings.Contains(workflowsBlocksListCmd.Example, ".data[].id") {
		t.Fatalf("blocks list example should read .data[].id, got:\n%s", workflowsBlocksListCmd.Example)
	}
}

func TestWorkflowBatchCreateRejectsEmptyArraysLocally(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	emptyArrayPath := t.TempDir() + "/empty.json"
	if err := os.WriteFile(emptyArrayPath, []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		cmd  *cobra.Command
		flag string
	}{
		{name: "blocks", cmd: workflowsBlocksCreateBatchCmd, flag: "blocks-file"},
		{name: "edges", cmd: workflowsEdgesCreateBatchCmd, flag: "edges-file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			tc.cmd.SetContext(context.Background())
			t.Cleanup(func() { tc.cmd.SetContext(nil) })
			if err := tc.cmd.Flags().Set(tc.flag, emptyArrayPath); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = tc.cmd.Flags().Set(tc.flag, "") })

			var err error
			_, stderr := captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, []string{"wf_empty"})
			})
			if err == nil {
				t.Fatalf("expected empty-array error")
			}
			if !strings.Contains(stderr, "empty JSON array") {
				t.Fatalf("stderr %q does not mention empty JSON array", stderr)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want no new requests", got-before)
			}
		})
	}
}

func TestWorkflowsDiagnoseGraphFileRejectsMalformedGraphLocally(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cases := []struct {
		name      string
		body      string
		wantError string
	}{
		{name: "blocks not array", body: `{"blocks":{},"edges":[]}`, wantError: "blocks must be an array"},
		{name: "block item not object", body: `{"blocks":[1],"edges":[]}`, wantError: "blocks[0] must be an object"},
		{name: "edges not array", body: `{"blocks":[],"edges":{}}`, wantError: "edges must be an array"},
		{name: "edge item not object", body: `{"blocks":[],"edges":[1]}`, wantError: "edges[0] must be an object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			path := t.TempDir() + "/graph.json"
			if err := os.WriteFile(path, []byte(tc.body), 0o600); err != nil {
				t.Fatal(err)
			}
			workflowsDiagnoseCmd.SetContext(context.Background())
			t.Cleanup(func() { workflowsDiagnoseCmd.SetContext(nil) })
			if err := workflowsDiagnoseCmd.Flags().Set("graph-file", path); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = workflowsDiagnoseCmd.Flags().Set("graph-file", "") })

			var err error
			_, stderr := captureStd(t, func() {
				err = workflowsDiagnoseCmd.RunE(workflowsDiagnoseCmd, []string{"wf_graph"})
			})
			if err == nil {
				t.Fatalf("expected malformed graph error")
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want no new requests", got-before)
			}
		})
	}
}

func TestWorkflowsDiagnoseGraphFileAcceptsEntitiesShape(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/workflows/wf_graph/diagnose-graph" {
			t.Fatalf("path = %s, want diagnose-graph", r.URL.Path)
		}
		var posted map[string]any
		if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		blocks, ok := posted["blocks"].([]any)
		if !ok || len(blocks) != 1 {
			http.Error(w, "missing blocks array", http.StatusBadRequest)
			return
		}
		block, ok := blocks[0].(map[string]any)
		if !ok {
			http.Error(w, "block is not object", http.StatusBadRequest)
			return
		}
		position, ok := block["position"].(map[string]any)
		if !ok || position["x"] != float64(11) || position["y"] != float64(22) {
			http.Error(w, "block position was not normalized", http.StatusBadRequest)
			return
		}
		edges, ok := posted["edges"].([]any)
		if !ok || len(edges) != 1 {
			http.Error(w, "missing edges array", http.StatusBadRequest)
			return
		}
		edge, ok := edges[0].(map[string]any)
		if !ok || edge["source"] != "start" || edge["target"] != "extract" {
			http.Error(w, "edge endpoints were not normalized", http.StatusBadRequest)
			return
		}
		if posted["re_propagate"] != false {
			http.Error(w, "re_propagate was not preserved", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_valid":    true,
			"issues":      []any{},
			"suggestions": []any{},
			"stats": map[string]any{
				"total_blocks": 1,
				"total_edges":  1,
				"block_types":  map[string]any{"start": 1},
				"start_blocks": 1,
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	graphPath := t.TempDir() + "/entities.json"
	graph := `{
		"workflow": {"id": "wf_graph"},
		"blocks": [
			{
				"id": "start",
				"type": "start",
				"label": "Start",
				"config": {},
				"position_x": 11,
				"position_y": 22,
				"width": 300,
				"height": 200
			}
		],
		"edges": [
			{
				"id": "edge_1",
				"source_block": "start",
				"target_block": "extract",
				"source_handle": "output",
				"target_handle": "input"
			}
		],
		"re_propagate": false
	}`
	if err := os.WriteFile(graphPath, []byte(graph), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowsDiagnoseCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsDiagnoseCmd.SetContext(nil) })
	if err := workflowsDiagnoseCmd.Flags().Set("graph-file", graphPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsDiagnoseCmd.Flags().Set("graph-file", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsDiagnoseCmd.RunE(workflowsDiagnoseCmd, []string{"wf_graph"}); err != nil {
			t.Fatalf("diagnose graph-file: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, `"is_valid": true`) {
		t.Fatalf("expected diagnosis response, got:\n%s", stdout)
	}
}

func TestWorkflowsSnapshotsCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"workflows", "snapshots", "wf_abc"})
	if err != nil {
		t.Fatalf("workflows snapshots not registered: %v", err)
	}
	if cmd.Name() != "snapshots" {
		t.Fatalf("resolved command = %q, want snapshots", cmd.Name())
	}
	if cmd.Flags().Lookup("limit") == nil {
		t.Fatal("workflows snapshots should expose --limit")
	}
}

func TestWorkflowsGetExampleUsesPublishedObjectShape(t *testing.T) {
	if strings.Contains(workflowsGetCmd.Example, ".published_version") {
		t.Fatalf("workflows get example references removed field:\n%s", workflowsGetCmd.Example)
	}
	if !strings.Contains(workflowsGetCmd.Example, ".published.version_id") {
		t.Fatalf("workflows get example should show current published version path:\n%s", workflowsGetCmd.Example)
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
