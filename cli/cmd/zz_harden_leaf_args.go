package cmd

import "github.com/spf13/cobra"

// hardenLeafArgs walks the command tree and gives every runnable command
// that never declared an Args validator a cobra.NoArgs one.
//
// Cobra's default when Args is nil is ArbitraryArgs: the command silently
// accepts and ignores any extra positional tokens. For the list commands
// (`retab files list`, `retab jobs list`, `retab workflows list`, ...) that
// means `retab files list file_abc123` — a plausible typo for a get/filter
// — exits 0 and returns the *entire* list, silently dropping the argument.
// Same trap for the flag-only create commands.
//
// The codebase convention is that any command which legitimately takes
// positional args declares an explicit Args validator (cobra.ExactArgs,
// cobra.RangeArgs, ...). So a runnable command with Args == nil is one that
// takes no positional args; making that explicit with cobra.NoArgs turns
// the silent swallow into a clear "unknown command" error with exit 1.
//
// Group commands are handled by hardenGroupCommands instead. By the time
// this runs they are either still non-runnable (skipped here) or already
// carry a NoArgs validator (skipped here too), so the two walks never
// fight over the same command regardless of init order.
func hardenLeafArgs(c *cobra.Command) {
	for _, sub := range c.Commands() {
		hardenLeafArgs(sub)
	}
	if c == rootCmd {
		return
	}
	if c.Runnable() && c.Args == nil {
		c.Args = cobra.NoArgs
	}
}

func init() {
	hardenLeafArgs(rootCmd)
}
