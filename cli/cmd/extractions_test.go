//go:build !retab_oagen_cli_extractions

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
	cmd.Flags().String("image-resolution-dpi", "", "")
	cmd.Flags().Var(&nonNegativeIntFlagValue{}, "n-consensus", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().StringArray("metadata", nil, "")
	cmd.Flags().String("messages-file", "", "")
	return cmd
}

// TestNewExtractionRequestGatesConsensusParam pins that unset --n-consensus is
// OMITTED from the request, not sent as 0. The legacy --image-resolution-dpi
// flag remains accepted for CLI compatibility but is no longer present in the
// generated request type.
func TestNewExtractionRequestGatesConsensusParam(t *testing.T) {
	t.Run("omitted when unset", func(t *testing.T) {
		cmd := newExtractionRequestTestCmd(t)
		for n, v := range map[string]string{
			"url":         "https://example.com/x.pdf",
			"model":       "gpt-4o",
			"json-schema": `{"type":"object"}`,
		} {
			if err := cmd.Flags().Set(n, v); err != nil {
				t.Fatalf("set --%s: %v", n, err)
			}
		}
		params, err := newExtractionRequest(cmd)
		if err != nil {
			t.Fatalf("newExtractionRequest: %v", err)
		}
		if params.NConsensus != nil {
			t.Fatalf("NConsensus must be nil when --n-consensus unset, got %d", *params.NConsensus)
		}
	})

	t.Run("consensus sent and legacy dpi ignored when set", func(t *testing.T) {
		cmd := newExtractionRequestTestCmd(t)
		for n, v := range map[string]string{
			"url":                  "https://example.com/x.pdf",
			"model":                "gpt-4o",
			"json-schema":          `{"type":"object"}`,
			"n-consensus":          "3",
			"image-resolution-dpi": "150",
		} {
			if err := cmd.Flags().Set(n, v); err != nil {
				t.Fatalf("set --%s: %v", n, err)
			}
		}
		params, err := newExtractionRequest(cmd)
		if err != nil {
			t.Fatalf("newExtractionRequest: %v", err)
		}
		if params.NConsensus == nil || *params.NConsensus != 3 {
			t.Fatalf("NConsensus = %v, want 3", params.NConsensus)
		}
	})
}

func TestNewExtractionRequestValidatesMetadataBeforeResolvingFileID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
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
