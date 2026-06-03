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
// This is "config v2": it adds named customer-environment profiles and a
// local default on top of the original legacy fields. The file may hold
// any mix of the shapes at once, and old config files (legacy `api_key`
// only, or `oauth` only) keep loading and behaving exactly as before.
//
// Auth-shape precedence is resolved by resolveCredential (see common.go);
// this struct is purely the storage format.
type retabConfig struct {
	// Version marks the config schema. Absent / 0 means a legacy file
	// written by an older CLI; the loader still accepts it. New writes
	// stamp configVersion so future migrations can branch on it.
	Version int `json:"version,omitempty"`

	// APIKey is the legacy auth path. Still fully supported — treated as
	// a production-scoped credential during the environments rollout.
	APIKey string `json:"api_key,omitempty"`

	// BaseURL overrides the default API endpoint. This selects a Retab
	// *deployment* (local dev, staging API host); it is NOT the customer
	// environment selector — that is determined by the API key.
	BaseURL string `json:"base_url,omitempty"`

	// EnvironmentID is the dashboard/API environment selected for OAuth
	// sessions. OAuth-backed API requests mint a short-lived Retab dashboard
	// context token for this environment. API keys are already
	// environment-scoped server-side, so this is primarily for user sessions
	// and local development workflows.
	EnvironmentID string `json:"environment_id,omitempty"`

	// OAuth holds tokens issued by WorkOS via `retab auth login`. Optional.
	OAuth *oauthTokens `json:"oauth,omitempty"`

	// Environments holds named customer-environment profiles, keyed by
	// slug ("test", "production", "staging", ...). Absent on legacy files,
	// in which case the CLI behaves exactly as it did pre-v2.
	Environments map[string]*environmentProfile `json:"environments,omitempty"`

	// DefaultEnvironment is the slug of the profile used when no --env,
	// --live, --api-key, or RETAB_API_KEY override is supplied.
	DefaultEnvironment string `json:"default_environment,omitempty"`
}

// configVersion is stamped onto every config file the v2 CLI writes.
const configVersion = 2

// environmentProfile is one named local credential profile. The slug is
// the map key, not a field. These are CLI-local names — the server
// decides the real environment from the API key document.
type environmentProfile struct {
	// Name is a display-only label ("Test", "Production", "QA").
	Name string `json:"name,omitempty"`

	// APIKey is the stored credential for this profile.
	APIKey string `json:"api_key,omitempty"`

	// APIKeyPreview is a redacted preview ("rt_test_...abcd") kept so list
	// views never need to touch the raw key.
	APIKeyPreview string `json:"api_key_preview,omitempty"`

	// ServerEnvironmentSlug / ServerEnvironmentID are filled from a
	// /v1/auth/status probe when available. Advisory only.
	ServerEnvironmentSlug string `json:"server_environment_slug,omitempty"`
	ServerEnvironmentID   string `json:"server_environment_id,omitempty"`

	// BaseURL optionally pins a Retab deployment for this profile
	// (advanced / local-dev use only).
	BaseURL string `json:"base_url,omitempty"`

	CreatedAt      string `json:"created_at,omitempty"`
	LastUsedAt     string `json:"last_used_at,omitempty"`
	LastVerifiedAt string `json:"last_verified_at,omitempty"`
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
	// Stamp the current schema version on every write so a file touched
	// by the v2 CLI is self-identifying. Legacy reads still tolerate a
	// missing version, so this never breaks downgrades that ignore it.
	cfg.Version = configVersion
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
