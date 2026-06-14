package cmd

import (
	"encoding/json"
	"runtime/debug"
	"strings"
	"testing"
)

func TestVersionHonorsOutputJSON(t *testing.T) {
	if err := rootCmd.PersistentFlags().Set("output", "json"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := versionCmd.RunE(versionCmd, nil); err != nil {
			t.Fatal(err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("expected JSON version output, got %v for:\n%s", err, stdout)
	}
	if got["version"] != version {
		t.Fatalf("version = %q, want %q", got["version"], version)
	}
	if got["commit"] != commit {
		t.Fatalf("commit = %q, want %q", got["commit"], commit)
	}
	if got["built"] != date {
		t.Fatalf("built = %q, want %q", got["built"], date)
	}
}

func TestVersionHonorsOutputTable(t *testing.T) {
	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := versionCmd.RunE(versionCmd, nil); err != nil {
			t.Fatal(err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	for _, want := range []string{"VERSION", "COMMIT", "BUILT", version, commit, date} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}

// withStampedDefaults pins the link-time vars to their unstamped defaults
// ("dev"/"none"/"unknown") for the duration of a test, so resolveVersionInfo's
// fallback branches are exercised regardless of how the test binary was built.
func withStampedDefaults(t *testing.T, v, c, d string) {
	t.Helper()
	origV, origC, origD := version, commit, date
	version, commit, date = v, c, d
	t.Cleanup(func() { version, commit, date = origV, origC, origD })
}

func TestResolveVersionInfoUsesLdflagStampsWhenPresent(t *testing.T) {
	// When the linker stamped real values, ReadBuildInfo must never override
	// them — the GoReleaser identity is authoritative.
	withStampedDefaults(t, "0.4.2", "a1b2c3d", "2026-05-14T15:03:21Z")

	orig := buildInfoSource
	t.Cleanup(func() { buildInfoSource = orig })
	buildInfoSource = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v9.9.9"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "deadbeef"},
				{Key: "vcs.time", Value: "2000-01-01T00:00:00Z"},
			},
		}, true
	}

	got := resolveVersionInfo()
	if got.Version != "0.4.2" || got.Commit != "a1b2c3d" || got.Built != "2026-05-14T15:03:21Z" {
		t.Fatalf("ldflag stamps must win, got %+v", got)
	}
}

func TestResolveVersionInfoFallsBackToBuildInfo(t *testing.T) {
	// A plain `go install`/`go build` binary: vars carry the unstamped
	// defaults, so the module version + VCS stamps must fill them in.
	withStampedDefaults(t, "dev", "none", "unknown")

	orig := buildInfoSource
	t.Cleanup(func() { buildInfoSource = orig })
	buildInfoSource = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v0.1.0"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "cafef00d"},
				{Key: "vcs.time", Value: "2026-06-14T10:00:00Z"},
			},
		}, true
	}

	got := resolveVersionInfo()
	if got.Version == "dev" || got.Version == "" {
		t.Fatalf("version should not stay %q when build info is available", got.Version)
	}
	if got.Commit == "none" || got.Commit == "" {
		t.Fatalf("commit should not stay %q when build info is available", got.Commit)
	}
	if got.Built == "unknown" || got.Built == "" {
		t.Fatalf("built should not stay %q when build info is available", got.Built)
	}
	if got.Version != "v0.1.0" || got.Commit != "cafef00d" || got.Built != "2026-06-14T10:00:00Z" {
		t.Fatalf("unexpected fallback values: %+v", got)
	}
}

func TestResolveVersionInfoKeepsDefaultsWithoutBuildInfo(t *testing.T) {
	// `(devel)` and a missing build info table both leave the honest
	// "dev"/"none"/"unknown" defaults in place.
	withStampedDefaults(t, "dev", "none", "unknown")

	orig := buildInfoSource
	t.Cleanup(func() { buildInfoSource = orig })

	buildInfoSource = func() (*debug.BuildInfo, bool) { return nil, false }
	if got := resolveVersionInfo(); got.Version != "dev" || got.Commit != "none" || got.Built != "unknown" {
		t.Fatalf("defaults must survive when build info is unavailable, got %+v", got)
	}

	buildInfoSource = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	}
	if got := resolveVersionInfo(); got.Version != "dev" {
		t.Fatalf("(devel) module version must not replace the dev default, got %q", got.Version)
	}
}
