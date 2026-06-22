//go:build !retab_oagen_cli_edits

package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// TestEditsCreateSendsCoercibleDocument pins the regression where
// `edits create` wrapped the resolved document in an extra pointer
// (`Document: &document`, i.e. *interface{}) before handing it to the SDK.
// The SDK's MIMEData coercion then rejected it with
// "unsupported MIME input type *interface {}", so every document-bearing
// `edits create` failed client-side before any request was dispatched.
// Every other primitive passes the resolved `doc` (an `any`) straight
// through; edits must too.
func TestEditsCreateSendsCoercibleDocument(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	var gotDocURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, _ := io.ReadAll(r.Body)
		var parsed struct {
			Document struct {
				URL string `json:"url"`
			} `json:"document"`
		}
		_ = json.Unmarshal(body, &parsed)
		gotDocURL = parsed.Document.URL
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"edt_test","object":"edit"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	const docURL = "https://example.com/contract.pdf"
	if err := editsCreateCmd.Flags().Set("url", docURL); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsCreateCmd.Flags().Set("url", "") })
	if err := editsCreateCmd.Flags().Set("instructions", "Redact phones"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsCreateCmd.Flags().Set("instructions", "") })

	_, err := captureStdAndRun(t, func() error {
		return editsCreateCmd.RunE(editsCreateCmd, nil)
	})
	if err != nil {
		t.Fatalf("edits create returned error: %v", err)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("server hit %d time(s), want 1", got)
	}
	if gotDocURL != docURL {
		t.Fatalf("request document.url = %q, want %q", gotDocURL, docURL)
	}
}

// TestEditsCreateTemplateOnlyOmitsDocument pins the template-only path:
// with `--template-id` and no document, the SDK must not reject the call
// for a missing document (document is optional in the spec — only
// `instructions` is required, and document is mutually exclusive with
// template_id). The request body must carry template_id and omit document.
func TestEditsCreateTemplateOnlyOmitsDocument(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	var hasDocument bool
	var gotTemplateID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]json.RawMessage
		_ = json.Unmarshal(body, &parsed)
		_, hasDocument = parsed["document"]
		if raw, ok := parsed["template_id"]; ok {
			_ = json.Unmarshal(raw, &gotTemplateID)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"edt_test","object":"edit"}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := editsCreateCmd.Flags().Set("template-id", "edittplt_abc"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsCreateCmd.Flags().Set("template-id", "") })
	if err := editsCreateCmd.Flags().Set("instructions", "Fill the fields"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = editsCreateCmd.Flags().Set("instructions", "") })

	_, err := captureStdAndRun(t, func() error {
		return editsCreateCmd.RunE(editsCreateCmd, nil)
	})
	if err != nil {
		t.Fatalf("edits create (template-only) returned error: %v", err)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("server hit %d time(s), want 1", got)
	}
	if hasDocument {
		t.Fatalf("request body included a document field for a template-only edit")
	}
	if gotTemplateID != "edittplt_abc" {
		t.Fatalf("request template_id = %q, want %q", gotTemplateID, "edittplt_abc")
	}
}

func TestEditTemplatesCreateValidatesFormFieldsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	serverHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHits++
		t.Fatalf("server should not be reached for invalid form fields, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

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

func TestEditTemplatesUpdateRejectsNoFieldsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached when no update fields are set, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// The `edits templates update` help promises "At least one of the two
	// flags must be set." Invoking it with neither --name nor
	// --form-fields-file must fail locally instead of issuing a no-op
	// PATCH that silently bumps updated_at.
	cmd := &cobra.Command{Use: "test-edit-template-update", RunE: editsTemplatesUpdateCmd.RunE}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("form-fields-file", "", "")

	err := cmd.RunE(cmd, []string{"tmpl_123"})
	if err == nil {
		t.Fatal("expected an error when no update fields are provided")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "at least one of --name or --form-fields-file is required") {
		t.Fatalf("error %q does not mention the required flags", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestEditTemplatesUpdateRejectsBlankNameBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Fatalf("server should not be reached for blank template name, got %s %s", r.Method, r.URL.String())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-edit-template-update", RunE: editsTemplatesUpdateCmd.RunE}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("form-fields-file", "", "")
	_ = cmd.Flags().Set("name", "   ")

	err := cmd.RunE(cmd, []string{"tmpl_123"})
	if err == nil {
		t.Fatal("expected blank template name error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "template name is required") {
		t.Fatalf("error %q does not mention blank template name", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want 0", got)
	}
}

func TestEditTemplatesUpdateReadsFormFieldsBeforeCredentials(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_API_BASE_URL", "")
	t.Setenv("HOME", t.TempDir())

	cmd := &cobra.Command{Use: "test-edit-template-update", RunE: editsTemplatesUpdateCmd.RunE}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("form-fields-file", "", "")
	_ = cmd.Flags().Set("form-fields-file", "/tmp/missing-fields.json")

	err := cmd.RunE(cmd, []string{"tmpl_123"})
	if err == nil {
		t.Fatal("expected form-fields-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--form-fields-file") {
		t.Fatalf("error %q does not mention --form-fields-file", err.Error())
	}
	if strings.Contains(err.Error(), "credentials") {
		t.Fatalf("error %q checked credentials before reading --form-fields-file", err.Error())
	}
}

func TestEditTemplatesListDoesNotExposeUnsupportedFileDateFilters(t *testing.T) {
	for _, name := range []string{"filename", "from-date", "to-date"} {
		if editsTemplatesListCmd.Flags().Lookup(name) != nil {
			t.Fatalf("edits templates list should not expose --%s", name)
		}
	}
}

// TestEditsCreateOmitsEmptyOptionalFields pins that unset optional fields are
// omitted from the request, not sent as empty strings. A non-nil *string("")
// survives omitempty, so template_id:"" on a document-only edit could trip the
// server's document-XOR-template rule, and model/color:"" could blank out
// server defaults.
func TestEditsCreateOmitsEmptyOptionalFields(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/edits" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"edit_1","status":"completed"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for n, v := range map[string]string{
		"url":          "https://example.com/contract.pdf",
		"instructions": "redact names",
	} {
		if err := editsCreateCmd.Flags().Set(n, v); err != nil {
			t.Fatalf("set --%s: %v", n, err)
		}
	}
	t.Cleanup(func() {
		for _, n := range []string{"url", "instructions", "template-id", "model", "color"} {
			_ = editsCreateCmd.Flags().Set(n, "")
			if f := editsCreateCmd.Flags().Lookup(n); f != nil {
				f.Changed = false
			}
		}
	})

	if _, err := captureStdAndRun(t, func() error {
		return editsCreateCmd.RunE(editsCreateCmd, nil)
	}); err != nil {
		t.Fatalf("edits create: %v", err)
	}
	for _, key := range []string{"template_id", "model", "config"} {
		if _, present := body[key]; present {
			t.Fatalf("%s must be omitted when its flag is unset, got %#v", key, body[key])
		}
	}
}
