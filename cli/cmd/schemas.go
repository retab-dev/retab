package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var schemasCmd = &cobra.Command{
	Use:   "schemas",
	Short: "Generate JSON schemas",
}

var schemasGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a JSON schema from one or more documents",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()

		documents := []any{}
		files, _ := cmd.Flags().GetStringArray("file")
		urls, _ := cmd.Flags().GetStringArray("url")
		fileIDs, _ := cmd.Flags().GetStringArray("file-id")
		docsFile, _ := cmd.Flags().GetString("documents-file")

		for _, path := range files {
			mime, err := retab.InferMIMEData(path)
			if err != nil {
				return fmt.Errorf("--file %s: %w", path, err)
			}
			documents = append(documents, mime)
		}
		for _, u := range urls {
			documents = append(documents, retab.MIMEData{URL: u})
		}
		for _, id := range fileIDs {
			documents = append(documents, retab.FileRef{ID: id})
		}
		if docsFile != "" {
			arr, err := readJSONArray(docsFile)
			if err != nil {
				return fmt.Errorf("--documents-file: %w", err)
			}
			documents = append(documents, arr...)
		}
		if len(documents) == 0 {
			return fmt.Errorf("at least one document is required (--file, --url, --file-id, or --documents-file)")
		}

		model, _ := cmd.Flags().GetString("model")
		result, err := client.Schemas.Generate(ctx, retab.GenerateSchemaRequest{
			Documents: documents,
			Model:     model,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	schemasGenerateCmd.Flags().StringArray("file", nil, "path to a document (repeatable)")
	schemasGenerateCmd.Flags().StringArray("url", nil, "document URL (repeatable)")
	schemasGenerateCmd.Flags().StringArray("file-id", nil, "Retab file id (repeatable)")
	schemasGenerateCmd.Flags().String("documents-file", "", "JSON array of documents (or - for stdin)")
	schemasGenerateCmd.Flags().String("model", "", "model identifier")

	schemasCmd.AddCommand(schemasGenerateCmd)
	rootCmd.AddCommand(schemasCmd)
}
