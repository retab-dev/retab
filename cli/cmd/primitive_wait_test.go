package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func TestPrimitiveWaitCommandPollsUntilCompleted(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/parses/parse_wait" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		status := "running"
		if hits.Add(1) >= 2 {
			status = "completed"
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"parse_wait","status":"%s","markdown":"done"}`, status)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := primitiveWaitCommand(parseWaitSpec)
	addPrimitiveWaitTuningFlags(cmd, false)
	if err := cmd.Flags().Set("poll-interval-ms", "1"); err != nil {
		t.Fatal(err)
	}
	stdout, err := captureStdAndRun(t, func() error {
		return cmd.RunE(cmd, []string{"parse_wait"})
	})
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if got := hits.Load(); got != 2 {
		t.Fatalf("GET count = %d, want 2", got)
	}
	var resource map[string]any
	if err := json.Unmarshal([]byte(stdout), &resource); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if resource["status"] != "completed" {
		t.Fatalf("status = %v, want completed", resource["status"])
	}
}

func TestPrimitiveWaitCommandPrintsFinalRecordAndErrorsOnTerminalFailure(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/extractions/extr_bad" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"extr_bad","status":"error","error":{"message":"failed"}}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cmd := primitiveWaitCommand(extractionWaitSpec)
	addPrimitiveWaitTuningFlags(cmd, false)
	var err error
	stdout, stderr := captureStd(t, func() {
		err = cmd.RunE(cmd, []string{"extr_bad"})
	})
	if err == nil {
		t.Fatal("expected terminal error")
	}
	if !strings.Contains(stderr, "extraction extr_bad ended with status error") {
		t.Fatalf("stderr = %q", stderr)
	}
	if !strings.Contains(stdout, `"status": "error"`) {
		t.Fatalf("stdout should include final error record, got:\n%s", stdout)
	}
}

func TestParsesCreateWaitPollsFreshResource(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var getHits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/parses":
			_, _ = fmt.Fprint(w, `{"id":"parse_created","status":"pending"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/parses/parse_created":
			status := "running"
			if getHits.Add(1) >= 2 {
				status = "completed"
			}
			_, _ = fmt.Fprintf(w, `{"id":"parse_created","status":"%s","markdown":"final"}`, status)
		default:
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for name, value := range map[string]string{
		"url":              "https://example.com/report.pdf",
		"model":            "gpt-4o",
		"wait":             "true",
		"poll-interval-ms": "1",
	} {
		if err := parsesCreateCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set --%s: %v", name, err)
		}
	}
	t.Cleanup(func() {
		for name, value := range map[string]string{
			"url":              "",
			"model":            "",
			"wait":             "false",
			"poll-interval-ms": "2000",
			"timeout-seconds":  "600",
		} {
			_ = parsesCreateCmd.Flags().Set(name, value)
		}
	})

	stdout, err := captureStdAndRun(t, func() error {
		return parsesCreateCmd.RunE(parsesCreateCmd, nil)
	})
	if err != nil {
		t.Fatalf("parses create --wait: %v", err)
	}
	if got := getHits.Load(); got != 2 {
		t.Fatalf("GET count = %d, want 2", got)
	}
	if !strings.Contains(stdout, `"status": "completed"`) || !strings.Contains(stdout, `"markdown": "final"`) {
		t.Fatalf("stdout should include final parse record, got:\n%s", stdout)
	}
}

func TestPrimitiveCreateCommandsExposeWaitFlags(t *testing.T) {
	for name, cmd := range map[string]*cobra.Command{
		"classifications create": classificationsCreateCmd,
		"edits create":           editsCreateCmd,
		"extractions create":     extractionsCreateCmd,
		"parses create":          parsesCreateCmd,
		"partitions create":      partitionsCreateCmd,
		"splits create":          splitsCreateCmd,
	} {
		for _, flag := range []string{"wait", "poll-interval-ms", "timeout-seconds"} {
			if cmd.Flags().Lookup(flag) == nil {
				t.Fatalf("%s missing --%s", name, flag)
			}
		}
	}
}
