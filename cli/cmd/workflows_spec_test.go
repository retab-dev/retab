package cmd

import (
	"bytes"
	"context"
	"encoding/json"
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
	defer f.Close()

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
		"export":   false,
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

func captureRootHelp(t *testing.T) string {
	t.Helper()
	var buf strings.Builder
	rootCmd.SetOut(&buf)
	t.Cleanup(func() { rootCmd.SetOut(nil) })
	rootCmd.HelpFunc()(rootCmd, nil)
	return buf.String()
}

// fakeExportResource builds the kind of envelope the server actually
// returns for `GET /workflows/spec/{id}`: a `workflow_id` plus a
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
	t.Setenv("RETAB_BASE_URL", server.URL)

	workflowsSpecExportCmd.SetContext(context.Background())
	t.Cleanup(func() { workflowsSpecExportCmd.SetContext(nil) })
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

// The --format flag must be registered on the export command (and only
// on the export command — other spec verbs print JSON unconditionally
// and a stray flag would be confusing). Default value is "yaml" so the
// out-of-the-box behaviour matches the bug-fix contract.
func TestWorkflowsSpecExport_FormatFlag(t *testing.T) {
	exp, _, err := rootCmd.Find([]string{"workflows", "spec", "export"})
	if err != nil {
		t.Fatalf("workflows spec export not registered: %v", err)
	}
	f := exp.Flags().Lookup("format")
	if f == nil {
		t.Fatalf("export command should expose a --format flag")
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
			t.Errorf("workflows spec %s should not expose --format (export-only)", name)
		}
	}
}
