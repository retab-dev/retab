package cmd

import "strings"

// Anchor is the discriminated location of a grep match inside a parsed file.
// It mirrors the Retab MCP files tools' anchor union so CLI output is
// shape-compatible with consumers of files_grep / files_get_sources:
//
//	pdf_page         -> Page, Line (+ optional Bbox when --bbox)
//	image            -> Page, Line (+ optional Bbox)
//	text_span        -> LineStart, LineEnd, CharStart, CharEnd
//	csv_cell         -> Row, Column, Coordinate
//	spreadsheet_cell -> SheetIndex, SheetName, Row, Column, Coordinate
//
// Kind is always set; the remaining fields are populated per kind and
// omitted otherwise. Char offsets and the sheet index use pointers so a
// legitimate zero value still serializes.
type Anchor struct {
	Kind string `json:"kind"`

	// pdf_page | image
	Page int `json:"page,omitempty"`
	Line int `json:"line,omitempty"`

	// text_span
	LineStart int  `json:"line_start,omitempty"`
	LineEnd   int  `json:"line_end,omitempty"`
	CharStart *int `json:"char_start,omitempty"`
	CharEnd   *int `json:"char_end,omitempty"`

	// csv_cell | spreadsheet_cell
	Row        int    `json:"row,omitempty"`
	Column     string `json:"column,omitempty"`
	Coordinate string `json:"coordinate,omitempty"`
	SheetIndex *int   `json:"sheet_index,omitempty"`
	SheetName  string `json:"sheet_name,omitempty"`

	// Spatial value-add available for pdf/image when --bbox is set. LiteParse
	// hands us per-item coordinates for free, so a grep match can also carry a
	// normalized [0,1] box for callers that want to draw it.
	Bbox *Bbox `json:"bbox,omitempty"`
}

const (
	anchorPDFPage         = "pdf_page"
	anchorImage           = "image"
	anchorTextSpan        = "text_span"
	anchorCSVCell         = "csv_cell"
	anchorSpreadsheetCell = "spreadsheet_cell"
)

// Bbox is a page-relative bounding box with all coordinates normalized to
// [0, 1] (left/top from the page's top-left origin).
type Bbox struct {
	Page   int     `json:"page,omitempty"`
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// colLetter converts a 1-based column index to its spreadsheet letter
// (1->A, 26->Z, 27->AA). A non-positive index returns "".
func colLetter(n int) string {
	if n < 1 {
		return ""
	}
	var b strings.Builder
	for n > 0 {
		n--
		b.WriteByte(byte('A' + n%26))
		n /= 26
	}
	// Built least-significant-first; reverse.
	runes := []byte(b.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// colIndex converts a spreadsheet column letter (A, Z, AA) to its 1-based
// index. Returns 0 for an empty or invalid string.
func colIndex(s string) int {
	s = strings.ToUpper(strings.TrimSpace(s))
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 'A' || c > 'Z' {
			return 0
		}
		n = n*26 + int(c-'A'+1)
	}
	return n
}
