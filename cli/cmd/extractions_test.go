//go:build !retab_oagen_cli_extractions

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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

// TestExtractionsStreamCommandSendsDeepExtractionOnTheWire is the stream twin
// of TestExtractionsCreateCommandSendsDeepExtractionOnTheWire, and the guard
// the stream path did not have. `stream` hand-assembles its request body
// (the generated SDK's CreateStream discards the response), so it does not
// inherit `create`'s serialization — and it was omitting deep_extraction while
// addExtractionBodyFlags still registered --deep-extraction on it. The flag was
// accepted and silently discarded: the stream ran an ordinary single-shot
// extraction. The server honors deep_extraction on /v1/extractions/stream (it
// has a deep-spreadsheet-specific 400 there), so nothing but this body was
// missing. Verified against staging: before the fix the flag never left the
// client; after it, streaming a CSV with --deep-extraction returns the server's
// "deep_extraction ... does not support stream" 400.
func TestExtractionsStreamCommandSendsDeepExtractionOnTheWire(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/stream+json")
		_, _ = fmt.Fprint(w, "{\"id\":\"chatcmpl_1\",\"extraction_id\":\"extr_1\"}\n")
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := extractionsStreamCmd
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
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
	t.Cleanup(func() {
		for _, name := range []string{"url", "model", "json-schema", "deep-extraction"} {
			_ = cmd.Flags().Set(name, cmd.Flags().Lookup(name).DefValue)
			cmd.Flags().Lookup(name).Changed = false
		}
	})

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("stream RunE: %v", err)
	}
	if path != "/v1/extractions/stream" {
		t.Fatalf("request path = %q, want /v1/extractions/stream", path)
	}
	if body == nil {
		t.Fatal("no request body captured")
	}
	if body["deep_extraction"] != true {
		t.Fatalf("stream body deep_extraction = %#v, want true", body["deep_extraction"])
	}
	if body["stream"] != true {
		t.Fatalf("stream body stream = %#v, want true", body["stream"])
	}
}

// streamBodyExclusions are the create-request fields that deliberately do NOT
// belong in a stream body, each with the reason. Everything else a caller can
// set MUST reach the stream, because `retab extractions stream --help` says
// "Flags and document/schema resolution are identical to
// retab extractions create".
var streamBodyExclusions = map[string]string{
	"background": "background and stream are mutually exclusive execution modes: one returns a queued record to poll, the other holds the connection open",
}

// TestStreamBodyCarriesEveryCreateField is the structural guard for the whole
// class of bug that --deep-extraction hit. `stream` hand-assembles its request
// body field by field instead of serializing the params struct, so any field
// added to (or already on) the create request is silently dropped from stream
// until someone remembers to add a line — the flag stays registered, the CLI
// accepts it, and the request runs without it. Enumerating the params struct
// and requiring an explicit exclusion reason turns that silent omission into a
// failing test.
func TestStreamBodyCarriesEveryCreateField(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/stream+json")
		_, _ = fmt.Fprint(w, "{\"id\":\"chatcmpl_1\"}\n")
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := extractionsStreamCmd
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	// Set every flag the stream command exposes, so no field is absent merely
	// because the test did not ask for it.
	flags := map[string]string{
		"url":             "https://example.com/ledger.pdf",
		"model":           "retab-large",
		"json-schema":     `{"type":"object"}`,
		"deep-extraction": "true",
		"instructions":    "steer me",
		"n-consensus":     "3",
		"bust-cache":      "true",
	}
	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set --%s: %v", name, err)
		}
	}
	if err := cmd.Flags().Set("metadata", "customer=acme"); err != nil {
		t.Fatalf("set --metadata: %v", err)
	}
	messagesFile := filepath.Join(t.TempDir(), "messages.json")
	if err := os.WriteFile(messagesFile, []byte(`[{"role":"user","content":"context"}]`), 0o600); err != nil {
		t.Fatalf("write messages file: %v", err)
	}
	if err := cmd.Flags().Set("messages-file", messagesFile); err != nil {
		t.Fatalf("set --messages-file: %v", err)
	}
	t.Cleanup(func() {
		for name := range flags {
			_ = cmd.Flags().Set(name, cmd.Flags().Lookup(name).DefValue)
			cmd.Flags().Lookup(name).Changed = false
		}
		if slice, ok := cmd.Flags().Lookup("metadata").Value.(pflag.SliceValue); ok {
			_ = slice.Replace(nil)
		}
		cmd.Flags().Lookup("metadata").Changed = false
		_ = cmd.Flags().Set("messages-file", "")
		cmd.Flags().Lookup("messages-file").Changed = false
	})

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("stream RunE: %v", err)
	}
	if body == nil {
		t.Fatal("no request body captured")
	}

	// The wire names the create params serialize to. Reflection over the params
	// struct is what makes this list self-maintaining: a new field appears here
	// automatically and must then be sent or excluded with a reason.
	params := reflect.TypeOf(retab.ExtractionsCreateParams{})
	var missing []string
	for i := 0; i < params.NumField(); i++ {
		field := params.Field(i)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			continue
		}
		if reason, excluded := streamBodyExclusions[name]; excluded {
			if _, present := body[name]; present {
				t.Errorf("%q is in the stream body but is documented as excluded: %s", name, reason)
			}
			continue
		}
		if _, present := body[name]; !present {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("the hand-assembled stream body drops create fields %v.\n"+
			"Either send them in extractionsStreamCmd's body, or add each to streamBodyExclusions with the reason it does not apply to streaming.\n"+
			"body was: %v", missing, sortedKeys(body))
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// --deep-extraction has three states on BOTH create and stream, and they must
// agree: unset omits the field entirely (the server's "absent == default"
// contract, and not a wire-shape change for every caller), --deep-extraction
// sends true, and an explicit --deep-extraction=false sends false so a user can
// override a wrapper that turned it on.
func TestDeepExtractionThreeStatesMatchOnCreateAndStream(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body = nil
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"extr_1","file":{"id":"file_1","filename":"a.pdf","mime_type":"application/pdf"},"model":"retab-large","json_schema":{"type":"object"},"output":{}}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cases := map[string]struct {
		set  bool
		flag string
		want any
	}{
		"unset":          {set: false, want: nil},
		"explicit true":  {set: true, flag: "true", want: true},
		"explicit false": {set: true, flag: "false", want: false},
	}

	for name, tc := range cases {
		// create
		createCmd := newExtractionRequestTestCmd(t)
		for flagName, value := range map[string]string{
			"url":         "https://example.com/a.pdf",
			"model":       "retab-large",
			"json-schema": `{"type":"object"}`,
		} {
			if err := createCmd.Flags().Set(flagName, value); err != nil {
				t.Fatalf("%s: set --%s: %v", name, flagName, err)
			}
		}
		if tc.set {
			if err := createCmd.Flags().Set("deep-extraction", tc.flag); err != nil {
				t.Fatalf("%s: set --deep-extraction: %v", name, err)
			}
		}
		createParams, err := newExtractionRequest(createCmd)
		if err != nil {
			t.Fatalf("%s: newExtractionRequest: %v", name, err)
		}
		client, err := newClient(createCmd)
		if err != nil {
			t.Fatalf("%s: client: %v", name, err)
		}
		if _, err := client.Extractions.Create(context.Background(), &createParams); err != nil {
			t.Fatalf("%s: Extractions.Create: %v", name, err)
		}
		createValue, createPresent := body["deep_extraction"]
		if tc.want == nil && createPresent {
			t.Errorf("%s: create body carries deep_extraction=%#v, want the field omitted", name, createValue)
		}
		if tc.want != nil && createValue != tc.want {
			t.Errorf("%s: create body deep_extraction = %#v, want %#v", name, createValue, tc.want)
		}

		// stream — same three states, same answers
		streamCmd := extractionsStreamCmd
		streamCmd.SetOut(&strings.Builder{})
		streamCmd.SetErr(&strings.Builder{})
		for flagName, value := range map[string]string{
			"url":         "https://example.com/a.pdf",
			"model":       "retab-large",
			"json-schema": `{"type":"object"}`,
		} {
			if err := streamCmd.Flags().Set(flagName, value); err != nil {
				t.Fatalf("%s: stream set --%s: %v", name, flagName, err)
			}
		}
		if tc.set {
			if err := streamCmd.Flags().Set("deep-extraction", tc.flag); err != nil {
				t.Fatalf("%s: stream set --deep-extraction: %v", name, err)
			}
		}
		if err := streamCmd.RunE(streamCmd, nil); err != nil {
			t.Fatalf("%s: stream RunE: %v", name, err)
		}
		streamValue, streamPresent := body["deep_extraction"]
		if tc.want == nil && streamPresent {
			t.Errorf("%s: stream body carries deep_extraction=%#v, want the field omitted", name, streamValue)
		}
		if tc.want != nil && streamValue != tc.want {
			t.Errorf("%s: stream body deep_extraction = %#v, want %#v", name, streamValue, tc.want)
		}

		for _, flagName := range []string{"url", "model", "json-schema", "deep-extraction"} {
			_ = streamCmd.Flags().Set(flagName, streamCmd.Flags().Lookup(flagName).DefValue)
			streamCmd.Flags().Lookup(flagName).Changed = false
		}
	}
}
