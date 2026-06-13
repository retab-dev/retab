package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Update-notifier for the Retab CLI.
//
// Design follows npm's `update-notifier`, not `bun upgrade`: bun has no
// passive notifier at all (it only checks the network when you explicitly
// run `bun upgrade`). The good, low-latency design is the update-notifier
// one — and it has two halves:
//
//  1. A detached background "daemon" (`retab __update-check`, hidden) does
//     the actual network fetch and writes the result to a small cache file.
//     We spawn it with Start()+Release() and never wait on it, so it adds
//     *zero* latency to the user's real command and survives the parent
//     exiting. On the next invocation the cache is already warm.
//
//  2. The foreground process only ever *reads* that cache. If a newer
//     version is cached it queues a one-line notice that prints to stderr
//     after the command finishes. Reading a local JSON file is instant, so
//     the notice never blocks anything.
//
// The check is suppressed for unreleased (`dev`) builds, in CI, on non-TTY
// stderr, when NO_UPDATE_NOTIFIER / RETAB_NO_UPDATE_NOTIFIER is set, and
// for the version/completion/daemon commands themselves.

const (
	updateCheckInterval = 24 * time.Hour
	updateFetchTimeout  = 5 * time.Second
	updateDaemonCommand = "__update-check"

	defaultUpdateRepo      = "retab-dev/retab"
	defaultUpdateTagPrefix = "cli-"
	defaultGitHubAPIDomain = "api.github.com"

	updateInstallCommand = "curl -fsSL https://retab.com/install.sh | sh"
)

// updateCache is the on-disk shape at ~/.retab/update-check.json. It is a
// pure cache — losing it just means one extra background fetch, so unlike
// config.json it carries no irreplaceable state.
type updateCache struct {
	LastCheckedAt time.Time `json:"last_checked_at"`
	LatestVersion string    `json:"latest_version"`
}

func updateCachePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update-check.json"), nil
}

func loadUpdateCache() (updateCache, error) {
	var c updateCache
	path, err := updateCachePath()
	if err != nil {
		return c, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		// A corrupt cache is just a stale cache — force a refresh rather
		// than surfacing an error the user can't act on.
		return updateCache{}, nil
	}
	return c, nil
}

func saveUpdateCache(c updateCache) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path, err := updateCachePath()
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	// temp + rename so a concurrent reader never sees a half-written file.
	tmp, err := os.CreateTemp(dir, "update-check.json.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return err
	}
	return nil
}

func (c updateCache) isStale(now time.Time, interval time.Duration) bool {
	if c.LastCheckedAt.IsZero() {
		return true
	}
	return now.Sub(c.LastCheckedAt) >= interval
}

// semver is a parsed major.minor.patch. Pre-release and build metadata are
// dropped — our release tags are plain `cli-vX.Y.Z`, and a notifier should
// never nag someone onto a pre-release they didn't ask for.
type semver [3]int

func parseSemver(v string) (semver, bool) {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	if v == "" {
		return semver{}, false
	}
	parts := strings.Split(v, ".")
	if len(parts) > 3 {
		return semver{}, false
	}
	var out semver
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return semver{}, false
		}
		out[i] = n
	}
	return out, true
}

func compareSemver(a, b semver) int {
	for i := range 3 {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}

// isNewerVersion reports whether latest is strictly greater than current.
// Unparseable inputs return false: we'd rather miss a notice than nag on
// garbage (e.g. a `dev` build or a malformed tag).
func isNewerVersion(current, latest string) bool {
	cur, okCur := parseSemver(current)
	lat, okLat := parseSemver(latest)
	if !okCur || !okLat {
		return false
	}
	return compareSemver(lat, cur) > 0
}

type githubRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// parseLatestCLIVersion extracts the highest stable CLI version from a
// GitHub `releases` API payload. Releases are tagged `cli-vX.Y.Z` (the SDK,
// docs, etc. live under other prefixes in the same repo), so we filter by
// tagPrefix, then strip it plus any leading `v` and keep the max semver.
func parseLatestCLIVersion(body []byte, tagPrefix string) (string, error) {
	var releases []githubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("parse releases: %w", err)
	}
	var best semver
	var bestStr string
	found := false
	for _, r := range releases {
		if r.Draft || r.Prerelease {
			continue
		}
		if !strings.HasPrefix(r.TagName, tagPrefix) {
			continue
		}
		raw := strings.TrimPrefix(r.TagName, tagPrefix)
		ver, ok := parseSemver(raw)
		if !ok {
			continue
		}
		if !found || compareSemver(ver, best) > 0 {
			best = ver
			bestStr = strings.TrimPrefix(strings.TrimSpace(raw), "v")
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("no %sX.Y.Z release found", tagPrefix)
	}
	return bestStr, nil
}

func updateRepo() string {
	if v := os.Getenv("RETAB_REPO"); v != "" {
		return v
	}
	return defaultUpdateRepo
}

func updateTagPrefix() string {
	if v := os.Getenv("RETAB_TAG_PREFIX"); v != "" {
		return v
	}
	return defaultUpdateTagPrefix
}

func githubAPIDomain() string {
	if v := os.Getenv("GITHUB_API_DOMAIN"); v != "" {
		return v
	}
	return defaultGitHubAPIDomain
}

// fetchLatestCLIVersion queries the GitHub releases API for the newest
// stable CLI release. A GITHUB_TOKEN, if present, is sent to dodge the
// 60-req/hr unauthenticated rate limit (mirrors what `bun upgrade` does).
func fetchLatestCLIVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("https://%s/repos/%s/releases?per_page=100", githubAPIDomain(), updateRepo())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "retab-cli")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if token := os.Getenv("GITHUB_ACCESS_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github releases: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return "", err
	}
	return parseLatestCLIVersion(body, updateTagPrefix())
}

// notifierEnabled decides whether the foreground process should bother with
// the update notifier at all. Pure function of its inputs so it's testable
// without faking a TTY, env, or os.Args.
func notifierEnabled(args []string, env func(string) string, currentVersion string, stderrIsTTY bool) bool {
	// Unreleased/local builds: the version isn't a real semver, so any
	// comparison is meaningless and the daemon would just spam GitHub.
	if _, ok := parseSemver(currentVersion); !ok {
		return false
	}
	if !stderrIsTTY {
		return false
	}
	if env("CI") != "" {
		return false
	}
	if env("NO_UPDATE_NOTIFIER") != "" || env("RETAB_NO_UPDATE_NOTIFIER") != "" {
		return false
	}
	if notifierSkippableCommand(args) {
		return false
	}
	return true
}

// notifierSkippableCommand suppresses the notice for invocations where it
// would be noise or recursive: the daemon itself, the version surfaces, and
// shell-completion output (which must stay machine-clean).
func notifierSkippableCommand(args []string) bool {
	for _, a := range args {
		// `retab --version` / `retab -v` before any subcommand is a version
		// query. Other flags are skipped — only the first positional token is
		// the actual subcommand, so a value like `files get version` must not
		// match.
		if a == "--version" || a == "-v" {
			return true
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		switch a {
		case updateDaemonCommand, "update", "completion", "version":
			return true
		}
		return false
	}
	return false
}

// stderrIsTTY reports whether stderr is an interactive terminal. It checks
// the fd directly rather than going through paletteFor, which also returns
// empty under NO_COLOR — and NO_COLOR should only drop colour, never
// suppress the notice itself.
func stderrIsTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// startUpdateNotifier is called once at the top of Execute. It reads the
// cache, spawns the background daemon if the cache is stale, and returns a
// closure that prints the deferred notice (if any) once the command has
// finished. The returned closure is always safe to call.
func startUpdateNotifier() func() {
	noop := func() {}
	if !notifierEnabled(os.Args[1:], os.Getenv, version, stderrIsTTY()) {
		return noop
	}

	cache, _ := loadUpdateCache()
	if cache.isStale(time.Now(), updateCheckInterval) {
		spawnUpdateDaemon()
	}

	if cache.LatestVersion == "" || !isNewerVersion(version, cache.LatestVersion) {
		return noop
	}
	latest := cache.LatestVersion
	return func() { printUpdateNotice(os.Stderr, version, latest) }
}

// spawnUpdateDaemon launches `retab __update-check` as a detached child and
// returns immediately. Start()+Release() orphans the process so it keeps
// running (and finishes its fetch) even after this command exits. All
// failures are swallowed — a notifier must never break the real command.
func spawnUpdateDaemon() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe, updateDaemonCommand)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return
	}
	_ = cmd.Process.Release()
}

func printUpdateNotice(w io.Writer, current, latest string) {
	s := paletteFor(w)
	helpFprintf(w, "\n%sUpdate available%s %s%s%s → %s%s%s\n",
		s.brand, s.reset, s.dim, current, s.reset, s.accent, latest, s.reset)
	helpFprintf(w, "Run %s%s%s to upgrade.\n\n",
		s.cyan, updateInstallCommand, s.reset)
}

// updateCheckCmd is the hidden background daemon. It performs the network
// fetch and writes the cache, then exits silently. It is never meant to be
// run by hand (though it is harmless to), so it produces no output and never
// returns an error that would print a stack of usage text.
var updateCheckCmd = &cobra.Command{
	Use:           updateDaemonCommand,
	Hidden:        true,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), updateFetchTimeout)
		defer cancel()
		prev, _ := loadUpdateCache()
		latest, err := fetchLatestCLIVersion(ctx)
		if err != nil {
			// Record the attempt anyway so a flaky GitHub doesn't make us
			// re-spawn on every single invocation for the next 24h, but
			// keep any version we resolved on a prior successful run.
			_ = saveUpdateCache(updateCache{
				LastCheckedAt: time.Now().UTC(),
				LatestVersion: prev.LatestVersion,
			})
			return nil
		}
		_ = saveUpdateCache(updateCache{
			LastCheckedAt: time.Now().UTC(),
			LatestVersion: latest,
		})
		return nil
	},
}

// simulatedNotice returns a believable (current → latest) pair for the
// `--test-dev` preview. On a real build it bumps the running version's minor;
// on a `dev` build (which has no parseable version) it falls back to a stand-in
// so the arrow still reads sensibly.
func simulatedNotice() (current, latest string) {
	cur, ok := parseSemver(version)
	current = version
	if !ok {
		cur = semver{0, 1, 0}
		current = "0.1.0"
	}
	latest = fmt.Sprintf("%d.%d.%d", cur[0], cur[1]+1, 0)
	return current, latest
}

// updateCmd is the hidden, user-facing counterpart to the background daemon.
//
//   - `retab update`            runs a real synchronous check against GitHub
//     and prints the notice (or "up to date"). Blocking is fine here because
//     the user asked for it explicitly — same contract as `bun upgrade`.
//   - `retab update --test-dev` renders the notice offline using a simulated
//     newer version. The real notifier is inert on `dev` builds, so this is
//     how you eyeball the styling locally without cutting a release.
//
// It does not replace the binary — upgrades go through the install script.
var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Check whether a newer Retab CLI release is available",
	Hidden:       true,
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if testDev, _ := cmd.Flags().GetBool("test-dev"); testDev {
			current, latest := simulatedNotice()
			printUpdateNotice(cmd.ErrOrStderr(), current, latest)
			return nil
		}

		if _, ok := parseSemver(version); !ok {
			fmt.Fprintf(cmd.ErrOrStderr(),
				"This is a development build (version %q) — there is no released version to compare against.\nRun `retab update --test-dev` to preview the update notice.\n",
				version)
			return nil
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), updateFetchTimeout)
		defer cancel()
		latest, err := fetchLatestCLIVersion(ctx)
		if err != nil {
			return fmt.Errorf("check for updates: %w", err)
		}
		if isNewerVersion(version, latest) {
			printUpdateNotice(cmd.ErrOrStderr(), version, latest)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "retab %s is up to date (latest release: %s)\n", version, latest)
		return nil
	},
}

func init() {
	updateCmd.Flags().Bool("test-dev", false, "preview the update notice with a simulated newer version (development aid)")
	rootCmd.AddCommand(updateCheckCmd)
	rootCmd.AddCommand(updateCmd)
}
