//go:build !retab_oagen_cli_classifications && !retab_oagen_cli_edits && !retab_oagen_cli_extractions && !retab_oagen_cli_parses && !retab_oagen_cli_partitions && !retab_oagen_cli_schemas && !retab_oagen_cli_splits

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// captureClassifyRequestBody runs the classifications create command against a
// fake server with the given flag setters and returns the decoded request body.
func captureClassifyRequestBody(t *testing.T, set func(*cobra.Command)) map[string]any {
	t.Helper()
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	docPath := filepath.Join(t.TempDir(), "doc.pdf")
	if err := os.WriteFile(docPath, []byte("%PDF-1.4 tiny"), 0o600); err != nil {
		t.Fatalf("write doc: %v", err)
	}

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/classifications" {
			t.Fatalf("path = %s, want /v1/classifications", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode request body: %v (raw=%s)", err, raw)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"clss_1","status":"completed","output":{"category":"invoice","reasoning":"x"}}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-classification", RunE: classificationsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "first-n-pages", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().StringArray("category", nil, "")
	cmd.Flags().String("categories-file", "", "")
	addPrimitiveBackgroundFlag(cmd)

	_ = cmd.Flags().Set("file", docPath)
	_ = cmd.Flags().Set("model", "retab-small")
	_ = cmd.Flags().Set("category", "invoice=vendor bills")
	set(cmd)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if body == nil {
		t.Fatal("server never received the classification request")
	}
	return body
}

// TestClassifyOmitsFirstNPagesWhenUnset guards the bug where an unset
// --first-n-pages was serialized as first_n_pages=0, which made the backend
// render a zero-page document entry so the classifier received no document.
func TestClassifyOmitsFirstNPagesWhenUnset(t *testing.T) {
	body := captureClassifyRequestBody(t, func(*cobra.Command) {})
	if _, present := body["first_n_pages"]; present {
		t.Fatalf("first_n_pages must be omitted when --first-n-pages is unset, got %v", body["first_n_pages"])
	}
}

// TestClassifyForwardsFirstNPagesWhenSet ensures an explicitly set value is sent.
func TestClassifyForwardsFirstNPagesWhenSet(t *testing.T) {
	body := captureClassifyRequestBody(t, func(cmd *cobra.Command) {
		_ = cmd.Flags().Set("first-n-pages", "2")
	})
	v, present := body["first_n_pages"]
	if !present {
		t.Fatal("first_n_pages must be forwarded when --first-n-pages is set")
	}
	if fmt.Sprintf("%v", v) != "2" {
		t.Fatalf("first_n_pages = %v, want 2", v)
	}
}
