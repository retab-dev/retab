package cmd

import "github.com/spf13/cobra"

// RootCommand returns the fully-assembled root command.
//
// It is exported so external tooling — shell-completion generators,
// documentation builders, and the cross-command conformance tests under
// cmd/tests — can introspect the command tree (walk subcommands, inspect
// flags, invoke RunE) without depending on unexported package state. It is the
// same command Execute runs, so what callers observe matches real behavior.
func RootCommand() *cobra.Command {
	return rootCmd
}
