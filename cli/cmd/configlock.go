package cmd

import (
	"os"
	"path/filepath"
)

// configLockPath is the advisory lock file that guards read-modify-write
// cycles on config.json across concurrent `retab` processes. It sits next to
// config.json in ~/.retab.
func configLockPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.lock"), nil
}

// withConfigLock runs fn while holding an exclusive cross-process lock on the
// config directory, serializing read-modify-write cycles on config.json — most
// importantly the OAuth refresh-token rotation persist, where two concurrent
// invocations could otherwise clobber each other's rotated refresh_token.
//
// The lock is advisory and strictly best-effort: if it cannot be created or
// acquired (unsupported filesystem, permission issue), fn still runs unlocked
// so locking never blocks or fails legitimate work. fn's error is returned
// verbatim.
func withConfigLock(fn func() error) error {
	path, err := configLockPath()
	if err != nil {
		return fn()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fn()
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fn()
	}
	defer f.Close()
	if err := lockFileExclusive(f); err != nil {
		// Could not acquire the OS lock; proceed unlocked rather than block
		// or fail. Worst case we fall back to the prior (unserialized)
		// behavior.
		return fn()
	}
	defer func() { _ = unlockFile(f) }()
	return fn()
}
