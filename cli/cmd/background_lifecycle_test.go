//go:build !retab_oagen_cli_parses && !retab_oagen_cli_splits && !retab_oagen_cli_extractions

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// This suite locks in the CLI-layer behavior of the background=true lifecycle
// (create --background, --wait polling, the standalone wait command, cancel, and
// error surfacing) that was verified by hand against production. It exercises the
// shared primitive_wait.go machinery and the parse/split/extraction create
// commands directly, with a fake API server — no network, deterministic timing.

// ---------------------------------------------------------------------------
// Pure helpers (no server)
// ---------------------------------------------------------------------------

func TestIsTerminalPrimitiveStatus(t *testing.T) {
	terminal := []string{"completed", "error", "failed", "cancelled"}
	nonTerminal := []string{"pending", "queued", "in_progress", "", "processing", "unknown"}
	for _, s := range terminal {
		if !isTerminalPrimitiveStatus(s) {
			t.Errorf("status %q should be terminal", s)
		}
	}
	for _, s := range nonTerminal {
		if isTerminalPrimitiveStatus(s) {
			t.Errorf("status %q should NOT be terminal", s)
		}
	}
}

func TestPrimitiveStatusReadsTopLevelAndLifecycle(t *testing.T) {
	t.Run("top-level status wins", func(t *testing.T) {
		got := primitiveStatus(map[string]any{"status": "completed"})
		if got != "completed" {
			t.Fatalf("status = %q, want completed", got)
		}
	})
	t.Run("falls back to lifecycle.status", func(t *testing.T) {
		got := primitiveStatus(map[string]any{"lifecycle": map[string]any{"status": "in_progress"}})
		if got != "in_progress" {
			t.Fatalf("status = %q, want in_progress", got)
		}
	})
	t.Run("empty when neither present", func(t *testing.T) {
		if got := primitiveStatus(map[string]any{"id": "x"}); got != "" {
			t.Fatalf("status = %q, want empty", got)
		}
	})
}

func TestPrimitiveTerminalError(t *testing.T) {
	spec := parseWaitSpec
	t.Run("completed is not an error", func(t *testing.T) {
		if err := primitiveTerminalError(spec, map[string]any{"id": "prs_1", "status": "completed"}); err != nil {
			t.Fatalf("completed should not error, got %v", err)
		}
	})
	t.Run("nil resource is not an error", func(t *testing.T) {
		if err := primitiveTerminalError(spec, nil); err != nil {
			t.Fatalf("nil should not error, got %v", err)
		}
	})
	for _, status := range []string{"error", "failed", "cancelled"} {
		t.Run(status+" surfaces with id", func(t *testing.T) {
			err := primitiveTerminalError(spec, map[string]any{"id": "prs_9", "status": status})
			if err == nil {
				t.Fatalf("status %q must surface an error", status)
			}
			if !strings.Contains(err.Error(), "prs_9") || !strings.Contains(err.Error(), status) {
				t.Fatalf("error %q must mention id and status", err.Error())
			}
		})
	}
	t.Run("error without id still surfaces", func(t *testing.T) {
		err := primitiveTerminalError(spec, map[string]any{"status": "failed"})
		if err == nil || !strings.Contains(err.Error(), "parse") {
			t.Fatalf("want an error naming the primitive, got %v", err)
		}
	})
}

func TestPrimitiveIDExtraction(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		id, err := primitiveID(map[string]any{"id": "extr_1"})
		if err != nil || id != "extr_1" {
			t.Fatalf("id=%q err=%v", id, err)
		}
	})
	t.Run("missing errors", func(t *testing.T) {
		if _, err := primitiveID(map[string]any{"status": "queued"}); err == nil {
			t.Fatal("missing id must error")
		}
	})
	t.Run("empty string errors", func(t *testing.T) {
		if _, err := primitiveID(map[string]any{"id": ""}); err == nil {
			t.Fatal("empty id must error")
		}
	})
}

func TestPrimitiveGetPathEscapesID(t *testing.T) {
	// Defensive: an id with URL-hostile characters must be path-escaped so a
	// crafted id can never break out of the primitive path.
	got := primitiveGetPath(extractionWaitSpec, "a/b?c#d")
	if strings.Contains(got, "?") || strings.Contains(got, "#") {
		t.Fatalf("path %q not escaped", got)
	}
	if !strings.HasPrefix(got, "/v1/extractions/") {
		t.Fatalf("path %q lost its prefix", got)
	}
}

func TestPrimitiveBackgroundParamAlwaysExplicit(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	addPrimitiveBackgroundFlag(cmd)

	// Default: explicit false, never nil — the server must receive an explicit
	// background value rather than inferring a default.
	if got := primitiveBackgroundParam(cmd); got == nil || *got {
		t.Fatalf("default background = %v, want explicit false", got)
	}
	if err := cmd.Flags().Set("background", "true"); err != nil {
		t.Fatal(err)
	}
	if got := primitiveBackgroundParam(cmd); got == nil || !*got {
		t.Fatalf("set background = %v, want explicit true", got)
	}
}

func TestPrimitiveWaitDurations(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	addPrimitiveWaitTuningFlags(cmd, true)

	t.Run("defaults", func(t *testing.T) {
		poll, timeout := primitiveWaitDurations(cmd)
		if poll != 2*time.Second {
			t.Errorf("poll = %v, want 2s", poll)
		}
		if timeout != 600*time.Second {
			t.Errorf("timeout = %v, want 600s", timeout)
		}
	})
	t.Run("custom", func(t *testing.T) {
		_ = cmd.Flags().Set("poll-interval-ms", "250")
		_ = cmd.Flags().Set("timeout-seconds", "5")
		poll, timeout := primitiveWaitDurations(cmd)
		if poll != 250*time.Millisecond {
			t.Errorf("poll = %v, want 250ms", poll)
		}
		if timeout != 5*time.Second {
			t.Errorf("timeout = %v, want 5s", timeout)
		}
	})
}

// ---------------------------------------------------------------------------
// waitForPrimitive against a fake server (the real polling loop)
// ---------------------------------------------------------------------------

// bgTestCmd returns a bare command whose request layer resolves base-url and
// api-key from the environment (set by the caller via t.Setenv), matching how
// the CLI resolves them when the persistent flags are unset.
func bgTestCmd() *cobra.Command { return &cobra.Command{Use: "bg-test"} }

func TestWaitForPrimitivePollsUntilCompleted(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		status := "queued"
		if n >= 3 {
			status = "completed"
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"prs_1","status":%q,"output":{"text":"hi"}}`, status)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	final, err := waitForPrimitive(context.Background(), bgTestCmd(), parseWaitSpec, "prs_1", time.Millisecond, 10*time.Second)
	if err != nil {
		t.Fatalf("waitForPrimitive: %v", err)
	}
	if primitiveStatus(final) != "completed" {
		t.Fatalf("final status = %q, want completed", primitiveStatus(final))
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("polled %d times, want 3 (stop as soon as terminal)", got)
	}
}

func TestWaitForPrimitiveReturnsTerminalFailedRecord(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"extr_2","status":"failed"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// waitForPrimitive stops at any terminal status with a nil error; the
	// non-zero exit is produced by primitiveTerminalError on the returned record.
	final, err := waitForPrimitive(context.Background(), bgTestCmd(), extractionWaitSpec, "extr_2", time.Millisecond, 5*time.Second)
	if err != nil {
		t.Fatalf("waitForPrimitive should not itself error on a terminal record: %v", err)
	}
	if termErr := primitiveTerminalError(extractionWaitSpec, final); termErr == nil {
		t.Fatal("a failed record must yield a terminal error for the exit code")
	}
}

func TestWaitForPrimitiveTimesOut(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"splt_3","status":"queued"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// Poll interval >> timeout: exactly one (successful) poll runs, then the
	// deadline fires deterministically during the long inter-poll sleep — no
	// request is in flight when it fires, so this is race-free.
	final, err := waitForPrimitive(context.Background(), bgTestCmd(), splitWaitSpec, "splt_3", 10*time.Second, 100*time.Millisecond)
	if err == nil {
		t.Fatal("a never-terminal primitive must time out")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("error %q should mention the timeout", err.Error())
	}
	// The last-seen (still queued) record is returned so the caller can print it.
	if final == nil || primitiveStatus(final) != "queued" {
		t.Fatalf("timeout should return the last-seen record, got %v", final)
	}
}

func TestWaitForPrimitiveToleratesTransientPollErrors(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			// Simulate a redeploy / brief unavailability — must NOT abort the wait.
			http.Error(w, `{"detail":"upstream unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"prs_4","status":"completed"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	final, err := waitForPrimitive(context.Background(), bgTestCmd(), parseWaitSpec, "prs_4", time.Millisecond, 10*time.Second)
	if err != nil {
		t.Fatalf("transient poll errors must be tolerated, got %v", err)
	}
	if primitiveStatus(final) != "completed" {
		t.Fatalf("final status = %q, want completed after recovery", primitiveStatus(final))
	}
}

func TestWaitForPrimitivePersistentErrorSurfacesAfterTimeout(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"still broken"}`, http.StatusBadGateway)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	// Poll interval >> timeout: the single fast-failing poll sets lastErr to the
	// 502 (a non-context error), then the deadline fires during the long sleep, so
	// the "repeated poll errors" branch is taken deterministically rather than the
	// bare "timed out" branch that a mid-request cancellation would produce.
	_, err := waitForPrimitive(context.Background(), bgTestCmd(), parseWaitSpec, "prs_5", 10*time.Second, 100*time.Millisecond)
	if err == nil {
		t.Fatal("a persistently failing poll must surface an error at timeout")
	}
	if !strings.Contains(err.Error(), "repeated poll errors") {
		t.Fatalf("error %q should attribute the failure to repeated poll errors", err.Error())
	}
}

func TestWaitForPrimitiveRecognizesLifecycleStatus(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"prs_6","lifecycle":{"status":"completed"}}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	final, err := waitForPrimitive(context.Background(), bgTestCmd(), parseWaitSpec, "prs_6", time.Millisecond, 5*time.Second)
	if err != nil {
		t.Fatalf("waitForPrimitive: %v", err)
	}
	if primitiveStatus(final) != "completed" {
		t.Fatalf("lifecycle.status not recognized, status = %q", primitiveStatus(final))
	}
}

// ---------------------------------------------------------------------------
// Create-command background wiring (drive the real RunE)
// ---------------------------------------------------------------------------

func bgTempDoc(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "doc.pdf")
	if err := os.WriteFile(p, []byte("%PDF-1.4 tiny"), 0o600); err != nil {
		t.Fatalf("write doc: %v", err)
	}
	return p
}

// runCreateCapturingBody drives a create command's RunE against a fake server
// that records the POST create body and answers any GET polls with `getStatus`.
// It returns the decoded create request body and the number of GET polls.
func runCreateCapturingBody(t *testing.T, createPath string, buildCmd func() *cobra.Command, set func(*cobra.Command), getStatus string) (map[string]any, int) {
	t.Helper()
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	var polls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == createPath:
			raw, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatalf("decode create body: %v (raw=%s)", err, raw)
			}
			// A background create returns a queued record; --wait then polls GET.
			_, _ = fmt.Fprint(w, `{"id":"prim_1","status":"queued"}`)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, createPath+"/"):
			polls.Add(1)
			_, _ = fmt.Fprintf(w, `{"id":"prim_1","status":%q,"output":{"text":"ok"}}`, getStatus)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := buildCmd()
	_ = cmd.Flags().Set("file", bgTempDoc(t))
	_ = cmd.Flags().Set("model", "retab-small")
	set(cmd)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if body == nil {
		t.Fatal("server never received the create request")
	}
	return body, int(polls.Load())
}

func newParseCreateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test-parse", RunE: parsesCreateCmd.RunE}
	addDocumentFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().String("table-parsing-format", "", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("bust-cache", false, "")
	addPrimitiveBackgroundFlag(cmd)
	addPrimitiveCreateWaitFlags(cmd)
	return cmd
}

func newSplitCreateTestCmd(t *testing.T) *cobra.Command {
	cmd := &cobra.Command{Use: "test-split", RunE: splitsCreateCmd.RunE}
	addDocumentFlags(cmd)
	subs := filepath.Join(t.TempDir(), "subs.json")
	if err := os.WriteFile(subs, []byte(`[{"name":"invoice","description":"bill"}]`), 0o600); err != nil {
		t.Fatalf("write subs: %v", err)
	}
	cmd.Flags().String("subdocuments-file", "", "")
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 8}, "n-consensus", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().String("instructions", "", "")
	addPrimitiveBackgroundFlag(cmd)
	addPrimitiveCreateWaitFlags(cmd)
	_ = cmd.Flags().Set("subdocuments-file", subs)
	return cmd
}

func newExtractionCreateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test-extraction", RunE: extractionsCreateCmd.RunE}
	addDocumentFlags(cmd)
	addSchemaFlags(cmd)
	cmd.Flags().String("model", "", "")
	cmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 16}, "n-consensus", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("bust-cache", false, "")
	cmd.Flags().StringArray("metadata", nil, "")
	cmd.Flags().String("messages-file", "", "")
	addPrimitiveBackgroundFlag(cmd)
	addPrimitiveCreateWaitFlags(cmd)
	_ = cmd.Flags().Set("json-schema", `{"type":"object","properties":{"x":{"type":"string"}}}`)
	return cmd
}

func TestCreateBackgroundFlagSerialization(t *testing.T) {
	cases := []struct {
		name       string
		createPath string
		build      func(*testing.T) func() *cobra.Command
	}{
		{"parse", "/v1/parses", func(*testing.T) func() *cobra.Command { return newParseCreateTestCmd }},
		{"split", "/v1/splits", func(t *testing.T) func() *cobra.Command {
			return func() *cobra.Command { return newSplitCreateTestCmd(t) }
		}},
		{"extraction", "/v1/extractions", func(*testing.T) func() *cobra.Command { return newExtractionCreateTestCmd }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" background=true forwarded", func(t *testing.T) {
			body, _ := runCreateCapturingBody(t, tc.createPath, tc.build(t), func(c *cobra.Command) {
				_ = c.Flags().Set("background", "true")
			}, "completed")
			bg, present := body["background"]
			if !present {
				t.Fatal("background must be sent explicitly")
			}
			if bg != true {
				t.Fatalf("background = %v, want true", bg)
			}
		})
		t.Run(tc.name+" background defaults to explicit false", func(t *testing.T) {
			body, _ := runCreateCapturingBody(t, tc.createPath, tc.build(t), func(*cobra.Command) {}, "completed")
			bg, present := body["background"]
			if !present {
				t.Fatal("background must be sent explicitly even when unset")
			}
			if bg != false {
				t.Fatalf("background = %v, want false", bg)
			}
		})
	}
}

func TestBackgroundCreateWaitPollsToCompletion(t *testing.T) {
	// --background --wait: create returns queued, then the CLI polls GET until
	// the record reaches a terminal status. Assert the poll actually happened.
	_, polls := runCreateCapturingBody(t, "/v1/parses", newParseCreateTestCmd, func(c *cobra.Command) {
		_ = c.Flags().Set("background", "true")
		_ = c.Flags().Set("wait", "true")
		_ = c.Flags().Set("poll-interval-ms", "1")
	}, "completed")
	if polls < 1 {
		t.Fatalf("--background --wait must poll GET at least once, polled %d", polls)
	}
}

func TestBackgroundCreateWithoutWaitDoesNotPoll(t *testing.T) {
	// Without --wait, a background create returns the queued record immediately
	// and must NOT poll (the user drives polling themselves).
	_, polls := runCreateCapturingBody(t, "/v1/extractions", newExtractionCreateTestCmd, func(c *cobra.Command) {
		_ = c.Flags().Set("background", "true")
	}, "completed")
	if polls != 0 {
		t.Fatalf("background create without --wait must not poll, polled %d", polls)
	}
}
