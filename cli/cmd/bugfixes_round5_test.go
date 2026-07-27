//go:build !retab_oagen_cli_files

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

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

			slug, err := loginEnvironmentSlug(child)
			if err != nil {
				t.Fatalf("loginEnvironmentSlug: %v", err)
			}
			if slug != tc.wantSlug {
				t.Fatalf("slug = %q, want %q", slug, tc.wantSlug)
			}

			captureStd(t, func() {
				if err := runAPIKeyLogin("rt_live_written", "", slug); err != nil {
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
	if _, err := loginEnvironmentSlug(child); err == nil {
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

// --live is an alias for --env production, so the two only conflict when they
// name different environments. resolveCredential already treats them as
// equivalent; login rejecting the agreeing pair made it stricter than the
// selector it feeds.
func TestLoginProfileSlugLiveAndEnvAgreeing(t *testing.T) {
	newRoot := func(live bool, env string) *cobra.Command {
		root := newTestRootCmd()
		if live {
			_ = root.PersistentFlags().Set("live", "true")
		}
		if env != "" {
			_ = root.PersistentFlags().Set("env", env)
		}
		child := &cobra.Command{Use: "child"}
		root.AddCommand(child)
		return child
	}
	// Agreeing: accepted.
	for _, env := range []string{"production", "live", "PRODUCTION"} {
		slug, err := loginEnvironmentSlug(newRoot(true, env))
		if err != nil {
			t.Errorf("--live --env %s should be accepted (they agree): %v", env, err)
		} else if slug != slugProduction {
			t.Errorf("--live --env %s = %q, want production", env, slug)
		}
	}
	// Disagreeing: rejected, and the message says why.
	_, err := loginEnvironmentSlug(newRoot(true, "staging"))
	if err == nil {
		t.Fatal("--live --env staging must be rejected")
	}
	if !strings.Contains(err.Error(), "production") {
		t.Errorf("error should explain that --live means production, got: %v", err)
	}
}

// withRootSelector sets the root's persistent --env/--live for the duration of a
// test and restores them, so a RunE that reads them via cmd.Root() can be driven
// without reparenting the global authLoginCmd (which corrupts the shared command
// tree for later tests).
func withRootSelector(t *testing.T, env string, live bool) {
	t.Helper()
	envF := rootCmd.PersistentFlags().Lookup("env")
	liveF := rootCmd.PersistentFlags().Lookup("live")
	if env != "" {
		_ = envF.Value.Set(env)
		envF.Changed = true
	}
	if live {
		_ = liveF.Value.Set("true")
		liveF.Changed = true
	}
	t.Cleanup(func() {
		_ = envF.Value.Set("")
		envF.Changed = false
		_ = liveF.Value.Set("false")
		liveF.Changed = false
	})
}

// resetAuthLoginFlags clears authLoginCmd's local flags before and after a test.
// authLoginCmd is a process-global command; merely Set()ing a flag marks it
// Changed, which trips earlier guards in the RunE.
func resetAuthLoginFlags(t *testing.T) {
	t.Helper()
	reset := func() {
		for _, name := range []string{"api-key", "access-token", "base-url"} {
			f := authLoginCmd.Flags().Lookup(name)
			_ = f.Value.Set("")
			f.Changed = false
		}
		b := authLoginCmd.Flags().Lookup("browser")
		_ = b.Value.Set("true")
		b.Changed = false
	}
	reset()
	t.Cleanup(reset)
}

// Environment profiles hold API keys only. Resolving the slug after the
// access-token branch meant `auth login --access-token X --env staging`
// silently ignored --env and stored the token in the default slot, while the
// user believed they had scoped it. This drives the real RunE guard, not just
// the slug helper — asserting the helper alone passed even with the guard
// removed.
func TestAuthLoginRejectsAccessTokenWithProfileSelector(t *testing.T) {
	isolateHome(t)
	withRootSelector(t, "staging", false)
	resetAuthLoginFlags(t)
	if err := authLoginCmd.Flags().Set("access-token", "acctk_secret"); err != nil {
		t.Fatal(err)
	}

	var err error
	captureStd(t, func() { err = authLoginCmd.RunE(authLoginCmd, nil) })
	if err == nil {
		t.Fatal("--access-token with --env must be rejected, not silently ignored")
	}
	if !strings.Contains(err.Error(), "cannot be combined with --access-token") {
		t.Errorf("expected the access-token/selector guard, got: %v", err)
	}
	// A refused login must not persist anything.
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		t.Fatal(cfgErr)
	}
	if cfg.AccessToken != "" || len(cfg.Environments) != 0 {
		t.Errorf("refused login wrote state: accessToken=%q environments=%v", cfg.AccessToken, cfg.Environments)
	}
}

// writeFileCreatingParents must only ever TIGHTEN permissions. Chmod-ing the
// non-secret files too would force a registry the user had deliberately
// chmod'd 0600 back open to 0644 on every setup/sync.
func TestWriteFileCreatingParentsNeverWidensNonSecretFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "install-registry.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	// The registry is written 0644 and carries no secret; a user who tightened
	// it must keep their choice.
	if err := writeFileCreatingParents(path, []byte(`{"version":1}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("non-secret write widened the file to %#o; it must not loosen permissions", perm)
	}
}

// A project-local install writes the key into files that are easy to commit.
// The file mode does not survive a commit or a fresh checkout, so the user has
// to be told.
func TestSetupWarnsWhenLocalInstallWritesCommittableKey(t *testing.T) {
	results := []setupResult{{Agent: "claude-code", MCPPath: "/repo/.mcp.json"}}

	var warned bytes.Buffer
	c := &cobra.Command{}
	c.SetErr(&warned)
	warnLocalMCPKeyIsCommittable(c, installScopeLocal, "rt_live_secret", results)
	out := warned.String()
	if !strings.Contains(out, ".mcp.json") {
		t.Errorf("warning should name the files, got: %q", out)
	}
	if !strings.Contains(out, ".gitignore") {
		t.Errorf("warning should tell the user what to do, got: %q", out)
	}

	// No warning without a key, or for a global install: nothing committable.
	for _, tc := range []struct {
		name   string
		scope  installScope
		apiKey string
	}{
		{"no key", installScopeLocal, ""},
		{"global scope", installScopeGlobal, "rt_live_secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			quiet := &cobra.Command{}
			quiet.SetErr(&buf)
			warnLocalMCPKeyIsCommittable(quiet, tc.scope, tc.apiKey, results)
			if buf.Len() != 0 {
				t.Errorf("unexpected warning: %q", buf.String())
			}
		})
	}
}

// grep must stay LINEAR in document size. Counting every match (so
// total_matches is a real total) removed the early exit, which exposed two
// quadratic costs: the line number was recomputed by scanning from offset 0 for
// every match, and every match — including the ones immediately discarded past
// --max-results — built a snippet and, with --bbox, a bounding box unioned from
// the page's text items. A 1.4 MB text document took ~30s and a dense PDF with
// --bbox ~46s; 200k matches never finished.
//
// This asserts the shape, not a wall-clock budget: doubling the document must
// roughly double the time, not quadruple it. The threshold is deliberately
// loose (4x for a 2x input) so it fails on a return to O(n^2) without flaking
// on a slow or noisy machine.
func TestGrepScalesLinearly(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-sensitive")
	}
	build := func(lines int) *ParseResult {
		var b strings.Builder
		for i := 0; i < lines; i++ {
			b.WriteString("lorem ipsum dolor sit amet e consectetur adipiscing\n")
		}
		return &ParseResult{TotalPages: 1, Pages: []ParsedPage{{Page: 1, Text: b.String()}}}
	}
	matcher, err := buildMatcher("e", false, false)
	if err != nil {
		t.Fatal(err)
	}
	measure := func(lines int) time.Duration {
		res := build(lines)
		start := time.Now()
		_, total, truncated := grepParseResult(res, kindText, matcher, 0, 50, false)
		elapsed := time.Since(start)
		if total <= 50 || !truncated {
			t.Fatalf("fixture should overflow the cap: total=%d truncated=%v", total, truncated)
		}
		return elapsed
	}
	// Warm up so the first allocation burst isn't attributed to the small run.
	measure(2000)
	small := measure(20000)
	large := measure(40000)
	if small <= 0 {
		small = time.Microsecond
	}
	if ratio := float64(large) / float64(small); ratio > 4 {
		t.Errorf("doubling the document multiplied the time by %.1fx (small=%v large=%v); "+
			"grep looks quadratic again", ratio, small, large)
	}
}

// Past --max-results, a match must cost a counter increment and nothing else.
// Bounding boxes are the dominant per-match cost, so a dense page with --bbox is
// the case that regressed worst.
//
// Asserts on the number of bounding boxes actually BUILT (grepBoundingBoxCalls),
// not on wall-clock time: a timing ceiling is inherently flaky and, with a
// fast-matching fixture, ~300x too loose to catch the regression it names. The
// count is exact — with the skip removed it equals the document total.
func TestGrepSkipsPerMatchWorkPastTheCap(t *testing.T) {
	items := make([]ParsedItem, 0, 2000)
	for i := 0; i < 2000; i++ {
		items = append(items, ParsedItem{Text: "e", X: float64(i % 100), Y: float64(i / 100), Width: 1, Height: 1})
	}
	var b strings.Builder
	for i := 0; i < 20000; i++ {
		b.WriteString("value e here\n")
	}
	page := ParsedPage{Page: 1, Width: 100, Height: 100, Text: b.String(), Items: items}
	result := &ParseResult{TotalPages: 1, Pages: []ParsedPage{page}}
	matcher, err := buildMatcher("e", false, false)
	if err != nil {
		t.Fatal(err)
	}

	grepBoundingBoxCalls = 0
	matches, total, truncated := grepParseResult(result, kindPDF, matcher, 0, 50, true)

	if len(matches) != 50 {
		t.Errorf("returned %d matches, want the cap of 50", len(matches))
	}
	if total <= 50 || !truncated {
		t.Errorf("total=%d truncated=%v, want a truncated count well above the cap", total, truncated)
	}
	// The load-bearing assertion: a bounding box is built only for the retained
	// matches. With the skip-past-cap removed this equals `total` (20000).
	if grepBoundingBoxCalls != len(matches) {
		t.Errorf("built %d bounding boxes for %d retained matches (of %d total); "+
			"per-match work is not being skipped past the cap", grepBoundingBoxCalls, len(matches), total)
	}
	// Only the retained matches carry a bounding box; the rest were never built.
	for i, m := range matches {
		if m.Anchor.Kind != anchorPDFPage {
			t.Errorf("match %d has anchor kind %q", i, m.Anchor.Kind)
		}
	}
}

// Writing an environment profile must NOT promote it to the default
// credential. Branch 5 of resolveCredential outranks both the stored access
// token and the OAuth session, so setting DefaultEnvironment here silently
// retired an existing OAuth login for every plain `retab ...` — and since
// nothing in the CLI ever clears DefaultEnvironment, a later
// `auth login --api-key ...` appeared to succeed while the stale profile kept
// winning. The only escape was `auth logout` or hand-editing JSON.
func TestProfileLoginDoesNotHijackTheDefaultCredential(t *testing.T) {
	isolateHome(t)
	// An existing OAuth session, as a browser login would leave it.
	if err := saveConfig(retabConfig{
		OAuth: &oauthTokens{AccessToken: "oauth_access", RefreshToken: "oauth_refresh", ExpiresAt: time.Now().Add(time.Hour)},
	}); err != nil {
		t.Fatal(err)
	}

	captureStd(t, func() {
		if err := runAPIKeyLogin("rt_test_scoped", "", "staging"); err != nil {
			t.Fatalf("profile login: %v", err)
		}
	})

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultEnvironment != "" {
		t.Errorf("profile login set default_environment=%q; it must stay an explicit selector", cfg.DefaultEnvironment)
	}
	if cfg.OAuth == nil || cfg.OAuth.AccessToken != "oauth_access" {
		t.Fatalf("the OAuth session must survive a profile login: %+v", cfg.OAuth)
	}

	// A plain invocation still resolves the OAuth session, not the profile.
	cred, err := resolveCredential(newTestRootCmd())
	if err != nil {
		t.Fatalf("resolveCredential: %v", err)
	}
	if cred.Source != sourceOAuth {
		t.Errorf("plain invocation resolved %q; the profile hijacked the default credential", cred.Source)
	}
	if cred.APIKey == "rt_test_scoped" {
		t.Error("plain invocation used the profile key")
	}

	// ...and the profile is still reachable when explicitly selected.
	selector := newTestRootCmd()
	_ = selector.PersistentFlags().Set("env", "staging")
	if cred, err = resolveCredential(selector); err != nil {
		t.Fatalf("--env staging: %v", err)
	}
	if cred.APIKey != "rt_test_scoped" {
		t.Errorf("--env staging resolved %q, want the profile key", cred.APIKey)
	}
}

// configuredLoginBaseURL falls back to the public default, so writing it
// unconditionally repointed a profile pinned at a local or self-hosted
// deployment back at api.retab.com on every key rotation.
func TestProfileLoginPreservesPinnedBaseURL(t *testing.T) {
	isolateHome(t)
	cfg := retabConfig{Environments: map[string]*environmentProfile{
		"staging": {Name: "staging", APIKey: "rt_test_old", BaseURL: "http://localhost:4000"},
	}}
	if err := saveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	// Rotate the key with no --base-url: the pin must survive.
	captureStd(t, func() {
		if err := runAPIKeyLogin("rt_test_new", "", "staging"); err != nil {
			t.Fatalf("rotate: %v", err)
		}
	})
	reloaded, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	profile := reloaded.Environments["staging"]
	if profile == nil {
		t.Fatal("profile disappeared")
	}
	if profile.APIKey != "rt_test_new" {
		t.Errorf("key = %q, want the rotated key", profile.APIKey)
	}
	if profile.BaseURL != "http://localhost:4000" {
		t.Errorf("base_url = %q; a key rotation must not repoint a pinned deployment", profile.BaseURL)
	}

	// An explicit --base-url still updates it.
	captureStd(t, func() {
		if err := runAPIKeyLogin("rt_test_new", "https://eu.retab.com", "staging"); err != nil {
			t.Fatalf("re-pin: %v", err)
		}
	})
	if reloaded, err = loadConfig(); err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Environments["staging"].BaseURL; got != "https://eu.retab.com" {
		t.Errorf("explicit --base-url = %q, want it applied", got)
	}
}

// The mounts writer must not mutate the caller's server config. It receives
// block.Config["mounts"] directly, and preservation writes local_path into the
// table maps in place — which leaked a local-only field back into the in-memory
// server config and poisoned push's reported config_hash.
func TestMountsWriterDoesNotMutateServerConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mounts.json")
	if err := writeJSONFile(path, map[string]any{
		"tables": []any{map[string]any{"table_id": "tbl_1", "path": "/sandbox/a.csv", "local_path": "./fixtures/a.csv"}},
	}); err != nil {
		t.Fatal(err)
	}

	serverConfig := map[string]any{
		"mounts": map[string]any{
			"tables": []any{map[string]any{"table_id": "tbl_1", "path": "/sandbox/a.csv"}},
		},
	}
	hashBefore := hashJSONMap(serverConfig)

	if err := writeMountsFilePreservingLocalPaths(path, serverConfig["mounts"]); err != nil {
		t.Fatalf("write: %v", err)
	}

	if got := hashJSONMap(serverConfig); got != hashBefore {
		t.Errorf("writer mutated the caller's config (hash %s -> %s)", hashBefore, got)
	}
	mounts, _ := serverConfig["mounts"].(map[string]any)
	tables, _ := mounts["tables"].([]any)
	table, _ := tables[0].(map[string]any)
	if _, leaked := table["local_path"]; leaked {
		t.Errorf("local_path leaked into the server config: %v", table)
	}
	// The file still gets the preserved binding.
	onDisk, err := readJSONMap(path)
	if err != nil {
		t.Fatal(err)
	}
	diskTables, _ := onDisk["tables"].([]any)
	diskTable, _ := diskTables[0].(map[string]any)
	if diskTable["local_path"] != "./fixtures/a.csv" {
		t.Errorf("local_path was not preserved on disk: %v", diskTable)
	}
}

// `auth login --env x` / `--live` with the browser flow used to fall straight
// through, minting an OAuth session and writing no profile — so the very next
// `retab --live ...` reported "no live credential configured. Run `retab auth
// login --live --api-key ...`", pointing back at a command the user had just
// run. That is the same infinite loop the profile fix set out to close, one
// flag over, so it must error rather than silently drop the selector.
func TestAuthLoginRejectsProfileSelectorWithoutAPIKey(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  string
		live bool
	}{
		{name: "env", env: "staging"},
		{name: "live", live: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			isolateHome(t)
			// Drive the real RunE without reparenting authLoginCmd: set the
			// selector on the root's persistent flags (which authLoginCmd.Root()
			// reads) and its own local flags, restoring both afterward.
			withRootSelector(t, tc.env, tc.live)
			resetAuthLoginFlags(t)

			var err error
			captureStd(t, func() { err = authLoginCmd.RunE(authLoginCmd, nil) })
			if err == nil {
				t.Fatal("a profile selector without --api-key must be refused, not silently ignored")
			}
			// Assert the GUARD's message specifically. The browser flow this
			// would otherwise fall through to also fails here (no network) with
			// an error that happens to mention --api-key, so a looser assertion
			// passes even when the guard is removed.
			if !strings.Contains(err.Error(), "store an API key in an environment profile") {
				t.Errorf("expected the profile-selector guard to reject this, got: %v", err)
			}

			// Nothing may have been written.
			cfg, cfgErr := loadConfig()
			if cfgErr != nil {
				t.Fatal(cfgErr)
			}
			if len(cfg.Environments) != 0 {
				t.Errorf("refused login still wrote a profile: %+v", cfg.Environments)
			}
			if cfg.OAuth != nil {
				t.Error("refused login started an OAuth session")
			}
		})
	}
}

// The real behaviour change — the MCP-config writers passing mcpConfigFileMode
// (0600) instead of 0644 — is what actually secures the API key on disk, and it
// went unguarded: reverting both call sites to 0644 left the suite green. These
// drive the writers themselves (through a pre-existing world-readable file, the
// case os.WriteFile would not re-perm) and assert owner-only.
func TestUpsertMCPConfigWritesOwnerOnlyFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Go reports a synthetic mode on Windows; secureConfigFile applies a
		// restrictive DACL there, covered by secure_file_windows_test.go.
		t.Skip("permission bits are not meaningful on windows")
	}
	dir := t.TempDir()

	jsonPath := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	serverConfig := mcpServerConfig{
		Type:    "http",
		URL:     "https://api.retab.com/mcp",
		Headers: map[string]string{"Authorization": "Bearer rt_live_secret"},
	}
	if err := upsertJSONMCPConfig(jsonPath, "mcpServers", "retab", serverConfig); err != nil {
		t.Fatalf("upsertJSONMCPConfig: %v", err)
	}
	assertOwnerOnly(t, jsonPath)

	tomlPath := filepath.Join(dir, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(tomlPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tomlPath, []byte("existing = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := upsertTomlMCPConfig(tomlPath, "retab", serverConfig); err != nil {
		t.Fatalf("upsertTomlMCPConfig: %v", err)
	}
	assertOwnerOnly(t, tomlPath)

	// The key really is in the files — so the mode assertion isn't guarding
	// empty content.
	for _, p := range []string{jsonPath, tomlPath} {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), "rt_live_secret") {
			t.Errorf("%s does not contain the key; the mode assertion would be vacuous", p)
		}
	}
}

func assertOwnerOnly(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != mcpConfigFileMode {
		t.Errorf("%s has mode %#o, want %#o (owner-only)", path, perm, mcpConfigFileMode)
	}
}
