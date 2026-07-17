package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// restoreCmdFor builds a cobra command carrying the flags the restore commands
// declare, with non-terminal stdin so confirmDestructive* takes the refusal
// branch (the branch a scripted caller actually hits).
func restoreCmdFor(use string) *cobra.Command {
	c := &cobra.Command{Use: use}
	c.Flags().BoolP("yes", "y", false, "")
	addConfirmFlag(c)
	c.SetIn(strings.NewReader(""))
	return c
}

// TestRestoreConfirmDoesNotSayDelete pins the verb. The restore commands
// (workflow draft / block version / edge version) overwrite the draft in place;
// they reused confirmDestructive, whose prompt reads "Permanently delete <kind>
// <id>?" — which describes the opposite of what restore does and could scare a
// reader into thinking `versions restore` destroys the workflow.
func TestRestoreConfirmDoesNotSayDelete(t *testing.T) {
	err := confirmDestructiveVerb(restoreCmdFor("restore"), "overwrite", "workflow draft (restoring from ver_1)", "wrk_1")
	if err == nil {
		t.Fatal("expected refusal without --yes in non-interactive mode")
	}
	msg := err.Error()
	if strings.Contains(msg, "delete") {
		t.Fatalf("a restore must not be described as a delete: %s", msg)
	}
	if !strings.Contains(msg, "overwrite") {
		t.Fatalf("restore refusal should name the real verb: %s", msg)
	}
}

// TestRestoreConfirmQuotesTheBareTypableID is the regression that matters. The
// `id` argument is BOTH displayed and the exact string confirmDestructive
// compares the typed answer against. `versions restore` passed
// fmt.Sprintf("%s from %s", workflowID, versionID), so the TTY prompt demanded
// the user type "wrk_x from ver_y" verbatim to confirm — unusable — and the
// non-TTY refusal quoted the whole phrase as if it were an id.
func TestRestoreConfirmQuotesTheBareTypableID(t *testing.T) {
	err := confirmDestructiveVerb(restoreCmdFor("restore"), "overwrite", "workflow draft (restoring from ver_1)", "wrk_1")
	if err == nil {
		t.Fatal("expected refusal")
	}
	msg := err.Error()
	// The quoted token is what the prompt asks the user to type back: it must be
	// the bare id, never a phrase.
	if !strings.Contains(msg, `"wrk_1"`) {
		t.Fatalf("the quoted token must be the bare workflow id: %s", msg)
	}
	if strings.Contains(msg, `"wrk_1 from ver_1"`) || strings.Contains(msg, "wrk_1 from ver_1") {
		t.Fatalf("the typable token must not be a compound phrase: %s", msg)
	}
	// The version being restored from is still reported — as context, not as
	// something to type.
	if !strings.Contains(msg, "ver_1") {
		t.Fatalf("refusal should still name the source version as context: %s", msg)
	}
}

// TestRestoreConfirmStillGates guards the safety property: renaming the verb
// must not weaken the gate. Without --yes/--confirm and without a TTY, restore
// still refuses; with either ack flag it proceeds.
func TestRestoreConfirmStillGates(t *testing.T) {
	if err := confirmDestructiveVerb(restoreCmdFor("restore"), "overwrite", "workflow draft", "wrk_1"); err == nil {
		t.Fatal("restore must refuse without --yes/--confirm in non-interactive mode")
	}
	cy := restoreCmdFor("restore")
	if err := cy.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	if err := confirmDestructiveVerb(cy, "overwrite", "workflow draft", "wrk_1"); err != nil {
		t.Fatalf("--yes should satisfy the restore gate: %v", err)
	}
	cc := restoreCmdFor("restore")
	if err := cc.Flags().Set("confirm", "true"); err != nil {
		t.Fatal(err)
	}
	if err := confirmDestructiveVerb(cc, "overwrite", "workflow draft", "wrk_1"); err != nil {
		t.Fatalf("--confirm should satisfy the restore gate: %v", err)
	}
}

// TestConfirmDestructiveStillSaysDelete pins that the delete callers are
// untouched: confirmDestructive keeps its exact wording, so this refactor is
// additive rather than a rename of every delete prompt.
func TestConfirmDestructiveStillSaysDelete(t *testing.T) {
	err := confirmDestructive(restoreCmdFor("delete"), "workflow", "wrk_1")
	if err == nil {
		t.Fatal("expected refusal")
	}
	if !strings.Contains(err.Error(), `refusing to delete workflow "wrk_1"`) {
		t.Fatalf("delete wording must be unchanged, got: %s", err.Error())
	}
}
