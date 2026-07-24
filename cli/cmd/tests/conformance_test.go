package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/retab-dev/retab/cli/cmd"
	"github.com/spf13/cobra"
)

// TestMain isolates this package from the developer's real ~/.retab config,
// mirroring the protection package cmd's own TestMain provides. os.UserHomeDir
// reads $HOME on unix and %USERPROFILE% on Windows, so repoint both at a
// throwaway dir; otherwise a command that loads/saves config could clobber a
// real environment selection or stored OAuth token.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "retab-cli-conformance-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create temp HOME:", err)
		os.Exit(1)
	}
	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

// --- discovery helpers -------------------------------------------------------

// walk visits cmd and every command beneath it.
func walk(c *cobra.Command, visit func(*cobra.Command)) {
	visit(c)
	for _, sub := range c.Commands() {
		walk(sub, visit)
	}
}

// commandsNamed returns every command in the tree whose final word is name
// (e.g. "list", "delete") and that has a RunE to invoke.
func commandsNamed(name string) []*cobra.Command {
	var out []*cobra.Command
	walk(cmd.RootCommand(), func(c *cobra.Command) {
		if c.Name() == name && c.RunE != nil {
			out = append(out, c)
		}
	})
	return out
}

func hasFlag(c *cobra.Command, name string) bool {
	return c.Flags().Lookup(name) != nil
}

// hasCobraMutex reports whether before/after are declared mutually exclusive via
// cobra's flag-group mechanism (enforced by cobra during arg parsing). The
// alternative is a hand-rolled RunE check; one of the two must be present.
func hasCobraMutex(c *cobra.Command, name string) bool {
	f := c.Flags().Lookup(name)
	if f == nil {
		return false
	}
	_, ok := f.Annotations["cobra_annotation_mutually_exclusive"]
	return ok
}

// --- invocation helpers ------------------------------------------------------

// setFlag sets a flag for the duration of the test and resets it afterward, so
// state never leaks between the shared, globally-registered commands.
func setFlag(t *testing.T, c *cobra.Command, name, value string) {
	t.Helper()
	if !hasFlag(c, name) {
		t.Fatalf("%s: missing --%s flag", c.CommandPath(), name)
	}
	if err := c.Flags().Set(name, value); err != nil {
		t.Fatalf("%s: set --%s=%q: %v", c.CommandPath(), name, value, err)
	}
	t.Cleanup(func() { _ = c.Flags().Set(name, "") })
}

// setRootOutput sets the persistent --output flag the renderers read, resetting
// it afterward.
func setRootOutput(t *testing.T, value string) {
	t.Helper()
	pf := cmd.RootCommand().PersistentFlags()
	if err := pf.Set("output", value); err != nil {
		t.Fatalf("set --output=%q: %v", value, err)
	}
	t.Cleanup(func() { _ = pf.Set("output", "") })
}

// scopeIfNeeded supplies a workflow id to commands that require one, so scoped
// list commands get past their "workflow id required" guard and exercise the
// behavior under test.
func scopeIfNeeded(t *testing.T, c *cobra.Command) {
	if hasFlag(c, "workflow-id") {
		setFlag(t, c, "workflow-id", "wf_conformance")
	}
}

// newCredentials points the CLI at a mock server with a test key, isolated per
// test via t.Setenv.
func newCredentials(t *testing.T, serverURL string) {
	t.Helper()
	// A test-scoped key prefix, not an arbitrary string. The production gate
	// now fails SAFE: a key whose environment cannot be placed from its prefix
	// is treated as production, so an unplaceable fixture key would make every
	// high-risk command in this suite hit the production confirmation before
	// reaching the behaviour under test (a delete would report "requires
	// --confirm" instead of its own "--yes" refusal). Real Retab keys carry a
	// recognizable prefix; the fixture should too.
	t.Setenv("RETAB_API_KEY", "rt_test_conformance")
	t.Setenv("RETAB_API_BASE_URL", serverURL)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
}

// runE invokes a command's RunE (bypassing cobra's arg parsing and the update
// notifier, exactly as the in-package tests do) while capturing stdout. A panic
// is converted to an error so one misbehaving command can't abort the whole
// suite; since RunE here skips cobra's arg validation, a panic usually just
// means the command needs positional args the caller didn't supply.
func runE(t *testing.T, c *cobra.Command, args []string) (out string, err error) {
	t.Helper()
	orig := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("pipe: %v", pipeErr)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var b strings.Builder
		_, _ = io.Copy(&b, r)
		done <- b.String()
	}()
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("panic: %v", rec)
			}
		}()
		err = c.RunE(c, args)
	}()
	_ = w.Close()
	os.Stdout = orig
	out = <-done
	_ = r.Close()
	return out, err
}

func looksLikeJSON(s string) bool {
	t := strings.TrimSpace(s)
	return strings.HasPrefix(t, "{") || strings.HasPrefix(t, "[")
}

// --- conformance tests -------------------------------------------------------

// Every primitive list command that exposes --filename and --from-date must
// actually forward those filters to the server. Regression: `partitions list`
// registered the flags but dropped them, returning an unfiltered list.
func TestListCommandsForwardFileDateFilters(t *testing.T) {
	var lists []*cobra.Command
	for _, c := range commandsNamed("list") {
		if hasFlag(c, "filename") && hasFlag(c, "from-date") && hasFlag(c, "to-date") {
			lists = append(lists, c)
		}
	}
	if len(lists) == 0 {
		t.Fatal("discovered no filterable list commands; discovery is broken")
	}

	for _, c := range lists {
		t.Run(c.CommandPath(), func(t *testing.T) {
			var got url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[],"list_metadata":{}}`)
			}))
			defer server.Close()
			newCredentials(t, server.URL)
			scopeIfNeeded(t, c)

			setFlag(t, c, "filename", "invoice.pdf")
			setFlag(t, c, "from-date", "2026-01-01T00:00:00Z")
			setFlag(t, c, "to-date", "2026-12-31T00:00:00Z")

			if _, err := runE(t, c, nil); err != nil {
				t.Fatalf("list: %v", err)
			}
			if got.Get("filename") != "invoice.pdf" {
				t.Errorf("--filename not forwarded: query=%v", got)
			}
			if got.Get("from_date") != "2026-01-01T00:00:00Z" {
				t.Errorf("--from-date not forwarded: query=%v", got)
			}
			if got.Get("to_date") != "2026-12-31T00:00:00Z" {
				t.Errorf("--to-date not forwarded: query=%v", got)
			}
		})
	}
}

// Every list command exposing both --before and --after must enforce their
// mutual exclusion — either via cobra's flag-group mechanism or a RunE check.
// Regression: `workflows experiments list` had neither, silently sending both.
func TestListCommandsEnforceBeforeAfterExclusivity(t *testing.T) {
	// Discover by FLAG, not by command name. Selecting only commands named
	// "list" left five flag-carrying commands unchecked — `usage runs`,
	// `usage blocks`, `usage primitives`, `workflows blocks history` and
	// `workflows blocks runs` — which mattered once cobra's flag-group backstop
	// was removed in favour of the uniform RunE message.
	var lists []*cobra.Command
	walk(cmd.RootCommand(), func(c *cobra.Command) {
		if c.RunE != nil && hasFlag(c, "before") && hasFlag(c, "after") {
			lists = append(lists, c)
		}
	})
	if len(lists) == 0 {
		t.Fatal("discovered no before/after list commands; discovery is broken")
	}

	for _, c := range lists {
		t.Run(c.CommandPath(), func(t *testing.T) {
			// Cobra's flag-group mechanism must NOT be used: its
			// validateFlagGroups stage runs before RunE and emits a noisy
			// message that shadows the concise one every other command prints.
			if hasCobraMutex(c, "before") {
				t.Fatalf("%s declares before/after via cobra's flag group; that shadows the concise RunE message", c.CommandPath())
			}
			// The command must reject the collision itself at RunE.
			// Pass the scope positionally (a dummy workflow id) rather than via
			// --workflow-id: scoped lists accept either form, and supplying both
			// would trip their "specified twice" guard before the mutex check.
			newCredentials(t, "http://127.0.0.1:0")
			setFlag(t, c, "before", "x")
			setFlag(t, c, "after", "y")

			_, err := runE(t, c, []string{"wf_conformance"})
			if err == nil {
				t.Fatalf("expected --before/--after collision to error")
			}
			if !strings.Contains(err.Error(), "mutually exclusive") {
				t.Fatalf("error should mention mutually exclusive, got: %v", err)
			}
		})
	}
}

// Every list command whose table renderer produces real output must also honor
// --output csv. Regression: `workflows tests list` only special-cased table and
// silently returned JSON for csv. Commands with no table renderer (JSON-only)
// are skipped — the rule is "if table works, csv must too".
func TestListCommandsHonorCSVOutput(t *testing.T) {
	for _, c := range commandsNamed("list") {
		c := c
		t.Run(c.CommandPath(), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"data":[{"id":"row_1","name":"row","status":"completed","created_at":"2026-01-01T00:00:00Z"}],"list_metadata":{}}`)
			}))
			defer server.Close()
			newCredentials(t, server.URL)
			// Pass the scope positionally so scoped lists run; non-scoped lists
			// ignore the extra arg.
			scopeArg := []string{"wf_conformance"}

			tableOut, tableErr := func() (string, error) {
				setRootOutput(t, "table")
				return runE(t, c, scopeArg)
			}()
			// Can't isolate csv behavior if the command can't render a table
			// here (needs args we don't supply, or is JSON-only). Not a csv bug.
			if tableErr != nil || looksLikeJSON(tableOut) {
				t.Skipf("no table rendering to compare against (err=%v)", tableErr)
			}

			setRootOutput(t, "csv")
			csvOut, csvErr := runE(t, c, scopeArg)
			if csvErr != nil {
				t.Fatalf("--output csv errored while table rendered: %v", csvErr)
			}
			if looksLikeJSON(csvOut) {
				t.Fatalf("--output csv fell back to JSON while table rendered:\n%s", csvOut)
			}
		})
	}
}

// Every delete command must refuse to act without --yes when stdin is not a
// terminal (as it is in tests), so piped/scripted invocations can't silently
// destroy data.
func TestDeleteCommandsRequireYesWhenNonInteractive(t *testing.T) {
	deletes := commandsNamed("delete")
	if len(deletes) == 0 {
		t.Fatal("discovered no delete commands; discovery is broken")
	}
	for _, c := range deletes {
		t.Run(c.CommandPath(), func(t *testing.T) {
			// A permissive server so any pre-confirm fetch succeeds; if the
			// command deletes anyway (no gate) the call returns nil and the
			// missing-gate assertion below fires.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{}`)
			}))
			defer server.Close()
			newCredentials(t, server.URL)
			scopeIfNeeded(t, c)

			_, err := runE(t, c, []string{"conformance_dummy_id"})
			if err == nil {
				t.Fatalf("delete without --yes should be refused on non-tty stdin")
			}
			if !strings.Contains(err.Error(), "--yes") {
				t.Fatalf("refusal should mention --yes, got: %v", err)
			}
		})
	}
}

// Every `wait` subcommand must expose --poll-interval-ms / --timeout-seconds.
// primitiveWaitCommand documents these flags (in its Example) and reads them
// at runtime, but they are registered by a separate addPrimitiveWaitTuningFlags
// call — easy to forget. Regression: `schemas wait` registered the command
// without the tuning flags, so the flags shown in its own help were unknown
// flags and the poll/timeout were silently pinned to the defaults.
func TestWaitCommandsExposeTuningFlags(t *testing.T) {
	waits := commandsNamed("wait")
	if len(waits) == 0 {
		t.Fatal("discovered no wait commands; discovery is broken")
	}
	for _, c := range waits {
		t.Run(c.CommandPath(), func(t *testing.T) {
			for _, flag := range []string{"poll-interval-ms", "timeout-seconds"} {
				if !hasFlag(c, flag) {
					t.Fatalf("%s: missing --%s flag", c.CommandPath(), flag)
				}
			}
		})
	}
}
