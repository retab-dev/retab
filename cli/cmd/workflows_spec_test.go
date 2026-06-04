package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

// readSpecYAML is the only piece of logic in workflows_spec.go that isn't
// already exercised by TestCommandTreeShape (which walks every registered
// command and checks RunE / sibling-name uniqueness for the 4 new leaves).
// It has three real paths — file, stdin, error — and each one matters
// because they're how users actually pipe specs into the CLI.

func TestReadSpecYAML_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	body := "name: test\nblocks:\n  - id: extract\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := readSpecYAML(path)
	if err != nil {
		t.Fatalf("readSpecYAML: %v", err)
	}
	if got != body {
		t.Errorf("body mismatch:\n got: %q\nwant: %q", got, body)
	}
}

// Stdin path ("-"). We can't reliably mock os.Stdin from a test, but we
// can validate the dispatch by swapping the underlying file: open a temp
// file, point os.Stdin at it, restore on cleanup. This exercises the
// branch the user hits with `cat workflow.yaml | retab spec apply -`.
func TestReadSpecYAML_FromStdin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stdin.yaml")
	body := "name: piped\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = f.Close() }()

	orig := os.Stdin
	os.Stdin = f
	t.Cleanup(func() { os.Stdin = orig })

	got, err := readSpecYAML("-")
	if err != nil {
		t.Fatalf("readSpecYAML(-): %v", err)
	}
	if got != body {
		t.Errorf("stdin body mismatch:\n got: %q\nwant: %q", got, body)
	}
}

func TestReadSpecYAML_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := readSpecYAML(path)
	if err == nil {
		t.Fatalf("expected error for empty spec, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention empty, got: %v", err)
	}
}

func TestReadSpecYAML_MissingFile(t *testing.T) {
	_, err := readSpecYAML(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected os.IsNotExist error, got: %v", err)
	}
}

// Cobra wiring sanity — TestCommandTreeShape covers it implicitly, but
// this test names the specific shape so a regression in the wiring (e.g.
// dropping a verb or moving spec out from under workflows) fails with a
// pointing-at-the-bug message instead of a generic tree-walk error.
func TestWorkflowsSpec_RouterAndChildrenRegistered(t *testing.T) {
	spec, _, err := rootCmd.Find([]string{"workflows", "spec"})
	if err != nil {
		t.Fatalf("workflows spec not registered: %v", err)
	}
	if spec.Parent() == nil || spec.Parent().Name() != "workflows" {
		t.Fatalf("spec must be a child of workflows, parent = %v", spec.Parent())
	}

	wantChildren := map[string]bool{
		"validate": false,
		"plan":     false,
		"apply":    false,
		"get":      false,
	}
	for _, c := range spec.Commands() {
		if _, want := wantChildren[c.Name()]; want {
			wantChildren[c.Name()] = true
			if c.RunE == nil && c.Run == nil {
				t.Errorf("spec %s leaf has no RunE/Run", c.Name())
			}
		}
	}
	for name, present := range wantChildren {
		if !present {
			t.Errorf("workflows spec %s is not registered", name)
		}
	}
}

// Spec must surface in the top-level help under Workflows, alongside
// artifacts/blocks/edges/runs/tests/experiments. The renderer's discovery
// rule is "any direct subcommand of workflows that has children", so the
// presence we just established above SHOULD imply this — but we pin it
// with an explicit assertion so a future help.go change that filters
// routers (e.g. by name) can't silently hide spec.
func TestWorkflowsSpec_AppearsInTopLevelHelp(t *testing.T) {
	got := captureRootHelp(t)
	if !strings.Contains(got, "    spec ") {
		t.Errorf("`spec` should appear as a router subcommand under workflows in top-level help:\n%s", got)
	}
}

func TestWorkflowsSpecHelpUsesWorkflowVocabulary(t *testing.T) {
	text := strings.Join([]string{
		workflowsSpecCmd.Long,
		workflowsSpecPlanCmd.Long,
	}, "\n")
	if strings.Contains(text, "Terraform") || strings.Contains(text, "Infrastructure") {
		t.Fatalf("spec help should use workflow vocabulary, got:\n%s", text)
	}
	for _, want := range []string{"declarative workflow", "review-then-apply"} {
		if !strings.Contains(text, want) {
			t.Fatalf("spec help should mention %q, got:\n%s", want, text)
		}
	}
}

func TestWorkflowsSpecValidateReturnsErrorWhenResultIsInvalid(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/workflows/spec/validate" {
			t.Fatalf("path = %s, want /v1/workflows/spec/validate", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workflow_id": "wf_invalid",
			"is_valid":    false,
			"block_count": 2,
			"edge_count":  1,
			"diagnostics": map[string]any{
				"is_valid": false,
				"issues": []map[string]any{
					{
						"severity": "error",
						"code":     "MISSING_USER_DECLARED_SCHEMA",
						"message":  "Function block has no output_schema.",
					},
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("apiVersion: workflows.retab.com/v1alpha2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsSpecValidateCmd.RunE(workflowsSpecValidateCmd, []string{path})
	})
	if err == nil {
		t.Fatal("expected spec validate to fail when response is_valid=false")
	}
	if !strings.Contains(errors.Unwrap(err).Error(), "validation failed") {
		t.Fatalf("error should mention validation failed, got: %v", err)
	}
	if !strings.Contains(stderr, "MISSING_USER_DECLARED_SCHEMA") {
		t.Fatalf("stderr should include diagnostic JSON, got:\n%s", stderr)
	}
}

func TestWorkflowsSpecValidateHonorsTableOutputFallback(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/workflows/spec/validate" {
			t.Fatalf("path = %s, want /v1/workflows/spec/validate", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workflow_id": "wf_valid",
			"is_valid":    true,
			"block_count": 1,
			"edge_count":  0,
			"diagnostics": map[string]any{"is_valid": true},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("apiVersion: workflows.retab.com/v1alpha2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsSpecValidateCmd.RunE(workflowsSpecValidateCmd, []string{path}); err != nil {
			t.Fatalf("spec validate: %v", err)
		}
	})
	if !strings.Contains(stderr, "falling back to json") {
		t.Fatalf("expected table fallback warning, got stderr %q", stderr)
	}
	if !strings.Contains(stdout, `"workflow_id": "wf_valid"`) {
		t.Fatalf("expected JSON fallback payload, got:\n%s", stdout)
	}
}

func TestWorkflowsSpecApplyReturnsErrorWhenResultIsInvalid(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/spec/plan":
			// Plan must return a valid (and non-destructive) shape so
			// the destructive-confirmation guard becomes a no-op and the
			// flow reaches the apply call this test is actually
			// exercising.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_invalid",
				"action":      "update",
				"diagnostics": map[string]any{"is_valid": true},
				"summary":     map[string]any{"add": 0, "change": 0, "destroy": 0},
			})
		case "/v1/workflows/spec/apply":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_invalid",
				"is_valid":    false,
				"diagnostics": map[string]any{
					"is_valid": false,
					"issues": []map[string]any{
						{
							"severity": "error",
							"code":     "INVALID_EDGE_ENDPOINT",
							"message":  "Edge target block is missing.",
						},
					},
				},
			})
		default:
			t.Fatalf("path = %s, want /v1/workflows/spec/plan or /v1/workflows/spec/apply", r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("apiVersion: workflows.retab.com/v1alpha2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsSpecApplyCmd.RunE(workflowsSpecApplyCmd, []string{path})
	})
	if err == nil {
		t.Fatal("expected spec apply to fail when response is_valid=false")
	}
	if !strings.Contains(errors.Unwrap(err).Error(), "validation failed") {
		t.Fatalf("error should mention validation failed, got: %v", err)
	}
	if !strings.Contains(stderr, "INVALID_EDGE_ENDPOINT") {
		t.Fatalf("stderr should include diagnostic JSON, got:\n%s", stderr)
	}
}

func TestWorkflowsSpecApplyToTargetsExistingWorkflowRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var planHits, applyHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var body struct {
			YamlDefinition string `json:"yaml_definition"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !strings.Contains(body.YamlDefinition, "name: Existing Target") {
			t.Fatalf("request YAML mismatch:\n%s", body.YamlDefinition)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/wf_target/spec/plan":
			planHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_target",
				"action":      "update",
				"diagnostics": map[string]any{"is_valid": true},
				"summary":     map[string]any{"add": 0, "change": 1, "destroy": 0},
			})
		case "/v1/workflows/wf_target/spec/apply":
			applyHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_target",
				"action":      "update",
				"created":     false,
				"diagnostics": map[string]any{"is_valid": true},
				"summary":     map[string]any{"add": 0, "change": 1, "destroy": 0},
			})
		default:
			t.Fatalf("path = %s, want target plan/apply route", r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("metadata:\n  name: Existing Target\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	captureStd(t, func() {
		if err := workflowsSpecApplyToCmd.RunE(workflowsSpecApplyToCmd, []string{"wf_target", path}); err != nil {
			t.Fatalf("apply-to: %v", err)
		}
	})
	if planHits.Load() != 1 {
		t.Fatalf("expected exactly 1 plan-to call, got %d", planHits.Load())
	}
	if applyHits.Load() != 1 {
		t.Fatalf("expected exactly 1 apply-to call, got %d", applyHits.Load())
	}
}

func TestWorkflowsSpecPlanToTargetsExistingWorkflowRoute(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var planHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/workflows/wf_target/spec/plan" {
			t.Fatalf("path = %s, want /v1/workflows/wf_target/spec/plan", r.URL.Path)
		}
		var body struct {
			YamlDefinition string `json:"yaml_definition"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !strings.Contains(body.YamlDefinition, "name: Existing Target") {
			t.Fatalf("request YAML mismatch:\n%s", body.YamlDefinition)
		}
		planHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workflow_id": "wf_target",
			"action":      "update",
			"diagnostics": map[string]any{"is_valid": true},
			"summary":     map[string]any{"add": 0, "change": 1, "destroy": 0},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("metadata:\n  name: Existing Target\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	captureStd(t, func() {
		if err := workflowsSpecPlanToCmd.RunE(workflowsSpecPlanToCmd, []string{"wf_target", path}); err != nil {
			t.Fatalf("plan-to: %v", err)
		}
	})
	if planHits.Load() != 1 {
		t.Fatalf("expected exactly 1 plan-to call, got %d", planHits.Load())
	}
}

func TestWorkflowsSpecApplyToWithoutYesAndNonTTYStdinRefusesOnDestroy(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var planHits, applyHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/wf_target/spec/plan":
			planHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_target",
				"action":      "update",
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 0, "change": 1, "destroy": 1,
					"replace": 0, "noop": 0, "total": 2,
					"has_changes": true,
				},
				"resource_changes": []map[string]any{
					{"address": "block.old", "target": "block", "name": "old", "actions": []string{"delete"}},
				},
			})
		case "/v1/workflows/wf_target/spec/apply":
			applyHits.Add(1)
			t.Fatalf("apply-to must NOT be called when destroy > 0 and no --yes")
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("metadata:\n  name: Existing Target\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := workflowsSpecApplyToCmd.Flags().Set("yes", "false"); err != nil {
		t.Fatal(err)
	}

	var err error
	_, _ = captureStd(t, func() {
		err = workflowsSpecApplyToCmd.RunE(workflowsSpecApplyToCmd, []string{"wf_target", path})
	})
	if err == nil {
		t.Fatal("expected refusal when destroy > 0 and stdin is not a TTY")
	}
	for _, want := range []string{"destroy", "--yes", "terminal"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %q", err.Error(), want)
		}
	}
	if planHits.Load() != 1 {
		t.Fatalf("expected 1 plan-to call, got %d", planHits.Load())
	}
	if applyHits.Load() != 0 {
		t.Fatalf("apply-to must not run, got %d hits", applyHits.Load())
	}
}

func TestWorkflowsSpecApplyPrintsForEachMinimumCanvasSize(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var planHits, applyHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var body struct {
			YamlDefinition string `json:"yaml_definition"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !strings.Contains(body.YamlDefinition, "type: for_each") {
			t.Fatalf("request YAML should contain for_each block, got:\n%s", body.YamlDefinition)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/spec/plan":
			planHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wrk_cli_for_each_min_size",
				"action":      "create",
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 3, "change": 0, "destroy": 0,
					"replace": 0, "noop": 0, "total": 3,
					"has_changes": true,
				},
			})
		case "/v1/workflows/spec/apply":
			applyHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wrk_cli_for_each_min_size",
				"action":      "create",
				"created":     true,
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 3, "change": 0, "destroy": 0,
					"replace": 0, "noop": 0, "total": 3,
					"has_changes": true,
				},
				"resource_changes": []map[string]any{
					{
						"address":   "workflow.wrk_cli_for_each_min_size.block.fanout",
						"target":    "block",
						"target_id": "block_fanout",
						"name":      "fanout",
						"type":      "for_each",
						"actions":   []string{"create"},
						"change": map[string]any{
							"after": map[string]any{
								"type":   "for_each",
								"width":  800.0,
								"height": 800.0,
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	spec := `apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  id: wrk_cli_for_each_min_size
  name: CLI For Each Minimum Size
spec:
  blocks:
    start:
      type: start_json
    fanout:
      type: for_each
      width: 320.0
      height: 640.0
      config:
        map_method: iterate_array
        map_source_path: orders
  edges:
    - from: { block: start, handle: output-json-0 }
      to:   { block: fanout, handle: fe-left-in }
`
	if err := os.WriteFile(path, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureStd(t, func() {
		if err := workflowsSpecApplyCmd.RunE(workflowsSpecApplyCmd, []string{path}); err != nil {
			t.Fatalf("spec apply: %v", err)
		}
	})
	if planHits.Load() != 1 || applyHits.Load() != 1 {
		t.Fatalf("plan hits = %d apply hits = %d, want 1 each", planHits.Load(), applyHits.Load())
	}
	if !strings.Contains(stdout, `"type": "for_each"`) {
		t.Fatalf("stdout should include for_each resource, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"width": 800`) || !strings.Contains(stdout, `"height": 800`) {
		t.Fatalf("stdout should include clamped 800x800 dimensions, got:\n%s", stdout)
	}
}

func TestWorkflowsSpecPlanReturnsErrorWhenDiagnosticsAreInvalid(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/workflows/spec/plan" {
			t.Fatalf("path = %s, want /v1/workflows/spec/plan", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workflow_id": "wf_invalid",
			"diagnostics": map[string]any{
				"is_valid": false,
				"issues": []map[string]any{
					{
						"severity": "error",
						"code":     "UNKNOWN_BLOCK_REFERENCE",
						"message":  "Edge references a missing block.",
					},
				},
			},
			"changes": []map[string]any{},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("apiVersion: workflows.retab.com/v1alpha2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsSpecPlanCmd.RunE(workflowsSpecPlanCmd, []string{path})
	})
	if err == nil {
		t.Fatal("expected spec plan to fail when diagnostics.is_valid=false")
	}
	if !strings.Contains(errors.Unwrap(err).Error(), "validation failed") {
		t.Fatalf("error should mention validation failed, got: %v", err)
	}
	if !strings.Contains(stderr, "UNKNOWN_BLOCK_REFERENCE") {
		t.Fatalf("stderr should include diagnostic JSON, got:\n%s", stderr)
	}
}

func captureRootHelp(t *testing.T) string {
	t.Helper()
	var buf strings.Builder
	rootCmd.SetOut(&buf)
	t.Cleanup(func() { rootCmd.SetOut(nil) })
	rootCmd.HelpFunc()(rootCmd, nil)
	return buf.String()
}

// fakeExportResource builds the kind of envelope the server actually
// returns for `GET /workflows/{id}/spec`: a `workflow_id` plus a
// `yaml_definition` string. Keeps each writeSpecExport test self-
// contained instead of stashing a global fixture.
func fakeExportResource(yaml string) *retab.Resource {
	r := retab.Resource{
		"workflow_id":     "wf_test",
		"yaml_definition": yaml,
	}
	return &r
}

// The whole point of `spec export` is to produce a file you can commit
// to git. Format=yaml must emit the YAML body verbatim — no JSON
// envelope, no escape sequences, exactly one trailing newline so a
// shell redirect doesn't double-newline the round-trip.
func TestWriteSpecExport_YAMLFormat(t *testing.T) {
	body := "apiVersion: workflows.retab.com/v1alpha2\nkind: Workflow\nmetadata:\n  name: demo\n"
	r := fakeExportResource(body)

	var buf bytes.Buffer
	if err := writeSpecExport(&buf, r, "yaml"); err != nil {
		t.Fatalf("writeSpecExport: %v", err)
	}
	got := buf.String()

	// Exact match: body verbatim plus a single trailing newline. No JSON
	// braces, no `yaml_definition` key — the file must be directly
	// re-importable by `spec apply`.
	want := body
	if !strings.HasSuffix(want, "\n") {
		want += "\n"
	}
	if got != want {
		t.Errorf("yaml output mismatch:\n got: %q\nwant: %q", got, want)
	}
}

// If the server's YAML already ends with a newline, the writer must
// still produce exactly one — not two. A redirected `> workflow.yaml`
// should round-trip without growing a blank line each time.
func TestWriteSpecExport_YAMLFormat_StripsTrailingNewlines(t *testing.T) {
	body := "name: demo\n\n\n"
	r := fakeExportResource(body)

	var buf bytes.Buffer
	if err := writeSpecExport(&buf, r, "yaml"); err != nil {
		t.Fatalf("writeSpecExport: %v", err)
	}
	got := buf.String()
	if got != "name: demo\n" {
		t.Errorf("expected exactly one trailing newline, got: %q", got)
	}
}

// Format=json preserves the legacy behaviour: the full Resource map
// pretty-printed. Power users opt into this when they want to read
// adjacent fields (e.g. `workflow_id`) with jq.
func TestWriteSpecExport_JSONFormat(t *testing.T) {
	body := "name: demo\n"
	r := fakeExportResource(body)

	var buf bytes.Buffer
	if err := writeSpecExport(&buf, r, "json"); err != nil {
		t.Fatalf("writeSpecExport: %v", err)
	}

	// Round-trip the JSON to confirm structure (indented output, key
	// presence) rather than pinning whitespace.
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json output not parseable: %v\nraw: %s", err, buf.String())
	}
	if decoded["yaml_definition"] != body {
		t.Errorf("yaml_definition mismatch: got %q", decoded["yaml_definition"])
	}
	if decoded["workflow_id"] != "wf_test" {
		t.Errorf("workflow_id missing or wrong: got %v", decoded["workflow_id"])
	}
	if !strings.Contains(buf.String(), "\n  \"") {
		t.Errorf("json output should be indented:\n%s", buf.String())
	}
}

// The error path the bug report cares about: if the server returns a
// response with no `yaml_definition` field (or an empty one), the
// command must fail loudly. An empty file would be worse than the
// JSON-wrapped behaviour we just fixed — users would commit a blank
// YAML to git and only notice on the next `apply`.
func TestWriteSpecExport_MissingYAMLDefinition(t *testing.T) {
	cases := []struct {
		name     string
		resource *retab.Resource
	}{
		{
			name:     "nil resource",
			resource: nil,
		},
		{
			name:     "field absent",
			resource: &retab.Resource{"workflow_id": "wf_test"},
		},
		{
			name:     "field empty string",
			resource: &retab.Resource{"yaml_definition": ""},
		},
		{
			name:     "field wrong type",
			resource: &retab.Resource{"yaml_definition": 42},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeSpecExport(&buf, tc.resource, "yaml")
			if err == nil {
				t.Fatalf("expected error for missing yaml_definition, got nil (output=%q)", buf.String())
			}
			if !strings.Contains(err.Error(), "missing yaml_definition") {
				t.Errorf("error should mention missing yaml_definition, got: %v", err)
			}
		})
	}
}

// Unknown format values surface as a CLI error rather than silently
// defaulting to yaml or json. Catches typos like `--format yml`.
func TestWriteSpecExport_UnknownFormat(t *testing.T) {
	r := fakeExportResource("name: demo\n")
	var buf bytes.Buffer
	err := writeSpecExport(&buf, r, "yml")
	if err == nil {
		t.Fatalf("expected error for unknown format, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --format") {
		t.Errorf("error should mention invalid --format, got: %v", err)
	}
}

func TestWorkflowsSpecExportRejectsUnknownFormatBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workflow_id":     "wf_test",
			"yaml_definition": "metadata:\n  id: wf_test\n",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	workflowsSpecExportCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsSpecExportCmd.SetContext(context.Background()) })
	if err := workflowsSpecExportCmd.Flags().Set("format", "yml"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsSpecExportCmd.Flags().Set("format", "yaml") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsSpecExportCmd.RunE(workflowsSpecExportCmd, []string{"wf_test"})
	})
	if err == nil {
		t.Fatalf("expected invalid format error")
	}
	if !strings.Contains(stderr, "invalid --format") {
		t.Fatalf("stderr %q does not mention invalid --format", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

// TestWorkflowsSpecApplyWithYesFlagSkipsPromptWhenDestroyPositive pins
// that `--yes` lets the apply through without prompting even when the
// plan would destroy resources. Mirrors the `--yes` contract used by
// `workflows blocks delete` / `workflows delete` — scripts must be able
// to opt in without a TTY.
func TestWorkflowsSpecApplyWithYesFlagSkipsPromptWhenDestroyPositive(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var planHits, applyHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/spec/plan":
			planHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_test",
				"action":      "update",
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 1, "change": 0, "destroy": 3,
					"replace": 0, "noop": 0, "total": 4,
					"has_changes": true,
				},
				"resource_changes": []map[string]any{
					{"address": "block.extract_invoice", "target": "block", "name": "extract_invoice", "actions": []string{"delete"}},
					{"address": "block.classify", "target": "block", "name": "classify", "actions": []string{"delete"}},
					{"address": "edge.e1", "target": "edge", "name": "e1", "actions": []string{"delete"}},
				},
			})
		case "/v1/workflows/spec/apply":
			applyHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_test",
				"created":     false,
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 1, "change": 0, "destroy": 3,
					"replace": 0, "noop": 0, "total": 4,
					"has_changes": true,
				},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("metadata:\n  id: wf_test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := workflowsSpecApplyCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsSpecApplyCmd.Flags().Set("yes", "false") })

	captureStd(t, func() {
		if err := workflowsSpecApplyCmd.RunE(workflowsSpecApplyCmd, []string{path}); err != nil {
			t.Fatalf("apply with --yes: %v", err)
		}
	})
	if planHits.Load() != 1 {
		t.Fatalf("expected exactly 1 plan call, got %d", planHits.Load())
	}
	if applyHits.Load() != 1 {
		t.Fatalf("expected exactly 1 apply call, got %d", applyHits.Load())
	}
}

// TestWorkflowsSpecApplyWithoutYesAndNonTTYStdinRefusesOnDestroy pins
// the safety guard: when the plan would destroy resources, stdin is not
// a terminal (CI, pipe, redirect), and `--yes` is not set, the command
// must refuse — apply is never invoked. The error message has to mention
// "destroy", "--yes", and "terminal" so users can see all three pieces
// of the diagnosis in one line.
func TestWorkflowsSpecApplyWithoutYesAndNonTTYStdinRefusesOnDestroy(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var planHits, applyHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/spec/plan":
			planHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_test",
				"action":      "update",
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 1, "change": 1, "destroy": 3,
					"replace": 0, "noop": 0, "total": 5,
					"has_changes": true,
				},
				"resource_changes": []map[string]any{
					{"address": "block.extract_invoice", "target": "block", "name": "extract_invoice", "actions": []string{"delete"}},
				},
			})
		case "/v1/workflows/spec/apply":
			applyHits.Add(1)
			t.Fatalf("apply must NOT be called when destroy > 0 and no --yes")
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("metadata:\n  id: wf_test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := workflowsSpecApplyCmd.Flags().Set("yes", "false"); err != nil {
		t.Fatal(err)
	}

	var err error
	_, _ = captureStd(t, func() {
		err = workflowsSpecApplyCmd.RunE(workflowsSpecApplyCmd, []string{path})
	})
	if err == nil {
		t.Fatal("expected refusal when destroy > 0 and stdin is not a TTY")
	}
	msg := err.Error()
	for _, want := range []string{"destroy", "--yes", "terminal"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not mention %q", msg, want)
		}
	}
	if planHits.Load() != 1 {
		t.Fatalf("expected 1 plan call, got %d", planHits.Load())
	}
	if applyHits.Load() != 0 {
		t.Fatalf("apply must not run, got %d hits", applyHits.Load())
	}
}

// TestWorkflowsSpecApplyDestroyZeroAppliesUnconditionally pins that a
// non-destructive plan applies without consulting --yes or TTY state.
// This is the byte-identical-to-pre-guard contract: a benign apply must
// not gain a prompt nor a refusal just because we added the gate.
func TestWorkflowsSpecApplyDestroyZeroAppliesUnconditionally(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var planHits, applyHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/spec/plan":
			planHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_test",
				"action":      "update",
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 2, "change": 1, "destroy": 0,
					"replace": 0, "noop": 0, "total": 3,
					"has_changes": true,
				},
				"resource_changes": []map[string]any{
					{"address": "block.new", "target": "block", "name": "new", "actions": []string{"create"}},
				},
			})
		case "/v1/workflows/spec/apply":
			applyHits.Add(1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow_id": "wf_test",
				"created":     false,
				"diagnostics": map[string]any{"is_valid": true},
				"summary": map[string]any{
					"add": 2, "change": 1, "destroy": 0,
					"replace": 0, "noop": 0, "total": 3,
					"has_changes": true,
				},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	path := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(path, []byte("metadata:\n  id: wf_test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Ensure --yes is explicitly false; the gate must be skipped on its
	// own when destroy == 0, not because of a leftover --yes from a
	// sibling test.
	if err := workflowsSpecApplyCmd.Flags().Set("yes", "false"); err != nil {
		t.Fatal(err)
	}

	var err error
	_, _ = captureStd(t, func() {
		err = workflowsSpecApplyCmd.RunE(workflowsSpecApplyCmd, []string{path})
	})
	if err != nil {
		t.Fatalf("non-destructive apply should succeed without --yes / TTY, got: %v", err)
	}
	if planHits.Load() != 1 || applyHits.Load() != 1 {
		t.Fatalf("expected 1 plan + 1 apply call, got plan=%d apply=%d",
			planHits.Load(), applyHits.Load())
	}
}

// The --yes flag must be registered on the apply command (and only on
// apply — the other spec verbs are read-only and a stray --yes would
// silently be accepted but ignored, which is confusing).
func TestWorkflowsSpecApply_YesFlag(t *testing.T) {
	apply, _, err := rootCmd.Find([]string{"workflows", "spec", "apply"})
	if err != nil {
		t.Fatalf("workflows spec apply not registered: %v", err)
	}
	f := apply.Flags().Lookup("yes")
	if f == nil {
		t.Fatalf("apply command should expose a --yes flag")
	}
	if f.Shorthand != "y" {
		t.Errorf("--yes shorthand should be -y, got %q", f.Shorthand)
	}
	for _, name := range []string{"validate", "plan", "get"} {
		sibling, _, err := rootCmd.Find([]string{"workflows", "spec", name})
		if err != nil {
			t.Fatalf("workflows spec %s not registered: %v", name, err)
		}
		if sibling.Flags().Lookup("yes") != nil {
			t.Errorf("workflows spec %s should not expose --yes (apply-only)", name)
		}
	}
}

// The --format flag must be registered on the get command (and only
// on the get command — other spec verbs print JSON unconditionally
// and a stray flag would be confusing). Default value is "yaml" so the
// out-of-the-box behaviour matches the bug-fix contract.
func TestWorkflowsSpecExport_FormatFlag(t *testing.T) {
	exp, _, err := rootCmd.Find([]string{"workflows", "spec", "get"})
	if err != nil {
		t.Fatalf("workflows spec get not registered: %v", err)
	}
	f := exp.Flags().Lookup("format")
	if f == nil {
		t.Fatalf("get command should expose a --format flag")
	}
	if f.DefValue != "yaml" {
		t.Errorf("--format default should be yaml, got %q", f.DefValue)
	}

	// The flag must NOT leak onto the sibling verbs. Each of them still
	// only prints JSON — adding --format would be silently ignored and
	// look like a real opt-out the user could rely on.
	for _, name := range []string{"validate", "plan", "apply"} {
		sibling, _, err := rootCmd.Find([]string{"workflows", "spec", name})
		if err != nil {
			t.Fatalf("workflows spec %s not registered: %v", name, err)
		}
		if sibling.Flags().Lookup("format") != nil {
			t.Errorf("workflows spec %s should not expose --format (get-only)", name)
		}
	}

	alias, _, err := rootCmd.Find([]string{"workflows", "spec", "export"})
	if err != nil {
		t.Fatalf("workflows spec export alias not registered: %v", err)
	}
	if alias != exp {
		t.Fatalf("workflows spec export should resolve to the get command")
	}
}
