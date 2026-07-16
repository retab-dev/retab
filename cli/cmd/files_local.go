package cmd

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// This file wires the local-first `files parse|grep|inspect` commands
// onto the existing `files` group. They are an intentional exception to the
// "CLI = thin overlay on the generated SDK" rule: they take a LOCAL path,
// never upload, never call the API, and shell out to LiteParse (`lit`) for
// PDFs/images while parsing text/csv/xlsx/docx natively in Go. See
// liteparse.go and native_parse.go for the machinery.
//
// Registration lives in its own untagged init() so these commands compile
// into both the default (`files.go`) and the generated-prototype
// (`zz_oagen_files.go`) builds — both define `filesCmd`, but under mutually
// exclusive build tags, so exactly one exists at link time.

func init() {
	addParseOptionFlags(filesParseCmd)
	filesParseCmd.Flags().String("file", "", "path to the local document (alternative to the positional <path>)")
	filesParseCmd.Flags().String("format", "text", "output format: text | json")
	filesParseCmd.Flags().Bool("bbox", false, "include per-item bounding boxes in JSON output (pdf/image)")
	filesParseCmd.Flags().StringP("out", "o", "", "write output to this path instead of stdout, - for stdout")

	addParseOptionFlags(filesGrepCmd)
	filesGrepCmd.Flags().String("file", "", "path to the local document (alternative to the positional <path>)")
	filesGrepCmd.Flags().Bool("regex", false, "treat the pattern as a regular expression")
	filesGrepCmd.Flags().Bool("case-sensitive", false, "match case-sensitively (default: case-insensitive)")
	filesGrepCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 500, value: "50"}, "max-results", "max matches to return (1-500)")
	filesGrepCmd.Flags().Var(&boundedIntFlagValue{min: 0, max: 500, value: "80"}, "context-chars", "characters of context on each side of a match (0-500)")
	filesGrepCmd.Flags().Bool("bbox", false, "attach a normalized bounding box to pdf/image matches")

	addParseOptionFlags(filesInspectCmd)
	filesInspectCmd.Flags().String("file", "", "path to the local document (alternative to the positional <path>)")
	filesInspectCmd.Flags().String("lines", "", "line range for text/docx, e.g. 10-40 (1-based, inclusive)")
	filesInspectCmd.Flags().String("cells", "", "cell range for csv/xlsx, e.g. A1:Z100")
	filesInspectCmd.Flags().String("sheet", "", "sheet name or 1-based index for xlsx (default: first sheet)")
	filesInspectCmd.Flags().String("render", "", "page range to render for pdf/image, e.g. 1,3,5 (max 3 pages)")
	filesInspectCmd.Flags().StringP("out", "o", "", "output directory for rendered page images (default: a temp dir)")
	filesInspectCmd.MarkFlagsMutuallyExclusive("lines", "cells", "render")

	filesCmd.AddCommand(filesParseCmd, filesGrepCmd, filesInspectCmd)
}

// localFilePath resolves a single local document path for the files
// parse/inspect/upload commands. The path may be given positionally (the
// idiomatic, documented form) or via --file — the flag form mirrors the document
// primitives' --file so muscle memory from `extractions create --file ...` works
// on the local `files` commands too. The positional wins when both are present.
func localFilePath(cmd *cobra.Command, args []string) (string, error) {
	if len(args) >= 1 && strings.TrimSpace(args[0]) != "" {
		return args[0], nil
	}
	if flag, _ := cmd.Flags().GetString("file"); strings.TrimSpace(flag) != "" {
		return flag, nil
	}
	return "", fmt.Errorf("a file path is required: pass it positionally (<path>) or with --file")
}

// localGrepInputs resolves grep's <path> and <pattern>. The path may come from
// the positional <path> or --file; the pattern is always a positional. So both
// `files grep doc.pdf "term"` and `files grep --file doc.pdf "term"` work.
func localGrepInputs(cmd *cobra.Command, args []string) (path string, pattern string, err error) {
	flag, _ := cmd.Flags().GetString("file")
	flag = strings.TrimSpace(flag)
	if flag != "" {
		if len(args) != 1 {
			return "", "", fmt.Errorf("with --file, provide exactly one positional argument: the <pattern>")
		}
		return flag, args[0], nil
	}
	if len(args) != 2 {
		return "", "", fmt.Errorf("provide <path> and <pattern> (or --file <path> with a single <pattern> positional)")
	}
	return args[0], args[1], nil
}

// explicitOutputJSON reports whether the user explicitly set the global
// --output flag to json. Unlike ResolveOutputFormat it does NOT collapse the
// "auto" default (which becomes json on a non-TTY) — only an explicit
// `--output json` counts, so piping `files parse` doesn't silently flip to JSON.
func explicitOutputJSON(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Changed {
		return f.Value.String() == string(OutputJSON)
	}
	if f := rootCmd.PersistentFlags().Lookup("output"); f != nil && f.Changed {
		return f.Value.String() == string(OutputJSON)
	}
	return false
}

// addParseOptionFlags attaches the LiteParse-tuning flags shared by parse,
// grep, and inspect. Defaults mirror defaultParseOptions().
func addParseOptionFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("no-ocr", false, "disable OCR (pdf/image; native text layer only)")
	cmd.Flags().String("ocr-language", "Latin", "Tesseract OCR language code (Latin = multilingual Latin-script; e.g. eng, fra)")
	cmd.Flags().String("ocr-server-url", "", "URL of an external OCR server (default: bundled Tesseract)")
	cmd.Flags().Var(&boundedIntFlagValue{min: 36, max: 600, value: "150"}, "dpi", "rendering DPI for OCR/screenshots (36-600)")
	cmd.Flags().String("pages", "", "limit parsing to these pages, e.g. 1-5,10 (pdf/image)")
	cmd.Flags().String("liteparse-bin", "", "path to the `lit` binary (default: $RETAB_LITEPARSE_BIN or `lit` on PATH)")
	cmd.Flags().Bool("no-cache", false, "skip the on-disk parse cache for pdf/image")
}

// parseOptionsFromCmd builds a ParseOptions from the shared flags.
func parseOptionsFromCmd(cmd *cobra.Command) ParseOptions {
	opt := defaultParseOptions()
	if noOCR, _ := cmd.Flags().GetBool("no-ocr"); noOCR {
		opt.OCR = false
	}
	if lang, _ := cmd.Flags().GetString("ocr-language"); lang != "" {
		opt.OCRLanguage = lang
	}
	if server, _ := cmd.Flags().GetString("ocr-server-url"); server != "" {
		opt.OCRServerURL = server
	}
	if dpi, _ := cmd.Flags().GetInt("dpi"); dpi > 0 {
		opt.DPI = dpi
	}
	if pages, _ := cmd.Flags().GetString("pages"); pages != "" {
		opt.TargetPages = pages
	}
	return opt
}

func liteBinFromCmd(cmd *cobra.Command) string {
	bin, _ := cmd.Flags().GetString("liteparse-bin")
	return bin
}

func useCacheFromCmd(cmd *cobra.Command) bool {
	noCache, _ := cmd.Flags().GetBool("no-cache")
	return !noCache
}

// tableSelected reports whether the user asked for table output on the root
// --output flag. Used by grep/inspect to choose between the MCP-shaped JSON
// (the default, for parity) and a compact TTY table.
func tableSelected(cmd *cobra.Command) bool {
	var w io.Writer
	if cmd != nil {
		w = cmd.OutOrStdout()
	}
	format, err := ResolveOutputFormat(cmd, w)
	if err != nil {
		return false
	}
	return format == OutputTable
}

// maxPageSpecPages bounds how many pages a --render page spec may select.
// Render caps output at 3 pages anyway; this guard only exists to reject
// absurd ranges before they are materialized in memory.
const maxPageSpecPages = 10000

// parsePageList parses a comma/range page spec ("1,3,5-7") into a sorted,
// de-duplicated 1-based page slice. Used by `files inspect --render`. Returns
// an error on malformed input or a zero/negative page.
func parsePageList(spec string) ([]int, error) {
	seen := map[int]bool{}
	var pages []int
	for _, chunk := range strings.Split(spec, ",") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		if lo, hi, ok := strings.Cut(chunk, "-"); ok {
			start, err := strconv.Atoi(strings.TrimSpace(lo))
			if err != nil {
				return nil, fmt.Errorf("invalid page range %q", chunk)
			}
			end, err := strconv.Atoi(strings.TrimSpace(hi))
			if err != nil {
				return nil, fmt.Errorf("invalid page range %q", chunk)
			}
			if start < 1 || end < 1 || end < start {
				return nil, fmt.Errorf("invalid page range %q", chunk)
			}
			for p := start; p <= end; p++ {
				if !seen[p] {
					seen[p] = true
					pages = append(pages, p)
				}
				// Bound the expansion before allocating: a spec like
				// "1-2000000000" would otherwise materialize billions of map
				// entries before any downstream page-count check runs.
				if len(pages) > maxPageSpecPages {
					return nil, fmt.Errorf("page spec %q selects more than %d pages", spec, maxPageSpecPages)
				}
			}
			continue
		}
		p, err := strconv.Atoi(chunk)
		if err != nil || p < 1 {
			return nil, fmt.Errorf("invalid page %q", chunk)
		}
		if !seen[p] {
			seen[p] = true
			pages = append(pages, p)
		}
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages specified")
	}
	sort.Ints(pages)
	return pages, nil
}
