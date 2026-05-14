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

// preferredColumnOrder lists the field names a generic auto-column table
// will pull from each row, in priority order. The first up to 5 names
// present on the row become the rendered columns. Aliases (filename for
// name, status for type) accept either spelling but only one wins per
// row so we never duplicate columns.
//
// The list is intentionally short and curated: id and name/filename are
// the load-bearing identifiers, type/status describes what kind of thing
// the row is, model is a hot field on extractions / parses, and
// created_at gives temporal ordering. Anything else falls back to JSON
// — the goal is a useful default for the five wired list commands, not
// per-resource customization (TableColumn already covers that path).
var preferredColumnOrder = []struct {
	header  string
	aliases []string
}{
	{"ID", []string{"id"}},
	{"NAME", []string{"name", "filename"}},
	{"TYPE", []string{"type", "status"}},
	{"MODEL", []string{"model"}},
	{"CREATED_AT", []string{"created_at"}},
}

// maxAutoColumns is the per-row column cap. Five matches
// preferredColumnOrder; the constant exists so the truncation rule is
// findable from one place.
const maxAutoColumns = 5

// autoTableTruncate is the per-cell width cap for trailing string
// columns. ~40 chars + ellipsis is enough to identify a row without
// blowing out terminal width on long URLs / filenames.
const autoTableTruncate = 40

// printResultTable renders v as a fixed-width text table to stdout when
// the shape is tabulable, falling back to printJSON otherwise.
//
// Tabulable shapes:
//
//  1. {"data": [...]} where the array contains objects. Reached via
//     extractDataSlice, which already handles the typed struct + JSON
//     round-trip cases used by RenderList.
//  2. A bare [] of objects passed directly. Handled by a separate
//     reflect path so callers don't have to wrap their slice in
//     {"data": ...} just to get a table.
//
// Non-tabulable shapes (single object, primitives, deeply nested
// without a usable data array, objects with no preferred-column keys)
// emit a single-line stderr warning and fall through to printJSON.
// The warning keeps the JSON output on stdout pipe-clean — anything
// piped into jq still works, the user just learns their --output
// table preference didn't apply this time.
//
// Column selection: scans the first row's keys, picks up to
// maxAutoColumns columns from preferredColumnOrder. Each preferred
// header collapses its aliases (e.g. "name" and "filename" both map
// to NAME) so a list whose rows have one or the other gets a single
// column rather than two half-empty ones.
func printResultTable(v any) error {
	rows, err := extractTabulableRows(v)
	if err != nil {
		return err
	}
	if rows == nil {
		fmt.Fprintln(os.Stderr, "note: --output table not applicable, falling back to json")
		return printJSON(v)
	}

	columns := pickAutoColumns(rows)
	if len(columns) == 0 {
		// Tabulable shape but no preferred column matched on any row —
		// rendering a header-less table would be more confusing than
		// useful. Fall back to JSON with the same warning so the user
		// knows their preference didn't apply.
		fmt.Fprintln(os.Stderr, "note: --output table not applicable, falling back to json")
		return printJSON(v)
	}

	// Empty data array is a valid tabulable shape — render headers alone.
	return renderAutoTable(os.Stdout, rows, columns)
}

// printResult dispatches between printResultTable and printJSON based
// on the --output persistent flag on the root command.
//
// Mapping:
//
//   - "table"           → printResultTable
//   - "json" / "" / "auto" → printJSON
//
// Per the bug spec, "auto" does NOT consult TTY state here — it's
// treated identically to "" (i.e. JSON). The DefaultOutputFormat /
// RenderList path that DOES auto-detect is preserved for callers
// using TableColumn specs; the generic path is intentionally simpler.
//
// Unknown --output values shouldn't reach this function because
// outputFlagValue rejects them at parse time. We still fall through
// to printJSON rather than erroring so programmatic callers (tests,
// embedding) get a sensible default.
func printResult(cmd *cobra.Command, v any) error {
	var raw string
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	if raw == string(OutputTable) {
		return printResultTable(v)
	}
	return printJSON(v)
}

// extractTabulableRows tries to coerce v into a slice of row values
// suitable for the auto-column table renderer. Returns:
//
//   - rows != nil → tabulable; render this slice (may be empty).
//   - rows == nil → not tabulable; caller falls back to JSON.
//   - err != nil  → genuine extraction failure (rare; extractDataSlice
//     only errors when the shape claims to be a list but the data
//     field is the wrong type).
//
// Two routes, tried in order:
//
//  1. extractDataSlice — covers {"data": [...]} for both typed list
//     responses (struct with Data field) and JSON-round-tripped maps.
//  2. Reflect on v for a top-level slice/array of objects (bare []
//     payloads passed directly).
//
// A single object, primitive, or anything that doesn't fit either
// shape returns (nil, nil) so the caller falls back to JSON.
func extractTabulableRows(v any) ([]any, error) {
	if v == nil {
		return nil, nil
	}

	// Route 1: {"data": [...]} via the existing extractor.
	if rows, err := extractDataSlice(v); err != nil {
		return nil, err
	} else if rows != nil {
		// rows can be empty — that's still a valid tabulable shape
		// (header-only render). extractDataSlice returns (nil, nil)
		// only when there was no data field at all.
		return rows, nil
	}

	// Route 2: bare slice / array.
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		// Only treat slices of objects (struct or map) as tabulable.
		// A []string or []int has no column structure to render.
		n := rv.Len()
		if n == 0 {
			// Empty top-level slice — still a valid (header-only) table.
			return []any{}, nil
		}
		first := rv.Index(0).Interface()
		if !rowIsObject(first) {
			return nil, nil
		}
		out := make([]any, n)
		for i := 0; i < n; i++ {
			out[i] = rv.Index(i).Interface()
		}
		return out, nil
	}

	return nil, nil
}

// rowIsObject reports whether row looks tabulable — a struct or
// map[string]any. Primitive types and []byte / strings are rejected
// so we don't try to render a flat list as a table of one column.
func rowIsObject(row any) bool {
	if row == nil {
		return false
	}
	rv := reflect.ValueOf(row)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if !rv.IsValid() || rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct:
		return true
	case reflect.Map:
		// Only string-keyed maps make sense as table rows.
		return rv.Type().Key().Kind() == reflect.String
	default:
		return false
	}
}

// pickAutoColumns selects up to maxAutoColumns columns from
// preferredColumnOrder by inspecting which aliases appear on at
// least one row's flattened key set.
//
// We scan every row (not just the first) because typed SDK structs
// sometimes omit zero-valued fields after JSON round-trip — a row
// missing "name" but with "filename" should still produce a NAME
// column. A column is included once any row exposes one of its
// aliases.
//
// Only the last (trailing) column gets per-cell truncation —
// truncating internal columns would mis-align everything to the
// right of them, which is uglier than letting tabwriter spread the
// columns wider. The trailing column is the natural place to absorb
// long values (timestamps, URLs, descriptions).
func pickAutoColumns(rows []any) []TableColumn {
	// Build the set of keys present across all rows. Use the alias
	// itself as the key (not the header) so lookup during Extract
	// is a plain map read.
	present := make(map[string]bool)
	for _, row := range rows {
		for _, k := range rowKeys(row) {
			present[k] = true
		}
	}

	// Collect the matching alias-sets in priority order, capped at
	// maxAutoColumns. We materialize the alias slice for each pick
	// here so the closure below captures by value rather than
	// re-reading preferredColumnOrder.
	type pick struct {
		header  string
		aliases []string
	}
	var picks []pick
	for _, spec := range preferredColumnOrder {
		hit := false
		for _, alias := range spec.aliases {
			if present[alias] {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		picks = append(picks, pick{header: spec.header, aliases: spec.aliases})
		if len(picks) == maxAutoColumns {
			break
		}
	}

	cols := make([]TableColumn, 0, len(picks))
	for i, p := range picks {
		aliases := p.aliases
		isTrailing := i == len(picks)-1
		cols = append(cols, TableColumn{
			Header: p.header,
			Extract: func(row any) string {
				for _, alias := range aliases {
					if v, ok := rowField(row, alias); ok {
						s := stringifyCell(v)
						if isTrailing && len(s) > autoTableTruncate {
							return s[:autoTableTruncate] + "…"
						}
						return s
					}
				}
				return ""
			},
		})
	}
	return cols
}

// rowKeys returns the JSON-visible field names of a row, normalized
// to lowercase. Struct rows are inspected via reflect using the
// `json` tag (falling back to the field name lowercased); map rows
// return their keys directly.
func rowKeys(row any) []string {
	if row == nil {
		return nil
	}
	rv := reflect.ValueOf(row)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if !rv.IsValid() || rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct:
		t := rv.Type()
		out := make([]string, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			out = append(out, jsonFieldName(f))
		}
		return out
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return nil
		}
		out := make([]string, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out = append(out, iter.Key().String())
		}
		return out
	default:
		return nil
	}
}

// jsonFieldName extracts the JSON field name for a struct field,
// matching encoding/json's resolution. Omitted (`json:"-"`) fields
// return an empty string which the caller skips.
func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if tag == "" {
		return f.Name
	}
	// Split on ',' to drop options like omitempty.
	if i := indexByte(tag, ','); i >= 0 {
		tag = tag[:i]
	}
	if tag == "" {
		return f.Name
	}
	return tag
}

// indexByte is strings.IndexByte inlined to avoid a strings import in
// this hot path — keeps the file's dependency footprint minimal.
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// rowField returns the value at key on row, or (nil, false) if the
// row doesn't expose that key.
func rowField(row any, key string) (any, bool) {
	if row == nil {
		return nil, false
	}
	rv := reflect.ValueOf(row)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if !rv.IsValid() || rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct:
		t := rv.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if jsonFieldName(f) == key {
				return rv.Field(i).Interface(), true
			}
		}
		return nil, false
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return nil, false
		}
		mv := rv.MapIndex(reflect.ValueOf(key))
		if !mv.IsValid() {
			return nil, false
		}
		return mv.Interface(), true
	default:
		return nil, false
	}
}

// stringifyCell renders a single cell value as plain text. Strings
// pass through unchanged; numbers and bools use fmt's default; nil
// renders as empty; everything else falls through to fmt.Sprintf("%v")
// for a reasonable approximation. Nested structs / maps print as
// Go's default format — acceptable because the auto-column path
// only targets the five preferred columns, none of which are
// expected to carry nested values.
func stringifyCell(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	}
	return fmt.Sprintf("%v", v)
}

// renderAutoTable writes the header + rows to w using text/tabwriter.
// Mirrors renderTable's settings (padding=2, tab separator) so the
// column-aligned look is identical between the auto-column and
// explicit-TableColumn paths.
func renderAutoTable(w io.Writer, rows []any, columns []TableColumn) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, col.Header)
	}
	fmt.Fprintln(tw)
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
