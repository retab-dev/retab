package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Every command in the `blocks config` group speaks one hash space: pull, diff,
// push and manifest.baseline.config_hash all hash the assembled bundle config
// the same way. Online `validate` used to report the BACKEND's hash of the
// executable config under that same `config_hash` key — a different hash space —
// so comparing `validate | jq -r .config_hash` against the manifest baseline
// showed drift on every online run, while `--offline` compared clean.
func validateConfigHashes(t *testing.T, stdout string) map[string]any {
	t.Helper()
	decoded := map[string]any{}
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("validate output is not JSON: %v\n%s", err, stdout)
	}
	return decoded
}

func runValidateConfig(t *testing.T, dir string, offline bool, serverURL string) map[string]any {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	if serverURL != "" {
		t.Setenv("RETAB_API_BASE_URL", serverURL)
	}
	if err := workflowsBlocksValidateConfigCmd.Flags().Set("dir", dir); err != nil {
		t.Fatal(err)
	}
	offlineValue := "false"
	if offline {
		offlineValue = "true"
	}
	if err := workflowsBlocksValidateConfigCmd.Flags().Set("offline", offlineValue); err != nil {
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
	return validateConfigHashes(t, stdout)
}

func TestValidateConfigHashMatchesBundleHashOnlineAndOffline(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{"prompt": "local"}
	writeTestBlockConfigBundle(t, dir, config)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          true,
			"workflow_id": "wf_cfg",
			"block_id":    "blk_cfg",
			"block_type":  "extract",
			"config_hash": "backend-hash-in-a-different-space",
		})
	}))
	defer server.Close()

	online := runValidateConfig(t, dir, false, server.URL)
	offline := runValidateConfig(t, dir, true, "")

	bundleHash := hashJSONMap(config)
	if online["config_hash"] != bundleHash {
		t.Fatalf("online config_hash = %v, want the bundle hash %s", online["config_hash"], bundleHash)
	}
	if offline["config_hash"] != bundleHash {
		t.Fatalf("offline config_hash = %v, want the bundle hash %s", offline["config_hash"], bundleHash)
	}
	if online["config_hash"] != offline["config_hash"] {
		t.Fatalf("config_hash must not change meaning with --offline: online %v vs offline %v",
			online["config_hash"], offline["config_hash"])
	}
	if online["executable_config_hash"] != "backend-hash-in-a-different-space" {
		t.Fatalf("executable_config_hash = %v, want the backend hash", online["executable_config_hash"])
	}
	if _, present := offline["executable_config_hash"]; present {
		t.Fatalf("offline validate has no backend hash to report, got %#v", offline["executable_config_hash"])
	}
}

// The bundle hash validate reports must be the one push refuses to overwrite,
// i.e. the value pull records as manifest.baseline.config_hash.
func TestValidateConfigHashMatchesManifestBaseline(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{"prompt": "local"}
	writeTestBlockConfigBundle(t, dir, config)

	offline := runValidateConfig(t, dir, true, "")

	manifest, assembled, err := readBlockConfigBundle(dir)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if got := hashJSONMap(assembled); offline["config_hash"] != got {
		t.Fatalf("config_hash = %v, want the assembled-config hash %s", offline["config_hash"], got)
	}
	if manifest.RemoteHash != "" && offline["config_hash"] != manifest.RemoteHash {
		t.Fatalf("config_hash %v should be comparable to manifest baseline %s",
			offline["config_hash"], manifest.RemoteHash)
	}
	if !strings.Contains(workflowsBlocksValidateConfigCmd.Long, "executable_config_hash") {
		t.Fatal("validate help should explain the two hash spaces")
	}
}
