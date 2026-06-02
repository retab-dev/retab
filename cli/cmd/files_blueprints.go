//go:build !retab_oagen_cli_files

package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var filesBlueprintsCmd = &cobra.Command{
	Use:   "blueprints",
	Short: "Create and manage file blueprints",
	Long: `Create and manage document blueprints for uploaded files.

A file blueprint analyzes an uploaded document and returns a reusable shape
that can guide extraction schemas, workflow design, and follow-up automation.`,
}

var filesBlueprintsCreateCmd = &cobra.Command{
	Use:   "create <file-id>",
	Short: "Create a file blueprint",
	Long: `Create a document blueprint for an uploaded file.

Pass an optional intent to guide the analysis. The command runs in the
background by default so it returns immediately with a queued blueprint. Use
` + "`--wait`" + ` to poll until the blueprint reaches a terminal status, or
` + "`--background=false`" + ` for synchronous creation.`,
	Example: `  # Create a blueprint for an uploaded statement
  retab files blueprints create file_abc123 \
    --mode reasoning \
    --intent "Identify account fields and the transaction table"

  # Start in the background and wait for the final record
  retab files blueprints create file_abc123 --wait`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		mode, _ := cmd.Flags().GetString("mode")
		intent, _ := cmd.Flags().GetString("intent")
		background, _ := cmd.Flags().GetBool("background")
		params := retab.FilesCreateBlueprintParams{
			FileID:     args[0],
			Background: ptr(background),
		}
		if mode != "" {
			typed := retab.CreateFileBlueprintRequestMode(mode)
			params.Mode = &typed
		}
		if intent != "" {
			params.Intent = &intent
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.CreateBlueprint(ctx, &params)
		if err != nil {
			return err
		}
		return maybeWaitForPrimitiveCreate(cmd, fileBlueprintWaitSpec, result)
	}),
}

var filesBlueprintsGetCmd = &cobra.Command{
	Use:   "get <blueprint-id>",
	Short: "Get a file blueprint",
	Long:  `Retrieve a file blueprint by id.`,
	Example: `  # Fetch a completed blueprint
  retab files blueprints get fbp_abc123

  # Fetch only the status projection
  retab files blueprints get fbp_abc123 --include-output=false`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		params := retab.FilesGetBlueprintParams{}
		if cmd.Flags().Changed("include-output") {
			includeOutput, _ := cmd.Flags().GetBool("include-output")
			params.IncludeOutput = &includeOutput
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.GetBlueprint(ctx, args[0], &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesBlueprintsCancelCmd = &cobra.Command{
	Use:   "cancel <blueprint-id>",
	Short: "Cancel a file blueprint",
	Long:  `Cancel an in-flight background file blueprint.`,
	Example: `  # Cancel a queued or running blueprint
  retab files blueprints cancel fbp_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.CreateBlueprintCancel(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	filesBlueprintsCreateCmd.Flags().Var(newEnumStringFlagValue("--mode", "instant", "reasoning"), "mode", "analysis mode: instant | reasoning")
	filesBlueprintsCreateCmd.Flags().String("intent", "", "optional blueprint intent")
	filesBlueprintsCreateCmd.Flags().Bool("background", true, "run asynchronously and return a queued blueprint")
	addPrimitiveCreateWaitFlags(filesBlueprintsCreateCmd)

	filesBlueprintsGetCmd.Flags().Bool("include-output", true, "include the blueprint output")

	filesBlueprintsWaitCmd := primitiveWaitCommand(fileBlueprintWaitSpec)
	addPrimitiveWaitTuningFlags(filesBlueprintsWaitCmd, false)

	filesBlueprintsCmd.AddCommand(filesBlueprintsCreateCmd, filesBlueprintsGetCmd, filesBlueprintsCancelCmd, filesBlueprintsWaitCmd)
	filesCmd.AddCommand(filesBlueprintsCmd)
}
