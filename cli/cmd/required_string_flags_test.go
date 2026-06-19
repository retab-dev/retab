//go:build !retab_oagen_cli_edits && !retab_oagen_cli_files && !retab_oagen_cli_parses

package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func TestParsesCreateRejectsBlankModelBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-parse", RunE: parsesCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().String("table-parsing-format", "", "")
	cmd.Flags().String("image-resolution-dpi", "", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("bust-cache", false, "")

	_ = cmd.Flags().Set("url", "https://example.com/doc.pdf")
	_ = cmd.Flags().Set("model", "   ")

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
		t.Fatalf("server was hit %d time(s), want model validation before request", got)
	}
}

func TestFilesCreateUploadRejectsBlankRequiredStringsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-file-upload", RunE: filesCreateUploadCmd.RunE}
	cmd.Flags().String("filename", "", "")
	cmd.Flags().String("content-type", "", "")
	cmd.Flags().Var(&nonNegativeInt64FlagValue{}, "size-bytes", "")
	cmd.Flags().Var(&sha256FlagValue{}, "sha256", "")

	_ = cmd.Flags().Set("filename", "   ")
	_ = cmd.Flags().Set("content-type", "application/pdf")
	_ = cmd.Flags().Set("size-bytes", "12")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected blank filename error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--filename must not be blank") {
		t.Fatalf("error %q does not mention blank filename", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want filename validation before request", got)
	}
}

func TestEditsCreateRejectsBlankInstructionsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := &cobra.Command{Use: "test-edit", RunE: editsCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().String("template-id", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().String("color", "", "")
	cmd.Flags().Bool("bust-cache", false, "")

	_ = cmd.Flags().Set("url", "https://example.com/doc.pdf")
	_ = cmd.Flags().Set("instructions", "   ")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected blank instructions error")
	}
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	if !strings.Contains(err.Error(), "--instructions must not be blank") {
		t.Fatalf("error %q does not mention blank instructions", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want instructions validation before request", got)
	}
}
