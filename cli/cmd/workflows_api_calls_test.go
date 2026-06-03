//go:build !retab_oagen_cli_workflows

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
)

func TestHydrateAPICallBundleWritesRuntimeAndEnvPlaceholders(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"method": "POST",
		"url":    "https://api.example.com/orders",
		"headers": map[string]any{
			"Authorization": "Bearer ${API_TOKEN}",
		},
		"mounts": map[string]any{
			"secrets": []any{
				map[string]any{"name": "api_token", "env": "API_TOKEN", "required": true},
			},
		},
	}
	block := retab.WorkflowBlock{
		ID:         "blk_api",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeAPICall,
		Config:     config,
		UpdatedAt:  time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatalf("write bundle: %v", err)
	}
	manifest, reassembled, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if manifest.BlockType != "api_call" {
		t.Fatalf("unexpected block type: %s", manifest.BlockType)
	}
	if err := hydrateAPICallBundle(dir, reassembled, false); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.py"), []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "curl.sh"), []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := hydrateAPICallBundle(dir, reassembled, true); err != nil {
		t.Fatalf("hydrate force: %v", err)
	}
	for _, rel := range []string{"run.sh", ".env.example", ".env.local", "samples", "rendered", "outputs", "traces"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected %s to be written: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "run.py")); !os.IsNotExist(err) {
		t.Fatalf("run.py should not be written for api_call bundles, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "curl.sh")); !os.IsNotExist(err) {
		t.Fatalf("curl.sh should not be written for api_call bundles, stat err=%v", err)
	}
	envLocal, err := os.ReadFile(filepath.Join(dir, ".env.local"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(envLocal), "API_TOKEN=__REPLACE_ME__") {
		t.Fatalf(".env.local should contain placeholder secret, got:\n%s", envLocal)
	}
}

func TestWorkflowsBlocksPullConfigAutoHydratesAPICallRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_api",
			"workflow_id": "wf_cfg",
			"type":        "api_call",
			"label":       "API",
			"updated_at":  "2026-06-03T10:00:00Z",
			"config": map[string]any{
				"method":  "POST",
				"url":     "https://api.example.com/orders",
				"headers": map[string]any{"Authorization": "Bearer ${API_TOKEN}"},
				"mounts": map[string]any{
					"secrets": []any{map[string]any{"name": "api_token", "env": "API_TOKEN"}},
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := filepath.Join(t.TempDir(), "bundle")
	if err := workflowsBlocksPullConfigCmd.Flags().Set("out", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPullConfigCmd.Flags().Set("out", "")
		_ = workflowsBlocksPullConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPullConfigCmd.Flags().Set("force", "false")
	})

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksPullConfigCmd.RunE(workflowsBlocksPullConfigCmd, []string{"wf_cfg", "blk_api"}); err != nil {
			t.Fatalf("pull-config: %v", err)
		}
	})
	if !strings.Contains(stdout, `"runtime_hydrated": true`) {
		t.Fatalf("api_call pull-config should report runtime_hydrated=true, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"runtime_adapter": "api_call"`) {
		t.Fatalf("api_call pull-config should report runtime_adapter=api_call, got:\n%s", stdout)
	}
	for _, rel := range []string{"run.sh", ".env.example", ".env.local", "rendered", "outputs", "traces"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("runtime support file %s should be written by pull-config: %v", rel, err)
		}
	}
	for _, rel := range []string{"run.py", "curl.sh"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(err) {
			t.Fatalf("stale runtime support file %s should not be written by pull-config, stat err=%v", rel, err)
		}
	}
}

func TestAPICallLocalRunRendersWithoutExecutingByDefault(t *testing.T) {
	t.Setenv("API_TOKEN", "")
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	dir := createHydratedAPICallBundle(t, map[string]any{
		"method": "POST",
		"url":    server.URL + "/orders",
		"headers": map[string]any{
			"Authorization": "Bearer ${API_TOKEN}",
		},
		"field_mappings": map[string]any{"order_id": "id"},
		"request_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string"},
			},
		},
		"mounts": map[string]any{
			"secrets": []any{
				map[string]any{"name": "api_token", "env": "API_TOKEN", "required": true},
			},
		},
	})
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("API_TOKEN=local-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sample := filepath.Join(dir, "samples", "order.json")
	if err := os.WriteFile(sample, []byte(`{"order_id":"ord_123","extra":"drop"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout := runAPICallCommandForTest(t, dir, sample, false)
	if !strings.Contains(stdout, `"request":"rendered/samples/order.request.json"`) {
		t.Fatalf("stdout should use relative request path, got:\n%s", stdout)
	}
	if called {
		t.Fatal("dry-run should not make HTTP requests")
	}
	body := readJSONMapFromPath(t, filepath.Join(dir, "rendered", "samples", "order.body.json"))
	if body["id"] != "ord_123" {
		t.Fatalf("rendered body did not map/filter input: %+v", body)
	}
	request := readJSONMapFromPath(t, filepath.Join(dir, "rendered", "samples", "order.request.json"))
	if request["input"] != "samples/order.json" {
		t.Fatalf("request artifact should use relative input path: %+v", request)
	}
	if request["body_path"] != "rendered/samples/order.body.json" {
		t.Fatalf("request artifact should use relative body path: %+v", request)
	}
	if request["execute_command_path"] != "rendered/samples/order.curl.sh" {
		t.Fatalf("request artifact should point at rendered curl script: %+v", request)
	}
	headers := request["headers"].(map[string]any)
	if headers["Authorization"] != "[REDACTED]" {
		t.Fatalf("request artifact should redact authorization header: %+v", request)
	}
	if _, err := os.Stat(filepath.Join(dir, "outputs", "samples", "order.out.json")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write an output response, stat err=%v", err)
	}
	curlScript, err := os.ReadFile(filepath.Join(dir, "rendered", "samples", "order.curl.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(curlScript), "curl") || !strings.Contains(string(curlScript), "--data-binary") {
		t.Fatalf("rendered curl script should be standalone curl, got:\n%s", curlScript)
	}
}

func TestAPICallLocalRunExecutePostsToLocalServer(t *testing.T) {
	t.Setenv("API_TOKEN", "")
	var gotAuth string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"accepted": true})
	}))
	defer server.Close()

	dir := createHydratedAPICallBundle(t, map[string]any{
		"method":  "POST",
		"url":     server.URL + "/orders",
		"headers": map[string]any{"Authorization": "Bearer ${API_TOKEN}"},
		"mounts": map[string]any{
			"secrets": []any{map[string]any{"name": "api_token", "env": "API_TOKEN", "required": true}},
		},
	})
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("API_TOKEN=local-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sample := filepath.Join(dir, "samples", "order.json")
	if err := os.WriteFile(sample, []byte(`{"id":"ord_123"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	runAPICallCommandForTest(t, dir, sample, true)
	if gotAuth != "Bearer local-token" {
		t.Fatalf("Authorization header = %q", gotAuth)
	}
	if gotBody["id"] != "ord_123" {
		t.Fatalf("request body = %+v", gotBody)
	}
	response := readJSONMapFromPath(t, filepath.Join(dir, "outputs", "samples", "order.out.json"))
	if response["status_code"].(float64) != 200 {
		t.Fatalf("unexpected response artifact: %+v", response)
	}
	if response["ok"] != true {
		t.Fatalf("response artifact should be ok: %+v", response)
	}
}

func TestAPICallLocalRunCanEmitAbsolutePaths(t *testing.T) {
	t.Setenv("API_TOKEN", "")
	dir := createHydratedAPICallBundle(t, map[string]any{
		"method":  "POST",
		"url":     "https://api.example.com/orders",
		"headers": map[string]any{"Authorization": "Bearer ${API_TOKEN}"},
		"mounts": map[string]any{
			"secrets": []any{map[string]any{"name": "api_token", "env": "API_TOKEN", "required": true}},
		},
	})
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("API_TOKEN=local-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sample := filepath.Join(dir, "samples", "order.json")
	if err := os.WriteFile(sample, []byte(`{"id":"ord_123"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	flags := workflowsAPICallsRunCmd.Flags()
	if err := flags.Set("absolute-paths", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = flags.Set("absolute-paths", "false")
	})
	stdout := runAPICallCommandForTest(t, dir, sample, false)
	if !strings.Contains(stdout, filepath.Join(dir, "rendered", "samples", "order.request.json")) {
		t.Fatalf("stdout should include absolute request path, got:\n%s", stdout)
	}
	request := readJSONMapFromPath(t, filepath.Join(dir, "rendered", "samples", "order.request.json"))
	if request["input"] != sample {
		t.Fatalf("request artifact should use absolute input path: %+v", request)
	}
}

func TestAPICallHydrateFillSecretsWritesEnvLocalWithoutPrintingValues(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/secrets/api_token/value" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"secret": map[string]any{
				"name":       "api_token",
				"value":      "server-secret-token",
				"updated_at": "2026-06-03T10:00:00Z",
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	dir := createHydratedAPICallBundle(t, map[string]any{
		"method": "POST",
		"url":    "https://api.example.com/orders",
		"mounts": map[string]any{
			"secrets": []any{map[string]any{"name": "api_token", "env": "API_TOKEN", "required": true}},
		},
	})
	flags := workflowsAPICallsHydrateCmd.Flags()
	if err := flags.Set("fill-secrets", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = flags.Set("force", "false")
		_ = flags.Set("fill-secrets", "false")
		_ = flags.Set("force-secrets", "false")
	})

	stdout, stderr := captureStd(t, func() {
		if err := workflowsAPICallsHydrateCmd.RunE(workflowsAPICallsHydrateCmd, []string{dir}); err != nil {
			t.Fatalf("hydrate --fill-secrets: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, "server-secret-token") {
		t.Fatalf("hydrate output leaked secret value:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"written": true`) {
		t.Fatalf("hydrate output should report written=true, got:\n%s", stdout)
	}
	envLocal, err := os.ReadFile(filepath.Join(dir, ".env.local"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(envLocal), "API_TOKEN=server-secret-token") {
		t.Fatalf(".env.local did not receive filled value:\n%s", envLocal)
	}
}

func TestFillSecretsEnvFilePreservesExistingValuesUnlessForced(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.local")
	if err := os.WriteFile(path, []byte("API_TOKEN=local-token\nOTHER=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	written, err := writeLocalSecretsEnvFile(path, map[string]string{"API_TOKEN": "server-token", "NEW_TOKEN": "new-token"}, false)
	if err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if written["API_TOKEN"] {
		t.Fatal("existing API_TOKEN should not be overwritten without force")
	}
	if !written["NEW_TOKEN"] {
		t.Fatal("missing NEW_TOKEN should be appended")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"API_TOKEN=local-token", "OTHER=value", "NEW_TOKEN=new-token"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf(".env.local missing %q:\n%s", want, content)
		}
	}

	written, err = writeLocalSecretsEnvFile(path, map[string]string{"API_TOKEN": "server-token"}, true)
	if err != nil {
		t.Fatalf("force write env file: %v", err)
	}
	if !written["API_TOKEN"] {
		t.Fatal("force should report API_TOKEN written")
	}
	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "API_TOKEN=server-token") {
		t.Fatalf("force did not overwrite API_TOKEN:\n%s", content)
	}
}

func createHydratedAPICallBundle(t *testing.T, config map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	block := retab.WorkflowBlock{
		ID:         "blk_api",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeAPICall,
		Config:     config,
		UpdatedAt:  time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatalf("write bundle: %v", err)
	}
	_, reassembled, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if err := hydrateAPICallBundle(dir, reassembled, false); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "samples"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runAPICallCommandForTest(t *testing.T, dir string, sample string, execute bool) string {
	t.Helper()
	flags := workflowsAPICallsRunCmd.Flags()
	if err := flags.Set("clean", "true"); err != nil {
		t.Fatal(err)
	}
	if err := flags.Set("execute", strconv.FormatBool(execute)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = flags.Set("out", "outputs")
		_ = flags.Set("jobs", "auto")
		_ = flags.Set("timeout", "0")
		_ = flags.Set("recursive", "false")
		_ = flags.Set("continue-on-error", "false")
		_ = flags.Set("clean", "false")
		_ = flags.Set("execute", "false")
	})
	stdout, _ := captureStd(t, func() {
		if err := workflowsAPICallsRunCmd.RunE(workflowsAPICallsRunCmd, []string{dir, sample}); err != nil {
			t.Fatalf("api-calls run failed: %v", err)
		}
	})
	return stdout
}

func readJSONMapFromPath(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
