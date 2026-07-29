//go:build !retab_oagen_cli_extractions

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"

	retab "github.com/retab-dev/retab/clients/go"
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
	cmd.Flags().Bool("deep-extraction", false, "")
	return cmd
}

// TestNewExtractionRequestGatesDeepExtractionParam pins that --deep-extraction
// is OMITTED when unset rather than sent as false. An unconditional false would
// put deep_extraction on every request body the CLI writes, which is both a
// wire-shape change for every caller and a needless divergence from the
// server's "absent == default" contract. An EXPLICIT --deep-extraction=false
// must still be sent, so a user can override a wrapper that sets it.
func TestNewExtractionRequestGatesDeepExtractionParam(t *testing.T) {
	baseFlags := map[string]string{
		"url":         "https://example.com/long.pdf",
		"model":       "retab-large",
		"json-schema": `{"type":"object"}`,
	}
	build := func(t *testing.T, extra map[string]string) (retab.ExtractionsCreateParams, error) {
		t.Helper()
		cmd := newExtractionRequestTestCmd(t)
		for n, v := range baseFlags {
			if err := cmd.Flags().Set(n, v); err != nil {
				t.Fatalf("set --%s: %v", n, err)
			}
		}
		for n, v := range extra {
			if err := cmd.Flags().Set(n, v); err != nil {
				t.Fatalf("set --%s: %v", n, err)
			}
		}
		return newExtractionRequest(cmd)
	}

	t.Run("omitted when unset", func(t *testing.T) {
		params, err := build(t, nil)
		if err != nil {
			t.Fatalf("newExtractionRequest: %v", err)
		}
		if params.DeepExtraction != nil {
			t.Fatalf("DeepExtraction must be nil when the flag is unset, got %v", *params.DeepExtraction)
		}
	})

	t.Run("sent when set", func(t *testing.T) {
		params, err := build(t, map[string]string{"deep-extraction": "true"})
		if err != nil {
			t.Fatalf("newExtractionRequest: %v", err)
		}
		if params.DeepExtraction == nil || !*params.DeepExtraction {
			t.Fatalf("DeepExtraction = %v, want true", params.DeepExtraction)
		}
	})

	t.Run("explicit false is sent", func(t *testing.T) {
		params, err := build(t, map[string]string{"deep-extraction": "false"})
		if err != nil {
			t.Fatalf("newExtractionRequest: %v", err)
		}
		if params.DeepExtraction == nil || *params.DeepExtraction {
			t.Fatalf("DeepExtraction = %v, want an explicit false", params.DeepExtraction)
		}
	})
}

// TestExtractionCommandsDropRetiredWindowingFlags pins that the retired
// spreadsheet-windowing knobs are gone from the real extraction commands.
//
// excel_windowing / auto_chunk_rows left the public contract (they are absent
// from public/docs/api-reference/openapi.json and from every generated client),
// so ExtractionsCreateParams has no field to carry them. Re-adding the flags
// without first restoring the request fields would either not compile or,
// worse, accept --excel-windowing=auto and silently return a whole-document
// extraction instead of the {"data": [...]} shape the caller asked for.
func TestExtractionCommandsDropRetiredWindowingFlags(t *testing.T) {
	for _, cmd := range []*cobra.Command{extractionsCreateCmd, extractionsStreamCmd} {
		for _, name := range []string{"excel-windowing", "auto-chunk-rows"} {
			if f := cmd.Flags().Lookup(name); f != nil {
				t.Errorf("%s: --%s is retired from the public contract and must not be registered", cmd.Name(), name)
			}
		}
	}
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

// TestExtractionsCreateCommandSendsDeepExtractionOnTheWire is the CLI's
// end-to-end guard. TestNewExtractionRequestGatesDeepExtractionParam covers the
// flag→params mapping; this runs the actual command against a stub server and
// asserts the field reaches the HTTP body. The two halves catch different
// breaks: params mapping catches a mis-wired flag, this catches a params field
// the SDK never serializes.
func TestExtractionsCreateCommandSendsDeepExtractionOnTheWire(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"extr_1","file":{"id":"file_1","filename":"ledger.pdf","mime_type":"application/pdf"},"model":"retab-large","json_schema":{"type":"object"},"output":{}}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := newExtractionRequestTestCmd(t)
	for name, value := range map[string]string{
		"url":             "https://example.com/ledger.pdf",
		"model":           "retab-large",
		"json-schema":     `{"type":"object"}`,
		"deep-extraction": "true",
	} {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set --%s: %v", name, err)
		}
	}

	params, err := newExtractionRequest(cmd)
	if err != nil {
		t.Fatalf("newExtractionRequest: %v", err)
	}
	client, err := newClient(cmd)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	if _, err := client.Extractions.Create(context.Background(), &params); err != nil {
		t.Fatalf("Extractions.Create: %v", err)
	}

	if body == nil {
		t.Fatal("no request body captured")
	}
	if body["deep_extraction"] != true {
		t.Fatalf("wire body deep_extraction = %#v, want true", body["deep_extraction"])
	}
}
