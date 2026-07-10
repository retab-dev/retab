package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func helpFprintf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

// Polished help renderer for the root command — inspired by `bun`'s layout:
// brand line, single-section grouped commands with aligned descriptions,
// minimal colour, footer with docs links.
//
// Only the root command uses this renderer. Subcommands fall back to
// cobra's default help via `defaultHelpFunc` captured during init() —
// subcommand help is already well-served by cobra's templates and rewriting
// each one would be a lot of yak-shaving with little payoff.
//
// Colour discipline:
//   - NO_COLOR=1 disables all ANSI.
//   - Non-TTY stdout disables all ANSI (so `retab | less` and `retab > x.txt`
//     produce plain text).
//   - Otherwise we use 256-colour for brand pink and 16-colour for everything
//     else, which works on every mainstream terminal including Windows Terminal.

// commandGroup is one section of the rendered help. The ordering of the
// `commands` slice is preserved on screen, so re-arrange here to re-arrange
// the help. Commands not listed in any group end up in "Other" (so a newly
// added subcommand never silently disappears from help — it just lands in
// the catch-all bucket until someone categorises it).
type commandGroup struct {
	title    string
	commands []string
}

// featuredSubcommands lists leaf subcommands worth advertising directly in the
// root help, keyed by parent command name. renderRouterSubcommands only
// expands *routers* (commands that themselves have subcommands), so leaf
// actions normally stay hidden until you run `retab <parent> --help`. The
// local-first `files` commands are the exception: parse/grep/inspect run
// entirely on-device — no upload, no API call — and are easy to miss behind the
// generic "Manage files" row, so we tease them on the front page where users
// will actually discover them.
var featuredSubcommands = map[string][]string{
	"files": {"parse", "grep", "inspect"},
}

var commandGroups = []commandGroup{
	{
		title:    "Primitives",
		commands: []string{"parses", "extractions", "edits", "splits", "partitions", "classifications"},
	},
	{
		// Standalone backend resources you store and reuse. `files` is a
		// first-class resource (FileService), not a utility — it leads here;
		// `schemas` is the generator helper; `tables` is the CSV data resource.
		title:    "Resources",
		commands: []string{"files", "schemas", "tables"},
	},
	{
		title:    "Workflows",
		commands: []string{"workflows"},
	},
	{
		// Org / environment-scoped backend resources (all are real API
		// routes — projects own workflows, secrets are environment-scoped,
		// members/invitations are org people-management).
		title:    "Organization",
		commands: []string{"org", "projects", "secrets", "members", "invitations", "usage"},
	},
	{
		title:    "Account",
		commands: []string{"auth", "env", "version"},
	},
	{
		// Local agent tooling — these install/refresh on-device config and
		// make no backend API calls. Grouped explicitly so they don't read
		// as backend resources in the anonymous "Other" bucket.
		title:    "Setup",
		commands: []string{"setup", "sync"},
	},
}

// styles holds the ANSI escapes used by the renderer. When colour is
// disabled (NO_COLOR, non-TTY) every field is the empty string and the
// printfs degrade to plain text without any conditional branches.
//
// Palette is intentionally ANSI 16-colour, not 256-colour. Bun's CLI
// does the same thing: its wordmark and command rows use bare `\x1b[35m`,
// `\x1b[34m`, etc., which means each user's terminal theme drives the
// actual rendered shade. Hard-coding a 256-colour slot overrides the
// theme — what we want is the opposite, so a Solarized user sees their
// Solarized magenta and a Dracula user sees their Dracula magenta.
//
// Role assignments — every colour is sourced from bun's CLI template
// (src/bun_core/output.zig color_map + src/runtime/cli/cli.zig template):
//   - brand        — bold magenta. Retab wordmark + <command> placeholder.
//     Same code bun uses for its "Bun" wordmark.
//   - groupHeader  — bold yellow. Section sub-headers like "Primitives:",
//     "Workflows:", "Other:". Repurposes the colour bun
//     assigns to its `build` row (we have explicit group
//     sub-headers; bun groups by colour with blank lines).
//   - accent       — bold blue. Every command name in the menu.
//   - headline     — plain bold. Top-level labels: "Usage:", "Flags:",
//     "Learn more:". Matches bun's plain `<b>` for those.
//   - cyan         — footer URLs.
//   - bold         — flag names (`--api-key`, `--debug`, …).
//   - dim          — version tag, parens, env hints, footer hint.
type styles struct {
	reset       string
	bold        string
	dim         string
	brand       string
	accent      string
	groupHeader string
	cyan        string
	headline    string
}

func paletteFor(w io.Writer) styles {
	if os.Getenv("NO_COLOR") != "" {
		return styles{}
	}
	f, ok := w.(*os.File)
	if !ok {
		return styles{}
	}
	if !term.IsTerminal(int(f.Fd())) {
		return styles{}
	}
	return styles{
		reset:       "\x1b[0m",
		bold:        "\x1b[1m",
		dim:         "\x1b[2m",
		brand:       "\x1b[1;35m", // bold magenta — Retab wordmark + <command>
		accent:      "\x1b[1;34m", // bold blue — command names
		groupHeader: "\x1b[1;33m", // bold yellow — section sub-headers (bun's `build` colour)
		cyan:        "\x1b[36m",
		headline:    "\x1b[1m",
	}
}

// renderRootHelp prints the polished top-level help. It reads the
// description of each command from the command's own `Short`, so adding a
// new subcommand with a `Short` is enough to make it appear correctly here
// (modulo categorisation in `commandGroups`).
func renderRootHelp(w io.Writer, root *cobra.Command) {
	renderRootHelpWithStyles(w, root, paletteFor(w))
}

// renderRootHelpWithStyles is the actual renderer. The palette is taken
// as a parameter so tests can pass sentinel escape strings and assert
// exactly which style each visual element receives, without having to
// fake a TTY for paletteFor's auto-detection.
func renderRootHelpWithStyles(w io.Writer, root *cobra.Command, s styles) {

	// ----- header: brand tagline + version -----
	//
	// Bun renders its first line as a full sentence with the wordmark
	// coloured: "<pink>Bun</pink> is a fast JavaScript runtime, ...".
	// We do the same: if rootCmd.Short already starts with "Retab"
	// (the recommended shape — the tagline reads as a complete sentence),
	// we split the wordmark off and colour it in place. Anything else
	// (legacy or test fixtures with arbitrary Short values) renders
	// uncoloured so we never duplicate the wordmark on the line.
	tagline := root.Short
	if tagline == "" {
		tagline = "Retab is the document intelligence CLI."
	}
	versionStr := root.Version
	if versionStr == "" {
		versionStr = "dev"
	}
	// Only prefix with `v` when the version looks like a semver — otherwise
	// values like "dev" / "snapshot-1134275" render as `(vdev)` which reads
	// like a typo.
	versionDisplay := versionStr
	if len(versionStr) > 0 && versionStr[0] >= '0' && versionStr[0] <= '9' {
		versionDisplay = "v" + versionStr
	}
	const wordmark = "Retab"
	if len(tagline) >= len(wordmark) && tagline[:len(wordmark)] == wordmark {
		rest := tagline[len(wordmark):]
		helpFprintf(w, "\n%s%s%s%s %s(%s)%s\n",
			s.brand, wordmark, s.reset, rest, s.dim, versionDisplay, s.reset)
	} else {
		helpFprintf(w, "\n%s %s(%s)%s\n",
			tagline, s.dim, versionDisplay, s.reset)
	}

	// ----- usage line -----
	// `<command>` in brand pink, matching bun's convention where the
	// placeholder is the same colour as the wordmark.
	helpFprintf(w, "\n%sUsage:%s retab %s<command>%s [flags]\n",
		s.headline, s.reset, s.brand, s.reset)

	// ----- index visible subcommands by name -----
	byName := map[string]*cobra.Command{}
	for _, c := range root.Commands() {
		// `help` and `completion` are cobra-auto-generated noise for users;
		// `--help` and `retab help <cmd>` still work, they just don't need
		// to appear in the top-level menu.
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		byName[c.Name()] = c
	}

	// Compute padding across every command name we'll print so the
	// description column lines up cleanly. We always pad from the widest
	// name so "Other"-bucket commands don't break alignment.
	pad := 0
	for name := range byName {
		if len(name) > pad {
			pad = len(name)
		}
	}
	const minPad = 16 // headroom so short layouts don't feel cramped
	if pad < minPad {
		pad = minPad
	}

	// ----- grouped commands -----
	rendered := map[string]bool{}
	for _, g := range commandGroups {
		// Skip the section entirely if none of its commands actually exist
		// — keeps the layout clean during partial rollouts.
		anyPresent := false
		for _, name := range g.commands {
			if _, ok := byName[name]; ok {
				anyPresent = true
				break
			}
		}
		if !anyPresent {
			continue
		}
		// Group sub-header and its commands all sit at col 2 — vertically
		// aligned in the same column. Hierarchy comes from the colon on
		// the header + reading order, not from horizontal indent.
		helpFprintf(w, "\n  %s%s:%s\n", s.groupHeader, g.title, s.reset)
		for _, name := range g.commands {
			c, ok := byName[name]
			if !ok {
				continue
			}
			rendered[name] = true
			// Commands in bold blue `accent` (bun convention). Pad with
			// plain spaces so escape codes don't bleed into trailing
			// whitespace (would corrupt terminal redraw on resize).
			spaces := pad - len(c.Name())
			helpFprintf(w, "  %s%s%s%s  %s\n",
				s.accent, c.Name(), s.reset, repeat(" ", spaces), c.Short)

			renderRouterSubcommands(w, c, s, pad)
			renderFeaturedSubcommands(w, c, s, pad)
		}
	}

	// ----- uncategorised stragglers -----
	var others []*cobra.Command
	for name, c := range byName {
		if !rendered[name] {
			others = append(others, c)
		}
	}
	if len(others) > 0 {
		sort.Slice(others, func(i, j int) bool { return others[i].Name() < others[j].Name() })
		// "Other" is a group sub-header just like Primitives/Utils/etc.
		// Same col 2 alignment for both the header and its commands.
		helpFprintf(w, "\n  %sOther:%s\n", s.groupHeader, s.reset)
		for _, c := range others {
			spaces := pad - len(c.Name())
			helpFprintf(w, "  %s%s%s%s  %s\n",
				s.accent, c.Name(), s.reset, repeat(" ", spaces), c.Short)
			renderRouterSubcommands(w, c, s, pad)
			renderFeaturedSubcommands(w, c, s, pad)
		}
	}

	// ----- global flags -----
	// Listed by hand (vs. iterating PersistentFlags) so we have full
	// control over wording and column alignment. Keep this in sync with
	// the PersistentFlags registered in init() below — there's a small
	// unit test that enforces that contract.
	helpFprintf(w, "\n%sFlags:%s\n", s.headline, s.reset)
	flagRows := []struct {
		name string // e.g. "--api-key"
		hint string // e.g. "KEY"   (empty for boolean flags)
		desc string // e.g. "Retab API key"
		env  string // e.g. "RETAB_API_KEY" (empty if not env-bound)
	}{
		{"--api-key", "KEY", "Retab API key", "RETAB_API_KEY"},
		{"--base-url", "URL", "Retab API base URL", "RETAB_API_BASE_URL"},
		{"--environment-id", "ID", "Retab environment id for OAuth dashboard context", "RETAB_ENVIRONMENT_ID"},
		{"--live", "", "use the stored production environment profile", ""},
		{"--env", "SLUG", "use the stored environment profile with this slug", ""},
		{"--debug", "", "verbose debug output", ""},
		{"--output", "FORMAT", "output format: json | table | csv (default: auto)", ""},
		{"--output-table", "", "shortcut for --output table", ""},
		{"-h, --help", "", "show this help", ""},
		{"-v, --version", "", "show version", ""},
	}
	// Pad the left column (flag name + optional hint) to a fixed width
	// so descriptions line up.
	leftWidth := 0
	for _, f := range flagRows {
		width := len(f.name)
		if f.hint != "" {
			width += 1 + len(f.hint) // " " + hint
		}
		if width > leftWidth {
			leftWidth = width
		}
	}
	for _, f := range flagRows {
		// Build the left column with bold name + dim hint, then compute
		// padding from the *uncoloured* visual width (ANSI escapes are
		// zero-width so we can't use %-*s on the styled string).
		visualWidth := len(f.name)
		left := s.bold + f.name + s.reset
		if f.hint != "" {
			left += " " + s.dim + f.hint + s.reset
			visualWidth += 1 + len(f.hint)
		}
		suffix := ""
		if f.env != "" {
			suffix = "   " + s.dim + "(env: " + f.env + ")" + s.reset
		}
		helpFprintf(w, "  %s%s  %s%s\n",
			left, repeat(" ", leftWidth-visualWidth), f.desc, suffix)
	}

	// ----- footer: docs + hint -----
	helpFprintf(w, "\n%sLearn more:%s\n", s.headline, s.reset)
	helpFprintf(w, "  Docs      %s%s%s\n", s.cyan, "https://docs.retab.com", s.reset)
	helpFprintf(w, "  GitHub    %s%s%s\n", s.cyan, "https://github.com/retab-dev/retab", s.reset)

	// Footer hint — same `<command>` colour rule as the Usage line above so
	// the placeholder reads as a consistent visual token throughout.
	helpFprintf(w, "\n%sRun%s retab %s<command>%s %s--help%s for command-specific options.\n\n",
		s.dim, s.reset, s.brand, s.reset, s.dim, s.reset)
}

// renderRouterSubcommands prints one row per "router" subcommand of c —
// each direct subcommand that itself has subcommands. Leaf actions
// (`list`, `get`, `create`, …) are intentionally NOT expanded — they're
// the body of the command, not its surface, and listing them all would
// quadruple the height of the help.
//
// Rendered at col 4 (one level deeper than command rows at col 2), with
// the description column aligned to the parent's so the eye reads the
// rows as a continuation of the same table.
//
// Today this fires only for `workflows` (its blocks/edges/runs/artifacts/
// tests/experiments sub-namespaces are routers, not leaves), but the
// detection rule is generic — any command that grows nested surface area
// later will surface here automatically.
func renderRouterSubcommands(w io.Writer, c *cobra.Command, s styles, parentPad int) {
	var rows []nestedRow
	for _, sub := range c.Commands() {
		if sub.Hidden || !sub.HasSubCommands() {
			continue
		}
		if sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		rows = append(rows, nestedRow{name: sub.Name(), short: sub.Short})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].name < rows[j].name })
	renderNestedRowList(w, rows, s, parentPad)
}

// renderFeaturedSubcommands prints the curated leaf subcommands of c (see
// featuredSubcommands) at col 4, matching renderRouterSubcommands' layout. The
// curated order is preserved (it's a hand-picked tour, not an alphabetical
// dump). Routers are skipped — renderRouterSubcommands already expands those —
// so a command can never appear twice.
func renderFeaturedSubcommands(w io.Writer, c *cobra.Command, s styles, parentPad int) {
	names := featuredSubcommands[c.Name()]
	if len(names) == 0 {
		return
	}
	byName := map[string]*cobra.Command{}
	for _, sub := range c.Commands() {
		byName[sub.Name()] = sub
	}
	var rows []nestedRow
	for _, name := range names {
		sub, ok := byName[name]
		if !ok || sub.Hidden || sub.HasSubCommands() {
			continue
		}
		rows = append(rows, nestedRow{name: sub.Name(), short: sub.Short})
	}
	renderNestedRowList(w, rows, s, parentPad)
}

// nestedRow is the row shape both nested renderers feed to renderNestedRowList.
type nestedRow struct{ name, short string }

// renderNestedRowList does the actual col-4 emission shared by the router and
// featured-leaf renderers.
func renderNestedRowList(w io.Writer, rows []nestedRow, s styles, parentPad int) {
	if len(rows) == 0 {
		return
	}
	// Parent description sits at col (2 + parentPad + 2). To align the
	// nested-row description to the same column from a col-4 indent, the
	// name field width is (parentPad - 2). If a sub-name happens to be
	// longer than that, force at least one separator space so the name
	// never runs into the description text.
	subPad := parentPad - 2
	for _, r := range rows {
		spaces := subPad - len(r.name)
		if spaces < 1 {
			spaces = 1
		}
		helpFprintf(w, "    %s%s%s%s  %s\n",
			s.accent, r.name, s.reset, repeat(" ", spaces), r.short)
	}
}

// repeat builds a string of n spaces (or whatever sep is). Used to pad
// columns *outside* of ANSI escape sequences so the codes don't bleed
// into trailing whitespace and ruin terminal redraw on resize.
// Returns the empty string for n <= 0 — a negative count would be a bug
// in the width math, but silently degrading beats panicking in help output.
func repeat(sep string, n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, n*len(sep))
	for i := 0; i < n; i++ {
		out = append(out, sep...)
	}
	return string(out)
}
