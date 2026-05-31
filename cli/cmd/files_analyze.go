package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
)

// `files analyze` is a CLI overlay on the public file-analysis job endpoint.
// It lives outside files.go so it is available in both the default and
// generated-prototype CLI builds, where the files router is provided by
// mutually exclusive source files.
var filesAnalyzeCmd = &cobra.Command{
	Use:   "analyze <file-id>",
	Short: "Analyze a file and create a blueprint job",
	Long:  `Analyze an uploaded file asynchronously and return the created job.`,
	Example: `  # Analyze an uploaded document
  retab files analyze file_abc123 --mode reasoning

  # Guide the blueprint with intent
  retab files analyze file_abc123 --intent "Extract policyholder and coverage fields"`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		mode, _ := cmd.Flags().GetString("mode")
		intent, _ := cmd.Flags().GetString("intent")
		metaPairs, _ := cmd.Flags().GetStringArray("metadata")
		metadata, err := parseKVStringList(metaPairs)
		if err != nil {
			return err
		}
		body := map[string]any{"file_id": args[0]}
		if mode != "" {
			body["mode"] = mode
		}
		if intent != "" {
			body["intent"] = intent
		}
		if len(metadata) > 0 {
			body["metadata"] = metadata
		}
		result, err := cliJSONRequest(cmd, http.MethodPost, "/v1/files/analyze", nil, body)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	filesAnalyzeCmd.Flags().Var(newEnumStringFlagValue("--mode", "instant", "reasoning"), "mode", "analysis mode: instant | reasoning")
	filesAnalyzeCmd.Flags().String("intent", "", "optional analysis intent")
	filesAnalyzeCmd.Flags().StringArray("metadata", nil, "metadata key=value (repeatable)")

	filesCmd.AddCommand(filesAnalyzeCmd)
}
