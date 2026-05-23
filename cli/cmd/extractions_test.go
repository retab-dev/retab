package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func newExtractionRequestTestCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test-extraction"}
	addDocumentFlags(cmd)
	addSchemaFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&boundedIntFlagValue{min: 96, max: 300}, "image-resolution-dpi", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().StringArray("metadata", nil, "")
	cmd.Flags().String("messages-file", "", "")
	return cmd
}

func TestNewExtractionRequestValidatesMetadataBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/v1/files/file_123/download-link" {
			t.Fatalf("path = %s, want /v1/files/file_123/download-link", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"download_url":"https://storage.example.com/file_123.pdf","filename":"file.pdf"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newExtractionRequestTestCmd(t)
	if err := cmd.Flags().Set("file-id", "file_123"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("json-schema", `{"type":"object"}`); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("metadata", "bad"); err != nil {
		t.Fatal(err)
	}

	_, err := newExtractionRequest(cmd)
	if err == nil {
		t.Fatal("expected invalid metadata error")
	}
	if !strings.Contains(err.Error(), "invalid key=value") {
		t.Fatalf("error %q does not mention invalid metadata", err.Error())
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want metadata validation before file-id resolution", got)
	}
}
