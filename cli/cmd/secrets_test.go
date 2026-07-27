//go:build !retab_oagen_cli_secrets

package cmd

import (
	"bytes"
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

func TestSecretsCommandsAreRegistered(t *testing.T) {
	for _, path := range [][]string{
		{"secrets", "list"},
		{"secrets", "get"},
		{"secrets", "value"},
		{"secrets", "set"},
		{"secrets", "delete"},
	} {
		cmd, _, err := rootCmd.Find(path)
		if err != nil {
			t.Fatalf("retab %v is not registered: %v", path, err)
		}
		if cmd == nil || cmd.Name() != path[len(path)-1] {
			t.Fatalf("retab %v resolved to %v", path, cmd)
		}
	}
	if secretsSetCmd.Flags().Lookup("value") != nil {
		t.Fatal("retab secrets set must not expose --value")
	}
}

func TestSecretsValuePrintsRawValueByDefaultAndJSONWhenRequested(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	resetOutputFlag(t)

	var sawValuePath bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/secrets/RESEND_API_KEY/value" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		sawValuePath = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"secret": map[string]any{
				"name":       "RESEND_API_KEY",
				"value":      "super-secret-value",
				"updated_at": "2026-06-03T10:00:00Z",
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "secrets", "value", "RESEND_API_KEY"); err != nil {
			t.Fatalf("secrets value: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !sawValuePath {
		t.Fatal("server did not receive secret value request")
	}
	if stdout != "super-secret-value" {
		t.Fatalf("stdout = %q, want raw secret value", stdout)
	}

	jsonStdout, jsonStderr := captureStd(t, func() {
		if err := runRootForTest(t, "--output", "json", "secrets", "value", "RESEND_API_KEY"); err != nil {
			t.Fatalf("secrets value --output json: %v", err)
		}
	})
	if jsonStderr != "" {
		t.Fatalf("stderr = %q, want empty", jsonStderr)
	}
	var payload map[string]map[string]any
	if err := json.Unmarshal([]byte(jsonStdout), &payload); err != nil {
		t.Fatalf("json output did not parse: %v\n%s", err, jsonStdout)
	}
	if payload["secret"]["value"] != "super-secret-value" {
		t.Fatalf("json output should include secret value envelope, got:\n%s", jsonStdout)
	}
}

func TestSecretsSetReadsValueFromFileAndDoesNotPrintIt(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	resetSecretsSetFlags(t)

	secretPath := filepath.Join(t.TempDir(), "resend-key")
	if err := os.WriteFile(secretPath, []byte("super-secret-value\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var sawSet bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/v1/secrets/RESEND_API_KEY" {
			t.Fatalf("path = %s, want /v1/secrets/RESEND_API_KEY", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("body is not JSON: %v", err)
		}
		if payload["value"] != "super-secret-value\n" {
			t.Fatalf("value = %q, want file contents", payload["value"])
		}
		sawSet = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"secret": map[string]any{
				"name":       "RESEND_API_KEY",
				"created_at": "2026-06-03T10:00:00Z",
				"updated_at": "2026-06-03T10:00:00Z",
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "secrets", "set", "RESEND_API_KEY", "--from-file", secretPath); err != nil {
			t.Fatalf("secrets set: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !sawSet {
		t.Fatal("server did not receive set request")
	}
	if !strings.Contains(stdout, "RESEND_API_KEY") {
		t.Fatalf("stdout %q does not contain secret name", stdout)
	}
	if strings.Contains(stdout, "super-secret-value") {
		t.Fatalf("stdout leaked secret value: %q", stdout)
	}
}

func TestSecretsSetReadsValueFromStdin(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	resetSecretsSetFlags(t)

	var gotBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v1/secrets/OPENAI_API_KEY" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"secret": map[string]any{
				"name":       "OPENAI_API_KEY",
				"created_at": "2026-06-03T10:00:00Z",
				"updated_at": "2026-06-03T10:01:00Z",
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	rootCmd.SetIn(bytes.NewBufferString("stdin-secret"))
	t.Cleanup(func() { rootCmd.SetIn(nil) })

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "secrets", "set", "OPENAI_API_KEY", "--from-stdin"); err != nil {
			t.Fatalf("secrets set --from-stdin: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if gotBody["value"] != "stdin-secret" {
		t.Fatalf("value = %q, want stdin-secret", gotBody["value"])
	}
	if strings.Contains(stdout, "stdin-secret") {
		t.Fatalf("stdout leaked secret value: %q", stdout)
	}
}

func TestSecretsSetRejectsMultipleInputSources(t *testing.T) {
	resetSecretsSetFlags(t)
	err := secretsSetCmd.Flags().Set("from-file", "secret.txt")
	if err != nil {
		t.Fatal(err)
	}
	err = secretsSetCmd.Flags().Set("from-stdin", "true")
	if err != nil {
		t.Fatal(err)
	}
	err = secretsSetCmd.RunE(secretsSetCmd, []string{"RESEND_API_KEY"})
	if err == nil {
		t.Fatal("expected mutually-exclusive input source error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %q, want mutually exclusive", err.Error())
	}
}

func TestSecretsListTableOutputDoesNotExposeValues(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	resetOutputFlag(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/secrets" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"secrets": []map[string]any{
				{
					"name":       "RESEND_API_KEY",
					"created_at": "2026-06-03T10:00:00Z",
					"updated_at": "2026-06-03T10:01:00Z",
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "--output", "table", "secrets", "list"); err != nil {
			t.Fatalf("secrets list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	for _, want := range []string{"NAME", "UPDATED_AT", "RESEND_API_KEY"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "value") || strings.Contains(stdout, "encrypted_value") {
		t.Fatalf("stdout exposed value fields:\n%s", stdout)
	}
}

func TestSecretsDeleteWithYesFlagProceedsWithoutPrompt(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	resetOutputFlag(t)

	var sawDelete atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/v1/secrets/RESEND_API_KEY" {
			sawDelete.Add(1)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "secrets", "delete", "RESEND_API_KEY", "--yes"); err != nil {
			t.Fatalf("secrets delete --yes: %v", err)
		}
	})
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "deleted secret: RESEND_API_KEY") {
		t.Fatalf("stderr = %q, want delete confirmation", stderr)
	}
	if sawDelete.Load() != 1 {
		t.Fatalf("expected one DELETE call, got %d", sawDelete.Load())
	}
}

func resetSecretsSetFlags(t *testing.T) {
	t.Helper()
	_ = secretsSetCmd.Flags().Set("from-file", "")
	_ = secretsSetCmd.Flags().Set("from-stdin", "false")
	t.Cleanup(func() {
		_ = secretsSetCmd.Flags().Set("from-file", "")
		_ = secretsSetCmd.Flags().Set("from-stdin", "false")
	})
}

func resetOutputFlag(t *testing.T) {
	t.Helper()
	_ = rootCmd.PersistentFlags().Set("output", "")
	t.Cleanup(func() {
		_ = rootCmd.PersistentFlags().Set("output", "")
	})
}
