//go:build !retab_oagen_cli_workflows_edges

package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func specEdgeObj(t *testing.T, body string) map[string]any {
	t.Helper()
	var obj map[string]any
	if err := json.Unmarshal([]byte(body), &obj); err != nil {
		t.Fatalf("bad test json: %v", err)
	}
	return obj
}

// All three spec edge spellings must normalize to identical create params, so
// the same fragment is portable between a spec file and `edges create --edge-file`.
func TestParseSpecEdgeObjectAcceptsAllSpellings(t *testing.T) {
	cases := map[string]string{
		"canonical": `{"source":{"block":"start","handle":"output-file-0"},"target":{"block":"pull","handle":"input-file-document"}}`,
		"legacy":    `{"from":{"block":"start","handle":"output-file-0"},"to":{"block":"pull","handle":"input-file-document"}}`,
		"flat":      `{"source_block":"start","source_handle":"output-file-0","target_block":"pull","target_handle":"input-file-document"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			req, err := parseSpecEdgeObject(specEdgeObj(t, body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if req.SourceBlock != "start" || req.TargetBlock != "pull" {
				t.Fatalf("blocks: got (%q,%q)", req.SourceBlock, req.TargetBlock)
			}
			if req.SourceHandle == nil || *req.SourceHandle != "output-file-0" {
				t.Fatalf("source handle: got %v", req.SourceHandle)
			}
			if req.TargetHandle == nil || *req.TargetHandle != "input-file-document" {
				t.Fatalf("target handle: got %v", req.TargetHandle)
			}
		})
	}
}

func TestParseSpecEdgeObjectRejectsUnknownAndConflicting(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"conflicting nested+flat": {
			`{"source":{"block":"start","handle":"h"},"source_block":"start","target":{"block":"pull","handle":"h"}}`,
			"both nested and flat",
		},
		"source given twice": {
			`{"source":{"block":"start","handle":"h"},"from":{"block":"start","handle":"h"},"target":{"block":"pull","handle":"h"}}`,
			"given twice",
		},
		"missing target": {
			`{"source":{"block":"start","handle":"h"}}`,
			"missing the target endpoint",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := parseSpecEdgeObject(specEdgeObj(t, tc.body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}
