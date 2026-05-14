package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// `auth status` is the single most common script-and-eyeball straddle in
// the CLI: humans run it at the prompt to verify they're logged in, and
// scripts pipe it into `jq` to gate other commands. The tests below pin
// the two contracts that, if broken, would silently degrade either side:
//
//  1. Non-TTY writer → JSON. Anything that looks like a pipe or redirect
//     must keep emitting the legacy JSON shape so `auth status | jq` keeps
//     working.
//  2. --json flag → JSON. Even on a TTY, scripts that want JSON should
//     get JSON by setting the flag (and never have to spoof a non-TTY).
//  3. Human output → 3-line block with the expected labels.
//  4. ANSI escapes NEVER leak into a non-TTY destination. `bytes.Buffer`
//     is the prototypical non-TTY case — a stray `\x1b[` there means
//     `auth status > status.txt` would corrupt the file with control
//     codes.

// Sample payload mirrors the shape the real handler builds when the user
// is authenticated via the env var and the workflows probe succeeded.
func sampleAuthStatus() map[string]any {
	return map[string]any{
		"authenticated":   true,
		"source":          "RETAB_API_KEY env",
		"valid":           true,
		"api_key_preview": "sk_r********************************************xUxQ",
	}
}

// Non-TTY destination (bytes.Buffer) must produce JSON, not the human
// block — so `auth status > out.txt` and `auth status | jq` keep working.
func TestWriteAuthStatus_NonTTYWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatus(&buf, sampleAuthStatus(), false); err != nil {
		t.Fatalf("writeAuthStatus: %v", err)
	}

	// Must parse as JSON.
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("non-TTY output should be JSON; unmarshal failed: %v\n%s", err, buf.String())
	}
	// And carry every key we put in.
	for _, k := range []string{"authenticated", "source", "valid", "api_key_preview"} {
		if _, ok := decoded[k]; !ok {
			t.Errorf("non-TTY JSON output missing key %q:\n%s", k, buf.String())
		}
	}
	// "Logged in as" is the human-mode header — must NOT appear when we
	// emit JSON, or scripts would choke parsing it.
	if strings.Contains(buf.String(), "Logged in as") {
		t.Errorf("non-TTY output should not contain human-mode prose:\n%s", buf.String())
	}
}

// --json forces JSON even when the destination would otherwise render
// human (e.g. a TTY). bytes.Buffer is already non-TTY so we get JSON
// regardless — what we're really pinning here is that --json doesn't
// flip us into human mode by accident.
func TestWriteAuthStatus_JSONFlagForcesJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatus(&buf, sampleAuthStatus(), true); err != nil {
		t.Fatalf("writeAuthStatus: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("--json forced JSON should parse; unmarshal failed: %v\n%s", err, buf.String())
	}
	if strings.Contains(buf.String(), "Logged in as") {
		t.Errorf("--json forced output should not contain human-mode prose:\n%s", buf.String())
	}
}

// JSON output must be byte-equivalent to the legacy `printJSON(out)` form
// — same field set, same alphabetical key order, same 2-space indent,
// same trailing newline. Anything else breaks existing `jq` callsites.
func TestWriteAuthStatusJSON_ShapeIsByteEquivalent(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatusJSON(&buf, sampleAuthStatus()); err != nil {
		t.Fatalf("writeAuthStatusJSON: %v", err)
	}
	// Reproduce the exact legacy form: encoding/json.Encoder with 2-space
	// indent, no HTML escape, trailing newline.
	var want bytes.Buffer
	enc := json.NewEncoder(&want)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(sampleAuthStatus()); err != nil {
		t.Fatalf("encode legacy form: %v", err)
	}
	if buf.String() != want.String() {
		t.Errorf("JSON byte-equivalence regressed.\n got: %q\nwant: %q", buf.String(), want.String())
	}
}

// Human mode (forced by calling writeAuthStatusHuman directly) must
// render the three expected lines. We don't pin exact wording beyond
// these labels so the prose can be tweaked without breaking the test.
func TestWriteAuthStatusHuman_Has3LineBlock(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatusHuman(&buf, sampleAuthStatus()); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Logged in as", "Source:", "Status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("human output missing %q:\n%s", want, out)
		}
	}
	// Status line should report "valid" when the probe succeeded.
	if !strings.Contains(out, "valid") {
		t.Errorf("human output should report 'valid' when out[\"valid\"]==true:\n%s", out)
	}
}

// Even though writeAuthStatusHuman renders the human block, the writer
// here is a bytes.Buffer (non-TTY) — paletteFor returns the zero palette
// for non-files, so NO ANSI escapes should appear. Catches the bug where
// someone hardcodes "\x1b[…" instead of going through the palette.
func TestWriteAuthStatusHuman_NoANSILeakForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatusHuman(&buf, sampleAuthStatus()); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("ANSI escape leaked into non-TTY human output:\n%s", buf.String())
	}
}

// writeAuthStatus (the dispatch entry point) routes a non-TTY writer to
// JSON regardless of the --json flag. Pair-check with the human test:
// the same payload that produces a JSON shape here produces a human block
// only when forced through writeAuthStatusHuman.
func TestWriteAuthStatus_NoANSILeakForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	if err := writeAuthStatus(&buf, sampleAuthStatus(), false); err != nil {
		t.Fatalf("writeAuthStatus: %v", err)
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("ANSI escape leaked into non-TTY dispatched output:\n%s", buf.String())
	}
}

// "Not logged in" path — when no credential resolved, api_key_preview is
// absent. Human output should fall back gracefully (no blank "Logged in
// as " line, no panic on a missing key).
func TestWriteAuthStatusHuman_NotLoggedIn(t *testing.T) {
	var buf bytes.Buffer
	payload := map[string]any{
		"authenticated": false,
		"source":        "",
		"hint":          "run `retab auth login` to authenticate",
	}
	if err := writeAuthStatusHuman(&buf, payload); err != nil {
		t.Fatalf("writeAuthStatusHuman: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Not logged in") {
		t.Errorf("expected 'Not logged in' on empty-source path:\n%s", out)
	}
	// No half-rendered "Logged in as " line with a missing preview.
	if strings.Contains(out, "Logged in as \n") || strings.Contains(out, "Logged in as <") {
		t.Errorf("rendered a malformed 'Logged in as' line for empty preview:\n%s", out)
	}
}
