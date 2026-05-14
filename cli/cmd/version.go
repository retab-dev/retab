package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build-time identity, populated by `-ldflags "-X .../cmd.version=..."`
// (see .goreleaser.yaml and .github/workflows/release-cli.yml). Two
// surfaces read these vars:
//
//   - `retab --version`  — cobra wires this from rootCmd.Version (see
//                          root.go's init()). Short, just the version.
//   - `retab version`    — registered below. Prints version + commit
//                          + ISO build date.
//
// We expose both because users coming from `git`/`go`/`node` instinctively
// type the subcommand form first; getting `unknown command "version"` is a
// needless first-run papercut. The subcommand also surfaces commit + date
// which the flag doesn't, and is the format we ask for in bug reports.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version, commit, and build date",
	Long: `Print the build identity of this binary.

Equivalent to ` + "`retab --version`" + ` but also prints the git commit
SHA and ISO build timestamp that GoReleaser injects at link time —
useful when filing bug reports or pinning a remote machine to a known
build.`,
	Example: `  retab version
  # retab 0.1.0 (commit a1b2c3d, built 2026-05-14T15:03:21Z)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "retab %s (commit %s, built %s)\n",
			version, commit, date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
