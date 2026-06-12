package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// grepMatch is one hit: the matched substring, a context snippet, and the
// format-specific anchor pointing at where it lives in the document.
type grepMatch struct {
	Match   string `json:"match"`
	Content string `json:"content"`
	Anchor  Anchor `json:"anchor"`
}

// grepResult mirrors the Retab MCP files_grep response shape so CLI output is
// drop-in compatible with consumers of the server tool (modulo file_id -> a
// local filename).
type grepResult struct {
	Filename      string      `json:"filename"`
	MIMEType      string      `json:"mime_type"`
	DocumentType  string      `json:"document_type"`
	Pattern       string      `json:"pattern"`
	Regex         bool        `json:"regex"`
	CaseSensitive bool        `json:"case_sensitive"`
	TotalPages    int         `json:"total_pages"`
	TotalMatches  int         `json:"total_matches"`
	Truncated     bool        `json:"truncated"`
	Matches       []grepMatch `json:"matches"`
}

var filesGrepCmd = &cobra.Command{
	Use:   "grep <path> <pattern>",
	Short: "Search a local document for a pattern with format-aware anchors",
	Long: `Search a local document for a pattern and return matches with anchors
that point at where each hit lives, entirely locally — no upload, no API call.

This mirrors the Retab MCP files_grep tool. The anchor shape depends on the
document type:

  pdf            -> pdf_page{page, line}      (+ bbox with --bbox)
  image          -> image{page, line}         (+ bbox with --bbox)
  text/md/json   -> text_span{line_start, line_end, char_start, char_end}
  csv/tsv        -> csv_cell{row, column, coordinate}
  xlsx           -> spreadsheet_cell{sheet_index, sheet_name, row, column, coordinate}

Patterns are literal by default (case-insensitive); pass --regex for a Go
regular expression and --case-sensitive to match case.`,
	Example: `  # Find a literal token in a PDF
  retab files grep invoice.pdf "Total Due"

  # Regex, case-sensitive, with bounding boxes
  retab files grep invoice.pdf "INV-\d+" --regex --case-sensitive --bbox

  # Search a spreadsheet (returns cell anchors)
  retab files grep data.xlsx 42000`,
	Args: cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		path, pattern := args[0], args[1]
		kind := detectKind(path)
		if kind == kindUnknown {
			return fmt.Errorf("unsupported file type for %s (supported: pdf, images, txt/md/json, csv/tsv, xlsx, docx)", path)
		}
		if pattern == "" {
			return fmt.Errorf("pattern must not be empty")
		}

		isRegex, _ := cmd.Flags().GetBool("regex")
		caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")
		maxResults, _ := cmd.Flags().GetInt("max-results")
		contextChars, _ := cmd.Flags().GetInt("context-chars")
		withBbox, _ := cmd.Flags().GetBool("bbox")

		matcher, err := buildMatcher(pattern, isRegex, caseSensitive)
		if err != nil {
			return err
		}

		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := loadParse(ctx, path, kind, liteBinFromCmd(cmd), parseOptionsFromCmd(cmd), useCacheFromCmd(cmd))
		if err != nil {
			return err
		}

		matches, truncated := grepParseResult(result, kind, matcher, contextChars, maxResults, withBbox)
		out := grepResult{
			Filename:      result.Filename,
			MIMEType:      result.MIMEType,
			DocumentType:  result.DocumentType,
			Pattern:       pattern,
			Regex:         isRegex,
			CaseSensitive: caseSensitive,
			TotalPages:    result.TotalPages,
			TotalMatches:  len(matches),
			Truncated:     truncated,
			Matches:       matches,
		}

		if tableSelected(cmd) {
			return renderGrepTable(cmd, out)
		}
		return printJSON(out)
	}),
}

// buildMatcher compiles the search pattern. Literal patterns are quoted so
// regex metacharacters are matched verbatim; case-insensitive search prepends
// the (?i) flag.
func buildMatcher(pattern string, isRegex, caseSensitive bool) (*regexp.Regexp, error) {
	expr := pattern
	if !isRegex {
		expr = regexp.QuoteMeta(pattern)
	}
	if !caseSensitive {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}
	return re, nil
}

// grepParseResult dispatches matching by document kind and returns the
// collected matches plus whether the result was truncated at maxResults.
func grepParseResult(result *ParseResult, kind docKind, matcher *regexp.Regexp, contextChars, maxResults int, withBbox bool) ([]grepMatch, bool) {
	var matches []grepMatch
	// Collect one past maxResults so we can distinguish "exactly filled" from
	// "there were more": the caller trims to maxResults and reports truncation.
	add := func(m grepMatch) bool {
		matches = append(matches, m)
		return len(matches) <= maxResults
	}
	switch kind {
	case kindCSV, kindSpreadsheet:
		grepSheets(result, kind, matcher, contextChars, add)
	case kindText, kindDocx:
		grepTextSpans(result, matcher, contextChars, add)
	default: // pdf, image
		grepPages(result, kind, matcher, contextChars, withBbox, add)
	}
	if len(matches) > maxResults {
		return matches[:maxResults], true
	}
	return matches, false
}

// grepPages matches against per-page projected text and emits pdf_page/image
// anchors (page + 1-based line). With withBbox it also unions the covering
// text_items into a normalized bounding box.
func grepPages(result *ParseResult, kind docKind, matcher *regexp.Regexp, contextChars int, withBbox bool, add func(grepMatch) bool) {
	anchorKind := anchorPDFPage
	if kind == kindImage {
		anchorKind = anchorImage
	}
	for _, page := range result.Pages {
		text := page.Text
		for _, loc := range matcher.FindAllStringIndex(text, -1) {
			line := strings.Count(text[:loc[0]], "\n") + 1
			anchor := Anchor{Kind: anchorKind, Page: page.Page, Line: line}
			if withBbox {
				if box := boundingBoxForMatch(page, text[loc[0]:loc[1]]); box != nil {
					anchor.Bbox = box
				}
			}
			cont := add(grepMatch{
				Match:   text[loc[0]:loc[1]],
				Content: snippet(text, loc[0], loc[1], contextChars),
				Anchor:  anchor,
			})
			if !cont {
				return
			}
		}
	}
}

// grepTextSpans matches against the single text page of a text/docx document
// and emits text_span anchors with 1-based line and 0-based char offsets.
func grepTextSpans(result *ParseResult, matcher *regexp.Regexp, contextChars int, add func(grepMatch) bool) {
	if len(result.Pages) == 0 {
		return
	}
	text := result.Pages[0].Text
	for _, loc := range matcher.FindAllStringIndex(text, -1) {
		startLine, startCol := lineColAt(text, loc[0])
		endLine, endCol := lineColAt(text, loc[1])
		anchor := Anchor{
			Kind:      anchorTextSpan,
			LineStart: startLine,
			LineEnd:   endLine,
			CharStart: ptr(startCol),
			CharEnd:   ptr(endCol),
		}
		cont := add(grepMatch{
			Match:   text[loc[0]:loc[1]],
			Content: snippet(text, loc[0], loc[1], contextChars),
			Anchor:  anchor,
		})
		if !cont {
			return
		}
	}
}

// grepSheets matches each cell value and emits csv_cell or spreadsheet_cell
// anchors. Row is 1-based (matching spreadsheet UIs); Column is the letter.
func grepSheets(result *ParseResult, kind docKind, matcher *regexp.Regexp, contextChars int, add func(grepMatch) bool) {
	for _, sheet := range result.Sheets {
		for r, row := range sheet.Rows {
			for c, cell := range row {
				if cell == "" {
					continue
				}
				for _, loc := range matcher.FindAllStringIndex(cell, -1) {
					col := colLetter(c + 1)
					coord := fmt.Sprintf("%s%d", col, r+1)
					anchor := Anchor{
						Row:        r + 1,
						Column:     col,
						Coordinate: coord,
					}
					if kind == kindCSV {
						anchor.Kind = anchorCSVCell
					} else {
						anchor.Kind = anchorSpreadsheetCell
						anchor.SheetIndex = ptr(sheet.Index)
						anchor.SheetName = sheet.Name
					}
					cont := add(grepMatch{
						Match:   cell[loc[0]:loc[1]],
						Content: snippet(cell, loc[0], loc[1], contextChars),
						Anchor:  anchor,
					})
					if !cont {
						return
					}
				}
			}
		}
	}
}

// lineColAt returns the 1-based line and 0-based column (rune offset within
// the line) of byte offset off in text.
func lineColAt(text string, off int) (line, col int) {
	line = 1
	lineStart := 0
	for i := 0; i < off && i < len(text); i++ {
		if text[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	col = len([]rune(text[lineStart:off]))
	return line, col
}

// snippet returns contextChars of context on each side of [start,end), clamped
// to the text bounds and trimmed to rune boundaries so multibyte characters
// aren't split.
func snippet(text string, start, end, contextChars int) string {
	if contextChars <= 0 {
		return text[start:end]
	}
	lo := start - contextChars
	if lo < 0 {
		lo = 0
	}
	hi := end + contextChars
	if hi > len(text) {
		hi = len(text)
	}
	lo = alignRuneStart(text, lo)
	hi = alignRuneStart(text, hi)
	return strings.TrimSpace(text[lo:hi])
}

// alignRuneStart moves i backward to the nearest UTF-8 rune boundary.
func alignRuneStart(s string, i int) int {
	for i > 0 && i < len(s) && (s[i]&0xC0) == 0x80 {
		i--
	}
	return i
}

// boundingBoxForMatch finds the page text_items whose concatenated text covers
// the matched phrase and unions their boxes into a normalized [0,1] bbox. It
// is a best-effort port of LiteParse's search_items merge: items are scanned in
// order, and the first contiguous run whose joined text contains the match
// wins. Returns nil when no covering run is found or the page has no items.
func boundingBoxForMatch(page ParsedPage, match string) *Bbox {
	if len(page.Items) == 0 || page.Width <= 0 || page.Height <= 0 {
		return nil
	}
	needle := strings.ToLower(strings.Join(strings.Fields(match), " "))
	if needle == "" {
		return nil
	}
	for i := range page.Items {
		var joined strings.Builder
		minX, minY := page.Items[i].X, page.Items[i].Y
		maxX := page.Items[i].X + page.Items[i].Width
		maxY := page.Items[i].Y + page.Items[i].Height
		for j := i; j < len(page.Items); j++ {
			it := page.Items[j]
			if joined.Len() > 0 {
				joined.WriteByte(' ')
			}
			// Collapse internal whitespace the same way `needle` is normalized
			// (strings.Fields), otherwise an item containing multiple spaces,
			// tabs, or newlines would never substring-match the needle.
			joined.WriteString(strings.ToLower(strings.Join(strings.Fields(it.Text), " ")))
			if it.X < minX {
				minX = it.X
			}
			if it.Y < minY {
				minY = it.Y
			}
			if it.X+it.Width > maxX {
				maxX = it.X + it.Width
			}
			if it.Y+it.Height > maxY {
				maxY = it.Y + it.Height
			}
			if strings.Contains(joined.String(), needle) {
				return &Bbox{
					Page:   page.Page,
					Left:   clamp01(minX / page.Width),
					Top:    clamp01(minY / page.Height),
					Width:  clamp01((maxX - minX) / page.Width),
					Height: clamp01((maxY - minY) / page.Height),
				}
			}
			// Stop growing a run once it clearly overshoots the needle length.
			if joined.Len() > len(needle)+64 {
				break
			}
		}
	}
	return nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// renderGrepTable prints a compact one-row-per-match table for TTY use. The
// JSON output remains the default (and the parity surface); this is a
// convenience for interactive scanning.
func renderGrepTable(cmd *cobra.Command, result grepResult) error {
	rows := make([]any, len(result.Matches))
	for i, m := range result.Matches {
		rows[i] = map[string]any{
			"location": anchorLocation(m.Anchor),
			"match":    m.Match,
			"content":  m.Content,
		}
	}
	columns := []TableColumn{
		{Header: "LOCATION", Extract: func(r any) string { return r.(map[string]any)["location"].(string) }},
		{Header: "MATCH", Extract: func(r any) string { return r.(map[string]any)["match"].(string) }},
		{Header: "CONTENT", Extract: func(r any) string { return flattenWhitespace(r.(map[string]any)["content"].(string)) }},
	}
	return renderAutoTable(cmd.OutOrStdout(), rows, columns)
}

// flattenWhitespace collapses runs of whitespace (including the newlines that
// the context window pulls in) into single spaces so a match renders on one
// table row instead of wrapping and breaking column alignment. The JSON output
// keeps the original content verbatim; this is table-view-only cosmetics.
func flattenWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// anchorLocation renders a short human-readable location for an anchor, used
// in the table view.
func anchorLocation(a Anchor) string {
	switch a.Kind {
	case anchorPDFPage, anchorImage:
		return fmt.Sprintf("p%d:L%d", a.Page, a.Line)
	case anchorTextSpan:
		return fmt.Sprintf("L%d:%d", a.LineStart, deref(a.CharStart))
	case anchorCSVCell:
		return a.Coordinate
	case anchorSpreadsheetCell:
		if a.SheetName != "" {
			return a.SheetName + "!" + a.Coordinate
		}
		return a.Coordinate
	default:
		return a.Kind
	}
}

func deref(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
