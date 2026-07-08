package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// replaceCSVHeaderColumns must read only the header row, auto-detect the
// delimiter, strip a UTF-8 BOM and surrounding quotes, and return nothing for
// stdin so `tables replace --preserve-schema` can restrict overrides safely.
func TestReplaceCSVHeaderColumns(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	cases := []struct {
		name string
		path string
		want []string
	}{
		{"comma", write("c.csv", "customer_name,code\nAGCO,0000511029\n"), []string{"customer_name", "code"}},
		{"semicolon", write("s.csv", "customer_name;code\nAGCO;0000511029\n"), []string{"customer_name", "code"}},
		{"bom+quotes", write("b.csv", "\ufeff\"customer_name\",\"code\"\nAGCO,1\n"), []string{"customer_name", "code"}},
		{"crlf", write("r.csv", "a,b\r\n1,2\r\n"), []string{"a", "b"}},
		{"single-header-no-newline", write("h.csv", "only"), []string{"only"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := replaceCSVHeaderColumns(tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, "|") != strings.Join(tc.want, "|") {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}

	// stdin and unset path must return no columns (header can't be peeked
	// without consuming the upload) rather than erroring.
	for _, p := range []string{"-", ""} {
		got, err := replaceCSVHeaderColumns(p)
		if err != nil || got != nil {
			t.Errorf("replaceCSVHeaderColumns(%q) = (%v, %v), want (nil, nil)", p, got, err)
		}
	}
}

// filesGetFlagErrorHint must redirect the confusing bare "-o" flag error to
// `files download`, while leaving unrelated flag errors untouched.
func TestFilesGetFlagErrorHint(t *testing.T) {
	oErr := errors.New("unknown shorthand flag: 'o' in -o")
	got := filesGetFlagErrorHint(oErr)
	if !strings.Contains(got.Error(), "files download") {
		t.Errorf("expected hint to mention `files download`, got %q", got.Error())
	}
	if !errors.Is(got, oErr) {
		t.Errorf("augmented error must wrap the original (errors.Is)")
	}

	other := errors.New("unknown flag: --nope")
	if got := filesGetFlagErrorHint(other); got != other {
		t.Errorf("unrelated flag error must pass through unchanged, got %q", got.Error())
	}
	if filesGetFlagErrorHint(nil) != nil {
		t.Errorf("nil must pass through as nil")
	}
}

// printDraftPublishHint must name the publish command with the workflow id on
// stderr, and stay silent for an empty id.
func TestPrintDraftPublishHint(t *testing.T) {
	var buf strings.Builder
	cmd := &cobra.Command{}
	cmd.SetErr(&buf)
	printDraftPublishHint(cmd, "wf_abc123")
	out := buf.String()
	if !strings.Contains(out, "retab workflows publish wf_abc123") {
		t.Errorf("hint should name the publish command with the id, got %q", out)
	}

	buf.Reset()
	printDraftPublishHint(cmd, "")
	if buf.String() != "" {
		t.Errorf("empty workflow id must produce no hint, got %q", buf.String())
	}
}
