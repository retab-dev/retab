package cmd

import (
	"strings"
	"testing"
)

// The root command sets SilenceUsage so a failed API call does not dump a wall of
// help. Cobra applies that to POSITIONAL ARGUMENT errors too, which is where the
// usage line is most useful: `retab workflows stats` used to answer only
// "accepts 1 arg(s), received 0" — never mentioning that it wants a workflow id,
// and never showing how to find out.
//
// enrichPositionalArgErrors restores the usage line for argument errors only.

func TestPositionalArgErrorsNameWhatTheCommandExpects(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	for _, tc := range []struct {
		args        []string
		wantInUsage string
	}{
		{[]string{"workflows", "stats"}, "<workflow-id>"},
		{[]string{"workflows", "get"}, "retab workflows get"},
		{[]string{"workflows", "runs", "get"}, "retab workflows runs get"},
		{[]string{"projects", "get"}, "retab projects get"},
		{[]string{"workflows", "artifacts", "list"}, "retab workflows artifacts list"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			err := runRootForTest(t, tc.args...)
			if err == nil {
				t.Fatalf("retab %s: expected an argument error", strings.Join(tc.args, " "))
			}
			message := err.Error()

			// The original cobra wording is preserved (wrapped, not replaced), so
			// anything matching on it keeps working.
			if !strings.Contains(message, "arg(s), received") {
				t.Fatalf("the original argument error was lost: %q", message)
			}
			if !strings.Contains(message, "Usage:") {
				t.Fatalf("argument errors must show the usage line, got: %q", message)
			}
			if !strings.Contains(message, tc.wantInUsage) {
				t.Fatalf("usage line does not mention %q, got: %q", tc.wantInUsage, message)
			}
			if !strings.Contains(message, "--help") {
				t.Fatalf("argument errors should point at --help, got: %q", message)
			}
		})
	}
}

// TestPositionalArgErrorsAreNotDoubleWrapped: executeRoot runs once per process
// normally but repeatedly across tests, and the wrapper is installed by walking
// the shared command tree. Without an idempotency guard the usage block would be
// appended once per run, producing steadily longer errors.
func TestPositionalArgErrorsAreNotDoubleWrapped(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var last string
	for range 3 {
		err := runRootForTest(t, "workflows", "stats")
		if err == nil {
			t.Fatal("expected an argument error")
		}
		if count := strings.Count(err.Error(), "Usage:"); count != 1 {
			t.Fatalf("usage block appeared %d times — the wrapper is being applied repeatedly:\n%s",
				count, err.Error())
		}
		if last != "" && err.Error() != last {
			t.Fatalf("the same argument error changed between runs:\n first %q\n then  %q", last, err.Error())
		}
		last = err.Error()
	}
}

// TestCommandsWithHandWrittenArgChecksAreUnchanged: several commands validate
// their inputs themselves and already produce a good message, which they RENDER
// themselves and then return as a silent sentinel. The wrapper only touches
// cobra's positional validators, so those commands must be untouched — no usage
// block bolted onto an already-clear message.
func TestCommandsWithHandWrittenArgChecksAreUnchanged(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	err := runRootForTest(t, "workflows", "evals", "list")
	if err == nil {
		t.Fatal("expected an error for a missing workflow id")
	}
	// These commands render their own message and return the silent sentinel, so
	// the error carries no text. What matters is that the wrapper did not turn it
	// into a cobra-style argument error with a usage block appended.
	if strings.Contains(err.Error(), "Usage:") {
		t.Fatalf("a hand-validated command should not gain a usage block: %q", err.Error())
	}
	if strings.Contains(err.Error(), "arg(s), received") {
		t.Fatalf("a hand-validated command should not report a cobra arg-count error: %q", err.Error())
	}
}
