package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// walkLeafArityCommands visits c and every command beneath it.
func walkLeafArityCommands(c *cobra.Command, visit func(*cobra.Command)) {
	visit(c)
	for _, sub := range c.Commands() {
		walkLeafArityCommands(sub, visit)
	}
}

// usePlaceholders counts the required (<arg>) and optional ([arg]) positional
// placeholders a command advertises in its Use line, and whether any of them is
// variadic (a trailing ...). The leading token is the command's own name and is
// skipped; the cobra-appended "[flags]" token is ignored. An optional group can
// itself wrap a required-looking token (e.g. "[<workflow-id>]"); a leading "["
// always wins, so that is counted as optional, not required.
func usePlaceholders(c *cobra.Command) (required, optional int, variadic bool) {
	fields := strings.Fields(c.Use)
	for i, tok := range fields {
		if i == 0 || tok == "[flags]" {
			continue
		}
		if strings.Contains(tok, "...") {
			variadic = true
		}
		switch {
		case strings.HasPrefix(tok, "["):
			optional++
		case strings.HasPrefix(tok, "<"):
			required++
		}
	}
	return required, optional, variadic
}

func makeDummyArgs(n int) []string {
	if n <= 0 {
		return nil
	}
	args := make([]string, n)
	for i := range args {
		args[i] = "x"
	}
	return args
}

// Every leaf command must actually accept the positional arguments it
// advertises in its Use line. The codebase convention is that a command taking
// positional args declares an explicit Args validator (cobra.ExactArgs,
// cobra.RangeArgs, ...), and hardenLeafArgs slaps cobra.NoArgs on any runnable
// command that forgot one. That safety net has a sharp edge: a command whose
// Use shows `get <id>` but whose author forgot the ExactArgs(1) silently becomes
// NoArgs and then *rejects its own documented argument* — `retab x get id` fails
// with a confusing "unknown command". Today only two commands (files get,
// workflows get) are guarded against that, by hand.
//
// This asserts the ACCEPT direction for every placeholder command: the Args
// validator must accept the documented required count and the documented
// maximum (required+optional). The opposite direction — rejecting too FEW args
// — is deliberately not asserted: several commands accept their `<id>` either
// positionally or via a flag (e.g. `workflows runs export <workflow-id>` also
// honours --workflow-id) or from stdin (`files upload <path>`), so a `<arg>`
// placeholder is not always a hard positional requirement here.
func TestLeafCommandArityMatchesUsePlaceholders(t *testing.T) {
	checked := 0
	walkLeafArityCommands(rootCmd, func(c *cobra.Command) {
		// Only leaf commands carry positional args; router/group commands are
		// covered by the unknown-subcommand invariants instead.
		if c == rootCmd || !c.Runnable() || c.HasSubCommands() {
			return
		}
		if c.Args == nil {
			// hardenLeafArgs guarantees a non-nil validator on every runnable
			// command; if that ever regresses there is nothing to assert here.
			return
		}
		required, optional, variadic := usePlaceholders(c)
		if required == 0 && optional == 0 {
			// Zero-positional commands are covered by
			// TestListCommandsRejectExtraPositionalArgs.
			return
		}
		checked++
		path := c.CommandPath()

		// Must accept the documented required count — this is the assertion
		// that catches the hardenLeafArgs over-reach (a forgotten validator
		// turned into NoArgs would reject this).
		if err := c.Args(c, makeDummyArgs(required)); err != nil {
			t.Errorf("%s: Use %q advertises %d required positional arg(s) but the Args validator rejects %d: %v "+
				"(did the command forget its cobra.ExactArgs/RangeArgs and get hardened to NoArgs?)",
				path, c.Use, required, required, err)
		}
		// Must also accept the documented maximum (required+optional), unless
		// the tail is variadic (no fixed upper bound to assert). Catches an
		// ExactArgs that is one short of a documented optional argument.
		if !variadic && optional > 0 {
			if err := c.Args(c, makeDummyArgs(required+optional)); err != nil {
				t.Errorf("%s: Use %q advertises %d optional positional arg(s) but the Args validator rejects the documented maximum of %d: %v",
					path, c.Use, optional, required+optional, err)
			}
		}
	})
	if checked == 0 {
		t.Fatal("discovered no commands with positional placeholders; the tree walk is broken")
	}
}

// Every visible runnable command must carry a non-empty Short description.
// Short is the line shown next to the command in `--help`; a blank one ships a
// command that looks broken and is hard to discover. Hidden plumbing commands
// (e.g. the detached update-check daemon) and cobra's auto-generated
// help/completion commands are exempt.
func TestRunnableCommandsHaveNonEmptyShort(t *testing.T) {
	checked := 0
	walkLeafArityCommands(rootCmd, func(c *cobra.Command) {
		if c == rootCmd || !c.Runnable() || c.Hidden {
			return
		}
		if c.Name() == "help" || c.Name() == "completion" {
			return
		}
		checked++
		if strings.TrimSpace(c.Short) == "" {
			t.Errorf("%s: command has an empty Short description (shown in --help)", c.CommandPath())
		}
	})
	if checked == 0 {
		t.Fatal("discovered no runnable commands; the tree walk is broken")
	}
}
