package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTestRootCmd builds a cobra command that carries the same persistent
// flags as the real root, so resolveCredential can read them. cmd.Root()
// returns the command itself when it has no parent.
func newTestRootCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("api-key", "", "")
	cmd.PersistentFlags().String("base-url", "", "")
	cmd.PersistentFlags().Bool("debug", false, "")
	cmd.PersistentFlags().Bool("live", false, "")
	cmd.PersistentFlags().String("env", "", "")
	return cmd
}

// isolateHome points HOME at an empty temp dir and clears the credential
// env vars so a test starts from a known-empty config state.
func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("RETAB_API_KEY", "")
	t.Setenv("RETAB_BASE_URL", "")
	return home
}

// --- slug validation -------------------------------------------------------

func TestValidateSlug(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"test", "test", false},
		{"production", "production", false},
		{"live", "production", false}, // alias folds onto production
		{"LIVE", "production", false}, // case-insensitive
		{"Test", "test", false},
		{"staging", "staging", false},
		{"qa", "qa", false},
		{"demo-1", "demo-1", false},
		{"a", "a", false},
		{"", "", true},
		{"   ", "", true},
		{"-bad", "", true},     // must start with alnum
		{"_bad", "", true},     // must start with alnum
		{"Bad Slug", "", true}, // space is rejected after lowercasing
		{"with.dot", "", true},
		{strings.Repeat("a", 64), "", true}, // 64 chars exceeds the 63-char cap
		{strings.Repeat("a", 63), strings.Repeat("a", 63), false},
	}
	for _, c := range cases {
		got, err := validateSlug(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("validateSlug(%q): expected error, got %q", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("validateSlug(%q): unexpected error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("validateSlug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEnvironmentFromKeyPrefix(t *testing.T) {
	cases := map[string]string{
		"rt_test_abc":       slugTest,
		"rt_live_abc":       slugProduction,
		"sk_retab_test_abc": slugTest,
		"sk_retab_abc":      slugProduction, // legacy -> production
		"unknown_abc":       "",
	}
	for key, want := range cases {
		if got := environmentFromKeyPrefix(key); got != want {
			t.Errorf("environmentFromKeyPrefix(%q) = %q, want %q", key, got, want)
		}
	}
}

// --- resolver precedence branches -----------------------------------------

// 1. --api-key flag wins over everything.
func TestResolveCredential_FlagKey(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		APIKey:             "rt_test_legacy",
		DefaultEnvironment: slugTest,
		Environments: map[string]*environmentProfile{
			slugTest: {APIKey: "rt_test_profile"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--api-key", "rt_live_flag"}); err != nil {
		t.Fatal(err)
	}
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceFlagKey {
		t.Errorf("source = %q, want %q", cred.Source, sourceFlagKey)
	}
	if cred.APIKey != "rt_live_flag" {
		t.Errorf("APIKey = %q, want rt_live_flag", cred.APIKey)
	}
	if cred.ExpectedEnvironment != slugProduction {
		t.Errorf("ExpectedEnvironment = %q, want production", cred.ExpectedEnvironment)
	}
}

// 2. RETAB_API_KEY env wins when no flag.
func TestResolveCredential_EnvKey(t *testing.T) {
	isolateHome(t)
	t.Setenv("RETAB_API_KEY", "rt_test_env")
	if err := saveConfig(retabConfig{
		Environments: map[string]*environmentProfile{
			slugTest: {APIKey: "rt_test_profile"},
		},
		DefaultEnvironment: slugTest,
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceEnvKey {
		t.Errorf("source = %q, want %q", cred.Source, sourceEnvKey)
	}
	if cred.APIKey != "rt_test_env" {
		t.Errorf("APIKey = %q, want rt_test_env", cred.APIKey)
	}
}

// 3. --env <slug> selects the named profile.
func TestResolveCredential_EnvFlagProfile(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		DefaultEnvironment: slugTest,
		Environments: map[string]*environmentProfile{
			slugTest:  {APIKey: "rt_test_abc"},
			"staging": {APIKey: "rt_test_staging"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--env", "staging"}); err != nil {
		t.Fatal(err)
	}
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceEnvFlag {
		t.Errorf("source = %q, want %q", cred.Source, sourceEnvFlag)
	}
	if cred.APIKey != "rt_test_staging" {
		t.Errorf("APIKey = %q, want rt_test_staging", cred.APIKey)
	}
	if cred.ProfileSlug != "staging" {
		t.Errorf("ProfileSlug = %q, want staging", cred.ProfileSlug)
	}
	if !cred.Override {
		t.Error("Override should be true for --env")
	}
}

// 3b. --env with no matching profile is an error.
func TestResolveCredential_EnvFlagMissingProfile(t *testing.T) {
	isolateHome(t)
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--env", "qa"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "qa") {
		t.Fatalf("expected missing-profile error mentioning qa, got %v", err)
	}
}

// 4. --live selects the production profile.
func TestResolveCredential_LiveFlagProfile(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		DefaultEnvironment: slugTest,
		Environments: map[string]*environmentProfile{
			slugTest:       {APIKey: "rt_test_abc"},
			slugProduction: {APIKey: "rt_live_xyz"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--live"}); err != nil {
		t.Fatal(err)
	}
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceLiveFlag {
		t.Errorf("source = %q, want %q", cred.Source, sourceLiveFlag)
	}
	if cred.APIKey != "rt_live_xyz" {
		t.Errorf("APIKey = %q, want rt_live_xyz", cred.APIKey)
	}
}

// 4b. --live with no production profile is a clear error.
func TestResolveCredential_LiveFlagMissingProfile(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		DefaultEnvironment: slugTest,
		Environments: map[string]*environmentProfile{
			slugTest: {APIKey: "rt_test_abc"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--live"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "live credential") {
		t.Fatalf("expected 'no live credential' error, got %v", err)
	}
}

// 5. stored default_environment profile.
func TestResolveCredential_DefaultEnvironment(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		DefaultEnvironment: "staging",
		Environments: map[string]*environmentProfile{
			"staging": {APIKey: "rt_test_staging"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceDefaultEnv {
		t.Errorf("source = %q, want %q", cred.Source, sourceDefaultEnv)
	}
	if cred.APIKey != "rt_test_staging" {
		t.Errorf("APIKey = %q, want rt_test_staging", cred.APIKey)
	}
	if cred.Override {
		t.Error("Override should be false for the stored default")
	}
}

func TestResolveCredential_StoredAccessToken(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		AccessToken: "acctk_production_stored",
		OAuth:       &oauthTokens{AccessToken: "oauth-access-token"},
		APIKey:      "sk_retab_legacy",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceAccessToken {
		t.Errorf("source = %q, want %q", cred.Source, sourceAccessToken)
	}
	if cred.AccessToken != "acctk_production_stored" {
		t.Errorf("AccessToken = %q, want acctk_production_stored", cred.AccessToken)
	}
	if cred.APIKey != "" || cred.OAuth != nil {
		t.Errorf("stored access token must not resolve APIKey/OAuth: %+v", cred)
	}
	if preview := cred.KeyPreview(); preview == "" || strings.Contains(preview, cred.AccessToken) {
		t.Errorf("access token preview = %q, want redacted non-empty preview", preview)
	}
}

// 7. stored OAuth session, when no API-key profile or access token is selected.
func TestResolveCredential_OAuth(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		OAuth: &oauthTokens{AccessToken: "oauth-access-token"},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceOAuth {
		t.Errorf("source = %q, want %q", cred.Source, sourceOAuth)
	}
	if cred.OAuth == nil || cred.OAuth.AccessToken != "oauth-access-token" {
		t.Errorf("OAuth not carried through: %+v", cred.OAuth)
	}
	if cred.APIKey != "" {
		t.Errorf("APIKey should be empty for OAuth, got %q", cred.APIKey)
	}
}

// 6b. OAuth session whose selected environment is production resolves to
// the production gate slug so high-risk commands are confirmed.
func TestResolveCredential_OAuth_ProductionEnvironment(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		OAuth:           &oauthTokens{AccessToken: "oauth-access-token"},
		EnvironmentType: "production",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceOAuth {
		t.Errorf("source = %q, want %q", cred.Source, sourceOAuth)
	}
	if cred.ExpectedEnvironment != slugProduction {
		t.Errorf("OAuth production session must resolve to %q, got %q", slugProduction, cred.ExpectedEnvironment)
	}
}

// 6c. OAuth session in an explicitly non-production environment resolves to
// "" so the gate never engages.
func TestResolveCredential_OAuth_NonProductionEnvironment(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		OAuth:           &oauthTokens{AccessToken: "oauth-access-token"},
		EnvironmentType: "non_production",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.ExpectedEnvironment != "" {
		t.Errorf("OAuth non-production session must resolve to \"\", got %q", cred.ExpectedEnvironment)
	}
}

// 6d. OAuth config written before environment-type persistence fails safe
// to production so the gate still engages.
func TestResolveCredential_OAuth_UnknownEnvironment_FailsSafeToProduction(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		OAuth: &oauthTokens{AccessToken: "oauth-access-token"},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.ExpectedEnvironment != slugProduction {
		t.Errorf("OAuth session without a persisted environment type must fail safe to %q, got %q", slugProduction, cred.ExpectedEnvironment)
	}
}

// 7. legacy stored api_key, treated as production-scoped.
func TestResolveCredential_LegacyKey(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{APIKey: "sk_retab_legacy"}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if cred.Source != sourceLegacyKey {
		t.Errorf("source = %q, want %q", cred.Source, sourceLegacyKey)
	}
	if cred.APIKey != "sk_retab_legacy" {
		t.Errorf("APIKey = %q, want sk_retab_legacy", cred.APIKey)
	}
	if cred.ExpectedEnvironment != slugProduction {
		t.Errorf("legacy key should resolve to production, got %q", cred.ExpectedEnvironment)
	}
}

// 8. no credentials -> error.
func TestResolveCredential_Unauthenticated(t *testing.T) {
	isolateHome(t)
	cmd := newTestRootCmd()
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("expected 'no credentials' error, got %v", err)
	}
}

// --- conflict rejection ----------------------------------------------------

func TestResolveCredential_ConflictFlagKeyAndLive(t *testing.T) {
	isolateHome(t)
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--api-key", "rt_live_abc", "--live"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "--live cannot be combined with --api-key") {
		t.Fatalf("expected --live/--api-key conflict error, got %v", err)
	}
}

func TestResolveCredential_ConflictFlagKeyAndEnv(t *testing.T) {
	isolateHome(t)
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--api-key", "rt_test_abc", "--env", "staging"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "--api-key cannot be combined with --env") {
		t.Fatalf("expected --api-key/--env conflict error, got %v", err)
	}
}

func TestResolveCredential_ConflictEnvKeyAndLive(t *testing.T) {
	isolateHome(t)
	t.Setenv("RETAB_API_KEY", "rt_test_env")
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--live"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "RETAB_API_KEY") {
		t.Fatalf("expected RETAB_API_KEY/--live conflict error, got %v", err)
	}
}

func TestResolveCredential_ConflictEnvKeyAndEnvFlag(t *testing.T) {
	isolateHome(t)
	t.Setenv("RETAB_API_KEY", "rt_test_env")
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--env", "staging"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "RETAB_API_KEY") {
		t.Fatalf("expected RETAB_API_KEY/--env conflict error, got %v", err)
	}
}

// --live + --env test is contradictory.
func TestResolveCredential_ConflictLiveAndEnvTest(t *testing.T) {
	isolateHome(t)
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--live", "--env", "test"}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveCredential(cmd)
	if err == nil || !strings.Contains(err.Error(), "--live conflicts with --env") {
		t.Fatalf("expected --live/--env test conflict error, got %v", err)
	}
}

// --live + --env production is equivalent and accepted.
func TestResolveCredential_LiveAndEnvProductionEquivalent(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		Environments: map[string]*environmentProfile{
			slugProduction: {APIKey: "rt_live_xyz"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newTestRootCmd()
	if err := cmd.ParseFlags([]string{"--live", "--env", "production"}); err != nil {
		t.Fatal(err)
	}
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatalf("--live + --env production should be accepted, got %v", err)
	}
	if cred.APIKey != "rt_live_xyz" {
		t.Errorf("APIKey = %q, want rt_live_xyz", cred.APIKey)
	}
}
