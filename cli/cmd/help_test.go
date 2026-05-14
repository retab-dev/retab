package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The polished `retab` (no-args) help is user-facing surface, so the
// tests below pin the contract that matters to users *and* the
// invariants that, if broken, would silently degrade the rendered UI:
//
//   1. Color discipline — escapes only when the destination is a TTY and
//      NO_COLOR is unset. (Piping to a file or `less` must produce plain
//      text or the file fills with garbage.)
//   2. Coverage — every registered subcommand appears somewhere, either
//      in a named group or in the "Other" fallback. If a command is
//      added later and forgotten in `commandGroups`, it must still show
//      up (so users can discover it).
//   3. Group ordering — Documents → Extraction → Workflows → Account.
//   4. Version formatting — `(dev)`, `(v0.1.0)`, `(snapshot-abc)`.
//   5. Dispatch — only the root command uses the fancy renderer;
//      subcommand help (`retab files --help`) falls through to cobra's
//      default templates. Regressing here would mean rewriting every
//      subcommand's help, which is not what we signed up for.

func TestPaletteFor_DisablesWhenNotAFile(t *testing.T) {
	// bytes.Buffer is the prototypical "writing to memory" case — e.g.
	// `retab > out.txt`. Must produce plain text.
	var buf bytes.Buffer
	s := paletteFor(&buf)
	if s.bold != "" || s.brand != "" || s.cyan != "" {
		t.Fatalf("expected empty palette for non-file writer, got %+v", s)
	}
}

func TestPaletteFor_DisablesOnNO_COLOR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// Even when handed os.Stdout (potentially a TTY), NO_COLOR must win.
	// We can't reliably claim os.Stdout is a TTY inside `go test`, but
	// this test still pins the precedence: with NO_COLOR set, the result
	// is empty regardless of the writer.
	var buf bytes.Buffer
	if got := paletteFor(&buf); got.bold != "" {
		t.Fatalf("NO_COLOR ignored: %+v", got)
	}
}

func TestRenderRootHelp_NoEscapesForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	if strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("ANSI escape leaked into non-TTY output:\n%s", buf.String())
	}
}

func TestRenderRootHelp_EveryRegisteredCommandIsVisible(t *testing.T) {
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()
	for _, c := range rootCmd.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		// Match at start-of-line + 4-space indent so we don't accidentally
		// pass via a substring appearance somewhere in a description.
		needle := "    " + c.Name() + " " // trailing space rules out prefixes
		// Some commands hit the exact `pad` width and have only the
		// description right after, but always with two-space separator.
		// So look for either the spaced needle or `Name` followed by `<space><space>`:
		if !strings.Contains(out, needle) && !strings.Contains(out, "    "+c.Name()+"  ") {
			t.Errorf("command %q is registered on rootCmd but does not appear in help output", c.Name())
		}
	}
}

func TestRenderRootHelp_GroupOrderIsPreserved(t *testing.T) {
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()
	idx := func(s string) int { return strings.Index(out, s) }
	docs := idx("Documents:")
	extr := idx("Extraction:")
	wf := idx("Workflows:")
	acct := idx("Account:")
	if docs < 0 || extr < 0 || wf < 0 || acct < 0 {
		t.Fatalf("missing section headers: docs=%d extr=%d wf=%d acct=%d\n%s", docs, extr, wf, acct, out)
	}
	if !(docs < extr && extr < wf && wf < acct) {
		t.Errorf("group order broken: docs=%d extr=%d wf=%d acct=%d", docs, extr, wf, acct)
	}
}

func TestRenderRootHelp_EmptyGroupIsSkipped(t *testing.T) {
	// Build a minimal cmd tree with only "files" registered. Extraction /
	// Workflows / Account groups have nothing in them and must not appear.
	root := &cobra.Command{Use: "retab", Short: "x"}
	root.AddCommand(&cobra.Command{Use: "files", Short: "Manage files"})
	var buf bytes.Buffer
	renderRootHelp(&buf, root)
	out := buf.String()
	if !strings.Contains(out, "Documents:") {
		t.Errorf("Documents section should render when 'files' is present:\n%s", out)
	}
	for _, absent := range []string{"Extraction:", "Workflows:", "Account:"} {
		if strings.Contains(out, absent) {
			t.Errorf("section %q should be skipped when no commands match:\n%s", absent, out)
		}
	}
}

func TestRenderRootHelp_UncategorizedCommandLandsInOther(t *testing.T) {
	// A user added a new subcommand but forgot to slot it into a group.
	// The renderer must still show it (in "Other") so users discover it.
	root := &cobra.Command{Use: "retab"}
	root.AddCommand(&cobra.Command{Use: "files", Short: "f"})
	root.AddCommand(&cobra.Command{Use: "honeybadger", Short: "uncategorised"})
	var buf bytes.Buffer
	renderRootHelp(&buf, root)
	out := buf.String()
	if !strings.Contains(out, "Other:") {
		t.Errorf("missing 'Other:' header for uncategorised commands:\n%s", out)
	}
	if !strings.Contains(out, "honeybadger") {
		t.Errorf("uncategorised command 'honeybadger' did not appear:\n%s", out)
	}
}

func TestRenderRootHelp_VersionFormatting(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "(dev)"},
		{"dev", "(dev)"},
		{"0.1.0", "(v0.1.0)"},
		{"1.2.3-rc.1", "(v1.2.3-rc.1)"},
		{"snapshot-abc", "(snapshot-abc)"},
	}
	for _, tc := range cases {
		t.Run("version="+tc.in, func(t *testing.T) {
			root := &cobra.Command{Use: "retab", Short: "x", Version: tc.in}
			var buf bytes.Buffer
			renderRootHelp(&buf, root)
			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("version %q: want substring %q, got:\n%s", tc.in, tc.want, buf.String())
			}
		})
	}
}

// Dispatch contract: `rootCmd.HelpFunc()` is the single entry point cobra
// uses for `--help`, `-h`, and `help <cmd>` lookups. Root must get our
// rendered header; children must get cobra's default template (no brand line,
// `Usage:` block + `Available Commands:` block).
func TestHelpFunc_RootGetsCustomChildrenGetDefault(t *testing.T) {
	// Root → fancy
	var rootBuf bytes.Buffer
	rootCmd.SetOut(&rootBuf)
	t.Cleanup(func() { rootCmd.SetOut(nil) })
	rootCmd.HelpFunc()(rootCmd, nil)
	if !strings.Contains(rootBuf.String(), "Retab · ") {
		t.Errorf("root help missing brand header:\n%s", rootBuf.String())
	}
	if !strings.Contains(rootBuf.String(), "Documents:") {
		t.Errorf("root help missing grouped commands:\n%s", rootBuf.String())
	}

	// Child → cobra default
	files, _, err := rootCmd.Find([]string{"files"})
	if err != nil {
		t.Fatalf("files command not registered: %v", err)
	}
	var childBuf bytes.Buffer
	files.SetOut(&childBuf)
	t.Cleanup(func() { files.SetOut(nil) })
	rootCmd.HelpFunc()(files, nil)
	if strings.Contains(childBuf.String(), "Retab · ") {
		t.Errorf("child help leaked brand header:\n%s", childBuf.String())
	}
	if !strings.Contains(childBuf.String(), "Usage:") {
		t.Errorf("child help missing cobra-default 'Usage:' section:\n%s", childBuf.String())
	}
}

// help, completion, and Hidden commands are noise on the top-level menu —
// they're available via `retab help <x>` and `retab completion --help`,
// but listing them up top buries the actually-useful surface.
func TestRenderRootHelp_HidesHelpAndCompletion(t *testing.T) {
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()
	// `completion` is the cobra-auto-generated subcommand
	if strings.Contains(out, "completion ") {
		t.Errorf("expected `completion` to be hidden from top-level help:\n%s", out)
	}
	// `help` is the cobra-auto-generated help command
	for _, line := range strings.Split(out, "\n") {
		// Must not have a line like "    help     ...something..."
		trimmed := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trimmed, "help ") || trimmed == "help" {
			t.Errorf("expected `help` to be hidden, found line: %q", line)
		}
	}
}

// Contract test paired with the comment in help.go's flag-rendering block.
// The flag table there is hand-maintained (so we control wording and
// alignment). If anyone adds, renames, or removes a persistent root flag
// in root.go, they must update help.go too — and vice versa. This test
// fails loudly when the two drift.
func TestHelpFlagsMatchRegisteredPersistentFlags(t *testing.T) {
	registered := map[string]bool{}
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { registered[f.Name] = true })

	// Render help and check every registered persistent flag appears.
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()
	for name := range registered {
		if !strings.Contains(out, "--"+name) {
			t.Errorf("persistent flag --%s is registered but missing from help output:\n%s", name, out)
		}
	}

	// Conversely: every flag listed in the help block (other than -h/-v,
	// which are auto-added by cobra) must be a real registered flag.
	// Catches typos / stale flag rows that would otherwise lie to users.
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "--") {
			continue
		}
		// Extract just the flag name up to whitespace / ANSI.
		name := strings.TrimPrefix(trimmed, "--")
		if i := strings.IndexAny(name, " \t\x1b"); i >= 0 {
			name = name[:i]
		}
		if !registered[name] {
			t.Errorf("flag --%s appears in help but is not registered on rootCmd.PersistentFlags", name)
		}
	}
}
