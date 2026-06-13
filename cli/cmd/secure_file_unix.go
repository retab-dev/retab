//go:build !windows

package cmd

import "os"

// secureConfigFile ensures path is readable/writable by its owner only (0600).
// saveConfig already chmods the temp file before publishing it; this is a
// cheap, defensive re-assert that also documents the cross-platform contract
// implemented differently on Windows.
func secureConfigFile(path string) error {
	return os.Chmod(path, 0o600)
}
