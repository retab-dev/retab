//go:build !retab_oagen_cli_workflows

package cmd

import (
	"context"
	"encoding/json"
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

  retab workflows blocks config pull <workflow-id> <block-id> --out tmp/fn

Function bundles pulled with config pull are hydrated automatically. Re-run
hydrate when you need to repair or regenerate local support files:

  retab workflows blocks functions hydrate tmp/fn

Hydration writes input.py, output.py, .env.local placeholders, local table
fixture paths in mounts.json, and a run.py wrapper. Secrets are not downloaded
by default. If your local environment already defines the variables declared in
mounts.secrets, the local runner uses them without prompting. Python functions
hydrate as Python files; TypeScript functions hydrate as TypeScript files.`,
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
		sourceSchema, err := readFunctionInputSchemaSidecar(args[0])
		if err != nil {
			return err
		}
		if err := hydrateFunctionBundle(args[0], config, sourceSchema, force); err != nil {
			return err
		}
		filledSecrets := []map[string]any{}
		if fillSecrets {
			filledSecrets, err = fillLocalSecretsFromRetab(cmd, args[0], config, forceSecrets)
			if err != nil {
				return err
			}
		}
		files := []string{
			"input.py",
			"output.py",
			"models.py",
			"run.py",
			".env.example",
			".env.local",
			".retab/runtime.py",
		}
		if functionLanguage(config) == "typescript" {
			files = []string{
				"input_schema.json",
				"models.generated.ts",
				"schemas.generated.ts",
				"tsconfig.json",
				"run.mjs",
				".env.example",
				".env.local",
				".retab/runtime.mjs",
			}
		}
		return printJSON(map[string]any{
			"ok":             true,
			"dir":            args[0],
			"mode":           "repair",
			"workflow_id":    manifest.WorkflowID,
			"block_id":       manifest.BlockID,
			"language":       functionLanguage(config),
			"files":          files,
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
		manifest, config, err := readBlockConfigBundle(dir)
		if err != nil {
			return err
		}
		if manifest.BlockType != "function" {
			return fmt.Errorf("bundle block_type is %q, expected function", manifest.BlockType)
		}
		language := functionLanguage(config)
		runFile := "run.py"
		if language == "typescript" {
			runFile = "run.mjs"
		}
		runPath := filepath.Join(dir, runFile)
		if _, err := os.Stat(runPath); err != nil {
			return fmt.Errorf("%s not found; run `retab workflows blocks functions hydrate %s` first: %w", runFile, dir, err)
		}
		python, _ := cmd.Flags().GetString("python")
		if strings.TrimSpace(python) == "" {
			python = "python3"
		}
		node, _ := cmd.Flags().GetString("node")
		if strings.TrimSpace(node) == "" {
			node = "node"
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
		executable := python
		if language == "typescript" {
			executable = node
		}
		return runFunctionChild(cmd.Context(), functionChildOptions{
			Executable: executable,
			Args:       childArgs,
			Dir:        dir,
			Timeout:    timeout,
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
	return runFunctionChild(ctx, functionChildOptions{
		Executable: opts.Python,
		Args:       opts.Args,
		Dir:        opts.Dir,
		Timeout:    opts.Timeout,
	})
}

type functionChildOptions struct {
	Executable string
	Args       []string
	Dir        string
	Timeout    time.Duration
}

func runFunctionChild(ctx context.Context, opts functionChildOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	child := exec.CommandContext(ctx, opts.Executable, opts.Args...)
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
	// filepath.IsAbs misses POSIX-rooted paths like "/tmp/out" on Windows
	// (it wants a drive letter there), which would let a rooted path escape
	// the bundle. Reject a leading "/" explicitly: on Unix every such path is
	// already IsAbs (so this is a no-op there), and on Windows it closes the
	// gap.
	if filepath.IsAbs(raw) || strings.HasPrefix(raw, "/") {
		return fmt.Errorf("--out must be a relative path inside the function bundle")
	}
	cleaned := filepath.Clean(raw)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("--out must stay inside the function bundle")
	}
	return nil
}

func hydrateFunctionBundle(dir string, config map[string]any, sourceSchema map[string]any, force bool) error {
	if functionLanguage(config) == "typescript" {
		return hydrateTypescriptFunctionBundle(dir, config, sourceSchema, force)
	}
	return hydratePythonFunctionBundle(dir, config, sourceSchema, force)
}

func hydratePythonFunctionBundle(dir string, config map[string]any, sourceSchema map[string]any, force bool) error {
	files := map[string]string{
		"input.py":          generateFunctionInputModule(sourceSchema),
		"output.py":         generateFunctionOutputModule(config),
		"models.py":         generateFunctionModelsModule(config),
		"run.py":            generatedFunctionRunPy,
		".retab/runtime.py": generatedFunctionRuntimePy,
	}
	for rel, content := range files {
		if err := writeTextFileIfAllowed(filepath.Join(dir, rel), content, force, 0o700); err != nil {
			return err
		}
	}
	if force {
		for _, rel := range []string{"input_models.py", "output_models.py", "retab_runtime.py", "mounts.local.json", "models.generated.ts", "schemas.generated.ts", "tsconfig.json", "run.mjs", ".retab/runtime.mjs"} {
			if err := os.Remove(filepath.Join(dir, rel)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	// Persist the upstream schema so a later bare `functions hydrate <dir>` can
	// rebuild input.py with the same typing instead of silently degrading to a
	// permissive Input: hydrate reads it back via readFunctionInputSchemaSidecar,
	// and only `config pull` can fetch it from the server.
	if err := writeJSONFile(filepath.Join(dir, "input_schema.json"), functionInputSchema(sourceSchema)); err != nil {
		return err
	}
	return hydrateFunctionCommonFiles(dir, config, force)
}

func hydrateTypescriptFunctionBundle(dir string, config map[string]any, sourceSchema map[string]any, force bool) error {
	files := map[string]string{
		"models.generated.ts":  generateTypescriptModelModule(sourceSchema, config),
		"schemas.generated.ts": generateTypescriptSchemaModule(sourceSchema, config),
		"tsconfig.json":        generateTypescriptTsConfig(),
		"run.mjs":              generatedFunctionRunMjs,
		".retab/runtime.mjs":   generatedFunctionRuntimeMjs,
	}
	for rel, content := range files {
		if err := writeTextFileIfAllowed(filepath.Join(dir, rel), content, force, 0o700); err != nil {
			return err
		}
	}
	if force {
		for _, rel := range []string{"input.py", "output.py", "run.py", ".retab/runtime.py", "models.py", "input_models.py", "output_models.py", "retab_runtime.py", "mounts.local.json"} {
			if err := os.Remove(filepath.Join(dir, rel)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	if err := writeJSONFile(filepath.Join(dir, "input_schema.json"), typescriptInputSchema(sourceSchema)); err != nil {
		return err
	}
	return hydrateFunctionCommonFiles(dir, config, force)
}

func readFunctionInputSchemaSidecar(dir string) (map[string]any, error) {
	path := filepath.Join(dir, "input_schema.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	schema, err := readJSONMap(path)
	if err != nil {
		return nil, fmt.Errorf("read input_schema.json: %w", err)
	}
	return schema, nil
}

func hydrateFunctionCommonFiles(dir string, config map[string]any, force bool) error {
	// Key the generated .env files by each secret's `env` field — that is the
	// environment-variable name the running function reads, and the name the
	// --fill-secrets path (fillLocalSecretsFromRetab keys by secret.Env) writes.
	// Using the display `name` instead produced a .env keyed by the wrong
	// identifier whenever name != env, so the runtime saw no value and
	// --fill-secrets appended a duplicate line. Mirror the api_call hydrate.
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

// generatedModelPreamble mirrors the import block the engine puts above the same
// emitted models in BuildExecutionScriptWithEntrypointWithError (core-go
// function_block.go). The emitted code references these names directly, so the
// preamble has to stay in sync with that script.
const generatedModelPreamble = `from __future__ import annotations
import datetime
from typing import Any, Optional, Literal, Annotated
from pydantic import BaseModel, Field, ConfigDict

`

// functionInputSchema mirrors typescriptInputSchema for Python bundles: a
// function block with no upstream schema still gets a valid (permissive) object
// schema rather than a nil the emitter would render as an empty Input.
func functionInputSchema(sourceSchema map[string]any) map[string]any {
	if sourceSchema == nil {
		return map[string]any{"type": "object"}
	}
	return sourceSchema
}

// emittedPydanticClassNames returns the class names actually declared by
// generated model code, in declaration order.
//
// It reads the emitted source rather than re-deriving names from $defs because
// the emitter -- not the caller -- decides the final names: it sanitizes them
// and de-duplicates collisions with numeric suffixes. Re-deriving produced
// names that did not exist in output.py, so models.py raised ImportError on
// exactly the schemas whose names needed rewriting.
func emittedPydanticClassNames(code string) []string {
	names := []string{}
	for _, line := range strings.Split(code, "\n") {
		if !strings.HasPrefix(line, "class ") || !strings.Contains(line, "(BaseModel):") {
			continue
		}
		name := strings.TrimPrefix(line, "class ")
		name = strings.Split(name, "(BaseModel):")[0]
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

// functionOutputClassNames returns the model class names generateFunctionOutputModule
// emits into output.py, as declared by the emitter itself.
func functionOutputClassNames(config map[string]any) []string {
	return emittedPydanticClassNames(generateFunctionOutputModule(config))
}

// generateFunctionOutputModule writes output.py using the same emitter the
// workflow engine runs server-side (GenerateOutputModelCode, vendored into
// zz_json_schema_to_pydantic.go). It previously wrote permissive stubs
// (`class X(_RetabModel): pass`), which could not reproduce production typing:
// a local run accepted values the engine rejects, and vice versa.
func generateFunctionOutputModule(config map[string]any) string {
	return generatedModelPreamble + GenerateOutputModelCode(JSONObject(typescriptOutputSchema(config))) + "\n"
}

// generateFunctionModelsModule generates the models.py compatibility module that
// mirrors the server sandbox (workspace_builder.build_python_models_module): it
// re-exports Input from input.py and the output classes from output.py so that
// server-canonical function code (`from models import Input, Output`) runs
// unchanged under `functions run`. Without it the local runtime fails with
// "ModuleNotFoundError: No module named 'models'".
func generateFunctionModelsModule(config map[string]any) string {
	outputNames := functionOutputClassNames(config)
	allNames := append([]string{"Input"}, outputNames...)
	quoted := make([]string, len(allNames))
	for i, name := range allNames {
		quoted[i] = fmt.Sprintf("%q", name)
	}
	var b strings.Builder
	b.WriteString("from __future__ import annotations\n\n")
	b.WriteString("from input import Input\n")
	fmt.Fprintf(&b, "from output import %s\n", strings.Join(outputNames, ", "))
	fmt.Fprintf(&b, "\n__all__ = [%s]\n", strings.Join(quoted, ", "))
	return b.String()
}

// generateFunctionInputModule writes input.py from the upstream schema with the
// engine's emitter, so `Input.model_validate` locally rejects what it rejects in
// production (e.g. a fractional value for an integer field) instead of the old
// stub silently accepting anything.
func generateFunctionInputModule(sourceSchema map[string]any) string {
	return generatedModelPreamble + GenerateInputModelCode(JSONObject(functionInputSchema(sourceSchema))) + "\n"
}

func generateTypescriptModelModule(sourceSchema map[string]any, config map[string]any) string {
	inputSchema := typescriptInputSchema(sourceSchema)
	outputSchema := typescriptOutputSchema(config)
	defs := collectTypescriptDefs(inputSchema)
	for name, schema := range collectTypescriptDefs(outputSchema) {
		defs[name] = schema
	}
	defNames := make([]string, 0, len(defs))
	for name := range defs {
		defNames = append(defNames, name)
	}
	sort.Strings(defNames)

	chunks := []string{}
	for _, defName := range defNames {
		defSchema, _ := defs[defName].(map[string]any)
		if defSchema == nil {
			continue
		}
		chunks = append(chunks, fmt.Sprintf(
			"export type %s = %s;",
			sanitizeTypescriptTypeName(defName),
			typescriptSchemaType(defSchema, defName),
		))
	}
	chunks = append(chunks, fmt.Sprintf(
		"export type Input = %s;",
		typescriptSchemaType(inputSchema, "Input"),
	))
	chunks = append(chunks, fmt.Sprintf(
		"export type Output = %s;",
		typescriptSchemaType(outputSchema, "Output"),
	))
	return strings.Join(chunks, "\n\n") + "\n"
}

func generateTypescriptSchemaModule(sourceSchema map[string]any, config map[string]any) string {
	inputRaw, _ := json.MarshalIndent(typescriptInputSchema(sourceSchema), "", "  ")
	outputRaw, _ := json.MarshalIndent(typescriptOutputSchema(config), "", "  ")
	return fmt.Sprintf("export const inputSchema = %s as const;\n\nexport const outputSchema = %s as const;\n", inputRaw, outputRaw)
}

func generateTypescriptTsConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "noEmit": true
  },
  "include": [
    "function.ts",
    "models.generated.ts",
    "schemas.generated.ts"
  ]
}
`
}

func typescriptInputSchema(sourceSchema map[string]any) map[string]any {
	if sourceSchema == nil {
		return map[string]any{"type": "object"}
	}
	return sourceSchema
}

func typescriptOutputSchema(config map[string]any) map[string]any {
	if schema, ok := config["output_schema"].(map[string]any); ok {
		return schema
	}
	return map[string]any{"type": "object"}
}

func collectTypescriptDefs(schema map[string]any) map[string]any {
	defs := map[string]any{}
	for _, key := range []string{"$defs", "definitions"} {
		rawDefs, _ := schema[key].(map[string]any)
		for name, value := range rawDefs {
			defs[name] = value
		}
	}
	return defs
}

func typescriptSchemaType(schema map[string]any, fallbackName string) string {
	if ref, ok := schema["$ref"].(string); ok && strings.TrimSpace(ref) != "" {
		parts := strings.Split(ref, "/")
		return sanitizeTypescriptTypeName(parts[len(parts)-1])
	}
	if enumValues, ok := schema["enum"].([]any); ok && len(enumValues) > 0 {
		parts := make([]string, 0, len(enumValues))
		for _, value := range enumValues {
			parts = append(parts, typescriptLiteralType(value))
		}
		return strings.Join(parts, " | ")
	}
	for _, unionKey := range []string{"anyOf", "oneOf"} {
		if variants, ok := schema[unionKey].([]any); ok && len(variants) > 0 {
			parts := make([]string, 0, len(variants))
			for _, variant := range variants {
				if variantSchema, ok := variant.(map[string]any); ok {
					parts = append(parts, typescriptSchemaType(variantSchema, fallbackName))
				} else {
					parts = append(parts, "unknown")
				}
			}
			return strings.Join(parts, " | ")
		}
	}
	if typeList, ok := schema["type"].([]any); ok {
		parts := make([]string, 0, len(typeList))
		for _, rawType := range typeList {
			if schemaType, ok := rawType.(string); ok {
				cloned := cloneShallowJSONMap(schema)
				cloned["type"] = schemaType
				parts = append(parts, typescriptSchemaType(cloned, fallbackName))
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, " | ")
		}
	}
	schemaType, _ := schema["type"].(string)
	switch schemaType {
	case "string":
		return "string"
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	case "null":
		return "null"
	case "array":
		items, _ := schema["items"].(map[string]any)
		itemType := "unknown"
		if items != nil {
			itemType = typescriptSchemaType(items, fallbackName+"Item")
		}
		return fmt.Sprintf("Array<%s>", itemType)
	case "object":
		return typescriptObjectType(schema, fallbackName)
	default:
		if _, ok := schema["properties"].(map[string]any); ok {
			return typescriptObjectType(schema, fallbackName)
		}
		return "unknown"
	}
}

func typescriptObjectType(schema map[string]any, fallbackName string) string {
	properties, _ := schema["properties"].(map[string]any)
	if len(properties) == 0 {
		if additional, ok := schema["additionalProperties"].(map[string]any); ok {
			return fmt.Sprintf("Record<string, %s>", typescriptSchemaType(additional, fallbackName+"Value"))
		}
		return "Record<string, unknown>"
	}
	required := map[string]bool{}
	if rawRequired, ok := schema["required"].([]any); ok {
		for _, raw := range rawRequired {
			if name, ok := raw.(string); ok {
				required[name] = true
			}
		}
	}
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	lines := []string{"{"}
	for _, name := range names {
		propSchema, _ := properties[name].(map[string]any)
		propType := "unknown"
		if propSchema != nil {
			propType = typescriptSchemaType(propSchema, fallbackName+sanitizeTypescriptTypeName(name))
		}
		optional := "?"
		if required[name] {
			optional = ""
		}
		lines = append(lines, fmt.Sprintf("  %s%s: %s;", quoteTypescriptProperty(name), optional, propType))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func typescriptLiteralType(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case float64, int, int64:
		return fmt.Sprintf("%v", typed)
	case string:
		raw, _ := json.Marshal(typed)
		return string(raw)
	default:
		return "unknown"
	}
}

var typescriptIdentifierRegexp = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

func quoteTypescriptProperty(name string) string {
	if typescriptIdentifierRegexp.MatchString(name) {
		return name
	}
	raw, _ := json.Marshal(name)
	return string(raw)
}

func sanitizeTypescriptTypeName(raw string) string {
	identifier := sanitizeIdentifier(raw)
	if identifier == "" {
		return "Model"
	}
	parts := strings.Split(identifier, "_")
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	name := strings.Join(parts, "")
	if name == "" {
		return "Model"
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "_" + name
	}
	return name
}

func cloneShallowJSONMap(value map[string]any) map[string]any {
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
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

const generatedFunctionRunMjs = `import { main } from "./.retab/runtime.mjs";

await main();
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
    # Mirror the engine's Input.model_validate(_raw_input). Input(**payload)
    # bypasses field aliases, so a schema whose property names are not Python
    # identifiers would silently drop those values locally but populate them in
    # production.
    input_data = Input.model_validate(payload) if isinstance(payload, dict) else payload
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

    bundle_dir = Path(__file__).resolve().parent.parent
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

const generatedFunctionRuntimeMjs = `import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL, fileURLToPath } from "node:url";
import { cpus } from "node:os";

function loadEnvFile(filePath) {
  if (!fs.existsSync(filePath)) return;
  for (const rawLine of fs.readFileSync(filePath, "utf8").split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#") || !line.includes("=")) continue;
    const [key, ...rest] = line.split("=");
    if (!(key.trim() in process.env)) {
      process.env[key.trim()] = rest.join("=").trim();
    }
  }
}

function prepareLocalMounts(bundleDir) {
  const mountsPath = path.join(bundleDir, "mounts.json");
  if (!fs.existsSync(mountsPath)) return;
  const mounts = JSON.parse(fs.readFileSync(mountsPath, "utf8"));
  for (const table of mounts.tables ?? []) {
    if (!table.local_path || !table.path) continue;
    const localPath = path.join(bundleDir, table.local_path);
    if (!fs.existsSync(localPath)) continue;
    fs.mkdirSync(path.dirname(table.path), { recursive: true });
    fs.copyFileSync(localPath, table.path);
  }
}

function parseArgs(argv) {
  const args = { inputs: [], out: "outputs", recursive: false, jobs: "auto", continueOnError: false, clean: false };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--out") args.out = argv[++i];
    else if (arg === "--recursive") args.recursive = true;
    else if (arg === "--jobs") args.jobs = argv[++i];
    else if (arg === "--continue-on-error") args.continueOnError = true;
    else if (arg === "--clean") args.clean = true;
    else args.inputs.push(arg);
  }
  if (args.inputs.length === 0) {
    throw new Error("at least one input JSON file or directory is required");
  }
  return args;
}

function walkInputs(rawInputs, recursive) {
  const out = [];
  for (const raw of rawInputs) {
    const resolved = path.resolve(raw);
    const stat = fs.statSync(resolved);
    if (!stat.isDirectory()) {
      out.push(resolved);
      continue;
    }
    const stack = [resolved];
    while (stack.length > 0) {
      const current = stack.pop();
      for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
        const child = path.join(current, entry.name);
        if (entry.isDirectory() && recursive) stack.push(child);
        else if (entry.isFile() && entry.name.endsWith(".json")) out.push(child);
      }
    }
  }
  return out.sort();
}

function outputStem(inputPath) {
  // Mirror the Python runtime: keep the input's directory structure relative
  // to the bundle dir (the cwd) so same-named inputs in different folders
  // don't collide on outputs/<name>.out.json. Inputs outside the bundle fall
  // back to the bare basename.
  const resolved = path.resolve(inputPath);
  const rel = path.relative(process.cwd(), resolved);
  const parsed = (rel === "" || rel.startsWith("..") || path.isAbsolute(rel))
    ? path.parse(path.basename(resolved))
    : path.parse(rel);
  return path.join(parsed.dir, parsed.name);
}

async function runOne({ bundleDir, transform, inputPath, outDir, traceDir }) {
  const payload = JSON.parse(fs.readFileSync(inputPath, "utf8"));
  const output = await transform(payload);
  const stem = outputStem(inputPath);
  const outPath = path.join(outDir, stem + ".out.json");
  fs.mkdirSync(path.dirname(outPath), { recursive: true });
  fs.writeFileSync(outPath, JSON.stringify(output, null, 2) + "\n", "utf8");
  const tracePath = path.join(traceDir, stem + ".trace.json");
  fs.mkdirSync(path.dirname(tracePath), { recursive: true });
  fs.writeFileSync(tracePath, JSON.stringify({ input: inputPath, output: outPath, ok: true }, null, 2) + "\n", "utf8");
  return { input: inputPath, output: outPath, ok: true };
}

export async function main(argv = process.argv.slice(2)) {
  const runtimeDir = path.dirname(fileURLToPath(import.meta.url));
  const bundleDir = path.dirname(runtimeDir);
  process.chdir(bundleDir);
  loadEnvFile(path.join(bundleDir, ".env.local"));
  prepareLocalMounts(bundleDir);
  const config = JSON.parse(fs.readFileSync(path.join(bundleDir, "config.json"), "utf8"));
  const entrypoint = config.entrypoint ?? "transform";
  let module;
  try {
    module = await import(pathToFileURL(path.join(bundleDir, "function.ts")).href);
  } catch (error) {
    throw new Error("could not load function.ts with local Node.js. Use Node.js with TypeScript execution support, or run the bundle with a TypeScript runner. " + error.message);
  }
  const transform = module[entrypoint];
  if (typeof transform !== "function") {
    throw new Error("function.ts must export " + entrypoint + "()");
  }
  const args = parseArgs(argv);
  const outDir = path.join(bundleDir, args.out);
  const traceDir = path.join(bundleDir, "traces");
  if (args.clean) {
    fs.rmSync(outDir, { recursive: true, force: true });
    fs.rmSync(traceDir, { recursive: true, force: true });
  }
  fs.mkdirSync(outDir, { recursive: true });
  fs.mkdirSync(traceDir, { recursive: true });
  const inputPaths = walkInputs(args.inputs, args.recursive);
  if (inputPaths.length === 0) throw new Error("no input JSON files matched");
  const jobs = args.jobs === "auto" ? Math.max(1, cpus().length) : Math.max(1, Number.parseInt(args.jobs, 10));
  const queue = [...inputPaths];
  let failed = false;
  async function worker() {
    while (queue.length > 0 && (!failed || args.continueOnError)) {
      const inputPath = queue.shift();
      try {
        const result = await runOne({ bundleDir, transform, inputPath, outDir, traceDir });
        console.log(JSON.stringify(result));
      } catch (error) {
        failed = true;
        const tracePath = path.join(traceDir, outputStem(inputPath) + ".trace.json");
        fs.mkdirSync(path.dirname(tracePath), { recursive: true });
        fs.writeFileSync(tracePath, JSON.stringify({ input: inputPath, ok: false, error: error.message }, null, 2) + "\n", "utf8");
        console.log(JSON.stringify({ input: inputPath, ok: false, error: error.message }));
      }
    }
  }
  await Promise.all(Array.from({ length: Math.min(jobs, inputPaths.length) }, () => worker()));
  if (failed) process.exitCode = 1;
}
`

func init() {
	workflowsFunctionsHydrateCmd.Flags().Bool("force", false, "overwrite generated runtime files if they already exist")
	workflowsFunctionsHydrateCmd.Flags().Bool("fill-secrets", false, "fill .env.local from Retab secrets when the API supports secret value reads")
	workflowsFunctionsHydrateCmd.Flags().Bool("force-secrets", false, "overwrite existing .env.local secret values when used with --fill-secrets")
	workflowsFunctionsRunCmd.Flags().String("python", "python3", "Python interpreter to use")
	workflowsFunctionsRunCmd.Flags().String("node", "node", "Node.js executable to use for TypeScript functions")
	workflowsFunctionsRunCmd.Flags().String("out", "outputs", "output directory inside the bundle")
	workflowsFunctionsRunCmd.Flags().String("jobs", "auto", "parallel local executions: auto or a positive integer")
	workflowsFunctionsRunCmd.Flags().String("timeout", "0", "maximum wall-clock duration for the local run, e.g. 30s or 5m; 0 disables")
	workflowsFunctionsRunCmd.Flags().Bool("recursive", false, "recurse into input directories")
	workflowsFunctionsRunCmd.Flags().Bool("continue-on-error", false, "continue processing remaining samples after a sample fails")
	workflowsFunctionsRunCmd.Flags().Bool("clean", false, "remove the selected output directory and traces before running")

	workflowsFunctionsCmd.AddCommand(workflowsFunctionsHydrateCmd, workflowsFunctionsRunCmd)
	workflowsBlocksCmd.AddCommand(workflowsFunctionsCmd)
}
