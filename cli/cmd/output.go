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
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

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
// row plus a "(no rows)" hint on stderr so the lone header isn't easily
// mistaken for "all rows hidden" (a UX trap copied from kubectl's
// "No resources found." behaviour). The hint goes to stderr so stdout
// stays clean for piping.
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
	if err := tw.Flush(); err != nil {
		return err
	}
	if len(rows) == 0 {
		emitEmptyRowsHint()
	}
	return nil
}

// emitEmptyRowsHint writes the single-line "(no rows)" indicator to
// stderr when a table render produced a header-only output. Kept in one
// place so the wording stays consistent between renderTable and
// renderAutoTable. Stderr (not stdout) keeps `--output table` pipeable
// — a downstream consumer awk-ing the header line still sees clean
// stdout, and the hint shows up in interactive use.
func emitEmptyRowsHint() {
	fmt.Fprintln(os.Stderr, "(no rows)")
}

// extractDataSlice pulls the `data` slice out of a list response.
// Two routes, tried in order:
//
//  1. Reflect on the top-level value (or its pointee) and look for a
//     struct field named `Data`. This matches the SDK's typed responses
//     (e.g. retab.FileListResponse, WorkflowTestListResponse) without an
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
			for i := range n {
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
	if raw == "" && cmd != rootCmd {
		if f := rootCmd.PersistentFlags().Lookup("output"); f != nil {
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
// preferredColumnOrder is the alias map for the auto-column table
// renderer. Each header collapses multiple JSON field names so different
// resource shapes (artifacts use “operation“, spec changes use
// “target“ / “actions“, etc.) light up under one consistent column
// in the TTY default. Without these aliases the TTY default for
// “spec plan“ and “artifacts list“ rendered empty TYPE/ACTION cells
// even though the JSON carries the value.
var preferredColumnOrder = []struct {
	header  string
	aliases []string
}{
	{"ID", []string{"id", "step_id", "block_id", "workflow_id", "run_id"}},
	{"SOURCE", []string{"source_block", "source.type", "source"}},
	{"TARGET", []string{"target_block", "target.block_id", "target", "address"}},
	{"NAME", []string{"name", "filename", "block_label", "name_at_run_time", "workflow.name_at_run_time", "summary"}},
	// ``operation`` covers workflow artifacts (parse/extract/split/etc.);
	// ``target_type``/``target`` covers spec-plan resource_changes; the
	// trailing ``actions`` (array of {add|update|delete}) also rounds
	// back into TYPE when the change-shape is what's being listed.
	{"TYPE", []string{"type", "status", "block_type", "trigger_type", "lifecycle.status", "trigger.type", "operation", "actions"}},
	{"MODEL", []string{"model"}},
	// Timestamp columns are split per field so the rendered header matches the
	// field actually shown. Lying about which timestamp a column shows is
	// worse than choosing a different column on a different response.
	{"CREATED_AT", []string{"created_at", "timing.created_at"}},
	{"STARTED_AT", []string{"started_at"}},
	{"UPDATED_AT", []string{"updated_at"}},
	{"COMPLETED_AT", []string{"completed_at"}},
}

var defaultEmptyAutoColumns = []TableColumn{
	{Header: "ID", Extract: func(any) string { return "" }},
	{Header: "NAME", Extract: func(any) string { return "" }},
	{Header: "TYPE", Extract: func(any) string { return "" }},
	{Header: "CREATED_AT", Extract: func(any) string { return "" }},
}

// maxAutoColumns is the per-row column cap. Five matches
// preferredColumnOrder; the constant exists so the truncation rule is
// findable from one place.
const maxAutoColumns = 5

// autoTableTruncate is the per-cell width cap for trailing string
// columns. ~40 chars + ellipsis is enough to identify a row without
// blowing out terminal width on long URLs / filenames.
const autoTableTruncate = 40

// autoTableInteriorTruncate caps interior cells too — high enough that
// normal names fit but low enough that an absurd outlier (e.g. a
// 350-char “name“ from a test/probe workflow) doesn't push the whole
// row off the screen. Set higher than autoTableTruncate so the look
// stays "trailing column truncates first" for typical content while
// still bounding pathological cells.
const autoTableInteriorTruncate = 80

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
	if len(columns) == 0 && len(rows) == 0 {
		return renderAutoTable(os.Stdout, rows, defaultEmptyAutoColumns)
	}
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
		for i := range n {
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
	// Collect the matching alias-sets in priority order, capped at
	// maxAutoColumns. We materialize the alias slice for each pick
	// here so the closure below captures by value rather than
	// re-reading preferredColumnOrder.
	type pick struct {
		header  string
		aliases []string
	}
	var picks []pick
	timestampPicked := false
	for _, spec := range preferredColumnOrder {
		hit := false
		for _, alias := range spec.aliases {
			if aliasHasDisplayableValue(rows, alias) {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		// Only one timestamp column at a time — the order in
		// preferredColumnOrder (created_at > started_at > updated_at >
		// completed_at) defines the precedence. Showing both CREATED_AT
		// and UPDATED_AT confuses readers about which value the column
		// is reporting.
		if isTimestampHeader(spec.header) {
			if timestampPicked {
				continue
			}
			timestampPicked = true
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
		isTimestamp := isTimestampHeader(p.header)
		cols = append(cols, TableColumn{
			Header: p.header,
			Extract: func(row any) string {
				for _, alias := range aliases {
					if v, ok := rowField(row, alias); ok {
						if cellIsEmpty(v) {
							continue
						}
						if !cellIsDisplayable(v) {
							continue
						}
						s := stringifyCell(v)
						if isTimestamp {
							s = normalizeTimestampCell(s)
						}
						if isTrailing && len(s) > autoTableTruncate {
							return s[:autoTableTruncate] + "…"
						}
						if !isTrailing && len(s) > autoTableInteriorTruncate {
							return s[:autoTableInteriorTruncate] + "…"
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

// isTimestampHeader reports whether a rendered column header names a
// timestamp column whose cells should be canonicalized by
// normalizeTimestampCell. Matches the headers produced by
// preferredColumnOrder and by the explicit files-list column specs.
func isTimestampHeader(header string) bool {
	switch header {
	case "CREATED_AT", "UPDATED_AT", "STARTED_AT", "COMPLETED_AT":
		return true
	}
	return false
}

// normalizeTimestampCell renders timestamp strings in one canonical form
// for table output. SDK list responses are inconsistent: workflows
// decode created_at into a time.Time (already RFC3339 after
// stringifyCell), while files / parses / extractions leave it as the raw
// JSON string carrying microsecond precision. Untreated the same column
// shows "2026-05-15T14:21:10.389000Z" for one resource and
// "2026-05-15T14:30:29Z" for another. Parse anything that is a full
// RFC3339 timestamp and re-emit it as second-precision UTC RFC3339 so
// every list table reads the same; pass through anything that does not
// parse (ids, filenames, bare dates) untouched.
func normalizeTimestampCell(s string) string {
	if ts, err := time.Parse(time.RFC3339, s); err == nil {
		return ts.UTC().Format(time.RFC3339)
	}
	return s
}

func aliasHasDisplayableValue(rows []any, alias string) bool {
	for _, row := range rows {
		v, ok := rowField(row, alias)
		if !ok || cellIsEmpty(v) {
			continue
		}
		if cellIsDisplayable(v) {
			return true
		}
	}
	return false
}

func cellIsEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	case reflect.String:
		return rv.Len() == 0
	default:
		return false
	}
}

func cellIsDisplayable(v any) bool {
	if v == nil {
		return false
	}
	if _, ok := v.(fmt.Stringer); ok {
		return true
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return false
		}
		if rv.CanInterface() {
			if _, ok := rv.Interface().(fmt.Stringer); ok {
				return true
			}
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
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
	if strings.Contains(key, ".") {
		var current any = row
		for part := range strings.SplitSeq(key, ".") {
			next, ok := rowFieldSingle(current, part)
			if !ok {
				return nil, false
			}
			current = next
		}
		return current, true
	}
	return rowFieldSingle(row, key)
}

func rowFieldSingle(row any, key string) (any, bool) {
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
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
		v = rv.Interface()
	}
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Slice:
		if rv.IsNil() {
			return ""
		}
	}
	switch t := v.(type) {
	case string:
		return t
	case time.Time:
		// time.Time implements fmt.Stringer, but its String() form
		// ("2026-05-15 13:24:54 +0000 UTC") is inconsistent with the
		// RFC3339 timestamps jobs/files list tables render. Match those
		// by special-casing before the fmt.Stringer branch below.
		return t.UTC().Format(time.RFC3339)
	case *time.Time:
		if t == nil {
			return ""
		}
		return t.UTC().Format(time.RFC3339)
	case float32:
		if math.Trunc(float64(t)) == float64(t) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		if math.Trunc(t) == t {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case fmt.Stringer:
		return t.String()
	}
	return fmt.Sprintf("%v", v)
}

// renderAutoTable writes the header + rows to w using text/tabwriter.
// Mirrors renderTable's settings (padding=2, tab separator) so the
// column-aligned look is identical between the auto-column and
// explicit-TableColumn paths. When the row set is empty an extra
// "(no rows)" hint goes to stderr — same UX as renderTable — so a lone
// header row isn't mistaken for hidden output.
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
	if err := tw.Flush(); err != nil {
		return err
	}
	if len(rows) == 0 {
		emitEmptyRowsHint()
	}
	return nil
}
