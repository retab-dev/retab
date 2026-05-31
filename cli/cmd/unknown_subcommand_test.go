package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// Regression guard: an unknown subcommand of a router command must fail
// loudly. Cobra's built-in unknown-command detection only fires for the
// root, so nested routers (`files`, `workflows runs`, ...) used
// to print help and exit 0 — silently swallowing typos in scripts.
//
// These tests drive the real command tree through Execute() (which
// applies hardenGroupCommands) and assert the error surfaces.

func runRootForTest(t *testing.T, args ...string) error {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetArgs(args)
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
	return Execute()
}

func TestUnknownSubcommandFailsOnRouters(t *testing.T) {
	cases := [][]string{
		{"files", "bogus"},
		{"workflows", "runs", "bogus"},
		{"workflows", "bogus"},
		{"parses", "bogus"},
		{"auth", "bogus"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := runRootForTest(t, args...)
			if err == nil {
				t.Fatalf("retab %s: expected an error for an unknown subcommand, got nil (would exit 0)", strings.Join(args, " "))
			}
			if !strings.Contains(err.Error(), "unknown command") {
				t.Fatalf("retab %s: expected an \"unknown command\" error, got: %v", strings.Join(args, " "), err)
			}
		})
	}
}

func TestUnknownRootCommandStillFails(t *testing.T) {
	err := runRootForTest(t, "bogus")
	if err == nil {
		t.Fatal("retab bogus: expected an error for an unknown command, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("retab bogus: expected an \"unknown command\" error, got: %v", err)
	}
}

func TestBareRouterPrintsHelpWithoutError(t *testing.T) {
	// A router invoked with no subcommand should still print help and
	// exit 0 — only *unknown* subcommands are an error.
	for _, router := range []string{"files", "workflows"} {
		t.Run(router, func(t *testing.T) {
			if err := runRootForTest(t, router); err != nil {
				t.Fatalf("retab %s: bare router should not error, got: %v", router, err)
			}
		})
	}
}
