package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Regression guard for hardenLeafArgs: list and other zero-positional-arg
// commands must reject unexpected positional arguments instead of silently
// ignoring them.
//
// Cobra's default when a command leaves Args nil is ArbitraryArgs, so
// `retab files list file_abc123` — a plausible typo for a get/filter —
// used to exit 0 and return the *entire* list, silently dropping the
// argument. hardenLeafArgs walks the tree and applies cobra.NoArgs to
// every runnable command that never declared its own Args validator.
//
// Reuses runRootForTest from unknown_subcommand_test.go (same package).
func TestListCommandsRejectExtraPositionalArgs(t *testing.T) {
	cases := [][]string{
		{"files", "list", "file_abc123"},
		{"extractions", "list", "extra"},
		{"parses", "list", "extra"},
		{"workflows", "list", "extra"},
		{"auth", "status", "extra"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := runRootForTest(t, args...)
			if err == nil {
				t.Fatalf("retab %s: expected an error for an unexpected positional arg, got nil (would exit 0)", strings.Join(args, " "))
			}
			if !strings.Contains(err.Error(), "unknown command") {
				t.Fatalf("retab %s: expected an \"unknown command\" error, got: %v", strings.Join(args, " "), err)
			}
		})
	}
}

// TestArgTakingCommandsStillAcceptTheirArg makes sure hardenLeafArgs did
// not over-reach: a command that declares an explicit Args validator must
// keep accepting its positional argument. We assert the failure (if any)
// is NOT an arg-count rejection — these ids don't exist, so a server
// round-trip error is fine; "accepts 0 arg(s)" is not.
func TestArgTakingCommandsStillAcceptTheirArg(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"detail": "not found"})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for _, args := range [][]string{
		{"files", "get", "file_nonexistent"},
		{"workflows", "get", "wf_nonexistent"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := runRootForTest(t, args...)
			if err != nil && strings.Contains(err.Error(), "accepts 0 arg") {
				t.Fatalf("retab %s: hardenLeafArgs wrongly stripped the positional arg: %v", strings.Join(args, " "), err)
			}
		})
	}
}
