package cmd

import (
	"strings"
	"testing"
)

// SDK list responses are not consistent about timestamp types: some
// decode created_at into a time.Time (workflows), others leave it as the
// raw JSON string carrying microsecond precision (files, parses,
// extractions). Rendered straight into a table that produced
// "2026-05-15T11:30:00.389000Z" for one resource and
// "2026-05-15T11:30:00Z" for another — same column, two formats.
//
// The auto-column renderer must canonicalize the timestamp column to
// second-precision UTC RFC3339 regardless of the incoming string form.
func TestAutoTableCanonicalizesTimestampColumn(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "microsecond precision", in: "2026-05-15T11:30:00.389000Z", want: "2026-05-15T11:30:00Z"},
		{name: "already canonical", in: "2026-05-15T11:30:00Z", want: "2026-05-15T11:30:00Z"},
		{name: "offset normalized to UTC", in: "2026-05-15T13:30:00+02:00", want: "2026-05-15T11:30:00Z"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows := []any{
				map[string]any{"id": "rsrc_1", "created_at": tc.in},
			}
			cols := pickAutoColumns(rows)
			var created *TableColumn
			for i := range cols {
				if cols[i].Header == "CREATED_AT" {
					created = &cols[i]
				}
			}
			if created == nil {
				t.Fatalf("no CREATED_AT column produced for rows %v", rows)
			}
			got := created.Extract(rows[0])
			if got != tc.want {
				t.Fatalf("CREATED_AT cell = %q, want canonical RFC3339 %q", got, tc.want)
			}
		})
	}
}

// Issue 2 regression: `workflows blocks list` rows carry `updated_at`
// (no `created_at`). The auto-column renderer previously labelled the
// column `CREATED_AT` because `updated_at` was a fallback alias under
// the CREATED_AT spec — that's a lie. The header must reflect the
// field actually being shown.
func TestAutoTableHeaderForUpdatedAtRowsIsUpdatedAt(t *testing.T) {
	rows := []any{
		map[string]any{
			"id":         "blk_1",
			"type":       "extract",
			"updated_at": "2026-05-21T21:33:59.557Z",
		},
	}
	cols := pickAutoColumns(rows)
	var matched *TableColumn
	headers := make([]string, 0, len(cols))
	for i := range cols {
		headers = append(headers, cols[i].Header)
		if cols[i].Header == "UPDATED_AT" {
			matched = &cols[i]
		}
	}
	for _, h := range headers {
		if h == "CREATED_AT" {
			t.Fatalf("rows carry only updated_at; auto-columns must not produce CREATED_AT header, got headers: %v", headers)
		}
	}
	if matched == nil {
		t.Fatalf("expected UPDATED_AT column for rows with updated_at only, got headers: %v", headers)
	}
	got := matched.Extract(rows[0])
	if got != "2026-05-21T21:33:59Z" {
		t.Fatalf("UPDATED_AT cell = %q, want canonical RFC3339 form", got)
	}
}

// Issue 2 regression: rows that DO carry `created_at` must still
// render as CREATED_AT. The fix must not regress the common case.
func TestAutoTableHeaderForCreatedAtRowsIsCreatedAt(t *testing.T) {
	rows := []any{
		map[string]any{
			"id":         "rsrc_1",
			"created_at": "2026-05-21T21:33:59Z",
		},
	}
	cols := pickAutoColumns(rows)
	for _, c := range cols {
		if c.Header == "UPDATED_AT" {
			t.Fatalf("rows carry created_at; auto-columns must not produce UPDATED_AT header")
		}
	}
	var matched *TableColumn
	for i := range cols {
		if cols[i].Header == "CREATED_AT" {
			matched = &cols[i]
		}
	}
	if matched == nil {
		t.Fatalf("expected CREATED_AT column for rows with created_at, got columns: %v", cols)
	}
}

// Issue 2 regression: when both fields are present on different rows,
// the column should reflect the dominant signal — `created_at` wins
// because it's the more specific identifier; the updated_at column
// would be implied as an artifact of an in-place edit.
func TestAutoTableHeaderPrefersCreatedAtOverUpdatedAt(t *testing.T) {
	rows := []any{
		map[string]any{"id": "a", "created_at": "2026-05-21T20:00:00Z"},
		map[string]any{"id": "b", "updated_at": "2026-05-21T21:00:00Z"},
	}
	cols := pickAutoColumns(rows)
	for _, c := range cols {
		if c.Header == "UPDATED_AT" {
			t.Fatalf("mixed rows with created_at present must not produce UPDATED_AT header")
		}
	}
}

// normalizeTimestampCell must leave non-timestamp strings untouched —
// IDs, filenames, model names all flow through stringifyCell and only
// genuine RFC3339 values should be rewritten.
func TestNormalizeTimestampCellPassesThroughNonTimestamps(t *testing.T) {
	for _, s := range []string{"", "file_abc123", "gpt-4.1-nano", "Invoice Template", "2026-05-15"} {
		if got := normalizeTimestampCell(s); got != s {
			t.Fatalf("normalizeTimestampCell(%q) = %q, want unchanged", s, got)
		}
	}
}

// End-to-end: a files list table (string-typed created_at) and a
// workflows list table (time.Time created_at) must render the timestamp
// column identically — second-precision RFC3339 with no fractional part.
func TestFilesListTableDropsFractionalSeconds(t *testing.T) {
	rows := []any{
		map[string]any{"id": "file_1", "filename": "inv.txt", "created_at": "2026-05-15T11:30:00.389000Z"},
	}
	var buf strings.Builder
	cols := pickAutoColumns(rows)
	if err := renderAutoTable(&buf, rows, cols); err != nil {
		t.Fatalf("renderAutoTable: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, ".389000") {
		t.Fatalf("table still carries raw fractional seconds:\n%s", out)
	}
	if !strings.Contains(out, "2026-05-15T11:30:00Z") {
		t.Fatalf("table missing canonical timestamp:\n%s", out)
	}
}
