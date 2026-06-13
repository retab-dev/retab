//go:build !retab_oagen_cli_workflows_blocks

package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

const blockConfigBundleTargetKind = "workflow_block"

type blockConfigBundleManifest struct {
	WorkflowID      string            `json:"workflow_id,omitempty"`
	BlockID         string            `json:"block_id"`
	BlockType       string            `json:"block_type"`
	BlockLabel      string            `json:"block_label,omitempty"`
	Adapter         string            `json:"adapter"`
	RemoteUpdatedAt string            `json:"remote_updated_at,omitempty"`
	RemoteHash      string            `json:"remote_hash"`
	Files           map[string]string `json:"files"`
}

type blockConfigBundleManifestJSON struct {
	Target   blockConfigBundleTargetJSON   `json:"target"`
	Adapter  blockConfigBundleAdapterJSON  `json:"adapter"`
	Baseline blockConfigBundleBaselineJSON `json:"baseline"`
	Files    []blockConfigBundleFileJSON   `json:"files"`
}

type blockConfigBundleTargetJSON struct {
	Kind       string `json:"kind"`
	WorkflowID string `json:"workflow_id"`
	BlockID    string `json:"block_id"`
	BlockType  string `json:"block_type"`
	BlockLabel string `json:"block_label,omitempty"`
}

type blockConfigBundleAdapterJSON struct {
	Name string `json:"name"`
}

type blockConfigBundleBaselineJSON struct {
	ConfigHash     string `json:"config_hash"`
	BlockUpdatedAt string `json:"block_updated_at,omitempty"`
}

type blockConfigBundleFileJSON struct {
	Role   string `json:"role"`
	Path   string `json:"path"`
	Format string `json:"format"`
}

var workflowsBlocksConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit workflow block config as local bundles",
	Long: `Pull a workflow block's config into editable local files, push edits back,
and diff/validate/doctor the local bundle.

Config is treated as a field on the workflow block, not a separate API
resource: ` + "`pull`" + ` composes from ` + "`workflows blocks get`" + ` and ` + "`push`" + `
from ` + "`workflows blocks get`/`update`" + `. The bundle records the remote
config hash so a later ` + "`push`" + ` can detect dashboard edits made after the
pull.`,
}

var workflowsBlocksPullConfigCmd = &cobra.Command{
	Use:   "pull [<workflow-id>] <block-id>",
	Short: "Pull a block config into an editable local bundle",
	Long: `Pull a workflow block's config into local editable files.

This command treats config as a field on the workflow block, not as a separate
API resource. It fetches the block through workflows blocks get, writes
manifest.json plus config files, and records the remote config hash so a later
push can detect dashboard edits made after the pull.

Large fields are split into separate files for block types where that makes the
bundle easier to edit. Push reassembles those files and updates the block with
config_mode=replace.`,
	Example: `  # Pull by org-unique block id
  retab workflows blocks config pull block_def456 --out tmp/block_def456

  # Legacy duplicate block-id disambiguation
  retab workflows blocks config pull wf_abc123 block_def456 --out tmp/block_def456`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, blockID, err := resolveBlockPositionalWorkflowID(cmd, args)
		if err != nil {
			return err
		}
		dir, _ := cmd.Flags().GetString("out")
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return fmt.Errorf("--out is required")
		}
		force, _ := cmd.Flags().GetBool("force")

		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		block, err := client.Workflows.Blocks.Get(ctx, blockID, &retab.WorkflowBlocksGetParams{WorkflowID: workflowID})
		if err != nil {
			return err
		}
		if err := writeBlockConfigBundle(dir, *block, force); err != nil {
			return err
		}
		runtimeHydrated := false
		runtimeAdapter := ""
		if block.Type == retab.WorkflowBlockTypeFunction {
			sourceSchema, err := fetchFunctionSourceSchema(cmd, *block)
			if err != nil {
				return err
			}
			if err := hydrateFunctionBundle(dir, block.Config, sourceSchema, force); err != nil {
				return err
			}
			runtimeHydrated = true
			runtimeAdapter = "function"
		}
		if block.Type == retab.WorkflowBlockTypeAPICall {
			if err := hydrateAPICallBundle(dir, block.Config, force); err != nil {
				return err
			}
			runtimeHydrated = true
			runtimeAdapter = "api_call"
		}
		return printJSON(map[string]any{
			"ok":                 true,
			"workflow_id":        block.WorkflowID,
			"block_id":           block.ID,
			"block_type":         block.Type,
			"dir":                dir,
			"config_hash":        hashJSONMap(block.Config),
			"runtime_hydratable": block.Type == retab.WorkflowBlockTypeFunction || block.Type == retab.WorkflowBlockTypeAPICall,
			"runtime_hydrated":   runtimeHydrated,
			"runtime_adapter":    runtimeAdapter,
		})
	}),
}

type blockResolvedSchemasResponse struct {
	WorkflowID string                    `json:"workflow_id"`
	BlockID    string                    `json:"block_id"`
	Schema     blockResolvedSchemasValue `json:"schema"`
}

type blockResolvedSchemasValue struct {
	InputSchemas map[string]any `json:"input_schemas"`
}

func fetchFunctionSourceSchema(cmd *cobra.Command, block retab.WorkflowBlock) (map[string]any, error) {
	if block.WorkflowID == "" || block.ID == "" {
		return nil, nil
	}
	var response blockResolvedSchemasResponse
	query := url.Values{"workflow_id": []string{block.WorkflowID}}
	err := cliJSONRequestInto(
		cmd,
		http.MethodGet,
		"/v1/workflows/blocks/"+url.PathEscape(block.ID)+"/resolved-schemas",
		query,
		nil,
		&response,
	)
	if err != nil {
		return nil, err
	}
	sourceSchema, _ := response.Schema.InputSchemas["default"].(map[string]any)
	return sourceSchema, nil
}

var workflowsBlocksValidateConfigCmd = &cobra.Command{
	Use:   "validate [<workflow-id>] <block-id>",
	Short: "Validate a local block config bundle",
	Long: `Validate a local workflow block config bundle without mutating remote state.

Validation checks manifest consistency, local JSON shape, adapter file
assembly, and local block-config rules that the CLI can enforce before a
network call. By default it then asks the backend to dry-run the assembled
config against the target block using the same validation policy as
push-config. It does not create workflow runs and does not publish.

Pass --offline for local-only validation when you do not want a network call.
Offline validation is useful but is not authoritative for backend block
semantics.`,
	Example: `  retab workflows blocks config validate block_def456 --dir tmp/block_def456

  retab workflows blocks config validate wf_abc123 block_def456 --dir tmp/block_def456

  retab workflows blocks config validate block_def456 --dir tmp/block_def456 --offline`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, blockID, err := resolveBlockPositionalWorkflowID(cmd, args)
		if err != nil {
			return err
		}
		dir, _ := cmd.Flags().GetString("dir")
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return fmt.Errorf("--dir is required")
		}
		manifest, config, err := readBlockConfigBundle(dir)
		if err != nil {
			return err
		}
		if err := validateBlockConfigTarget(dir, manifest, blockID, workflowID); err != nil {
			return err
		}
		if err := validateBlockConfigBundle(manifest, config); err != nil {
			return err
		}
		localChecks := runLocalBlockConfigChecks(dir, manifest, config)
		offline, _ := cmd.Flags().GetBool("offline")
		if !offline {
			client, err := newClient(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := ctxFor(cmd)
			defer cancel()
			mode := retab.UpdateWorkflowBlockRequestConfigModeReplace
			result, err := client.Workflows.Blocks.CreateBlockValidateConfig(ctx, manifest.BlockID, &retab.WorkflowBlocksCreateBlockValidateConfigParams{
				Config:     config,
				ConfigMode: &mode,
				WorkflowID: workflowIDFromManifestOrArg(manifest, workflowID),
			})
			if err != nil {
				return err
			}
			return printJSON(map[string]any{
				"ok":            result.Ok,
				"mode":          "remote",
				"authoritative": true,
				"workflow_id":   result.WorkflowID,
				"block_id":      result.BlockID,
				"block_type":    result.BlockType,
				"adapter":       manifest.Adapter,
				"config_hash":   result.ConfigHash,
				"local_checks":  localChecks,
			})
		}
		return printJSON(map[string]any{
			"ok":            true,
			"mode":          "offline",
			"authoritative": false,
			"workflow_id":   manifest.WorkflowID,
			"block_id":      manifest.BlockID,
			"block_type":    manifest.BlockType,
			"adapter":       manifest.Adapter,
			"config_hash":   hashJSONMap(config),
			"local_checks":  localChecks,
		})
	}),
}

var workflowsBlocksPushConfigCmd = &cobra.Command{
	Use:   "push [<workflow-id>] <block-id>",
	Short: "Push a local block config bundle back to Retab",
	Long: `Push a local workflow block config bundle back to the remote draft.

The command reassembles the full block config from local files and updates the
block with config_mode=replace. It fetches the current remote block first and
refuses to push when the remote config hash differs from the hash recorded at
pull time, unless --force is passed.

This mutates only the workflow draft. It does not publish and does not create
workflow runs.`,
	Example: `  retab workflows blocks config push block_def456 --dir tmp/block_def456

  retab workflows blocks config push wf_abc123 block_def456 --dir tmp/block_def456`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, blockID, err := resolveBlockPositionalWorkflowID(cmd, args)
		if err != nil {
			return err
		}
		dir, _ := cmd.Flags().GetString("dir")
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return fmt.Errorf("--dir is required")
		}
		manifest, config, err := readBlockConfigBundle(dir)
		if err != nil {
			return err
		}
		if err := validateBlockConfigTarget(dir, manifest, blockID, workflowID); err != nil {
			return err
		}
		if err := validateBlockConfigBundle(manifest, config); err != nil {
			return err
		}
		force, _ := cmd.Flags().GetBool("force")
		if !force && strings.TrimSpace(manifest.RemoteHash) == "" {
			return fmt.Errorf("manifest baseline.config_hash is required for drift-safe push; re-pull the block or pass --force")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		remote, err := client.Workflows.Blocks.Get(ctx, blockID, &retab.WorkflowBlocksGetParams{WorkflowID: workflowIDFromManifestOrArg(manifest, workflowID)})
		if err != nil {
			return err
		}
		if err := validateRemoteBlockConfigTarget(dir, manifest, *remote); err != nil {
			return err
		}
		remoteHash := hashJSONMap(remote.Config)
		if !force && manifest.RemoteHash != "" && remoteHash != manifest.RemoteHash {
			return fmt.Errorf("remote config changed since pull: manifest hash %s, remote hash %s; re-pull or pass --force", manifest.RemoteHash, remoteHash)
		}
		mode := retab.UpdateWorkflowBlockRequestConfigModeReplace
		result, err := client.Workflows.Blocks.Update(ctx, blockID, &retab.WorkflowBlocksUpdateParams{
			WorkflowID: workflowIDFromManifestOrArg(manifest, workflowID),
			Config:     &config,
			ConfigMode: &mode,
			Label:      nil,
			PositionX:  nil,
			PositionY:  nil,
			Width:      nil,
			Height:     nil,
			ParentID:   nil,
		})
		if err != nil {
			return err
		}
		if err := refreshBlockConfigBundle(dir, *result); err != nil {
			return err
		}
		return printJSON(map[string]any{
			"ok":          true,
			"workflow_id": result.WorkflowID,
			"block_id":    result.ID,
			"block_type":  result.Type,
			"config_hash": hashJSONMap(result.Config),
		})
	}),
}

var workflowsBlocksDiffConfigCmd = &cobra.Command{
	Use:   "diff [<workflow-id>] <block-id>",
	Short: "Diff a local block config bundle against the current remote block",
	Long: `Diff a local workflow block config bundle against the current remote block.

The command reassembles local config, fetches the current remote block, and
prints a structural JSON-path summary. It does not mutate remote state.`,
	Example: `  retab workflows blocks config diff block_def456 --dir tmp/block_def456`,
	Args:    cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, blockID, err := resolveBlockPositionalWorkflowID(cmd, args)
		if err != nil {
			return err
		}
		dir, _ := cmd.Flags().GetString("dir")
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return fmt.Errorf("--dir is required")
		}
		manifest, localConfig, err := readBlockConfigBundle(dir)
		if err != nil {
			return err
		}
		if err := validateBlockConfigTarget(dir, manifest, blockID, workflowID); err != nil {
			return err
		}
		if err := validateBlockConfigBundle(manifest, localConfig); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		remote, err := client.Workflows.Blocks.Get(ctx, blockID, &retab.WorkflowBlocksGetParams{WorkflowID: workflowIDFromManifestOrArg(manifest, workflowID)})
		if err != nil {
			return err
		}
		if err := validateRemoteBlockConfigTarget(dir, manifest, *remote); err != nil {
			return err
		}
		changes := diffJSONPaths(remote.Config, localConfig)
		return printJSON(map[string]any{
			"ok":                            true,
			"workflow_id":                   remote.WorkflowID,
			"block_id":                      remote.ID,
			"block_type":                    remote.Type,
			"remote_config_hash":            hashJSONMap(remote.Config),
			"local_config_hash":             hashJSONMap(localConfig),
			"remote_changed_since_baseline": manifest.RemoteHash != "" && hashJSONMap(remote.Config) != manifest.RemoteHash,
			"local_differs_from_remote":     len(changes) > 0,
			"changed_paths":                 changes,
			"baseline_config_hash":          manifest.RemoteHash,
			"assembled_local_config_valid":  true,
		})
	}),
}

var workflowsBlocksDoctorConfigCmd = &cobra.Command{
	Use:   "doctor <bundle-dir>",
	Short: "Diagnose a local block config bundle",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		result := diagnoseBlockConfigBundle(args[0])
		return printJSON(result)
	}),
}

func diagnoseBlockConfigBundle(dir string) map[string]any {
	problems := []map[string]any{}
	manifestPath := filepath.Join(dir, "manifest.json")
	manifest, err := readBlockConfigManifest(manifestPath)
	if err != nil {
		problems = append(problems, map[string]any{
			"kind":  "invalid_or_missing_manifest",
			"path":  "manifest.json",
			"error": err.Error(),
			"fix":   "retab workflows blocks config pull <workflow-id> <block-id> --out " + dir,
		})
		return map[string]any{
			"ok":       false,
			"dir":      dir,
			"problems": problems,
		}
	}
	_, config, err := readBlockConfigBundle(dir)
	if err != nil {
		problems = append(problems, map[string]any{
			"kind":  "invalid_bundle_files",
			"error": err.Error(),
			"fix":   "retab workflows blocks config pull " + manifest.WorkflowID + " " + manifest.BlockID + " --out " + dir + " --force",
		})
	} else if err := validateBlockConfigBundle(manifest, config); err != nil {
		problems = append(problems, map[string]any{
			"kind":  "invalid_bundle",
			"error": err.Error(),
			"fix":   "retab workflows blocks config pull " + manifest.WorkflowID + " " + manifest.BlockID + " --out " + dir + " --force",
		})
	}
	for role, rel := range manifest.Files {
		path, pathErr := bundleFilePath(dir, rel)
		if pathErr != nil {
			problems = append(problems, map[string]any{
				"kind":  "invalid_manifest_path",
				"role":  role,
				"path":  rel,
				"error": pathErr.Error(),
			})
			continue
		}
		if _, statErr := os.Stat(path); statErr != nil {
			problems = append(problems, map[string]any{
				"kind":  "missing_bundle_file",
				"role":  role,
				"path":  rel,
				"error": statErr.Error(),
				"fix":   "retab workflows blocks config pull " + manifest.WorkflowID + " " + manifest.BlockID + " --out " + dir + " --force",
			})
		}
	}
	problems = append(problems, diagnoseBlockRuntimeFiles(dir, manifest)...)
	localChecks := []localBlockConfigCheck{}
	if config != nil {
		localChecks = runLocalBlockConfigChecks(dir, manifest, config)
	}
	return map[string]any{
		"ok":           len(problems) == 0,
		"dir":          dir,
		"workflow_id":  manifest.WorkflowID,
		"block_id":     manifest.BlockID,
		"block_type":   manifest.BlockType,
		"adapter":      manifest.Adapter,
		"problems":     problems,
		"local_checks": localChecks,
	}
}

func readBlockConfigBundleConfigOnly(dir string, manifest blockConfigBundleManifest) (map[string]any, error) {
	configPath, err := bundleFilePath(dir, manifest.Files["config"])
	if err != nil {
		return nil, fmt.Errorf("manifest files.config: %w", err)
	}
	config, err := readJSONMap(configPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", manifest.Files["config"], err)
	}
	return config, nil
}

func diagnoseBlockRuntimeFiles(dir string, manifest blockConfigBundleManifest) []map[string]any {
	problems := []map[string]any{}
	switch manifest.BlockType {
	case "api_call":
		for _, rel := range []string{"run.sh", ".env.example", ".env.local", "samples", "rendered", "outputs", "traces"} {
			if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
				problems = append(problems, map[string]any{
					"kind":  "missing_runtime_file",
					"path":  rel,
					"error": err.Error(),
					"fix":   "retab workflows blocks api-calls hydrate " + dir,
				})
			}
		}
		for _, rel := range []string{"run.py", "curl.sh"} {
			if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
				problems = append(problems, map[string]any{
					"kind": "stale_runtime_file",
					"path": rel,
					"fix":  "retab workflows blocks api-calls hydrate " + dir + " --force",
				})
			}
		}
	case "function":
		for _, rel := range functionRuntimeFiles(manifest) {
			if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
				problems = append(problems, map[string]any{
					"kind":  "missing_runtime_file",
					"path":  rel,
					"error": err.Error(),
					"fix":   "retab workflows blocks functions hydrate " + dir,
				})
			}
		}
		for _, rel := range staleFunctionRuntimeFiles(manifest) {
			if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
				problems = append(problems, map[string]any{
					"kind": "stale_runtime_file",
					"path": rel,
					"fix":  "retab workflows blocks functions hydrate " + dir + " --force",
				})
			}
		}
	}
	return problems
}

func writeBlockConfigBundle(dir string, block retab.WorkflowBlock, force bool) error {
	config, err := cloneJSONMap(block.Config)
	if err != nil {
		return fmt.Errorf("clone block config: %w", err)
	}
	files := map[string]string{"config": "config.json"}
	splitLargeConfigFields(string(block.Type), config, files)
	manifest := blockConfigBundleManifest{
		WorkflowID:      block.WorkflowID,
		BlockID:         block.ID,
		BlockType:       string(block.Type),
		Adapter:         blockConfigAdapter(string(block.Type)),
		RemoteUpdatedAt: block.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		RemoteHash:      hashJSONMap(block.Config),
		Files:           files,
	}
	if block.Label != nil {
		manifest.BlockLabel = *block.Label
	}
	if err := ensureBundleDirWritable(dir, force); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for key, rel := range files {
		path := filepath.Join(dir, rel)
		switch key {
		case "config":
			if err := writeJSONFile(path, config); err != nil {
				return err
			}
		case "function":
			code, _ := block.Config["code"].(string)
			if err := os.WriteFile(path, []byte(code), 0o600); err != nil {
				return err
			}
		case "json_schema":
			if err := writeJSONFile(path, block.Config["json_schema"]); err != nil {
				return err
			}
		case "output_schema":
			if err := writeJSONFile(path, block.Config["output_schema"]); err != nil {
				return err
			}
		case "request_schema":
			if err := writeJSONFile(path, block.Config["request_schema"]); err != nil {
				return err
			}
		case "mounts":
			if err := writeJSONFile(path, block.Config["mounts"]); err != nil {
				return err
			}
		case "subdocuments":
			if err := writeJSONFile(path, block.Config["subdocuments"]); err != nil {
				return err
			}
		case "categories":
			if err := writeJSONFile(path, block.Config["categories"]); err != nil {
				return err
			}
		}
	}
	return writeBlockConfigManifest(filepath.Join(dir, "manifest.json"), manifest)
}

func splitLargeConfigFields(blockType string, config map[string]any, files map[string]string) {
	if _, ok := config["json_schema"]; ok && (blockType == "start_json" || blockType == "extract") {
		files["json_schema"] = "json_schema.json"
		delete(config, "json_schema")
	}
	if blockType == "function" {
		if _, ok := config["code"].(string); ok {
			files["function"] = functionCodeFile(config)
			delete(config, "code")
		}
		if _, ok := config["output_schema"]; ok {
			files["output_schema"] = "output_schema.json"
			delete(config, "output_schema")
		}
		if _, ok := config["mounts"]; ok {
			files["mounts"] = "mounts.json"
			delete(config, "mounts")
		}
	}
	if blockType == "api_call" {
		if _, ok := config["request_schema"]; ok {
			files["request_schema"] = "request_schema.json"
			delete(config, "request_schema")
		}
		if _, ok := config["output_schema"]; ok {
			files["output_schema"] = "output_schema.json"
			delete(config, "output_schema")
		}
		if _, ok := config["mounts"]; ok {
			files["mounts"] = "mounts.json"
			delete(config, "mounts")
		}
	}
	if blockType == "split" {
		if _, ok := config["subdocuments"]; ok {
			files["subdocuments"] = "subdocuments.json"
			delete(config, "subdocuments")
		}
	}
	if blockType == "classifier" {
		if _, ok := config["categories"]; ok {
			files["categories"] = "categories.json"
			delete(config, "categories")
		}
	}
}

func readBlockConfigBundle(dir string) (blockConfigBundleManifest, map[string]any, error) {
	manifest, err := readBlockConfigManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return manifest, nil, fmt.Errorf("read manifest.json: %w", err)
	}
	if manifest.Files == nil {
		return manifest, nil, fmt.Errorf("manifest files.config is required")
	}
	if strings.TrimSpace(manifest.Files["config"]) == "" {
		return manifest, nil, fmt.Errorf("manifest files.config is required")
	}
	configPath, err := bundleFilePath(dir, manifest.Files["config"])
	if err != nil {
		return manifest, nil, fmt.Errorf("manifest files.config: %w", err)
	}
	config, err := readJSONMap(configPath)
	if err != nil {
		return manifest, nil, fmt.Errorf("read %s: %w", manifest.Files["config"], err)
	}
	for key, rel := range manifest.Files {
		if key == "config" {
			continue
		}
		path, err := bundleFilePath(dir, rel)
		if err != nil {
			return manifest, nil, fmt.Errorf("manifest files.%s: %w", key, err)
		}
		switch key {
		case "function":
			raw, err := os.ReadFile(path)
			if err != nil {
				return manifest, nil, fmt.Errorf("read %s: %w", rel, err)
			}
			config["code"] = string(raw)
		case "json_schema":
			value, err := readJSONMap(path)
			if err != nil {
				return manifest, nil, fmt.Errorf("read %s: %w", rel, err)
			}
			config["json_schema"] = value
		case "output_schema":
			value, err := readJSONMap(path)
			if err != nil {
				return manifest, nil, fmt.Errorf("read %s: %w", rel, err)
			}
			config["output_schema"] = value
		case "request_schema":
			value, err := readJSONMap(path)
			if err != nil {
				return manifest, nil, fmt.Errorf("read %s: %w", rel, err)
			}
			config["request_schema"] = value
		case "mounts":
			value, err := readJSONMap(path)
			if err != nil {
				return manifest, nil, fmt.Errorf("read %s: %w", rel, err)
			}
			config["mounts"] = stripLocalMountFields(value)
		case "subdocuments":
			value, err := readJSONArray(path)
			if err != nil {
				return manifest, nil, fmt.Errorf("read %s: %w", rel, err)
			}
			config["subdocuments"] = value
		case "categories":
			value, err := readJSONArray(path)
			if err != nil {
				return manifest, nil, fmt.Errorf("read %s: %w", rel, err)
			}
			config["categories"] = value
		default:
			return manifest, nil, fmt.Errorf("manifest files.%s is not a supported block config file role", key)
		}
	}
	return manifest, config, nil
}

func stripLocalMountFields(mounts map[string]any) map[string]any {
	out := make(map[string]any, len(mounts))
	for key, value := range mounts {
		out[key] = value
	}
	rawTables, ok := mounts["tables"].([]any)
	if !ok {
		return out
	}
	tables := make([]any, 0, len(rawTables))
	for _, raw := range rawTables {
		table, ok := raw.(map[string]any)
		if !ok {
			tables = append(tables, raw)
			continue
		}
		cleanTable := make(map[string]any, len(table))
		for key, value := range table {
			if key == "local_path" {
				continue
			}
			cleanTable[key] = value
		}
		tables = append(tables, cleanTable)
	}
	out["tables"] = tables
	return out
}

func readBlockConfigManifest(path string) (blockConfigBundleManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return blockConfigBundleManifest{}, err
	}
	var manifestJSON blockConfigBundleManifestJSON
	if err := decodeJSONStrict(raw, &manifestJSON); err != nil {
		return blockConfigBundleManifest{}, err
	}
	if manifestJSON.Target.Kind != blockConfigBundleTargetKind {
		return blockConfigBundleManifest{}, fmt.Errorf("manifest target.kind %q is unsupported", manifestJSON.Target.Kind)
	}
	files := make(map[string]string, len(manifestJSON.Files))
	for _, file := range manifestJSON.Files {
		role := strings.TrimSpace(file.Role)
		if role == "" {
			return blockConfigBundleManifest{}, fmt.Errorf("manifest files role is required")
		}
		expectedFormat := blockConfigFileFormat(role, file.Path)
		if file.Format != expectedFormat {
			return blockConfigBundleManifest{}, fmt.Errorf("manifest files.%s format %q does not match expected %q", role, file.Format, expectedFormat)
		}
		if _, exists := files[role]; exists {
			return blockConfigBundleManifest{}, fmt.Errorf("manifest files.%s is duplicated", role)
		}
		files[role] = file.Path
	}
	return blockConfigBundleManifest{
		WorkflowID:      manifestJSON.Target.WorkflowID,
		BlockID:         manifestJSON.Target.BlockID,
		BlockType:       manifestJSON.Target.BlockType,
		BlockLabel:      manifestJSON.Target.BlockLabel,
		Adapter:         manifestJSON.Adapter.Name,
		RemoteUpdatedAt: manifestJSON.Baseline.BlockUpdatedAt,
		RemoteHash:      manifestJSON.Baseline.ConfigHash,
		Files:           files,
	}, nil
}

func writeBlockConfigManifest(path string, manifest blockConfigBundleManifest) error {
	return writeJSONFile(path, blockConfigBundleManifestJSON{
		Target: blockConfigBundleTargetJSON{
			Kind:       blockConfigBundleTargetKind,
			WorkflowID: manifest.WorkflowID,
			BlockID:    manifest.BlockID,
			BlockType:  manifest.BlockType,
			BlockLabel: manifest.BlockLabel,
		},
		Adapter: blockConfigBundleAdapterJSON{
			Name: manifest.Adapter,
		},
		Baseline: blockConfigBundleBaselineJSON{
			ConfigHash:     manifest.RemoteHash,
			BlockUpdatedAt: manifest.RemoteUpdatedAt,
		},
		Files: blockConfigManifestFiles(manifest.Files),
	})
}

func blockConfigManifestFiles(files map[string]string) []blockConfigBundleFileJSON {
	roles := make([]string, 0, len(files))
	for role := range files {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	if len(roles) > 1 {
		for index, role := range roles {
			if role == "config" {
				copy(roles[1:index+1], roles[0:index])
				roles[0] = "config"
				break
			}
		}
	}
	out := make([]blockConfigBundleFileJSON, 0, len(roles))
	for _, role := range roles {
		out = append(out, blockConfigBundleFileJSON{
			Role:   role,
			Path:   files[role],
			Format: blockConfigFileFormat(role, files[role]),
		})
	}
	return out
}

func blockConfigFileFormat(role string, path string) string {
	switch role {
	case "function":
		if strings.EqualFold(filepath.Ext(path), ".ts") || strings.EqualFold(filepath.Ext(path), ".tsx") {
			return "typescript"
		}
		return "python"
	default:
		return "json"
	}
}

func validateBlockConfigBundle(manifest blockConfigBundleManifest, config map[string]any) error {
	if strings.TrimSpace(manifest.BlockID) == "" {
		return fmt.Errorf("manifest block_id is required")
	}
	if strings.TrimSpace(manifest.WorkflowID) == "" {
		return fmt.Errorf("manifest workflow_id is required")
	}
	if strings.TrimSpace(manifest.BlockType) == "" {
		return fmt.Errorf("manifest block_type is required")
	}
	expectedAdapter := blockConfigAdapter(manifest.BlockType)
	if strings.TrimSpace(manifest.Adapter) != expectedAdapter {
		return fmt.Errorf("manifest adapter %q does not match %s block config adapter %q", manifest.Adapter, manifest.BlockType, expectedAdapter)
	}
	if _, ok := manifest.Files["config"]; !ok {
		return fmt.Errorf("manifest files.config is required")
	}
	allowedFileRoles := blockConfigFileRoles(manifest.BlockType)
	for role := range manifest.Files {
		if !allowedFileRoles[role] {
			return fmt.Errorf("manifest files.%s is not supported for %s block config bundles", role, manifest.BlockType)
		}
	}
	if manifest.BlockType == "function" {
		if code, ok := config["code"].(string); !ok || strings.TrimSpace(code) == "" {
			return fmt.Errorf("function bundle requires non-empty %s/config.code", manifest.Files["function"])
		}
	}
	return rejectLegacyReviewConfig(&config)
}

func validateBlockConfigTarget(dir string, manifest blockConfigBundleManifest, blockID string, workflowID *string) error {
	if manifest.BlockID != blockID {
		return fmt.Errorf("directory %s belongs to block %s, not %s", dir, manifest.BlockID, blockID)
	}
	if workflowID != nil && manifest.WorkflowID != "" && manifest.WorkflowID != *workflowID {
		return fmt.Errorf("directory %s belongs to workflow %s, not %s", dir, manifest.WorkflowID, *workflowID)
	}
	return nil
}

func validateRemoteBlockConfigTarget(dir string, manifest blockConfigBundleManifest, remote retab.WorkflowBlock) error {
	if manifest.BlockID != remote.ID {
		return fmt.Errorf("directory %s belongs to block %s, but remote returned block %s", dir, manifest.BlockID, remote.ID)
	}
	if manifest.WorkflowID != "" && manifest.WorkflowID != remote.WorkflowID {
		return fmt.Errorf("directory %s belongs to workflow %s, but remote block is in workflow %s", dir, manifest.WorkflowID, remote.WorkflowID)
	}
	if manifest.BlockType != string(remote.Type) {
		return fmt.Errorf("directory %s manifest block_type %s does not match remote block type %s", dir, manifest.BlockType, remote.Type)
	}
	return nil
}

func bundleFilePath(dir string, rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("path must be relative to the bundle directory")
	}
	cleaned := filepath.Clean(rel)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path must stay inside the bundle directory")
	}
	return filepath.Join(dir, cleaned), nil
}

func workflowIDFromManifestOrArg(manifest blockConfigBundleManifest, workflowID *string) *string {
	if workflowID != nil {
		return workflowID
	}
	if strings.TrimSpace(manifest.WorkflowID) == "" {
		return nil
	}
	return ptr(manifest.WorkflowID)
}

func refreshBlockConfigBundle(dir string, block retab.WorkflowBlock) error {
	return writeBlockConfigBundle(dir, block, true)
}

func blockConfigAdapter(blockType string) string {
	switch blockType {
	case "function":
		return "function"
	case "start_json", "extract":
		return "json_schema"
	case "split":
		return "split"
	case "classifier":
		return "classifier"
	case "api_call":
		return "api_call"
	default:
		return "generic"
	}
}

func blockConfigFileRoles(blockType string) map[string]bool {
	roles := map[string]bool{"config": true}
	switch blockType {
	case "function":
		roles["function"] = true
		roles["output_schema"] = true
		roles["mounts"] = true
	case "start_json", "extract":
		roles["json_schema"] = true
	case "split":
		roles["subdocuments"] = true
	case "classifier":
		roles["categories"] = true
	case "api_call":
		roles["request_schema"] = true
		roles["output_schema"] = true
		roles["mounts"] = true
	}
	return roles
}

func functionLanguage(config map[string]any) string {
	language := strings.ToLower(strings.TrimSpace(stringFromAny(config["language"])))
	switch language {
	case "typescript", "ts":
		return "typescript"
	default:
		return "python"
	}
}

func functionManifestLanguage(manifest blockConfigBundleManifest) string {
	path := manifest.Files["function"]
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ts", ".tsx":
		return "typescript"
	default:
		return "python"
	}
}

func functionCodeFile(config map[string]any) string {
	if functionLanguage(config) == "typescript" {
		return "function.ts"
	}
	return "function.py"
}

func functionRuntimeFiles(manifest blockConfigBundleManifest) []string {
	if functionManifestLanguage(manifest) == "typescript" {
		return []string{
			"input_schema.json",
			"models.generated.ts",
			"schemas.generated.ts",
			"tsconfig.json",
			"run.mjs",
			".retab/runtime.mjs",
			".env.example",
			".env.local",
			"samples",
			"outputs",
			"traces",
		}
	}
	return []string{"input.py", "output.py", "run.py", ".retab/runtime.py", ".env.example", ".env.local", "samples", "outputs", "traces"}
}

func staleFunctionRuntimeFiles(manifest blockConfigBundleManifest) []string {
	if functionManifestLanguage(manifest) == "typescript" {
		return []string{"input.py", "output.py", "run.py", ".retab/runtime.py", "models.py", "input_models.py", "output_models.py", "retab_runtime.py", "mounts.local.json"}
	}
	return []string{"input_schema.json", "models.generated.ts", "schemas.generated.ts", "tsconfig.json", "run.mjs", ".retab/runtime.mjs", "models.py", "input_models.py", "output_models.py", "retab_runtime.py", "mounts.local.json"}
}

const localCheckOutputLimit = 8 * 1024
const localCheckTimeout = 30 * time.Second

type localBlockConfigCheck struct {
	Name        string `json:"name"`
	Language    string `json:"language"`
	Status      string `json:"status"`
	Required    bool   `json:"required"`
	Command     string `json:"command,omitempty"`
	ExitCode    *int   `json:"exit_code,omitempty"`
	Stdout      string `json:"stdout,omitempty"`
	Stderr      string `json:"stderr,omitempty"`
	Message     string `json:"message,omitempty"`
	InstallHint string `json:"install_hint,omitempty"`
}

func runLocalBlockConfigChecks(dir string, manifest blockConfigBundleManifest, config map[string]any) []localBlockConfigCheck {
	if manifest.BlockType != "function" {
		return []localBlockConfigCheck{}
	}
	switch functionManifestLanguage(manifest) {
	case "typescript":
		return runTypescriptFunctionChecks(dir)
	default:
		return runPythonFunctionChecks(dir, manifest)
	}
}

func runTypescriptFunctionChecks(dir string) []localBlockConfigCheck {
	if _, err := os.Stat(filepath.Join(dir, "tsconfig.json")); err != nil {
		return []localBlockConfigCheck{{
			Name:        "tsc",
			Language:    "typescript",
			Status:      "skipped",
			Required:    false,
			Message:     "tsconfig.json is missing",
			InstallHint: "run `retab workflows blocks functions hydrate <bundle-dir> --force`",
		}}
	}
	return []localBlockConfigCheck{
		runOptionalLocalToolCheck(localToolCheckSpec{
			Name:        "tsc",
			Language:    "typescript",
			EnvVar:      "RETAB_TSC",
			DefaultBin:  "tsc",
			Args:        []string{"--noEmit", "--pretty", "false", "--project", "tsconfig.json"},
			Dir:         dir,
			InstallHint: "install TypeScript, for example `npm install -g typescript`",
		}),
	}
}

func runPythonFunctionChecks(dir string, manifest blockConfigBundleManifest) []localBlockConfigCheck {
	functionPath := manifest.Files["function"]
	if strings.TrimSpace(functionPath) == "" {
		functionPath = "function.py"
	}
	return []localBlockConfigCheck{
		runOptionalLocalToolCheck(localToolCheckSpec{
			Name:        "ruff",
			Language:    "python",
			EnvVar:      "RETAB_RUFF",
			DefaultBin:  "ruff",
			Args:        []string{"check", functionPath},
			Dir:         dir,
			InstallHint: "install Ruff, for example `pip install ruff`",
		}),
		runOptionalLocalToolCheck(localToolCheckSpec{
			Name:        "pyright",
			Language:    "python",
			EnvVar:      "RETAB_PYRIGHT",
			DefaultBin:  "pyright",
			Args:        []string{functionPath},
			Dir:         dir,
			InstallHint: "install Pyright, for example `npm install -g pyright`",
		}),
	}
}

type localToolCheckSpec struct {
	Name        string
	Language    string
	EnvVar      string
	DefaultBin  string
	Args        []string
	Dir         string
	InstallHint string
}

func runOptionalLocalToolCheck(spec localToolCheckSpec) localBlockConfigCheck {
	result := localBlockConfigCheck{
		Name:        spec.Name,
		Language:    spec.Language,
		Status:      "skipped",
		Required:    false,
		InstallHint: spec.InstallHint,
	}
	bin, source, err := resolveOptionalLocalTool(spec.EnvVar, spec.DefaultBin)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	args := append([]string{}, spec.Args...)
	result.Command = strings.Join(append([]string{source}, args...), " ")

	ctx, cancel := context.WithTimeout(context.Background(), localCheckTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = spec.Dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	result.Stdout = truncateLocalCheckOutput(stdout.String())
	result.Stderr = truncateLocalCheckOutput(stderr.String())
	if ctx.Err() == context.DeadlineExceeded {
		result.Status = "failed"
		result.Message = fmt.Sprintf("%s timed out after %s", spec.Name, localCheckTimeout)
		return result
	}
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			result.ExitCode = &exitCode
		}
		return result
	}
	exitCode := 0
	result.ExitCode = &exitCode
	result.Status = "passed"
	result.Message = ""
	return result
}

func resolveOptionalLocalTool(envVar string, defaultBin string) (string, string, error) {
	if override := strings.TrimSpace(os.Getenv(envVar)); override != "" {
		if filepath.IsAbs(override) || strings.ContainsRune(override, filepath.Separator) {
			if info, err := os.Stat(override); err != nil {
				return "", override, fmt.Errorf("%s is set to %q, but that executable was not found", envVar, override)
			} else if info.IsDir() {
				return "", override, fmt.Errorf("%s is set to %q, but it is a directory", envVar, override)
			} else if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
				return "", override, fmt.Errorf("%s is set to %q, but that file is not executable", envVar, override)
			}
			return override, override, nil
		}
		resolved, err := exec.LookPath(override)
		if err != nil {
			return "", override, fmt.Errorf("%s is set to %q, but that executable was not found on PATH", envVar, override)
		}
		return resolved, override, nil
	}
	resolved, err := exec.LookPath(defaultBin)
	if err != nil {
		return "", defaultBin, fmt.Errorf("%s is not installed", defaultBin)
	}
	return resolved, defaultBin, nil
}

func truncateLocalCheckOutput(value string) string {
	if len(value) <= localCheckOutputLimit {
		return value
	}
	return value[:localCheckOutputLimit] + "\n[truncated]\n"
}

func cloneJSONMap(value map[string]any) (map[string]any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func hashJSONMap(value map[string]any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(bytes.TrimSpace(raw))
	return hex.EncodeToString(sum[:])
}

func ensureBundleDirWritable(dir string, force bool) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists and is not a directory", dir)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		if len(entries) > 0 && !force {
			return fmt.Errorf("%s is not empty; pass --force to overwrite bundle files", dir)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func diffJSONPaths(remote map[string]any, local map[string]any) []string {
	paths := make([]string, 0)
	diffJSONValue("", remote, local, &paths)
	sort.Strings(paths)
	return paths
}

func diffJSONValue(path string, a any, b any, paths *[]string) {
	if aMap, ok := a.(map[string]any); ok {
		if bMap, ok := b.(map[string]any); ok {
			keys := make(map[string]struct{}, len(aMap)+len(bMap))
			for key := range aMap {
				keys[key] = struct{}{}
			}
			for key := range bMap {
				keys[key] = struct{}{}
			}
			for key := range keys {
				childPath := key
				if path != "" {
					childPath = path + "." + key
				}
				av, aOK := aMap[key]
				bv, bOK := bMap[key]
				if !aOK || !bOK {
					*paths = append(*paths, childPath)
					continue
				}
				diffJSONValue(childPath, av, bv, paths)
			}
			return
		}
	}
	aRaw, aErr := json.Marshal(a)
	bRaw, bErr := json.Marshal(b)
	if aErr != nil || bErr != nil || !bytes.Equal(aRaw, bRaw) {
		if path == "" {
			*paths = append(*paths, "$")
			return
		}
		*paths = append(*paths, path)
	}
}

func writeJSONFile(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o600)
}

func readJSONFileStrict(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return decodeJSONStrict(raw, out)
}

func decodeJSONStrict(raw []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("invalid trailing JSON data")
	}
	return nil
}

func init() {
	workflowsBlocksPullConfigCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksPullConfigCmd.Flags().String("out", "", "directory to write the editable bundle")
	workflowsBlocksPullConfigCmd.Flags().Bool("force", false, "overwrite files when the output directory is not empty")

	workflowsBlocksPushConfigCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksPushConfigCmd.Flags().String("dir", "", "directory containing the editable bundle")
	workflowsBlocksPushConfigCmd.Flags().Bool("force", false, "push even if the remote config hash changed since pull")

	workflowsBlocksDiffConfigCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksDiffConfigCmd.Flags().String("dir", "", "directory containing the editable bundle")

	workflowsBlocksValidateConfigCmd.Flags().String("workflow-id", "", "scope flat block-id lookup to one workflow (for legacy duplicate ids)")
	workflowsBlocksValidateConfigCmd.Flags().String("dir", "", "directory containing the editable bundle")
	workflowsBlocksValidateConfigCmd.Flags().Bool("offline", false, "validate only local bundle structure without contacting the backend")

	workflowsBlocksConfigCmd.AddCommand(
		workflowsBlocksPullConfigCmd,
		workflowsBlocksPushConfigCmd,
		workflowsBlocksDiffConfigCmd,
		workflowsBlocksValidateConfigCmd,
		workflowsBlocksDoctorConfigCmd,
	)
	workflowsBlocksCmd.AddCommand(workflowsBlocksConfigCmd)
}
