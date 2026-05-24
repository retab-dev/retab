package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Pins the security contract for `--debug`: when the user dumps wire-level
// request/response for a bug report, the API key (and any OAuth Bearer
// token) MUST be redacted. The unredacted version is a leak — bug reports
// get pasted into chat, GitHub issues, screenshots.

func TestDebugTransport_RedactsApiKey(t *testing.T) {
	// Stand up a noop upstream that just answers 200.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	stderr := captureStderr(t)
	defer stderr.restore()

	tr := &debugTransport{wrapped: http.DefaultTransport}
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/v1/files/x", nil)
	req.Header.Set("Api-Key", "sk_retab_VERYSECRETVALUE_dont_leak_me")
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	got := stderr.read()
	if strings.Contains(got, "sk_retab_VERYSECRETVALUE_dont_leak_me") {
		t.Errorf("Api-Key value leaked into --debug output:\n%s", got)
	}
	// The redacted form must STILL appear so the user can confirm WHICH
	// key got used. redactKey reveals the first 4 + last 4 chars; the
	// rest is asterisks. For our input "...dont_leak_me", that's "k_me".
	if !strings.Contains(got, "sk_r") || !strings.Contains(got, "k_me") {
		t.Errorf("redacted Api-Key preview missing — user can't disambiguate keys:\n%s", got)
	}
	if !strings.Contains(got, "Api-Key: ") {
		t.Errorf("Api-Key header missing entirely (over-redaction):\n%s", got)
	}
}

func TestDebugTransport_RedactsBearerTokenButKeepsScheme(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stderr := captureStderr(t)
	defer stderr.restore()

	tr := &debugTransport{wrapped: http.DefaultTransport}
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("Authorization", "Bearer wkos_at_VERYSECRETACCESSTOKEN_xxx")
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	got := stderr.read()
	if strings.Contains(got, "VERYSECRETACCESSTOKEN") {
		t.Errorf("Bearer token leaked:\n%s", got)
	}
	// Scheme is preserved — useful for diagnosing "why is the CLI sending
	// Bearer instead of Api-Key?" auth flow questions.
	if !strings.Contains(got, "Authorization: Bearer ") {
		t.Errorf("expected scheme preserved (`Authorization: Bearer ...`):\n%s", got)
	}
}

func TestDebugTransport_NonSensitiveHeadersUnchanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stderr := captureStderr(t)
	defer stderr.restore()

	tr := &debugTransport{wrapped: http.DefaultTransport}
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("Api-Key", "sk_retab_secret123456789012")
	req.Header.Set("X-Request-Id", "req_safe_to_show")
	req.Header.Set("User-Agent", "retab-cli/test")
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	got := stderr.read()
	// Non-sensitive headers must still appear in full so debug output stays
	// useful for diagnosing routing / tracing / user-agent issues.
	if !strings.Contains(got, "req_safe_to_show") {
		t.Errorf("X-Request-Id was redacted (it shouldn't be):\n%s", got)
	}
	if !strings.Contains(got, "retab-cli/test") {
		t.Errorf("User-Agent was redacted (it shouldn't be):\n%s", got)
	}
}

// The clone-before-mutate property: the outgoing wire request must NOT
// have its Api-Key header redacted by the dump. A regression here would
// silently break every authenticated request when --debug is on.
func TestDebugTransport_OutgoingRequestStillAuthenticated(t *testing.T) {
	var sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.Header.Get("Api-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stderr := captureStderr(t)
	defer stderr.restore()

	tr := &debugTransport{wrapped: http.DefaultTransport}
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("Api-Key", "sk_retab_real_token_should_reach_server_abc123")
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if sawAPIKey != "sk_retab_real_token_should_reach_server_abc123" {
		t.Errorf("server received redacted Api-Key %q — CLI broke auth while trying to redact debug output", sawAPIKey)
	}
}

func TestDebugTransport_PreservesRequestAndResponseBodies(t *testing.T) {
	var sawBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		sawBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	stderr := captureStderr(t)
	defer stderr.restore()

	tr := &debugTransport{wrapped: http.DefaultTransport}
	req, _ := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(`{"hello":"world"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if sawBody != `{"hello":"world"}` {
		t.Fatalf("server saw body %q, want original request body", sawBody)
	}
	if string(responseBody) != `{"ok":true}` {
		t.Fatalf("client saw response body %q, want original response body", string(responseBody))
	}
	got := stderr.read()
	if !strings.Contains(got, `{"hello":"world"}`) || !strings.Contains(got, `{"ok":true}`) {
		t.Fatalf("debug dump missing request or response body:\n%s", got)
	}
}

// stderrCapture redirects os.Stderr to a buffer for assertions, with a
// proper restore on cleanup. This pattern is common enough that a couple
// of the other tests in this package would benefit from it too, but
// we're scoping the helper here to avoid scope creep on this PR.
type stderrCapture struct {
	t        *testing.T
	original *os.File
	r        *os.File
	w        *os.File
	buf      bytes.Buffer
	done     chan struct{}
}

func captureStderr(t *testing.T) *stderrCapture {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	s := &stderrCapture{
		t: t, original: os.Stderr, r: r, w: w, done: make(chan struct{}),
	}
	os.Stderr = w
	go func() {
		_, _ = io.Copy(&s.buf, r)
		close(s.done)
	}()
	return s
}

// read closes the writer to flush, then returns everything captured.
// Safe to call once; subsequent calls return the same accumulated buffer.
func (s *stderrCapture) read() string {
	if s.w != nil {
		_ = s.w.Close()
		s.w = nil
		<-s.done
	}
	return s.buf.String()
}

func (s *stderrCapture) restore() {
	if s.w != nil {
		_ = s.w.Close()
		s.w = nil
		<-s.done
	}
	os.Stderr = s.original
	_ = s.r.Close()
}
