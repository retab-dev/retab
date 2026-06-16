package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

//go:embed embedded_skills/retab/SKILL.md
var embeddedRetabSkill embed.FS

const (
	retabMCPServerName       = "retab"
	retabMCPURL              = "https://mcp.retab.com/mcp"
	retabDocsMCPServerName   = "retab-docs"
	retabDocsMCPURL          = "https://docs.retab.com/mcp"
	retabInstallRegistryV1   = 1
	retabInstallerName       = "retab-cli"
	retabSkillInstallName    = "retab"
	retabSkillBundleRootPath = "bundles/current"
)

type setupOptions struct {
	scope  installScope
	cwd    string
	agents []string
	apiKey string
}

type installScope string

const (
	installScopeGlobal installScope = "global"
	installScopeLocal  installScope = "local"
)

type setupAgent struct {
	Name             string
	Label            string
	GlobalSkillDir   func() (string, error)
	LocalSkillDir    func(string) string
	GlobalMCPPath    func() (string, error)
	LocalMCPPath     func(string) string
	MCPFormat        mcpConfigFormat
	MCPKey           string
	TransformMCP     func(mcpServerConfig) any
	SupportsLocalMCP bool
}

type mcpConfigFormat string

const (
	mcpConfigJSON mcpConfigFormat = "json"
	mcpConfigTOML mcpConfigFormat = "toml"
)

type mcpServerConfig struct {
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type installRegistry struct {
	Version  int                    `json:"version"`
	Installs []installRegistryEntry `json:"installs"`
}

type installRegistryEntry struct {
	ID           string       `json:"id"`
	Agent        string       `json:"agent"`
	Scope        installScope `json:"scope"`
	CWD          string       `json:"cwd,omitempty"`
	Installer    string       `json:"installer"`
	RetabVersion string       `json:"retab_version"`
	InstalledAt  string       `json:"installed_at"`
	UpdatedAt    string       `json:"updated_at"`
}

type setupResult struct {
	Agent    string
	SkillDir string
	MCPPath  string
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install Retab skills and MCP config for AI coding agents",
	Long: `Install Retab into supported AI coding agents.

This always writes the bundled Retab skill to the universal .agents/skills
directory used by mainstream agents such as Amp, Antigravity, Cline, Codex,
Cursor, Deep Agents, Dexto, Firebender, Gemini CLI, GitHub Copilot, Kimi Code
CLI, OpenCode, and Warp. It also writes agent-specific skills and MCP entries
for Claude Code, Codex, Cursor, OpenCode, and Windsurf. Use --local to install
project-local config where supported.`,
	Example: `  retab setup
  retab setup --local
  retab setup --agent codex --agent claude-code
  RETAB_API_KEY=rtb_... retab setup`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		local, _ := cmd.Flags().GetBool("local")
		agents, _ := cmd.Flags().GetStringArray("agent")
		apiKey, _ := cmd.Flags().GetString("mcp-api-key")
		if apiKey == "" {
			apiKey = os.Getenv("RETAB_API_KEY")
		}
		if apiKey == "" {
			cfg, _ := loadConfig()
			apiKey = cfg.APIKey
		}
		scope := installScopeGlobal
		if local {
			scope = installScopeLocal
		}
		results, err := runSetup(setupOptions{
			scope:  scope,
			cwd:    mustGetwd(),
			agents: agents,
			apiKey: apiKey,
		})
		if err != nil {
			return err
		}
		for _, result := range results {
			if result.SkillDir != "" {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "installed %s skill: %s\n", result.Agent, result.SkillDir); err != nil {
					return err
				}
			}
			if result.MCPPath != "" {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "installed %s MCP:   %s\n", result.Agent, result.MCPPath); err != nil {
					return err
				}
			}
		}
		if apiKey == "" {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "note: no API key was available, so MCP entries were written without an Authorization header"); err != nil {
				return err
			}
		}
		return nil
	}),
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Refresh tracked Retab skills and MCP config",
	Long: `Refresh Retab skills and MCP entries for every install tracked in
~/.retab/install-registry.json.`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("mcp-api-key")
		if apiKey == "" {
			apiKey = os.Getenv("RETAB_API_KEY")
		}
		if apiKey == "" {
			cfg, _ := loadConfig()
			apiKey = cfg.APIKey
		}
		results, err := runSync(apiKey)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "refreshed Retab skills/MCP (%d setup item%s)\n", len(results), pluralS(len(results)))
		return err
	}),
}

func init() {
	setupCmd.Flags().Bool("local", false, "install into the current project instead of global agent config")
	setupCmd.Flags().StringArray("agent", nil, "target agent: claude-code, codex, cursor, opencode, windsurf (repeatable)")
	setupCmd.Flags().String("mcp-api-key", "", "API key value to write as Authorization: Bearer in MCP config (env: RETAB_API_KEY)")
	syncCmd.Flags().String("mcp-api-key", "", "API key value to write as Authorization: Bearer in MCP config (env: RETAB_API_KEY)")
	rootCmd.AddCommand(setupCmd, syncCmd)
}

func runSetup(opts setupOptions) ([]setupResult, error) {
	if opts.cwd == "" {
		opts.cwd = mustGetwd()
	}
	if opts.scope == "" {
		opts.scope = installScopeGlobal
	}
	agents, err := resolveSetupAgents(opts.agents, opts.scope)
	if err != nil {
		return nil, err
	}
	bundleDir, err := materializeRetabSkillBundle()
	if err != nil {
		return nil, err
	}
	registry, err := loadInstallRegistry()
	if err != nil {
		return nil, err
	}
	universalSkillDir, err := setupUniversalSkill(opts.scope, opts.cwd, bundleDir)
	if err != nil {
		return nil, fmt.Errorf("install universal skill: %w", err)
	}
	results := []setupResult{{Agent: "universal", SkillDir: universalSkillDir}}
	installedSkillDirs := map[string]bool{universalSkillDir: true}
	for _, agent := range agents {
		skillBase, err := skillBaseDir(agent, opts.scope, opts.cwd)
		if err != nil {
			return nil, fmt.Errorf("install %s skill: %w", agent.Name, err)
		}
		skillDir := filepath.Join(skillBase, retabSkillInstallName)
		if !installedSkillDirs[skillDir] {
			skillDir, err = setupSkillForAgent(agent, opts.scope, opts.cwd, bundleDir)
			if err != nil {
				return nil, fmt.Errorf("install %s skill: %w", agent.Name, err)
			}
			installedSkillDirs[skillDir] = true
			results = append(results, setupResult{Agent: agent.Name, SkillDir: skillDir})
		}
		mcpPath, err := setupMCPForAgent(agent, opts.scope, opts.cwd, opts.apiKey)
		if err != nil {
			return nil, fmt.Errorf("install %s MCP: %w", agent.Name, err)
		}
		upsertRegistryInstall(&registry, registryEntryFor(agent.Name, opts.scope, opts.cwd))
		results = append(results, setupResult{Agent: agent.Name, MCPPath: mcpPath})
	}
	if err := saveInstallRegistry(registry); err != nil {
		return nil, err
	}
	return results, nil
}

func runSync(apiKey string) ([]setupResult, error) {
	registry, err := loadInstallRegistry()
	if err != nil {
		return nil, err
	}
	if len(registry.Installs) == 0 {
		return nil, nil
	}
	bundleDir, err := materializeRetabSkillBundle()
	if err != nil {
		return nil, err
	}
	agents := supportedSetupAgents()
	installedSkillDirs := map[string]bool{}
	installedUniversalSkills := map[string]bool{}
	var results []setupResult
	for _, entry := range registry.Installs {
		cwd := entry.CWD
		if cwd == "" {
			cwd = mustGetwd()
		}
		universalKey := string(entry.Scope) + ":" + cwd
		if !installedUniversalSkills[universalKey] {
			universalSkillDir, err := setupUniversalSkill(entry.Scope, cwd, bundleDir)
			if err != nil {
				return nil, fmt.Errorf("install universal skill: %w", err)
			}
			installedUniversalSkills[universalKey] = true
			installedSkillDirs[universalSkillDir] = true
			results = append(results, setupResult{Agent: "universal", SkillDir: universalSkillDir})
		}
		agent, ok := agents[entry.Agent]
		if !ok {
			return nil, fmt.Errorf("unsupported tracked agent %q", entry.Agent)
		}
		skillBase, err := skillBaseDir(agent, entry.Scope, cwd)
		if err != nil {
			return nil, err
		}
		skillDir := filepath.Join(skillBase, retabSkillInstallName)
		if !installedSkillDirs[skillDir] {
			skillDir, err = setupSkillForAgent(agent, entry.Scope, cwd, bundleDir)
			if err != nil {
				return nil, fmt.Errorf("install %s skill: %w", agent.Name, err)
			}
			installedSkillDirs[skillDir] = true
			results = append(results, setupResult{Agent: agent.Name, SkillDir: skillDir})
		}
		mcpPath, err := setupMCPForAgent(agent, entry.Scope, cwd, apiKey)
		if err != nil {
			return nil, fmt.Errorf("install %s MCP: %w", agent.Name, err)
		}
		upsertRegistryInstall(&registry, registryEntryFor(agent.Name, entry.Scope, cwd))
		results = append(results, setupResult{Agent: agent.Name, MCPPath: mcpPath})
	}
	if err := saveInstallRegistry(registry); err != nil {
		return nil, err
	}
	return results, nil
}

func resolveSetupAgents(requested []string, scope installScope) ([]setupAgent, error) {
	all := supportedSetupAgents()
	if len(requested) == 0 {
		var agents []setupAgent
		for _, agent := range all {
			if scope == installScopeLocal && !agent.SupportsLocalMCP {
				continue
			}
			agents = append(agents, agent)
		}
		return agents, nil
	}
	seen := map[string]bool{}
	var agents []setupAgent
	for _, raw := range requested {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" || seen[name] {
			continue
		}
		agent, ok := all[name]
		if !ok {
			return nil, fmt.Errorf("unsupported agent %q (want: %s)", raw, strings.Join(setupAgentNames(all), ", "))
		}
		if scope == installScopeLocal && !agent.SupportsLocalMCP {
			return nil, fmt.Errorf("%s does not support local MCP setup", name)
		}
		seen[name] = true
		agents = append(agents, agent)
	}
	return agents, nil
}

func supportedSetupAgents() map[string]setupAgent {
	homePath := func(parts ...string) func() (string, error) {
		return func() (string, error) {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(append([]string{home}, parts...)...), nil
		}
	}
	configPath := func(parts ...string) func() (string, error) {
		return func() (string, error) {
			return filepath.Join(append([]string{xdgConfigHome()}, parts...)...), nil
		}
	}
	claudeHome := func() (string, error) {
		if value := strings.TrimSpace(os.Getenv("CLAUDE_CONFIG_DIR")); value != "" {
			return value, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude"), nil
	}
	codexHome := func() (string, error) {
		if value := strings.TrimSpace(os.Getenv("CODEX_HOME")); value != "" {
			return value, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".codex"), nil
	}
	return map[string]setupAgent{
		"claude-code": {
			Name: "claude-code", Label: "Claude Code",
			GlobalSkillDir: func() (string, error) { dir, err := claudeHome(); return filepath.Join(dir, "skills"), err },
			LocalSkillDir:  func(cwd string) string { return filepath.Join(cwd, ".claude", "skills") },
			GlobalMCPPath:  homePath(".claude.json"),
			LocalMCPPath:   func(cwd string) string { return filepath.Join(cwd, ".mcp.json") },
			MCPFormat:      mcpConfigJSON, MCPKey: "mcpServers", SupportsLocalMCP: true,
		},
		"codex": {
			Name: "codex", Label: "Codex",
			GlobalSkillDir: func() (string, error) { dir, err := codexHome(); return filepath.Join(dir, "skills"), err },
			LocalSkillDir:  func(cwd string) string { return filepath.Join(cwd, ".agents", "skills") },
			GlobalMCPPath:  func() (string, error) { dir, err := codexHome(); return filepath.Join(dir, "config.toml"), err },
			LocalMCPPath:   func(cwd string) string { return filepath.Join(cwd, ".codex", "config.toml") },
			MCPFormat:      mcpConfigTOML, MCPKey: "mcp_servers", SupportsLocalMCP: true,
		},
		"cursor": {
			Name: "cursor", Label: "Cursor",
			GlobalSkillDir: homePath(".cursor", "skills"),
			LocalSkillDir:  func(cwd string) string { return filepath.Join(cwd, ".agents", "skills") },
			GlobalMCPPath:  homePath(".cursor", "mcp.json"),
			LocalMCPPath:   func(cwd string) string { return filepath.Join(cwd, ".cursor", "mcp.json") },
			MCPFormat:      mcpConfigJSON, MCPKey: "mcpServers", SupportsLocalMCP: true,
		},
		"opencode": {
			Name: "opencode", Label: "OpenCode",
			GlobalSkillDir: configPath("opencode", "skills"),
			LocalSkillDir:  func(cwd string) string { return filepath.Join(cwd, ".agents", "skills") },
			GlobalMCPPath:  configPath("opencode", "opencode.json"),
			LocalMCPPath:   func(cwd string) string { return filepath.Join(cwd, "opencode.json") },
			MCPFormat:      mcpConfigJSON, MCPKey: "mcp", TransformMCP: opencodeMCPConfig, SupportsLocalMCP: true,
		},
		"windsurf": {
			Name: "windsurf", Label: "Windsurf",
			GlobalSkillDir: homePath(".codeium", "windsurf", "skills"),
			LocalSkillDir:  func(cwd string) string { return filepath.Join(cwd, ".windsurf", "skills") },
			GlobalMCPPath:  homePath(".codeium", "windsurf", "mcp_config.json"),
			MCPFormat:      mcpConfigJSON, MCPKey: "mcpServers", SupportsLocalMCP: false,
		},
	}
}

func setupAgentNames(agents map[string]setupAgent) []string {
	names := make([]string, 0, len(agents))
	for name := range agents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func materializeRetabSkillBundle() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	skillsRoot := filepath.Join(dir, retabSkillBundleRootPath, "skills")
	if err := os.RemoveAll(skillsRoot); err != nil {
		return "", err
	}
	skillDir := filepath.Join(skillsRoot, retabSkillInstallName)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return "", err
	}
	raw, err := embeddedRetabSkill.ReadFile("embedded_skills/retab/SKILL.md")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), raw, 0o644); err != nil {
		return "", err
	}
	return skillDir, nil
}

func setupSkillForAgent(agent setupAgent, scope installScope, cwd string, bundleSkillDir string) (string, error) {
	baseDir, err := skillBaseDir(agent, scope, cwd)
	if err != nil {
		return "", err
	}
	return setupSkillInBaseDir(baseDir, bundleSkillDir)
}

func setupUniversalSkill(scope installScope, cwd string, bundleSkillDir string) (string, error) {
	baseDir, err := universalSkillBaseDir(scope, cwd)
	if err != nil {
		return "", err
	}
	return setupSkillInBaseDir(baseDir, bundleSkillDir)
}

func setupSkillInBaseDir(baseDir string, bundleSkillDir string) (string, error) {
	target := filepath.Join(baseDir, retabSkillInstallName)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", err
	}
	if err := os.RemoveAll(target); err != nil {
		return "", err
	}
	relative, err := filepath.Rel(baseDir, bundleSkillDir)
	if err == nil && runtime.GOOS != "windows" {
		if err := os.Symlink(relative, target); err == nil {
			return target, nil
		}
	}
	if err := copyDir(bundleSkillDir, target); err != nil {
		return "", err
	}
	return target, nil
}

func universalSkillBaseDir(scope installScope, cwd string) (string, error) {
	if scope == installScopeLocal {
		return filepath.Join(cwd, ".agents", "skills"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agents", "skills"), nil
}

func skillBaseDir(agent setupAgent, scope installScope, cwd string) (string, error) {
	if scope == installScopeLocal {
		return agent.LocalSkillDir(cwd), nil
	}
	return agent.GlobalSkillDir()
}

func setupMCPForAgent(agent setupAgent, scope installScope, cwd string, apiKey string) (string, error) {
	path, err := mcpConfigPath(agent, scope, cwd)
	if err != nil {
		return "", err
	}
	retabConfig := mcpServerConfig{Type: "streamable-http", URL: retabMCPURL}
	if apiKey != "" {
		retabConfig.Headers = map[string]string{"Authorization": "Bearer " + apiKey}
	}
	if err := writeMCPServer(agent, path, retabMCPServerName, retabConfig); err != nil {
		return "", err
	}
	// The Retab docs MCP is the public Mintlify docs server; it carries no
	// API key because the documentation is unauthenticated.
	docsConfig := mcpServerConfig{Type: "streamable-http", URL: retabDocsMCPURL}
	if err := writeMCPServer(agent, path, retabDocsMCPServerName, docsConfig); err != nil {
		return "", err
	}
	return path, nil
}

func writeMCPServer(agent setupAgent, path string, serverName string, config mcpServerConfig) error {
	var value any = config
	if agent.TransformMCP != nil {
		value = agent.TransformMCP(config)
	}
	switch agent.MCPFormat {
	case mcpConfigJSON:
		return upsertJSONMCPConfig(path, agent.MCPKey, serverName, value)
	case mcpConfigTOML:
		return upsertTomlMCPConfig(path, serverName, config)
	default:
		return fmt.Errorf("unsupported MCP config format %q", agent.MCPFormat)
	}
}

func mcpConfigPath(agent setupAgent, scope installScope, cwd string) (string, error) {
	if scope == installScopeLocal {
		if agent.LocalMCPPath == nil {
			return "", fmt.Errorf("%s has no local MCP config path", agent.Name)
		}
		return agent.LocalMCPPath(cwd), nil
	}
	return agent.GlobalMCPPath()
}

func opencodeMCPConfig(config mcpServerConfig) any {
	value := map[string]any{
		"type":    "remote",
		"url":     config.URL,
		"enabled": true,
	}
	if len(config.Headers) > 0 {
		value["headers"] = config.Headers
	}
	return value
}

func upsertJSONMCPConfig(path string, key string, serverName string, serverConfig any) error {
	existing := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil && strings.TrimSpace(string(raw)) != "" {
		if err := json.Unmarshal(raw, &existing); err != nil {
			return fmt.Errorf("parse %s as JSON: %w", path, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	servers, _ := existing[key].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[serverName] = serverConfig
	existing[key] = servers
	raw, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return writeFileCreatingParents(path, append(raw, '\n'), 0o644)
}

func upsertTomlMCPConfig(path string, serverName string, config mcpServerConfig) error {
	var existing string
	if raw, err := os.ReadFile(path); err == nil {
		existing = string(raw)
	} else if !os.IsNotExist(err) {
		return err
	}
	existing = removeTomlSection(existing, "mcp_servers."+serverName)
	var block strings.Builder
	if strings.TrimSpace(existing) != "" {
		block.WriteString(strings.TrimRight(existing, "\n"))
		block.WriteString("\n\n")
	}
	block.WriteString("[mcp_servers.")
	block.WriteString(serverName)
	block.WriteString("]\n")
	block.WriteString(`type = "`)
	block.WriteString(escapeTomlString(config.Type))
	block.WriteString("\"\n")
	block.WriteString(`url = "`)
	block.WriteString(escapeTomlString(config.URL))
	block.WriteString("\"\n")
	if len(config.Headers) > 0 {
		block.WriteString("\n[mcp_servers.")
		block.WriteString(serverName)
		block.WriteString(".headers]\n")
		keys := make([]string, 0, len(config.Headers))
		for key := range config.Headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			block.WriteString(tomlKey(key))
			block.WriteString(` = "`)
			block.WriteString(escapeTomlString(config.Headers[key]))
			block.WriteString("\"\n")
		}
	}
	return writeFileCreatingParents(path, []byte(block.String()), 0o644)
}

func removeTomlSection(raw string, section string) string {
	lines := strings.Split(raw, "\n")
	header := "[" + section + "]"
	prefix := "[" + section + "."
	var kept []string
	skip := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if trimmed == header || strings.HasPrefix(trimmed, prefix) {
				skip = true
				continue
			}
			skip = false
		}
		if !skip {
			kept = append(kept, line)
		}
	}
	return strings.TrimRight(strings.Join(kept, "\n"), "\n")
}

func escapeTomlString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func tomlKey(key string) string {
	if key == "" {
		return `""`
	}
	for _, r := range key {
		if r != '_' && r != '-' && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return `"` + escapeTomlString(key) + `"`
		}
	}
	return key
}

func registryPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "install-registry.json"), nil
}

func loadInstallRegistry() (installRegistry, error) {
	registry := installRegistry{Version: retabInstallRegistryV1}
	path, err := registryPath()
	if err != nil {
		return registry, err
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return registry, nil
	}
	if err != nil {
		return registry, err
	}
	if strings.TrimSpace(string(raw)) == "" {
		return registry, nil
	}
	if err := json.Unmarshal(raw, &registry); err != nil {
		return registry, fmt.Errorf("parse %s: %w", path, err)
	}
	if registry.Version == 0 {
		registry.Version = retabInstallRegistryV1
	}
	return registry, nil
}

func saveInstallRegistry(registry installRegistry) error {
	path, err := registryPath()
	if err != nil {
		return err
	}
	if registry.Version == 0 {
		registry.Version = retabInstallRegistryV1
	}
	raw, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	return writeFileCreatingParents(path, append(raw, '\n'), 0o644)
}

func registryEntryFor(agent string, scope installScope, cwd string) installRegistryEntry {
	now := time.Now().UTC().Format(time.RFC3339)
	entry := installRegistryEntry{
		ID:           registryInstallID(agent, scope, cwd),
		Agent:        agent,
		Scope:        scope,
		Installer:    retabInstallerName,
		RetabVersion: version,
		InstalledAt:  now,
		UpdatedAt:    now,
	}
	if scope == installScopeLocal {
		entry.CWD = cwd
	}
	return entry
}

func registryInstallID(agent string, scope installScope, cwd string) string {
	if scope == installScopeGlobal {
		return "global:" + agent
	}
	return "local:" + agent + ":" + cwd
}

func upsertRegistryInstall(registry *installRegistry, entry installRegistryEntry) {
	for i := range registry.Installs {
		if registry.Installs[i].ID == entry.ID {
			entry.InstalledAt = registry.Installs[i].InstalledAt
			registry.Installs[i] = entry
			return
		}
	}
	registry.Installs = append(registry.Installs, entry)
	sort.Slice(registry.Installs, func(i, j int) bool {
		return registry.Installs[i].ID < registry.Installs[j].ID
	})
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(src string, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func writeFileCreatingParents(path string, raw []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, perm)
}

func xdgConfigHome() string {
	if value := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config"
	}
	return filepath.Join(home, ".config")
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
