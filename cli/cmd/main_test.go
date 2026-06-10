package cmd

import (
	"fmt"
	"os"
	"testing"
)

// TestMain isolates the whole cmd test package from the developer's real
// config directory.
//
// Config paths are resolved via os.UserHomeDir(), which reads $HOME on
// unix/plan9 but %USERPROFILE% on Windows. Individual tests only set $HOME for
// isolation, so on Windows saveConfig/loadConfig would otherwise hit the real
// ~/.retab/config.json — silently overwriting a developer's selected
// environment and (worse) their stored OAuth refresh token. Repointing BOTH
// variables at a throwaway directory here closes that gap regardless of OS and
// regardless of whether a given test remembers to isolate itself.
func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

// runTests is split out so the temp-dir cleanup (deferred) still runs before
// os.Exit; os.Exit in TestMain would otherwise skip deferred calls.
func runTests(m *testing.M) int {
	home, err := os.MkdirTemp("", "retab-cli-test-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create temp HOME for tests:", err)
		return 1
	}
	defer os.RemoveAll(home)

	// Set both so the config dir resolves into the temp home on every OS.
	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)

	return m.Run()
}
