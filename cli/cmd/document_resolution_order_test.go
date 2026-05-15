package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func writeJSONTempFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp JSON file: %v", err)
	}
	return path
}

func fakeFileLinkServer(t *testing.T, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/files/file_123/download-link" {
			t.Fatalf("path = %s, want /files/file_123/download-link", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"download_url":"https://storage.example.com/file_123.pdf","filename":"file.pdf"}`)
	}))
}

func TestClassificationValidatesCategoriesFileBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-classification", RunE: classificationsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "first-n-pages", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().StringArray("category", nil, "")
	cmd.Flags().String("categories-file", "", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("model", "retab-small")
	_ = cmd.Flags().Set("categories-file", "/does/not/exist.json")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected categories-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--categories-file") {
		t.Fatalf("error %q does not mention --categories-file", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want categories validation before file-id resolution", got)
	}
}

func TestClassificationRejectsMalformedCategoriesBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	categoriesPath := writeJSONTempFile(t, `[{"description":"missing name"}]`)

	cmd := &cobra.Command{Use: "test-classification", RunE: classificationsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "first-n-pages", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().StringArray("category", nil, "")
	cmd.Flags().String("categories-file", "", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("model", "retab-small")
	_ = cmd.Flags().Set("categories-file", categoriesPath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected malformed categories-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--categories-file[0].name is required") {
		t.Fatalf("error %q does not mention missing category name", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want categories validation before file-id resolution", got)
	}
}

func TestSplitValidatesSubdocumentsFileBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-split", RunE: splitsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("subdocuments-file", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().String("instructions", "", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("model", "retab-small")
	_ = cmd.Flags().Set("subdocuments-file", "/does/not/exist.json")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected subdocuments-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--subdocuments-file") {
		t.Fatalf("error %q does not mention --subdocuments-file", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want subdocuments validation before file-id resolution", got)
	}
}

func TestSplitRejectsMalformedSubdocumentsBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	subdocumentsPath := writeJSONTempFile(t, `[{"description":"missing name"}]`)

	cmd := &cobra.Command{Use: "test-split", RunE: splitsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("subdocuments-file", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().String("instructions", "", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("model", "retab-small")
	_ = cmd.Flags().Set("subdocuments-file", subdocumentsPath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected malformed subdocuments-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--subdocuments-file[0].name is required") {
		t.Fatalf("error %q does not mention missing subdocument name", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want subdocuments validation before file-id resolution", got)
	}
}

func TestEditTemplateCreateValidatesFormFieldsBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-edit-template", RunE: editsTemplatesCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("form-fields-file", "", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("name", "bad-template")
	_ = cmd.Flags().Set("form-fields-file", "/does/not/exist.json")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected form-fields-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--form-fields-file") {
		t.Fatalf("error %q does not mention --form-fields-file", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want form field validation before file-id resolution", got)
	}
}

func TestEditTemplateCreateRejectsBlankNameBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	fieldsPath := writeJSONTempFile(t, `[{"key":"name","description":"Name","type":"text","bbox":{"left":0,"top":0,"width":1,"height":1,"page":1}}]`)

	cmd := &cobra.Command{Use: "test-edit-template", RunE: editsTemplatesCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("form-fields-file", "", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("name", "   ")
	_ = cmd.Flags().Set("form-fields-file", fieldsPath)

	err := cmd.RunE(cmd, nil)
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
		t.Fatalf("server was hit %d time(s), want name validation before file-id resolution", got)
	}
}

func TestSchemasGenerateValidatesDocumentsFileBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-schema-generate", RunE: schemasGenerateCmd.RunE}
	cmd.Flags().StringArray("file", nil, "")
	cmd.Flags().StringArray("url", nil, "")
	cmd.Flags().StringArray("file-id", nil, "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().String("format", "schema", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("documents-file", "/does/not/exist.json")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected documents-file error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--documents-file") {
		t.Fatalf("error %q does not mention --documents-file", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want documents-file validation before file-id resolution", got)
	}
}

func TestSchemasGenerateRejectsBlankURLBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-schema-generate", RunE: schemasGenerateCmd.RunE}
	cmd.Flags().StringArray("file", nil, "")
	cmd.Flags().StringArray("url", nil, "")
	cmd.Flags().StringArray("file-id", nil, "")
	cmd.Flags().String("documents-file", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().String("format", "schema", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("url", "   ")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected blank url error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--url must not be blank") {
		t.Fatalf("error %q does not mention blank url", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want url validation before file-id resolution", got)
	}
}

func TestClassificationRejectsBlankModelBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	categoriesPath := writeJSONTempFile(t, `[{"name":"invoice","description":"invoice docs"}]`)

	cmd := &cobra.Command{Use: "test-classification", RunE: classificationsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "first-n-pages", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().StringArray("category", nil, "")
	cmd.Flags().String("categories-file", "", "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("model", "   ")
	_ = cmd.Flags().Set("categories-file", categoriesPath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected blank model error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--model must not be blank") {
		t.Fatalf("error %q does not mention blank model", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want model validation before file-id resolution", got)
	}
}

func TestPartitionRejectsBlankRequiredStringsBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := fakeFileLinkServer(t, &hits)
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-partition", RunE: partitionsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("key", "", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")

	_ = cmd.Flags().Set("file-id", "file_123")
	_ = cmd.Flags().Set("key", "   ")
	_ = cmd.Flags().Set("instructions", "split invoices")
	_ = cmd.Flags().Set("model", "retab-small")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected blank key error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--key must not be blank") {
		t.Fatalf("error %q does not mention blank key", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want key validation before file-id resolution", got)
	}
}
