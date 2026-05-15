package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

func TestWriteGeneratedSchemaDefaultEmitsReusableSchema(t *testing.T) {
	response := retab.Resource{
		"object": "schema",
		"json_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"invoice_number": map[string]any{"type": "string"},
			},
		},
	}

	var buf bytes.Buffer
	if err := writeGeneratedSchema(&buf, &response, "schema"); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
	}
	if got["type"] != "object" {
		t.Fatalf("type = %#v, want object", got["type"])
	}
	if _, ok := got["json_schema"]; ok {
		t.Fatalf("default output should be the reusable schema body, got envelope:\n%s", buf.String())
	}
}

func TestWriteGeneratedSchemaJSONFormatPreservesEnvelope(t *testing.T) {
	response := retab.Resource{
		"object":      "schema",
		"created_at":  "2026-05-15T11:42:06Z",
		"json_schema": map[string]any{"type": "object"},
	}

	var buf bytes.Buffer
	if err := writeGeneratedSchema(&buf, &response, "json"); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
	}
	if got["object"] != "schema" {
		t.Fatalf("object = %#v, want schema", got["object"])
	}
	if _, ok := got["json_schema"]; !ok {
		t.Fatalf("json format should preserve envelope:\n%s", buf.String())
	}
}

func TestSchemasGenerateRejectsUnknownFormatBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"json_schema": map[string]any{"type": "object"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	schemasGenerateCmd.SetContext(context.Background())
	t.Cleanup(func() { schemasGenerateCmd.SetContext(nil) })
	if err := schemasGenerateCmd.Flags().Set("url", "https://example.com/sample.pdf"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if slice, ok := schemasGenerateCmd.Flags().Lookup("url").Value.(interface{ Replace([]string) error }); ok {
			_ = slice.Replace(nil)
		}
	})
	if err := schemasGenerateCmd.Flags().Set("format", "yaml"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = schemasGenerateCmd.Flags().Set("format", "schema") })

	var err error
	_, stderr := captureStd(t, func() {
		err = schemasGenerateCmd.RunE(schemasGenerateCmd, nil)
	})
	if err == nil {
		t.Fatalf("expected invalid format error")
	}
	if !strings.Contains(stderr, "invalid --format") {
		t.Fatalf("stderr %q does not mention invalid --format", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}
