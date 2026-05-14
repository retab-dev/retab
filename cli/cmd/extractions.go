package cmd

import (
	"errors"
	"fmt"
	"io"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var extractionsCmd = &cobra.Command{
	Use:   "extractions",
	Short: "Run and manage extractions",
}

func newExtractionRequest(cmd *cobra.Command) (retab.ExtractionCreateRequest, error) {
	doc, err := resolveDocument(cmd)
	if err != nil {
		return retab.ExtractionCreateRequest{}, err
	}
	schema, err := resolveSchema(cmd)
	if err != nil {
		return retab.ExtractionCreateRequest{}, err
	}
	model, _ := cmd.Flags().GetString("model")
	imageDPI, _ := cmd.Flags().GetInt("image-resolution-dpi")
	nConsensus, _ := cmd.Flags().GetInt("n-consensus")
	instructions, _ := cmd.Flags().GetString("instructions")
	bustCache, _ := cmd.Flags().GetBool("bust-cache")
	metaPairs, _ := cmd.Flags().GetStringArray("metadata")
	metadata, err := parseKVStringList(metaPairs)
	if err != nil {
		return retab.ExtractionCreateRequest{}, err
	}
	messagesFile, _ := cmd.Flags().GetString("messages-file")
	var messages []retab.Resource
	if messagesFile != "" {
		arr, err := readJSONArray(messagesFile)
		if err != nil {
			return retab.ExtractionCreateRequest{}, fmt.Errorf("--messages-file: %w", err)
		}
		for _, m := range arr {
			obj, ok := m.(map[string]any)
			if !ok {
				return retab.ExtractionCreateRequest{}, fmt.Errorf("--messages-file: each item must be a JSON object")
			}
			messages = append(messages, retab.Resource(obj))
		}
	}
	return retab.ExtractionCreateRequest{
		Document:           doc,
		JSONSchema:         schema,
		Model:              model,
		ImageResolutionDPI: imageDPI,
		NConsensus:         nConsensus,
		Instructions:       instructions,
		Metadata:           metadata,
		AdditionalMessages: messages,
		BustCache:          bustCache,
	}, nil
}

var extractionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an extraction",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req, err := newExtractionRequest(cmd)
		if err != nil {
			return err
		}
		result, err := client.Extractions.Create(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var extractionsStreamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream an extraction; emits one JSON object per line",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req, err := newExtractionRequest(cmd)
		if err != nil {
			return err
		}
		stream, err := client.Extractions.CreateStream(ctx, req)
		if err != nil {
			return err
		}
		defer stream.Close()
		for {
			item, err := stream.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
			if err := printNDJSON(item); err != nil {
				return err
			}
		}
	}),
}

var extractionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List extractions",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListExtractionsParams{ListParams: collectListParams(cmd)}
		params.OriginType, _ = cmd.Flags().GetString("origin-type")
		params.OriginID, _ = cmd.Flags().GetString("origin-id")
		metaPairs, _ := cmd.Flags().GetStringArray("metadata")
		md, err := parseKVStringList(metaPairs)
		if err != nil {
			return err
		}
		params.Metadata = md
		result, err := client.Extractions.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var extractionsGetCmd = &cobra.Command{
	Use:   "get <extraction-id>",
	Short: "Get an extraction by id",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Extractions.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var extractionsSourcesCmd = &cobra.Command{
	Use:   "sources <extraction-id>",
	Short: "Get the provenance for an extraction",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Extractions.Sources(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var extractionsDeleteCmd = &cobra.Command{
	Use:   "delete <extraction-id>",
	Short: "Delete an extraction",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Extractions.Delete(ctx, args[0])
	}),
}

func addExtractionBodyFlags(cmd *cobra.Command) {
	addDocumentFlags(cmd)
	addSchemaFlags(cmd)
	cmd.Flags().String("model", "", "model identifier (required)")
	cmd.Flags().Int("image-resolution-dpi", 0, "image resolution DPI")
	cmd.Flags().Int("n-consensus", 0, "consensus count")
	cmd.Flags().String("instructions", "", "extra instructions")
	cmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	cmd.Flags().StringArray("metadata", nil, "metadata key=value (repeatable)")
	cmd.Flags().String("messages-file", "", "JSON array of additional_messages (or - for stdin)")
	_ = cmd.MarkFlagRequired("model")
}

func init() {
	addExtractionBodyFlags(extractionsCreateCmd)
	addExtractionBodyFlags(extractionsStreamCmd)

	addListFlags(extractionsListCmd, false)
	extractionsListCmd.Flags().String("origin-type", "", "filter by origin type")
	extractionsListCmd.Flags().String("origin-id", "", "filter by origin id")
	extractionsListCmd.Flags().StringArray("metadata", nil, "metadata key=value filter (repeatable)")

	extractionsCmd.AddCommand(extractionsCreateCmd, extractionsStreamCmd, extractionsListCmd, extractionsGetCmd, extractionsSourcesCmd, extractionsDeleteCmd)
	rootCmd.AddCommand(extractionsCmd)
}
