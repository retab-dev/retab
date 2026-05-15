package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// workflows spec — declarative (YAML) management for workflows.
//
// Mirrors the four methods on `WorkflowSpecsService` in the Go SDK:
// validate, plan, apply (POST a YAML body), and export (GET an existing
// workflow back to YAML by id). Aimed at the Terraform-style flow:
// edit YAML in your repo, plan, apply, commit.

var workflowsSpecCmd = &cobra.Command{
	Use:   "spec",
	Short: "Manage workflows declaratively from YAML",
	Long: `Validate, plan, apply, and export YAML workflow specs.

A spec is a single YAML file describing a workflow's blocks, edges, and
metadata. The four verbs form a Terraform-style loop:

  validate   parse + type-check the spec, no server changes
  plan       diff the spec against the live workflow without applying
  apply      create or update the workflow from the spec
  export     dump a live workflow's spec back to YAML

The three POST verbs read YAML from a file path, or from stdin when the
path is "-". Output is JSON on stdout, suitable for piping into ` + "`jq`" + `.`,
	Example: `  # Round-trip a workflow through git
  retab workflows spec export wf_abc123 > workflow.yaml
  $EDITOR workflow.yaml
  retab workflows spec validate workflow.yaml
  retab workflows spec plan     workflow.yaml
  retab workflows spec apply    workflow.yaml

  # Tail of a pipe
  cat workflow.yaml | retab workflows spec apply -`,
}

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

var workflowsSpecValidateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Parse + type-check a YAML spec without touching the server",
	Long: `Send the YAML to the server's spec validator. No workflow is
created or modified — the server only returns whether the spec parses
and whether its referenced blocks / schemas / model ids exist.

Returns a non-zero exit if validation fails, with the server's diagnostic
JSON on stderr — wire it into pre-commit / CI to keep broken specs out
of main.`,
	Example: `  retab workflows spec validate ./workflow.yaml
  cat workflow.yaml | retab workflows spec validate -`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		yaml, err := readSpecYAML(args[0])
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Specs.Validate(ctx, yaml)
		if err != nil {
			return translateSpecAPIError(err)
		}
		return printJSON(result)
	}),
}

var workflowsSpecPlanCmd = &cobra.Command{
	Use:   "plan <path>",
	Short: "Diff a YAML spec against the live workflow without applying",
	Long: `Compute what would change if the spec were applied: which
blocks would be created, updated, deleted; which edges would be re-wired.

Plan is read-only — safe to run on production specs. Pair it with
` + "`apply`" + ` for a Terraform-style review-then-apply loop.`,
	Example: `  retab workflows spec plan ./workflow.yaml
  cat workflow.yaml | retab workflows spec plan - | jq .changes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		yaml, err := readSpecYAML(args[0])
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Specs.Plan(ctx, yaml)
		if err != nil {
			return translateSpecAPIError(err)
		}
		return printJSON(result)
	}),
}

var workflowsSpecApplyCmd = &cobra.Command{
	Use:   "apply <path>",
	Short: "Create or update a workflow from a YAML spec",
	Long: `Apply the YAML spec: create the workflow if it doesn't exist,
or update it (block + edge diff) if it does. The workflow's id is read
from the spec body, so the same file always targets the same workflow.

Mutating — gate behind ` + "`plan`" + ` in CI if the workflow is in production.`,
	Example: `  retab workflows spec apply ./workflow.yaml
  cat workflow.yaml | retab workflows spec apply -`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		yaml, err := readSpecYAML(args[0])
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Specs.Apply(ctx, yaml)
		if err != nil {
			return translateSpecAPIError(err)
		}
		return printJSON(result)
	}),
}

var workflowsSpecExportCmd = &cobra.Command{
	Use:   "export <workflow-id>",
	Short: "Dump a live workflow's spec back to YAML",
	Long: `Fetch the YAML spec representing the workflow's current
state. Useful for round-tripping a workflow created in the dashboard
back into a git-managed YAML file, or for diffing two environments.

By default, the raw YAML body is written to stdout so the command can
be redirected straight into a file. Pass ` + "`--format json`" + ` to see
the full server response object (handy for piping into ` + "`jq`" + ` to
pull out other fields).`,
	Example: `  retab workflows spec export wf_abc123 > workflow.yaml
  retab workflows spec export wf_abc123 --format json | jq .`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return err
		}
		// `--json` is the shorthand for `--format json`. When both are
		// passed and they disagree, the explicit `--format` value wins
		// (`--json` is the cheap-to-type alias, `--format` is the strict
		// version). Both being set to compatible values is a no-op.
		if jsonShortcut, _ := cmd.Flags().GetBool("json"); jsonShortcut && format == "yaml" {
			format = "json"
		}
		if format != "yaml" && format != "json" {
			return fmt.Errorf("invalid --format value %q (want: yaml | json)", format)
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Specs.Export(ctx, args[0])
		if err != nil {
			return err
		}
		return writeSpecExport(os.Stdout, result, format)
	}),
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
	return fmt.Errorf(`spec: metadata.id is required.
  For new workflows, use any unique identifier (e.g. metadata.id: wrk_my-pipeline).
  For existing workflows, use the id returned by ` + "`retab workflows list`" + `.
  Use --debug to see the full server response.`)
}

func init() {
	workflowsSpecExportCmd.Flags().String("format", "yaml", "output format: yaml | json")
	workflowsSpecExportCmd.Flags().Bool("json", false, "shorthand for --format json")

	workflowsSpecCmd.AddCommand(
		workflowsSpecValidateCmd,
		workflowsSpecPlanCmd,
		workflowsSpecApplyCmd,
		workflowsSpecExportCmd,
	)
	workflowsCmd.AddCommand(workflowsSpecCmd)
}
