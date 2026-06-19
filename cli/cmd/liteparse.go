package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// LiteParse (https://github.com/run-llama/liteparse) is an Apache-2.0 OSS
// document parser: PDFium text extraction with selective Tesseract OCR, run
// entirely locally. The `files parse|grep|inspect` commands shell out to its
// `lit` CLI for PDFs and images — the formats where local spatial parsing and
// OCR are the value — and keep text/csv/xlsx/docx on native Go readers (see
// native_parse.go), which preserves exact cell/line anchors that a
// LibreOffice->PDF round-trip would destroy.

// ParseOptions controls a LiteParse `lit parse` invocation. The zero value is
// not valid; build it via defaultParseOptions so OCR/DPI/language carry the
// documented defaults.
type ParseOptions struct {
	OCR          bool   // OCR enabled (LiteParse default true)
	OCRLanguage  string // Tesseract language code, e.g. "eng"
	OCRServerURL string // optional HTTP OCR server; empty uses bundled Tesseract
	DPI          int    // rendering DPI for OCR/screenshots
	TargetPages  string // e.g. "1-5,10,15-20"; empty parses all pages
}

func defaultParseOptions() ParseOptions {
	// Latin = the multilingual Latin-script Tesseract model (en/fr/de/es/it/pt/…),
	// matching the AI server's PaddleOCR lang='la' and the bundled
	// Latin.traineddata. Override per-invocation with --ocr-language.
	return ParseOptions{OCR: true, OCRLanguage: "Latin", DPI: 150}
}

// fingerprint is a stable string identifying the parse options, used as part
// of the parse cache key so a language/DPI/OCR change naturally misses.
func (o ParseOptions) fingerprint() string {
	return fmt.Sprintf("ocr=%t;lang=%s;server=%s;dpi=%d;pages=%s",
		o.OCR, o.OCRLanguage, o.OCRServerURL, o.DPI, o.TargetPages)
}

// ParsedItem is a single positioned text fragment from a page. Coordinates
// are LiteParse viewport space (top-left origin, 72 DPI). Confidence is 1.0
// for native PDF text and <1.0 for OCR'd text.
type ParsedItem struct {
	Text       string   `json:"text"`
	X          float64  `json:"x"`
	Y          float64  `json:"y"`
	Width      float64  `json:"width"`
	Height     float64  `json:"height"`
	FontName   string   `json:"font_name,omitempty"`
	FontSize   float64  `json:"font_size,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"`
}

// ParsedPage is one page of layout-preserved text plus its positioned items.
type ParsedPage struct {
	Page   int          `json:"page"`
	Width  float64      `json:"width"`
	Height float64      `json:"height"`
	Text   string       `json:"text"`
	Items  []ParsedItem `json:"items,omitempty"`
}

// SheetData holds the cell grid for a csv (single sheet) or xlsx (one per
// worksheet). Rows are row-major, 0-based; Rows[r][c] is the cell text.
type SheetData struct {
	Index int        `json:"index"`
	Name  string     `json:"name,omitempty"`
	Rows  [][]string `json:"rows"`
}

// ParseResult is the normalized output of parsing any supported file. Pages
// carries the textual/positioned view (always populated); Sheets carries the
// structured cell grid for csv/spreadsheet so grep/inspect can emit cell
// anchors.
type ParseResult struct {
	Filename     string       `json:"filename"`
	MIMEType     string       `json:"mime_type"`
	DocumentType string       `json:"document_type"` // pdf|image|text|csv|spreadsheet|docx
	Source       string       `json:"source"`        // pdf_text_layer|ocr|native
	TotalPages   int          `json:"total_pages"`
	Pages        []ParsedPage `json:"pages"`
	Sheets       []SheetData  `json:"sheets,omitempty"`
}

// ScreenshotOptions controls `lit screenshot`.
type ScreenshotOptions struct {
	TargetPages string
	DPI         int
	OutDir      string
}

// ScreenshotPage is one rendered page image on disk.
type ScreenshotPage struct {
	Page     int    `json:"page"`
	Path     string `json:"path"`
	MIMEType string `json:"mime_type"`
}

// LiteParser is the seam over the `lit` binary. The exec-backed litCLI is the
// production implementation; tests inject a fake to avoid requiring `lit` on
// PATH and to assert argv construction.
type LiteParser interface {
	Parse(ctx context.Context, path string, opt ParseOptions) (*ParseResult, error)
	Screenshot(ctx context.Context, path string, opt ScreenshotOptions) ([]ScreenshotPage, error)
	Version(ctx context.Context) (string, error)
}

// resolveLiteParserFn is the indirection tests override to supply a fake
// without touching PATH. Production resolution honors --liteparse-bin, then
// RETAB_LITEPARSE_BIN, then a `lit` on PATH.
var resolveLiteParserFn = defaultResolveLiteParser

func resolveLiteParser(bin string) (LiteParser, error) { return resolveLiteParserFn(bin) }

func defaultResolveLiteParser(bin string) (LiteParser, error) {
	// An explicitly requested binary (flag or env) must resolve as given; we
	// never silently substitute a managed download for it.
	explicit := bin != "" || os.Getenv("RETAB_LITEPARSE_BIN") != ""
	if bin == "" {
		bin = os.Getenv("RETAB_LITEPARSE_BIN")
	}
	if bin == "" {
		bin = "lit"
	}
	if resolved, err := exec.LookPath(bin); err == nil {
		return &litCLI{bin: resolved}, nil
	}
	if explicit {
		return nil, fmt.Errorf("liteparse binary %q not found; point --liteparse-bin / RETAB_LITEPARSE_BIN at an existing `lit`", bin)
	}

	// No `lit` on PATH and none requested: fall back to the self-contained
	// bundle (lit + libpdfium + OCR data) embedded in this binary, or
	// downloaded + checksum-verified for source builds. errBundleUnavailable
	// means we ship no bundle for this platform — surface the install hint.
	managed, err := resolveManagedLit()
	if err == nil {
		return managed, nil
	}
	if !errors.Is(err, errBundleUnavailable) {
		return nil, err
	}
	return nil, fmt.Errorf(
		"liteparse binary %q not found. Install it with one of:\n"+
			"  npm i -g @llamaindex/liteparse\n"+
			"  pip install liteparse\n"+
			"  cargo install liteparse\n"+
			"or point --liteparse-bin / RETAB_LITEPARSE_BIN at an existing `lit`", bin)
}

// litCLI execs the `lit` binary. pdfiumDir, when non-empty, is the directory
// holding a bundled libpdfium; it's exported to `lit` as PDFIUM_LIB_PATH so a
// managed bundle can dlopen its own pdfium instead of requiring one on the
// system. tessdataDir, when non-empty, is the directory holding the bundled
// Latin.traineddata; it's passed via --tessdata-path so OCR finds its language
// data. For a user-provided `lit` (PATH/flag/env) both are empty and we leave
// the environment + args untouched.
type litCLI struct {
	bin         string
	pdfiumDir   string
	tessdataDir string
}

// command builds an exec.Cmd for `lit`, wiring PDFIUM_LIB_PATH when this is a
// managed bundle so the dlopen'd libpdfium resolves next to the binary.
func (c *litCLI) command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.bin, args...)
	if c.pdfiumDir != "" {
		cmd.Env = append(os.Environ(), "PDFIUM_LIB_PATH="+c.pdfiumDir)
	}
	return cmd
}

// liteParseJSON mirrors `lit parse --format json` output
// (crates/liteparse/src/output/json.rs).
type liteParseJSON struct {
	Pages []struct {
		Page      int     `json:"page"`
		Width     float64 `json:"width"`
		Height    float64 `json:"height"`
		Text      string  `json:"text"`
		TextItems []struct {
			Text       string   `json:"text"`
			X          float64  `json:"x"`
			Y          float64  `json:"y"`
			Width      float64  `json:"width"`
			Height     float64  `json:"height"`
			FontName   string   `json:"fontName"`
			FontSize   float64  `json:"fontSize"`
			Confidence *float64 `json:"confidence"`
		} `json:"textItems"`
	} `json:"pages"`
}

func (c *litCLI) parseArgs(path string, opt ParseOptions) []string {
	args := []string{"parse", path, "--format", "json", "--quiet"}
	if !opt.OCR {
		args = append(args, "--no-ocr")
	} else {
		if opt.OCRLanguage != "" {
			args = append(args, "--ocr-language", opt.OCRLanguage)
		}
		if opt.OCRServerURL != "" {
			args = append(args, "--ocr-server-url", opt.OCRServerURL)
		}
		// Point Tesseract at the bundled language data when this is a managed
		// bundle; a user-provided `lit` uses its own TESSDATA_PREFIX.
		if c.tessdataDir != "" {
			args = append(args, "--tessdata-path", c.tessdataDir)
		}
	}
	if opt.DPI > 0 {
		args = append(args, "--dpi", fmt.Sprintf("%d", opt.DPI))
	}
	if opt.TargetPages != "" {
		args = append(args, "--target-pages", opt.TargetPages)
	}
	return args
}

func (c *litCLI) Parse(ctx context.Context, path string, opt ParseOptions) (*ParseResult, error) {
	// Standalone images are wrapped into a temp PDF so OCR runs through
	// PDFium+Tesseract instead of `lit`'s ImageMagick image path.
	input, cleanup, err := litInputPath(path, opt.DPI)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	cmd := c.command(ctx, c.parseArgs(input, opt)...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("liteparse parse failed: %s", msg)
	}
	var raw liteParseJSON
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("decode liteparse json: %w", err)
	}
	return convertLiteParse(&raw), nil
}

// convertLiteParse maps the `lit` JSON into a ParseResult. document_type and
// filename/mime are filled by the caller; here we only derive Source from
// whether any item carries an OCR confidence (<1.0).
func convertLiteParse(raw *liteParseJSON) *ParseResult {
	result := &ParseResult{Source: "pdf_text_layer", Pages: make([]ParsedPage, 0, len(raw.Pages))}
	ocrUsed := false
	for _, p := range raw.Pages {
		page := ParsedPage{Page: p.Page, Width: p.Width, Height: p.Height, Text: p.Text}
		for _, it := range p.TextItems {
			if it.Confidence != nil && *it.Confidence < 1.0 {
				ocrUsed = true
			}
			page.Items = append(page.Items, ParsedItem{
				Text: it.Text, X: it.X, Y: it.Y, Width: it.Width, Height: it.Height,
				FontName: it.FontName, FontSize: it.FontSize, Confidence: it.Confidence,
			})
		}
		result.Pages = append(result.Pages, page)
	}
	result.TotalPages = len(result.Pages)
	if ocrUsed {
		result.Source = "ocr"
	}
	return result
}

func (c *litCLI) Screenshot(ctx context.Context, path string, opt ScreenshotOptions) ([]ScreenshotPage, error) {
	if opt.OutDir == "" {
		return nil, fmt.Errorf("screenshot output directory is required")
	}
	if err := os.MkdirAll(opt.OutDir, 0o755); err != nil {
		return nil, fmt.Errorf("create screenshot dir: %w", err)
	}
	// Wrap standalone images into a temp PDF (same reason as Parse) so the
	// renderer goes through PDFium rather than `lit`'s ImageMagick path.
	input, cleanup, err := litInputPath(path, opt.DPI)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	args := []string{"screenshot", input, "--output-dir", opt.OutDir, "--quiet"}
	if opt.TargetPages != "" {
		args = append(args, "--target-pages", opt.TargetPages)
	}
	if opt.DPI > 0 {
		args = append(args, "--dpi", fmt.Sprintf("%d", opt.DPI))
	}
	cmd := c.command(ctx, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("liteparse screenshot failed: %s", msg)
	}
	return collectRenderedPages(opt.OutDir, opt.TargetPages)
}

// pageImageNameRe matches the page number liteparse encodes in a rendered image
// filename, e.g. "page_2.png" → 2.
var pageImageNameRe = regexp.MustCompile(`(?i)page[_-]?(\d+)\.[a-z0-9]+$`)

// pageNumberFromImagePath parses the real 1-based page number from a rendered
// image filename. Reporting this number — rather than the image's ordinal
// position among the output files — is what lets `files inspect --render 2-3`
// correctly label page_2.png as page 2 instead of "page 1".
func pageNumberFromImagePath(path string) (int, bool) {
	m := pageImageNameRe.FindStringSubmatch(portableBase(path))
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// collectRenderedPages reads the images liteparse wrote into outDir and returns
// one ScreenshotPage per page, each labeled with the real page number parsed
// from its filename. When targetSpec names specific pages, images for other
// pages already present in outDir (e.g. left by an earlier render into the same
// directory) are excluded — but a requested page whose PNG already existed is
// still reported, so re-rendering into a populated directory is idempotent
// rather than returning an empty/null page list.
func collectRenderedPages(outDir, targetSpec string) ([]ScreenshotPage, error) {
	images, err := imagesIn(outDir)
	if err != nil {
		return nil, err
	}
	requested := map[int]bool{}
	if nums, err := parsePageList(targetSpec); err == nil {
		for _, n := range nums {
			requested[n] = true
		}
	}
	var pages []ScreenshotPage
	for _, p := range images {
		page, ok := pageNumberFromImagePath(p)
		if !ok {
			continue
		}
		if len(requested) > 0 && !requested[page] {
			continue
		}
		pages = append(pages, ScreenshotPage{Page: page, Path: p, MIMEType: mimeForExt(filepath.Ext(p))})
	}
	sort.Slice(pages, func(i, j int) bool { return pages[i].Page < pages[j].Page })
	return pages, nil
}

// pageNumberFromScreenshotFilename extracts N from a `page-<N>.<ext>` file
// name produced by `lit screenshot`. Returns 0 when the name does not match.
func pageNumberFromScreenshotFilename(path string) int {
	base := portableBase(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	const prefix = "page-"
	if !strings.HasPrefix(base, prefix) {
		return 0
	}
	n, err := strconv.Atoi(base[len(prefix):])
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func portableBase(path string) string {
	if i := strings.LastIndexAny(path, `/\`); i >= 0 {
		return path[i+1:]
	}
	return path
}

func imagesIn(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".png", ".jpg", ".jpeg", ".webp":
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

// imageModTimes maps each image file in dir to its modtime (UnixNano). Used to
// tell which screenshots a render run actually produced, including in-place
// overwrites of pre-existing page-N.png files.
func imageModTimes(dir string) map[string]int64 {
	out := map[string]int64{}
	paths, err := imagesIn(dir)
	if err != nil {
		return out
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		out[p] = info.ModTime().UnixNano()
	}
	return out
}

func (c *litCLI) Version(ctx context.Context) (string, error) {
	out, err := c.command(ctx, "--version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// --- parse cache -----------------------------------------------------------

// parseCacheEnabled is flipped off by --no-cache. Caching only applies to the
// liteparse (pdf/image) path; native parsing is cheap enough to always redo.
func liteParseCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "retab", "liteparse"), nil
}

func parseCacheKey(fileBytes []byte, documentType string, opt ParseOptions) string {
	h := sha256.New()
	h.Write(fileBytes)
	h.Write([]byte("\x00" + documentType + "\x00" + opt.fingerprint()))
	return hex.EncodeToString(h.Sum(nil))
}

func readParseCache(key string) (*ParseResult, bool) {
	dir, err := liteParseCacheDir()
	if err != nil {
		return nil, false
	}
	data, err := os.ReadFile(filepath.Join(dir, key+".json"))
	if err != nil {
		return nil, false
	}
	var result ParseResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}
	return &result, true
}

func writeParseCache(key string, result *ParseResult) {
	dir, err := liteParseCacheDir()
	if err != nil {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	data, err := json.Marshal(result)
	if err != nil {
		return
	}
	tmp := filepath.Join(dir, key+".json.tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	if err := os.Rename(tmp, filepath.Join(dir, key+".json")); err != nil {
		// Don't leave the temp file orphaned if the atomic swap fails (e.g. a
		// concurrent reader holds the destination open on Windows).
		_ = os.Remove(tmp)
	}
}
