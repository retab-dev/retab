package cmd

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// docKind classifies an input file by extension. PDFs and images go through
// LiteParse; the rest are parsed natively in Go to keep exact cell/line
// anchors (a LibreOffice->PDF round-trip would lose them).
type docKind int

const (
	kindUnknown docKind = iota
	kindPDF
	kindImage
	kindText
	kindCSV
	kindSpreadsheet
	kindDocx
)

func (k docKind) documentType() string {
	switch k {
	case kindPDF:
		return "pdf"
	case kindImage:
		return "image"
	case kindText:
		return "text"
	case kindCSV:
		return "csv"
	case kindSpreadsheet:
		return "spreadsheet"
	case kindDocx:
		return "docx"
	default:
		return "unknown"
	}
}

// usesLiteParse reports whether the kind is parsed by the `lit` binary rather
// than a native Go reader.
func (k docKind) usesLiteParse() bool { return k == kindPDF || k == kindImage }

var textExtensions = map[string]bool{
	".txt": true, ".md": true, ".markdown": true, ".json": true, ".jsonl": true,
	".yaml": true, ".yml": true, ".xml": true, ".html": true, ".htm": true,
	".log": true, ".rst": true, ".text": true, ".ini": true, ".toml": true,
}

func detectKind(path string) docKind {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".pdf":
		return kindPDF
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff", ".tif", ".webp":
		return kindImage
	case ".csv", ".tsv":
		return kindCSV
	case ".xlsx", ".xlsm":
		return kindSpreadsheet
	case ".docx":
		return kindDocx
	default:
		if textExtensions[ext] {
			return kindText
		}
		return kindUnknown
	}
}

var extMIME = map[string]string{
	".pdf": "application/pdf", ".png": "image/png", ".jpg": "image/jpeg",
	".jpeg": "image/jpeg", ".gif": "image/gif", ".bmp": "image/bmp",
	".tiff": "image/tiff", ".tif": "image/tiff", ".webp": "image/webp",
	".csv": "text/csv", ".tsv": "text/tab-separated-values",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".xlsm": "application/vnd.ms-excel.sheet.macroEnabled.12",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".txt":  "text/plain", ".md": "text/markdown", ".json": "application/json",
}

func mimeForExt(ext string) string {
	if m, ok := extMIME[strings.ToLower(ext)]; ok {
		return m
	}
	return "application/octet-stream"
}

// loadParse is the single entry point for parse/grep/inspect. Native kinds are
// parsed directly; pdf/image are routed through LiteParse with a content-hash
// cache (unless useCache is false). The returned ParseResult always has
// Filename/MIMEType/DocumentType populated.
func loadParse(ctx context.Context, path string, kind docKind, liteBin string, opt ParseOptions, useCache bool) (*ParseResult, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	filename := filepath.Base(path)
	mime := mimeForExt(filepath.Ext(path))

	if !kind.usesLiteParse() {
		var (
			result *ParseResult
			err    error
		)
		switch kind {
		case kindText:
			result, err = parseTextFile(path)
		case kindCSV:
			result, err = parseCSVFile(path)
		case kindSpreadsheet:
			result, err = parseXLSXFile(path)
		case kindDocx:
			result, err = parseDocxFile(path)
		default:
			return nil, fmt.Errorf("unsupported file type %q (supported: pdf, images, txt/md/json, csv/tsv, xlsx, docx)", filepath.Ext(path))
		}
		if err != nil {
			return nil, err
		}
		result.Filename = filename
		result.MIMEType = mime
		result.DocumentType = kind.documentType()
		result.Source = "native"
		return result, nil
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	key := parseCacheKey(fileBytes, kind.documentType(), opt)
	if useCache {
		if cached, ok := readParseCache(key); ok {
			return cached, nil
		}
	}
	parser, err := resolveLiteParser(liteBin)
	if err != nil {
		return nil, err
	}
	result, err := parser.Parse(ctx, path, opt)
	if err != nil {
		return nil, err
	}
	result.Filename = filename
	result.MIMEType = mime
	result.DocumentType = kind.documentType()
	if useCache {
		writeParseCache(key, result)
	}
	return result, nil
}

func parseTextFile(path string) (*ParseResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	text := string(data)
	return &ParseResult{
		TotalPages: 1,
		Pages:      []ParsedPage{{Page: 1, Text: text}},
	}, nil
}

func parseCSVFile(path string) (*ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // ragged rows are fine
	reader.LazyQuotes = true
	if strings.EqualFold(filepath.Ext(path), ".tsv") {
		reader.Comma = '\t'
	}
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv %s: %w", path, err)
	}
	return &ParseResult{
		TotalPages: 1,
		Pages:      []ParsedPage{{Page: 1, Text: rowsToText(rows)}},
		Sheets:     []SheetData{{Index: 0, Rows: rows}},
	}, nil
}

func rowsToText(rows [][]string) string {
	var b strings.Builder
	for _, row := range rows {
		b.WriteString(strings.Join(row, "\t"))
		b.WriteByte('\n')
	}
	return b.String()
}

// --- xlsx (zero-dependency: archive/zip + encoding/xml) --------------------

type xlsxWorkbook struct {
	Sheets struct {
		Sheet []struct {
			Name    string `xml:"name,attr"`
			SheetID string `xml:"sheetId,attr"`
			RID     string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
		} `xml:"sheet"`
	} `xml:"sheets"`
}

type xlsxRels struct {
	Rels []struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

type xlsxSharedStrings struct {
	SI []struct {
		// Either a single <t> or rich-text runs <r><t>.
		T string   `xml:"t"`
		R []string `xml:"r>t"`
	} `xml:"si"`
}

func (s xlsxSharedStrings) at(i int) string {
	if i < 0 || i >= len(s.SI) {
		return ""
	}
	si := s.SI[i]
	if len(si.R) > 0 {
		return strings.Join(si.R, "")
	}
	return si.T
}

type xlsxSheetXML struct {
	Rows []struct {
		R     int `xml:"r,attr"`
		Cells []struct {
			R  string `xml:"r,attr"` // cell ref e.g. "B14"
			T  string `xml:"t,attr"` // type; "s" = shared string
			V  string `xml:"v"`      // value (index for shared strings)
			IS struct {
				T string   `xml:"t"`
				R []string `xml:"r>t"`
			} `xml:"is"` // inline string
		} `xml:"c"`
	} `xml:"sheetData>row"`
}

var cellRefPattern = regexp.MustCompile(`^([A-Za-z]+)(\d+)$`)

func parseXLSXFile(path string) (*ParseResult, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open xlsx %s: %w", path, err)
	}
	defer func() { _ = zr.Close() }()

	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[f.Name] = f
	}

	var shared xlsxSharedStrings
	if f := files["xl/sharedStrings.xml"]; f != nil {
		if err := unmarshalZip(f, &shared); err != nil {
			return nil, err
		}
	}

	var workbook xlsxWorkbook
	if f := files["xl/workbook.xml"]; f != nil {
		if err := unmarshalZip(f, &workbook); err != nil {
			return nil, err
		}
	}
	var rels xlsxRels
	if f := files["xl/_rels/workbook.xml.rels"]; f != nil {
		if err := unmarshalZip(f, &rels); err != nil {
			return nil, err
		}
	}
	relTarget := map[string]string{}
	for _, r := range rels.Rels {
		relTarget[r.ID] = r.Target
	}

	result := &ParseResult{}
	for idx, sheet := range workbook.Sheets.Sheet {
		target := relTarget[sheet.RID]
		if target == "" {
			target = fmt.Sprintf("worksheets/sheet%d.xml", idx+1)
		}
		// Targets are usually relative to xl/ ("worksheets/sheet1.xml") but
		// may be absolute ("/xl/worksheets/sheet1.xml"); normalize both.
		rel := strings.TrimPrefix(strings.TrimPrefix(target, "/"), "xl/")
		name := "xl/" + rel
		f := files[name]
		if f == nil {
			f = files["xl/worksheets/sheet"+strconv.Itoa(idx+1)+".xml"]
		}
		if f == nil {
			continue
		}
		rows, err := parseSheetRows(f, &shared)
		if err != nil {
			return nil, err
		}
		result.Sheets = append(result.Sheets, SheetData{Index: idx, Name: sheet.Name, Rows: rows})
	}
	if len(result.Sheets) == 0 {
		return nil, fmt.Errorf("no worksheets found in %s", filepath.Base(path))
	}

	// Build a textual view: one page per sheet.
	for _, s := range result.Sheets {
		header := s.Name
		if header == "" {
			header = fmt.Sprintf("Sheet%d", s.Index+1)
		}
		result.Pages = append(result.Pages, ParsedPage{
			Page: s.Index + 1,
			Text: header + "\n" + rowsToText(s.Rows),
		})
	}
	result.TotalPages = len(result.Pages)
	return result, nil
}

// maxExcelColumns is Excel's hard column limit (XFD). Cell refs beyond it are
// rejected so a crafted/garbage .xlsx (e.g. a ref like "ZZZZZZZ1") cannot drive
// an unbounded per-row slice allocation.
const maxExcelColumns = 16384

// maxExcelRows is Excel's hard row limit (1,048,576). Rows whose r attribute
// exceeds it are dropped so a crafted/garbage .xlsx (e.g. a single row at
// r="999999999") cannot drive an unbounded padding allocation.
const maxExcelRows = 1048576

func parseSheetRows(f *zip.File, shared *xlsxSharedStrings) ([][]string, error) {
	var sx xlsxSheetXML
	if err := unmarshalZip(f, &sx); err != nil {
		return nil, err
	}
	// Place rows by their r attribute so omitted (empty) rows keep their
	// spreadsheet position, mirroring the per-column gap handling below.
	// Sparse sheets routinely skip empty rows in sheetData, so honoring r is
	// what keeps reported coordinates (e.g. "B5") aligned with the UI.
	rowByNum := map[int][]string{}
	maxRow := 0
	implicit := 0 // last assigned row number, for rows missing the r attribute
	for _, row := range sx.Rows {
		// Place cells by their column ref so gaps stay aligned.
		maxCol := 0
		cells := map[int]string{}
		for _, c := range row.Cells {
			col := 0
			if m := cellRefPattern.FindStringSubmatch(c.R); m != nil {
				col = colIndex(m[1])
			}
			if col == 0 || col > maxExcelColumns {
				continue
			}
			value := c.V
			switch c.T {
			case "s":
				if i, err := strconv.Atoi(strings.TrimSpace(c.V)); err == nil {
					value = shared.at(i)
				}
			case "inlineStr":
				if len(c.IS.R) > 0 {
					value = strings.Join(c.IS.R, "")
				} else {
					value = c.IS.T
				}
			}
			cells[col] = value
			if col > maxCol {
				maxCol = col
			}
		}
		rowSlice := make([]string, maxCol)
		for col, v := range cells {
			rowSlice[col-1] = v
		}
		rowNum := row.R
		if rowNum <= 0 {
			rowNum = implicit + 1
		}
		if rowNum > maxExcelRows {
			continue
		}
		implicit = rowNum
		rowByNum[rowNum] = rowSlice
		if rowNum > maxRow {
			maxRow = rowNum
		}
	}
	rows := make([][]string, maxRow)
	for n, rowSlice := range rowByNum {
		rows[n-1] = rowSlice
	}
	return rows, nil
}

func unmarshalZip(f *zip.File, v any) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open %s: %w", f.Name, err)
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read %s: %w", f.Name, err)
	}
	if err := xml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse %s: %w", f.Name, err)
	}
	return nil
}

// --- docx (zero-dependency: zip + word/document.xml) -----------------------

func parseDocxFile(path string) (*ParseResult, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open docx %s: %w", path, err)
	}
	defer func() { _ = zr.Close() }()
	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return nil, fmt.Errorf("not a valid docx (missing word/document.xml): %s", filepath.Base(path))
	}
	rc, err := docFile.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	text, err := docxXMLToText(rc)
	if err != nil {
		return nil, err
	}
	return &ParseResult{
		TotalPages: 1,
		Pages:      []ParsedPage{{Page: 1, Text: text}},
	}, nil
}

// docxXMLToText streams word/document.xml and emits one line per paragraph
// (<w:p>), concatenating the text runs (<w:t>) inside. Tabs (<w:tab>) and
// breaks (<w:br>) become whitespace. This mirrors the server's
// "convert to text then grep" behavior closely enough for line anchors.
func docxXMLToText(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)
	var out strings.Builder
	var para strings.Builder
	inText := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("parse docx xml: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab":
				para.WriteByte('\t')
			case "br", "cr":
				para.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				para.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				out.WriteString(para.String())
				out.WriteByte('\n')
				para.Reset()
			}
		}
	}
	return out.String(), nil
}
