package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// envCmd groups the local environment-profile management commands. These
// are a WorkOS-style environment manager: named local profiles, each
// holding one API key, plus a local default. None of them mutate server
// state — `env switch` only changes which stored key the CLI sends.
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage local Retab environment profiles",
	Long: `Manage named local environment profiles in ~/.retab/config.json.

Each profile stores one API key under a slug ("test", "production",
"staging", ...). The slug is a CLI-local name; the server decides the real
customer environment from the API key. Profile commands never change
server state.

Reserved slugs: ` + "`test`" + ` and ` + "`production`" + `. ` + "`live`" + ` is accepted as an
alias for ` + "`production`" + `. Custom slugs must match ^[a-z0-9][a-z0-9_-]{0,62}$.`,
	Example: `  # Add the two canonical environments.
  retab env add test --api-key rt_test_...
  retab env add production --api-key rt_live_...

  # Add a custom environment.
  retab env add staging --api-key rt_test_...

  # Inspect and switch the local default.
  retab env list
  retab env switch production
  retab env remove staging`,
}

var envAddCmd = &cobra.Command{
	Use:   "add <slug>",
	Short: "Add or update a local environment profile",
	Long: `Store an API key under a named environment profile.

If no other profile exists yet, the new profile also becomes the local
default. Re-running ` + "`env add`" + ` for an existing slug overwrites that
profile's key in place (key rotation).`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		slug, err := validateSlug(args[0])
		if err != nil {
			return err
		}
		apiKey, _ := cmd.Flags().GetString("api-key")
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			prompted, perr := promptSecret(fmt.Sprintf("API key for %q: ", slug))
			if perr != nil {
				return perr
			}
			apiKey = strings.TrimSpace(prompted)
		}
		if apiKey == "" {
			return fmt.Errorf("an API key is required: pass --api-key")
		}

		name, _ := cmd.Flags().GetString("name")
		baseURL, _ := cmd.Flags().GetString("base-url")

		cfg, _ := loadConfig()
		if cfg.Environments == nil {
			cfg.Environments = map[string]*environmentProfile{}
		}
		existing := cfg.Environments[slug]
		profile := &environmentProfile{
			Name:          name,
			APIKey:        apiKey,
			APIKeyPreview: redactKey(apiKey),
			BaseURL:       baseURL,
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		}
		if existing != nil {
			// Preserve metadata on update; only the key/name/base-url change.
			profile.CreatedAt = existing.CreatedAt
			if name == "" {
				profile.Name = existing.Name
			}
			if baseURL == "" {
				profile.BaseURL = existing.BaseURL
			}
			profile.ServerEnvironmentSlug = existing.ServerEnvironmentSlug
			profile.ServerEnvironmentID = existing.ServerEnvironmentID
		}
		if profile.Name == "" {
			profile.Name = defaultDisplayName(slug)
		}
		cfg.Environments[slug] = profile

		// First profile becomes the default automatically.
		setDefault, _ := cmd.Flags().GetBool("default")
		if cfg.DefaultEnvironment == "" || setDefault {
			cfg.DefaultEnvironment = slug
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		verb := "Added"
		if existing != nil {
			verb = "Updated"
		}
		fmt.Fprintf(os.Stderr, "%s environment %q (%s).\n", verb, slug, redactKey(apiKey))
		if cfg.DefaultEnvironment == slug {
			fmt.Fprintf(os.Stderr, "Default CLI environment is now %q.\n", slug)
		}
		return nil
	}),
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local environment profiles",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		cfg, _ := loadConfig()
		return writeEnvList(cmd.OutOrStdout(), cfg)
	}),
}

var envSwitchCmd = &cobra.Command{
	Use:   "switch <slug>",
	Short: "Change the local default environment profile",
	Long: `Change which stored profile the CLI uses when no --env, --live,
--api-key, or RETAB_API_KEY override is supplied.

This only changes local state. It never mutates server state and never
changes what an API key can access.`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		slug, err := validateSlug(args[0])
		if err != nil {
			return err
		}
		cfg, _ := loadConfig()
		if cfg.Environments[slug] == nil {
			return fmt.Errorf("no environment profile %q. Run `retab env add %s --api-key <key>` first", slug, slug)
		}
		cfg.DefaultEnvironment = slug
		if err := saveConfig(cfg); err != nil {
			return err
		}
		if slug == slugProduction {
			fmt.Fprintln(os.Stderr, "Switched default CLI environment to production.")
			fmt.Fprintln(os.Stderr, "Future commands without --env or --live will use production credentials.")
		} else {
			fmt.Fprintf(os.Stderr, "Switched default CLI environment to %q.\n", slug)
		}
		if os.Getenv("RETAB_API_KEY") != "" {
			fmt.Fprintln(os.Stderr, "warning: RETAB_API_KEY is set in your shell and will override the active CLI environment.")
			fmt.Fprintln(os.Stderr, "  Unset RETAB_API_KEY to use the `retab env switch` default.")
		}
		return nil
	}),
}

var envRemoveCmd = &cobra.Command{
	Use:   "remove <slug>",
	Short: "Remove a local environment profile",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		slug, err := validateSlug(args[0])
		if err != nil {
			return err
		}
		cfg, _ := loadConfig()
		if cfg.Environments[slug] == nil {
			return fmt.Errorf("no environment profile %q", slug)
		}
		delete(cfg.Environments, slug)
		// If the removed profile was the default, clear it. Don't guess a
		// replacement — the user should pick one explicitly.
		if cfg.DefaultEnvironment == slug {
			cfg.DefaultEnvironment = ""
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Removed environment %q.\n", slug)
		if cfg.DefaultEnvironment == "" && len(cfg.Environments) > 0 {
			fmt.Fprintln(os.Stderr, "No default environment set. Run `retab env switch <slug>` to pick one.")
		}
		return nil
	}),
}

// defaultDisplayName returns a human label for a slug when none was given.
func defaultDisplayName(slug string) string {
	switch slug {
	case slugTest:
		return "Test"
	case slugProduction:
		return "Production"
	default:
		// Title-case the first rune for custom slugs.
		if slug == "" {
			return ""
		}
		return strings.ToUpper(slug[:1]) + slug[1:]
	}
}

// profileType classifies a profile as a test- or production-type
// environment, using the server-resolved slug when available and falling
// back to the key prefix, then the local slug.
func profileType(slug string, profile *environmentProfile) string {
	if profile.ServerEnvironmentSlug == slugProduction {
		return slugProduction
	}
	if profile.ServerEnvironmentSlug == slugTest {
		return slugTest
	}
	if env := environmentFromKeyPrefix(profile.APIKey); env != "" {
		return env
	}
	if slug == slugProduction {
		return slugProduction
	}
	return slugTest
}

// writeEnvList renders the profile table (TTY) or JSON (non-TTY / --output
// json), mirroring the auth-status dual-output discipline.
func writeEnvList(w io.Writer, cfg retabConfig) error {
	slugs := make([]string, 0, len(cfg.Environments))
	for slug := range cfg.Environments {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	jsonMode := !isTerminalWriter(w)
	if f := rootCmd.PersistentFlags().Lookup("output"); f != nil && f.Value.String() == string(OutputJSON) {
		jsonMode = true
	}

	if jsonMode {
		data := make([]map[string]any, 0, len(slugs))
		for _, slug := range slugs {
			p := cfg.Environments[slug]
			entry := map[string]any{
				"slug":                    slug,
				"name":                    p.Name,
				"type":                    profileType(slug, p),
				"server_environment_slug": p.ServerEnvironmentSlug,
				"active":                  slug == cfg.DefaultEnvironment,
				"api_key_preview":         p.APIKeyPreview,
			}
			if p.LastVerifiedAt != "" {
				entry["last_verified_at"] = p.LastVerifiedAt
			}
			data = append(data, entry)
		}
		return writeAuthStatusJSON(w, map[string]any{"data": data})
	}

	if len(slugs) == 0 {
		fmt.Fprintln(w, "No environment profiles configured.")
		fmt.Fprintln(w, "Add one with `retab env add test --api-key rt_test_...`.")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "SLUG\tNAME\tTYPE\tACTIVE\tCREDENTIAL")
	for _, slug := range slugs {
		p := cfg.Environments[slug]
		active := "no"
		if slug == cfg.DefaultEnvironment {
			active = "yes"
		}
		preview := p.APIKeyPreview
		if preview == "" {
			preview = redactKey(p.APIKey)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", slug, p.Name, profileType(slug, p), active, preview)
	}
	return tw.Flush()
}

func init() {
	envAddCmd.Flags().String("api-key", "", "API key to store for this environment")
	envAddCmd.Flags().String("name", "", "display name for the environment profile")
	envAddCmd.Flags().String("base-url", "", "optional Retab deployment base URL for this profile")
	envAddCmd.Flags().Bool("default", false, "also set this profile as the local default")

	envCmd.AddCommand(envAddCmd, envListCmd, envSwitchCmd, envRemoveCmd)
	rootCmd.AddCommand(envCmd)
}
