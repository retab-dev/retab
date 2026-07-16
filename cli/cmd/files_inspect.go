package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// inspectLinesResult is the JSON shape for `files inspect --lines`.
type inspectLinesResult struct {
	Filename     string   `json:"filename"`
	DocumentType string   `json:"document_type"`
	LineStart    int      `json:"line_start"`
	LineEnd      int      `json:"line_end"`
	Lines        []string `json:"lines"`
}

// inspectCellsResult is the JSON shape for `files inspect --cells`.
type inspectCellsResult struct {
	Filename     string     `json:"filename"`
	DocumentType string     `json:"document_type"`
	SheetIndex   int        `json:"sheet_index"`
	SheetName    string     `json:"sheet_name,omitempty"`
	Range        string     `json:"range"`
	Rows         [][]string `json:"rows"`
}

// inspectRenderResult is the JSON shape for `files inspect --render`.
type inspectRenderResult struct {
	Filename     string           `json:"filename"`
	DocumentType string           `json:"document_type"`
	OutputDir    string           `json:"output_dir"`
	Pages        []ScreenshotPage `json:"pages"`
}

var filesInspectCmd = &cobra.Command{
	Use:   "inspect <path>",
	Short: "Inspect lines/cells or render PDF/image pages to PNG",
	Long: `Inspect lines/cells or render PDF/image pages to PNG.

It works on local documents entirely locally: no upload, no API call.

This mirrors the Retab MCP files_inspect tool. Pick exactly one mode:

  --lines A-B          read a line range from a text/docx document
  --cells A1:Z100      read a cell range from a csv/xlsx document
  --render 1,3,5       render pdf/image pages to PNG (via LiteParse)

Use --render when an agent or human needs to see the page visually. It shells
out to the ` + "`lit`" + ` binary, writes one PNG per selected page into --out
(or a temp directory), and prints JSON with the image paths. If the requested
range runs past the end of the document, only existing pages are rendered.

The selected mode must match the document type.`,
	Example: `  # Read lines 10-40 of a text file
  retab files inspect notes.md --lines 10-40

  # Read a cell range from the second sheet
  retab files inspect data.xlsx --cells A1:D20 --sheet 2

  # Render the first three PDF pages to PNG images
  retab files inspect report.pdf --render 1-3 --out ./pages

  # Render one local image to a normalized PNG and print its path
  retab files inspect receipt.jpg --render 1 --out ./pages`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		path, err := localFilePath(cmd, args)
		if err != nil {
			return err
		}
		kind := detectKind(path)
		if kind == kindUnknown {
			return fmt.Errorf("unsupported file type for %s (supported: pdf, images, txt/md/json, csv/tsv, xlsx, docx)", path)
		}

		lines, _ := cmd.Flags().GetString("lines")
		cells, _ := cmd.Flags().GetString("cells")
		render, _ := cmd.Flags().GetString("render")

		set := 0
		for _, v := range []string{lines, cells, render} {
			if v != "" {
				set++
			}
		}
		if set == 0 {
			return fmt.Errorf("one of --lines, --cells, or --render is required")
		}
		if set > 1 {
			return fmt.Errorf("--lines, --cells, and --render are mutually exclusive")
		}

		ctx, cancel := ctxFor(cmd)
		defer cancel()

		switch {
		case lines != "":
			return inspectLines(ctx, cmd, path, kind, lines)
		case cells != "":
			return inspectCells(ctx, cmd, path, kind, cells)
		default:
			return inspectRender(ctx, cmd, path, kind, render)
		}
	}),
}

func inspectLines(ctx context.Context, cmd *cobra.Command, path string, kind docKind, spec string) error {
	if kind != kindText && kind != kindDocx {
		return fmt.Errorf("--lines applies to text/docx documents, not %s", kind.documentType())
	}
	start, end, err := parseLineRange(spec)
	if err != nil {
		return err
	}
	result, err := loadParse(ctx, path, kind, liteBinFromCmd(cmd), parseOptionsFromCmd(cmd), useCacheFromCmd(cmd))
	if err != nil {
		return err
	}
	if len(result.Pages) == 0 {
		return fmt.Errorf("%s has no text", result.Filename)
	}
	all := strings.Split(result.Pages[0].Text, "\n")
	// Drop a single trailing empty element produced by a final newline.
	if n := len(all); n > 0 && all[n-1] == "" {
		all = all[:n-1]
	}
	if start > len(all) {
		return fmt.Errorf("line %d is past end of file (%d lines)", start, len(all))
	}
	if end > len(all) {
		end = len(all)
	}
	return printJSON(inspectLinesResult{
		Filename:     result.Filename,
		DocumentType: result.DocumentType,
		LineStart:    start,
		LineEnd:      end,
		Lines:        all[start-1 : end],
	})
}

func inspectCells(ctx context.Context, cmd *cobra.Command, path string, kind docKind, spec string) error {
	if kind != kindCSV && kind != kindSpreadsheet {
		return fmt.Errorf("--cells applies to csv/xlsx documents, not %s", kind.documentType())
	}
	startCol, startRow, endCol, endRow, err := parseCellRange(spec)
	if err != nil {
		return err
	}
	result, err := loadParse(ctx, path, kind, liteBinFromCmd(cmd), parseOptionsFromCmd(cmd), useCacheFromCmd(cmd))
	if err != nil {
		return err
	}
	sheet, err := selectSheet(cmd, result, kind)
	if err != nil {
		return err
	}
	rows := sliceCells(sheet.Rows, startCol, startRow, endCol, endRow)
	return printJSON(inspectCellsResult{
		Filename:     result.Filename,
		DocumentType: result.DocumentType,
		SheetIndex:   sheet.Index,
		SheetName:    sheet.Name,
		Range:        strings.ToUpper(spec),
		Rows:         rows,
	})
}

func inspectRender(ctx context.Context, cmd *cobra.Command, path string, kind docKind, spec string) error {
	if !kind.usesLiteParse() {
		return fmt.Errorf("--render applies to pdf/image documents, not %s", kind.documentType())
	}
	pages, err := parsePageList(spec)
	if err != nil {
		return err
	}

	outDir, _ := cmd.Flags().GetString("out")
	if outDir == "" {
		outDir, err = os.MkdirTemp("", "retab-inspect-*")
		if err != nil {
			return fmt.Errorf("create temp output dir: %w", err)
		}
	}

	parser, err := resolveLiteParser(liteBinFromCmd(cmd))
	if err != nil {
		return err
	}
	dpi, _ := cmd.Flags().GetInt("dpi")
	countOpt := parseOptionsFromCmd(cmd)
	countOpt.OCR = false
	countOpt.TargetPages = ""
	parsed, err := parser.Parse(ctx, path, countOpt)
	if err != nil {
		return err
	}
	pages, err = clipPagesToDocument(pages, parsed.TotalPages)
	if err != nil {
		return err
	}
	if len(pages) > 3 {
		return fmt.Errorf("--render supports at most 3 pages (got %d after clipping to document bounds)", len(pages))
	}
	shots, err := parser.Screenshot(ctx, path, ScreenshotOptions{
		TargetPages: pageListToSpec(pages),
		DPI:         dpi,
		OutDir:      outDir,
	})
	if err != nil {
		return err
	}
	return printJSON(inspectRenderResult{
		Filename:     fileBase(path),
		DocumentType: kind.documentType(),
		OutputDir:    outDir,
		Pages:        shots,
	})
}

func clipPagesToDocument(pages []int, totalPages int) ([]int, error) {
	if totalPages <= 0 {
		return nil, fmt.Errorf("document has no renderable pages")
	}
	clipped := make([]int, 0, len(pages))
	for _, page := range pages {
		if page <= totalPages {
			clipped = append(clipped, page)
		}
	}
	if len(clipped) == 0 {
		return nil, fmt.Errorf("requested pages are outside document bounds (document has %d pages)", totalPages)
	}
	return clipped, nil
}

// selectSheet resolves --sheet (name or 1-based index) against a parse result.
// Empty --sheet selects the first sheet. CSV always has exactly one sheet.
func selectSheet(cmd *cobra.Command, result *ParseResult, kind docKind) (SheetData, error) {
	if len(result.Sheets) == 0 {
		return SheetData{}, fmt.Errorf("%s has no sheets", result.Filename)
	}
	sel, _ := cmd.Flags().GetString("sheet")
	if sel == "" || kind == kindCSV {
		return result.Sheets[0], nil
	}
	if idx, err := strconv.Atoi(sel); err == nil {
		if idx < 1 || idx > len(result.Sheets) {
			return SheetData{}, fmt.Errorf("sheet index %d out of range (1-%d)", idx, len(result.Sheets))
		}
		return result.Sheets[idx-1], nil
	}
	for _, s := range result.Sheets {
		if strings.EqualFold(s.Name, sel) {
			return s, nil
		}
	}
	return SheetData{}, fmt.Errorf("no sheet named %q", sel)
}

// sliceCells extracts the rectangular [startRow,endRow] x [startCol,endCol]
// region (all 1-based, inclusive) from rows, padding short rows with empty
// strings so the result is rectangular.
func sliceCells(rows [][]string, startCol, startRow, endCol, endRow int) [][]string {
	// Clamp the column span to Excel's real maximum width (XFD = 16384).
	// Without this, a range like `--cells A1:ZZZZZZ1` (ZZZZZZ ≈ column
	// 321,272,406) drives a multi-gigabyte rectangular allocation of empty
	// padding cells — an effective out-of-memory DoS from a single flag. Real
	// spreadsheets never exceed maxExcelColumns and the xlsx parser already
	// enforces the same cap, so this changes no legitimate result. minMax
	// guarantees startCol <= endCol, and capping both at the same ceiling keeps
	// that invariant (so the width below is never negative). Mirrors the row
	// dimension, which is already clipped to len(rows) below.
	if startCol > maxExcelColumns {
		startCol = maxExcelColumns
	}
	if endCol > maxExcelColumns {
		endCol = maxExcelColumns
	}
	// Initialize non-nil so a range entirely past the last data row emits
	// "rows": [] rather than "rows": null in the JSON output.
	out := [][]string{}
	for r := startRow; r <= endRow; r++ {
		if r-1 >= len(rows) {
			break
		}
		row := rows[r-1]
		slice := make([]string, 0, endCol-startCol+1)
		for c := startCol; c <= endCol; c++ {
			if c-1 < len(row) {
				slice = append(slice, row[c-1])
			} else {
				slice = append(slice, "")
			}
		}
		out = append(out, slice)
	}
	return out
}

var lineRangePattern = regexp.MustCompile(`^(\d+)\s*-\s*(\d+)$`)

// parseLineRange parses "A-B" or a single "N" into a 1-based inclusive range.
func parseLineRange(spec string) (start, end int, err error) {
	spec = strings.TrimSpace(spec)
	if m := lineRangePattern.FindStringSubmatch(spec); m != nil {
		start, _ = strconv.Atoi(m[1])
		end, _ = strconv.Atoi(m[2])
		if start < 1 || end < start {
			return 0, 0, fmt.Errorf("invalid line range %q", spec)
		}
		return start, end, nil
	}
	if n, e := strconv.Atoi(spec); e == nil && n >= 1 {
		return n, n, nil
	}
	return 0, 0, fmt.Errorf("invalid line range %q (want A-B or N)", spec)
}

var cellRefRangePattern = regexp.MustCompile(`^([A-Za-z]+)(\d+)$`)

// parseCellRange parses "A1:Z100" (or a single "B14") into 1-based column/row
// bounds. Bounds are normalized so the top-left precedes the bottom-right.
func parseCellRange(spec string) (startCol, startRow, endCol, endRow int, err error) {
	spec = strings.TrimSpace(spec)
	lo, hi, ok := strings.Cut(spec, ":")
	if !ok {
		hi = lo
	}
	c1, r1, e := parseCellRef(lo)
	if e != nil {
		return 0, 0, 0, 0, e
	}
	c2, r2, e := parseCellRef(hi)
	if e != nil {
		return 0, 0, 0, 0, e
	}
	startCol, endCol = minMax(c1, c2)
	startRow, endRow = minMax(r1, r2)
	return startCol, startRow, endCol, endRow, nil
}

func parseCellRef(ref string) (col, row int, err error) {
	m := cellRefRangePattern.FindStringSubmatch(strings.TrimSpace(ref))
	if m == nil {
		return 0, 0, fmt.Errorf("invalid cell reference %q (want e.g. B14)", ref)
	}
	col = colIndex(m[1])
	row, _ = strconv.Atoi(m[2])
	if col < 1 || row < 1 {
		return 0, 0, fmt.Errorf("invalid cell reference %q", ref)
	}
	return col, row, nil
}

func minMax(a, b int) (int, int) {
	if a <= b {
		return a, b
	}
	return b, a
}

// pageListToSpec renders a sorted page slice back to a compact LiteParse
// --target-pages spec (just comma-joined; ranges aren't required).
func pageListToSpec(pages []int) string {
	parts := make([]string, len(pages))
	for i, p := range pages {
		parts[i] = strconv.Itoa(p)
	}
	return strings.Join(parts, ",")
}

func fileBase(path string) string {
	// Treat both separators so Windows paths (C:\dir\file.pdf) and URL/POSIX
	// paths yield the base name rather than leaking the full path.
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}
