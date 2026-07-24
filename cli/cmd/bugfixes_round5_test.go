//go:build !retab_oagen_cli_files

package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// inspectRender must KEEP the temp directory it reports on success and REMOVE
// it on every failure. Only the failure half was covered; deleting the
// success-path `ownedTempDir = ""` reset made a successful --render delete the
// very directory its JSON pointed at, with the whole suite still green.
//
// The LiteParser seam (resolveLiteParserFn) makes both halves testable without
// a real `lit` binary on PATH.
func TestInspectRenderTempDirLifecycle(t *testing.T) {
	newRenderCmd := func(outDir string) *cobra.Command {
		c := &cobra.Command{}
		c.Flags().String("render", "1", "")
		c.Flags().String("out", outDir, "")
		c.Flags().String("liteparse-bin", "", "")
		c.Flags().Var(&boundedIntFlagValue{min: 36, max: 600, value: "150"}, "dpi", "")
		c.Flags().Bool("ocr", false, "")
		c.Flags().Bool("no-cache", false, "")
		return c
	}
	countTempDirs := func(t *testing.T) int {
		t.Helper()
		entries, err := filepath.Glob(filepath.Join(os.TempDir(), "retab-inspect-*"))
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		return len(entries)
	}

	dir := t.TempDir()
	doc := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(doc, []byte("%PDF-1.4 fake"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Run("success keeps the reported directory", func(t *testing.T) {
		shotDir := t.TempDir()
		shot := filepath.Join(shotDir, "page-1.png")
		if err := os.WriteFile(shot, []byte("png"), 0o600); err != nil {
			t.Fatal(err)
		}
		withFakeLiteParser(t, &fakeLiteParser{
			parse: &ParseResult{TotalPages: 1, Pages: []ParsedPage{{Page: 1, Width: 10, Height: 10}}},
			shots: []ScreenshotPage{{Page: 1, Path: shot, MIMEType: "image/png"}},
		})

		before := countTempDirs(t)
		stdout, _ := captureStd(t, func() {
			if err := inspectRender(context.Background(), newRenderCmd(""), doc, kindPDF, "1"); err != nil {
				t.Fatalf("inspect --render: %v", err)
			}
		})

		var out struct {
			OutputDir string `json:"output_dir"`
		}
		if err := json.Unmarshal([]byte(stdout), &out); err != nil {
			t.Fatalf("decode output: %v\n%s", err, stdout)
		}
		if out.OutputDir == "" {
			t.Fatal("render reported no output_dir")
		}
		// The whole point: the directory named in the output must still exist.
		if _, err := os.Stat(out.OutputDir); err != nil {
			t.Fatalf("render reported %s but it does not exist: %v", out.OutputDir, err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(out.OutputDir) })
		if got := countTempDirs(t); got != before+1 {
			t.Errorf("temp dir count = %d, want %d (the reported dir must survive)", got, before+1)
		}
	})

	t.Run("failure after the dir is created leaks nothing", func(t *testing.T) {
		// Fails at the >3-page cap, which fires AFTER the temp dir is created —
		// one of the paths that used to orphan a directory per invocation.
		withFakeLiteParser(t, &fakeLiteParser{
			parse: &ParseResult{TotalPages: 10, Pages: []ParsedPage{{Page: 1}}},
		})
		before := countTempDirs(t)
		err := inspectRender(context.Background(), newRenderCmd(""), doc, kindPDF, "1-4")
		if err == nil {
			t.Fatal("expected the >3-page cap to reject this render")
		}
		if !strings.Contains(err.Error(), "at most 3 pages") {
			t.Fatalf("unexpected error (test no longer exercises the post-mkdir path): %v", err)
		}
		if got := countTempDirs(t); got != before {
			t.Errorf("failed render leaked a temp dir: before=%d after=%d", before, got)
		}
	})

	t.Run("user-supplied --out is never removed", func(t *testing.T) {
		withFakeLiteParser(t, &fakeLiteParser{
			parse: &ParseResult{TotalPages: 10, Pages: []ParsedPage{{Page: 1}}},
		})
		userDir := t.TempDir()
		_ = inspectRender(context.Background(), newRenderCmd(userDir), doc, kindPDF, "1-4")
		if _, err := os.Stat(userDir); err != nil {
			t.Errorf("a --out directory the user supplied must never be deleted: %v", err)
		}
	})
}

// The production safety gate skips entirely when the expected environment is
// not "production", and environmentFromKeyPrefix returns "" for any prefix
// outside its table — so an unrecognized key DISABLED the gate. `sk_live_`
// appears in this CLI's own `auth login` examples and `rtb_` in `setup`'s, so
// the highest-risk credentials were the ones silently exempted. An unplaceable
// key must fail SAFE to production, matching OAuth and stored legacy keys.
func TestExpectedEnvironmentForKeyFailsSafe(t *testing.T) {
	for key, want := range map[string]string{
		"rt_test_abc":       slugTest,
		"sk_retab_test_abc": slugTest,
		"sk_test_abc":       slugTest,
		"rt_live_abc":       slugProduction,
		"sk_live_abc":       slugProduction,
		"sk_retab_abc":      slugProduction,
		// Unplaceable prefixes: gated rather than silently exempt.
		"rtb_abc":         slugProduction,
		"totally_unknown": slugProduction,
		"":                slugProduction,
	} {
		if got := expectedEnvironmentForKey(key); got != want {
			t.Errorf("expectedEnvironmentForKey(%q) = %q, want %q", key, got, want)
		}
	}
}

// Nothing in the CLI ever wrote cfg.Environments, so resolveCredential's --env
// and --live branches were unreachable and their own error text pointed at
// commands that could not populate them. `auth login --env <slug>` / `--live`
// now writes the profile the selector reads.
func TestAuthLoginWritesEnvironmentProfile(t *testing.T) {
	cases := []struct {
		name     string
		live     bool
		envSlug  string
		wantSlug string
	}{
		{name: "live_writes_production", live: true, wantSlug: slugProduction},
		{name: "env_writes_named_profile", envSlug: "staging", wantSlug: "staging"},
		{name: "live_alias_normalizes", envSlug: "live", wantSlug: slugProduction},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateHome(t)
			root := newTestRootCmd()
			if tc.live {
				if err := root.PersistentFlags().Set("live", "true"); err != nil {
					t.Fatal(err)
				}
			}
			if tc.envSlug != "" {
				if err := root.PersistentFlags().Set("env", tc.envSlug); err != nil {
					t.Fatal(err)
				}
			}
			child := &cobra.Command{Use: "child"}
			root.AddCommand(child)

			slug, err := loginProfileSlug(child)
			if err != nil {
				t.Fatalf("loginProfileSlug: %v", err)
			}
			if slug != tc.wantSlug {
				t.Fatalf("slug = %q, want %q", slug, tc.wantSlug)
			}

			captureStd(t, func() {
				if err := runAPIKeyLoginForProfile("rt_live_written", "", slug); err != nil {
					t.Fatalf("login: %v", err)
				}
			})

			cfg, err := loadConfig()
			if err != nil {
				t.Fatalf("loadConfig: %v", err)
			}
			profile := cfg.Environments[tc.wantSlug]
			if profile == nil || profile.APIKey != "rt_live_written" {
				t.Fatalf("profile %q was not written: %+v", tc.wantSlug, cfg.Environments)
			}
			if profile.APIKeyPreview == "" {
				t.Error("profile should record a redacted preview")
			}
			// A profile login must not fall back to the legacy single-key slot.
			if cfg.APIKey != "" {
				t.Errorf("profile login wrote the legacy api_key slot: %q", cfg.APIKey)
			}

			// The selector must now resolve the profile it just wrote — that
			// round trip is the whole bug: the remediation the error text
			// suggested used to leave the selector failing identically.
			resolveRoot := newTestRootCmd()
			if tc.wantSlug == slugProduction {
				_ = resolveRoot.PersistentFlags().Set("live", "true")
			} else {
				_ = resolveRoot.PersistentFlags().Set("env", tc.wantSlug)
			}
			cred, err := resolveCredential(resolveRoot)
			if err != nil {
				t.Fatalf("selector cannot resolve the profile it just wrote: %v", err)
			}
			if cred.APIKey != "rt_live_written" {
				t.Errorf("resolved key = %q, want the profile key", cred.APIKey)
			}
		})
	}
}

// --live and --env select different profiles; accepting both would pick one
// arbitrarily.
func TestLoginProfileSlugRejectsLivePlusEnv(t *testing.T) {
	root := newTestRootCmd()
	_ = root.PersistentFlags().Set("live", "true")
	_ = root.PersistentFlags().Set("env", "staging")
	child := &cobra.Command{Use: "child"}
	root.AddCommand(child)
	if _, err := loginProfileSlug(child); err == nil {
		t.Fatal("--live combined with --env must be rejected")
	}
}

// Files that can carry the API key must be owner-only, and the mode has to be
// enforced on files that ALREADY exist — os.WriteFile only applies perm when it
// creates the file, and these writers upsert into an existing .mcp.json.
func TestMCPConfigFilesAreOwnerOnly(t *testing.T) {
	if mcpConfigFileMode != 0o600 {
		t.Fatalf("mcpConfigFileMode = %#o, want 0600", mcpConfigFileMode)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// Pre-existing and world-readable: the case os.WriteFile would not fix.
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeFileCreatingParents(path, []byte(`{"mcpServers":{}}`), mcpConfigFileMode); err != nil {
		t.Fatalf("write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		// Go reports a synthetic mode on Windows; secureConfigFile applies a
		// restrictive DACL there, covered by secure_file_windows_test.go.
		t.Skipf("permission bits are not meaningful on windows (mode=%v)", info.Mode().Perm())
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("existing MCP config kept mode %#o, want 0600", perm)
	}
}

// `local_path` is a purely local fixture binding the server never returns, so
// rewriting mounts.json from the server config erased it. push refreshed the
// bundle through that writer and never re-hydrated, silently unbinding every
// table fixture; pull --force clobbered it the same way.
func TestMountsWriterPreservesLocalPaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mounts.json")
	existing := map[string]any{
		"tables": []any{
			map[string]any{"table_id": "tbl_1", "path": "/sandbox/a.csv", "local_path": "./fixtures/a.csv"},
			map[string]any{"table_id": "tbl_2", "path": "/sandbox/b.csv"},
		},
	}
	if err := writeJSONFile(path, existing); err != nil {
		t.Fatal(err)
	}
	// What the server returns: the same tables, with no local_path anywhere.
	fromServer := map[string]any{
		"tables": []any{
			map[string]any{"table_id": "tbl_1", "path": "/sandbox/a.csv"},
			map[string]any{"table_id": "tbl_2", "path": "/sandbox/b.csv"},
		},
	}
	if err := writeMountsFilePreservingLocalPaths(path, fromServer); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := readJSONMap(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	tables, _ := got["tables"].([]any)
	if len(tables) != 2 {
		t.Fatalf("tables = %v", tables)
	}
	first, _ := tables[0].(map[string]any)
	if first["local_path"] != "./fixtures/a.csv" {
		t.Errorf("local_path was dropped on refresh: %v", first)
	}
	second, _ := tables[1].(map[string]any)
	if _, ok := second["local_path"]; ok {
		t.Errorf("a table with no binding must not gain one: %v", second)
	}
	// A non-map payload must still round-trip through the writer.
	if err := writeMountsFilePreservingLocalPaths(filepath.Join(dir, "other.json"), []any{"x"}); err != nil {
		t.Fatalf("non-map payload: %v", err)
	}
}

// total_matches must be the DOCUMENT total, not the length of the truncated
// slice — it was by construction identical to len(matches), so the CLI printed
// "2 of 2" for a document holding 5.
func TestGrepReportsTrueTotalMatches(t *testing.T) {
	res := &ParseResult{
		TotalPages: 1,
		Pages:      []ParsedPage{{Page: 1, Text: "a\na\na\na\na\n"}},
	}
	matcher, err := buildMatcher("a", false, false)
	if err != nil {
		t.Fatal(err)
	}
	matches, total, truncated := grepParseResult(res, kindText, matcher, 0, 2, false)
	if total != 5 {
		t.Errorf("total_matches = %d, want 5 (the document total)", total)
	}
	if len(matches) != 2 {
		t.Errorf("returned %d matches, want the --max-results cap of 2", len(matches))
	}
	if !truncated {
		t.Error("truncated must be true when the cap bit")
	}
	matches, total, truncated = grepParseResult(res, kindText, matcher, 0, 50, false)
	if total != 5 || len(matches) != 5 || truncated {
		t.Errorf("untruncated: total=%d returned=%d truncated=%v", total, len(matches), truncated)
	}
}
