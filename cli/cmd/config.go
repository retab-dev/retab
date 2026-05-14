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

// saveConfig writes cfg to ~/.retab/config.json *atomically*.
//
// Atomicity matters here because the file holds the only copy of the
// rotated refresh_token on this machine. WorkOS rotates refresh tokens on
// every refresh and invalidates the previous one; if a non-atomic write
// is interrupted (signal, crash, disk full) the file can end up empty or
// truncated, the old refresh_token on disk is already dead server-side,
// and the user is silently logged out until they re-run `auth login`.
//
// The strategy is the canonical temp-file + fsync + rename: rename(2) is
// atomic within a filesystem on POSIX, so either the new contents fully
// land or the old contents remain. fsync on the temp file ensures the
// data is durable before the rename publishes it.
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

	// Use os.CreateTemp in the SAME directory so rename is intra-fs (atomic).
	// The pattern keeps a stable prefix so a half-written file is easy to
	// spot during incident triage.
	tmp, err := os.CreateTemp(dir, "config.json.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	// Best-effort cleanup if we error out before the rename.
	cleanupTmp := func() { _ = os.Remove(tmpPath) }

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		cleanupTmp()
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		cleanupTmp()
		return err
	}
	// Sync before rename: rename publishes whatever bytes are in the inode,
	// so the data needs to be durable first or a power cut between sync()
	// and the rename's metadata flush could still leave a zero-byte file.
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanupTmp()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanupTmp()
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanupTmp()
		return err
	}
	return nil
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
