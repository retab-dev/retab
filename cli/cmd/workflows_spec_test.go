package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
