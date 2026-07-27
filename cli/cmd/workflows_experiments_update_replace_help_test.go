//go:build !retab_oagen_cli_workflows_experiments

package cmd

import (
	"strings"
	"testing"
)

// `experiments update` REPLACES the experiment's document set — the PATCH $sets
// `documents` from the request body alone and never merges what is already
// stored. The help said the opposite:
//
//	# Add more documents from production
//	retab workflows experiments update exp_pqr678 --captures-file ./more-captures.json
//
// Following that example against an experiment holding two documents left it
// holding one, and the next run scored that single document — the other two, and
// the consensus passes already paid for, were gone with no warning. The example is
// the part users copy, so it must not describe a destructive operation as additive.
func TestWorkflowsExperimentsUpdateExampleDoesNotPromiseAdditiveDocuments(t *testing.T) {
	example := workflowsExperimentsUpdateCmd.Example
	if example == "" {
		t.Fatal("workflowsExperimentsUpdateCmd.Example is empty")
	}
	if strings.Contains(strings.ToLower(example), "add more documents") {
		t.Fatalf("update --help still advertises additive document semantics:\n%s", example)
	}
	if !strings.Contains(example, "REPLACE") {
		t.Fatalf("update --help example must state that the document set is replaced:\n%s", example)
	}
}

// The Long block already described replacement; keep it that way so the example
// and the prose cannot drift apart again.
func TestWorkflowsExperimentsUpdateLongDocumentsReplacement(t *testing.T) {
	long := workflowsExperimentsUpdateCmd.Long
	for _, want := range []string{"replace", "invalidates"} {
		if !strings.Contains(strings.ToLower(long), want) {
			t.Fatalf("expected update --help to document replacement (%q):\n%s", want, long)
		}
	}
}

// Each document flag on `update` must carry the replacement warning on its own
// line: the flag list is what `--help` shows, and these three flags share their
// wording with `create`, where "capture a document" correctly reads as additive.
func TestWorkflowsExperimentsUpdateDocumentFlagsSayTheyReplace(t *testing.T) {
	for _, name := range []string{"capture", "captures-file", "documents-file"} {
		flag := workflowsExperimentsUpdateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("update has no --%s flag", name)
		}
		if !strings.Contains(flag.Usage, "REPLACES the document set") {
			t.Fatalf("--%s usage does not warn that it replaces the document set: %q", name, flag.Usage)
		}
	}
}

// The same flags on `create` must NOT carry the warning — there is nothing to
// replace when the experiment is being created, and the shared registration
// helper makes it easy to overwrite both commands' wording at once.
func TestWorkflowsExperimentsCreateDocumentFlagsAreNotMarkedAsReplacing(t *testing.T) {
	for _, name := range []string{"capture", "captures-file", "documents-file"} {
		flag := workflowsExperimentsCreateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("create has no --%s flag", name)
		}
		if strings.Contains(flag.Usage, "REPLACES") {
			t.Fatalf("--%s on create wrongly warns about replacement: %q", name, flag.Usage)
		}
	}
}
