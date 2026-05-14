package cmd

import (
	"bufio"
	"fmt"
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
		if source == "" {
			out["hint"] = "run `retab auth login` to authenticate"
			return printJSON(out)
		}

		// Best-effort verification — list workflows with limit=1.
		client, err := newClient(cmd)
		if err != nil {
			out["valid"] = false
			out["error"] = err.Error()
			return printJSON(out)
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
		return printJSON(out)
	}),
}

func redactKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
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

	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
