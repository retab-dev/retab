package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseSemver(t *testing.T) {
	cases := []struct {
		in   string
		want semver
		ok   bool
	}{
		{"1.2.3", semver{1, 2, 3}, true},
		{"v1.2.3", semver{1, 2, 3}, true},
		{"0.1.0", semver{0, 1, 0}, true},
		{"1.2", semver{1, 2, 0}, true},
		{"2", semver{2, 0, 0}, true},
		{"1.2.3-rc.1", semver{1, 2, 3}, true},
		{"1.2.3+build.5", semver{1, 2, 3}, true},
		{"dev", semver{}, false},
		{"", semver{}, false},
		{"1.2.3.4", semver{}, false},
		{"1.x.0", semver{}, false},
		{"-1.0.0", semver{}, false},
	}
	for _, c := range cases {
		got, ok := parseSemver(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("parseSemver(%q) = %v,%v want %v,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"0.1.0", "0.2.0", true},
		{"0.1.0", "0.1.1", true},
		{"1.0.0", "2.0.0", true},
		{"0.2.0", "0.2.0", false},
		{"0.3.0", "0.2.0", false},
		{"2.0.0", "1.9.9", false},
		{"dev", "1.0.0", false}, // unreleased current never nags
		{"1.0.0", "garbage", false},
	}
	for _, c := range cases {
		if got := isNewerVersion(c.current, c.latest); got != c.want {
			t.Errorf("isNewerVersion(%q,%q) = %v want %v", c.current, c.latest, got, c.want)
		}
	}
}

func TestParseLatestCLIVersion(t *testing.T) {
	body := []byte(`[
		{"tag_name": "clients/go/v9.9.9", "draft": false, "prerelease": false},
		{"tag_name": "cli-v0.1.0", "draft": false, "prerelease": false},
		{"tag_name": "cli-v0.3.0", "draft": false, "prerelease": false},
		{"tag_name": "cli-v0.2.0", "draft": false, "prerelease": false},
		{"tag_name": "cli-v0.4.0", "draft": true, "prerelease": false},
		{"tag_name": "cli-v0.5.0-rc.1", "draft": false, "prerelease": true}
	]`)
	got, err := parseLatestCLIVersion(body, "cli-")
	if err != nil {
		t.Fatalf("parseLatestCLIVersion: %v", err)
	}
	// Highest stable cli- release, ignoring the SDK tag, the draft, and the
	// prerelease.
	if got != "0.3.0" {
		t.Fatalf("got %q want 0.3.0", got)
	}
}

func TestParseLatestCLIVersionNoMatch(t *testing.T) {
	body := []byte(`[{"tag_name": "clients/go/v1.0.0", "draft": false, "prerelease": false}]`)
	if _, err := parseLatestCLIVersion(body, "cli-"); err == nil {
		t.Fatal("expected error when no cli- release present")
	}
}

func TestUpdateCacheRoundTrip(t *testing.T) {
	withTempHome(t)

	want := updateCache{
		LastCheckedAt: time.Now().UTC().Truncate(time.Second),
		LatestVersion: "0.9.0",
	}
	if err := saveUpdateCache(want); err != nil {
		t.Fatalf("saveUpdateCache: %v", err)
	}
	got, err := loadUpdateCache()
	if err != nil {
		t.Fatalf("loadUpdateCache: %v", err)
	}
	if !got.LastCheckedAt.Equal(want.LastCheckedAt) || got.LatestVersion != want.LatestVersion {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestLoadUpdateCacheMissing(t *testing.T) {
	withTempHome(t)
	got, err := loadUpdateCache()
	if err != nil {
		t.Fatalf("loadUpdateCache on missing file: %v", err)
	}
	if got.LatestVersion != "" || !got.LastCheckedAt.IsZero() {
		t.Fatalf("expected zero cache, got %+v", got)
	}
}

func TestLoadUpdateCacheCorrupt(t *testing.T) {
	withTempHome(t)
	path, err := updateCachePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := loadUpdateCache()
	if err != nil {
		t.Fatalf("corrupt cache should not error: %v", err)
	}
	if got.LatestVersion != "" {
		t.Fatalf("corrupt cache should read as empty, got %+v", got)
	}
}

func TestUpdateCacheIsStale(t *testing.T) {
	now := time.Now()
	if !(updateCache{}).isStale(now, updateCheckInterval) {
		t.Error("zero cache should be stale")
	}
	fresh := updateCache{LastCheckedAt: now.Add(-time.Hour)}
	if fresh.isStale(now, updateCheckInterval) {
		t.Error("1h-old cache should be fresh under 24h interval")
	}
	old := updateCache{LastCheckedAt: now.Add(-48 * time.Hour)}
	if !old.isStale(now, updateCheckInterval) {
		t.Error("48h-old cache should be stale")
	}
}

func TestNotifierEnabled(t *testing.T) {
	noEnv := func(string) string { return "" }

	if !notifierEnabled([]string{"files", "list"}, noEnv, "0.1.0", true) {
		t.Error("baseline: released version on a TTY with a normal command should be enabled")
	}
	if notifierEnabled([]string{"files"}, noEnv, "dev", true) {
		t.Error("dev build should disable the notifier")
	}
	if notifierEnabled([]string{"files"}, noEnv, "0.1.0", false) {
		t.Error("non-TTY stderr should disable the notifier")
	}
	ciEnv := func(k string) string {
		if k == "CI" {
			return "true"
		}
		return ""
	}
	if notifierEnabled([]string{"files"}, ciEnv, "0.1.0", true) {
		t.Error("CI should disable the notifier")
	}
	optOut := func(k string) string {
		if k == "NO_UPDATE_NOTIFIER" {
			return "1"
		}
		return ""
	}
	if notifierEnabled([]string{"files"}, optOut, "0.1.0", true) {
		t.Error("NO_UPDATE_NOTIFIER should disable the notifier")
	}
	retabOptOut := func(k string) string {
		if k == "RETAB_NO_UPDATE_NOTIFIER" {
			return "1"
		}
		return ""
	}
	if notifierEnabled([]string{"files"}, retabOptOut, "0.1.0", true) {
		t.Error("RETAB_NO_UPDATE_NOTIFIER should disable the notifier")
	}
	for _, args := range [][]string{
		{updateDaemonCommand},
		{"update"},
		{"version"},
		{"completion", "zsh"},
		{"--version"},
		{"-v"},
	} {
		if notifierEnabled(args, noEnv, "0.1.0", true) {
			t.Errorf("args %v should disable the notifier", args)
		}
	}
}

func TestPrintUpdateNotice(t *testing.T) {
	var b strings.Builder
	printUpdateNotice(&b, "0.1.0", "0.2.0")
	out := b.String()
	for _, want := range []string{"Update available", "0.1.0", "0.2.0", updateInstallCommand} {
		if !strings.Contains(out, want) {
			t.Errorf("notice missing %q:\n%s", want, out)
		}
	}
	// A non-*os.File writer gets the empty palette, so the notice must be
	// plain text with no ANSI escapes.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI escapes for non-TTY writer:\n%q", out)
	}
}

func TestUpdateCmdTestDevRendersNotice(t *testing.T) {
	if err := updateCmd.Flags().Set("test-dev", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = updateCmd.Flags().Set("test-dev", "false") })

	var out, errBuf strings.Builder
	updateCmd.SetOut(&out)
	updateCmd.SetErr(&errBuf)
	t.Cleanup(func() {
		updateCmd.SetOut(nil)
		updateCmd.SetErr(nil)
	})

	if err := updateCmd.RunE(updateCmd, nil); err != nil {
		t.Fatalf("update --test-dev: %v", err)
	}

	notice := errBuf.String()
	// The preview always renders a notice (a simulated newer version),
	// even though the real notifier is inert on a dev build.
	if !strings.Contains(notice, "Update available") {
		t.Fatalf("expected a simulated notice on stderr, got:\n%s", notice)
	}
	_, latest := simulatedNotice()
	if !strings.Contains(notice, latest) {
		t.Fatalf("notice should mention simulated latest %q:\n%s", latest, notice)
	}
	if out.String() != "" {
		t.Fatalf("preview should not write to stdout, got: %q", out.String())
	}
}

func TestSimulatedNotice(t *testing.T) {
	current, latest := simulatedNotice()
	cur, okCur := parseSemver(current)
	lat, okLat := parseSemver(latest)
	if !okCur || !okLat {
		t.Fatalf("simulatedNotice produced unparseable versions: %q -> %q", current, latest)
	}
	if compareSemver(lat, cur) <= 0 {
		t.Fatalf("simulated latest %q should be newer than current %q", latest, current)
	}
}

// withTempHome points os.UserHomeDir at a throwaway dir so cache reads and
// writes never touch the developer's real ~/.retab.
func withTempHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	// os.UserHomeDir reads $HOME on unix and %USERPROFILE% on Windows.
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}
