package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// `env add` stores a profile, makes the first one the default, and folds
// the `live` alias onto the `production` slug.
func TestEnvAdd_FirstProfileBecomesDefault(t *testing.T) {
	isolateHome(t)
	envAddCmd.Flags().Set("api-key", "rt_test_abc")
	t.Cleanup(func() { envAddCmd.Flags().Set("api-key", "") })

	if err := envAddCmd.RunE(envAddCmd, []string{"test"}); err != nil {
		t.Fatal(err)
	}
	cfg, _ := loadConfig()
	if cfg.Environments["test"] == nil || cfg.Environments["test"].APIKey != "rt_test_abc" {
		t.Fatalf("test profile not stored: %+v", cfg.Environments)
	}
	if cfg.DefaultEnvironment != "test" {
		t.Errorf("first profile should become default, got %q", cfg.DefaultEnvironment)
	}
	// The raw key must not be the stored preview.
	if cfg.Environments["test"].APIKeyPreview == "rt_test_abc" {
		t.Error("APIKeyPreview should be redacted, not the raw key")
	}
}

func TestEnvAdd_LiveAliasStoresProductionSlug(t *testing.T) {
	isolateHome(t)
	envAddCmd.Flags().Set("api-key", "rt_live_xyz")
	t.Cleanup(func() { envAddCmd.Flags().Set("api-key", "") })

	if err := envAddCmd.RunE(envAddCmd, []string{"live"}); err != nil {
		t.Fatal(err)
	}
	cfg, _ := loadConfig()
	if cfg.Environments[slugProduction] == nil {
		t.Fatalf("`live` alias should store under production slug: %+v", cfg.Environments)
	}
	if cfg.Environments["live"] != nil {
		t.Error("should not store a literal `live` slug")
	}
}

func TestEnvAdd_RejectsInvalidSlug(t *testing.T) {
	isolateHome(t)
	envAddCmd.Flags().Set("api-key", "rt_test_abc")
	t.Cleanup(func() { envAddCmd.Flags().Set("api-key", "") })

	err := envAddCmd.RunE(envAddCmd, []string{"Bad Slug"})
	if err == nil || !strings.Contains(err.Error(), "invalid environment slug") {
		t.Fatalf("expected invalid-slug error, got %v", err)
	}
}

// `env switch` changes the default but errors if the profile is missing.
func TestEnvSwitch(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		DefaultEnvironment: "test",
		Environments: map[string]*environmentProfile{
			"test":         {APIKey: "rt_test_abc"},
			slugProduction: {APIKey: "rt_live_xyz"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	if err := envSwitchCmd.RunE(cmd, []string{"production"}); err != nil {
		t.Fatal(err)
	}
	cfg, _ := loadConfig()
	if cfg.DefaultEnvironment != slugProduction {
		t.Errorf("DefaultEnvironment = %q, want production", cfg.DefaultEnvironment)
	}

	err := envSwitchCmd.RunE(cmd, []string{"staging"})
	if err == nil || !strings.Contains(err.Error(), "staging") {
		t.Fatalf("expected missing-profile error, got %v", err)
	}
}

// `env remove` deletes the profile and clears the default when it pointed
// at the removed profile.
func TestEnvRemove_ClearsDefault(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		DefaultEnvironment: "test",
		Environments: map[string]*environmentProfile{
			"test": {APIKey: "rt_test_abc"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	if err := envRemoveCmd.RunE(cmd, []string{"test"}); err != nil {
		t.Fatal(err)
	}
	cfg, _ := loadConfig()
	if cfg.Environments["test"] != nil {
		t.Error("profile not removed")
	}
	if cfg.DefaultEnvironment != "" {
		t.Errorf("default should be cleared, got %q", cfg.DefaultEnvironment)
	}
}

// `env list` on a non-TTY writer emits JSON with a `data` array, never the
// raw key.
func TestWriteEnvList_JSONForNonTTY(t *testing.T) {
	cfg := retabConfig{
		DefaultEnvironment: "test",
		Environments: map[string]*environmentProfile{
			"test": {Name: "Test", APIKey: "rt_test_abcd1234", APIKeyPreview: "rt_t...1234"},
		},
	}
	var buf bytes.Buffer
	if err := writeEnvList(&buf, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"data"`) {
		t.Errorf("non-TTY env list should be JSON with data array:\n%s", out)
	}
	if strings.Contains(out, "rt_test_abcd1234") {
		t.Errorf("raw API key leaked into env list output:\n%s", out)
	}
}

func TestWriteEnvList_EmptyHumanHint(t *testing.T) {
	var buf bytes.Buffer
	// bytes.Buffer is non-TTY → JSON; an empty config yields an empty data
	// array rather than the human hint, which is fine for scripts.
	if err := writeEnvList(&buf, retabConfig{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"data"`) {
		t.Errorf("empty env list should still emit a data array:\n%s", buf.String())
	}
}

// decodeAuthStatus maps the server payload into the typed struct.
func TestDecodeAuthStatus(t *testing.T) {
	raw := map[string]any{
		"authenticated":    true,
		"organization_id":  "org_1",
		"environment_id":   "env_1",
		"environment_slug": "test",
		"auth_method":      "api_key",
		"region":           "eu",
		"key_prefix":       "rt_test",
		"key_name":         "CI key",
	}
	status := decodeAuthStatus(raw)
	if !status.Authenticated {
		t.Error("Authenticated should be true")
	}
	if status.EnvironmentSlug != "test" {
		t.Errorf("EnvironmentSlug = %q", status.EnvironmentSlug)
	}
	if status.KeyPrefix != "rt_test" {
		t.Errorf("KeyPrefix = %q", status.KeyPrefix)
	}
	if status.OrganizationID != "org_1" {
		t.Errorf("OrganizationID = %q", status.OrganizationID)
	}

	// Defensive: a non-object payload yields a zero struct, not a panic.
	if got := decodeAuthStatus("not-an-object"); got.Authenticated {
		t.Error("non-object payload should decode to zero struct")
	}
}
