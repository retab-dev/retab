package cmd

import (
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

// mimeDataFromDocument must preserve inline content/mime-type for a local-file
// descriptor; copying only filename/url sent an empty document over the wire.
func TestMimeDataFromDocumentPreservesContent(t *testing.T) {
	doc := retab.MIMEData{Filename: "a.pdf", Content: "aGVsbG8=", MIMEType: "application/pdf"}
	got, err := mimeDataFromDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != doc.Content || got.MIMEType != doc.MIMEType || got.Filename != doc.Filename {
		t.Fatalf("content/mime-type dropped: %+v", got)
	}

	gotPtr, err := mimeDataFromDocument(&doc)
	if err != nil {
		t.Fatal(err)
	}
	if gotPtr.Content != doc.Content || gotPtr.MIMEType != doc.MIMEType {
		t.Fatalf("pointer form dropped content/mime-type: %+v", gotPtr)
	}
}

// fileBase must strip Windows backslash separators, not just '/'.
func TestFileBaseSeparators(t *testing.T) {
	cases := map[string]string{
		`C:\docs\report.pdf`:       "report.pdf",
		"/home/user/report.pdf":    "report.pdf",
		"https://x.test/a/b/c.pdf": "c.pdf",
		"plain.pdf":                "plain.pdf",
	}
	for in, want := range cases {
		if got := fileBase(in); got != want {
			t.Errorf("fileBase(%q) = %q, want %q", in, got, want)
		}
	}
}

// pageNumberFromScreenshotFilename extracts the document page from page-N.ext.
func TestPageNumberFromScreenshotFilename(t *testing.T) {
	cases := map[string]int{
		"page-1.png":           1,
		"page-12.jpg":          12,
		"/tmp/out/page-3.webp": 3,
		`C:\out\page-7.png`:    7,
		"thumbnail.png":        0,
		"page-.png":            0,
		"page-abc.png":         0,
	}
	for in, want := range cases {
		if got := pageNumberFromScreenshotFilename(in); got != want {
			t.Errorf("pageNumberFromScreenshotFilename(%q) = %d, want %d", in, got, want)
		}
	}
}

// notifierSkippableCommand must only match the real subcommand, not an
// argument value that happens to equal a keyword.
func TestNotifierSkippableCommand(t *testing.T) {
	skip := [][]string{
		{"version"},
		{"update"},
		{"completion", "bash"},
		{"--global-flag", "completion"}, // value-less flags are skipped over
		{"--version"},
		{"-v"},
	}
	for _, args := range skip {
		if !notifierSkippableCommand(args) {
			t.Errorf("expected skip for %v", args)
		}
	}
	keep := [][]string{
		{"files", "get", "version"}, // "version" is an arg, not the command
		{"workflows", "runs", "list"},
		{"files", "list", "--after", "update"},
	}
	for _, args := range keep {
		if notifierSkippableCommand(args) {
			t.Errorf("expected NO skip for %v", args)
		}
	}
}

// redactKey must never reveal more than half of a short credential.
func TestRedactKeyShortCredential(t *testing.T) {
	for _, key := range []string{"abcdefghi", "abcdefghijkl", "retab_sk_abcd1234"} {
		got := redactKey(key)
		if len(got) < len(key) {
			// fixed-width mask may shorten very long keys; short keys keep length
			continue
		}
		revealed := 0
		for _, c := range got {
			if c != '*' {
				revealed++
			}
		}
		if revealed*2 > len(key) {
			t.Errorf("redactKey(%q)=%q revealed %d of %d chars (>half)", key, got, revealed, len(key))
		}
	}
	if got := redactKey("short"); got != "*****" {
		t.Errorf("redactKey(short) = %q", got)
	}
}

// workflowASCIISyntheticLevels must terminate on a cyclic edge set rather than
// looping forever.
func TestWorkflowSyntheticLevelsTerminatesOnCycle(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		{ID: "a", Type: retab.WorkflowBlockType("start")},
		{ID: "b", Type: retab.WorkflowBlockType("extract")},
		{ID: "c", Type: retab.WorkflowBlockType("extract")},
	}
	edges := []retab.WorkflowEdgeDoc{
		{SourceBlock: "a", TargetBlock: "b"},
		{SourceBlock: "b", TargetBlock: "c"},
		{SourceBlock: "c", TargetBlock: "a"}, // cycle back to start
	}
	// Completing at all proves termination; the package test timeout would
	// otherwise fire on the old unbounded loop.
	levels := workflowASCIISyntheticLevels(blocks, edges)
	for id, lvl := range levels {
		if lvl < 0 || lvl >= len(blocks) {
			t.Errorf("block %q has out-of-bounds level %d", id, lvl)
		}
	}
}

// Sanity: redactKey still shows first/last 4 for long tokens.
func TestRedactKeyLongToken(t *testing.T) {
	long := "abcd" + strings.Repeat("x", 1000) + "wxyz"
	got := redactKey(long)
	if !strings.HasPrefix(got, "abcd") || !strings.HasSuffix(got, "wxyz") {
		t.Errorf("redactKey long token = %q", got)
	}
}
