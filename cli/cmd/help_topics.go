package cmd

import (
	"github.com/spf13/cobra"
)

// Topical help subsystem — `retab help <topic>` style entries for
// cross-cutting concerns that don't belong to a single subcommand.
//
// Pattern is borrowed from the Retab MCP server's `help(topic='…')` tool
// and from `gh help <topic>`: short markdown-ish prose surfaced through
// cobra's auto-generated `help` command. Cobra finds a topic the same way
// it finds any subcommand — they're registered on rootCmd — so
// `retab help quickstart` Just Works without a custom dispatcher.
//
// Topics are registered Hidden so they don't pollute the polished command
// menu (see renderRootHelp in help.go, which filters Hidden). They're
// surfaced instead through the dedicated "Topics:" section appended to the
// rendered root help below, so users still discover them.
//
// Adding a topic: append to `helpTopics` below. The slice ordering drives
// the on-screen ordering both in `retab --help` and in any test snapshots.

// helpTopic is one entry in the topical help registry. It maps cleanly
// onto a cobra.Command — the `Use` field becomes the topic name on the
// command line (`retab help <Use>`), the `Short` lands in the "Topics:"
// menu, and `Long` is rendered when the topic is requested.
type helpTopic struct {
	use   string
	short string
	long  string
}

// helpTopics is the canonical, ordered list of topics. To add or
// reorganise topics, edit this slice — everything else (registration in
// init(), menu rendering, the consistency test in help_topics_test.go)
// reads from here.
var helpTopics = []helpTopic{
	{
		use:   "quickstart",
		short: "Authenticate and run your first extraction in under a minute",
		long:  topicQuickstart,
	},
	{
		use:   "authentication",
		short: "How --api-key, env, and the config file interact",
		long:  topicAuthentication,
	},
	{
		use:   "configuration",
		short: "Layout of ~/.retab/config.json and how to edit it",
		long:  topicConfiguration,
	},
	{
		use:   "formats",
		short: "Output formats — JSON, NDJSON streaming, exit codes",
		long:  topicFormats,
	},
	{
		use:   "shell-completion",
		short: "Enable tab-completion in bash / zsh / fish / powershell",
		long:  topicShellCompletion,
	},
}

func init() {
	for _, t := range helpTopics {
		// Hidden so the polished menu skips them (cf. help.go's filter on
		// `c.Hidden`); the dedicated "Topics:" section in renderRootHelp
		// lists them by reading from helpTopics directly.
		//
		// No Run/RunE — cobra treats this as a help-only command and
		// prints `Long` when the topic is invoked. That's also what
		// `retab help <topic>` ultimately calls.
		topic := t // capture for closure-safety against future refactors
		rootCmd.AddCommand(&cobra.Command{
			Use:    topic.use,
			Short:  topic.short,
			Long:   topic.long,
			Hidden: true,
		})
	}
}

// ---- topic content -------------------------------------------------------
//
// Kept as constants rather than markdown files so the binary ships
// self-contained — `retab help quickstart` works offline on a fresh
// install before the user has cloned any docs.
//
// Style notes for editors:
//   - Wrap prose at ~78 columns so help renders cleanly in 80-col terminals.
//   - Use two-space indent for shell snippets; cobra prints Long verbatim.
//   - Cross-reference other topics by their `use` name in backticks.

const topicQuickstart = `Quick Start — go from zero to extracted JSON in about a minute.

  # 1. Install (once, on this machine)
  curl -fsSL https://retab.com/install.sh | sh

  # 2. Authenticate. Either:
  retab auth login                  # interactive prompt; saved to ~/.retab/config.json
  # ...or set the env var:
  export RETAB_API_KEY=sk_...       # picked up by every subsequent command

  # 3. Confirm you're wired up
  retab auth status                 # prints credential source and verifies it

  # 4. Run your first extraction
  retab extractions create \
    --file ./invoice.pdf \
    --json-schema-file ./schema.json \
    --model gpt-4o

Output is indented JSON on stdout — pipe into "jq" or redirect to a file.

To stream tokens as the model produces them (useful for slow extractions):

  retab extractions stream \
    --file ./invoice.pdf \
    --json-schema-file ./schema.json \
    --model gpt-4o

This emits one JSON event per line (NDJSON), suitable for "jq -c", "head",
or piping into a tail-style consumer.

Next steps:
  retab help authentication      Credential precedence and rotation
  retab files --help             Upload documents and reuse by id
  retab workflows --help         Multi-step extraction pipelines
`

const topicAuthentication = `Authentication — how the CLI decides which API key to use.

The CLI resolves credentials in this order, first match wins:

  1. --api-key flag         (highest priority — explicit per-invocation override)
  2. RETAB_API_KEY env var  (process-wide override; great for CI)
  3. ~/.retab/config.json   (written by "retab auth login"; persistent)

Same precedence applies to --base-url / RETAB_BASE_URL / config base_url —
the lever for pointing the CLI at regional endpoints (EU, dedicated, or
self-hosted instances).

Common patterns:

  # Interactive login — secret entry doesn't echo
  retab auth login

  # Non-interactive login for CI/CD
  retab auth login --api-key="$RETAB_API_KEY"

  # One-shot override without persisting
  retab --api-key=sk_test_... files list

  # Inspect what the CLI is using right now
  retab auth status

  # Clear local credentials (does not revoke the key server-side)
  retab auth logout

Key rotation: "retab auth login" overwrites the saved key in place, so
re-running is always safe. Revoking happens in the Retab dashboard.

The config file lives at ~/.retab/config.json with mode 0600 (owner read /
write only). The parent directory is 0700. See "retab help configuration"
for layout.
`

const topicConfiguration = `Configuration File — ~/.retab/config.json

Written by "retab auth login" and read by every command that needs a
credential. Format is plain JSON with two optional fields:

  {
    "api_key":  "sk_live_...",
    "base_url": "https://api.eu.retab.com"
  }

File permissions are 0600 (parent dir 0700) — the file contains a secret.

You can edit it by hand to change "base_url" (e.g. switching between EU
and US regions) without touching the API key. The CLI re-reads the file
on every invocation, so there's nothing to restart.

Precedence with flags + env: see "retab help authentication".

The path is fixed at $HOME/.retab/config.json — there's no --config flag
on purpose. One canonical location per machine keeps debugging simple.
If you need per-project credentials, prefer RETAB_API_KEY scoped to a
direnv .envrc or a CI secret instead of a second config file.
`

const topicFormats = `Output Formats — what the CLI writes to stdout.

JSON (default for most commands)
  Indented JSON, one document per invocation. Suitable for piping to jq:

    retab files list | jq '.data[].id'

  HTML-significant characters (<, >, &) are NOT escaped — payloads
  round-trip through jq without re-encoding.

NDJSON (newline-delimited JSON, streaming)
  Used by "retab extractions stream" and similar streaming commands.
  One JSON object per line, flushed as it is produced:

    retab extractions stream \
        --file ./doc.pdf --json-schema-file ./s.json --model gpt-4o \
      | jq -c 'select(.type=="delta") | .data'

Binary
  "retab files download <id>" writes the raw file body. Pass "-o -" to
  send to stdout (handy for piping into another tool); otherwise it
  writes to the original filename, or a path you specify with "-o".

Exit codes
  0 on success. Non-zero on any error; the human-readable message is
  printed to stderr so stdout stays parseable:

    retab extractions get $ID >result.json 2>error.log
    test $? -eq 0 && jq . result.json

  130 is the conventional Ctrl-C — the CLI propagates SIGINT through the
  API request context, so long-running streams cancel cleanly.
`

const topicShellCompletion = `Shell Completion — tab-complete commands, subcommands, and flags.

Cobra generates completion scripts for bash, zsh, fish, and powershell.
Pick the line for your shell and add it to your rc file once:

  # bash (Linux: ~/.bashrc; macOS: ~/.bash_profile)
  source <(retab completion bash)

  # zsh — must have "compinit" loaded in your .zshrc already
  source <(retab completion zsh)

  # fish
  retab completion fish | source

  # powershell — append to $PROFILE
  retab completion powershell | Out-String | Invoke-Expression

To make it permanent system-wide on Linux:

  retab completion bash | sudo tee /etc/bash_completion.d/retab >/dev/null

After restarting your shell:
  - "retab e<TAB>" completes to "retab extractions"
  - "retab files l<TAB>" completes to "retab files list"
  - "retab files list --<TAB>" lists all flags
`
