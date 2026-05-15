package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditTemplatesCreateValidatesFormFieldsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	serverHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHits++
		t.Fatalf("server should not be reached for invalid form fields, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	dir := t.TempDir()
	documentPath := filepath.Join(dir, "template.txt")
	fieldsPath := filepath.Join(dir, "fields.json")
	if err := os.WriteFile(documentPath, []byte("Name: Alice\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fieldsPath, []byte(`[{"name":"name","label":"Name","type":"text"}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := editsTemplatesCreateCmd.Flags().Set("name", "bad-template"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsTemplatesCreateCmd.Flags().Set("name", "") })
	if err := editsTemplatesCreateCmd.Flags().Set("file", documentPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsTemplatesCreateCmd.Flags().Set("file", "") })
	if err := editsTemplatesCreateCmd.Flags().Set("form-fields-file", fieldsPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsTemplatesCreateCmd.Flags().Set("form-fields-file", "") })

	var err error
	_, stderr := captureStd(t, func() {
		err = editsTemplatesCreateCmd.RunE(editsTemplatesCreateCmd, nil)
	})
	if err == nil {
		t.Fatal("expected local form field validation error")
	}
	if !strings.Contains(stderr, `form_fields[0] missing required field "key"`) {
		t.Fatalf("stderr %q does not contain missing key error", stderr)
	}
	if serverHits != 0 {
		t.Fatalf("server was hit %d time(s), want 0", serverHits)
	}
}
