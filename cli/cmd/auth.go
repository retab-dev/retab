package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Retab authentication",
	Long: `Manage credentials the CLI uses to talk to the Retab API.

Covers interactive OAuth login, headless API-key login, inspecting the
currently active credential, and clearing local state. Credentials are
resolved with --api-key flag > RETAB_API_KEY env > ~/.retab/config.json
(mode 0600, parent dir 0700).`,
	Example: `  # Interactive login (browser OAuth by default)
  retab auth login

  # Headless / CI login with a long-lived key
  retab auth login --api-key=sk_live_abc123

  # Show which credential the CLI would use right now
  retab auth status

  # Forget the local credential (does not revoke the key server-side)
  retab auth logout`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Retab",
	Long: `Log in to Retab and persist the credential locally.

By default, opens a browser for an OAuth login flow via WorkOS. The
resulting tokens are saved to ~/.retab/config.json (mode 0600) and
refreshed transparently when needed.

For headless setups (CI, servers, scripts) skip the browser flow and
store a long-lived API key instead with --api-key. Re-running login is
always safe: it overwrites the saved credential in place, which is also
how key rotation works. RETAB_API_KEY remains honored as a process-wide
override and takes precedence over anything written to disk.`,
	Example: `  # Interactive OAuth flow (opens a browser)
  retab auth login

  # Headless / CI: pass the key inline
  retab auth login --api-key=sk_live_abc123

  # Headless without echoing the key to history
  retab auth login --api-key="$RETAB_API_KEY"

  # Prompt for an API key without opening a browser
  retab auth login --browser=false`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")
		baseURL, _ := cmd.Flags().GetString("base-url")
		useBrowser, _ := cmd.Flags().GetBool("browser")

		// Explicit API-key path. Direct, no browser. Backward compatible
		// with the original login UX.
		if apiKey != "" || (!useBrowser && os.Getenv("RETAB_API_KEY") != "") {
			return runAPIKeyLogin(apiKey, baseURL)
		}

		// If `--no-browser` was requested and no key was given, prompt.
		if !useBrowser {
			prompted, err := promptSecret("API key (leave blank to use browser flow): ")
			if err != nil {
				return err
			}
			if strings.TrimSpace(prompted) != "" {
				return runAPIKeyLogin(prompted, baseURL)
			}
			// fall through to browser flow
		}

		ctx, cancel := ctxFor(cmd)
		defer cancel()

		// Discovery + browser OAuth flow.
		effectiveBaseURL := baseURL
		if effectiveBaseURL == "" {
			effectiveBaseURL = os.Getenv("RETAB_API_BASE_URL")
		}
		if effectiveBaseURL == "" {
			effectiveBaseURL = os.Getenv("RETAB_BASE_URL")
		}
		disc, err := fetchOAuthDiscovery(ctx, effectiveBaseURL)
		if err != nil {
			return fmt.Errorf("OAuth discovery failed: %w", err)
		}
		tokens, err := runLoginFlow(ctx, disc, openBrowser)
		if err != nil {
			return err
		}

		cfg, _ := loadConfig()
		cfg.OAuth = tokens
		// Switching to OAuth wipes the legacy API key; users who want
		// both can re-run `retab auth login --api-key …` afterward.
		cfg.APIKey = ""
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		path, _ := configPath()
		fmt.Fprintf(os.Stderr, "Logged in. Saved OAuth tokens to %s\n", path)
		return nil
	}),
}

// runAPIKeyLogin persists an API key — the legacy auth path. Kept first-
// class so anyone scripting `retab auth login --api-key …` keeps working.
func runAPIKeyLogin(apiKey, baseURL string) error {
	if apiKey == "" {
		apiKey = os.Getenv("RETAB_API_KEY")
	}
	if apiKey == "" {
		prompted, err := promptSecret("API key: ")
		if err != nil {
			return err
		}
		apiKey = strings.TrimSpace(prompted)
	}
	if apiKey == "" {
		return fmt.Errorf("api key is required")
	}
	cfg, _ := loadConfig()
	cfg.APIKey = apiKey
	// Wipe stale OAuth state — explicit key login is the user's intent.
	cfg.OAuth = nil
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if err := saveConfig(cfg); err != nil {
		return err
	}
	path, _ := configPath()
	fmt.Fprintf(os.Stderr, "Saved API key to %s\n", path)
	return nil
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove ~/.retab/config.json",
	Long: `Delete the local credential file at ~/.retab/config.json.

Only clears LOCAL state — the API key (or OAuth refresh token) is not
revoked on the server. To revoke a key for real, rotate it in the Retab
dashboard. After logout, commands that need authentication will fail
until ` + "`retab auth login`" + ` runs again or RETAB_API_KEY is set in
the environment.`,
	Example: `  # Forget the credential on this machine
  retab auth logout

  # Switch accounts: forget, then log in again
  retab auth logout && retab auth login`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := deleteConfig(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Cleared local credentials.")
		return nil
	}),
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show which credentials the CLI is using and verify them",
	Long: `Report the credential the CLI would use for the next request, and
verify it works by making a best-effort API call (a tiny ` + "`workflows list`" + `
probe). Output is JSON with the credential ` + "`source`" + ` (flag, env, or
config), a redacted ` + "`api_key_preview`" + `, the effective ` + "`base_url`" + `, and a
` + "`valid`" + ` boolean reflecting the probe.

Useful for debugging "why is the wrong account being used?" — the
` + "`source`" + ` field disambiguates --api-key vs RETAB_API_KEY vs the config
file. Resolution order: --api-key > RETAB_API_KEY > ~/.retab/config.json.`,
	Example: `  # Quick check
  retab auth status

  # In a script — assert authenticated and key is valid
  retab auth status | jq -e '.valid == true'

  # Verify a specific key without persisting it
  retab --api-key=sk_test_xyz auth status`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		flagKey, _ := cmd.Root().PersistentFlags().GetString("api-key")
		envKey := os.Getenv("RETAB_API_KEY")
		cfg, _ := loadConfig()

		var source string
		switch {
		case flagKey != "":
			source = "--api-key flag"
		case envKey != "":
			source = "RETAB_API_KEY env"
		case cfg.OAuth != nil && cfg.OAuth.AccessToken != "":
			source = "~/.retab/config.json (oauth)"
		case cfg.APIKey != "":
			source = "~/.retab/config.json (api_key)"
		}
		baseURL, _ := cmd.Root().PersistentFlags().GetString("base-url")
		if baseURL == "" {
			baseURL = os.Getenv("RETAB_API_BASE_URL")
		}
		if baseURL == "" {
			baseURL = os.Getenv("RETAB_BASE_URL")
		}
		if baseURL == "" {
			baseURL = cfg.BaseURL
		}

		out := map[string]any{
			"authenticated": source != "",
			"source":        source,
		}
		if baseURL != "" {
			out["base_url"] = baseURL
		}
		if cfg.OAuth != nil && cfg.OAuth.AccessToken != "" {
			out["oauth"] = map[string]any{
				"authkit_domain": cfg.OAuth.AuthKitDomain,
				"expires_at":     cfg.OAuth.ExpiresAt,
				"has_refresh":    cfg.OAuth.RefreshToken != "",
			}
		}
		if flagKey != "" || envKey != "" || cfg.APIKey != "" {
			key := flagKey
			if key == "" {
				key = envKey
			}
			if key == "" {
				key = cfg.APIKey
			}
			out["api_key_preview"] = redactKey(key)
		}

		jsonOnly, _ := cmd.Flags().GetBool("json")

		if source == "" {
			out["hint"] = "run `retab auth login` to authenticate"
			return writeAuthStatus(cmd.OutOrStdout(), out, jsonOnly)
		}

		// Best-effort verification — list workflows with limit=1.
		client, err := newClient(cmd)
		if err != nil {
			out["valid"] = false
			out["error"] = err.Error()
			return writeAuthStatus(cmd.OutOrStdout(), out, jsonOnly)
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		_, err = client.Workflows.List(ctx, nil)
		if err != nil {
			out["valid"] = false
			out["error"] = err.Error()
		} else {
			out["valid"] = true
		}
		return writeAuthStatus(cmd.OutOrStdout(), out, jsonOnly)
	}),
}

// writeAuthStatus decides whether `auth status` renders JSON or a human
// status block, and writes the result to w.
//
// Output mode:
//   - --json flag set         → JSON
//   - w is not a TTY          → JSON (so pipes / redirects stay parseable)
//   - otherwise               → human block
//
// JSON output is byte-equivalent to the legacy `printJSON(out)` form
// (encoding/json's encoder with 2-space indent, trailing newline) so any
// script consuming `auth status | jq …` keeps working.
func writeAuthStatus(w io.Writer, out map[string]any, jsonOnly bool) error {
	if jsonOnly || !isTerminalWriter(w) {
		return writeAuthStatusJSON(w, out)
	}
	return writeAuthStatusHuman(w, out)
}

// writeAuthStatusJSON renders the legacy JSON shape. Encoder (not
// MarshalIndent) keeps the trailing-newline behaviour of printJSON.
func writeAuthStatusJSON(w io.Writer, out map[string]any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

// writeAuthStatusHuman renders the three-line status block. Colour is
// applied only when the destination is a real TTY and NO_COLOR is unset,
// reusing paletteFor's discipline so the rules match `retab --help`.
//
// Layout:
//
//	Logged in as <preview>
//	Source:  <source>
//	Status:  <valid|invalid|not authenticated>
//
// "Source" / "Status" labels are dim; the api-key preview is bold magenta
// (matching the Retab wordmark elsewhere in the CLI).
func writeAuthStatusHuman(w io.Writer, out map[string]any) error {
	s := paletteFor(w)

	preview, _ := out["api_key_preview"].(string)
	source, _ := out["source"].(string)

	// First line — "Logged in as <preview>" when we have any credential to
	// preview, otherwise "Not logged in" (matches the JSON's empty-source
	// case, where api_key_preview is absent because there's no key).
	if preview != "" {
		fmt.Fprintf(w, "Logged in as %s%s%s\n", s.brand, preview, s.reset)
	} else {
		fmt.Fprintln(w, "Not logged in")
	}

	// Second line — credential source.
	if source == "" {
		source = "none"
	}
	fmt.Fprintf(w, "%sSource:%s  %s\n", s.dim, s.reset, source)

	// Third line — verification result. The `valid` key is absent when
	// we never got far enough to probe (no creds, or hint path); treat
	// that as "not authenticated" so the user gets a definite answer.
	var status string
	if v, ok := out["valid"]; ok {
		if b, _ := v.(bool); b {
			status = "valid"
		} else {
			// Just "invalid" — full error detail (often multi-line, with
			// URL/code/body) stays in the JSON's `error` field. Splatting
			// it onto the Status: line would shred the calm 3-line block
			// the human view exists to provide.
			status = "invalid"
		}
	} else {
		status = "not authenticated"
	}
	fmt.Fprintf(w, "%sStatus:%s  %s\n", s.dim, s.reset, status)
	return nil
}

// isTerminalWriter mirrors paletteFor's TTY check — true only when w is
// an *os.File pointing at a terminal. bytes.Buffer, pipes, and redirected
// files all return false, which is what we want for "render JSON for
// machines, human block for humans".
func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// redactKey masks a credential for display: the first 4 and last 4
// characters are kept, the middle replaced by a FIXED-WIDTH asterisk run.
//
// The width is capped at redactMaskWidth rather than len(key)-8. An OAuth
// access token is a ~900-1200 character JWT; the old len-8 mask dumped a
// full screen of asterisks into `--debug` output and `auth status`, and
// reproduced the exact length of the secret. A fixed mask is both readable
// and leaks nothing about the credential's size.
func redactKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	const redactMaskWidth = 8
	mask := len(key) - 8
	if mask > redactMaskWidth {
		mask = redactMaskWidth
	}
	return key[:4] + strings.Repeat("*", mask) + key[len(key)-4:]
}

func promptSecret(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func init() {
	authLoginCmd.Flags().String("api-key", "", "skip the browser flow and store this API key (also reads RETAB_API_KEY)")
	authLoginCmd.Flags().String("base-url", "", "override the default API base URL")
	authLoginCmd.Flags().Bool("browser", true, "open a browser for OAuth login (set --browser=false to prompt for an API key)")

	// `--json` forces JSON output even on a TTY. Without it, status auto-
	// detects: TTY → human block, pipe/redirect → JSON (same shape as today).
	authStatusCmd.Flags().Bool("json", false, "emit JSON even when stdout is a TTY (default: human-readable on TTY, JSON when piped)")

	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
