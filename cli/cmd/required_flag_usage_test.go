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

// `workflows tests create` accepts each of target/source/assertion as EITHER
// a JSON file OR an inline flag form, so none of the file flags is
// unconditionally required. The convention here is the inverse of
// TestRequiredFlagsAdvertiseRequiredInUsage: the file flags must NOT claim
// "(required)" (that would mislead users away from the inline form), and each
// must point at its inline alternative so `--help` documents both paths.
func TestWorkflowsTestsCreateFileFlagsAdvertiseInlineAlternative(t *testing.T) {
	cases := []struct {
		flag        string
		alternative string
	}{
		{"target-file", "--block-id"},
		{"source-file", "--run-id"},
		{"assertion-file", "--equals"},
	}
	for _, tc := range cases {
		f := workflowsTestsCreateCmd.Flags().Lookup(tc.flag)
		if f == nil {
			t.Errorf("workflows tests create has no flag --%s", tc.flag)
			continue
		}
		if strings.Contains(f.Usage, "(required)") {
			t.Errorf("flag --%s is no longer unconditionally required (an inline form exists) but its help text %q still says \"(required)\"",
				tc.flag, f.Usage)
		}
		if !strings.Contains(f.Usage, tc.alternative) {
			t.Errorf("flag --%s help %q should point at its inline alternative %q", tc.flag, f.Usage, tc.alternative)
		}
	}
}
