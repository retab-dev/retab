package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeRawConfig writes raw JSON bytes directly to ~/.retab/config.json,
// bypassing saveConfig — so an test can simulate a config file produced by
// an older CLI version.
func writeRawConfig(t *testing.T, home, contents string) {
	t.Helper()
	dir := filepath.Join(home, ".retab")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

// A legacy config file with only a top-level api_key must keep loading and
// resolve as a production-scoped credential — no `version`, no
// `environments`, no `default_environment`.
func TestLoadConfig_LegacyAPIKeyOnly(t *testing.T) {
	home := isolateHome(t)
	writeRawConfig(t, home, `{"api_key": "sk_retab_legacy", "base_url": "https://api.test/v1"}`)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("legacy config should load, got %v", err)
	}
	if cfg.APIKey != "sk_retab_legacy" {
		t.Errorf("APIKey = %q, want sk_retab_legacy", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.test/v1" {
		t.Errorf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.Version != 0 {
		t.Errorf("legacy file should have Version 0, got %d", cfg.Version)
	}
	if cfg.Environments != nil {
		t.Errorf("legacy file should have no Environments, got %+v", cfg.Environments)
	}

	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatalf("resolveCredential on legacy config: %v", err)
	}
	if cred.Source != sourceLegacyKey {
		t.Errorf("source = %q, want %q", cred.Source, sourceLegacyKey)
	}
}

// A legacy OAuth-only config file must keep loading.
func TestLoadConfig_LegacyOAuthOnly(t *testing.T) {
	home := isolateHome(t)
	writeRawConfig(t, home, `{"oauth": {"access_token": "tok", "expires_at": "2030-01-01T00:00:00Z"}}`)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("legacy oauth config should load, got %v", err)
	}
	if cfg.OAuth == nil || cfg.OAuth.AccessToken != "tok" {
		t.Fatalf("OAuth not loaded: %+v", cfg.OAuth)
	}
}

// A v2 config file with environments + default_environment round-trips.
func TestLoadConfig_V2WithEnvironments(t *testing.T) {
	home := isolateHome(t)
	writeRawConfig(t, home, `{
		"version": 2,
		"default_environment": "staging",
		"api_key": "sk_retab_legacy",
		"environments": {
			"test": {"name": "Test", "api_key": "rt_test_abc", "api_key_preview": "rt_t...abc"},
			"staging": {"name": "Staging", "api_key": "rt_test_stg", "server_environment_slug": "test"}
		}
	}`)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("v2 config should load, got %v", err)
	}
	if cfg.Version != 2 {
		t.Errorf("Version = %d, want 2", cfg.Version)
	}
	if cfg.DefaultEnvironment != "staging" {
		t.Errorf("DefaultEnvironment = %q, want staging", cfg.DefaultEnvironment)
	}
	if len(cfg.Environments) != 2 {
		t.Fatalf("expected 2 environments, got %d", len(cfg.Environments))
	}
	if cfg.Environments["test"].APIKey != "rt_test_abc" {
		t.Errorf("test profile key = %q", cfg.Environments["test"].APIKey)
	}
	if cfg.Environments["staging"].ServerEnvironmentSlug != "test" {
		t.Errorf("staging server slug = %q", cfg.Environments["staging"].ServerEnvironmentSlug)
	}
	// The legacy api_key must survive alongside the new fields.
	if cfg.APIKey != "sk_retab_legacy" {
		t.Errorf("legacy api_key dropped: %q", cfg.APIKey)
	}
}

// saveConfig must stamp the version, preserve the legacy api_key, the
// OAuth block, profiles, and default_environment — without losing data.
func TestSaveConfig_PreservesAllFields(t *testing.T) {
	isolateHome(t)
	in := retabConfig{
		APIKey:             "sk_retab_legacy",
		AccessToken:        "acctk_production_saved",
		BaseURL:            "https://api.test/v1",
		OAuth:              &oauthTokens{AccessToken: "tok", WorkosAPIBaseURL: "https://api.workos.com"},
		DefaultEnvironment: "test",
		Environments: map[string]*environmentProfile{
			"test": {Name: "Test", APIKey: "rt_test_abc", APIKeyPreview: "rt_t...abc"},
		},
	}
	if err := saveConfig(in); err != nil {
		t.Fatal(err)
	}
	got, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != configVersion {
		t.Errorf("Version = %d, want %d", got.Version, configVersion)
	}
	if got.APIKey != "sk_retab_legacy" {
		t.Errorf("APIKey = %q", got.APIKey)
	}
	if got.AccessToken != "acctk_production_saved" {
		t.Errorf("AccessToken = %q", got.AccessToken)
	}
	if got.OAuth == nil || got.OAuth.AccessToken != "tok" {
		t.Errorf("OAuth not preserved: %+v", got.OAuth)
	}
	if got.DefaultEnvironment != "test" {
		t.Errorf("DefaultEnvironment = %q", got.DefaultEnvironment)
	}
	if got.Environments["test"].APIKey != "rt_test_abc" {
		t.Errorf("environment profile not preserved: %+v", got.Environments)
	}
}

// The config file is written at mode 0600 and its parent dir at 0700.
func TestSaveConfig_FilePermissions(t *testing.T) {
	// POSIX permission bits aren't represented on Windows/NTFS (Go reports
	// 0666/0777 there regardless of the requested mode), so this contract is
	// only meaningful on POSIX platforms.
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not represented on Windows/NTFS")
	}
	isolateHome(t)
	if err := saveConfig(retabConfig{APIKey: "rt_test_abc"}); err != nil {
		t.Fatal(err)
	}
	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("config file mode = %o, want 600", perm)
	}
	dir, err := configDir()
	if err != nil {
		t.Fatal(err)
	}
	di, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := di.Mode().Perm(); perm != 0o700 {
		t.Errorf("config dir mode = %o, want 700", perm)
	}
}

// An invalid JSON file is rejected with a clear error rather than silently
// loading an empty config.
func TestLoadConfig_InvalidJSON(t *testing.T) {
	home := isolateHome(t)
	writeRawConfig(t, home, `{not json`)
	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected error for invalid JSON config")
	}
}
