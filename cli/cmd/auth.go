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
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Retab",
	Long: `Log in to Retab.

By default, opens your browser for an OAuth login flow via WorkOS. The
resulting tokens are saved to ~/.retab/config.json and refreshed
transparently when needed.

For headless setups (CI, servers, scripts) you can skip the browser flow
and store a long-lived API key instead:

  retab auth login --api-key sk_live_…

In both cases, RETAB_API_KEY remains honored as an environment-only
override and takes precedence over anything stored on disk.`,
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
