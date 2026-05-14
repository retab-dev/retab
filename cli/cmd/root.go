package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "retab",
	Short:         "Retab is the platform to build, evaluate and scale software-defined document processing automations.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
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
