package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseDocumentArgs_DocumentFlagOnly_NoWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs([]string{"start=./a.pdf"}, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./a.pdf" {
		t.Fatalf("got %v, want {start: ./a.pdf}", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestParseDocumentArgs_MultipleDocumentFlags(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./a.pdf", "classify=./b.pdf"},
		nil,
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want both keys", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestParseDocumentArgs_LegacyFlagEmitsOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(nil, []string{"start=./a.pdf"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./a.pdf" {
		t.Fatalf("got %v, want {start: ./a.pdf}", got)
	}
	if !strings.Contains(warn.String(), "--document-file is deprecated") {
		t.Fatalf("expected deprecation warning, got %q", warn.String())
	}
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_LegacyMultipleEntriesOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		nil,
		[]string{"start=./a.pdf", "classify=./b.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want both keys", got)
	}
	// Two legacy entries must still produce exactly one warning line.
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_MixedFlagsUnionOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./a.pdf"},
		[]string{"classify=./b.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want union of both keys", got)
	}
	if !strings.Contains(warn.String(), "--document-file is deprecated") {
		t.Fatalf("expected deprecation warning, got %q", warn.String())
	}
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_NewFlagWinsOnCollision(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./new.pdf"},
		[]string{"start=./legacy.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./new.pdf" {
		t.Fatalf("got %v, want {start: ./new.pdf} (--document overrides --document-file)", got)
	}
	// Still exactly one warning line because the legacy flag was used.
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_NoFlagsEmptyMap(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(nil, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty map", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestParseDocumentArgs_BadShapes(t *testing.T) {
	cases := []struct {
		name string
		docs []string
		legs []string
	}{
		{name: "missing equals on --document", docs: []string{"./a.pdf"}},
		{name: "empty key on --document", docs: []string{"=./a.pdf"}},
		{name: "empty value on --document", docs: []string{"start="}},
		{name: "missing equals on --document-file", legs: []string{"./a.pdf"}},
		{name: "empty key on --document-file", legs: []string{"=./a.pdf"}},
		{name: "empty value on --document-file", legs: []string{"start="}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var warn bytes.Buffer
			_, err := parseDocumentArgs(tc.docs, tc.legs, &warn)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestParseDocumentArgs_NilWarnSinkDoesNotPanic(t *testing.T) {
	// Smoke test: when the legacy flag is used but warnTo is nil (e.g. tests
	// that don't care about warnings), the helper must not panic.
	_, err := parseDocumentArgs(nil, []string{"start=./a.pdf"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
