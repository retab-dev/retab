package cmd

// Build-time identity, populated by `-ldflags "-X .../cmd.version=..."`
// (see .goreleaser.yaml and .github/workflows/release-cli.yml). The values
// surface to users via `retab --version`, which cobra wires up from
// rootCmd.Version — see root.go's init() for the assignment.
//
// There is intentionally NO `retab version` subcommand. The flag does the
// job and a dedicated subcommand was just menu noise.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)
