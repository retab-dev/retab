package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("retab %s (commit %s, built %s)\n", version, commit, date)
		return nil
	},
}

func init() {
	rootCmd.Version = version
	rootCmd.AddCommand(versionCmd)
}
