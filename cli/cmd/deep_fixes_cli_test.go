package cmd

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

// A crafted/garbage .xlsx cell ref like "ZZZZZZZ1" decodes to a column index in
// the hundreds of millions. Before the clamp, parseSheetRows allocated a row
// slice of that size (effective OOM). The hostile cell must be dropped.
func TestParseSheetRowsClampsHostileColumnRef(t *testing.T) {
	const sheetXML = `<worksheet><sheetData>` +
		`<row r="1"><c r="ZZZZZZZ1"><v>9</v></c><c r="A1"><v>1</v></c></row>` +
		`</sheetData></worksheet>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("sheet.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(sheetXML)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	rows, err := parseSheetRows(zr.File[0], &xlsxSharedStrings{})
	if err != nil {
		t.Fatalf("parseSheetRows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if got := len(rows[0]); got != 1 {
		t.Fatalf("row width = %d, want 1 (hostile out-of-range column must be clamped)", got)
	}
	if rows[0][0] != "1" {
		t.Fatalf("cell = %q, want \"1\"", rows[0][0])
	}
}

// Cell truncation must be rune-aware: slicing a multibyte string by bytes cuts
// mid-rune and emits invalid UTF-8.
func TestCleanWorkflowTableCellTruncatesByRune(t *testing.T) {
	cell := strings.Repeat("é", 50) // 2 bytes/rune, exceeds the width below
	got := cleanWorkflowTableCell(cell, tableQueryRenderOptions{MaxWidth: 10})
	if !utf8.ValidString(got) {
		t.Fatalf("truncated cell is not valid UTF-8: %q", got)
	}
	if n := utf8.RuneCountInString(got); n != 10 {
		t.Fatalf("rune count = %d, want 10 (7 runes + ellipsis)", n)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", got)
	}
}
