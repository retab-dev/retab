// Tests for the CLI-DX friction fixes:
//   - `--json` shortcut for `--output json`            (root.go)
//   - `versions get/diff/restore` rejecting a wph_ id  (workflows_versions.go)
//   - an org hint appended to "... not found" 404s     (common.go)
//
// resolveWorkflowVersionRef lives in a `!retab_oagen_cli_workflows` file, so this
// test file carries the same tag to stay out of the prototype (oagen) build.
//go:build !retab_oagen_cli_workflows

package cmd

import (
	"context"
	"net/http"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

// TestOutputJSONFlagShortcut: `--json` routes to `--output json`, and `--json=false`
// is a no-op (never forces json), mirroring the existing `--output-table` shortcut.
func TestOutputJSONFlagShortcut(t *testing.T) {
	out := &outputFlagValue{}
	f := &outputJSONFlagValue{output: out}
	if !f.IsBoolFlag() {
		t.Fatal("--json must present as a bool flag so `--json` needs no value")
	}
	if err := f.Set("true"); err != nil {
		t.Fatalf("Set(true): %v", err)
	}
	if out.value != string(OutputJSON) {
		t.Fatalf("--json should select json output, got %q", out.value)
	}

	out2 := &outputFlagValue{}
	f2 := &outputJSONFlagValue{output: out2}
	if err := f2.Set("false"); err != nil {
		t.Fatalf("Set(false): %v", err)
	}
	if out2.value != "" {
		t.Fatalf("--json=false must not change the format, got %q", out2.value)
	}
	if err := f2.Set("notabool"); err == nil {
		t.Fatal("Set(notabool) must error")
	}
}

// TestResolveWorkflowVersionRefRejectsPublishRecordID: a wph_ publish-record id is
// caught locally with guidance (no server round-trip), while a real ver_ id passes
// straight through. Both paths avoid any client call, so a nil client is safe.
func TestResolveWorkflowVersionRefRejectsPublishRecordID(t *testing.T) {
	_, err := resolveWorkflowVersionRef(context.Background(), nil, "wf_123", "wph_abcDEF")
	if err == nil {
		t.Fatal("a wph_ publish-record id must be rejected with guidance, not passed to the server")
	}
	for _, want := range []string{"publish-record", "workflow_version_id", "versions list"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("guidance should mention %q, got: %v", want, err)
		}
	}

	got, err := resolveWorkflowVersionRef(context.Background(), nil, "wf_123", "ver_realid")
	if err != nil || got != "ver_realid" {
		t.Fatalf("a literal ver_ id must pass through untouched, got (%q, %v)", got, err)
	}
}

// TestRenderAPIErrorAppendsOrgHintOn404NotFound: a 404 whose message reads
// "... not found" gets an org-context hint appended; other statuses do not.
func TestRenderAPIErrorAppendsOrgHintOn404NotFound(t *testing.T) {
	cmd := commandWithDebugFlagForTest(t, false)

	notFound := &retab.APIError{StatusCode: http.StatusNotFound, Message: "Workflow not found"}
	got := renderAPIErrorForCLI(cmd, notFound)
	if !strings.Contains(got, "org switch") || !strings.Contains(strings.ToLower(got), "different organization") {
		t.Fatalf("a 404 not-found should append an org hint, got:\n%s", got)
	}

	badRequest := &retab.APIError{StatusCode: http.StatusBadRequest, Message: "bad thing"}
	if strings.Contains(renderAPIErrorForCLI(cmd, badRequest), "org switch") {
		t.Fatalf("a non-404 error must not get the org hint")
	}
}
