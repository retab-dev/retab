//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
)

func TestWorkflowsBlocksPullConfigWritesEditableBundle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	var gotPath, gotQuery, gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_cfg",
			"workflow_id": "wf_cfg",
			"type":        "extract",
			"label":       "Extract",
			"updated_at":  "2026-06-03T10:00:00Z",
			"config": map[string]any{
				"model":  "retab-small",
				"prompt": "Extract fields.",
				"json_schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"order_id": map[string]any{"type": "string"},
					},
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
		if err := workflowsBlocksPullConfigCmd.RunE(workflowsBlocksPullConfigCmd, []string{"wf_cfg", "blk_cfg"}); err != nil {
			t.Fatalf("pull-config: %v", err)
		}
	})

	if gotMethod != http.MethodGet {
		t.Fatalf("method = %s, want GET", gotMethod)
	}
	if gotPath != "/v1/workflows/blocks/blk_cfg" {
		t.Fatalf("path = %s, want block get", gotPath)
	}
	if !strings.Contains(gotQuery, "workflow_id=wf_cfg") {
		t.Fatalf("query = %s, want workflow_id=wf_cfg", gotQuery)
	}
	if !strings.Contains(stdout, `"block_id": "blk_cfg"`) {
		t.Fatalf("stdout should include block id, got:\n%s", stdout)
	}

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if manifest.BlockID != "blk_cfg" || manifest.WorkflowID != "wf_cfg" || manifest.Adapter != "json_schema" {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	var rawManifest map[string]any
	if err := readJSONFileStrict(filepath.Join(dir, "manifest.json"), &rawManifest); err != nil {
		t.Fatalf("raw manifest: %v", err)
	}
	if _, ok := rawManifest["workflow_id"]; ok {
		t.Fatalf("manifest should not use flat workflow_id: %+v", rawManifest)
	}
	if _, ok := rawManifest["api_version"]; ok {
		t.Fatalf("manifest should not use api_version: %+v", rawManifest)
	}
	if _, ok := rawManifest["target"].(map[string]any); !ok {
		t.Fatalf("manifest should include target object: %+v", rawManifest)
	}
	adapter, ok := rawManifest["adapter"].(map[string]any)
	if !ok {
		t.Fatalf("manifest should include adapter object: %+v", rawManifest)
	}
	if _, ok := adapter["version"]; ok {
		t.Fatalf("manifest adapter should not use version: %+v", adapter)
	}
	files, ok := rawManifest["files"].([]any)
	if !ok || len(files) != 2 {
		t.Fatalf("manifest should include file descriptors: %+v", rawManifest)
	}
	for _, rawFile := range files {
		file, ok := rawFile.(map[string]any)
		if !ok {
			t.Fatalf("manifest file descriptor should be an object: %+v", rawFile)
		}
		if file["format"] != "json" {
			t.Fatalf("json sidecars should use public json format, got %+v", file)
		}
	}

	config, err := readJSONMap(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("config.json: %v", err)
	}
	if _, ok := config["json_schema"]; ok {
		t.Fatalf("config.json should split json_schema into json_schema.json, got %+v", config)
	}
	if _, err := os.Stat(filepath.Join(dir, "json_schema.json")); err != nil {
		t.Fatalf("json_schema.json should exist: %v", err)
	}
}

func TestWorkflowsBlocksPullConfigAutoHydratesFunctionRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "block_fn",
			"workflow_id": "wf_cfg",
			"type":        "function",
			"label":       "Function",
			"updated_at":  "2026-06-03T10:00:00Z",
			"config": map[string]any{
				"entrypoint": "transform",
				"code":       "from output import Output\n\ndef transform(input):\n    return Output(ok=True)\n",
				"output_schema": map[string]any{
					"type":       "object",
					"properties": map[string]any{"ok": map[string]any{"type": "boolean"}},
				},
				"mounts": map[string]any{
					"secrets": []any{map[string]any{"env": "API_TOKEN"}},
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
		if err := workflowsBlocksPullConfigCmd.RunE(workflowsBlocksPullConfigCmd, []string{"wf_cfg", "block_fn"}); err != nil {
			t.Fatalf("pull-config: %v", err)
		}
	})
	if !strings.Contains(stdout, `"runtime_hydrated": true`) {
		t.Fatalf("function pull-config should report runtime_hydrated=true, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"runtime_adapter": "function"`) {
		t.Fatalf("function pull-config should report runtime_adapter=function, got:\n%s", stdout)
	}
	for _, rel := range []string{"function.py", "output_schema.json", "mounts.json"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected %s to be written: %v", rel, err)
		}
	}
	for _, rel := range []string{"input.py", "output.py", "run.py", ".retab/runtime.py", ".env.example", ".env.local"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected runtime support file %s to exist: %v", rel, err)
		}
	}
	for _, rel := range []string{"models.py", "input_models.py", "output_models.py", "retab_runtime.py", "mounts.local.json"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(err) {
			t.Fatalf("obsolete support file %s should not exist, stat err=%v", rel, err)
		}
	}
	var manifestJSON blockConfigBundleManifestJSON
	if err := readJSONFileStrict(filepath.Join(dir, "manifest.json"), &manifestJSON); err != nil {
		t.Fatalf("function manifest: %v", err)
	}
	formats := map[string]string{}
	for _, file := range manifestJSON.Files {
		formats[file.Role] = file.Format
	}
	if formats["function"] != "python" {
		t.Fatalf("function file format = %s, want python", formats["function"])
	}
	for _, role := range []string{"config", "output_schema", "mounts"} {
		if formats[role] != "json" {
			t.Fatalf("%s file format = %s, want json", role, formats[role])
		}
	}
}

func TestWorkflowsBlocksPullConfigAutoHydratesTypescriptFunctionRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "block_fn_ts",
			"workflow_id": "wf_cfg",
			"type":        "function",
			"label":       "TypeScript Function",
			"updated_at":  "2026-06-03T10:00:00Z",
			"config": map[string]any{
				"language": "typescript",
				"code":     "import type { Input, Output } from \"./models.generated\";\n\nexport function transform(input: Input): Output {\n  return { ok: true };\n}\n",
				"output_schema": map[string]any{
					"type":       "object",
					"properties": map[string]any{"ok": map[string]any{"type": "boolean"}},
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
		if err := workflowsBlocksPullConfigCmd.RunE(workflowsBlocksPullConfigCmd, []string{"wf_cfg", "block_fn_ts"}); err != nil {
			t.Fatalf("pull-config: %v", err)
		}
	})
	if !strings.Contains(stdout, `"runtime_hydrated": true`) {
		t.Fatalf("function pull-config should report runtime_hydrated=true, got:\n%s", stdout)
	}
	for _, rel := range []string{"function.ts", "output_schema.json", "models.generated.ts", "schemas.generated.ts", "run.mjs", ".retab/runtime.mjs", ".env.example", ".env.local"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected %s to be written: %v", rel, err)
		}
	}
	for _, rel := range []string{"function.py", "input.py", "output.py", "run.py", ".retab/runtime.py"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(err) {
			t.Fatalf("python support file %s should not exist for TypeScript bundle, stat err=%v", rel, err)
		}
	}

	var manifestJSON blockConfigBundleManifestJSON
	if err := readJSONFileStrict(filepath.Join(dir, "manifest.json"), &manifestJSON); err != nil {
		t.Fatalf("function manifest: %v", err)
	}
	paths := map[string]string{}
	formats := map[string]string{}
	for _, file := range manifestJSON.Files {
		paths[file.Role] = file.Path
		formats[file.Role] = file.Format
	}
	if paths["function"] != "function.ts" {
		t.Fatalf("function file path = %s, want function.ts", paths["function"])
	}
	if formats["function"] != "typescript" {
		t.Fatalf("function file format = %s, want typescript", formats["function"])
	}
	manifest, config, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if err := validateBlockConfigBundle(manifest, config); err != nil {
		t.Fatalf("validate bundle: %v", err)
	}
	if got := config["code"].(string); !strings.Contains(got, "export function transform") {
		t.Fatalf("reassembled code should come from function.ts, got:\n%s", got)
	}
	stdout, _ = captureStd(t, func() {
		if err := workflowsBlocksDoctorConfigCmd.RunE(workflowsBlocksDoctorConfigCmd, []string{dir}); err != nil {
			t.Fatalf("doctor-config: %v", err)
		}
	})
	if !strings.Contains(stdout, `"ok": true`) {
		t.Fatalf("doctor-config should accept freshly pulled TypeScript bundle, got:\n%s", stdout)
	}
}

func TestBlockConfigBundleRoundTripsSmokeMatrix(t *testing.T) {
	matrix := []struct {
		blockType string
		config    map[string]any
	}{
		{blockType: "start_document", config: map[string]any{}},
		{blockType: "start_json", config: map[string]any{"json_schema": map[string]any{"type": "object"}}},
		{blockType: "note", config: map[string]any{"text": "Remember this"}},
		{blockType: "parse", config: map[string]any{"model": "retab-small"}},
		{blockType: "edit", config: map[string]any{"instructions": "Fill the form"}},
		{blockType: "extract", config: map[string]any{"json_schema": map[string]any{"type": "object"}}},
		{blockType: "split", config: map[string]any{"subdocuments": []any{map[string]any{"name": "Invoice"}}}},
		{blockType: "classifier", config: map[string]any{"categories": []any{map[string]any{"name": "Invoice"}}}},
		{blockType: "conditional", config: map[string]any{"conditions": []any{}}},
		{blockType: "api_call", config: map[string]any{
			"request_schema": map[string]any{"type": "object"},
			"output_schema":  map[string]any{"type": "object"},
			"mounts":         map[string]any{},
		}},
		{blockType: "function", config: map[string]any{
			"code":          "def transform(input):\n    return {}\n",
			"output_schema": map[string]any{"type": "object"},
			"mounts":        map[string]any{},
		}},
		{blockType: "while_loop", config: map[string]any{"max_iterations": float64(3)}},
		{blockType: "for_each", config: map[string]any{"map_method": "split_by_key"}},
		{blockType: "merge_dicts", config: map[string]any{}},
	}

	for _, item := range matrix {
		t.Run(item.blockType, func(t *testing.T) {
			dir := t.TempDir()
			block := retab.WorkflowBlock{
				ID:         "blk_" + strings.ReplaceAll(item.blockType, "_", "-"),
				WorkflowID: "wf_cfg",
				Type:       retab.WorkflowBlockType(item.blockType),
				Config:     item.config,
				UpdatedAt:  time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
			}
			if err := writeBlockConfigBundle(dir, block, false); err != nil {
				t.Fatalf("write bundle: %v", err)
			}
			manifest, reassembled, err := readBlockConfigBundle(dir)
			if err != nil {
				t.Fatalf("read bundle: %v", err)
			}
			if manifest.BlockType != item.blockType {
				t.Fatalf("manifest block type = %s, want %s", manifest.BlockType, item.blockType)
			}
			if hashJSONMap(reassembled) != hashJSONMap(item.config) {
				t.Fatalf("reassembled config hash mismatch: got %s want %s", hashJSONMap(reassembled), hashJSONMap(item.config))
			}
		})
	}
}

func TestSplitBlockConfigBundleSplitsSubdocumentsAndReassembles(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"model":       "retab-small",
		"n_consensus": float64(2),
		"subdocuments": []any{
			map[string]any{"name": "Invoice", "handle_key": "invoice", "description": "Invoice pages"},
			map[string]any{"name": "Terms", "handle_key": "terms", "description": "Legal pages"},
		},
	}
	block := retab.WorkflowBlock{
		ID:         "blk_split",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeSplit,
		Config:     config,
		UpdatedAt:  time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatalf("write bundle: %v", err)
	}

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if manifest.Adapter != "split" || manifest.Files["subdocuments"] != "subdocuments.json" {
		t.Fatalf("unexpected split manifest: %+v", manifest)
	}
	baseConfig, err := readJSONMap(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("config.json: %v", err)
	}
	if _, ok := baseConfig["subdocuments"]; ok {
		t.Fatalf("config.json should not duplicate split subdocuments: %+v", baseConfig)
	}
	if _, err := os.Stat(filepath.Join(dir, "subdocuments.json")); err != nil {
		t.Fatalf("subdocuments.json should exist: %v", err)
	}

	_, reassembled, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if hashJSONMap(reassembled) != hashJSONMap(config) {
		t.Fatalf("reassembled config hash mismatch: got %s want %s", hashJSONMap(reassembled), hashJSONMap(config))
	}
}

func TestClassifierBlockConfigBundleSplitsCategoriesAndReassembles(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"model":         "retab-small",
		"first_n_pages": float64(3),
		"categories": []any{
			map[string]any{"name": "Invoice", "handle_key": "invoice", "description": "Invoice documents"},
			map[string]any{"name": "Receipt", "handle_key": "receipt", "description": "Payment receipts"},
		},
	}
	block := retab.WorkflowBlock{
		ID:         "blk_classifier",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeClassifier,
		Config:     config,
		UpdatedAt:  time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatalf("write bundle: %v", err)
	}

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if manifest.Adapter != "classifier" || manifest.Files["categories"] != "categories.json" {
		t.Fatalf("unexpected classifier manifest: %+v", manifest)
	}
	baseConfig, err := readJSONMap(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("config.json: %v", err)
	}
	if _, ok := baseConfig["categories"]; ok {
		t.Fatalf("config.json should not duplicate classifier categories: %+v", baseConfig)
	}
	if _, err := os.Stat(filepath.Join(dir, "categories.json")); err != nil {
		t.Fatalf("categories.json should exist: %v", err)
	}

	_, reassembled, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if hashJSONMap(reassembled) != hashJSONMap(config) {
		t.Fatalf("reassembled config hash mismatch: got %s want %s", hashJSONMap(reassembled), hashJSONMap(config))
	}
}

func TestAPICallBlockConfigBundleSplitsSchemasAndReassembles(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"method": "POST",
		"url":    "https://example.com/orders",
		"request_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"order_id": map[string]any{"type": "string"},
			},
		},
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{"type": "string"},
			},
		},
	}
	block := retab.WorkflowBlock{
		ID:         "blk_api_call",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeAPICall,
		Config:     config,
		UpdatedAt:  time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatalf("write bundle: %v", err)
	}

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if manifest.Adapter != "api_call" {
		t.Fatalf("unexpected api_call adapter: %+v", manifest)
	}
	if manifest.Files["request_schema"] != "request_schema.json" || manifest.Files["output_schema"] != "output_schema.json" {
		t.Fatalf("unexpected api_call manifest files: %+v", manifest.Files)
	}
	baseConfig, err := readJSONMap(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("config.json: %v", err)
	}
	if _, ok := baseConfig["request_schema"]; ok {
		t.Fatalf("config.json should not duplicate api_call request_schema: %+v", baseConfig)
	}
	if _, ok := baseConfig["output_schema"]; ok {
		t.Fatalf("config.json should not duplicate api_call output_schema: %+v", baseConfig)
	}
	if _, err := os.Stat(filepath.Join(dir, "request_schema.json")); err != nil {
		t.Fatalf("request_schema.json should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "output_schema.json")); err != nil {
		t.Fatalf("output_schema.json should exist: %v", err)
	}

	_, reassembled, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if hashJSONMap(reassembled) != hashJSONMap(config) {
		t.Fatalf("reassembled config hash mismatch: got %s want %s", hashJSONMap(reassembled), hashJSONMap(config))
	}
}

func TestReadBlockConfigBundleRejectsVersionedManifest(t *testing.T) {
	dir := t.TempDir()
	if err := writeJSONFile(filepath.Join(dir, "config.json"), map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONFile(filepath.Join(dir, "manifest.json"), map[string]any{
		"api_version": "retab.dev/workflow-block-config/v1beta1",
		"target": map[string]any{
			"kind":        "workflow_block",
			"workflow_id": "wf_cfg",
			"block_id":    "blk_cfg",
			"block_type":  "extract",
		},
		"adapter":  map[string]any{"name": "json_schema"},
		"baseline": map[string]any{"config_hash": hashJSONMap(map[string]any{})},
		"files":    []any{map[string]any{"role": "config", "path": "config.json", "format": "json"}},
	}); err != nil {
		t.Fatal(err)
	}

	_, _, bundleErr := readBlockConfigBundle(dir)
	if bundleErr == nil {
		t.Fatal("expected versioned manifest to be rejected")
	}
	if !strings.Contains(bundleErr.Error(), "unknown field") || !strings.Contains(bundleErr.Error(), "api_version") {
		t.Fatalf("unexpected error: %v", bundleErr)
	}
}

func TestReadBlockConfigBundleRejectsEscapingManifestPath(t *testing.T) {
	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest.Files["config"] = "../outside.json"
	if err := writeBlockConfigManifest(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}

	_, _, bundleErr := readBlockConfigBundle(dir)
	if bundleErr == nil {
		t.Fatal("expected escaping manifest path to be rejected")
	}
	if !strings.Contains(bundleErr.Error(), "files.config") || !strings.Contains(bundleErr.Error(), "inside the bundle") {
		t.Fatalf("unexpected error: %v", bundleErr)
	}
}

func TestWorkflowsBlocksDoctorConfigReportsMissingRuntimeFile(t *testing.T) {
	dir := t.TempDir()
	block := retab.WorkflowBlock{
		ID:         "blk_api",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeAPICall,
		Config: map[string]any{
			"method": "POST",
			"url":    "https://example.com",
		},
		UpdatedAt: time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatal(err)
	}
	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksDoctorConfigCmd.RunE(workflowsBlocksDoctorConfigCmd, []string{dir}); err != nil {
			t.Fatalf("doctor-config: %v", err)
		}
	})
	if !strings.Contains(stdout, `"ok": false`) {
		t.Fatalf("doctor-config should report problems, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"path": "run.sh"`) {
		t.Fatalf("doctor-config should report missing run.sh, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "retab workflows blocks api-calls hydrate") {
		t.Fatalf("doctor-config should include hydrate fix, got:\n%s", stdout)
	}
}

func TestReadBlockConfigBundleRejectsUnknownManifestFileRole(t *testing.T) {
	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest.Files["typo"] = "typo.json"
	if err := writeJSONFile(filepath.Join(dir, "typo.json"), map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if err := writeBlockConfigManifest(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}

	_, _, bundleErr := readBlockConfigBundle(dir)
	if bundleErr == nil {
		t.Fatal("expected unknown file role to be rejected")
	}
	if !strings.Contains(bundleErr.Error(), "files.typo") || !strings.Contains(bundleErr.Error(), "not a supported") {
		t.Fatalf("unexpected error: %v", bundleErr)
	}
}

func TestValidateBlockConfigBundleRejectsFileRoleForWrongBlockType(t *testing.T) {
	dir := t.TempDir()
	block := retab.WorkflowBlock{
		ID:         "blk_api_call",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeAPICall,
		Config: map[string]any{
			"method": "POST",
			"url":    "https://example.com/orders",
		},
		UpdatedAt: time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatal(err)
	}

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest.Files["subdocuments"] = "subdocuments.json"
	if err := writeJSONFile(filepath.Join(dir, "subdocuments.json"), []any{map[string]any{"name": "Bad"}}); err != nil {
		t.Fatal(err)
	}
	if err := writeBlockConfigManifest(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}

	manifest, config, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	err = validateBlockConfigBundle(manifest, config)
	if err == nil {
		t.Fatal("expected wrong file role for block type to be rejected")
	}
	if !strings.Contains(err.Error(), "files.subdocuments") || !strings.Contains(err.Error(), "api_call") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateBlockConfigBundleRejectsWrongAdapter(t *testing.T) {
	dir := t.TempDir()
	block := retab.WorkflowBlock{
		ID:         "blk_api_call",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeAPICall,
		Config: map[string]any{
			"method": "POST",
			"url":    "https://example.com/orders",
		},
		UpdatedAt: time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatal(err)
	}

	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest.Adapter = "generic"
	if err := writeBlockConfigManifest(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}

	manifest, config, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	err = validateBlockConfigBundle(manifest, config)
	if err == nil {
		t.Fatal("expected wrong adapter to be rejected")
	}
	if !strings.Contains(err.Error(), "manifest adapter") || !strings.Contains(err.Error(), "api_call") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadBlockConfigBundleRejectsNonArraySplitSidecar(t *testing.T) {
	dir := t.TempDir()
	block := retab.WorkflowBlock{
		ID:         "blk_split",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeSplit,
		Config: map[string]any{
			"model":        "retab-small",
			"subdocuments": []any{map[string]any{"name": "Invoice"}},
		},
		UpdatedAt: time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONFile(filepath.Join(dir, "subdocuments.json"), map[string]any{"name": "Invoice"}); err != nil {
		t.Fatal(err)
	}

	_, _, err := readBlockConfigBundle(dir)
	if err == nil {
		t.Fatal("expected non-array subdocuments sidecar to be rejected")
	}
	if !strings.Contains(err.Error(), "subdocuments.json") || !strings.Contains(err.Error(), "expected JSON array") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadBlockConfigBundleRejectsNonArrayClassifierSidecar(t *testing.T) {
	dir := t.TempDir()
	block := retab.WorkflowBlock{
		ID:         "blk_classifier",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeClassifier,
		Config: map[string]any{
			"model":      "retab-small",
			"categories": []any{map[string]any{"name": "Invoice"}},
		},
		UpdatedAt: time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, false); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONFile(filepath.Join(dir, "categories.json"), map[string]any{"name": "Invoice"}); err != nil {
		t.Fatal(err)
	}

	_, _, bundleErr := readBlockConfigBundle(dir)
	if bundleErr == nil {
		t.Fatal("expected non-array categories sidecar to be rejected")
	}
	if !strings.Contains(bundleErr.Error(), "categories.json") || !strings.Contains(bundleErr.Error(), "expected JSON array") {
		t.Fatalf("unexpected error: %v", bundleErr)
	}
}

func TestWorkflowsBlocksValidateConfigRejectsWrongBlock(t *testing.T) {
	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})

	if err := workflowsBlocksValidateConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workflowsBlocksValidateConfigCmd.Flags().Set("dir", "") })

	err := workflowsBlocksValidateConfigCmd.RunE(workflowsBlocksValidateConfigCmd, []string{"blk_other"})
	if err == nil {
		t.Fatal("expected wrong block id to be rejected")
	}
	if !strings.Contains(err.Error(), "belongs to block blk_cfg") {
		t.Fatalf("error should name manifest block, got %q", err.Error())
	}
}

func TestWorkflowsBlocksValidateConfigReportsOfflineMode(t *testing.T) {
	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})

	if err := workflowsBlocksValidateConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksValidateConfigCmd.Flags().Set("offline", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksValidateConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksValidateConfigCmd.Flags().Set("offline", "false")
	})

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksValidateConfigCmd.RunE(workflowsBlocksValidateConfigCmd, []string{"blk_cfg"}); err != nil {
			t.Fatalf("validate-config: %v", err)
		}
	})

	if !strings.Contains(stdout, `"mode": "offline"`) {
		t.Fatalf("validate output should report offline mode, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"authoritative": false`) {
		t.Fatalf("validate output should report authoritative=false, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"adapter": "json_schema"`) {
		t.Fatalf("validate output should report adapter, got:\n%s", stdout)
	}
}

func TestWorkflowsBlocksValidateConfigDefaultsToRemoteDryRun(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "local"})

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/workflows/blocks/blk_cfg/validate-config" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "workflow_id=wf_cfg") {
			t.Fatalf("query = %s, want workflow_id=wf_cfg", r.URL.RawQuery)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          true,
			"workflow_id": "wf_cfg",
			"block_id":    "blk_cfg",
			"block_type":  "extract",
			"config_hash": "server-hash",
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksValidateConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksValidateConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksValidateConfigCmd.Flags().Set("offline", "false")
	})

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksValidateConfigCmd.RunE(workflowsBlocksValidateConfigCmd, []string{"blk_cfg"}); err != nil {
			t.Fatalf("validate-config: %v", err)
		}
	})

	if gotBody["config_mode"] != "replace" {
		t.Fatalf("config_mode = %#v, want replace", gotBody["config_mode"])
	}
	cfg, ok := gotBody["config"].(map[string]any)
	if !ok || cfg["prompt"] != "local" {
		t.Fatalf("config = %#v, want local prompt", gotBody["config"])
	}
	if !strings.Contains(stdout, `"mode": "remote"`) {
		t.Fatalf("validate output should report remote mode, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"authoritative": true`) {
		t.Fatalf("validate output should report authoritative=true, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"config_hash": "server-hash"`) {
		t.Fatalf("validate output should use server hash, got:\n%s", stdout)
	}
}

func TestWorkflowsBlocksValidateConfigHelpDescribesRemoteAndOfflineModes(t *testing.T) {
	if !strings.Contains(workflowsBlocksValidateConfigCmd.Long, "backend to dry-run") {
		t.Fatalf("validate-config long help should describe remote validation:\n%s", workflowsBlocksValidateConfigCmd.Long)
	}
	if !strings.Contains(workflowsBlocksValidateConfigCmd.Long, "--offline") {
		t.Fatalf("validate-config long help should describe offline validation:\n%s", workflowsBlocksValidateConfigCmd.Long)
	}
}

func TestWorkflowsBlocksPushConfigUsesReplaceAndPreservesDriftCheck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})
	if err := writeJSONFile(filepath.Join(dir, "config.json"), map[string]any{"prompt": "local edit"}); err != nil {
		t.Fatal(err)
	}

	var patchCalled bool
	var gotPatchBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/v1/workflows/blocks/blk_cfg" || !strings.Contains(r.URL.RawQuery, "workflow_id=wf_cfg") {
				t.Fatalf("unexpected GET %s?%s", r.URL.Path, r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_cfg",
				"workflow_id": "wf_cfg",
				"type":        "extract",
				"updated_at":  "2026-06-03T10:00:00Z",
				"config":      map[string]any{"prompt": "original"},
			})
		case http.MethodPatch:
			patchCalled = true
			if r.URL.Path != "/v1/workflows/blocks/blk_cfg" || !strings.Contains(r.URL.RawQuery, "workflow_id=wf_cfg") {
				t.Fatalf("unexpected PATCH %s?%s", r.URL.Path, r.URL.RawQuery)
			}
			if err := json.NewDecoder(r.Body).Decode(&gotPatchBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_cfg",
				"workflow_id": "wf_cfg",
				"type":        "extract",
				"updated_at":  "2026-06-03T10:01:00Z",
				"config":      map[string]any{"prompt": "server canonical"},
			})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksPushConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPushConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("force", "false")
	})

	captureStd(t, func() {
		if err := workflowsBlocksPushConfigCmd.RunE(workflowsBlocksPushConfigCmd, []string{"blk_cfg"}); err != nil {
			t.Fatalf("push-config: %v", err)
		}
	})

	if !patchCalled {
		t.Fatal("expected PATCH to be called")
	}
	if gotPatchBody["config_mode"] != "replace" {
		t.Fatalf("config_mode = %#v, want replace", gotPatchBody["config_mode"])
	}
	cfg, ok := gotPatchBody["config"].(map[string]any)
	if !ok || cfg["prompt"] != "local edit" {
		t.Fatalf("PATCH config = %#v, want local edit", gotPatchBody["config"])
	}

	refreshed, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("refreshed manifest: %v", err)
	}
	if refreshed.RemoteHash != hashJSONMap(map[string]any{"prompt": "server canonical"}) {
		t.Fatalf("manifest remote hash was not refreshed, got %s", refreshed.RemoteHash)
	}
	refreshedConfig, err := readJSONMap(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("refreshed config: %v", err)
	}
	if refreshedConfig["prompt"] != "server canonical" {
		t.Fatalf("config.json was not refreshed from server response: %#v", refreshedConfig)
	}
}

func TestWorkflowsBlocksPushConfigSendsEmptyConfigForReplace(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{})

	var gotPatchBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_cfg",
				"workflow_id": "wf_cfg",
				"type":        "extract",
				"updated_at":  "2026-06-03T10:00:00Z",
				"config":      map[string]any{},
			})
		case http.MethodPatch:
			if err := json.NewDecoder(r.Body).Decode(&gotPatchBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_cfg",
				"workflow_id": "wf_cfg",
				"type":        "extract",
				"updated_at":  "2026-06-03T10:01:00Z",
				"config":      map[string]any{},
			})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksPushConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPushConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("force", "false")
	})

	captureStd(t, func() {
		if err := workflowsBlocksPushConfigCmd.RunE(workflowsBlocksPushConfigCmd, []string{"blk_cfg"}); err != nil {
			t.Fatalf("push-config: %v", err)
		}
	})

	cfg, ok := gotPatchBody["config"].(map[string]any)
	if !ok {
		t.Fatalf("PATCH body should include config, got %#v", gotPatchBody)
	}
	if len(cfg) != 0 {
		t.Fatalf("empty replace config = %#v, want empty object", cfg)
	}
	if gotPatchBody["config_mode"] != "replace" {
		t.Fatalf("config_mode = %#v, want replace", gotPatchBody["config_mode"])
	}
}

func TestWorkflowsBlocksPushConfigRejectsRemoteDriftBeforePatch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})
	if err := writeJSONFile(filepath.Join(dir, "config.json"), map[string]any{"prompt": "local edit"}); err != nil {
		t.Fatal(err)
	}

	patchCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			patchCalled = true
			t.Fatal("PATCH should not be called on drift")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_cfg",
			"workflow_id": "wf_cfg",
			"type":        "extract",
			"updated_at":  "2026-06-03T10:02:00Z",
			"config":      map[string]any{"prompt": "remote edit"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksPushConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPushConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("force", "false")
	})

	runErr := workflowsBlocksPushConfigCmd.RunE(workflowsBlocksPushConfigCmd, []string{"blk_cfg"})
	if runErr == nil {
		t.Fatal("expected drift error")
	}
	if !strings.Contains(runErr.Error(), "remote config changed since pull") {
		t.Fatalf("unexpected error: %v", runErr)
	}
	if patchCalled {
		t.Fatal("PATCH must not be called after drift")
	}
}

func TestWorkflowsBlocksPushConfigForceBypassesOnlyRemoteDrift(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})
	if err := writeJSONFile(filepath.Join(dir, "config.json"), map[string]any{"prompt": "local edit"}); err != nil {
		t.Fatal(err)
	}

	patchCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_cfg",
				"workflow_id": "wf_cfg",
				"type":        "extract",
				"updated_at":  "2026-06-03T10:02:00Z",
				"config":      map[string]any{"prompt": "remote edit"},
			})
		case http.MethodPatch:
			patchCalled = true
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			cfg, ok := body["config"].(map[string]any)
			if !ok || cfg["prompt"] != "local edit" {
				t.Fatalf("PATCH config = %#v, want local edit", body["config"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_cfg",
				"workflow_id": "wf_cfg",
				"type":        "extract",
				"updated_at":  "2026-06-03T10:03:00Z",
				"config":      map[string]any{"prompt": "local edit"},
			})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksPushConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	if err := workflowsBlocksPushConfigCmd.Flags().Set("force", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPushConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("force", "false")
	})

	captureStd(t, func() {
		if err := workflowsBlocksPushConfigCmd.RunE(workflowsBlocksPushConfigCmd, []string{"blk_cfg"}); err != nil {
			t.Fatalf("push-config --force: %v", err)
		}
	})
	if !patchCalled {
		t.Fatal("expected PATCH with --force despite remote drift")
	}
}

func TestWorkflowsBlocksPushConfigPreservesBackendSemanticError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})
	if err := writeJSONFile(filepath.Join(dir, "config.json"), map[string]any{"prompt": "bad local field"}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "blk_cfg",
				"workflow_id": "wf_cfg",
				"type":        "extract",
				"updated_at":  "2026-06-03T10:00:00Z",
				"config":      map[string]any{"prompt": "original"},
			})
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"detail": "Invalid config fields for 'extract' block: prompt. Valid fields: image_resolution_dpi, inputs, json_schema, model, n_consensus, review.",
			})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksPushConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPushConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("force", "false")
	})

	_, stderr := captureStd(t, func() {
		err := workflowsBlocksPushConfigCmd.RunE(workflowsBlocksPushConfigCmd, []string{"blk_cfg"})
		if err == nil {
			t.Fatal("expected backend validation error")
		}
	})
	if !strings.Contains(stderr, "Invalid config fields for 'extract' block") {
		t.Fatalf("backend detail should be preserved, got: %s", stderr)
	}
}

func TestWorkflowsBlocksPushConfigRequiresRemoteHashUnlessForced(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})
	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest.RemoteHash = ""
	if err := writeBlockConfigManifest(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}

	httpCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksPushConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPushConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("force", "false")
	})

	err = workflowsBlocksPushConfigCmd.RunE(workflowsBlocksPushConfigCmd, []string{"blk_cfg"})
	if err == nil {
		t.Fatal("expected missing remote_hash error")
	}
	if !strings.Contains(err.Error(), "baseline.config_hash") || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("unexpected error: %v", err)
	}
	if httpCalled {
		t.Fatal("missing remote_hash should be rejected before HTTP")
	}
}

func TestWorkflowsBlocksPushConfigRejectsRemoteTypeMismatchBeforePatch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{"prompt": "original"})
	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest.BlockType = "split"
	manifest.Adapter = "split"
	if err := writeBlockConfigManifest(filepath.Join(dir, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}

	patchCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			patchCalled = true
			t.Fatal("PATCH should not be called on block type mismatch")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_cfg",
			"workflow_id": "wf_cfg",
			"type":        "extract",
			"updated_at":  "2026-06-03T10:00:00Z",
			"config":      map[string]any{"prompt": "original"},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksPushConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksPushConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("workflow-id", "")
		_ = workflowsBlocksPushConfigCmd.Flags().Set("force", "false")
	})

	err = workflowsBlocksPushConfigCmd.RunE(workflowsBlocksPushConfigCmd, []string{"blk_cfg"})
	if err == nil {
		t.Fatal("expected block type mismatch error")
	}
	if !strings.Contains(err.Error(), "block_type split") || !strings.Contains(err.Error(), "extract") {
		t.Fatalf("unexpected error: %v", err)
	}
	if patchCalled {
		t.Fatal("PATCH must not be called after block type mismatch")
	}
}

func TestWorkflowsBlocksDiffConfigReportsChangedPaths(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "test-key")

	dir := t.TempDir()
	writeTestBlockConfigBundle(t, dir, map[string]any{
		"prompt": "original",
		"nested": map[string]any{"keep": "same"},
	})
	if err := writeJSONFile(filepath.Join(dir, "config.json"), map[string]any{
		"prompt": "local edit",
		"nested": map[string]any{"keep": "same", "added": true},
	}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "blk_cfg",
			"workflow_id": "wf_cfg",
			"type":        "extract",
			"updated_at":  "2026-06-03T10:00:00Z",
			"config": map[string]any{
				"prompt": "original",
				"nested": map[string]any{"keep": "same"},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := workflowsBlocksDiffConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = workflowsBlocksDiffConfigCmd.Flags().Set("dir", "")
		_ = workflowsBlocksDiffConfigCmd.Flags().Set("workflow-id", "")
	})

	stdout, _ := captureStd(t, func() {
		if err := workflowsBlocksDiffConfigCmd.RunE(workflowsBlocksDiffConfigCmd, []string{"blk_cfg"}); err != nil {
			t.Fatalf("diff-config: %v", err)
		}
	})

	if !strings.Contains(stdout, `"local_differs_from_remote": true`) {
		t.Fatalf("diff output should report a diff, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"ok": true`) {
		t.Fatalf("diff output should report ok=true, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, `"prompt"`) || !strings.Contains(stdout, `"nested.added"`) {
		t.Fatalf("diff output should include changed paths, got:\n%s", stdout)
	}
}

func writeTestBlockConfigBundle(t *testing.T, dir string, config map[string]any) {
	t.Helper()
	label := "Config"
	block := retab.WorkflowBlock{
		ID:         "blk_cfg",
		WorkflowID: "wf_cfg",
		Type:       retab.WorkflowBlockTypeExtract,
		Label:      &label,
		Config:     config,
		UpdatedAt:  time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC),
	}
	if err := writeBlockConfigBundle(dir, block, true); err != nil {
		t.Fatalf("write bundle: %v", err)
	}
}
