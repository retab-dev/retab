package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		// Execute parses args onto the SHARED global command tree, and the
		// parsed values — including each flag's Changed bit — persist after the
		// run. Resetting only a handful of root persistent flags left local
		// subcommand flags (e.g. the required `project-id` on `tables create`)
		// leaking their Changed=true state into the next test, which then either
		// skipped a required-flag check or read a stale value depending on test
		// order. Reset the WHOLE tree so every runRootForTest starts clean.
		resetCommandTreeFlags(rootCmd)
	})
	return ExecuteArgs(args)
}

// setFlagClean sets a flag's value on a shared command singleton without
// leaving the Changed bit set. pflag's FlagSet.Set marks the flag Changed, so
// restoring a flag with Set(name, "") leaves Changed=true — which a later test
// inspecting Changed(name) misreads as "the user passed --name". Setting the
// underlying Value directly and clearing Changed restores a truly pristine
// flag. A missing flag is a no-op.
func setFlagClean(cmd *cobra.Command, name, value string) {
	f := cmd.Flags().Lookup(name)
	if f == nil {
		return
	}
	_ = f.Value.Set(value)
	f.Changed = false
}

// resetCommandTreeFlags restores every flag on cmd and its descendants to its
// default value and clears the Changed bit, so parsed state from one test can't
// bleed into the next through the shared global command tree. Slice flags are
// cleared via SliceValue.Replace (Set(DefValue) would append, not reset).
func resetCommandTreeFlags(cmd *cobra.Command) {
	reset := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				_ = sv.Replace(nil)
			} else {
				_ = f.Value.Set(f.DefValue)
			}
			f.Changed = false
		})
	}
	reset(cmd.Flags())
	reset(cmd.PersistentFlags())
	for _, sub := range cmd.Commands() {
		resetCommandTreeFlags(sub)
	}
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

func TestUnknownSubcommandWithHelpFailsOnRouters(t *testing.T) {
	cases := [][]string{
		{"files", "delete", "--help"},
		{"workflows", "nope", "--help"},
		{"workflows", "runs", "nope", "--help"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := runRootForTest(t, args...)
			if err == nil {
				t.Fatalf("retab %s: expected an error for an unknown subcommand before --help, got nil", strings.Join(args, " "))
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

func TestLeafHelpAfterPositionalArgSucceeds(t *testing.T) {
	// `retab <leaf> <id> --help` must print the leaf's help and exit 0, not
	// error with `unknown command "<id>"`. The pre-Execute help-path walk used
	// to treat the positional id as a subcommand of a leaf and reject it.
	cases := [][]string{
		{"workflows", "get", "wf_123", "--help"},
		{"parses", "get", "some_id", "--help"},
		{"parses", "wait", "some_id", "-h"},
		{"splits", "delete", "split_x", "--help"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if err := runRootForTest(t, args...); err != nil {
				t.Fatalf("retab %s: leaf help after a positional arg should not error, got: %v", strings.Join(args, " "), err)
			}
		})
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

func TestUnicodeDashHelpFlagIsNormalized(t *testing.T) {
	if err := runRootForTest(t, "workflows", "—help"); err != nil {
		t.Fatalf("retab workflows —help should behave like --help, got: %v", err)
	}
}
