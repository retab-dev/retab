//go:build !retab_oagen_cli_workflows

package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestValidateFunctionRunJobs(t *testing.T) {
	for _, raw := range []string{"", "auto", "1", "8"} {
		if err := validateFunctionRunJobs(raw); err != nil {
			t.Fatalf("validateFunctionRunJobs(%q) returned unexpected error: %v", raw, err)
		}
	}
	for _, raw := range []string{"0", "-1", "many"} {
		if err := validateFunctionRunJobs(raw); err == nil {
			t.Fatalf("validateFunctionRunJobs(%q) returned nil error", raw)
		}
	}
}

func TestValidateFunctionRunOutDir(t *testing.T) {
	for _, raw := range []string{"outputs", "tmp/results", "a/b"} {
		if err := validateFunctionRunOutDir(raw); err != nil {
			t.Fatalf("validateFunctionRunOutDir(%q) returned unexpected error: %v", raw, err)
		}
	}
	for _, raw := range []string{"", ".", "..", "../outside", "/tmp/out"} {
		if err := validateFunctionRunOutDir(raw); err == nil {
			t.Fatalf("validateFunctionRunOutDir(%q) returned nil error", raw)
		}
	}
}

func TestParseFunctionRunTimeout(t *testing.T) {
	got, err := parseFunctionRunTimeout("5m")
	if err != nil {
		t.Fatalf("parse timeout: %v", err)
	}
	if got != 5*time.Minute {
		t.Fatalf("timeout = %s, want 5m", got)
	}
	for _, raw := range []string{"", "0"} {
		got, err := parseFunctionRunTimeout(raw)
		if err != nil {
			t.Fatalf("parseFunctionRunTimeout(%q): %v", raw, err)
		}
		if got != 0 {
			t.Fatalf("parseFunctionRunTimeout(%q) = %s, want 0", raw, got)
		}
	}
	for _, raw := range []string{"abc", "-1s"} {
		if _, err := parseFunctionRunTimeout(raw); err == nil {
			t.Fatalf("parseFunctionRunTimeout(%q) returned nil error", raw)
		}
	}
}

func TestHydrateFunctionBundleWritesLocalMountsWithoutMutatingConfigMounts(t *testing.T) {
	dir := t.TempDir()
	configMounts := map[string]any{
		"tables": []any{
			map[string]any{
				"name":     "customers",
				"table_id": "tbl_customers",
				"path":     "/mnt/tables/customers.csv",
			},
		},
	}
	config := map[string]any{
		"mounts": configMounts,
	}
	if err := writeJSONFile(filepath.Join(dir, "mounts.json"), configMounts); err != nil {
		t.Fatalf("write mounts.json: %v", err)
	}

	if err := hydrateFunctionBundle(dir, config, nil, false); err != nil {
		t.Fatalf("hydrate: %v", err)
	}

	mounts, err := readJSONMap(filepath.Join(dir, "mounts.json"))
	if err != nil {
		t.Fatalf("read mounts.json: %v", err)
	}
	rawConfigTables, _ := mounts["tables"].([]any)
	if len(rawConfigTables) != 1 {
		t.Fatalf("mounts.json tables = %#v", rawConfigTables)
	}
	configTable, _ := rawConfigTables[0].(map[string]any)
	if configTable["local_path"] != "fixtures/tables/customers.csv" {
		t.Fatalf("mounts.json local_path = %#v, want fixtures/tables/customers.csv", configTable["local_path"])
	}
	if _, err := os.Stat(filepath.Join(dir, "mounts.local.json")); !os.IsNotExist(err) {
		t.Fatalf("mounts.local.json should not exist, stat err=%v", err)
	}
}

// Regression: function hydrate keyed the generated .env files by each secret's
// display `name` instead of its `env` field. When the two differ, the runtime
// (which reads the secret via its env-var name) saw no value, and --fill-secrets
// (which keys by env) appended a duplicate line. The sibling api_call hydrate
// already keys by env; function hydrate must match.
func TestHydrateFunctionBundleEnvFilesKeyedByEnvNotName(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"mounts": map[string]any{
			"secrets": []any{
				map[string]any{"name": "prod-openai-key", "env": "OPENAI_API_KEY"},
			},
		},
	}
	if err := writeJSONFile(filepath.Join(dir, "mounts.json"), config["mounts"]); err != nil {
		t.Fatalf("write mounts.json: %v", err)
	}
	if err := hydrateFunctionBundle(dir, config, nil, false); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	for _, name := range []string{".env.example", ".env.local"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		content := string(raw)
		if !strings.Contains(content, "OPENAI_API_KEY=") {
			t.Fatalf("%s should be keyed by the env var OPENAI_API_KEY, got:\n%s", name, content)
		}
		if strings.Contains(content, "prod-openai-key") {
			t.Fatalf("%s leaked the secret display name instead of the env var, got:\n%s", name, content)
		}
	}
}

func TestHydrateFunctionBundleWritesTinyRunPyAndRuntimeSupport(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"entrypoint": "transform",
	}
	if err := hydrateFunctionBundle(dir, config, nil, false); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	runPy, err := os.ReadFile(filepath.Join(dir, "run.py"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(runPy), "ThreadPoolExecutor") {
		t.Fatalf("run.py should be a tiny launcher, got:\n%s", runPy)
	}
	if !strings.Contains(string(runPy), ".retab") || !strings.Contains(string(runPy), "runtime.py") {
		t.Fatalf("run.py should load .retab/runtime.py, got:\n%s", runPy)
	}
	runtimePy, err := os.ReadFile(filepath.Join(dir, ".retab", "runtime.py"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(runtimePy), "ThreadPoolExecutor") {
		t.Fatalf(".retab/runtime.py should contain the generated runtime implementation")
	}
}

// TestHydratePythonFunctionBundleWritesModelsCompatModule reproduces the
// `functions run` failure "ModuleNotFoundError: No module named 'models'".
// Server-run function code uses the canonical `from models import Input, Output`
// (the sandbox writes a models.py compat module), but the local hydrator only
// wrote input.py/output.py, so a pulled block could not run locally. The
// hydrator must emit a models.py re-exporting Input + the output classes.
func TestHydratePythonFunctionBundleWritesModelsCompatModule(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"entrypoint": "transform",
		"output_schema": map[string]any{
			"type":       "object",
			"properties": map[string]any{"doubled": map[string]any{"type": "integer"}},
		},
	}
	if err := hydrateFunctionBundle(dir, config, nil, false); err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "models.py"))
	if err != nil {
		t.Fatalf("models.py not written by hydrator (functions run would fail with ModuleNotFoundError: 'models'): %v", err)
	}
	if !strings.Contains(string(body), "from input import Input") {
		t.Fatalf("models.py must re-export Input from input.py, got:\n%s", body)
	}
	if !strings.Contains(string(body), "from output import") || !strings.Contains(string(body), "Output") {
		t.Fatalf("models.py must re-export Output from output.py, got:\n%s", body)
	}
}

// TestHydratePythonFunctionBundleForcePreservesModelsModule ensures --force
// (re-hydrate) keeps the freshly written models.py rather than deleting it as a
// stale artifact.
func TestHydratePythonFunctionBundleForcePreservesModelsModule(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{"entrypoint": "transform"}
	if err := hydrateFunctionBundle(dir, config, nil, true); err != nil {
		t.Fatalf("hydrate --force: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(dir, "models.py")); err != nil {
		t.Fatalf("models.py must survive a --force re-hydrate: %v", err)
	}
}

func TestRunFunctionPythonChildForwardsArgsAndDetachesStdin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is unix-specific")
	}
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "fake-python.sh")
	argsPath := filepath.Join(dir, "args.txt")
	stdinPath := filepath.Join(dir, "stdin.txt")
	script := `#!/bin/sh
printf '%s\n' "$@" > args.txt
if [ -t 0 ]; then
  printf 'tty\n' > stdin.txt
else
  printf 'detached\n' > stdin.txt
fi
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake python: %v", err)
	}

	err := runFunctionPythonChild(nil, functionPythonChildOptions{
		Python: scriptPath,
		Args:   []string{"run.py", "--jobs", "2", "sample.json"},
		Dir:    dir,
	})
	if err != nil {
		t.Fatalf("run child: %v", err)
	}
	argsRaw, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	if strings.TrimSpace(string(argsRaw)) != "run.py\n--jobs\n2\nsample.json" {
		t.Fatalf("unexpected forwarded args:\n%s", argsRaw)
	}
	stdinRaw, err := os.ReadFile(stdinPath)
	if err != nil {
		t.Fatalf("read stdin marker: %v", err)
	}
	if strings.TrimSpace(string(stdinRaw)) != "detached" {
		t.Fatalf("stdin marker = %q, want detached", strings.TrimSpace(string(stdinRaw)))
	}
}

func TestRunFunctionPythonChildTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is unix-specific")
	}
	err := runFunctionPythonChild(nil, functionPythonChildOptions{
		Python:  "/bin/sh",
		Args:    []string{"-c", "sleep 2"},
		Dir:     t.TempDir(),
		Timeout: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("expected timeout message, got %v", err)
	}
}

func TestRunFunctionPythonChildExitErrorMentionsTraces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is unix-specific")
	}
	dir := t.TempDir()
	err := runFunctionPythonChild(nil, functionPythonChildOptions{
		Python: "/bin/sh",
		Args:   []string{"-c", "exit 7"},
		Dir:    dir,
	})
	if err == nil {
		t.Fatal("expected child exit error, got nil")
	}
	if !strings.Contains(err.Error(), "inspect trace files") {
		t.Fatalf("expected trace hint, got %v", err)
	}
	if !strings.Contains(err.Error(), filepath.Join(dir, "traces")) {
		t.Fatalf("expected trace directory in error, got %v", err)
	}
}

func TestGenerateTypescriptModelModuleUsesInputAndOutputSchemas(t *testing.T) {
	sourceSchema := map[string]any{
		"type": "object",
		"$defs": map[string]any{
			"line_item": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sku":      map[string]any{"type": "string"},
					"quantity": map[string]any{"type": "integer"},
				},
				"required": []any{"sku"},
			},
		},
		"properties": map[string]any{
			"customer_name": map[string]any{"type": "string"},
			"items": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/$defs/line_item"},
			},
			"priority": map[string]any{"enum": []any{"standard", "rush"}},
		},
		"required": []any{"customer_name", "items"},
	}
	config := map[string]any{
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"ok":     map[string]any{"type": "boolean"},
				"status": map[string]any{"type": []any{"string", "null"}},
			},
			"required": []any{"ok"},
		},
	}

	got := generateTypescriptModelModule(sourceSchema, config)
	for _, want := range []string{
		"export type LineItem = {",
		"  quantity?: number;",
		"  sku: string;",
		"export type Input = {",
		"  customer_name: string;",
		"  items: Array<LineItem>;",
		"  priority?: \"standard\" | \"rush\";",
		"export type Output = {",
		"  ok: boolean;",
		"  status?: string | null;",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript models should contain %q, got:\n%s", want, got)
		}
	}
}
