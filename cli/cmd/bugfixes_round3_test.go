//go:build !retab_oagen_cli_edits && !retab_oagen_cli_tables && !retab_oagen_cli_workflows_runs

package cmd

import (
	"strings"
	"testing"
)

// tableCell with csv=true must be faithful (no "-" placeholder that
// sanitizeCSVCell would mangle to "'-", no 96-rune truncation); csv=false
// keeps the human-readable table behavior.
func TestTableCellCSVFaithful(t *testing.T) {
	if got := tableCell(nil, true); got != "" {
		t.Errorf("tableCell(nil, csv) = %q, want empty", got)
	}
	if got := tableCell(nil, false); got != "-" {
		t.Errorf("tableCell(nil, table) = %q, want -", got)
	}
	long := strings.Repeat("x", 200)
	if got := tableCell(long, true); got != long {
		t.Errorf("tableCell(long, csv) truncated: got %d runes, want 200", len(got))
	}
	if got := tableCell(long, false); len([]rune(got)) != 96 || !strings.HasSuffix(got, "...") {
		t.Errorf("tableCell(long, table) = %q, want 96-rune ellipsized cell", got)
	}
}

// tableJSONSchemaField must not emit the "-" placeholder into CSV output when
// a column has no json_schema.
func TestTableJSONSchemaFieldCSVEmpty(t *testing.T) {
	column := map[string]any{"name": "a"}
	if got := tableJSONSchemaField(column, "type", true); got != "" {
		t.Errorf("tableJSONSchemaField(csv) = %q, want empty", got)
	}
	if got := tableJSONSchemaField(column, "type", false); got != "-" {
		t.Errorf("tableJSONSchemaField(table) = %q, want -", got)
	}
}

// sliceCells must return an empty (non-nil) slice when the requested range
// starts past the last data row, so JSON output shows "rows": [] not null.
func TestSliceCellsPastEndIsEmptyNotNil(t *testing.T) {
	rows := [][]string{{"a", "b"}}
	got := sliceCells(rows, 1, 100, 2, 200)
	if got == nil {
		t.Fatal("sliceCells past end returned nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("sliceCells past end returned %d rows, want 0", len(got))
	}
}

// appendKVPairs must trim block-id keys like the --document-id/--document-url
// parsers do, so a padded key can't dodge the cross-flag conflict checks or
// start-block alias resolution.
func TestAppendKVPairsTrimsKeys(t *testing.T) {
	into := map[string]string{}
	source := map[string]string{}
	if err := appendKVPairs(into, source, []string{" start =a.pdf"}, "--document"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := into["start"]; !ok {
		t.Fatalf("key not trimmed: got keys %v", into)
	}
	// A duplicate that differs only in padding must now collide.
	if err := appendKVPairs(into, source, []string{"start=b.pdf"}, "--document"); err == nil {
		t.Fatal("expected duplicate-key error for padded duplicate, got nil")
	}
	// A whitespace-only key must be rejected, not silently accepted.
	if err := appendKVPairs(map[string]string{}, map[string]string{}, []string{"  =c.pdf"}, "--document"); err == nil {
		t.Fatal("expected error for blank key, got nil")
	}
}

// Arrays of scalars must be displayable and comma-joined so the "actions"
// alias in preferredColumnOrder can populate the TYPE column for spec-plan
// rows ("actions": ["delete"] renders as "delete", not blank or "[delete]").
func TestScalarArrayCellsDisplayable(t *testing.T) {
	actions := []any{"add", "delete"}
	if !cellIsDisplayable(actions) {
		t.Fatal("cellIsDisplayable([]any{scalars}) = false, want true")
	}
	if got := stringifyCell(actions); got != "add,delete" {
		t.Errorf("stringifyCell = %q, want add,delete", got)
	}
	// Empty arrays and []byte stay non-displayable.
	if cellIsDisplayable([]any{}) {
		t.Error("cellIsDisplayable(empty slice) = true, want false")
	}
	if cellIsDisplayable([]byte("abc")) {
		t.Error("cellIsDisplayable([]byte) = true, want false")
	}
	// Arrays containing non-scalars stay non-displayable.
	if cellIsDisplayable([]any{map[string]any{"k": "v"}}) {
		t.Error("cellIsDisplayable([]any{map}) = true, want false")
	}
}

// edits create must reject a document plus --template-id client-side (the API
// contract is exactly-one-of; the server responds 400).
func TestEditsCreateRejectsDocumentPlusTemplate(t *testing.T) {
	flags := editsCreateCmd.Flags()
	if err := flags.Set("instructions", "fill"); err != nil {
		t.Fatal(err)
	}
	if err := flags.Set("template-id", "tmpl_x"); err != nil {
		t.Fatal(err)
	}
	if err := flags.Set("file-id", "file_x"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Reset values AND Changed state: editsCreateCmd is a package
		// global, and flags.Set marks Changed=true, which would leak into
		// later tests in the package.
		for _, name := range []string{"instructions", "template-id", "file-id"} {
			_ = flags.Set(name, "")
			flags.Lookup(name).Changed = false
		}
	})
	err := editsCreateCmd.RunE(editsCreateCmd, nil)
	if err == nil {
		t.Fatal("expected mutual-exclusion error, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("unexpected error: %v", err)
	}
}
