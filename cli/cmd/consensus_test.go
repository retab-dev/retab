//go:build !retab_oagen_cli_consensus

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// writeConsensusFile writes body to a temp file and returns its path.
func writeConsensusFile(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// resetConsensusFlags restores the shared cobra command's flags between cases;
// the command vars are package-level singletons.
func resetConsensusFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_ = consensusCreateCmd.Flags().Set("inputs", "")
		_ = consensusCreateCmd.Flags().Set("json-schema", "")
		_ = consensusCreateCmd.Flags().Set("include-alignment", "false")
	})
}

// TestConsensusCreateRejectsNonObjectInputsBeforeRequest pins that a malformed
// --inputs payload fails locally. The API would 422 anyway, but the caller's
// mistake is in their file, so the CLI should say which element is wrong
// instead of surfacing a server-side validation error.
func TestConsensusCreateRejectsNonObjectInputsBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	resetConsensusFlags(t)
	for _, tc := range []struct {
		name, body, want string
	}{
		{"not-an-array", `{"name":"Acme"}`, "must be a JSON array of objects"},
		{"empty-array", `[]`, "at least one object"},
		{"scalar-element", `[{"name":"Acme"}, 3]`, "--inputs[1] must be a JSON object"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeConsensusFile(t, "inputs.json", tc.body)
			if err := consensusCreateCmd.Flags().Set("inputs", path); err != nil {
				t.Fatal(err)
			}
			var err error
			_, stderr := captureStd(t, func() {
				err = consensusCreateCmd.RunE(consensusCreateCmd, nil)
			})
			if err == nil {
				t.Fatal("expected an inputs validation error")
			}
			if !strings.Contains(stderr, tc.want) {
				t.Fatalf("stderr %q does not mention %q", stderr, tc.want)
			}
		})
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}

// TestConsensusCreateSendsInputsSchemaAndAlignment pins the wire request: the
// inputs array, the schema under json_schema (the field that changes how
// numeric leaves reconcile, so a silent drop would quietly degrade results),
// and include_alignment only when asked for.
func TestConsensusCreateSendsInputsSchemaAndAlignment(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"consensus":{"qty":101},"likelihoods":{"qty":1},"fields":[],"alignment":null}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	inputsPath := writeConsensusFile(t, "inputs.json", `[{"qty":100},{"qty":102}]`)
	schemaPath := writeConsensusFile(t, "schema.json", `{"type":"object","properties":{"qty":{"type":"number"}}}`)

	resetConsensusFlags(t)
	for _, f := range []struct{ name, value string }{
		{"inputs", inputsPath},
		{"json-schema", schemaPath},
		{"include-alignment", "true"},
	} {
		if err := consensusCreateCmd.Flags().Set(f.name, f.value); err != nil {
			t.Fatal(err)
		}
	}

	var err error
	stdout, _ := captureStd(t, func() {
		err = consensusCreateCmd.RunE(consensusCreateCmd, nil)
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if gotPath != "/v1/consensus" {
		t.Fatalf("path = %q, want /v1/consensus", gotPath)
	}
	inputs, _ := gotBody["inputs"].([]any)
	if len(inputs) != 2 {
		t.Fatalf("inputs = %v, want 2 objects", gotBody["inputs"])
	}
	schema, ok := gotBody["json_schema"].(map[string]any)
	if !ok || schema["properties"] == nil {
		t.Fatalf("json_schema not sent intact: %v", gotBody["json_schema"])
	}
	if gotBody["include_alignment"] != true {
		t.Fatalf("include_alignment = %v, want true", gotBody["include_alignment"])
	}
	if !strings.Contains(stdout, "\"qty\"") {
		t.Fatalf("stdout %q does not contain the consensus result", stdout)
	}
}

// TestConsensusCreateOmitsIncludeAlignmentByDefault pins that the flag is not
// sent unless asked for, so the default request stays minimal.
func TestConsensusCreateOmitsIncludeAlignmentByDefault(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"consensus":{},"likelihoods":{},"fields":[],"alignment":null}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	resetConsensusFlags(t)
	if err := consensusCreateCmd.Flags().Set("inputs", writeConsensusFile(t, "inputs.json", `[{"a":1}]`)); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() {
		err = consensusCreateCmd.RunE(consensusCreateCmd, nil)
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, present := gotBody["include_alignment"]; present {
		t.Fatalf("include_alignment sent by default: %v", gotBody)
	}
	if _, present := gotBody["json_schema"]; present {
		t.Fatalf("json_schema sent without the flag: %v", gotBody)
	}
}
