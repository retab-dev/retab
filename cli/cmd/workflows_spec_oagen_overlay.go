//go:build retab_oagen_cli_workflows_spec

package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// readSpecYAML reads the YAML body for validate/plan/apply. The server
// is the source of truth on YAML schema, so we don't pre-parse here —
// any bytes go through, including comments and unrecognised fields.
// Path "-" reads from stdin so the command works as the tail of a pipe.
func readSpecYAML(path string) (string, error) {
	var raw []byte
	var err error
	if path == "-" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		return "", fmt.Errorf("empty spec input")
	}
	return string(raw), nil
}

// confirmDestructiveApply inspects a plan response and gates the
// subsequent apply when one or more resources would be destroyed.
//
// Three branches, matching `confirmDestructive` for delete:
//
//	--yes               → proceed unconditionally, no prompt.
//	TTY stdin           → list the doomed resources to stderr, prompt
//	                      "Apply this change? [y/N] ". Accept y/yes
//	                      only; anything else aborts.
//	non-TTY, no --yes   → refuse outright. A typo in metadata.id that
//	                      points at an unrelated workflow could
//	                      otherwise wipe its blocks/edges silently;
//	                      scripts must opt in with --yes.
//
// When the plan's destroy count is zero, this is a no-op — the apply
// proceeds immediately so the non-destructive happy path stays
// byte-identical to the pre-guard behaviour.
func confirmDestructiveApply(cmd *cobra.Command, plan *retab.Resource) error {
	destroy, doomed := planDestroyCountAndResources(plan)
	if destroy == 0 {
		return nil
	}
	if yes, _ := cmd.Flags().GetBool("yes"); yes {
		return nil
	}
	stdin, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(stdin.Fd())) {
		return fmt.Errorf(
			"refusing to apply spec that would destroy %d resources without --yes (stdin is not a terminal)",
			destroy,
		)
	}
	out := cmd.ErrOrStderr()
	if _, err := fmt.Fprintf(out, "This apply will destroy %d resource(s):\n", destroy); err != nil {
		return err
	}
	for _, r := range doomed {
		if _, err := fmt.Fprintf(out, "  - %s\n", r); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(out, "Apply this change? [y/N] "); err != nil {
		return err
	}
	answer, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("aborted: spec apply not run")
	}
}

// planDestroyCountAndResources extracts the destroy total and the list
// of resource addresses scheduled for deletion from a plan response.
// Returns (0, nil) when the response is missing the expected shape so
// the caller treats absence-of-evidence as "nothing to gate on" — the
// server-side validation path already failed loudly upstream if the
// response is malformed.
func planDestroyCountAndResources(plan *retab.Resource) (int, []string) {
	if plan == nil {
		return 0, nil
	}
	summary, _ := (*plan)["summary"].(map[string]any)
	destroy := 0
	if raw, ok := summary["destroy"]; ok {
		switch n := raw.(type) {
		case float64:
			destroy = int(n)
		case int:
			destroy = n
		}
	}
	if destroy == 0 {
		return 0, nil
	}
	changes, _ := (*plan)["resource_changes"].([]any)
	var doomed []string
	for _, raw := range changes {
		change, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		actions, _ := change["actions"].([]any)
		hasDelete := false
		for _, a := range actions {
			if s, _ := a.(string); s == "delete" {
				hasDelete = true
				break
			}
		}
		if !hasDelete {
			continue
		}
		address, _ := change["address"].(string)
		target, _ := change["target"].(string)
		name, _ := change["name"].(string)
		label := address
		switch {
		case label != "" && target != "":
			label = fmt.Sprintf("%s (%s)", address, target)
		case label == "" && name != "":
			label = name
		case label == "":
			label = "<unknown resource>"
		}
		doomed = append(doomed, label)
	}
	return destroy, doomed
}

// writeSpecExport renders the export response onto w in the requested
// format. Pulled out of RunE so the format dispatch — which is the only
// non-trivial branch in the command — is testable without a live server
// or a stubbed cobra invocation.
//
// "yaml" (the default) extracts result["yaml_definition"] and writes it
// as bare text terminated by exactly one trailing newline, so a shell
// redirect produces a clean YAML file. A missing or empty field is an
// error rather than an empty file: the user asked for a workflow's YAML
// and got nothing back, that's worth surfacing.
//
// "json" preserves the legacy behaviour — the full Resource map pretty-
// printed with the same settings as printJSON. This keeps the response
// envelope available for power users who want to read other fields
// (e.g. `workflow_id`) with `jq`.
func writeSpecExport(w io.Writer, result *retab.Resource, format string) error {
	switch format {
	case "yaml":
		if result == nil {
			return fmt.Errorf("server response missing yaml_definition field")
		}
		raw, ok := (*result)["yaml_definition"]
		if !ok {
			return fmt.Errorf("server response missing yaml_definition field")
		}
		yaml, ok := raw.(string)
		if !ok || yaml == "" {
			return fmt.Errorf("server response missing yaml_definition field")
		}
		// Guarantee exactly one trailing newline regardless of whether the
		// server already included one — a redirected `> workflow.yaml`
		// should end cleanly, and a doubled newline would dirty the diff
		// on the very next round-trip.
		for len(yaml) > 0 && yaml[len(yaml)-1] == '\n' {
			yaml = yaml[:len(yaml)-1]
		}
		_, err := fmt.Fprintln(w, yaml)
		return err
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(result)
	default:
		return fmt.Errorf("invalid --format value %q (want: yaml | json)", format)
	}
}

func failIfSpecValidationInvalid(result *retab.Resource) error {
	if result == nil {
		return nil
	}
	if !specValidationIsInvalid(result) {
		return nil
	}
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(result); err != nil {
		return err
	}
	return fmt.Errorf("validation failed")
}

func specValidationIsInvalid(result *retab.Resource) bool {
	if isValid, ok := (*result)["is_valid"].(bool); ok {
		return !isValid
	}
	diagnostics, ok := (*result)["diagnostics"].(map[string]any)
	if !ok {
		return false
	}
	isValid, ok := diagnostics["is_valid"].(bool)
	return ok && !isValid
}

// validationResponseAsResource round-trips a typed validation response
// through JSON to produce an untyped map shape. The validation and
// destructive-confirmation helpers walk the response by string keys
// (`is_valid`, `diagnostics`, `summary`, `resource_changes`) so they keep
// working with whatever the server actually returns — including new
// fields the typed struct doesn't know about yet.
func validationResponseAsResource(r *retab.DeclarativeValidationResponse) *retab.Resource {
	return typedResponseAsResource(r)
}

func planResponseAsResource(r *retab.DeclarativePlanResponse) *retab.Resource {
	return typedResponseAsResource(r)
}

func applyResponseAsResource(r *retab.DeclarativeApplyResponse) *retab.Resource {
	return typedResponseAsResource(r)
}

func exportResponseAsResource(r *retab.DeclarativeExportResponse) *retab.Resource {
	return typedResponseAsResource(r)
}

func typedResponseAsResource(v any) *retab.Resource {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	out := retab.Resource{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return &out
}

// translateSpecAPIError catches the most common failure mode of the spec
// validate/plan/apply surface — a YAML body without `metadata.id` — and
// rewrites the server's giant pydantic blob into a single actionable
// sentence. The original error is still surfaced by `--debug` (which dumps
// the full HTTP trace) so we don't hide debugging detail.
//
// The marker is the substring `"yaml_path":"metadata.id"` returned in the
// pydantic error envelope when the field is missing. Matching on this
// string is intentional: it's stable, it's what the server emits today,
// and switching to structural unmarshalling would couple us to the exact
// shape of the server's error wrapper (which has churned).
//
// Any error that doesn't match this pattern is passed through unchanged.
func translateSpecAPIError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if !strings.Contains(msg, `"yaml_path":"metadata.id"`) {
		return err
	}
	if !strings.Contains(msg, `"type":"missing"`) {
		return err
	}
	return fmt.Errorf("spec: metadata.id is required; for new workflows, use any unique identifier (e.g. metadata.id: wrk_my-pipeline); for existing workflows, use the id returned by `retab workflows list`; use --debug to see the full server response")
}
