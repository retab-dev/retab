package cmd

import (
	"strings"
	"testing"
)

func TestWorkflowsBlocksCreateHelpShowsExtractReviewConfig(t *testing.T) {
	help := workflowsBlocksCreateCmd.Long + "\n" + workflowsBlocksCreateCmd.Example

	for _, want := range []string{
		`"type": "extract"`,
		`"review"`,
		`"predicate"`,
		`"kind": "always"`,
		`"inputs": [{"name": "document", "type": "file", "is_primary": true}]`,
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("blocks create help should show review config fragment %q, got:\n%s", want, help)
		}
	}
}
