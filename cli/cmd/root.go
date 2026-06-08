package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// outputFlagValue is a pflag.Value for --output that rejects unknown
// strings at parse time. We can't do this from a cobra PreRunE because
// `--output bogus --help` shortcuts past PreRunE; pflag's Set is the
// earliest hook that runs regardless of whether help is requested.
type outputFlagValue struct{ value string }

func (o *outputFlagValue) String() string { return o.value }
func (o *outputFlagValue) Type() string   { return "string" }
func (o *outputFlagValue) Set(raw string) error {
	switch raw {
	case "", "auto", string(OutputJSON), string(OutputTable), string(OutputCSV):
		o.value = raw
		return nil
	default:
		return fmt.Errorf("invalid --output value %q (want: json | table | csv | auto)", raw)
	}
}

type outputTableFlagValue struct {
	output *outputFlagValue
}

func (o *outputTableFlagValue) String() string { return "false" }
func (o *outputTableFlagValue) Type() string   { return "bool" }
func (o *outputTableFlagValue) IsBoolFlag() bool {
	return true
}
func (o *outputTableFlagValue) Set(raw string) error {
	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		return fmt.Errorf("invalid --output-table value %q", raw)
	}
	if enabled {
		return o.output.Set(string(OutputTable))
	}
	return nil
}

var rootOutputFlag = &outputFlagValue{}

var rootCmd = &cobra.Command{
	Use:           "retab",
	Short:         "Retab is the platform to build, evaluate and scale software-defined document processing automations.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	hardenGroupCommands(rootCmd)
	notify := startUpdateNotifier()
	err := rootCmd.Execute()
	notify()
	if err == nil {
		return nil
	}
	var rendered renderedError
	if errors.As(err, &rendered) {
		return errSilent
	}
	return err
}

// hardenGroupCommands makes router commands (those that only group
// subcommands and have no action of their own) reject unknown
// subcommands instead of silently printing help and exiting 0.
//
// Cobra's built-in unknown-command detection (legacyArgs) only fires for
// the root command — it explicitly bails for any command that has a
// parent. So `retab bogus` errors as expected, but `retab files bogus`,
// `retab files bogus`, `retab workflows runs bogus` etc. all fall through
// to a help dump with exit code 0, which silently swallows typos in
// scripts.
//
// A non-runnable command short-circuits to help *before* cobra validates
// args, so setting Args alone is not enough — the command must also be
// made runnable. We give each router a RunE that just prints help (so a
// bare router command keeps working) plus Args=NoArgs (so any leftover
// token surfaces as `unknown command "..."` for that router).
func hardenGroupCommands(c *cobra.Command) {
	for _, sub := range c.Commands() {
		hardenGroupCommands(sub)
	}
	if c == rootCmd || !c.HasSubCommands() || c.Runnable() {
		return
	}
	c.Args = cobra.NoArgs
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}
}

func replaceTabsInHelpText(c *cobra.Command) {
	c.Short = strings.ReplaceAll(c.Short, "\t", "  ")
	c.Long = strings.ReplaceAll(c.Long, "\t", "  ")
	c.Example = strings.ReplaceAll(c.Example, "\t", "  ")
	for _, sub := range c.Commands() {
		replaceTabsInHelpText(sub)
	}
}

func hideDefaultCompletionCommand(c *cobra.Command) {
	for _, sub := range c.Commands() {
		if sub.Name() == "completion" {
			sub.Hidden = true
			return
		}
	}
}

func init() {
	// `version` lives in version.go and is overwritten at link time by
	// GoReleaser's ldflags. Surfacing it on rootCmd makes both
	// `retab --version` and the polished help header pick up the same
	// string — no risk of drift between them.
	rootCmd.Version = version

	rootCmd.PersistentFlags().String("api-key", "", "Retab API key (env: RETAB_API_KEY)")
	rootCmd.PersistentFlags().String("base-url", "", "Retab API base URL (env: RETAB_API_BASE_URL)")
	rootCmd.PersistentFlags().String("environment-id", "", "Retab environment id for OAuth dashboard context (env: RETAB_ENVIRONMENT_ID)")
	rootCmd.PersistentFlags().Bool("live", false, "use the stored production environment profile (alias for --env production)")
	rootCmd.PersistentFlags().String("env", "", "use the stored environment profile with this slug")
	// `--confirm` is intentionally NOT a persistent flag: it only has an
	// effect on high-risk commands (see highRiskCommands in
	// zz_safety_classification.go), so it is registered locally on just
	// those commands via addConfirmFlag. Registering it globally would
	// advertise it on every command's help where it does nothing.
	rootCmd.PersistentFlags().Bool("debug", false, "verbose debug output")
	rootCmd.PersistentFlags().Var(rootOutputFlag, "output", "output format: json | table | csv (default: auto-detect)")
	rootCmd.PersistentFlags().Var(&outputTableFlagValue{output: rootOutputFlag}, "output-table", "shortcut for --output table")
	if flag := rootCmd.PersistentFlags().Lookup("output-table"); flag != nil {
		flag.NoOptDefVal = "true"
	}
	rootCmd.InitDefaultCompletionCmd()
	replaceTabsInHelpText(rootCmd)
	hideDefaultCompletionCommand(rootCmd)

	// Capture cobra's default help func *before* overriding so we can
	// delegate to it for non-root commands. If we set our func first
	// then asked for HelpFunc() we'd get back our own (recursion).
	cobraDefault := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c == rootCmd {
			renderRootHelp(c.OutOrStdout(), c)
			return
		}
		cobraDefault(c, args)
	})

	hardenGroupCommands(rootCmd)
}
