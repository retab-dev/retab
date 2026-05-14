package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// makeCmdWithWorkflowIDFlag mimics the flag wiring used by the workflows
// subcommands that have migrated workflow-id from a required flag to a
// positional argument with a deprecated --workflow-id fallback.
func makeCmdWithWorkflowIDFlag() *cobra.Command {
	c := &cobra.Command{Use: "fake"}
	c.Flags().String("workflow-id", "", "workflow id (deprecated; pass as positional)")
	_ = c.Flags().MarkDeprecated("workflow-id", "use the positional argument instead")
	return c
}

func TestResolveWorkflowIDArg_PositionalAlone(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags(nil); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, []string{"wf_abc"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_abc" {
		t.Fatalf("got %q want wf_abc", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestResolveWorkflowIDArg_FlagAloneEmitsWarning(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags([]string{"--workflow-id", "wf_flag"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_flag" {
		t.Fatalf("got %q want wf_flag", got)
	}
	if !strings.Contains(warn.String(), "deprecated") {
		t.Fatalf("expected deprecation warning, got %q", warn.String())
	}
}

func TestResolveWorkflowIDArg_PositionalWinsOverFlag(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags([]string{"--workflow-id", "wf_flag"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	got, err := resolveWorkflowIDArgTo(c, []string{"wf_pos"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wf_pos" {
		t.Fatalf("got %q want wf_pos (positional must win)", got)
	}
	if !strings.Contains(warn.String(), "deprecated") {
		t.Fatalf("expected deprecation warning when both supplied, got %q", warn.String())
	}
}

func TestResolveWorkflowIDArg_NeitherErrors(t *testing.T) {
	c := makeCmdWithWorkflowIDFlag()
	if err := c.ParseFlags(nil); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var warn bytes.Buffer
	_, err := resolveWorkflowIDArgTo(c, nil, &warn)
	if err == nil {
		t.Fatalf("expected error when neither positional nor flag is set")
	}
	if !strings.Contains(err.Error(), "workflow id required") {
		t.Fatalf("expected 'workflow id required' message, got %q", err.Error())
	}
}
