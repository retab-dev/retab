//go:build !retab_oagen_cli_classifications

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/pflag"
)

// TestClassificationsCreateOmitsUnsetBoundedIntParams pins, at the wire level,
// that unset --n-consensus / --first-n-pages are omitted from the request body
// rather than serialized as 0 (a value both flags' own ranges reject). This
// exercises the inline-RunE construction path and confirms the SDK's omitempty
// actually drops the nil *int pointers on the wire.
func TestClassificationsCreateOmitsUnsetBoundedIntParams(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/classifications" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"cls_1","status":"completed"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := classificationsCreateCmd.Flags().Set("url", "https://example.com/x.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := classificationsCreateCmd.Flags().Set("model", "gpt-4o"); err != nil {
		t.Fatal(err)
	}
	if err := classificationsCreateCmd.Flags().Set("category", "invoice=an invoice"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = classificationsCreateCmd.Flags().Set("url", "")
		_ = classificationsCreateCmd.Flags().Set("model", "")
		for _, n := range []string{"url", "model"} {
			if f := classificationsCreateCmd.Flags().Lookup(n); f != nil {
				f.Changed = false
			}
		}
		if f := classificationsCreateCmd.Flags().Lookup("category"); f != nil {
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				_ = sv.Replace(nil)
			}
			f.Changed = false
		}
	})

	if _, err := captureStdAndRun(t, func() error {
		return classificationsCreateCmd.RunE(classificationsCreateCmd, nil)
	}); err != nil {
		t.Fatalf("classifications create: %v", err)
	}
	for _, key := range []string{"n_consensus", "first_n_pages"} {
		if _, present := body[key]; present {
			t.Fatalf("%s must be omitted when its flag is unset, got %#v", key, body[key])
		}
	}
}
