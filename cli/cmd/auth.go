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
	Short: "Store the Retab API key in ~/.retab/config.json",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Root().PersistentFlags().GetString("api-key")
		baseURL, _ := cmd.Root().PersistentFlags().GetString("base-url")
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
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		path, _ := configPath()
		fmt.Fprintf(os.Stderr, "Saved credentials to %s\n", path)
		return nil
	}),
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

		var source, apiKey, baseURL string
		switch {
		case flagKey != "":
			apiKey, source = flagKey, "--api-key flag"
		case envKey != "":
			apiKey, source = envKey, "RETAB_API_KEY env"
		case cfg.APIKey != "":
			apiKey, source = cfg.APIKey, "~/.retab/config.json"
		}
		baseURL, _ = cmd.Root().PersistentFlags().GetString("base-url")
		if baseURL == "" {
			baseURL = os.Getenv("RETAB_BASE_URL")
		}
		if baseURL == "" {
			baseURL = cfg.BaseURL
		}

		out := map[string]any{
			"authenticated": apiKey != "",
			"source":        source,
			"api_key_set":   apiKey != "",
		}
		if apiKey != "" {
			out["api_key_preview"] = redactKey(apiKey)
		}
		if baseURL != "" {
			out["base_url"] = baseURL
		}
		if apiKey == "" {
			out["hint"] = "run `retab auth login` or set RETAB_API_KEY"
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
	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
