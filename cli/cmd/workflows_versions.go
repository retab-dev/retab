//go:build !retab_oagen_cli_workflows

package cmd

import (
	"fmt"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsVersionsCmd = &cobra.Command{
	Use:   "versions <workflow-id>",
	Short: "List workflow versions",
	Long: `List immutable versions for a workflow.

Use this to find version ids for diffing, inspection, or restoring a prior
published graph back into the draft.`,
	Example: `  retab workflows versions wf_abc123
  retab workflows versions wf_abc123 --limit 20`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		params := &retab.WorkflowsListVersionsParams{WorkflowID: args[0]}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			params.Limit = ptr(limit)
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.ListVersions(ctx, params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsDiffCmd = &cobra.Command{
	Use:   "diff <workflow-id> [from-version-id] [to-version-id]",
	Short: "Diff workflow versions",
	Long: `Diff two immutable workflow graph versions for the same workflow.

The two version ids can be passed co-equally either positionally
(` + "`diff <workflow-id> <from-version-id> <to-version-id>`" + `) or via the
` + "`--from-version-id` / `--to-version-id`" + ` flags. Mixing the two forms is
allowed as long as a positional and its matching flag agree.

Pass version ids from ` + "`retab workflows versions <workflow-id>`" + `.`,
	Example: `  retab workflows diff wf_abc123 wfv_old wfv_new
  retab workflows diff wf_abc123 \
    --from-version-id wfv_old --to-version-id wfv_new`,
	Args: func(cmd *cobra.Command, args []string) error {
		switch len(args) {
		case 1, 3:
			return nil
		case 2:
			return fmt.Errorf("the positional form needs both <from-version-id> and <to-version-id>: pass all three (\"diff <wf> <from> <to>\") or use --from-version-id/--to-version-id")
		default:
			return fmt.Errorf("accepts the workflow id plus an optional <from-version-id> <to-version-id> pair (1 or 3 args), received %d", len(args))
		}
	},
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		fromFlag, _ := cmd.Flags().GetString("from-version-id")
		toFlag, _ := cmd.Flags().GetString("to-version-id")
		fromVersionID, toVersionID, err := resolveDiffVersions(args, fromFlag, toFlag)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.ListDiff(ctx, &retab.WorkflowsListDiffParams{
			WorkflowID:            args[0],
			FromWorkflowVersionID: fromVersionID,
			ToWorkflowVersionID:   toVersionID,
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionCmd = &cobra.Command{
	Use:   "version <workflow-id> <workflow-version-id>",
	Short: "Get a workflow version",
	Long: `Fetch one immutable workflow graph version.

The workflow id disambiguates content-addressed version ids that may be reused
across workflows.`,
	Example: `  retab workflows version wf_abc123 wfv_456`,
	Args:    cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.GetVersion(ctx, args[1], &retab.WorkflowsGetVersionParams{
			WorkflowID: args[0],
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsVersionRestoreCmd = &cobra.Command{
	Use:   "version-restore <workflow-id> [workflow-version-id]",
	Short: "Restore a workflow version",
	Long: `Restore an immutable workflow graph version into the workflow draft.

The version id can be passed co-equally either positionally
(` + "`version-restore <workflow-id> <workflow-version-id>`" + `) or via the
` + "`--version-id`" + ` flag.

This mutates the draft graph. The restored version remains immutable; only the
current draft is replaced.`,
	Example: `  retab workflows version-restore wf_abc123 wfv_456 --yes
  retab workflows version-restore wf_abc123 --version-id wfv_456 --yes`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		versionFlag, _ := cmd.Flags().GetString("version-id")
		versionID, err := resolveRestoreVersionID(args, versionFlag)
		if err != nil {
			return err
		}
		if err := confirmDestructive(cmd, "workflow draft", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.CreateVersionRestore(ctx, versionID, &retab.WorkflowsCreateVersionRestoreParams{
			WorkflowID: args[0],
		})
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

// resolveDiffVersions implements the uniform positional/flag contract for
// `workflows diff`. The two version ids may be supplied co-equally either
// positionally (`diff <workflow-id> <from> <to>`) or via the
// `--from-version-id` / `--to-version-id` flags. Args validation already
// guarantees len(args) is 1 (flags supply both) or 3 (positionals supply both),
// so each id has at most one positional source plus its flag. When both a
// positional and its flag are present they must agree; identical values are
// allowed for symmetry with callers that pass both forms, differing values are
// rejected naming both forms. A wholly missing id errors mentioning both forms.
func resolveDiffVersions(args []string, fromFlag, toFlag string) (from, to string, err error) {
	var fromPos, toPos string
	if len(args) == 3 {
		fromPos = strings.TrimSpace(args[1])
		toPos = strings.TrimSpace(args[2])
	}
	fromFlag = strings.TrimSpace(fromFlag)
	toFlag = strings.TrimSpace(toFlag)

	from, err = pickVersion("from", fromPos, fromFlag, "--from-version-id")
	if err != nil {
		return "", "", err
	}
	to, err = pickVersion("to", toPos, toFlag, "--to-version-id")
	if err != nil {
		return "", "", err
	}
	return from, to, nil
}

// pickVersion reconciles a single version id supplied positionally and/or via a
// flag for the `diff` command. label is "from"/"to" and flagName is the user-
// facing flag (e.g. "--from-version-id").
func pickVersion(label, positional, flag, flagName string) (string, error) {
	switch {
	case positional != "" && flag != "" && positional != flag:
		return "", fmt.Errorf("%s version id supplied twice and they differ: positional %q vs %s %q", label, positional, flagName, flag)
	case positional != "":
		return positional, nil
	case flag != "":
		return flag, nil
	default:
		return "", fmt.Errorf("%s version id required: pass it positionally (\"diff <wf> <from> <to>\") or via %s", label, flagName)
	}
}

// resolveRestoreVersionID implements the uniform positional/flag contract for
// `workflows version-restore`. The version id may be supplied co-equally either
// as the second positional or via the `--version-id` flag. Args validation
// guarantees len(args) is 1 or 2. When both are present they must agree;
// identical values are allowed, differing values are rejected naming both
// forms. A wholly missing id errors mentioning both forms.
func resolveRestoreVersionID(args []string, versionFlag string) (string, error) {
	var positional string
	if len(args) == 2 {
		positional = strings.TrimSpace(args[1])
	}
	versionFlag = strings.TrimSpace(versionFlag)
	switch {
	case positional != "" && versionFlag != "" && positional != versionFlag:
		return "", fmt.Errorf("version id supplied twice and they differ: positional %q vs --version-id %q", positional, versionFlag)
	case positional != "":
		return positional, nil
	case versionFlag != "":
		return versionFlag, nil
	default:
		return "", fmt.Errorf("version id required: pass it positionally or via --version-id")
	}
}

func init() {
	workflowsVersionsCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 100}, "limit", "max items to return (1-100)")

	workflowsDiffCmd.Flags().String("from-version-id", "", "base workflow version id (alternative to the positional form)")
	workflowsDiffCmd.Flags().String("to-version-id", "", "target workflow version id (alternative to the positional form)")

	workflowsVersionRestoreCmd.Flags().String("version-id", "", "workflow version id to restore (alternative to the positional form)")
	workflowsVersionRestoreCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsCmd.AddCommand(workflowsVersionsCmd, workflowsDiffCmd, workflowsVersionCmd, workflowsVersionRestoreCmd)
}
