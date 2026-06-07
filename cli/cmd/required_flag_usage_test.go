//go:build !retab_oagen_cli_workflows_tests

package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Cobra's MarkFlagRequired enforces a flag at parse time but does NOT
// append anything to the rendered help. This CLI's convention is to
// hand-write " (required)" into the usage string of every required flag
// (see `--model`, `--instructions`, `--name`, ...). When the two drift
// apart users see a flag with no "(required)" marker, run the command
// without it, and get a parse error the help never warned them about.
//
// Walk every registered command and assert: if a flag carries cobra's
// required annotation, its usage text mentions "(required)". This pins
// the convention so a future MarkFlagRequired call can't silently ship
// without the matching help text.
func TestRequiredFlagsAdvertiseRequiredInUsage(t *testing.T) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if _, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; !ok {
				return
			}
			if !strings.Contains(f.Usage, "(required)") {
				t.Errorf("command %q flag --%s is MarkFlagRequired'd but its help text %q does not contain \"(required)\"",
					c.CommandPath(), f.Name, f.Usage)
			}
		})
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}

// Some flags are required but enforced inside RunE rather than via
// MarkFlagRequired (e.g. when several flags are jointly required, or
// when "-" stdin handling rules out cobra's required-flag machinery).
// TestRequiredFlagsAdvertiseRequiredInUsage cannot see those, so pin
// the ones we know about explicitly: their help text must still tell
// the user they are required, exactly like the MarkFlagRequired'd ones.
func TestRunEEnforcedRequiredFlagsAdvertiseRequiredInUsage(t *testing.T) {
	cases := []struct {
		cmd   *cobra.Command
		flags []string
	}{
		// workflows tests create rejects a missing target/source/
		// assertion file in RunE with "... are required".
		{workflowsTestsCreateCmd, []string{"target-file", "source-file", "assertion-file"}},
	}
	for _, tc := range cases {
		for _, name := range tc.flags {
			f := tc.cmd.Flags().Lookup(name)
			if f == nil {
				t.Errorf("command %q has no flag --%s", tc.cmd.CommandPath(), name)
				continue
			}
			if !strings.Contains(f.Usage, "(required)") {
				t.Errorf("command %q flag --%s is required (enforced in RunE) but its help text %q does not contain \"(required)\"",
					tc.cmd.CommandPath(), name, f.Usage)
			}
		}
	}
}
