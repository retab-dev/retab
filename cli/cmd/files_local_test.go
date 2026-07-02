package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- column-letter math ----------------------------------------------------

func TestColLetterAndIndexRoundTrip(t *testing.T) {
	cases := []struct {
		n      int
		letter string
	}{
		{1, "A"}, {2, "B"}, {26, "Z"}, {27, "AA"}, {28, "AB"}, {52, "AZ"}, {53, "BA"}, {702, "ZZ"}, {703, "AAA"},
	}
	for _, tc := range cases {
		if got := colLetter(tc.n); got != tc.letter {
			t.Errorf("colLetter(%d) = %q, want %q", tc.n, got, tc.letter)
		}
		if got := colIndex(tc.letter); got != tc.n {
			t.Errorf("colIndex(%q) = %d, want %d", tc.letter, got, tc.n)
		}
	}
	if colLetter(0) != "" {
		t.Errorf("colLetter(0) should be empty")
	}
	if colIndex("") != 0 || colIndex("a1") != 0 {
		t.Errorf("colIndex should reject invalid input")
	}
}

// --- format detection ------------------------------------------------------

func TestDetectKind(t *testing.T) {
	cases := map[string]docKind{
		"a.pdf": kindPDF, "b.PNG": kindImage, "c.jpeg": kindImage, "d.tiff": kindImage,
		"e.csv": kindCSV, "f.tsv": kindCSV, "g.xlsx": kindSpreadsheet, "h.docx": kindDocx,
		"i.txt": kindText, "j.md": kindText, "k.json": kindText, "l.bin": kindUnknown,
		"noext": kindUnknown,
	}
	for path, want := range cases {
		if got := detectKind(path); got != want {
			t.Errorf("detectKind(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestDocKindUsesLiteParse(t *testing.T) {
	for k, want := range map[docKind]bool{
		kindPDF: true, kindImage: true, kindText: false, kindCSV: false, kindSpreadsheet: false, kindDocx: false,
	} {
		if got := k.usesLiteParse(); got != want {
			t.Errorf("%v.usesLiteParse() = %v, want %v", k, got, want)
		}
	}
}

// --- native parsers --------------------------------------------------------

func TestParseTextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("line one\nline two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := loadParse(context.Background(), path, kindText, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if res.Source != "native" || res.DocumentType != "text" || res.TotalPages != 1 {
		t.Fatalf("unexpected metadata: %+v", res)
	}
	if res.Pages[0].Text != "line one\nline two\n" {
		t.Fatalf("text = %q", res.Pages[0].Text)
	}
}

func TestParseCSVFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("name,amount\nACME,42000\nBeta,7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := loadParse(context.Background(), path, kindCSV, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if len(res.Sheets) != 1 {
		t.Fatalf("want 1 sheet, got %d", len(res.Sheets))
	}
	want := [][]string{{"name", "amount"}, {"ACME", "42000"}, {"Beta", "7"}}
	if !reflect.DeepEqual(res.Sheets[0].Rows, want) {
		t.Fatalf("rows = %v, want %v", res.Sheets[0].Rows, want)
	}
}

// Regression: a .txt/.md/.json authored on Windows carries a UTF-8 BOM
// (PowerShell `Out-File -Encoding utf8`) or is UTF-16 LE (default `>` redirect).
// parseTextFile read raw bytes and skipped the normalization that readJSON and
// readTextFileOrStdin already apply, so `files parse`/`grep`/`inspect` surfaced
// the first line with a stray U+FEFF (or fully garbled for UTF-16).
func TestParseTextFile_NormalizesBOMAndUTF16(t *testing.T) {
	dir := t.TempDir()

	utf8BOM := filepath.Join(dir, "utf8bom.txt")
	if err := os.WriteFile(utf8BOM, append([]byte{0xEF, 0xBB, 0xBF}, []byte("hello\nworld\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := loadParse(context.Background(), utf8BOM, kindText, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if res.Pages[0].Text != "hello\nworld\n" {
		t.Fatalf("utf-8 BOM text = %q, want clean", res.Pages[0].Text)
	}

	utf16LE := filepath.Join(dir, "utf16le.txt")
	raw := []byte{0xFF, 0xFE}
	for _, r := range "hello\nworld\n" {
		raw = append(raw, byte(r), 0x00)
	}
	if err := os.WriteFile(utf16LE, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = loadParse(context.Background(), utf16LE, kindText, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if res.Pages[0].Text != "hello\nworld\n" {
		t.Fatalf("utf-16 LE text = %q, want decoded", res.Pages[0].Text)
	}
}

// Regression: a Windows-authored CSV with a UTF-8 BOM used to glue the BOM onto
// the first header cell (U+FEFF prefix on "name"), corrupting grep
// matches and inspect --cells anchors against column A row 1.
func TestParseCSVFile_NormalizesBOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	body := append([]byte{0xEF, 0xBB, 0xBF}, []byte("name,amount\nACME,42000\n")...)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := loadParse(context.Background(), path, kindCSV, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	want := [][]string{{"name", "amount"}, {"ACME", "42000"}}
	if !reflect.DeepEqual(res.Sheets[0].Rows, want) {
		t.Fatalf("rows = %v, want %v (BOM should be stripped from first cell)", res.Sheets[0].Rows, want)
	}
}

func TestParseXLSXFile(t *testing.T) {
	path := writeTestXLSX(t)
	res, err := loadParse(context.Background(), path, kindSpreadsheet, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if len(res.Sheets) != 1 {
		t.Fatalf("want 1 sheet, got %d", len(res.Sheets))
	}
	s := res.Sheets[0]
	if s.Name != "Budget" {
		t.Errorf("sheet name = %q, want Budget", s.Name)
	}
	// A1=Item B1=Cost ; A2=Rent B2=1200
	if len(s.Rows) < 2 || s.Rows[0][0] != "Item" || s.Rows[1][1] != "1200" {
		t.Fatalf("unexpected rows: %v", s.Rows)
	}
}

// Sparse sheets omit empty rows in sheetData (e.g. r="1" then r="4"). The
// parser must honor the row's r attribute so reported coordinates stay aligned
// with the spreadsheet UI — mirroring the per-column gap handling. Before the
// fix, rows were appended in document order and a value physically in row 4 was
// reported as row 2.
func TestParseXLSXFileHonorsSparseRowNumbers(t *testing.T) {
	path := writeSparseTestXLSX(t)
	res, err := loadParse(context.Background(), path, kindSpreadsheet, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if len(res.Sheets) != 1 {
		t.Fatalf("want 1 sheet, got %d", len(res.Sheets))
	}
	rows := res.Sheets[0].Rows
	// A1=Item B1=Cost ; rows 2 and 3 are absent; A4=Rent B4=1200.
	if len(rows) != 4 {
		t.Fatalf("want 4 rows (positions preserved through the gap), got %d: %v", len(rows), rows)
	}
	if rows[0][0] != "Item" || rows[0][1] != "Cost" {
		t.Errorf("row 1 = %v, want [Item Cost]", rows[0])
	}
	if len(rows[1]) != 0 || len(rows[2]) != 0 {
		t.Errorf("rows 2-3 should be empty, got %v / %v", rows[1], rows[2])
	}
	if len(rows[3]) < 2 || rows[3][0] != "Rent" || rows[3][1] != "1200" {
		t.Fatalf("row 4 = %v, want [Rent 1200]", rows[3])
	}

	// The grep anchor must report the spreadsheet row (4), not the dense index.
	matcher, err := buildMatcher("Rent", false, false)
	if err != nil {
		t.Fatalf("buildMatcher: %v", err)
	}
	var got *Anchor
	grepSheets(res, kindSpreadsheet, matcher, 0, func(m grepMatch) bool {
		a := m.Anchor
		got = &a
		return false
	})
	if got == nil {
		t.Fatal("expected a grep match for Rent")
	}
	if got.Row != 4 || got.Coordinate != "A4" {
		t.Errorf("anchor = row %d coord %q, want row 4 coord A4", got.Row, got.Coordinate)
	}
}

// The OOXML cell `r` attribute is optional: an absent ref means "the column
// after the previous cell". Some writers omit it entirely. Before the fix
// every such cell resolved to column 0 and was dropped, so a sheet written
// without cell refs parsed as completely empty. The parser must place
// ref-less cells in document order, mirroring the implicit-row handling.
func TestParseXLSXFileHonorsImplicitCellRefs(t *testing.T) {
	path := writeImplicitRefTestXLSX(t)
	res, err := loadParse(context.Background(), path, kindSpreadsheet, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if len(res.Sheets) != 1 {
		t.Fatalf("want 1 sheet, got %d", len(res.Sheets))
	}
	rows := res.Sheets[0].Rows
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d: %v", len(rows), rows)
	}
	if len(rows[0]) < 2 || rows[0][0] != "Item" || rows[0][1] != "Cost" {
		t.Fatalf("row 1 = %v, want [Item Cost]", rows[0])
	}
	if len(rows[1]) < 2 || rows[1][0] != "Rent" || rows[1][1] != "1200" {
		t.Fatalf("row 2 = %v, want [Rent 1200]", rows[1])
	}
}

func TestParseDocxFile(t *testing.T) {
	path := writeTestDOCX(t)
	res, err := loadParse(context.Background(), path, kindDocx, "", defaultParseOptions(), false)
	if err != nil {
		t.Fatalf("loadParse: %v", err)
	}
	if res.DocumentType != "docx" || res.Source != "native" {
		t.Fatalf("unexpected metadata: %+v", res)
	}
	if !strings.Contains(res.Pages[0].Text, "Hello world") {
		t.Fatalf("text = %q", res.Pages[0].Text)
	}
	if !strings.Contains(res.Pages[0].Text, "Second paragraph") {
		t.Fatalf("missing second paragraph: %q", res.Pages[0].Text)
	}
}

func TestLoadParseMissingFile(t *testing.T) {
	_, err := loadParse(context.Background(), "/no/such/file.txt", kindText, "", defaultParseOptions(), false)
	if err == nil || !strings.Contains(err.Error(), "file not found") {
		t.Fatalf("want file not found, got %v", err)
	}
}

// --- matcher / anchor math -------------------------------------------------

func TestBuildMatcher(t *testing.T) {
	// Literal pattern with regex metacharacters is matched verbatim.
	m, err := buildMatcher("a.b", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if m.MatchString("axb") {
		t.Error("literal pattern should not treat . as wildcard")
	}
	if !m.MatchString("A.B") {
		t.Error("default should be case-insensitive")
	}
	// Case-sensitive literal.
	m2, _ := buildMatcher("ACME", false, true)
	if m2.MatchString("acme") {
		t.Error("case-sensitive should not match lowercase")
	}
	// Regex mode.
	m3, _ := buildMatcher(`INV-\d+`, true, true)
	if !m3.MatchString("INV-42") {
		t.Error("regex should match")
	}
	if _, err := buildMatcher(`(unclosed`, true, false); err == nil {
		t.Error("invalid regex should error")
	}
}

func TestLineColAt(t *testing.T) {
	text := "abc\ndé\nxyz"
	// offset of 'x' (after second newline). bytes: a0 b1 c2 \n3 d4 é5,6 \n7 x8
	line, col := lineColAt(text, 8)
	if line != 3 || col != 0 {
		t.Fatalf("lineColAt x = (%d,%d), want (3,0)", line, col)
	}
	// offset just after 'é' on line 2: col should count runes, not bytes.
	line, col = lineColAt(text, 7)
	if line != 2 || col != 2 {
		t.Fatalf("lineColAt after é = (%d,%d), want (2,2)", line, col)
	}
}

func TestSnippet(t *testing.T) {
	text := "the quick brown fox jumps"
	// match "brown" at [10,15]; 4 chars of context each side -> text[6:19]
	got := snippet(text, 10, 15, 4)
	if got != "ick brown fox" {
		t.Fatalf("snippet = %q", got)
	}
	// zero context returns exactly the match.
	if snippet(text, 10, 15, 0) != "brown" {
		t.Fatalf("zero-context snippet wrong")
	}
}

func TestFlattenWhitespace(t *testing.T) {
	// Context windows pull in newlines; the table view must collapse them so a
	// match stays on one row instead of wrapping and breaking alignment.
	got := flattenWhitespace("# Title\n\nRevenue grew 18%\tquarter over\n  quarter.")
	want := "# Title Revenue grew 18% quarter over quarter."
	if got != want {
		t.Fatalf("flattenWhitespace = %q, want %q", got, want)
	}
	if flattenWhitespace("   \n\t ") != "" {
		t.Errorf("all-whitespace should flatten to empty string")
	}
	if flattenWhitespace("single") != "single" {
		t.Errorf("single token should be unchanged")
	}
}

func TestClamp01(t *testing.T) {
	if clamp01(-0.5) != 0 || clamp01(1.5) != 1 || clamp01(0.3) != 0.3 {
		t.Fatal("clamp01 broken")
	}
}

// --- grep dispatch per kind ------------------------------------------------

func TestGrepTextSpans(t *testing.T) {
	res := &ParseResult{
		Filename: "n.txt", DocumentType: "text", TotalPages: 1,
		Pages: []ParsedPage{{Page: 1, Text: "alpha beta\ngamma beta delta\n"}},
	}
	m, _ := buildMatcher("beta", false, false)
	matches, trunc := grepParseResult(res, kindText, m, 0, 50, false)
	if trunc {
		t.Fatal("unexpected truncation")
	}
	if len(matches) != 2 {
		t.Fatalf("want 2 matches, got %d", len(matches))
	}
	a := matches[0].Anchor
	if a.Kind != anchorTextSpan || a.LineStart != 1 || deref(a.CharStart) != 6 {
		t.Fatalf("first anchor wrong: %+v", a)
	}
	if matches[1].Anchor.LineStart != 2 {
		t.Fatalf("second match should be on line 2: %+v", matches[1].Anchor)
	}
}

func TestGrepSheetsCSV(t *testing.T) {
	res := &ParseResult{
		Filename: "d.csv", DocumentType: "csv",
		Sheets: []SheetData{{Index: 0, Rows: [][]string{{"name", "amount"}, {"ACME", "42000"}}}},
	}
	m, _ := buildMatcher("42000", false, false)
	matches, _ := grepParseResult(res, kindCSV, m, 0, 50, false)
	if len(matches) != 1 {
		t.Fatalf("want 1 match, got %d", len(matches))
	}
	a := matches[0].Anchor
	if a.Kind != anchorCSVCell || a.Row != 2 || a.Column != "B" || a.Coordinate != "B2" {
		t.Fatalf("csv anchor wrong: %+v", a)
	}
	if a.SheetIndex != nil {
		t.Errorf("csv anchor should not carry sheet index")
	}
}

func TestGrepSheetsSpreadsheet(t *testing.T) {
	res := &ParseResult{
		Filename: "d.xlsx", DocumentType: "spreadsheet",
		Sheets: []SheetData{{Index: 0, Name: "Budget", Rows: [][]string{{"Item", "Cost"}, {"Rent", "1200"}}}},
	}
	m, _ := buildMatcher("1200", false, false)
	matches, _ := grepParseResult(res, kindSpreadsheet, m, 0, 50, false)
	if len(matches) != 1 {
		t.Fatalf("want 1 match, got %d", len(matches))
	}
	a := matches[0].Anchor
	if a.Kind != anchorSpreadsheetCell || a.Coordinate != "B2" || a.SheetName != "Budget" {
		t.Fatalf("xlsx anchor wrong: %+v", a)
	}
	if a.SheetIndex == nil || *a.SheetIndex != 0 {
		t.Fatalf("xlsx anchor sheet index wrong: %+v", a)
	}
}

func TestGrepPagesPDFAndBbox(t *testing.T) {
	res := &ParseResult{
		Filename: "r.pdf", DocumentType: "pdf", TotalPages: 1,
		Pages: []ParsedPage{{
			Page: 1, Width: 100, Height: 200,
			Text: "header line\nTotal Due here\n",
			Items: []ParsedItem{
				{Text: "Total", X: 10, Y: 20, Width: 30, Height: 10},
				{Text: "Due", X: 45, Y: 20, Width: 20, Height: 10},
			},
		}},
	}
	m, _ := buildMatcher("Total Due", false, false)
	matches, _ := grepParseResult(res, kindPDF, m, 0, 50, true)
	if len(matches) != 1 {
		t.Fatalf("want 1 match, got %d", len(matches))
	}
	a := matches[0].Anchor
	if a.Kind != anchorPDFPage || a.Page != 1 || a.Line != 2 {
		t.Fatalf("pdf anchor wrong: %+v", a)
	}
	if a.Bbox == nil {
		t.Fatal("expected a bbox with --bbox")
	}
	// left = minX/width = 10/100 = 0.1 ; width = (65-10)/100 = 0.55
	if a.Bbox.Left < 0.09 || a.Bbox.Left > 0.11 || a.Bbox.Width < 0.54 || a.Bbox.Width > 0.56 {
		t.Fatalf("bbox math wrong: %+v", a.Bbox)
	}
}

func TestGrepTruncation(t *testing.T) {
	res := &ParseResult{
		Pages: []ParsedPage{{Page: 1, Text: "x x x x x"}},
	}
	m, _ := buildMatcher("x", false, false)
	matches, trunc := grepParseResult(res, kindPDF, m, 0, 3, false)
	if !trunc || len(matches) != 3 {
		t.Fatalf("want truncated to 3, got %d trunc=%v", len(matches), trunc)
	}
}

// --- inspect parsing helpers -----------------------------------------------

func TestParseLineRange(t *testing.T) {
	cases := []struct {
		spec       string
		start, end int
		wantErr    bool
	}{
		{"10-40", 10, 40, false},
		{"5", 5, 5, false},
		{"3 - 7", 3, 7, false},
		{"40-10", 0, 0, true},
		{"0-5", 0, 0, true},
		{"abc", 0, 0, true},
	}
	for _, tc := range cases {
		s, e, err := parseLineRange(tc.spec)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseLineRange(%q) want error", tc.spec)
			}
			continue
		}
		if err != nil || s != tc.start || e != tc.end {
			t.Errorf("parseLineRange(%q) = (%d,%d,%v), want (%d,%d)", tc.spec, s, e, err, tc.start, tc.end)
		}
	}
}

func TestParseCellRange(t *testing.T) {
	sc, sr, ec, er, err := parseCellRange("A1:C3")
	if err != nil || sc != 1 || sr != 1 || ec != 3 || er != 3 {
		t.Fatalf("A1:C3 = (%d,%d,%d,%d,%v)", sc, sr, ec, er, err)
	}
	// reversed corners normalize.
	sc, sr, ec, er, err = parseCellRange("C3:A1")
	if err != nil || sc != 1 || sr != 1 || ec != 3 || er != 3 {
		t.Fatalf("C3:A1 normalize failed: (%d,%d,%d,%d,%v)", sc, sr, ec, er, err)
	}
	// single cell.
	sc, sr, ec, er, _ = parseCellRange("B14")
	if sc != 2 || sr != 14 || ec != 2 || er != 14 {
		t.Fatalf("single cell wrong: (%d,%d,%d,%d)", sc, sr, ec, er)
	}
	if _, _, _, _, err := parseCellRange("nope"); err == nil {
		t.Error("invalid cell range should error")
	}
}

func TestParsePageList(t *testing.T) {
	got, err := parsePageList("3,1,1,5-7")
	if err != nil {
		t.Fatal(err)
	}
	want := []int{1, 3, 5, 6, 7}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parsePageList = %v, want %v", got, want)
	}
	if _, err := parsePageList("0"); err == nil {
		t.Error("page 0 should error")
	}
	if _, err := parsePageList(""); err == nil {
		t.Error("empty spec should error")
	}
}

func TestSliceCells(t *testing.T) {
	rows := [][]string{
		{"a1", "b1", "c1"},
		{"a2", "b2"},
		{"a3", "b3", "c3"},
	}
	got := sliceCells(rows, 1, 1, 3, 2) // A1:C2
	want := [][]string{{"a1", "b1", "c1"}, {"a2", "b2", ""}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sliceCells = %v, want %v", got, want)
	}
}

// --- litCLI argv construction ----------------------------------------------

func TestLitCLIParseArgs(t *testing.T) {
	c := &litCLI{bin: "lit"}
	args := c.parseArgs("doc.pdf", ParseOptions{OCR: true, OCRLanguage: "fra", DPI: 200, TargetPages: "1-3"})
	joined := strings.Join(args, " ")
	for _, want := range []string{"parse doc.pdf", "--format json", "--quiet", "--ocr-language fra", "--dpi 200", "--target-pages 1-3"} {
		if !strings.Contains(joined, want) {
			t.Errorf("args %q missing %q", joined, want)
		}
	}
	if strings.Contains(joined, "--no-ocr") {
		t.Errorf("OCR enabled should not pass --no-ocr: %q", joined)
	}
	// no-ocr path
	noOCR := strings.Join(c.parseArgs("doc.pdf", ParseOptions{OCR: false}), " ")
	if !strings.Contains(noOCR, "--no-ocr") {
		t.Errorf("OCR disabled should pass --no-ocr: %q", noOCR)
	}
}

func TestLitCLIParseSurfacesOCRFailureOnSuccessfulExit(t *testing.T) {
	dir := t.TempDir()
	fakeLit := filepath.Join(dir, "lit")
	script := `#!/bin/sh
printf '{"pages":[{"page":1,"width":612,"height":792,"text":"","textItems":[]}]}'
printf 'Error opening data file /tmp/home/tessdata/eng.traineddata\n' >&2
printf 'Failed loading language '\''eng'\''\n' >&2
printf '[ocr] failed for page 1: Failed to initialize Tesseract\n' >&2
exit 0
`
	if err := os.WriteFile(fakeLit, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake lit: %v", err)
	}

	c := &litCLI{bin: fakeLit}
	_, err := c.Parse(context.Background(), filepath.Join(dir, "doc.pdf"), ParseOptions{OCR: true, OCRLanguage: "eng", DPI: 150})
	if err == nil {
		t.Fatal("expected OCR stderr diagnostic to fail parse")
	}
	if got := err.Error(); !strings.Contains(got, "liteparse OCR failed") || !strings.Contains(got, "Failed to initialize Tesseract") {
		t.Fatalf("error did not surface OCR diagnostic:\n%s", got)
	}
}

// --- fake LiteParser end-to-end (pdf grep + inspect render) ----------------

type fakeLiteParser struct {
	parse       *ParseResult
	shots       []ScreenshotPage
	lastOpt     ParseOptions
	lastShotOpt ScreenshotOptions
}

func (f *fakeLiteParser) Parse(_ context.Context, _ string, opt ParseOptions) (*ParseResult, error) {
	f.lastOpt = opt
	return f.parse, nil
}
func (f *fakeLiteParser) Screenshot(_ context.Context, _ string, opt ScreenshotOptions) ([]ScreenshotPage, error) {
	f.lastShotOpt = opt
	if f.parse != nil && f.parse.TotalPages > 0 && opt.TargetPages != "" {
		pages, err := parsePageList(opt.TargetPages)
		if err != nil {
			return nil, err
		}
		for _, page := range pages {
			if page > f.parse.TotalPages {
				return nil, fmt.Errorf("page %d out of range (document has %d pages)", page, f.parse.TotalPages)
			}
		}
	}
	return f.shots, nil
}
func (f *fakeLiteParser) Version(_ context.Context) (string, error) { return "lit 9.9.9", nil }

func withFakeLiteParser(t *testing.T, fake LiteParser) {
	t.Helper()
	orig := resolveLiteParserFn
	resolveLiteParserFn = func(string) (LiteParser, error) { return fake, nil }
	t.Cleanup(func() { resolveLiteParserFn = orig })
}

func TestFilesGrepPDFEndToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4 fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	withFakeLiteParser(t, &fakeLiteParser{parse: &ParseResult{
		TotalPages: 1,
		Pages:      []ParsedPage{{Page: 1, Width: 100, Height: 100, Text: "intro\nfind me here\n"}},
	}})

	resetGrepFlags(t)
	stdout, _ := captureStd(t, func() {
		if err := filesGrepCmd.RunE(filesGrepCmd, []string{path, "find me"}); err != nil {
			t.Fatalf("grep: %v", err)
		}
	})
	var out grepResult
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, stdout)
	}
	if out.TotalMatches != 1 || out.Matches[0].Anchor.Page != 1 || out.Matches[0].Anchor.Line != 2 {
		t.Fatalf("unexpected grep result: %+v", out)
	}
	if out.DocumentType != "pdf" {
		t.Errorf("document_type = %q", out.DocumentType)
	}
}

func TestFilesInspectRenderEndToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4 fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "out")
	withFakeLiteParser(t, &fakeLiteParser{
		parse: &ParseResult{
			TotalPages: 1,
			Pages:      []ParsedPage{{Page: 1, Text: "page\n"}},
		},
		shots: []ScreenshotPage{
			{Page: 1, Path: filepath.Join(outDir, "page-1.png"), MIMEType: "image/png"},
		},
	})

	resetInspectFlags(t)
	if err := filesInspectCmd.Flags().Set("render", "1"); err != nil {
		t.Fatal(err)
	}
	if err := filesInspectCmd.Flags().Set("out", outDir); err != nil {
		t.Fatal(err)
	}
	stdout, _ := captureStd(t, func() {
		if err := filesInspectCmd.RunE(filesInspectCmd, []string{path}); err != nil {
			t.Fatalf("inspect: %v", err)
		}
	})
	var out inspectRenderResult
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, stdout)
	}
	if len(out.Pages) != 1 || out.Pages[0].Page != 1 {
		t.Fatalf("unexpected render result: %+v", out)
	}
}

func TestFilesInspectRenderClipsPagesPastDocumentEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4 fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "out")
	fake := &fakeLiteParser{
		parse: &ParseResult{
			TotalPages: 1,
			Pages:      []ParsedPage{{Page: 1, Text: "only page\n"}},
		},
		shots: []ScreenshotPage{
			{Page: 1, Path: filepath.Join(outDir, "page-1.png"), MIMEType: "image/png"},
		},
	}
	withFakeLiteParser(t, fake)

	resetInspectFlags(t)
	if err := filesInspectCmd.Flags().Set("render", "1-3"); err != nil {
		t.Fatal(err)
	}
	if err := filesInspectCmd.Flags().Set("out", outDir); err != nil {
		t.Fatal(err)
	}
	stdout, _ := captureStd(t, func() {
		if err := filesInspectCmd.RunE(filesInspectCmd, []string{path}); err != nil {
			t.Fatalf("inspect: %v", err)
		}
	})
	var out inspectRenderResult
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, stdout)
	}
	if fake.lastShotOpt.TargetPages != "1" {
		t.Fatalf("target pages = %q, want 1", fake.lastShotOpt.TargetPages)
	}
	if len(out.Pages) != 1 || out.Pages[0].Page != 1 {
		t.Fatalf("unexpected render result: %+v", out)
	}
}

func TestFilesInspectLinesEndToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "n.txt")
	if err := os.WriteFile(path, []byte("l1\nl2\nl3\nl4\nl5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resetInspectFlags(t)
	if err := filesInspectCmd.Flags().Set("lines", "2-4"); err != nil {
		t.Fatal(err)
	}
	stdout, _ := captureStd(t, func() {
		if err := filesInspectCmd.RunE(filesInspectCmd, []string{path}); err != nil {
			t.Fatalf("inspect: %v", err)
		}
	})
	var out inspectLinesResult
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, stdout)
	}
	if !reflect.DeepEqual(out.Lines, []string{"l2", "l3", "l4"}) {
		t.Fatalf("lines = %v", out.Lines)
	}
}

func TestFilesInspectModeMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "n.txt")
	if err := os.WriteFile(path, []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resetInspectFlags(t)
	if err := filesInspectCmd.Flags().Set("cells", "A1:B2"); err != nil {
		t.Fatal(err)
	}
	err := filesInspectCmd.RunE(filesInspectCmd, []string{path})
	if err == nil || !strings.Contains(err.Error(), "csv/xlsx") {
		t.Fatalf("want mode mismatch error, got %v", err)
	}
}

func TestFilesParseTextFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "n.txt")
	if err := os.WriteFile(path, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resetParseFlags(t)
	stdout, _ := captureStd(t, func() {
		if err := filesParseCmd.RunE(filesParseCmd, []string{path}); err != nil {
			t.Fatalf("parse: %v", err)
		}
	})
	if stdout != "hello\nworld\n" {
		t.Fatalf("parse text output = %q", stdout)
	}
}

// --- helpers to reset persistent command flags between table-driven runs ---

func resetParseFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_ = filesParseCmd.Flags().Set("format", "text")
		_ = filesParseCmd.Flags().Set("bbox", "false")
		_ = filesParseCmd.Flags().Set("out", "")
		_ = filesParseCmd.Flags().Set("no-cache", "false")
	})
	_ = filesParseCmd.Flags().Set("no-cache", "true")
}

func resetGrepFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_ = filesGrepCmd.Flags().Set("regex", "false")
		_ = filesGrepCmd.Flags().Set("case-sensitive", "false")
		_ = filesGrepCmd.Flags().Set("bbox", "false")
		_ = filesGrepCmd.Flags().Set("no-cache", "false")
	})
	_ = filesGrepCmd.Flags().Set("no-cache", "true")
}

func resetInspectFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_ = filesInspectCmd.Flags().Set("lines", "")
		_ = filesInspectCmd.Flags().Set("cells", "")
		_ = filesInspectCmd.Flags().Set("render", "")
		_ = filesInspectCmd.Flags().Set("sheet", "")
		_ = filesInspectCmd.Flags().Set("out", "")
		_ = filesInspectCmd.Flags().Set("no-cache", "false")
	})
	_ = filesInspectCmd.Flags().Set("no-cache", "true")
}

// --- fixture builders ------------------------------------------------------

func writeTestXLSX(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "book.xlsx")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, content string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	add("xl/workbook.xml", `<?xml version="1.0"?>
<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="Budget" sheetId="1" r:id="rId1"/></sheets>
</workbook>`)
	add("xl/_rels/workbook.xml.rels", `<?xml version="1.0"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Target="worksheets/sheet1.xml"/>
</Relationships>`)
	add("xl/sharedStrings.xml", `<?xml version="1.0"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <si><t>Item</t></si><si><t>Cost</t></si><si><t>Rent</t></si>
</sst>`)
	add("xl/worksheets/sheet1.xml", `<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c></row>
    <row r="2"><c r="A2" t="s"><v>2</v></c><c r="B2"><v>1200</v></c></row>
  </sheetData>
</worksheet>`)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeSparseTestXLSX(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sparse.xlsx")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, content string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	add("xl/workbook.xml", `<?xml version="1.0"?>
<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="Budget" sheetId="1" r:id="rId1"/></sheets>
</workbook>`)
	add("xl/_rels/workbook.xml.rels", `<?xml version="1.0"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Target="worksheets/sheet1.xml"/>
</Relationships>`)
	add("xl/sharedStrings.xml", `<?xml version="1.0"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <si><t>Item</t></si><si><t>Cost</t></si><si><t>Rent</t></si>
</sst>`)
	// Rows 2 and 3 are omitted, as real spreadsheets do for empty rows.
	add("xl/worksheets/sheet1.xml", `<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c></row>
    <row r="4"><c r="A4" t="s"><v>2</v></c><c r="B4"><v>1200</v></c></row>
  </sheetData>
</worksheet>`)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeImplicitRefTestXLSX builds a sheet whose cells omit the optional `r`
// attribute (only the rows carry `r`). Column position is then implied by
// document order: first <c> = column A, next = column B, etc.
func writeImplicitRefTestXLSX(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "implicit.xlsx")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, content string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	add("xl/workbook.xml", `<?xml version="1.0"?>
<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="Budget" sheetId="1" r:id="rId1"/></sheets>
</workbook>`)
	add("xl/_rels/workbook.xml.rels", `<?xml version="1.0"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Target="worksheets/sheet1.xml"/>
</Relationships>`)
	add("xl/sharedStrings.xml", `<?xml version="1.0"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <si><t>Item</t></si><si><t>Cost</t></si><si><t>Rent</t></si>
</sst>`)
	// Cells carry no r attribute; column is implied by order.
	add("xl/worksheets/sheet1.xml", `<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c t="s"><v>0</v></c><c t="s"><v>1</v></c></row>
    <row r="2"><c t="s"><v>2</v></c><c><v>1200</v></c></row>
  </sheetData>
</worksheet>`)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTestDOCX(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.docx")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(`<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Hello world</w:t></w:r></w:p>
    <w:p><w:r><w:t>Second paragraph</w:t></w:r></w:p>
  </w:body>
</w:document>`)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
