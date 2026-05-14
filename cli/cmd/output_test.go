package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// fileItem is a tiny test fixture mirroring the SDK's list-response shape
// — `Data []T` plus a few sibling fields — so the reflect path in
// extractDataSlice gets exercised against a real typed struct, not just
// map[string]any.
type fileItem struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

type fileList struct {
	Data    []fileItem `json:"data"`
	HasMore bool       `json:"has_more"`
}

// TestDefaultOutputFormatNonFile pins the "non-TTY falls back to JSON"
// half of DefaultOutputFormat. A bytes.Buffer isn't an *os.File and never
// will be, so this case is unambiguous and stable.
func TestDefaultOutputFormatNonFile(t *testing.T) {
	got := DefaultOutputFormat(&bytes.Buffer{})
	if got != OutputJSON {
		t.Fatalf("want OutputJSON for non-file writer, got %q", got)
	}
}

// TestRenderListJSONBytesPin enforces that RenderList(OutputJSON, …)
// produces exactly the same bytes as the historical encoder pattern used
// by printJSON / printNDJSON in common.go. If this drifts, anything piping
// `retab files list` into jq will see a behavioural change — that's a
// breaking change and the test should fail loudly first.
func TestRenderListJSONBytesPin(t *testing.T) {
	list := fileList{
		Data: []fileItem{
			{ID: "f_1", Filename: "a.pdf"},
			{ID: "f_2", Filename: "b.pdf"},
		},
		HasMore: false,
	}

	// Reference encoder — copied verbatim from printJSON in common.go.
	var want bytes.Buffer
	enc := json.NewEncoder(&want)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(list); err != nil {
		t.Fatalf("reference encode failed: %v", err)
	}

	var got bytes.Buffer
	if err := RenderList(&got, OutputJSON, list, nil); err != nil {
		t.Fatalf("RenderList failed: %v", err)
	}

	if got.String() != want.String() {
		t.Fatalf("JSON output drifted from printJSON\n--- want ---\n%s\n--- got ---\n%s",
			want.String(), got.String())
	}
}

// TestRenderListTableTypedStruct exercises the reflect path: we hand the
// renderer the typed struct directly (no json round-trip) and check the
// header + row layout. Two-space minimum separation is tabwriter's
// configured padding; we assert it explicitly so a future flag change
// doesn't silently collapse the columns.
//
// Header style choice: plain text, no bold / no underline. The renderer
// is going to be used in pipes and terminals that don't speak ANSI; a
// plain header keeps the output paste-able into issues, markdown, and
// pagers without escape-code noise. The "header is just the first line"
// convention is what kubectl, gh, and stripe-cli all do.
func TestRenderListTableTypedStruct(t *testing.T) {
	list := fileList{
		Data: []fileItem{
			{ID: "f_1", Filename: "alpha.pdf"},
			{ID: "f_22", Filename: "beta.pdf"},
		},
	}
	cols := []TableColumn{
		{Header: "ID", Extract: func(r any) string { return r.(fileItem).ID }},
		{Header: "FILENAME", Extract: func(r any) string { return r.(fileItem).Filename }},
	}

	var buf bytes.Buffer
	if err := RenderList(&buf, OutputTable, list, cols); err != nil {
		t.Fatalf("RenderList table failed: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d:\n%s", len(lines), out)
	}

	// Header content and column order.
	if !strings.HasPrefix(lines[0], "ID") || !strings.Contains(lines[0], "FILENAME") {
		t.Fatalf("header line malformed: %q", lines[0])
	}
	// Column separation: tabwriter's padding=2 means at least two spaces
	// between the rightmost char of column N and the leftmost of N+1.
	if !strings.Contains(lines[0], "  ") {
		t.Fatalf("header columns not separated by >= 2 spaces: %q", lines[0])
	}

	// Row content lives in the right column.
	if !strings.Contains(lines[1], "f_1") || !strings.Contains(lines[1], "alpha.pdf") {
		t.Fatalf("row 1 missing data: %q", lines[1])
	}
	if !strings.Contains(lines[2], "f_22") || !strings.Contains(lines[2], "beta.pdf") {
		t.Fatalf("row 2 missing data: %q", lines[2])
	}
}

// TestRenderListTableMapFallback exercises the JSON-round-trip path used
// when the input isn't a struct with a Data field — e.g. payloads coming
// back from readJSON. The Extract func sees the row as map[string]any.
func TestRenderListTableMapFallback(t *testing.T) {
	list := map[string]any{
		"data": []any{
			map[string]any{"id": "x_1", "filename": "one.pdf"},
			map[string]any{"id": "x_2", "filename": "two.pdf"},
		},
	}
	cols := []TableColumn{
		{Header: "ID", Extract: func(r any) string { return r.(map[string]any)["id"].(string) }},
		{Header: "FILENAME", Extract: func(r any) string { return r.(map[string]any)["filename"].(string) }},
	}

	var buf bytes.Buffer
	if err := RenderList(&buf, OutputTable, list, cols); err != nil {
		t.Fatalf("RenderList table failed: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ID", "FILENAME", "x_1", "one.pdf", "x_2", "two.pdf"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in table output, got:\n%s", want, out)
		}
	}
}

// TestRenderListTableEmpty pins the "no rows" behaviour: header row only,
// no error. An empty list is a valid response and the user benefits from
// seeing the headers so they know the column shape.
func TestRenderListTableEmpty(t *testing.T) {
	list := fileList{Data: nil}
	cols := []TableColumn{
		{Header: "ID", Extract: func(r any) string { return r.(fileItem).ID }},
	}
	var buf bytes.Buffer
	if err := RenderList(&buf, OutputTable, list, cols); err != nil {
		t.Fatalf("RenderList table empty failed: %v", err)
	}
	out := strings.TrimRight(buf.String(), "\n")
	if out != "ID" {
		t.Fatalf("want bare header 'ID', got %q", out)
	}
}

// fakeRoot builds a cobra command tree with a persistent --output flag.
// Uses a plain string flag (not the validating outputFlagValue) so the
// "bogus" case still reaches ResolveOutputFormat — we need to assert
// ResolveOutputFormat's own error branch independently of the parse-time
// guard installed by root.go. The two layers defend the same property
// from different sides.
func fakeRoot() *cobra.Command {
	root := &cobra.Command{Use: "retab"}
	root.PersistentFlags().String("output", "", "output format")
	child := &cobra.Command{Use: "list"}
	root.AddCommand(child)
	return child
}

func TestResolveOutputFormat(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		want    OutputFormat
		wantErr bool
	}{
		// Empty flag → DefaultOutputFormat(buf) → OutputJSON because
		// bytes.Buffer isn't a *os.File.
		{"empty defaults to json on buffer", "", OutputJSON, false},
		{"auto defaults to json on buffer", "auto", OutputJSON, false},
		{"json explicit", "json", OutputJSON, false},
		{"table explicit", "table", OutputTable, false},
		{"bad value rejected", "bogus", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := fakeRoot()
			if err := cmd.Root().PersistentFlags().Set("output", tc.value); err != nil {
				t.Fatalf("set flag: %v", err)
			}
			got, err := ResolveOutputFormat(cmd, &bytes.Buffer{})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error for %q, got %q", tc.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
