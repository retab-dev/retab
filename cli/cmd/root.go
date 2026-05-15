package cmd

import (
	"errors"
	"fmt"

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
	case "", "auto", string(OutputJSON), string(OutputTable):
		o.value = raw
		return nil
	default:
		return fmt.Errorf("invalid --output value %q (want: json | table | auto)", raw)
	}
}

var rootCmd = &cobra.Command{
	Use:           "retab",
	Short:         "Retab is the platform to build, evaluate and scale software-defined document processing automations.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	err := rootCmd.Execute()
	if err == nil {
		return nil
	}
	var rendered renderedError
	if errors.As(err, &rendered) {
		return errSilent
	}
	return err
}

func init() {
	// `version` lives in version.go and is overwritten at link time by
	// GoReleaser's ldflags. Surfacing it on rootCmd makes both
	// `retab --version` and the polished help header pick up the same
	// string — no risk of drift between them.
	rootCmd.Version = version

	rootCmd.PersistentFlags().String("api-key", "", "Retab API key (env: RETAB_API_KEY)")
	rootCmd.PersistentFlags().String("base-url", "", "Retab API base URL (env: RETAB_BASE_URL)")
	rootCmd.PersistentFlags().Bool("debug", false, "verbose debug output")
	rootCmd.PersistentFlags().Var(&outputFlagValue{}, "output", "output format: json | table (default: auto-detect)")

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
}
