package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSetupInstallsSkillMCPAndRegistry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// configDir/UserHomeDir read %USERPROFILE% on Windows; isolate it too so
	// the bundle + registry land in this temp home, not the shared profile.
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	results, err := runSetup(setupOptions{
		scope:  installScopeGlobal,
		cwd:    t.TempDir(),
		agents: []string{"codex", "claude-code"},
		apiKey: "rtb_test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 5 {
		t.Fatalf("len(results) = %d, want 5", len(results))
	}

	bundledSkill := filepath.Join(home, ".retab", "bundles", "current", "skills", "retab", "SKILL.md")
	assertFileContains(t, bundledSkill, "name: retab")
	assertFileContains(t, filepath.Join(home, ".agents", "skills", "retab", "SKILL.md"), "Six document primitives")
	assertFileContains(t, filepath.Join(home, ".codex", "skills", "retab", "SKILL.md"), "Six document primitives")
	assertFileContains(t, filepath.Join(home, ".claude", "skills", "retab", "SKILL.md"), "Six document primitives")

	assertFileContains(t, filepath.Join(home, ".codex", "config.toml"), `[mcp_servers.retab]`)
	assertFileContains(t, filepath.Join(home, ".codex", "config.toml"), `type = "streamable-http"`)
	assertFileContains(t, filepath.Join(home, ".codex", "config.toml"), `Authorization = "Bearer rtb_test"`)
	assertFileContains(t, filepath.Join(home, ".codex", "config.toml"), `[mcp_servers.retab-docs]`)
	assertFileContains(t, filepath.Join(home, ".codex", "config.toml"), retabDocsMCPURL)

	var claude map[string]any
	readJSONFile(t, filepath.Join(home, ".claude.json"), &claude)
	servers := claude["mcpServers"].(map[string]any)
	retabServer := servers["retab"].(map[string]any)
	if retabServer["url"] != retabMCPURL {
		t.Fatalf("claude retab url = %v, want %s", retabServer["url"], retabMCPURL)
	}
	docsServer := servers["retab-docs"].(map[string]any)
	if docsServer["url"] != retabDocsMCPURL {
		t.Fatalf("claude retab-docs url = %v, want %s", docsServer["url"], retabDocsMCPURL)
	}
	if _, hasHeaders := docsServer["headers"]; hasHeaders {
		t.Fatalf("claude retab-docs should carry no headers, got %v", docsServer["headers"])
	}

	var registry installRegistry
	readJSONFile(t, filepath.Join(home, ".retab", "install-registry.json"), &registry)
	if len(registry.Installs) != 2 {
		t.Fatalf("registry installs = %d, want 2", len(registry.Installs))
	}
}

func TestRunSetupLocalSkipsWindsurfByDefault(t *testing.T) {
	home := t.TempDir()
	cwd := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))

	results, err := runSetup(setupOptions{scope: installScopeLocal, cwd: cwd})
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Agent == "windsurf" {
			t.Fatal("local default setup included windsurf")
		}
	}
	assertFileContains(t, filepath.Join(cwd, ".agents", "skills", "retab", "SKILL.md"), "Six document primitives")
	assertFileContains(t, filepath.Join(cwd, ".codex", "config.toml"), `[mcp_servers.retab]`)
	assertFileContains(t, filepath.Join(cwd, ".mcp.json"), `"retab"`)
}

func TestRunSyncRefreshesUniversalSkillAndMCP(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Isolate %USERPROFILE% too (Windows) so runSync reads only this test's
	// install registry, not a sibling test's leaked installs.
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))

	_, err := runSetup(setupOptions{
		scope:  installScopeGlobal,
		cwd:    t.TempDir(),
		agents: []string{"codex"},
		apiKey: "rtb_old",
	})
	if err != nil {
		t.Fatal(err)
	}

	universalSkill := filepath.Join(home, ".agents", "skills", "retab")
	if err := os.RemoveAll(universalSkill); err != nil {
		t.Fatal(err)
	}

	results, err := runSync("rtb_fresh")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	assertFileContains(t, filepath.Join(universalSkill, "SKILL.md"), "Six document primitives")
	assertFileContains(t, filepath.Join(home, ".codex", "config.toml"), `Authorization = "Bearer rtb_fresh"`)
}

func TestUpsertTomlMCPConfigReplacesExistingSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[mcp_servers.old]\nurl = \"old\"\n\n[mcp_servers.retab]\nurl = \"stale\"\n\n[mcp_servers.retab.headers]\nAuthorization = \"Bearer stale\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := upsertTomlMCPConfig(path, "retab", mcpServerConfig{
		Type:    "streamable-http",
		URL:     retabMCPURL,
		Headers: map[string]string{"Authorization": "Bearer fresh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if strings.Count(text, "[mcp_servers.retab]") != 1 {
		t.Fatalf("retab section count = %d, want 1\n%s", strings.Count(text, "[mcp_servers.retab]"), text)
	}
	if !strings.Contains(text, `Authorization = "Bearer fresh"`) || strings.Contains(text, "stale") {
		t.Fatalf("toml was not replaced correctly:\n%s", text)
	}
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(raw), want) {
		t.Fatalf("%s does not contain %q:\n%s", path, want, string(raw))
	}
}

func readJSONFile(t *testing.T, path string, out any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal %s: %v\n%s", path, err, string(raw))
	}
}
