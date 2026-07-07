package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// safetyClass classifies a command by how dangerous it is to run against
// the customer's production environment. The blueprint's "Command Safety
// Classes" / "Production Safety" tables define six classes; the gate only
// distinguishes "needs a production confirmation" from "does not", so the
// classes collapse to three CLI-relevant buckets:
//
//   - readOnly      — list/get/view/status/diagnose/export. Never gated.
//   - normalWrite   — create-draft, update-draft, upload. Never gated.
//     This is the default for an unclassified command, so
//     a command left unmarked never accidentally prompts.
//   - highRisk      — runtime side effect, external side effect,
//     destructive, release/promotion. Gated when the
//     resolved environment is production.
type safetyClass string

const (
	classReadOnly    safetyClass = "read-only"
	classNormalWrite safetyClass = "normal-write"
	classHighRisk    safetyClass = "high-risk"
)

// safetyAnnotationKey is the cobra Annotations map key under which a
// command's safety class is stored. Using Annotations (rather than a
// parallel registry) keeps the classification attached to the command
// object itself and survives the command-tree walk in init().
const safetyAnnotationKey = "retab.safetyClass"

// confirmFlagName is the name of the per-command production pre-approval
// flag. It is registered locally on high-risk commands (addConfirmFlag),
// never globally, so it only appears in help where it actually does
// something.
const confirmFlagName = "confirm"

// addConfirmFlag registers the local --confirm flag on a high-risk command
// so it shows up in that command's own help and can pre-approve a
// production mutation in CI. It is idempotent: a command that already
// declares the flag is left untouched.
func addConfirmFlag(cmd *cobra.Command) {
	if cmd.Flags().Lookup(confirmFlagName) != nil {
		return
	}
	cmd.Flags().Bool(confirmFlagName, false, "pre-approve this production-mutating command (skips the confirmation prompt)")
}

// markSafety records cls on cmd's Annotations map. Commands call this from
// their package init() (see the per-resource classification at the bottom
// of each resource file's init, wired through classifyCommands below).
func markSafety(cmd *cobra.Command, cls safetyClass) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[safetyAnnotationKey] = string(cls)
}

// safetyClassOf returns the classification recorded for cmd. An unmarked
// command defaults to normalWrite — never highRisk — so forgetting to
// classify a command can only ever weaken the gate, not spuriously block a
// user. Genuinely destructive commands are explicitly marked highRisk.
func safetyClassOf(cmd *cobra.Command) safetyClass {
	if cmd.Annotations != nil {
		if v, ok := cmd.Annotations[safetyAnnotationKey]; ok {
			return safetyClass(v)
		}
	}
	return classNormalWrite
}

// confirmDecider abstracts the two pieces of environment the production
// gate needs from the outside world: whether the session is interactive
// (a real TTY) and, when it is, what the user typed at the prompt. Making
// this an interface lets unit tests exercise the gate without a real
// terminal — see safety_test.go.
type confirmDecider interface {
	// Interactive reports whether the CLI is attached to a TTY and may
	// therefore prompt the user.
	Interactive() bool

	// ReadConfirmation prints prompt and returns the line the user typed.
	// Only called when Interactive() is true.
	ReadConfirmation(prompt string) (string, error)
}

// ttyConfirmDecider is the production confirmDecider: it inspects stdin/
// stdout for a real terminal and reads a line from stdin when prompting.
type ttyConfirmDecider struct{}

func (ttyConfirmDecider) Interactive() bool {
	// Both stdin (to read the answer) and stderr (to show the prompt)
	// must be terminals. A piped stdin with a TTY stderr would hang on
	// ReadConfirmation; treat it as non-interactive instead.
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stderr.Fd()))
}

func (ttyConfirmDecider) ReadConfirmation(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// productionGate is the single reusable confirmation gate. It is applied
// centrally (see applySafetyGate) to every command's RunE, so behaviour is
// identical across the whole CLI.
//
// Behaviour:
//
//   - Resolved environment is not production -> never gated.
//   - Command class is read-only or normal-write -> never gated.
//   - --confirm passed -> gate passes without prompting (CI path).
//   - Interactive TTY -> print the blueprint prompt; require the user to
//     type exactly "production".
//   - Non-interactive / piped -> do NOT prompt; fail with the blueprint's
//     "production write requires --confirm in non-interactive mode" error.
//
// resolveErr is tolerated: if the credential cannot be resolved the gate
// stays out of the way and lets the command's own RunE surface the real
// credential error.
func productionGate(cmd *cobra.Command, decider confirmDecider) error {
	cls := safetyClassOf(cmd)
	if cls != classHighRisk {
		return nil
	}

	cred, baseURL, cfg, err := resolveCLIRequest(cmd)
	if err != nil {
		// Let the command body report the credential failure itself.
		return nil
	}
	if expectedEnvironmentForSafety(cmd, cfg, cred, baseURL) != slugProduction {
		return nil
	}

	confirmed, _ := cmd.Flags().GetBool(confirmFlagName)
	if confirmed {
		return nil
	}

	if !decider.Interactive() {
		return fmt.Errorf("production write requires --confirm in non-interactive mode")
	}

	prompt := productionConfirmPrompt(cmd, cred)
	answer, err := decider.ReadConfirmation(prompt)
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	if answer != slugProduction {
		return fmt.Errorf("aborted: confirmation did not match %q", slugProduction)
	}
	return nil
}

func expectedEnvironmentForSafety(cmd *cobra.Command, cfg retabConfig, cred resolvedCredential, baseURL string) string {
	expected := cred.ExpectedEnvironment
	if cred.OAuth == nil {
		return expected
	}

	// An explicit --environment-id flag / RETAB_ENVIRONMENT_ID env var routes
	// the request regardless of the persisted session environment, so the gate
	// must judge the override — in both directions: a production session
	// pointed at a non-production environment disengages the gate, and a
	// non-production session pointed at production (or an unprovable
	// environment) engages it.
	environmentID, source := selectedEnvironmentIDWithSource(cmd, cfg)
	if !isExplicitEnvironmentSelector(source) || strings.TrimSpace(environmentID) == "" {
		return expected
	}

	environmentType, ok := lookupEnvironmentTypeForSafety(cmd, cred, baseURL, environmentID)
	if ok {
		if environmentType == cliEnvironmentTypeNonProduction {
			return ""
		}
		// The override provably points at production: gate, regardless of
		// the persisted session environment.
		return slugProduction
	}
	// Lookup failed (network blip, token scope, ...): fall back to the
	// session's own expectation. A production session stays gated
	// (fail safe); a non-production session is not hard-blocked in CI by
	// a transient lookup error.
	return expected
}

func isExplicitEnvironmentSelector(source string) bool {
	return source == "--environment-id flag" || source == "RETAB_ENVIRONMENT_ID env"
}

func lookupEnvironmentTypeForSafety(cmd *cobra.Command, cred resolvedCredential, baseURL string, environmentID string) (cliEnvironmentType, bool) {
	if cred.OAuth == nil {
		return "", false
	}
	ctx, cancel := ctxFor(cmd)
	defer cancel()

	rawToken, err := makeOAuthTokenProvider(cred.OAuth)(ctx)
	if err != nil {
		return "", false
	}
	environment, err := fetchEnvironmentForSafety(ctx, baseURL, rawToken, environmentID)
	if err != nil {
		return "", false
	}
	return environment.Type, true
}

func fetchEnvironmentForSafety(ctx context.Context, baseURL string, rawOAuthToken string, environmentID string) (*cliEnvironment, error) {
	if strings.TrimSpace(rawOAuthToken) == "" {
		return nil, fmt.Errorf("OAuth access token is empty")
	}
	if strings.TrimSpace(environmentID) == "" {
		return nil, fmt.Errorf("environment id is required")
	}
	requestURL, err := buildCLIRequestURL(baseURL, "/v1/environments/"+url.PathEscape(environmentID), nil)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+rawOAuthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("environment lookup failed: %s", resp.Status)
	}
	var environment cliEnvironment
	if err := json.NewDecoder(resp.Body).Decode(&environment); err != nil {
		return nil, err
	}
	return &environment, nil
}

// productionConfirmPrompt renders the blueprint's confirmation prompt: it
// names the environment, shows the redacted credential, echoes the command
// path, and asks the user to type "production". The credential is always
// redacted via resolvedCredential.KeyPreview — full keys are never printed.
func productionConfirmPrompt(cmd *cobra.Command, cred resolvedCredential) string {
	credLine := cred.KeyPreview()
	if credLine == "" {
		credLine = string(cred.Source)
	}
	var b strings.Builder
	b.WriteString("You are about to modify Retab production.\n\n")
	fmt.Fprintf(&b, "environment: %s\n", slugProduction)
	fmt.Fprintf(&b, "credential: %s\n", credLine)
	fmt.Fprintf(&b, "command: %s\n\n", commandPathWithoutRoot(cmd))
	b.WriteString(`Type "production" to continue: `)
	return b.String()
}

// commandPathWithoutRoot returns the command invocation path with the
// leading "retab " stripped, e.g. "workflows publish". Used in the prompt
// so it reads as the command the user typed.
func commandPathWithoutRoot(cmd *cobra.Command) string {
	path := cmd.CommandPath()
	root := cmd.Root().Name() + " "
	return strings.TrimPrefix(path, root)
}

// applySafetyGate wraps every runnable command's RunE so the production
// gate runs before the command body. It is a tree walk in the same spirit
// as hardenGroupCommands / hardenLeafArgs, so the gate is enforced
// consistently without every command opting in by hand.
//
// Read-only and normal-write commands pass through productionGate as a
// no-op, so the wrap is cheap and uniform.
func applySafetyGate(c *cobra.Command) {
	for _, sub := range c.Commands() {
		applySafetyGate(sub)
	}
	if c.RunE == nil {
		return
	}
	inner := c.RunE
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if err := productionGate(cmd, ttyConfirmDecider{}); err != nil {
			// runE-style rendering: the gate error is a plain CLI error.
			fmt.Fprintln(os.Stderr, "error: "+err.Error())
			return renderedError{err: err}
		}
		return inner(cmd, args)
	}
}
