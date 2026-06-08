package cmd

import "github.com/spf13/cobra"

// This file is the single place where every CLI command is assigned a
// safety class and where the production confirmation gate is wired onto
// the command tree.
//
// It is named `zz_` so its init() runs last — after every resource file's
// init() has finished building the command tree with AddCommand. Two
// things must be true before the gate is applied:
//
//  1. All commands exist and are attached to rootCmd (resource inits done).
//  2. Each command carries its safety annotation (classifyCommands below).
//
// applySafetyGate then wraps every command's RunE so productionGate runs
// first. See safety.go for the gate mechanism itself.
//
// Classification policy (from the blueprint's "Production Safety" table):
//
//   - read-only   : list, get, view, status, diagnose, export, snapshots,
//                   config, execution-order, document-url, eligible-blocks,
//                   results/metrics get/list, download-link, env list,
//                   auth status. Never prompts.
//   - normal-write: create-draft, update-draft, upload, duplicate, create
//                   test/experiment, batch block/edge edits. Never prompts.
//                   This is also the default for any unmarked command.
//   - high-risk   : runtime side effect (run workflow, restart/cancel run,
//                   resume/approve HIL, retry job), external side effect,
//                   destructive (delete, delete-all), and release/
//                   promotion (publish). Prompts / requires --confirm
//                   against production.
//
// The lists below deliberately enumerate only read-only and high-risk
// commands; anything else is left at the normalWrite default. That way a
// newly added command can never silently become high-risk, and a missed
// read-only classification only costs a harmless no-op gate.

// readOnlyCommands are commands that only observe state. They never cause
// a write and so are never gated, even against production.
func readOnlyCommands() []*cobra.Command {
	return []*cobra.Command{
		authStatusCmd, envListCmd,
		workflowsListCmd, workflowsGetCmd,
		projectsListCmd, projectsGetCmd,
	}
}

// highRiskCommands cause a production side effect (runtime, external,
// destructive, or release) and require confirmation against production.
func highRiskCommands() []*cobra.Command {
	return []*cobra.Command{
		workflowsPublishCmd,
		workflowsDeleteCmd,
		workflowsRunsCreateCmd,
	}
}

// classifyCommands writes the safety annotation onto every command. Run
// once, in this file's init(), before applySafetyGate.
func classifyCommands() {
	for _, c := range readOnlyCommands() {
		markSafety(c, classReadOnly)
	}
	for _, c := range highRiskCommands() {
		markSafety(c, classHighRisk)
		// --confirm only does anything on high-risk commands, so register
		// it locally here rather than as a global persistent flag. This is
		// the single place it gets wired onto the command tree, alongside
		// the safety class itself.
		addConfirmFlag(c)
	}
	// Everything else keeps the normalWrite default from safetyClassOf.
}

func init() {
	classifyCommands()
	applySafetyGate(rootCmd)
}
