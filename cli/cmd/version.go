package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Build-time identity, populated by `-ldflags "-X .../cmd.version=..."`
// (see .goreleaser.yaml and .github/workflows/release-cli.yml). Two
// surfaces read these vars:
//
//   - `retab --version`  — cobra wires this from rootCmd.Version (see
//     root.go's init()). Short, just the version.
//   - `retab version`    — registered below. Prints version + commit
//   - ISO build date.
//
// We expose both because users coming from `git`/`go`/`node` instinctively
// type the subcommand form first; getting `unknown command "version"` is a
// needless first-run papercut. The subcommand also surfaces commit + date
// which the flag doesn't, and is the format we ask for in bug reports.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Built   string `json:"built"`
}

// buildInfoSource is swapped in tests; in production it reads the module
// build metadata Go embeds into every binary (module version + VCS stamps).
var buildInfoSource = debug.ReadBuildInfo

// resolveVersionInfo returns the build identity to display.
//
// When the binary is built by GoReleaser the linker stamps `version`,
// `commit`, and `date` via `-ldflags -X` and those win unconditionally.
// But a binary produced by a plain `go build`/`go install` carries no
// such stamps, so the package-level defaults ("dev"/"none"/"unknown")
// would otherwise make `retab version` unattributable. For that path we
// fall back to `runtime/debug.ReadBuildInfo()`, which exposes the module
// version (e.g. `v0.1.0` for a tagged `go install …@v0.1.0`, or
// `(devel)`) plus the `vcs.revision` / `vcs.time` settings the Go
// toolchain records from the source repository.
func resolveVersionInfo() versionInfo {
	info := versionInfo{Version: version, Commit: commit, Built: date}

	bi, ok := buildInfoSource()
	if !ok {
		return info
	}

	if info.Version == "dev" || info.Version == "" {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			info.Version = v
		}
	}

	if info.Commit == "none" || info.Commit == "" {
		for _, setting := range bi.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				info.Commit = setting.Value
				break
			}
		}
	}

	if info.Built == "unknown" || info.Built == "" {
		for _, setting := range bi.Settings {
			if setting.Key == "vcs.time" && setting.Value != "" {
				info.Built = setting.Value
				break
			}
		}
	}

	return info
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version, commit, and build date",
	Long: `Print the build identity of this binary.

Equivalent to ` + "`retab --version`" + ` but also prints the git commit
SHA and ISO build timestamp that GoReleaser injects at link time —
useful when filing bug reports or pinning a remote machine to a known
build.`,
	Example: `  retab version
  # retab 0.1.0 (commit a1b2c3d, built 2026-05-14T15:03:21Z)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := resolveVersionInfo()
		var raw string
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
		switch raw {
		case string(OutputJSON):
			return printJSON(info)
		case string(OutputTable), string(OutputCSV):
			return RenderList(cmd.OutOrStdout(), OutputFormat(raw), map[string]any{
				"data": []map[string]string{
					{
						"version": info.Version,
						"commit":  info.Commit,
						"built":   info.Built,
					},
				},
			}, []TableColumn{
				{Header: "VERSION", Extract: func(row any) string { return versionTableCell(row, "version") }},
				{Header: "COMMIT", Extract: func(row any) string { return versionTableCell(row, "commit") }},
				{Header: "BUILT", Extract: func(row any) string { return versionTableCell(row, "built") }},
			})
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "retab %s (commit %s, built %s)\n",
			info.Version, info.Commit, info.Built)
		return err
	},
}

func versionTableCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok {
		return ""
	}
	return stringifyCell(value)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
