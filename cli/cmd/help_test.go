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

// Cobra renders Long / Short / Example verbatim into help output. A stray
// tab inside a raw-string literal (easy to introduce when a multi-line
// description is indented to match surrounding Go code) shows up as a
// misaligned hard tab in the user's terminal. Walk every registered
// command and reject embedded tabs so the help stays clean.
func TestCommandHelpTextHasNoStrayTabs(t *testing.T) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for label, text := range map[string]string{
			"Short":   c.Short,
			"Long":    c.Long,
			"Example": c.Example,
		} {
			if strings.Contains(text, "\t") {
				t.Errorf("command %q has a stray tab in its %s text:\n%q", c.CommandPath(), label, text)
			}
		}
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}

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
		// Match at the command row's 2-space gutter (rows are "  <name>"
		// with at least one trailing space before the description). The
		// trailing space rules out e.g. matching "files" inside a
		// hypothetical "filesystem" description.
		if !strings.Contains(out, "  "+c.Name()+" ") {
			t.Errorf("command %q is registered on rootCmd but does not appear in help output:\n%s", c.Name(), out)
		}
	}
}

// Groups must appear in source order so re-ordering `commandGroups` in
// help.go is the *only* lever for changing on-screen layout. We don't
// hardcode names here — that lets the maintainer rename / re-order
// groups without touching tests, and still catches accidental
// reshufflings introduced by other edits.
func TestRenderRootHelp_GroupOrderIsPreserved(t *testing.T) {
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()

	prev := -1
	for _, g := range commandGroups {
		idx := strings.Index(out, g.title+":")
		if idx < 0 {
			// Group with zero present commands is allowed to be skipped.
			continue
		}
		if idx < prev {
			t.Errorf("group %q (idx=%d) appears before a previous group (idx=%d):\n%s",
				g.title, idx, prev, out)
		}
		prev = idx
	}
}

// Build a minimal cmd tree with just one command, find the group that
// owns it, and assert: that one group renders; all the others are
// skipped. Group identity is read from `commandGroups`, so renames are
// safe.
func TestRenderRootHelp_EmptyGroupIsSkipped(t *testing.T) {
	const pickedCmd = "files"
	var owningGroup string
	for _, g := range commandGroups {
		for _, n := range g.commands {
			if n == pickedCmd {
				owningGroup = g.title
				break
			}
		}
	}
	if owningGroup == "" {
		t.Fatalf("test setup: %q is not in any group", pickedCmd)
	}

	root := &cobra.Command{Use: "retab", Short: "x"}
	root.AddCommand(&cobra.Command{Use: pickedCmd, Short: "Manage files"})
	var buf bytes.Buffer
	renderRootHelp(&buf, root)
	out := buf.String()

	if !strings.Contains(out, owningGroup+":") {
		t.Errorf("group %q should render when %q is present:\n%s", owningGroup, pickedCmd, out)
	}
	for _, g := range commandGroups {
		if g.title == owningGroup {
			continue
		}
		if strings.Contains(out, g.title+":") {
			t.Errorf("group %q should be skipped when none of its commands are registered:\n%s", g.title, out)
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

// Subcommands that are themselves routers (their `Use` has children of
// its own — like `workflows blocks`, `workflows experiments`) should be
// surfaced one indent level deeper than the parent command row, so users
// can discover nested CLI surface from the top-level help. Leaf actions
// like `list`/`get`/`create` are NOT expanded — those are the body of the
// command, not its surface.
func TestRenderRootHelp_RouterSubcommandsAreExpanded(t *testing.T) {
	// Build a stand-in tree so the test doesn't depend on which routers
	// the real workflows command happens to have today.
	root := &cobra.Command{Use: "retab", Short: "x"}
	workflows := &cobra.Command{Use: "workflows", Short: "Manage workflows"}
	// Two routers (they have subcommands), one leaf (no subcommands).
	blocks := &cobra.Command{Use: "blocks", Short: "Manage workflow blocks"}
	blocks.AddCommand(&cobra.Command{Use: "list", Short: "List blocks"})
	experiments := &cobra.Command{Use: "experiments", Short: "Manage experiments"}
	experiments.AddCommand(&cobra.Command{Use: "create", Short: "Create"})
	leaf := &cobra.Command{Use: "create", Short: "Leaf — should NOT be expanded"}
	workflows.AddCommand(blocks, experiments, leaf)
	root.AddCommand(workflows)

	var buf bytes.Buffer
	renderRootHelp(&buf, root)
	out := buf.String()

	// Both routers should appear with their Short description, indented
	// 4 cols (one level deeper than the parent at col 2).
	for _, want := range []string{"blocks", "experiments"} {
		if !strings.Contains(out, "    "+want+" ") {
			t.Errorf("router subcommand %q should be expanded at col 4:\n%s", want, out)
		}
	}
	// Router subcommands should land *under* the parent's row.
	parentIdx := strings.Index(out, "workflows ")
	subIdx := strings.Index(out, "    blocks ")
	if parentIdx < 0 || subIdx < 0 || subIdx < parentIdx {
		t.Errorf("router subcommand should appear after the parent row (parent=%d, sub=%d):\n%s",
			parentIdx, subIdx, out)
	}
	// Leaf subcommand `create` must NOT appear as a nested row (it's a
	// leaf action, not a router). Note: this test's `create` has no
	// subcommands; we check that the unique short "Leaf — should NOT be
	// expanded" never gets rendered.
	if strings.Contains(out, "Leaf — should NOT be expanded") {
		t.Errorf("leaf subcommand was expanded (it shouldn't be):\n%s", out)
	}
}

func TestRenderRootHelp_FeaturedFilesSubcommandsAreTeased(t *testing.T) {
	// The local-first `files` commands (parse/grep/inspect) are leaves, so the
	// generic router expansion won't surface them. featuredSubcommands should
	// tease exactly those three under the `files` row, and NOT the API-facing
	// leaves like `upload`.
	root := &cobra.Command{Use: "retab", Short: "x"}
	files := &cobra.Command{Use: "files", Short: "Manage files"}
	files.AddCommand(
		&cobra.Command{Use: "parse <path>", Short: "Parse a local document"},
		&cobra.Command{Use: "grep <path> <pattern>", Short: "Search a local document"},
		&cobra.Command{Use: "inspect <path>", Short: "Inspect a region of a local document"},
		&cobra.Command{Use: "upload <path>", Short: "Upload — should NOT be teased"},
		&cobra.Command{Use: "doctor", Short: "Doctor — not in the featured list"},
	)
	root.AddCommand(files)

	var buf bytes.Buffer
	renderRootHelp(&buf, root)
	out := buf.String()

	// The three featured leaves appear at col 4 (one level under `files`).
	for _, want := range []string{"parse", "grep", "inspect"} {
		if !strings.Contains(out, "    "+want+" ") {
			t.Errorf("featured files subcommand %q should be teased at col 4:\n%s", want, out)
		}
	}
	// They sit *under* the files row, in curated (not alphabetical) order.
	filesIdx := strings.Index(out, "files ")
	parseIdx := strings.Index(out, "    parse ")
	grepIdx := strings.Index(out, "    grep ")
	inspectIdx := strings.Index(out, "    inspect ")
	if filesIdx < 0 || parseIdx < filesIdx || grepIdx < parseIdx || inspectIdx < grepIdx {
		t.Errorf("featured order should be files -> parse -> grep -> inspect (files=%d parse=%d grep=%d inspect=%d):\n%s",
			filesIdx, parseIdx, grepIdx, inspectIdx, out)
	}
	// Non-featured leaves must not be teased.
	if strings.Contains(out, "Upload — should NOT be teased") {
		t.Errorf("non-featured leaf `upload` was teased (it shouldn't be):\n%s", out)
	}
	if strings.Contains(out, "Doctor — not in the featured list") {
		t.Errorf("non-featured leaf `doctor` was teased (it shouldn't be):\n%s", out)
	}
}

func TestFilesHelpHighlightsLocalParseAndRender(t *testing.T) {
	out := commandHelpOutput(t, "files")
	for _, want := range []string{
		"Local document tools do not require an API key and never upload data:",
		"inspect --render   render PDF/image pages to PNG files for visual review",
		"retab files parse ./invoice.pdf --format json --bbox",
		"retab files inspect ./statement.pdf --render 1-3 --out ./pages",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("files help missing %q:\n%s", want, out)
		}
	}
}

func TestFilesParseHelpMakesLocalJSONAndCacheObvious(t *testing.T) {
	out := commandHelpOutput(t, "files", "parse")
	for _, want := range []string{
		"Parse a local document to text/JSON without uploading",
		"--format json      print normalized JSON",
		"--bbox             include positioned text items for PDFs/images",
		"PDF/image parses are cached by file content and parse options",
		"retab files parse invoice.pdf --format json --bbox -o invoice.parse.json",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("parse help missing %q:\n%s", want, out)
		}
	}
}

func TestFilesInspectHelpMakesRenderOutputObvious(t *testing.T) {
	out := commandHelpOutput(t, "files", "inspect")
	for _, want := range []string{
		"Inspect lines/cells or render PDF/image pages to PNG",
		"Use --render when an agent or human needs to see the page visually.",
		"prints JSON with the image paths",
		"only existing pages are rendered",
		"retab files inspect report.pdf --render 1-3 --out ./pages",
		"retab files inspect receipt.jpg --render 1 --out ./pages",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("inspect help missing %q:\n%s", want, out)
		}
	}
}

func commandHelpOutput(t *testing.T, path ...string) string {
	t.Helper()
	cmd, _, err := rootCmd.Find(path)
	if err != nil {
		t.Fatalf("find %v: %v", path, err)
	}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	t.Cleanup(func() {
		cmd.SetOut(nil)
		cmd.SetErr(nil)
	})
	if err := cmd.Help(); err != nil {
		t.Fatalf("help %v: %v", path, err)
	}
	return buf.String()
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
	// The brand line inlines the wordmark into the tagline sentence
	// (matching bun: "<Bun> is a fast JavaScript runtime…"). Looking for
	// the literal "Retab is" lets the test stay agnostic to wordmark
	// colouring and tagline rewording — both can change freely as long
	// as the sentence still starts with the product name.
	if !strings.Contains(rootBuf.String(), "Retab is") {
		t.Errorf("root help missing brand sentence:\n%s", rootBuf.String())
	}
	// Any of the configured groups appearing is sufficient evidence the
	// custom renderer ran (vs. cobra's default, which uses "Available
	// Commands:"). Reading group names from commandGroups avoids hardcoded
	// titles that drift when the groupings get re-organised.
	gotAnyGroup := false
	for _, g := range commandGroups {
		if strings.Contains(rootBuf.String(), g.title+":") {
			gotAnyGroup = true
			break
		}
	}
	if !gotAnyGroup {
		t.Errorf("root help missing any of the configured group headers:\n%s", rootBuf.String())
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
	// Cobra's default child help doesn't print our brand sentence (it
	// uses the child's own Short / Long), so "Retab is" must not appear.
	if strings.Contains(childBuf.String(), "Retab is") {
		t.Errorf("child help leaked brand sentence:\n%s", childBuf.String())
	}
	if !strings.Contains(childBuf.String(), "Usage:") {
		t.Errorf("child help missing cobra-default 'Usage:' section:\n%s", childBuf.String())
	}
}

// Pin the colour role of each visual element by injecting a sentinel
// palette and asserting exact wrappers. Five roles all live in one test
// so a regression on any of them points at the specific role that drifted:
//
//   - command names                       → bold blue   `<A>` (accent)
//   - <command> placeholder (Usage+footer) → bold magenta `<P>` (brand)
//   - group sub-headers (Primitives:, …)  → bold yellow  `<G>` (groupHeader)
//   - top-level labels (Usage:, Flags:, Learn more:) → plain bold `<H>` (headline)
//   - flag names (--api-key, --debug, …)  → plain bold `<B>` (bold)
//
// These match bun's role assignments — see palette docs in help.go.
func TestRenderRootHelp_ColourRolesAreCorrect(t *testing.T) {
	sentinel := styles{
		reset: "<R>", bold: "<B>", dim: "<D>",
		brand: "<P>", accent: "<A>", groupHeader: "<G>",
		cyan: "<C>", headline: "<H>",
	}
	var buf bytes.Buffer
	renderRootHelpWithStyles(&buf, rootCmd, sentinel)
	out := buf.String()

	// Every visible command name should be wrapped in the accent role.
	for _, c := range rootCmd.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		if !strings.Contains(out, "<A>"+c.Name()+"<R>") {
			t.Errorf("command %q is not wrapped in <A>…<R> (accent / bold blue):\n%s", c.Name(), out)
		}
	}

	// <command> placeholder uses brand magenta in both Usage line and footer.
	if got := strings.Count(out, "<P><command><R>"); got < 2 {
		t.Errorf("<command> should appear in <P>…<R> in both Usage and footer (got %d occurrences):\n%s", got, out)
	}

	// Group sub-headers use the yellow accent. Sample the first group that
	// actually appears (some groups may have zero commands and be skipped).
	gotGroupHeader := false
	for _, g := range commandGroups {
		if strings.Contains(out, "<G>"+g.title+":<R>") {
			gotGroupHeader = true
			break
		}
	}
	if !gotGroupHeader {
		t.Errorf("no group sub-header wrapped in <G>…<R> (groupHeader / bold yellow):\n%s", out)
	}

	// Top-level labels stay in plain bold (the spec headline role).
	for _, label := range []string{"Usage", "Flags", "Learn more"} {
		if !strings.Contains(out, "<H>"+label+":<R>") {
			t.Errorf("top-level label %q should be wrapped in <H>…<R> (headline / plain bold):\n%s", label, out)
		}
	}

	// Flag names use plain bold. Sample --debug — no value-hint so its
	// row is the simplest to match unambiguously.
	if !strings.Contains(out, "<B>--debug<R>") {
		t.Errorf("flag --debug should be wrapped in <B>…<R> (bold):\n%s", out)
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
	// `registered` is every persistent root flag; `visible` excludes ones
	// marked hidden. Hidden persistent flags (e.g. `--confirm`, which is a
	// blanket parse-everywhere flag whose visible copy is registered locally
	// on high-risk commands) are intentionally absent from the root help, so
	// only `visible` flags are required to appear there. The converse check
	// below still uses the full `registered` set: a help line must map to some
	// real flag, hidden or not.
	registered := map[string]bool{}
	visible := map[string]bool{}
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		registered[f.Name] = true
		if !f.Hidden {
			visible[f.Name] = true
		}
	})

	// Render help and check every visible registered persistent flag appears.
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()
	for name := range visible {
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
