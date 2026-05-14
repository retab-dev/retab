package cmd

// Output rendering infrastructure for list/get/create commands.
//
// Two formats are supported:
//
//   - OutputJSON — indented JSON to stdout, byte-equivalent to the existing
//     printJSON helper in common.go. This is the lossless, pipe-friendly
//     default whenever stdout is being captured (e.g. `retab files list | jq`).
//   - OutputTable — a compact text table rendered with stdlib text/tabwriter.
//     Intended for interactive TTY use where the user wants to see two or
//     three load-bearing columns instead of every internal field on every
//     record.
//
// OutputAuto picks between the two based on whether the writer is a TTY,
// using the same detection rule as paletteFor in help.go (must be a real
// *os.File whose fd is a terminal). Anything else falls back to JSON so
// redirects, pipes, and bytes.Buffer in tests all stay deterministic.
//
// This file deliberately does NOT wire RenderList into any specific list
// command. The follow-up commit is going to adopt the helper one command
// at a time so the column choices for each resource can be reviewed
// individually — landing both the plumbing and ~20 column specs in a
// single change makes the diff unreviewable.

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// OutputFormat names the rendering mode chosen for a command's output.
type OutputFormat string

const (
	// OutputJSON renders indented JSON, byte-equivalent to printJSON.
	OutputJSON OutputFormat = "json"
	// OutputTable renders a compact text table via text/tabwriter.
	OutputTable OutputFormat = "table"
	// OutputAuto picks between JSON and Table based on whether the writer
	// is a TTY. Never the final stored value — resolved before rendering.
	OutputAuto OutputFormat = "auto"
)

// DefaultOutputFormat returns the format to use when --output is empty.
// Mirrors paletteFor in help.go: only a real *os.File whose fd is a terminal
// gets table mode; bytes.Buffer, os.Pipe write halves, captured stdout in
// tests — all stay on JSON so machine-consumed output is deterministic.
func DefaultOutputFormat(w io.Writer) OutputFormat {
	f, ok := w.(*os.File)
	if !ok {
		return OutputJSON
	}
	if !term.IsTerminal(int(f.Fd())) {
		return OutputJSON
	}
	return OutputTable
}

// TableColumn describes one column in a rendered table.
//
// Extract receives a single row value (whatever element type the list's
// `data` slice contains, decoded as `any`) and returns its display string.
// Implementations should be defensive about types — the row may arrive as
// a map[string]any (JSON round-trip) or as a typed struct pointer (direct
// pass-through), so callers usually handle both via a small helper.
type TableColumn struct {
	Header  string
	Extract func(row any) string
}

// RenderList writes data to w using the given format.
//
// For OutputJSON the output is byte-equivalent to the long-standing
// printJSON helper: indented two spaces, HTML escaping off, trailing
// newline. Existing consumers piping into jq see no change.
//
// For OutputTable the renderer extracts data["data"] as a slice and prints
// one row per element using text/tabwriter. The header row is printed
// first; columns are separated by two spaces (tabwriter padding=2,
// minwidth=0, tabwidth=0). No bold / no underline — the header gains
// emphasis from being first and the row separation alone is enough at
// terminal width. This keeps the output paste-able into issues and
// markdown without ANSI noise.
//
// OutputAuto is resolved against w before dispatching.
func RenderList(w io.Writer, format OutputFormat, data any, columns []TableColumn) error {
	if format == OutputAuto {
		format = DefaultOutputFormat(w)
	}
	switch format {
	case OutputJSON:
		return renderJSON(w, data)
	case OutputTable:
		return renderTable(w, data, columns)
	default:
		return fmt.Errorf("unknown output format: %q", format)
	}
}

// renderJSON encodes v to w with the same settings as printJSON. Kept as a
// dedicated function so it remains trivially comparable to the printJSON
// behaviour in common.go — the pinning test in output_test.go relies on
// the two producing identical bytes for identical inputs.
func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// renderTable extracts the `data` slice from `v` and renders it via
// text/tabwriter. The slice can be reached in either of two ways:
//
//  1. v is a struct (typed list response) with a `Data` field — reflect
//     directly. This is the hot path for typed SDK responses.
//  2. v has no usable struct field — fall back to JSON round-trip into
//     map[string]any, then read map["data"]. Slower but copes with
//     anonymous payloads and decoded `any` values from readJSON.
//
// If `data` is missing, not a slice, or empty, we render only the header
// row. An explicit empty result is a perfectly valid table — printing an
// error or hiding the headers would just be confusing.
func renderTable(w io.Writer, v any, columns []TableColumn) error {
	rows, err := extractDataSlice(v)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Header row.
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, col.Header)
	}
	fmt.Fprintln(tw)

	// Data rows.
	for _, row := range rows {
		for i, col := range columns {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, col.Extract(row))
		}
		fmt.Fprintln(tw)
	}
	return tw.Flush()
}

// extractDataSlice pulls the `data` slice out of a list response.
// Two routes, tried in order:
//
//  1. Reflect on the top-level value (or its pointee) and look for a
//     struct field named `Data`. This matches the SDK's typed responses
//     (e.g. retab.FileListResponse, BlockTestListResponse) without an
//     extra serialize/deserialize roundtrip.
//  2. Marshal to JSON, unmarshal into map[string]any, and read "data".
//     Catches the map-shaped payloads readJSON produces and anything
//     custom a caller passes in.
//
// Returns a slice of `any` so the column Extract funcs see exactly one
// row per element regardless of source. An empty / missing data field
// returns (nil, nil) — tables with zero rows are still a valid render.
func extractDataSlice(v any) ([]any, error) {
	if v == nil {
		return nil, nil
	}

	// Route 1: reflect for a typed struct with a Data field.
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		f := rv.FieldByName("Data")
		if f.IsValid() && (f.Kind() == reflect.Slice || f.Kind() == reflect.Array) {
			n := f.Len()
			out := make([]any, n)
			for i := 0; i < n; i++ {
				out[i] = f.Index(i).Interface()
			}
			return out, nil
		}
	}

	// Route 2: JSON round-trip into a map and read the "data" key.
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("render table: marshal failed: %w", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		// Not a JSON object — nothing to render. Return empty rather
		// than erroring; the header row alone is still a valid table.
		return nil, nil
	}
	data, ok := obj["data"]
	if !ok || data == nil {
		return nil, nil
	}
	slice, ok := data.([]any)
	if !ok {
		return nil, fmt.Errorf("render table: expected `data` to be an array, got %T", data)
	}
	return slice, nil
}

// ResolveOutputFormat reads the --output persistent flag set on the root
// command and resolves it against w. An empty flag means auto-detect.
// Unknown values produce an error so typos surface immediately rather
// than silently falling back to JSON.
//
// In normal CLI use the unknown-value path is unreachable because root.go
// registers --output as a custom pflag.Value that rejects invalid strings
// at parse time. ResolveOutputFormat still keeps the error branch because
// callers may build flag values programmatically (e.g. tests) and rely on
// this function as the single canonical validation point.
func ResolveOutputFormat(cmd *cobra.Command, w io.Writer) (OutputFormat, error) {
	var raw string
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	switch raw {
	case "", "auto":
		return DefaultOutputFormat(w), nil
	case string(OutputJSON):
		return OutputJSON, nil
	case string(OutputTable):
		return OutputTable, nil
	default:
		return "", fmt.Errorf("invalid --output value %q (want: json | table | auto)", raw)
	}
}
