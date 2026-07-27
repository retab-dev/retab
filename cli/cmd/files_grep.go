package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

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
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		path, pattern, err := localGrepInputs(cmd, args)
		if err != nil {
			return err
		}
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

		matches, totalMatches, truncated := grepParseResult(result, kind, matcher, contextChars, maxResults, withBbox)
		out := grepResult{
			Filename:      result.Filename,
			MIMEType:      result.MIMEType,
			DocumentType:  result.DocumentType,
			Pattern:       pattern,
			Regex:         isRegex,
			CaseSensitive: caseSensitive,
			TotalPages:    result.TotalPages,
			TotalMatches:  totalMatches,
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

// grepParseResult dispatches matching by document kind and returns the matches
// retained (at most maxResults), the TOTAL number of matches in the document,
// and whether the retained set was truncated.
//
// The walk deliberately runs to completion instead of stopping one past
// maxResults. It used to stop early, and the caller then reported
// `total_matches: len(matches)` — a count that was by construction identical to
// the number of matches returned, so with `truncated: true` it said "showing 2
// of 2" while the document held more. A field named total_matches has to be the
// document total or it carries no information a caller can't already compute
// from the array length.
//
// Cost: the document is already parsed in memory, so this is a regex walk over
// extracted text; matches beyond maxResults are counted and discarded rather
// than retained, so memory stays bounded by maxResults.
func grepParseResult(result *ParseResult, kind docKind, matcher *regexp.Regexp, contextChars, maxResults int, withBbox bool) ([]grepMatch, int, bool) {
	collector := newGrepCollector(maxResults)
	switch kind {
	case kindCSV, kindSpreadsheet:
		grepSheets(result, kind, matcher, contextChars, collector)
	case kindText, kindDocx:
		grepTextSpans(result, matcher, contextChars, collector)
	default: // pdf, image
		grepPages(result, kind, matcher, contextChars, withBbox, collector)
	}
	return collector.matches, collector.total, collector.total > len(collector.matches)
}

// grepCollector counts every match while materializing only the first
// maxResults of them.
//
// Counting a match must stay O(1). Building one is not: it computes an anchor,
// a context snippet, and — with --bbox — a bounding box unioned from the page's
// text items. Handing the walkers a plain `add(grepMatch)` forced them to build
// every match just to have it thrown away, which made a full walk pathologically
// slow (a 1.4 MB text document with 40k hits took ~30s, and --bbox on a PDF with
// 20k hits ~46s). `collect` takes a constructor instead, so a match past the cap
// costs one increment.
type grepCollector struct {
	matches    []grepMatch
	total      int
	maxResults int
}

func newGrepCollector(maxResults int) *grepCollector {
	if maxResults < 0 {
		// Defensive: make() panics on a negative capacity. The CLI flag is
		// bounded 1..500, but this keeps an internal caller from crashing.
		maxResults = 0
	}
	return &grepCollector{matches: make([]grepMatch, 0, maxResults), maxResults: maxResults}
}

// collecting reports whether the next match will be kept, so a walker can skip
// per-match work (bounding boxes especially) once the cap is reached.
func (c *grepCollector) collecting() bool { return len(c.matches) < c.maxResults }

// collect counts a match and, only while under the cap, builds and keeps it.
func (c *grepCollector) collect(build func() grepMatch) {
	c.total++
	if len(c.matches) < c.maxResults {
		c.matches = append(c.matches, build())
	}
}

// count records a match the caller has already decided not to materialize.
func (c *grepCollector) count() { c.total++ }

// lineScanner converts byte offsets to 1-based line numbers in a single forward
// pass. regexp returns non-overlapping matches in increasing order, so the
// scanner only ever moves forward and the whole walk costs O(len(text)) —
// instead of the O(offset) rescan-from-zero that lineColAt and strings.Count
// were doing once (or twice) per match, which is what made the walk quadratic.
type lineScanner struct {
	text      string
	pos       int
	line      int
	lineStart int
}

func newLineScanner(text string) *lineScanner {
	return &lineScanner{text: text, line: 1}
}

// lineAt returns the 1-based line containing off. off must be >= the previous
// call's off.
func (s *lineScanner) lineAt(off int) int {
	for s.pos < off && s.pos < len(s.text) {
		if s.text[s.pos] == '\n' {
			s.line++
			s.lineStart = s.pos + 1
		}
		s.pos++
	}
	return s.line
}

// lineColAt returns the 1-based line and 0-based rune column of off.
func (s *lineScanner) lineColAt(off int) (int, int) {
	line := s.lineAt(off)
	return line, len([]rune(s.text[s.lineStart:off]))
}

// grepPages matches against per-page projected text and emits pdf_page/image
// anchors (page + 1-based line). With withBbox it also unions the covering
// text_items into a normalized bounding box.
func grepPages(result *ParseResult, kind docKind, matcher *regexp.Regexp, contextChars int, withBbox bool, collector *grepCollector) {
	anchorKind := anchorPDFPage
	if kind == kindImage {
		anchorKind = anchorImage
	}
	for _, page := range result.Pages {
		text := page.Text
		scanner := newLineScanner(text)
		for _, loc := range matcher.FindAllStringIndex(text, -1) {
			if !collector.collecting() {
				// Past the cap. Skipping this is what keeps --bbox usable: a
				// bounding box is unioned from the page's text items, so
				// building one per match on a dense page dominated everything
				// else even though the result was immediately discarded.
				collector.count()
				continue
			}
			line := scanner.lineAt(loc[0])
			anchor := Anchor{Kind: anchorKind, Page: page.Page, Line: line}
			if withBbox {
				if box := boundingBoxForMatch(page, text[loc[0]:loc[1]]); box != nil {
					anchor.Bbox = box
				}
			}
			collector.collect(func() grepMatch {
				return grepMatch{
					Match:   text[loc[0]:loc[1]],
					Content: snippet(text, loc[0], loc[1], contextChars),
					Anchor:  anchor,
				}
			})
		}
	}
}

// grepTextSpans matches against the single text page of a text/docx document
// and emits text_span anchors with 1-based line and 0-based char offsets.
func grepTextSpans(result *ParseResult, matcher *regexp.Regexp, contextChars int, collector *grepCollector) {
	if len(result.Pages) == 0 {
		return
	}
	text := result.Pages[0].Text
	scanner := newLineScanner(text)
	for _, loc := range matcher.FindAllStringIndex(text, -1) {
		if !collector.collecting() {
			// Past the cap: count it and skip the anchor/snippet work entirely.
			collector.count()
			continue
		}
		startLine, startCol := scanner.lineColAt(loc[0])
		endLine, endCol := scanner.lineColAt(loc[1])
		collector.collect(func() grepMatch {
			return grepMatch{
				Match:   text[loc[0]:loc[1]],
				Content: snippet(text, loc[0], loc[1], contextChars),
				Anchor: Anchor{
					Kind:      anchorTextSpan,
					LineStart: startLine,
					LineEnd:   endLine,
					CharStart: ptr(startCol),
					CharEnd:   ptr(endCol),
				},
			}
		})
	}
}

// grepSheets matches each cell value and emits csv_cell or spreadsheet_cell
// anchors. Row is 1-based (matching spreadsheet UIs); Column is the letter.
func grepSheets(result *ParseResult, kind docKind, matcher *regexp.Regexp, contextChars int, collector *grepCollector) {
	for _, sheet := range result.Sheets {
		for r, row := range sheet.Rows {
			for c, cell := range row {
				if cell == "" {
					continue
				}
				for _, loc := range matcher.FindAllStringIndex(cell, -1) {
					if !collector.collecting() {
						collector.count()
						continue
					}
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
					collector.collect(func() grepMatch {
						return grepMatch{
							Match:   cell[loc[0]:loc[1]],
							Content: snippet(cell, loc[0], loc[1], contextChars),
							Anchor:  anchor,
						}
					})
				}
			}
		}
	}
}

// snippet returns contextChars of context on each side of [start,end), clamped
// to the text bounds. contextChars counts whole characters (runes), not bytes,
// matching the flag's documented unit — walking rune-by-rune also guarantees
// the window edges land on rune boundaries so multibyte characters aren't split.
func snippet(text string, start, end, contextChars int) string {
	if contextChars <= 0 {
		return text[start:end]
	}
	lo := start
	for n := 0; n < contextChars && lo > 0; n++ {
		_, size := utf8.DecodeLastRuneInString(text[:lo])
		lo -= size
	}
	hi := end
	for n := 0; n < contextChars && hi < len(text); n++ {
		_, size := utf8.DecodeRuneInString(text[hi:])
		hi += size
	}
	return strings.TrimSpace(text[lo:hi])
}

// boundingBoxForMatch finds the page text_items whose concatenated text covers
// the matched phrase and unions their boxes into a normalized [0,1] bbox. It
// is a best-effort port of LiteParse's search_items merge: items are scanned in
// order, and the first contiguous run whose joined text contains the match
// wins. Returns nil when no covering run is found or the page has no items.
// grepBoundingBoxCalls counts boundingBoxForMatch invocations. It is a test
// observability hook: computing a bounding box is the dominant per-match cost
// (an O(items) union over the page's text items), so a test asserts this counter
// stays at the retained-match count rather than the document total, proving the
// walk skips that work past --max-results. Deterministic where a wall-clock
// threshold would be flaky.
var grepBoundingBoxCalls int

func boundingBoxForMatch(page ParsedPage, match string) *Bbox {
	grepBoundingBoxCalls++
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
