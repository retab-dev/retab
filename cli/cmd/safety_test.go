//go:build !retab_oagen_cli_files && !retab_oagen_cli_workflows && !retab_oagen_cli_workflows_runs

package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newSafetyTestCmd builds a root + child command pair carrying the same
// persistent flags resolveCredential and the gate read. The child is what
// the gate inspects; it gets the given safety class. apiKey, when set, is
// passed via the --api-key flag so resolveCredential picks a deterministic
// credential without touching config files.
func newSafetyTestCmd(t *testing.T, cls safetyClass, apiKey string) *cobra.Command {
	t.Helper()
	root := &cobra.Command{Use: "retab"}
	root.PersistentFlags().String("api-key", "", "")
	root.PersistentFlags().String("base-url", "", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.PersistentFlags().Bool("live", false, "")
	root.PersistentFlags().String("env", "", "")

	child := &cobra.Command{Use: "publish", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(child)
	markSafety(child, cls)
	// Mirror production: --confirm is a local flag on high-risk commands
	// (see addConfirmFlag), not a global persistent flag.
	if cls == classHighRisk {
		addConfirmFlag(child)
	}

	if apiKey != "" {
		if err := root.PersistentFlags().Set("api-key", apiKey); err != nil {
			t.Fatalf("set --api-key: %v", err)
		}
	}
	return child
}

// fakeDecider is an injectable confirmDecider so the gate can be tested
// without a real TTY.
type fakeDecider struct {
	interactive bool
	answer      string
	answerErr   error
	prompted    bool
}

func (f *fakeDecider) Interactive() bool { return f.interactive }

func (f *fakeDecider) ReadConfirmation(string) (string, error) {
	f.prompted = true
	return f.answer, f.answerErr
}

func setConfirm(t *testing.T, cmd *cobra.Command, v bool) {
	t.Helper()
	val := "false"
	if v {
		val = "true"
	}
	if err := cmd.Flags().Set(confirmFlagName, val); err != nil {
		t.Fatalf("set --confirm: %v", err)
	}
}

// --- non-interactive (no TTY) -------------------------------------------

func TestProductionGate_HighRisk_NonInteractive_NoConfirm_Errors(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classHighRisk, "rt_live_secret")
	err := productionGate(cmd, &fakeDecider{interactive: false})
	if err == nil {
		t.Fatal("expected an error for a high-risk production command in non-interactive mode without --confirm")
	}
	if !strings.Contains(err.Error(), "production write requires --confirm in non-interactive mode") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestProductionGate_HighRisk_NonInteractive_WithConfirm_Allowed(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classHighRisk, "rt_live_secret")
	setConfirm(t, cmd, true)
	dec := &fakeDecider{interactive: false}
	if err := productionGate(cmd, dec); err != nil {
		t.Fatalf("expected --confirm to allow the command, got: %v", err)
	}
	if dec.prompted {
		t.Fatal("--confirm must skip the prompt entirely")
	}
}

// --- OAuth sessions are gated by the selected environment's type --------

// An OAuth session (no environment-scoped API key) whose selected
// environment is production must gate high-risk commands. This is the
// reported regression: the OAuth branch left ExpectedEnvironment empty, so
// the gate short-circuited and production mutations ran unconfirmed.
func TestProductionGate_HighRisk_OAuthProductionEnv_NonInteractive_NoConfirm_Errors(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		OAuth:           &oauthTokens{AccessToken: "oauth-access-token"},
		EnvironmentID:   "env_prod123",
		EnvironmentType: "production",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newSafetyTestCmd(t, classHighRisk, "")
	err := productionGate(cmd, &fakeDecider{interactive: false})
	if err == nil {
		t.Fatal("OAuth session in a production environment must gate high-risk commands in non-interactive mode without --confirm")
	}
	if !strings.Contains(err.Error(), "production write requires --confirm in non-interactive mode") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// An OAuth config written before environment-type persistence (no
// EnvironmentType) fails safe to production so the gate still engages,
// mirroring how legacy keys are treated as production-scoped.
func TestProductionGate_HighRisk_OAuthUnknownEnv_FailsSafeToGated(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		OAuth:         &oauthTokens{AccessToken: "oauth-access-token"},
		EnvironmentID: "env_legacy123",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newSafetyTestCmd(t, classHighRisk, "")
	err := productionGate(cmd, &fakeDecider{interactive: false})
	if err == nil {
		t.Fatal("OAuth session with an unknown environment type must fail safe to gated")
	}
	if !strings.Contains(err.Error(), "production write requires --confirm in non-interactive mode") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// An OAuth session in an explicitly non-production environment is never
// gated — it must not spuriously prompt staging/test users.
func TestProductionGate_HighRisk_OAuthNonProductionEnv_NeverGated(t *testing.T) {
	isolateHome(t)
	if err := saveConfig(retabConfig{
		OAuth:           &oauthTokens{AccessToken: "oauth-access-token"},
		EnvironmentID:   "env_staging123",
		EnvironmentType: "non_production",
	}); err != nil {
		t.Fatal(err)
	}
	cmd := newSafetyTestCmd(t, classHighRisk, "")
	dec := &fakeDecider{interactive: false}
	if err := productionGate(cmd, dec); err != nil {
		t.Fatalf("OAuth session in a non-production environment must never be gated, got: %v", err)
	}
	if dec.prompted {
		t.Fatal("non-production OAuth session must never prompt")
	}
}

// --- test environment is never gated ------------------------------------

func TestProductionGate_HighRisk_TestEnvironment_NeverGated(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classHighRisk, "rt_test_secret")
	dec := &fakeDecider{interactive: false}
	if err := productionGate(cmd, dec); err != nil {
		t.Fatalf("test-environment high-risk command must never be gated, got: %v", err)
	}
	if dec.prompted {
		t.Fatal("test environment must never prompt")
	}
}

// --- read-only is never gated -------------------------------------------

func TestProductionGate_ReadOnly_Production_NeverGated(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classReadOnly, "rt_live_secret")
	dec := &fakeDecider{interactive: false}
	if err := productionGate(cmd, dec); err != nil {
		t.Fatalf("read-only production command must never be gated, got: %v", err)
	}
	if dec.prompted {
		t.Fatal("read-only command must never prompt")
	}
}

func TestProductionGate_NormalWrite_Production_NeverGated(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classNormalWrite, "rt_live_secret")
	dec := &fakeDecider{interactive: false}
	if err := productionGate(cmd, dec); err != nil {
		t.Fatalf("normal-write production command must never be gated, got: %v", err)
	}
	if dec.prompted {
		t.Fatal("normal-write command must never prompt")
	}
}

// --- interactive prompt -------------------------------------------------

func TestProductionGate_HighRisk_Interactive_CorrectAnswer_Allowed(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classHighRisk, "rt_live_secret")
	dec := &fakeDecider{interactive: true, answer: "production"}
	if err := productionGate(cmd, dec); err != nil {
		t.Fatalf("typing \"production\" must allow the command, got: %v", err)
	}
	if !dec.prompted {
		t.Fatal("interactive high-risk production command must prompt")
	}
}

func TestProductionGate_HighRisk_Interactive_WrongAnswer_Aborts(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classHighRisk, "rt_live_secret")
	dec := &fakeDecider{interactive: true, answer: "yes"}
	err := productionGate(cmd, dec)
	if err == nil {
		t.Fatal("a non-matching confirmation answer must abort the command")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestProductionGate_HighRisk_Interactive_WithConfirm_SkipsPrompt(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classHighRisk, "rt_live_secret")
	setConfirm(t, cmd, true)
	dec := &fakeDecider{interactive: true, answer: "production"}
	if err := productionGate(cmd, dec); err != nil {
		t.Fatalf("--confirm must allow the command without error, got: %v", err)
	}
	if dec.prompted {
		t.Fatal("--confirm must skip the prompt even on an interactive TTY")
	}
}

// --- unclassified command defaults to normal-write ----------------------

func TestSafetyClassOf_DefaultsToNormalWrite(t *testing.T) {
	cmd := &cobra.Command{Use: "mystery"}
	if got := safetyClassOf(cmd); got != classNormalWrite {
		t.Fatalf("unmarked command should default to %q, got %q", classNormalWrite, got)
	}
}

func TestProductionGate_Unclassified_Production_NeverGated(t *testing.T) {
	isolateHome(t)
	root := &cobra.Command{Use: "retab"}
	root.PersistentFlags().String("api-key", "", "")
	root.PersistentFlags().String("base-url", "", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.PersistentFlags().Bool("live", false, "")
	root.PersistentFlags().String("env", "", "")
	child := &cobra.Command{Use: "mystery", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(child)
	if err := root.PersistentFlags().Set("api-key", "rt_live_secret"); err != nil {
		t.Fatalf("set --api-key: %v", err)
	}
	// child carries no safety annotation -> defaults to normal-write.
	if err := productionGate(child, &fakeDecider{interactive: false}); err != nil {
		t.Fatalf("an unclassified command must never be gated, got: %v", err)
	}
}

// --- prompt content -----------------------------------------------------

func TestProductionConfirmPrompt_RedactsCredentialAndNamesCommand(t *testing.T) {
	isolateHome(t)
	cmd := newSafetyTestCmd(t, classHighRisk, "rt_live_supersecretkey")
	cred, err := resolveCredential(cmd)
	if err != nil {
		t.Fatalf("resolveCredential: %v", err)
	}
	prompt := productionConfirmPrompt(cmd, cred)
	if strings.Contains(prompt, "rt_live_supersecretkey") {
		t.Fatal("prompt must never contain the full API key")
	}
	if !strings.Contains(prompt, "environment: production") {
		t.Fatalf("prompt must name the production environment: %q", prompt)
	}
	if !strings.Contains(prompt, "publish") {
		t.Fatalf("prompt must echo the command: %q", prompt)
	}
	if !strings.Contains(prompt, `Type "production" to continue`) {
		t.Fatalf("prompt must ask the user to type production: %q", prompt)
	}
}

// --- classification wiring ----------------------------------------------

func TestClassification_KnownCommands(t *testing.T) {
	if got := safetyClassOf(workflowsPublishCmd); got != classHighRisk {
		t.Errorf("workflows publish should be high-risk, got %q", got)
	}
	if got := safetyClassOf(workflowsDeleteCmd); got != classHighRisk {
		t.Errorf("workflows delete should be high-risk, got %q", got)
	}
	if got := safetyClassOf(workflowsRunsCreateCmd); got != classHighRisk {
		t.Errorf("workflows runs create should be high-risk, got %q", got)
	}
	if got := safetyClassOf(workflowsListCmd); got != classReadOnly {
		t.Errorf("workflows list should be read-only, got %q", got)
	}
	if got := safetyClassOf(workflowsGetCmd); got != classReadOnly {
		t.Errorf("workflows get should be read-only, got %q", got)
	}
	if got := safetyClassOf(authStatusCmd); got != classReadOnly {
		t.Errorf("auth status should be read-only, got %q", got)
	}
	if got := safetyClassOf(envListCmd); got != classReadOnly {
		t.Errorf("env list should be read-only, got %q", got)
	}
	// create-draft / update-draft / upload stay at the normal-write default.
	if got := safetyClassOf(workflowsCreateCmd); got != classNormalWrite {
		t.Errorf("workflows create should be normal-write, got %q", got)
	}
	if got := safetyClassOf(workflowsUpdateCmd); got != classNormalWrite {
		t.Errorf("workflows update should be normal-write, got %q", got)
	}
	if got := safetyClassOf(filesUploadCmd); got != classNormalWrite {
		t.Errorf("files upload should be normal-write, got %q", got)
	}
}
