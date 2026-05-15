package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestParseDocumentArgs_DocumentFlagOnly_NoWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs([]string{"start=./a.pdf"}, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./a.pdf" {
		t.Fatalf("got %v, want {start: ./a.pdf}", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestWorkflowsRunsCreateResolvesStartAliasToGeneratedStartBlock(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var postedDocuments map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/blocks":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "block_generated", "type": "start", "label": "Document"},
					{"id": "parse", "type": "parse", "label": "Parse"},
				},
				"list_metadata": map[string]any{},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/run":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			postedDocuments, _ = body["documents"].(map[string]any)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "run_123",
				"status": "running",
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	dir := t.TempDir()
	docPath := filepath.Join(dir, "invoice.txt")
	if err := os.WriteFile(docPath, []byte("invoice\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := workflowsRunsCreateCmd.Flags().Set("document", "start="+docPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsRunsCreateCmd.Flags().Set("document", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsRunsCreateCmd.RunE(workflowsRunsCreateCmd, []string{"wf_123"}); err != nil {
			t.Fatalf("runs create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_123") {
		t.Fatalf("expected run response on stdout, got:\n%s", stdout)
	}
	if _, ok := postedDocuments["block_generated"]; !ok {
		t.Fatalf("documents posted under keys %#v, want block_generated", keysOfAnyMap(postedDocuments))
	}
	if _, ok := postedDocuments["start"]; ok {
		t.Fatalf("friendly alias leaked into request body: %#v", postedDocuments)
	}
}

func keysOfAnyMap(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func TestWorkflowsRunsListRejectsInvalidListFlagsLocally(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		value     string
		wantError string
		reset     string
	}{
		{name: "negative limit", flag: "limit", value: "-1", wantError: "non-negative", reset: "0"},
		{name: "invalid order", flag: "order", value: "sideways", wantError: "asc", reset: ""},
		{name: "invalid from date", flag: "from-date", value: "not-a-date", wantError: "YYYY-MM-DD", reset: ""},
		{name: "invalid to date", flag: "to-date", value: "not-a-date", wantError: "YYYY-MM-DD", reset: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := workflowsRunsListCmd.Flags().Set(tc.flag, tc.value)
			if err == nil {
				t.Fatalf("expected local parse error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
			}
			if resetErr := workflowsRunsListCmd.Flags().Set(tc.flag, tc.reset); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestWorkflowsRunsCommandsRejectInvalidEnumFiltersBeforeRequest(t *testing.T) {
	cases := []struct {
		name      string
		cmd       *cobra.Command
		args      []string
		flag      string
		value     string
		wantError string
	}{
		{name: "list invalid status", cmd: workflowsRunsListCmd, flag: "status", value: "banana", wantError: "invalid --status"},
		{name: "list invalid statuses", cmd: workflowsRunsListCmd, flag: "statuses", value: "running,banana", wantError: "invalid --statuses"},
		{name: "list invalid exclude status", cmd: workflowsRunsListCmd, flag: "exclude-status", value: "banana", wantError: "invalid --exclude-status"},
		{name: "list invalid trigger type", cmd: workflowsRunsListCmd, flag: "trigger-type", value: "banana", wantError: "invalid --trigger-type"},
		{name: "list invalid trigger types", cmd: workflowsRunsListCmd, flag: "trigger-types", value: "api,banana", wantError: "invalid --trigger-types"},
		{name: "export invalid status", cmd: workflowsRunsExportCmd, flag: "status", value: "banana", wantError: "invalid --status"},
		{name: "export invalid exclude status", cmd: workflowsRunsExportCmd, flag: "exclude-status", value: "banana", wantError: "invalid --exclude-status"},
		{name: "export invalid trigger types", cmd: workflowsRunsExportCmd, flag: "trigger-types", value: "api,banana", wantError: "invalid --trigger-types"},
		{name: "export invalid export source", cmd: workflowsRunsExportCmd, flag: "export-source", value: "banana", wantError: "invalid --export-source"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RETAB_API_KEY", "test-key")
			t.Setenv("HOME", t.TempDir())

			hits := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				t.Fatalf("server should not be reached for invalid local filter, got %s %s", r.Method, r.URL.String())
			}))
			defer server.Close()
			t.Setenv("RETAB_BASE_URL", server.URL)

			if err := tc.cmd.Flags().Set(tc.flag, tc.value); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { resetWorkflowRunsFlag(t, tc.cmd, tc.flag) })

			var err error
			_, stderr := captureStd(t, func() {
				err = tc.cmd.RunE(tc.cmd, tc.args)
			})
			if err == nil {
				t.Fatalf("expected local validation error for --%s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(stderr, tc.wantError) {
				t.Fatalf("stderr %q does not contain %q", stderr, tc.wantError)
			}
			if hits != 0 {
				t.Fatalf("server was hit %d time(s), want 0", hits)
			}
		})
	}
}

func TestWorkflowsRunsRestartSendsDefaultConfigSource(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs/run_123/restart" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "run_456",
			"workflow_id": "wf_123",
			"workflow": map[string]any{
				"workflow_id":       "wf_123",
				"version_id":        "ver_123",
				"name_at_run_time":  "Workflow",
				"requested_version": "production",
			},
			"trigger": map[string]any{"type": "api"},
			"lifecycle": map[string]any{
				"status": "running",
			},
			"timing": map[string]any{
				"created_at": "2026-05-15T00:00:00Z",
			},
			"inputs": map[string]any{
				"documents": map[string]any{},
				"json_data": map[string]any{},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := workflowsRunsRestartCmd.Flags().Set("command-id", "cmd_restart"); err != nil {
		t.Fatal(err)
	}
	if err := workflowsRunsRestartCmd.Flags().Set("config-source", "published"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsRunsRestartCmd.Flags().Set("command-id", "")
		_ = workflowsRunsRestartCmd.Flags().Set("config-source", "published")
	})

	stdout, stderr := captureStd(t, func() {
		if err := workflowsRunsRestartCmd.RunE(workflowsRunsRestartCmd, []string{"run_123"}); err != nil {
			t.Fatalf("runs restart: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "run_456") {
		t.Fatalf("expected restart response on stdout, got:\n%s", stdout)
	}
	if body["command_id"] != "cmd_restart" || body["config_source"] != "published" {
		t.Fatalf("restart body = %#v", body)
	}
}

func TestWorkflowsRunsRestartRejectsInvalidConfigSourceLocally(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		t.Fatalf("server should not be reached for invalid config source, got %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	if err := workflowsRunsRestartCmd.Flags().Set("config-source", "preview"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsRunsRestartCmd.Flags().Set("config-source", "published") })

	var err error
	_, stderr := captureStd(t, func() {
		err = workflowsRunsRestartCmd.RunE(workflowsRunsRestartCmd, []string{"run_123"})
	})
	if err == nil {
		t.Fatal("expected invalid config source error")
	}
	if !strings.Contains(errors.Unwrap(err).Error(), "--config-source must be published or draft") {
		t.Fatalf("error %q does not mention valid config sources", err.Error())
	}
	if !strings.Contains(stderr, "--config-source must be published or draft") {
		t.Fatalf("stderr %q does not mention valid config sources", stderr)
	}
	if hits != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits)
	}
}

func resetWorkflowRunsFlag(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing workflow runs flag %q", name)
	}
	if slice, ok := flag.Value.(pflag.SliceValue); ok {
		if err := slice.Replace(nil); err != nil {
			t.Fatalf("reset --%s: %v", name, err)
		}
		return
	}
	if err := cmd.Flags().Set(name, ""); err != nil {
		t.Fatalf("reset --%s: %v", name, err)
	}
}

func TestWorkflowsRunsStepsListExampleUsesPaginatedEnvelope(t *testing.T) {
	example := workflowsRunsStepsListCmd.Example
	if !strings.Contains(example, ".data[]") {
		t.Fatalf("steps list example should iterate over .data[], got:\n%s", example)
	}
	if !strings.Contains(example, ".lifecycle.status") {
		t.Fatalf("steps list example should read lifecycle status, got:\n%s", example)
	}
}

func TestWorkflowsRunsGetHelpUsesLifecycleStatusPath(t *testing.T) {
	if !strings.Contains(workflowsRunsGetCmd.Example, ".lifecycle.status") {
		t.Fatalf("runs get example should read lifecycle.status:\n%s", workflowsRunsGetCmd.Example)
	}
	if strings.Contains(workflowsRunsGetCmd.Example, "jq -r '.status'") {
		t.Fatalf("runs get example should not read top-level status:\n%s", workflowsRunsGetCmd.Example)
	}
}

func TestWorkflowsRunsListExamplesUseBackendStatusNames(t *testing.T) {
	if strings.Contains(workflowsRunsListCmd.Example, "--status failed") {
		t.Fatalf("runs list example should use backend status name error, got:\n%s", workflowsRunsListCmd.Example)
	}
	if !strings.Contains(workflowsRunsListCmd.Example, "--status error") {
		t.Fatalf("runs list example should include --status error, got:\n%s", workflowsRunsListCmd.Example)
	}
}

func TestWorkflowsRunsExportRejectsInvalidDateFlagsLocally(t *testing.T) {
	cases := []struct {
		name string
		flag string
	}{
		{name: "invalid from date", flag: "from-date"},
		{name: "invalid to date", flag: "to-date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := workflowsRunsExportCmd.Flags().Set(tc.flag, "not-a-date")
			if err == nil {
				t.Fatalf("expected local parse error for --%s=not-a-date", tc.flag)
			}
			if !strings.Contains(err.Error(), "YYYY-MM-DD") {
				t.Fatalf("error %q does not contain YYYY-MM-DD", err.Error())
			}
			if resetErr := workflowsRunsExportCmd.Flags().Set(tc.flag, ""); resetErr != nil {
				t.Fatalf("reset --%s: %v", tc.flag, resetErr)
			}
		})
	}
}

func TestParseDocumentArgs_MultipleDocumentFlags(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./a.pdf", "classify=./b.pdf"},
		nil,
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want both keys", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestParseDocumentArgs_LegacyFlagEmitsOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(nil, []string{"start=./a.pdf"}, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./a.pdf" {
		t.Fatalf("got %v, want {start: ./a.pdf}", got)
	}
	if !strings.Contains(warn.String(), "--document-file is deprecated") {
		t.Fatalf("expected deprecation warning, got %q", warn.String())
	}
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_LegacyMultipleEntriesOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		nil,
		[]string{"start=./a.pdf", "classify=./b.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want both keys", got)
	}
	// Two legacy entries must still produce exactly one warning line.
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_MixedFlagsUnionOneWarning(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./a.pdf"},
		[]string{"classify=./b.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got["start"] != "./a.pdf" || got["classify"] != "./b.pdf" {
		t.Fatalf("got %v, want union of both keys", got)
	}
	if !strings.Contains(warn.String(), "--document-file is deprecated") {
		t.Fatalf("expected deprecation warning, got %q", warn.String())
	}
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_NewFlagWinsOnCollision(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(
		[]string{"start=./new.pdf"},
		[]string{"start=./legacy.pdf"},
		&warn,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["start"] != "./new.pdf" {
		t.Fatalf("got %v, want {start: ./new.pdf} (--document overrides --document-file)", got)
	}
	// Still exactly one warning line because the legacy flag was used.
	if strings.Count(warn.String(), "\n") != 1 {
		t.Fatalf("expected exactly one warning line, got %q", warn.String())
	}
}

func TestParseDocumentArgs_NoFlagsEmptyMap(t *testing.T) {
	var warn bytes.Buffer
	got, err := parseDocumentArgs(nil, nil, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty map", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("expected no warning, got %q", warn.String())
	}
}

func TestParseDocumentArgs_BadShapes(t *testing.T) {
	cases := []struct {
		name string
		docs []string
		legs []string
	}{
		{name: "missing equals on --document", docs: []string{"./a.pdf"}},
		{name: "empty key on --document", docs: []string{"=./a.pdf"}},
		{name: "empty value on --document", docs: []string{"start="}},
		{name: "missing equals on --document-file", legs: []string{"./a.pdf"}},
		{name: "empty key on --document-file", legs: []string{"=./a.pdf"}},
		{name: "empty value on --document-file", legs: []string{"start="}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var warn bytes.Buffer
			_, err := parseDocumentArgs(tc.docs, tc.legs, &warn)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestParseDocumentArgs_NilWarnSinkDoesNotPanic(t *testing.T) {
	// Smoke test: when the legacy flag is used but warnTo is nil (e.g. tests
	// that don't care about warnings), the helper must not panic.
	_, err := parseDocumentArgs(nil, []string{"start=./a.pdf"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
