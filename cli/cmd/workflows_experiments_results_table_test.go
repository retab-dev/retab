package cmd

import (
	"strings"
	"testing"
)

// TestExperimentResultsListTableRendersRichColumns pins that the dedicated
// results column spec surfaces the per-document fields (document, block kind,
// status, duration) instead of the old generic ID+TYPE pair.
func TestExperimentResultsListTableRendersRichColumns(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":          "expjob_x",
				"document_id": "expdoc_y",
				"block_type":  "extract",
				"lifecycle":   map[string]any{"status": "completed"},
				"timing":      map[string]any{"duration_ms": 42},
			},
		},
	}
	var buf strings.Builder
	if err := RenderList(&buf, OutputTable, resource, experimentResultColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"DOCUMENT", "BLOCK_KIND", "STATUS", "DURATION_MS",
		"expdoc_y", "extract", "completed", "42",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("results table missing %q:\n%s", want, out)
		}
	}
}
