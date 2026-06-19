package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// TestTruncateCellRunesKeepsValidUTF8 pins the rune-based truncation: byte
// slicing a multi-byte string mid-rune produced invalid UTF-8 in table/CSV
// output. The cut must land on a rune boundary.
func TestTruncateCellRunesKeepsValidUTF8(t *testing.T) {
	s := strings.Repeat("é", 60) // 60 runes, 120 bytes
	got := truncateCellRunes(s, 40)
	if !utf8.ValidString(got) {
		t.Fatalf("truncated output is not valid UTF-8: %q", got)
	}
	if n := utf8.RuneCountInString(got); n != 41 { // 40 runes + ellipsis
		t.Fatalf("rune count = %d, want 41 (40 + ellipsis)", n)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("want ellipsis suffix, got %q", got)
	}
	if truncateCellRunes("abc", 40) != "abc" {
		t.Fatal("short string should pass through unchanged")
	}
	if truncateCellRunes("anything", 0) != "anything" {
		t.Fatal("non-positive limit should pass through unchanged")
	}
}

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

// TestRenderListCSVTypedStruct pins that RenderList(OutputCSV, …) emits
// RFC 4180 CSV using the same columns as the table renderer. Before this,
// `--output csv` silently fell back to JSON on some list commands and
// errored ("unknown output format") on others.
func TestRenderListCSVTypedStruct(t *testing.T) {
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
	if err := RenderList(&buf, OutputCSV, list, cols); err != nil {
		t.Fatalf("RenderList csv failed: %v", err)
	}

	want := "ID,FILENAME\nf_1,alpha.pdf\nf_22,beta.pdf\n"
	if buf.String() != want {
		t.Fatalf("csv output mismatch:\n got: %q\nwant: %q", buf.String(), want)
	}
}

// TestRenderListCSVQuotesSpecialValues pins that encoding/csv quoting is
// applied — a cell containing a comma or quote must not corrupt columns.
func TestRenderListCSVQuotesSpecialValues(t *testing.T) {
	list := fileList{Data: []fileItem{{ID: "f_1", Filename: "a,b \"c\".pdf"}}}
	cols := []TableColumn{
		{Header: "ID", Extract: func(r any) string { return r.(fileItem).ID }},
		{Header: "FILENAME", Extract: func(r any) string { return r.(fileItem).Filename }},
	}

	var buf bytes.Buffer
	if err := RenderList(&buf, OutputCSV, list, cols); err != nil {
		t.Fatalf("RenderList csv failed: %v", err)
	}

	want := "ID,FILENAME\nf_1,\"a,b \"\"c\"\".pdf\"\n"
	if buf.String() != want {
		t.Fatalf("csv quoting mismatch:\n got: %q\nwant: %q", buf.String(), want)
	}
}

// TestPrintResultCSVAutoColumns pins the generic auto-column CSV path used
// by commands that route through printResult (e.g. `workflows list`).
func TestPrintResultCSVAutoColumns(t *testing.T) {
	list := map[string]any{
		"data": []any{
			map[string]any{"id": "wrk_1", "name": "Alpha"},
			map[string]any{"id": "wrk_2", "name": "Beta"},
		},
	}
	stdout, _ := captureStd(t, func() {
		if err := printResultCSV(list); err != nil {
			t.Fatalf("printResultCSV failed: %v", err)
		}
	})
	if !strings.HasPrefix(stdout, "ID,NAME\n") {
		t.Fatalf("expected CSV header ID,NAME, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "wrk_1,Alpha\n") || !strings.Contains(stdout, "wrk_2,Beta\n") {
		t.Fatalf("expected CSV rows for both workflows, got:\n%s", stdout)
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
	// renderTable emits its zero-rows hint to stderr (R2-9) so the
	// buffered stdout under test still ends as a clean header line.
	// Capture stderr separately and discard for this assertion — the
	// hint behaviour is pinned by TestRenderListEmitsEmptyHintForZeroRows.
	_, _ = captureStd(t, func() {
		if err := RenderList(&buf, OutputTable, list, cols); err != nil {
			t.Fatalf("RenderList table empty failed: %v", err)
		}
	})
	out := strings.TrimRight(buf.String(), "\n")
	if out != "ID" {
		t.Fatalf("want bare header 'ID', got %q", out)
	}
}

// TestRenderListEmitsEmptyHintForZeroRows pins R2-9: a list rendered as a
// table with zero data rows emits "(no rows)" to stderr so the lone
// header row isn't mistaken for "all rows hidden". Stdout stays clean
// for piping — the hint is for humans staring at a terminal, never for
// downstream consumers.
func TestRenderListEmitsEmptyHintForZeroRows(t *testing.T) {
	list := fileList{Data: nil}
	cols := []TableColumn{
		{Header: "ID", Extract: func(r any) string { return r.(fileItem).ID }},
	}
	var stdoutBuf bytes.Buffer
	_, stderr := captureStd(t, func() {
		if err := RenderList(&stdoutBuf, OutputTable, list, cols); err != nil {
			t.Fatalf("RenderList table empty failed: %v", err)
		}
	})
	if !strings.Contains(stderr, "(no rows)") {
		t.Fatalf("expected (no rows) hint on stderr, got: %q", stderr)
	}
	if strings.Contains(stdoutBuf.String(), "(no rows)") {
		t.Fatalf("(no rows) hint should not appear in stdout buffer, got:\n%s", stdoutBuf.String())
	}
	if !strings.Contains(stdoutBuf.String(), "ID") {
		t.Fatalf("expected ID header in stdout buffer, got:\n%s", stdoutBuf.String())
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

// captureStd swaps os.Stdout and os.Stderr for pipes, runs fn, and
// returns whatever fn wrote. Used by the printResultTable / printResult
// tests because those helpers write to the real stdout/stderr (this
// matches how every other "print*" helper in common.go works — they
// don't accept an io.Writer arg).
//
// We use os.Pipe rather than a bytes.Buffer because os.Stdout/Stderr
// are *os.File-typed; we can't reassign them to a buffer without
// breaking other code that reflects on them.
func captureStd(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr
	t.Cleanup(func() {
		os.Stdout, os.Stderr = origOut, origErr
	})

	// Drain the pipes in background so a verbose render can't block on
	// the pipe buffer.
	outCh := make(chan string, 1)
	errCh := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(rOut)
		outCh <- string(b)
	}()
	go func() {
		b, _ := io.ReadAll(rErr)
		errCh <- string(b)
	}()

	fn()
	_ = wOut.Close()
	_ = wErr.Close()
	return <-outCh, <-errCh
}

// TestPrintResultTableDataWrappedList covers acceptance criterion #1:
// a {"data": [...]} resource shape with id / name / created_at →
// auto-column table with header + N data rows, columns lined up.
//
// We assert structurally (header present, every row appears, columns
// separated by ≥ 2 spaces — tabwriter's padding setting) rather than
// pinning the exact whitespace. The latter would couple this test to
// tabwriter's internal alignment math and break on any width change.
func TestPrintResultTableDataWrappedList(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{
				"id":         "f_1",
				"name":       "alpha.pdf",
				"created_at": "2026-01-01T00:00:00Z",
			},
			map[string]any{
				"id":         "f_22",
				"name":       "beta.pdf",
				"created_at": "2026-01-02T00:00:00Z",
			},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(resource); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr (data-shape should be tabulable): %q", stderr)
	}

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines (header + 2 rows), got %d:\n%s", len(lines), stdout)
	}
	// Header — ID, NAME (from name alias), CREATED_AT must appear in order.
	for _, h := range []string{"ID", "NAME", "CREATED_AT"} {
		if !strings.Contains(lines[0], h) {
			t.Fatalf("header missing %q: %q", h, lines[0])
		}
	}
	if strings.Index(lines[0], "ID") > strings.Index(lines[0], "NAME") {
		t.Fatalf("ID should come before NAME: %q", lines[0])
	}
	if strings.Index(lines[0], "NAME") > strings.Index(lines[0], "CREATED_AT") {
		t.Fatalf("NAME should come before CREATED_AT: %q", lines[0])
	}
	// Columns separated by ≥ 2 spaces (tabwriter padding=2).
	if !strings.Contains(lines[0], "  ") {
		t.Fatalf("header columns not space-separated: %q", lines[0])
	}
	// Each row contains its id + filename.
	if !strings.Contains(lines[1], "f_1") || !strings.Contains(lines[1], "alpha.pdf") {
		t.Fatalf("row 1 missing data: %q", lines[1])
	}
	if !strings.Contains(lines[2], "f_22") || !strings.Contains(lines[2], "beta.pdf") {
		t.Fatalf("row 2 missing data: %q", lines[2])
	}
}

// TestPrintResultTableBareArray covers acceptance criterion #2:
// a bare [] of objects (no wrapping {"data": ...}) renders the same
// way as the wrapped shape. Catches the reflect-on-top-level-slice
// branch in extractTabulableRows.
func TestPrintResultTableBareArray(t *testing.T) {
	bare := []any{
		map[string]any{"id": "x_1", "filename": "one.pdf"},
		map[string]any{"id": "x_2", "filename": "two.pdf"},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(bare); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr for bare array: %q", stderr)
	}
	for _, want := range []string{"ID", "NAME", "x_1", "one.pdf", "x_2", "two.pdf"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in output:\n%s", want, stdout)
		}
	}
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d:\n%s", len(lines), stdout)
	}
}

type stepItem struct {
	ID        string     `json:"id"`
	StepID    string     `json:"step_id"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
}

type stepList struct {
	Data []stepItem `json:"data"`
}

func TestPrintResultTableTypedNilPointerCellDoesNotPanic(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 10, 15, 49, 0, time.UTC)
	steps := stepList{
		Data: []stepItem{
			{ID: "step_1", StepID: "run_1_block_1", CreatedAt: nil, StartedAt: &startedAt},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(steps); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr for typed list: %q", stderr)
	}
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "step_1") {
		t.Fatalf("expected table output with step id, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "2026-05-15") {
		t.Fatalf("expected fallback timestamp alias in table output, got:\n%s", stdout)
	}
}

func TestPrintResultTableWorkflowStepColumns(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 10, 15, 49, 0, time.UTC)
	steps := map[string]any{
		"data": []any{
			map[string]any{
				"step_id":     "run_1_block_extract",
				"block_label": "Extract invoice",
				"block_type":  "extract",
				"started_at":  startedAt.Format(time.RFC3339),
				"model":       "retab-small",
			},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(steps); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr for workflow step list: %q", stderr)
	}
	// Workflow steps carry `started_at` (no `created_at`) — the auto
	// renderer must label the timestamp column STARTED_AT, not
	// CREATED_AT, so the header is honest about which field is shown.
	for _, want := range []string{"ID", "NAME", "TYPE", "MODEL", "STARTED_AT", "run_1_block_extract", "Extract invoice", "extract", "retab-small", "2026-05-15"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in workflow step table, got:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "CREATED_AT") {
		t.Fatalf("workflow step rows expose started_at only; CREATED_AT header would be a lie:\n%s", stdout)
	}
}

func TestPrintResultTableWorkflowRunNestedColumns(t *testing.T) {
	runs := map[string]any{
		"data": []any{
			map[string]any{
				"id": "run_1",
				"workflow": map[string]any{
					"name_at_run_time": "Invoice workflow",
				},
				"lifecycle": map[string]any{
					"status": "completed",
				},
				"timing": map[string]any{
					"created_at": "2026-05-15T10:15:48Z",
				},
			},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(runs); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr for workflow run list: %q", stderr)
	}
	for _, want := range []string{"ID", "NAME", "TYPE", "CREATED_AT", "run_1", "Invoice workflow", "completed", "2026-05-15"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in workflow run table, got:\n%s", want, stdout)
		}
	}
}

// TestPrintResultTableEnvelopeFieldNavigable pins the table renderer against
// a *typed* row whose `lifecycle` is a discriminated-union envelope, not a
// map[string]any. The envelope keeps its payload in an unexported
// `raw json.RawMessage` reachable only through MarshalJSON, so the dotted-path
// alias `lifecycle.status` can only resolve if rowFieldSingle falls back to
// marshalling the value and looking the key up in the JSON object. Without that
// fallback the TYPE column would silently go blank for every list endpoint
// whose rows carry a union field — exactly the regression this guards.
func TestPrintResultTableEnvelopeFieldNavigable(t *testing.T) {
	status := "completed"
	type runRow struct {
		ID        string                     `json:"id"`
		Lifecycle retab.WorkflowRunLifecycle `json:"lifecycle"`
	}
	type runList struct {
		Data []runRow `json:"data"`
	}
	runs := runList{Data: []runRow{{
		ID: "run_1",
		// Build a coherent completed variant: the union's From* constructors force
		// the "status" discriminator to match the chosen variant, so the asserted
		// value must be the real discriminator ("completed"), not an arbitrary
		// string pinned onto a mismatched variant.
		Lifecycle: retab.WorkflowRunLifecycleFromCompletedTerminal(retab.CompletedTerminal{Status: &status}),
	}}}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(runs); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr for typed envelope run list: %q", stderr)
	}
	for _, want := range []string{"ID", "TYPE", "run_1", "completed"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in typed-envelope run table, got:\n%s", want, stdout)
		}
	}
}

func TestPrintResultTableWorkflowEdgeColumns(t *testing.T) {
	edges := map[string]any{
		"data": []any{
			map[string]any{
				"id":           "edge_1",
				"source_block": "start",
				"target_block": "extract",
				"updated_at":   "2026-05-15T10:15:13Z",
			},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(edges); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr for workflow edge list: %q", stderr)
	}
	// Edges carry `updated_at` only (no `created_at`) — header must
	// be UPDATED_AT to reflect the field actually being rendered.
	for _, want := range []string{"ID", "SOURCE", "TARGET", "UPDATED_AT", "edge_1", "start", "extract", "2026-05-15"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in workflow edge table, got:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "CREATED_AT") {
		t.Fatalf("edge rows expose updated_at only; CREATED_AT header would be a lie:\n%s", stdout)
	}
}

func TestPrintResultTableSkipsNestedObjectCells(t *testing.T) {
	tests := map[string]any{
		"data": []any{
			map[string]any{
				"id": "wfnodeeval_1",
				"source": map[string]any{
					"type": "manual",
					"handle_inputs": map[string]any{
						"input-json-0": map[string]any{"type": "json"},
					},
				},
				"target": map[string]any{
					"type":     "block",
					"block_id": "block_calc_total",
				},
				"name":       "calc-total-baseline",
				"created_at": "2026-05-15T13:12:16Z",
			},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(tests); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr for workflow eval list: %q", stderr)
	}
	for _, want := range []string{"ID", "SOURCE", "TARGET", "NAME", "CREATED_AT", "wfnodeeval_1", "manual", "block_calc_total", "calc-total-baseline"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in workflow eval table, got:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "map[") {
		t.Fatalf("nested object leaked into table output:\n%s", stdout)
	}
}

// TestPrintResultTableSingleObjectFallsBackToJSON covers acceptance
// criterion #3: a single object (not a list) isn't tabulable, so the
// helper warns on stderr and falls back to JSON on stdout.
//
// Important: the JSON on stdout must remain valid JSON so existing
// `| jq` pipelines keep working when a user has --output table set
// globally but happens to hit a non-list response.
func TestPrintResultTableSingleObjectFallsBackToJSON(t *testing.T) {
	obj := map[string]any{"id": "single", "filename": "lone.pdf"}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(obj); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if !strings.Contains(stderr, "not applicable") {
		t.Fatalf("expected stderr warning about non-tabulable input, got: %q", stderr)
	}
	// Stdout must be valid JSON encoding the original object.
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("fallback stdout is not valid JSON: %v\nraw:\n%s", err, stdout)
	}
	if got["id"] != "single" || got["filename"] != "lone.pdf" {
		t.Fatalf("fallback JSON missing fields: %v", got)
	}
}

// TestPrintResultTableNoPreferredColumns covers acceptance criterion
// #4: a list of objects whose keys don't match ANY preferred column
// (no id/name/filename/type/status/model/created_at) must fall back
// to JSON rather than crash or render a column-less header.
//
// This is the failure mode the auto-column path is most exposed to —
// an unknown resource shape lands in printResultTable from a generic
// `printResult` dispatch and we never want it to panic.
func TestPrintResultTableNoPreferredColumns(t *testing.T) {
	weird := map[string]any{
		"data": []any{
			map[string]any{"foo": "bar", "baz": 42},
			map[string]any{"foo": "qux", "baz": 7},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(weird); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if !strings.Contains(stderr, "not applicable") {
		t.Fatalf("expected stderr warning, got: %q", stderr)
	}
	// Stdout must be valid JSON containing the original payload.
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("fallback stdout is not valid JSON: %v\nraw:\n%s", err, stdout)
	}
	if _, ok := got["data"]; !ok {
		t.Fatalf("fallback JSON missing data key: %v", got)
	}
}

func TestPrintResultTableEmptyListHonorsTableOutput(t *testing.T) {
	empty := map[string]any{
		"data":          []any{},
		"list_metadata": map[string]any{},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(empty); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	// The lone "(no rows)" hint goes to stderr so an empty result is
	// visually distinguishable from a hidden one (R2-9). The hint must
	// NOT leak into stdout — downstream pipes still see the clean
	// header row.
	if strings.Contains(stderr, "falling back to json") {
		t.Fatalf("expected table render, got JSON-fallback warning: %q", stderr)
	}
	if !strings.Contains(stderr, "(no rows)") {
		t.Fatalf("expected (no rows) hint on stderr for empty list, got: %q", stderr)
	}
	if strings.Contains(stdout, "(no rows)") {
		t.Fatalf("(no rows) hint should not leak into stdout, got:\n%s", stdout)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Fatalf("expected table output for empty list, got JSON:\n%s", stdout)
	}
	for _, want := range []string{"ID", "NAME", "TYPE", "CREATED_AT"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected empty table header %q, got:\n%s", want, stdout)
		}
	}
}

// TestPrintResultDispatch pins the printResult dispatcher: --output
// table routes to printResultTable, everything else routes to
// printJSON. Per the bug spec, "auto" and "" both behave like JSON
// (no TTY detection on this path).
func TestPrintResultDispatch(t *testing.T) {
	resource := map[string]any{
		"data": []any{
			map[string]any{"id": "a_1", "name": "first"},
		},
	}

	cases := []struct {
		name      string
		flagValue string
		wantTable bool // true → assert table layout, false → assert JSON
	}{
		{"empty defaults to json", "", false},
		{"auto defaults to json", "auto", false},
		{"json explicit", "json", false},
		{"table renders table", "table", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := fakeRoot()
			if err := cmd.Root().PersistentFlags().Set("output", tc.flagValue); err != nil {
				t.Fatalf("set flag: %v", err)
			}
			stdout, _ := captureStd(t, func() {
				if err := printResult(cmd, resource); err != nil {
					t.Fatalf("printResult: %v", err)
				}
			})
			if tc.wantTable {
				if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "a_1") {
					t.Fatalf("expected table layout, got:\n%s", stdout)
				}
				// JSON output starts with '{' or '[' — table output starts
				// with the header line "ID  …".
				if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
					t.Fatalf("expected table but got JSON:\n%s", stdout)
				}
			} else {
				var got map[string]any
				if err := json.Unmarshal([]byte(stdout), &got); err != nil {
					t.Fatalf("expected JSON, parse failed: %v\n%s", err, stdout)
				}
			}
		})
	}
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
		{"csv explicit", "csv", OutputCSV, false},
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

// timeRow mirrors the SDK's workflow list-response element shape: a
// `created_at` field typed as time.Time (not a pre-formatted string).
// `retab workflows list --output table` decodes into exactly this shape.
type timeRow struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type timeRowList struct {
	Data []timeRow `json:"data"`
}

// TestPrintResultTableFormatsTimeDotTimeAsRFC3339 pins that a time.Time
// cell renders as RFC3339 in --output table, not Go's default
// time.Time.String() form ("2026-05-15 13:24:54.014 +0000 UTC").
//
// Regression: workflow list responses type created_at as time.Time, and
// stringifyCell's fmt.Stringer branch caught it first, emitting the Go
// default format — visibly inconsistent with jobs/files list tables,
// which render clean RFC3339 timestamps.
func TestPrintResultTableFormatsTimeDotTimeAsRFC3339(t *testing.T) {
	createdAt := time.Date(2026, 5, 15, 13, 24, 54, 0, time.UTC)
	list := timeRowList{
		Data: []timeRow{{ID: "wrk_1", Name: "probe", CreatedAt: createdAt}},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printResultTable(list); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "2026-05-15T13:24:54Z") {
		t.Fatalf("expected RFC3339 timestamp in table output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "+0000 UTC") {
		t.Fatalf("table leaked Go default time.Time format, got:\n%s", stdout)
	}
}

// Regression for the 2026-05-22 CLI exploration: “workflows list --output
// table“ against a workflow with a 350-char “name“ rendered the row at
// 350+ characters wide because the auto-table renderer only truncated the
// TRAILING column (CREATED_AT), and NAME is an interior column. The row
// became unreadable; the rest of the table was forced to that width too.
//
// Cap interior cells at a generous width (high enough to fit normal names,
// low enough that a runaway 350-char outlier doesn't blow out the row).
func TestPrintResultTableTruncatesAbsurdInteriorCellWidth(t *testing.T) {
	absurd := strings.Repeat("a", 400)
	list := timeRowList{
		Data: []timeRow{
			{ID: "wrk_1", Name: absurd, CreatedAt: time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)},
		},
	}

	stdout, _ := captureStd(t, func() {
		if err := printResultTable(list); err != nil {
			t.Fatalf("printResultTable: %v", err)
		}
	})

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least header + 1 row, got:\n%s", stdout)
	}
	rowLen := len(lines[1])
	if rowLen >= 400 {
		t.Fatalf("interior cell was not truncated; row width=%d, want <200:\n%s", rowLen, lines[1])
	}
	if !strings.Contains(lines[1], "…") {
		t.Fatalf("expected ellipsis on truncated interior cell, got:\n%s", lines[1])
	}
}
