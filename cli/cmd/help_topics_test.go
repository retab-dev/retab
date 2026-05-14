package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// Tests for the topical help subsystem (help_topics.go).
//
// Topics are user-facing surface — these tests pin the contracts that
// would silently degrade UX if broken:
//
//   1. Every topic in `helpTopics` is registered as a hidden subcommand
//      reachable via `retab help <topic>`. If a topic is added to the
//      slice but the init() registration drifts, users get "unknown
//      command" instead of help text.
//   2. Topics appear in a "Topics:" section of the polished root help,
//      in the order declared in `helpTopics`. Re-ordering the slice is
//      the only lever for changing on-screen ordering.
//   3. Topics are Hidden — they must NOT appear in any of the command
//      groups (Primitives, Workflows, …). Mixing them would mislead
//      users into thinking topics are runnable commands.
//   4. Every topic has non-trivial `Long` content. A topic that ships
//      with just `Short` is effectively a 404 — `retab help <topic>`
//      would show no body.

func TestHelpTopics_RegisteredOnRootInOrder(t *testing.T) {
	for _, want := range helpTopics {
		found, _, err := rootCmd.Find([]string{want.use})
		if err != nil {
			t.Errorf("topic %q is in helpTopics but not registered on rootCmd: %v", want.use, err)
			continue
		}
		if found.Name() != want.use {
			t.Errorf("topic lookup for %q resolved to %q (cobra disambiguation?)", want.use, found.Name())
		}
		if !found.Hidden {
			t.Errorf("topic %q must be Hidden so it doesn't pollute the command menu", want.use)
		}
		if found.Runnable() {
			t.Errorf("topic %q must have no Run/RunE so cobra prints Long on invoke", want.use)
		}
	}
}

func TestHelpTopics_AllHaveNontrivialLong(t *testing.T) {
	for _, topic := range helpTopics {
		// Arbitrary threshold — short enough to not require padding, long
		// enough to catch accidental empty / placeholder strings.
		if len(strings.TrimSpace(topic.long)) < 80 {
			t.Errorf("topic %q has trivial Long (%d bytes after trim) — looks like a stub",
				topic.use, len(strings.TrimSpace(topic.long)))
		}
		// Each topic should mention its own name somewhere in its body so
		// users searching for "quickstart" in the rendered output get a
		// confirming signal that they're in the right place. Defensive
		// check rather than a strict contract.
		if !strings.Contains(topic.long, topic.use) {
			t.Logf("note: topic %q does not mention its own name in body — may be fine",
				topic.use)
		}
	}
}

func TestRenderRootHelp_TopicsSectionIsPresent(t *testing.T) {
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()
	if !strings.Contains(out, "Topics:") {
		t.Fatalf("rendered root help is missing the 'Topics:' section:\n%s", out)
	}
	for _, topic := range helpTopics {
		// The menu row is "  <name><pad>  <short>"; the trailing space after
		// the name is reliable across colour modes (ANSI escapes wrap the
		// name itself, so the bytes after `</R>` are plain spaces).
		// We assert the name appears AFTER the "Topics:" header to rule out
		// false matches against other parts of the output.
		topicsIdx := strings.Index(out, "Topics:")
		if !strings.Contains(out[topicsIdx:], topic.use) {
			t.Errorf("topic %q is missing from the rendered Topics section:\n%s",
				topic.use, out)
		}
	}
}

func TestRenderRootHelp_TopicsAppearInDeclaredOrder(t *testing.T) {
	var buf bytes.Buffer
	renderRootHelp(&buf, rootCmd)
	out := buf.String()
	// Only consider output from the "Topics:" header onward; earlier
	// sections may legitimately mention topic-like words.
	topicsIdx := strings.Index(out, "Topics:")
	if topicsIdx < 0 {
		t.Fatalf("Topics: section missing — see TestRenderRootHelp_TopicsSectionIsPresent")
	}
	tail := out[topicsIdx:]
	prev := -1
	for _, topic := range helpTopics {
		idx := strings.Index(tail, topic.use)
		if idx < 0 {
			continue // caught by the other test
		}
		if idx < prev {
			t.Errorf("topic %q (idx=%d) appears before a previously-declared topic (idx=%d)",
				topic.use, idx, prev)
		}
		prev = idx
	}
}

func TestRenderRootHelp_TopicsAreNotInCommandGroups(t *testing.T) {
	// A topic accidentally listed in `commandGroups` would render twice
	// (once as a top-level command, once in Topics:), and users would
	// expect it to be runnable. Hard-stop this class of bug.
	for _, g := range commandGroups {
		for _, cmdName := range g.commands {
			for _, topic := range helpTopics {
				if cmdName == topic.use {
					t.Errorf("topic %q is also listed in command group %q — pick one",
						topic.use, g.title)
				}
			}
		}
	}
}

// `retab help <topic>` is the documented entrypoint — verify cobra's
// auto-help routes through our topic command and renders its Long.
func TestHelpTopics_HelpCommandRendersTopicLong(t *testing.T) {
	for _, topic := range helpTopics {
		t.Run(topic.use, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			t.Cleanup(func() { rootCmd.SetOut(nil); rootCmd.SetErr(nil) })
			rootCmd.SetArgs([]string{"help", topic.use})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("rootCmd.Execute([help %s]) returned error: %v", topic.use, err)
			}
			out := buf.String()
			// First line of Long should appear — strong evidence the topic
			// command's help got invoked and its body printed.
			firstLine := strings.SplitN(strings.TrimSpace(topic.long), "\n", 2)[0]
			if !strings.Contains(out, firstLine) {
				t.Errorf("`retab help %s` did not print topic body. First line of Long:\n  %q\nOutput:\n%s",
					topic.use, firstLine, out)
			}
		})
	}
}
