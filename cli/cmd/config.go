package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// retabConfig is the on-disk shape at ~/.retab/config.json.
//
// Two auth shapes are supported, and the file may hold both at once. At
// request time the CLI prefers OAuth tokens when present; the legacy
// `api_key` field is still honored so that long-standing setups don't
// break when users upgrade.
type retabConfig struct {
	// APIKey is the legacy auth path. Still fully supported.
	APIKey string `json:"api_key,omitempty"`

	// BaseURL overrides the default API endpoint. Useful for staging.
	BaseURL string `json:"base_url,omitempty"`

	// OAuth holds tokens issued by WorkOS via `retab auth login`. Optional.
	OAuth *oauthTokens `json:"oauth,omitempty"`
}

// oauthTokens is the persisted OAuth state. Mirrors the WorkOS token
// endpoint response, plus an absolute expiry computed at write time so
// the CLI can decide whether to refresh without re-reading clock skew.
type oauthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"` // "Bearer"
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`

	// Echoed from the /v1/auth/cli/config discovery call at login time so
	// that the refresh path doesn't need to re-discover.
	AuthKitDomain string `json:"authkit_domain,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".retab"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func loadConfig() (retabConfig, error) {
	var cfg retabConfig
	path, err := configPath()
	if err != nil {
		return cfg, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("invalid %s: %w", path, err)
	}
	return cfg, nil
}

func saveConfig(cfg retabConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func deleteConfig() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
