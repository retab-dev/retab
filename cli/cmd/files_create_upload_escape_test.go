//go:build !retab_oagen_cli_files

package cmd

import (
	"strings"
	"testing"
)

// TestUploadResponseMarshalJSONDoesNotHTMLEscape pins that `files
// create-upload` output leaves `&`, `<`, `>` unescaped — matching every
// other CLI JSON output.
//
// Regression: uploadResponse.MarshalJSON encoded each value with the
// package-level json.Marshal, which always HTML-escapes. Signed upload URLs
// are full of `&` query separators, so the bytes MarshalJSON returned had
// the escape baked in. printJSON's SetEscapeHTML(false) only suppresses
// re-escaping at the outer encoder — it cannot undo an escape already
// present in a custom marshaler's output. The fix routes MarshalJSON
// through an encoder with HTML escaping disabled.
//
// The test drives printJSON directly (the exact path `files create-upload`
// uses) rather than json.Marshal, because json.Marshal unconditionally
// re-escapes custom-marshaler output and would not reflect real CLI output.
func TestUploadResponseMarshalJSONDoesNotHTMLEscape(t *testing.T) {
	const signedURL = "https://storage.example.com/o/file.txt?X-Goog-Algorithm=GOOG4&X-Goog-Expires=3600&X-Goog-Signature=abc"

	resp := uploadResponse{
		pairs: []uploadResponseField{
			{"id", "file_123"},
			{"upload_url", signedURL},
		},
	}

	stdout, stderr := captureStd(t, func() {
		if err := printJSON(resp); err != nil {
			t.Fatalf("printJSON: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	if strings.Contains(stdout, "\\u0026") {
		t.Fatalf("upload_url HTML-escaped (& -> \\u0026):\n%s", stdout)
	}
	if !strings.Contains(stdout, signedURL) {
		t.Fatalf("expected verbatim signed URL in output:\n%s", stdout)
	}
}
