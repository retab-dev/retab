package cmd

import (
	"image"
	"image/color"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

// flattenImageToRGB must composite transparency over white, not black.
func TestFlattenImageToRGBCompositesOverWhite(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	img.Set(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255}) // opaque red
	img.Set(1, 0, color.NRGBA{R: 0, G: 0, B: 0, A: 0})     // fully transparent
	img.Set(2, 0, color.NRGBA{R: 0, G: 0, B: 255, A: 128}) // semi-transparent blue
	rgb := flattenImageToRGB(img)
	if len(rgb) != 9 {
		t.Fatalf("expected 9 bytes, got %d", len(rgb))
	}
	// Opaque red unchanged.
	if rgb[0] != 255 || rgb[1] != 0 || rgb[2] != 0 {
		t.Errorf("opaque red = %v, want [255 0 0]", rgb[0:3])
	}
	// Fully transparent -> white, NOT black.
	if rgb[3] != 255 || rgb[4] != 255 || rgb[5] != 255 {
		t.Errorf("transparent pixel = %v, want white [255 255 255]", rgb[3:6])
	}
	// Semi-transparent (50%) blue over white = (0,0,255)*0.5 + white*0.5 ≈
	// (127,127,255). R/G lifted toward white, B stays 255.
	if rgb[6] < 120 || rgb[6] > 135 || rgb[7] < 120 || rgb[7] > 135 || rgb[8] != 255 {
		t.Errorf("semi-transparent blue = %v, want ~[127 127 255] composited over white", rgb[6:9])
	}
}

// coerceTableScalarValue must preserve large integers (no float64 rounding).
func TestCoerceTableScalarValuePreservesLargeInt(t *testing.T) {
	got := coerceTableScalarValue("12345678901234567890")
	num, ok := got.(interface{ String() string }) // json.Number implements Stringer
	if !ok {
		t.Fatalf("expected json.Number, got %T (%v)", got, got)
	}
	if num.String() != "12345678901234567890" {
		t.Errorf("large int mangled: got %q", num.String())
	}
	// Trailing junk stays a string (matches json.Unmarshal whole-input rule).
	if got := coerceTableScalarValue("12 34"); got != "12 34" {
		t.Errorf("coerceTableScalarValue(\"12 34\") = %v, want unchanged string", got)
	}
	// Booleans/null still coerce.
	if got := coerceTableScalarValue("true"); got != true {
		t.Errorf("true coercion failed: %v", got)
	}
}

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
		{"__complete", "files", "get", ""}, // cobra's hidden completion driver
		{"__completeNoDesc", "workflows"},   // no-description variant
		{"--global-flag", "completion"},     // value-less flags are skipped over
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

// stringifyCell must not overflow int64 on large integral floats.
func TestStringifyCellLargeFloat(t *testing.T) {
	// 1e19 is integral but > math.MaxInt64 (~9.22e18); must not wrap negative.
	if got := stringifyCell(float64(1e19)); got != "10000000000000000000" {
		t.Errorf("stringifyCell(1e19) = %q, want 10000000000000000000", got)
	}
	// In-range integral floats still render as plain integers.
	if got := stringifyCell(float64(42)); got != "42" {
		t.Errorf("stringifyCell(42) = %q, want 42", got)
	}
	// Non-integral keeps decimals.
	if got := stringifyCell(float64(3.5)); got != "3.5" {
		t.Errorf("stringifyCell(3.5) = %q, want 3.5", got)
	}
}

// sanitizeCSVCell must neutralize formula triggers but preserve numbers.
func TestSanitizeCSVCell(t *testing.T) {
	neutralized := map[string]string{
		"=cmd|'/c calc'!A1": "'=cmd|'/c calc'!A1",
		"@SUM(A1)":          "'@SUM(A1)",
		"+cmd":              "'+cmd",
		"-1+1":              "'-1+1",
		"\tinjected":        "'\tinjected",
	}
	for in, want := range neutralized {
		if got := sanitizeCSVCell(in); got != want {
			t.Errorf("sanitizeCSVCell(%q) = %q, want %q", in, got, want)
		}
	}
	// Legitimate values (including negative numbers) pass through untouched.
	for _, in := range []string{"", "report.pdf", "-42", "-3.14", "+7", "1e9", "hello"} {
		if got := sanitizeCSVCell(in); got != in {
			t.Errorf("sanitizeCSVCell(%q) = %q, want unchanged", in, got)
		}
	}
}

// boundingBoxForMatch must match across items with internal whitespace.
func TestBoundingBoxForMatchWhitespace(t *testing.T) {
	page := ParsedPage{
		Page: 1, Width: 100, Height: 100,
		Items: []ParsedItem{
			{Text: "Invoice   Total", X: 10, Y: 10, Width: 30, Height: 5}, // internal double space
			{Text: "Due", X: 42, Y: 10, Width: 10, Height: 5},
		},
	}
	if box := boundingBoxForMatch(page, "invoice total"); box == nil {
		t.Fatal("expected a bbox for a needle spanning whitespace-normalized item text")
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
