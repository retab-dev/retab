package cmd

import (
	"strings"
	"testing"
)

// A --where whose column contains a rune that changes byte length under
// ToLower (U+0130) must still split at the operator correctly.
func TestParseTableWhereFlagMultibyteFoldColumn(t *testing.T) {
	where, err := parseTableWhereFlag("İd eq 5")
	if err != nil {
		t.Fatalf("parseTableWhereFlag: %v", err)
	}
	if where["column"] != "İd" {
		t.Fatalf("column = %q, want %q", where["column"], "İd")
	}
	if where["operator"] != "eq" {
		t.Fatalf("operator = %q, want eq", where["operator"])
	}
}

// Operator matching stays case-insensitive after the fold fix.
func TestParseTableWhereFlagUppercaseOperator(t *testing.T) {
	where, err := parseTableWhereFlag("amount GTE 10")
	if err != nil {
		t.Fatalf("parseTableWhereFlag: %v", err)
	}
	if where["column"] != "amount" || where["operator"] != "gte" {
		t.Fatalf("got column=%q operator=%q", where["column"], where["operator"])
	}
}

// An absurdly wide page range must be rejected instead of materialized.
func TestParsePageListRejectsHugeRange(t *testing.T) {
	_, err := parsePageList("1-2000000000")
	if err == nil {
		t.Fatal("expected error for huge page range")
	}
	if !strings.Contains(err.Error(), "more than") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Normal specs still parse after the cap was added.
func TestParsePageListNormalSpecStillWorks(t *testing.T) {
	pages, err := parsePageList("1,3,5-7")
	if err != nil {
		t.Fatalf("parsePageList: %v", err)
	}
	want := []int{1, 3, 5, 6, 7}
	if len(pages) != len(want) {
		t.Fatalf("pages = %v, want %v", pages, want)
	}
	for i := range want {
		if pages[i] != want[i] {
			t.Fatalf("pages = %v, want %v", pages, want)
		}
	}
}

func TestIndexASCIIFold(t *testing.T) {
	cases := []struct {
		s, needle string
		want      int
	}{
		{"a EQ b", " eq ", 1},
		{"a eq b", " EQ ", 1},
		{"İd eq 5", " eq ", 3},
		{"no operator here", " eq ", -1},
	}
	for _, tc := range cases {
		if got := indexASCIIFold(tc.s, tc.needle); got != tc.want {
			t.Errorf("indexASCIIFold(%q, %q) = %d, want %d", tc.s, tc.needle, got, tc.want)
		}
	}
}
