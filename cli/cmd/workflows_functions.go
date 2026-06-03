//go:build !retab_oagen_cli_workflows

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var workflowsFunctionsCmd = &cobra.Command{
	Use:   "functions",
	Short: "Develop workflow function blocks locally",
	Long: `Hydrate and run local workflow function bundles.

Start by pulling a function block with:

  retab workflows blocks pull-config <workflow-id> <block-id> --out tmp/fn

Function bundles pulled with pull-config are hydrated automatically. Re-run
hydrate when you need to repair or regenerate local support files:

  retab workflows blocks functions hydrate tmp/fn

Hydration writes input.py, output.py, .env.local placeholders, local table
fixture paths in mounts.json, and a run.py wrapper. Secrets are not downloaded
by default. If your local environment already defines the variables declared in
mounts.secrets, run.py uses them without prompting.`,
}

var workflowsFunctionsHydrateCmd = &cobra.Command{
	Use:   "hydrate <bundle-dir>",
	Short: "Create local runtime files for a pulled function block bundle",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		fillSecrets, _ := cmd.Flags().GetBool("fill-secrets")
		forceSecrets, _ := cmd.Flags().GetBool("force-secrets")
		manifest, config, err := readBlockConfigBundle(args[0])
		if err != nil {
			return err
		}
		if manifest.BlockType != "function" {
			return fmt.Errorf("bundle block_type is %q, expected function", manifest.BlockType)
		}
		if err := hydrateFunctionBundle(args[0], config, force); err != nil {
			return err
		}
		filledSecrets := []map[string]any{}
		if fillSecrets {
			filledSecrets, err = fillLocalSecretsFromRetab(cmd, config, forceSecrets)
			if err != nil {
				return err
			}
		}
		return printJSON(map[string]any{
			"ok":          true,
			"dir":         args[0],
			"workflow_id": manifest.WorkflowID,
			"block_id":    manifest.BlockID,
			"files": []string{
				"input.py",
				"output.py",
				"run.py",
				".env.example",
				".env.local",
				"mounts.json",
				".retab/runtime.py",
			},
			"filled_secrets": filledSecrets,
		})
	}),
}

var workflowsFunctionsRunCmd = &cobra.Command{
	Use:   "run <bundle-dir> <input-json>...",
	Short: "Run a hydrated function bundle against local JSON samples",
	Args:  cobra.MinimumNArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		dir, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		runPath := filepath.Join(dir, "run.py")
		if _, err := os.Stat(runPath); err != nil {
			return fmt.Errorf("run.py not found; run `retab workflows blocks functions hydrate %s` first: %w", dir, err)
		}
		python, _ := cmd.Flags().GetString("python")
		if strings.TrimSpace(python) == "" {
			python = "python3"
		}
		outDir, _ := cmd.Flags().GetString("out")
		jobs, _ := cmd.Flags().GetString("jobs")
		timeoutRaw, _ := cmd.Flags().GetString("timeout")
		recursive, _ := cmd.Flags().GetBool("recursive")
		continueOnError, _ := cmd.Flags().GetBool("continue-on-error")
		clean, _ := cmd.Flags().GetBool("clean")
		timeout, err := parseFunctionRunTimeout(timeoutRaw)
		if err != nil {
			return err
		}
		if err := validateFunctionRunJobs(jobs); err != nil {
			return err
		}
		if err := validateFunctionRunOutDir(outDir); err != nil {
			return err
		}

		inputs := make([]string, 0, len(args)-1)
		for _, raw := range args[1:] {
			inputPath, err := filepath.Abs(raw)
			if err != nil {
				return err
			}
			inputs = append(inputs, inputPath)
		}
		childArgs := []string{runPath, "--out", outDir, "--jobs", jobs}
		if recursive {
			childArgs = append(childArgs, "--recursive")
		}
		if continueOnError {
			childArgs = append(childArgs, "--continue-on-error")
		}
		if clean {
			childArgs = append(childArgs, "--clean")
		}
		childArgs = append(childArgs, inputs...)
		return runFunctionPythonChild(cmd.Context(), functionPythonChildOptions{
			Python:  python,
			Args:    childArgs,
			Dir:     dir,
			Timeout: timeout,
		})
	}),
}

type functionPythonChildOptions struct {
	Python  string
	Args    []string
	Dir     string
	Timeout time.Duration
}

func runFunctionPythonChild(ctx context.Context, opts functionPythonChildOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	child := exec.CommandContext(ctx, opts.Python, opts.Args...)
	child.Dir = opts.Dir
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	// The generated local runner is file-driven and never reads stdin.
	// Leaving stdin detached avoids hangs in non-interactive shells and CI.
	child.Stdin = nil
	if err := child.Run(); err != nil {
		if opts.Timeout > 0 && ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("function run timed out after %s", opts.Timeout)
		}
		if _, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("local function run failed; inspect trace files under %s: %w", filepath.Join(opts.Dir, "traces"), err)
		}
		return err
	}
	return nil
}

func parseFunctionRunTimeout(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "0" {
		return 0, nil
	}
	timeout, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("--timeout must be a Go duration like 30s, 5m, or 1h: %w", err)
	}
	if timeout < 0 {
		return 0, fmt.Errorf("--timeout must be non-negative")
	}
	return timeout, nil
}

func validateFunctionRunJobs(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "auto" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fmt.Errorf("--jobs must be \"auto\" or a positive integer")
	}
	return nil
}

func validateFunctionRunOutDir(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("--out is required")
	}
	if filepath.IsAbs(raw) {
		return fmt.Errorf("--out must be a relative path inside the function bundle")
	}
	cleaned := filepath.Clean(raw)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("--out must stay inside the function bundle")
	}
	return nil
}

func hydrateFunctionBundle(dir string, config map[string]any, force bool) error {
	files := map[string]string{
		"input.py":          generateFunctionInputModule(),
		"output.py":         generateFunctionOutputModule(config),
		"run.py":            generatedFunctionRunPy,
		".retab/runtime.py": generatedFunctionRuntimePy,
	}
	for rel, content := range files {
		if err := writeTextFileIfAllowed(filepath.Join(dir, rel), content, force, 0o700); err != nil {
			return err
		}
	}
	if force {
		for _, rel := range []string{"models.py", "input_models.py", "output_models.py", "retab_runtime.py", "mounts.local.json"} {
			if err := os.Remove(filepath.Join(dir, rel)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	secrets := collectFunctionSecretEnvNames(config)
	if err := writeTextFileIfAllowed(filepath.Join(dir, ".env.example"), renderEnvFile(secrets, false), force, 0o600); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, ".env.local")); os.IsNotExist(err) {
		if err := writeTextFileIfAllowed(filepath.Join(dir, ".env.local"), renderEnvFile(secrets, true), true, 0o600); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "fixtures", "tables"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "samples"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "outputs"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "traces"), 0o755); err != nil {
		return err
	}
	return writeHydratedMountsFile(dir, config)
}

func writeTextFileIfAllowed(path string, content string, force bool, perm os.FileMode) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), perm)
}

func collectFunctionSecretEnvNames(config map[string]any) []string {
	mounts, _ := config["mounts"].(map[string]any)
	rawSecrets, _ := mounts["secrets"].([]any)
	seen := map[string]bool{}
	var names []string
	for _, raw := range rawSecrets {
		secret, _ := raw.(map[string]any)
		name, _ := secret["env"].(string)
		if strings.TrimSpace(name) == "" {
			name, _ = secret["name"].(string)
		}
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func renderEnvFile(names []string, placeholders bool) string {
	if len(names) == 0 {
		return "# No mounts.secrets env vars are declared for this function.\n"
	}
	var b strings.Builder
	for _, name := range names {
		if placeholders {
			fmt.Fprintf(&b, "%s=__REPLACE_ME__\n", name)
		} else {
			fmt.Fprintf(&b, "%s=\n", name)
		}
	}
	return b.String()
}

func writeHydratedMountsFile(dir string, config map[string]any) error {
	mounts := buildHydratedMounts(config)
	if len(mounts) == 0 {
		return nil
	}
	path := filepath.Join(dir, "mounts.json")
	if existing, err := readJSONMap(path); err == nil {
		preserveExistingTableLocalPaths(mounts, existing)
	}
	return writeJSONFile(path, mounts)
}

func buildHydratedMounts(config map[string]any) map[string]any {
	mounts, _ := config["mounts"].(map[string]any)
	if mounts == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(mounts))
	for key, value := range mounts {
		out[key] = value
	}
	rawTables, _ := mounts["tables"].([]any)
	tables := make([]any, 0, len(rawTables))
	for _, raw := range rawTables {
		table, _ := raw.(map[string]any)
		sandboxPath, _ := table["path"].(string)
		tableID, _ := table["table_id"].(string)
		if strings.TrimSpace(sandboxPath) == "" {
			tables = append(tables, raw)
			continue
		}
		localTable := make(map[string]any, len(table)+1)
		for key, value := range table {
			localTable[key] = value
		}
		localName := filepath.Base(sandboxPath)
		if localName == "." || localName == string(filepath.Separator) {
			localName = sanitizeIdentifier(tableID)
			if localName == "" {
				localName = "table.csv"
			}
		}
		if strings.TrimSpace(stringFromAny(localTable["local_path"])) == "" {
			localTable["local_path"] = filepath.ToSlash(filepath.Join("fixtures", "tables", localName))
		}
		tables = append(tables, localTable)
	}
	if len(rawTables) > 0 {
		out["tables"] = tables
	}
	return out
}

func preserveExistingTableLocalPaths(mounts map[string]any, existing map[string]any) {
	rawTables, _ := mounts["tables"].([]any)
	existingTables, _ := existing["tables"].([]any)
	localPaths := map[string]string{}
	for index, raw := range existingTables {
		table, _ := raw.(map[string]any)
		localPath := strings.TrimSpace(stringFromAny(table["local_path"]))
		if localPath == "" {
			continue
		}
		key := tableMountKey(index, table)
		localPaths[key] = localPath
	}
	for index, raw := range rawTables {
		table, _ := raw.(map[string]any)
		if table == nil {
			continue
		}
		key := tableMountKey(index, table)
		if localPath := localPaths[key]; localPath != "" {
			table["local_path"] = localPath
		}
	}
}

func tableMountKey(index int, table map[string]any) string {
	tableID := strings.TrimSpace(stringFromAny(table["table_id"]))
	path := strings.TrimSpace(stringFromAny(table["path"]))
	if tableID != "" || path != "" {
		return tableID + "\x00" + path
	}
	return fmt.Sprintf("#%d", index)
}

func stringFromAny(value any) string {
	str, _ := value.(string)
	return str
}

func generateFunctionOutputModule(config map[string]any) string {
	classNames := map[string]bool{"Output": true}
	if schema, ok := config["output_schema"].(map[string]any); ok {
		for _, defsKey := range []string{"$defs", "definitions"} {
			if defs, ok := schema[defsKey].(map[string]any); ok {
				for name := range defs {
					if className := sanitizePythonClassName(name); className != "" {
						classNames[className] = true
					}
				}
			}
		}
	}
	names := []string{}
	for name := range classNames {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\nfrom input import _RetabModel\n")
	for _, name := range names {
		fmt.Fprintf(&b, "\nclass %s(_RetabModel):\n    pass\n", name)
	}
	return b.String()
}

func generateFunctionInputModule() string {
	var b strings.Builder
	b.WriteString(generatedModelBasePy)
	b.WriteString("\nclass Input(_RetabModel):\n    pass\n")
	return b.String()
}

func sanitizePythonClassName(raw string) string {
	identifier := sanitizeIdentifier(raw)
	if identifier == "" {
		return ""
	}
	parts := strings.Split(identifier, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	name := strings.Join(parts, "")
	if name == "" {
		return ""
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "Model" + name
	}
	return name
}

var nonIdentifierChars = regexp.MustCompile(`[^A-Za-z0-9_]+`)

func sanitizeIdentifier(raw string) string {
	value := nonIdentifierChars.ReplaceAllString(strings.TrimSpace(raw), "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return ""
	}
	return value
}

const generatedModelBasePy = `from __future__ import annotations


def _wrap(value):
    if isinstance(value, dict):
        return _RetabModel(**value)
    if isinstance(value, list):
        return [_wrap(item) for item in value]
    return value


def _dump(value):
    if isinstance(value, _RetabModel):
        return {key: _dump(val) for key, val in value.__dict__.items()}
    if isinstance(value, list):
        return [_dump(item) for item in value]
    if isinstance(value, dict):
        return {key: _dump(val) for key, val in value.items()}
    return value


class _RetabModel:
    def __init__(self, **kwargs):
        for key, value in kwargs.items():
            setattr(self, key, _wrap(value))

    def model_dump(self, *args, **kwargs):
        return _dump(self)

    def dict(self, *args, **kwargs):
        return self.model_dump()

    def __repr__(self):
        return f"{type(self).__name__}({self.__dict__!r})"
`

const generatedFunctionRunPy = `from __future__ import annotations

import importlib.util
import os
import sys
from pathlib import Path


def _load_runtime():
    bundle_dir = Path(__file__).resolve().parent
    os.chdir(bundle_dir)
    sys.path.insert(0, str(bundle_dir))
    runtime_path = bundle_dir / ".retab" / "runtime.py"
    spec = importlib.util.spec_from_file_location("retab_local_runtime", runtime_path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"could not load Retab local runtime from {runtime_path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


if __name__ == "__main__":
    _load_runtime().main()
`

const generatedFunctionRuntimePy = `from __future__ import annotations

import argparse
import importlib.util
import json
import os
import shutil
import sys
import traceback
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

from input import Input


def load_env_file(path: str | Path) -> None:
    env_path = Path(path)
    if not env_path.exists():
        return
    for raw_line in env_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        os.environ.setdefault(key.strip(), value.strip())


def prepare_local_mounts(bundle_dir: str | Path) -> None:
    bundle_path = Path(bundle_dir)
    mounts_path = bundle_path / "mounts.json"
    if not mounts_path.exists():
        return
    mounts = json.loads(mounts_path.read_text(encoding="utf-8"))
    for table in mounts.get("tables", []):
        local_path_raw = table.get("local_path")
        sandbox_path_raw = table.get("path")
        if not local_path_raw or not sandbox_path_raw:
            continue
        local_path = bundle_path / local_path_raw
        sandbox_path = Path(sandbox_path_raw)
        if not local_path.exists():
            continue
        sandbox_path.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(local_path, sandbox_path)


def load_function(bundle_dir: Path):
    function_path = bundle_dir / "function.py"
    spec = importlib.util.spec_from_file_location("retab_user_function", function_path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"could not load function module from {function_path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def dump_value(value):
    if hasattr(value, "model_dump"):
        return value.model_dump()
    if hasattr(value, "dict"):
        return value.dict()
    if isinstance(value, dict):
        return {key: dump_value(val) for key, val in value.items()}
    if isinstance(value, list):
        return [dump_value(item) for item in value]
    if hasattr(value, "__dict__"):
        return {key: dump_value(val) for key, val in value.__dict__.items()}
    return value


def iter_inputs(paths, recursive):
    for raw in paths:
        path = Path(raw)
        if path.is_dir():
            pattern = "**/*.json" if recursive else "*.json"
            yield from sorted(path.glob(pattern))
        else:
            yield path


def output_stem_for(input_path):
    resolved = input_path.resolve()
    try:
        rel = resolved.relative_to(Path.cwd().resolve())
    except ValueError:
        rel = Path(input_path.name)
    return rel.with_suffix("")


def run_one(bundle_dir, entrypoint, input_path, out_dir, trace_dir):
    payload = json.loads(input_path.read_text(encoding="utf-8"))
    input_data = Input(**payload) if isinstance(payload, dict) else payload
    result = entrypoint(input_data)
    output = dump_value(result)
    output_stem = output_stem_for(input_path)
    out_path = out_dir / output_stem.parent / f"{output_stem.name}.out.json"
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(output, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    trace_path = trace_dir / output_stem.parent / f"{output_stem.name}.trace.json"
    trace_path.parent.mkdir(parents=True, exist_ok=True)
    trace_path.write_text(
        json.dumps({"input": str(input_path), "output": str(out_path), "ok": True}, indent=2) + "\n",
        encoding="utf-8",
    )
    return {"input": str(input_path), "output": str(out_path), "ok": True}


def main(argv=None):
    parser = argparse.ArgumentParser()
    parser.add_argument("inputs", nargs="+")
    parser.add_argument("--out", default="outputs")
    parser.add_argument("--recursive", action="store_true")
    parser.add_argument("--jobs", default="auto")
    parser.add_argument("--continue-on-error", action="store_true")
    parser.add_argument("--clean", action="store_true")
    args = parser.parse_args(argv)

    bundle_dir = Path(__file__).resolve().parent
    os.chdir(bundle_dir)
    sys.path.insert(0, str(bundle_dir))
    load_env_file(bundle_dir / ".env.local")
    prepare_local_mounts(bundle_dir)
    config = json.loads((bundle_dir / "config.json").read_text(encoding="utf-8"))
    entrypoint_name = config.get("entrypoint", "transform")
    module = load_function(bundle_dir)
    entrypoint = getattr(module, entrypoint_name)

    out_dir = bundle_dir / args.out
    trace_dir = bundle_dir / "traces"
    if args.clean:
        shutil.rmtree(out_dir, ignore_errors=True)
        shutil.rmtree(trace_dir, ignore_errors=True)
    out_dir.mkdir(parents=True, exist_ok=True)
    trace_dir.mkdir(parents=True, exist_ok=True)
    input_paths = list(iter_inputs(args.inputs, args.recursive))
    if not input_paths:
        raise SystemExit("no input JSON files matched")
    max_workers = os.cpu_count() or 1 if args.jobs == "auto" else int(args.jobs)
    max_workers = max(1, min(max_workers, len(input_paths)))

    results = []
    failed = False
    with ThreadPoolExecutor(max_workers=max_workers) as pool:
        futures = {
            pool.submit(run_one, bundle_dir, entrypoint, path, out_dir, trace_dir): path
            for path in input_paths
        }
        for future in as_completed(futures):
            path = futures[future]
            try:
                result = future.result()
                print(json.dumps(result), flush=True)
                results.append(result)
            except Exception as exc:
                failed = True
                output_stem = output_stem_for(path)
                trace_path = trace_dir / output_stem.parent / f"{output_stem.name}.trace.json"
                trace_path.parent.mkdir(parents=True, exist_ok=True)
                trace_path.write_text(
                    json.dumps(
                        {
                            "input": str(path),
                            "ok": False,
                            "error": str(exc),
                            "traceback": traceback.format_exc(),
                        },
                        indent=2,
                    )
                    + "\n",
                    encoding="utf-8",
                )
                print(json.dumps({"input": str(path), "ok": False, "error": str(exc)}), flush=True)
                if not args.continue_on_error:
                    break
    if failed:
        raise SystemExit(1)


if __name__ == "__main__":
    main()
`

func init() {
	workflowsFunctionsHydrateCmd.Flags().Bool("force", false, "overwrite generated runtime files if they already exist")
	workflowsFunctionsHydrateCmd.Flags().Bool("fill-secrets", false, "fill .env.local from Retab secrets when the API supports secret value reads")
	workflowsFunctionsHydrateCmd.Flags().Bool("force-secrets", false, "overwrite existing .env.local secret values when used with --fill-secrets")
	workflowsFunctionsRunCmd.Flags().String("python", "python3", "Python interpreter to use")
	workflowsFunctionsRunCmd.Flags().String("out", "outputs", "output directory inside the bundle")
	workflowsFunctionsRunCmd.Flags().String("jobs", "auto", "parallel local executions: auto or a positive integer")
	workflowsFunctionsRunCmd.Flags().String("timeout", "0", "maximum wall-clock duration for the local run, e.g. 30s or 5m; 0 disables")
	workflowsFunctionsRunCmd.Flags().Bool("recursive", false, "recurse into input directories")
	workflowsFunctionsRunCmd.Flags().Bool("continue-on-error", false, "continue processing remaining samples after a sample fails")
	workflowsFunctionsRunCmd.Flags().Bool("clean", false, "remove the selected output directory and traces before running")

	workflowsFunctionsCmd.AddCommand(workflowsFunctionsHydrateCmd, workflowsFunctionsRunCmd)
	workflowsBlocksCmd.AddCommand(workflowsFunctionsCmd)
}
