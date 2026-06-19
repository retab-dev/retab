//go:build !retab_oagen_cli_workflows_evals

package cmd

import (
	"strings"
	"testing"
)

// Bug 10: the `--source-file` help string only said "the input. Usually
// a reference to a run/step that supplied known-good input" and did not
// describe the actual discriminated-union schema, so users hit
// `Extra inputs are not permitted` 422s after guessing the field names.
//
// Pin the documented schema for both source variants in the Long-form
// help block (so `evals create --help` is enough to answer "what fields
// does --source-file expect?").
func TestWorkflowsEvalsCreateLongDocumentsBothSourceShapes(t *testing.T) {
	long := workflowsEvalsCreateCmd.Long
	if long == "" {
		t.Fatal("workflowsEvalsCreateCmd.Long is empty")
	}
	// `manual` variant — must document handle_inputs.
	for _, want := range []string{
		"manual",
		"handle_inputs",
	} {
		if !strings.Contains(long, want) {
			t.Fatalf("expected --help to document the manual source variant (%q):\n%s", want, long)
		}
	}
	// `run_step` variant — must document run_id (required) and step_id
	// (optional).
	for _, want := range []string{
		"run_step",
		"run_id",
		"step_id",
	} {
		if !strings.Contains(long, want) {
			t.Fatalf("expected --help to document the run_step source variant (%q):\n%s", want, long)
		}
	}
	// Two JSON example fragments — one per variant. Be tolerant about
	// exact spacing/quoting since the doc is rendered as backtick-quoted
	// Go.
	if !strings.Contains(long, `"type": "manual"`) && !strings.Contains(long, `"type":"manual"`) {
		t.Fatalf("expected manual JSON example in --help:\n%s", long)
	}
	if !strings.Contains(long, `"type": "run_step"`) && !strings.Contains(long, `"type":"run_step"`) {
		t.Fatalf("expected run_step JSON example in --help:\n%s", long)
	}
}

// And the per-flag help string for --source-file must at least gesture
// at the schema so `--help` from a shell tab-completion or short usage
// banner isn't actively misleading.
func TestWorkflowsEvalsCreateSourceFileFlagHelpMentionsBothShapes(t *testing.T) {
	flag := workflowsEvalsCreateCmd.Flag("source-file")
	if flag == nil {
		t.Fatal("workflows evals create missing --source-file flag")
	}
	usage := flag.Usage
	for _, want := range []string{"manual", "run_step"} {
		if !strings.Contains(usage, want) {
			t.Fatalf("--source-file usage should mention %q (got %q)", want, usage)
		}
	}
}
