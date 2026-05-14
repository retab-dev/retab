package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var partitionsCmd = &cobra.Command{
	Use:   "partitions",
	Short: "Run and manage partitions",
}

var partitionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a partition",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		doc, err := resolveDocument(cmd)
		if err != nil {
			return err
		}
		key, _ := cmd.Flags().GetString("key")
		instructions, _ := cmd.Flags().GetString("instructions")
		model, _ := cmd.Flags().GetString("model")
		nConsensus, _ := cmd.Flags().GetInt("n-consensus")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		result, err := client.Partitions.Create(ctx, retab.PartitionCreateRequest{
			Document:     doc,
			Key:          key,
			Instructions: instructions,
			Model:        model,
			NConsensus:   nConsensus,
			BustCache:    bustCache,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var partitionsGetCmd = &cobra.Command{
	Use:   "get <partition-id>",
	Short: "Get a partition by id",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Partitions.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var partitionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List partitions",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := collectListParams(cmd)
		result, err := client.Partitions.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var partitionsDeleteCmd = &cobra.Command{
	Use:   "delete <partition-id>",
	Short: "Delete a partition",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Partitions.Delete(ctx, args[0])
	}),
}

func init() {
	addDocumentFlags(partitionsCreateCmd)
	partitionsCreateCmd.Flags().String("key", "", "partition key (required)")
	partitionsCreateCmd.Flags().String("instructions", "", "instructions (required)")
	partitionsCreateCmd.Flags().String("model", "", "model identifier (required)")
	partitionsCreateCmd.Flags().Int("n-consensus", 0, "consensus count")
	partitionsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	_ = partitionsCreateCmd.MarkFlagRequired("key")
	_ = partitionsCreateCmd.MarkFlagRequired("instructions")
	_ = partitionsCreateCmd.MarkFlagRequired("model")

	addListFlags(partitionsListCmd, false)

	partitionsCmd.AddCommand(partitionsCreateCmd, partitionsGetCmd, partitionsListCmd, partitionsDeleteCmd)
	rootCmd.AddCommand(partitionsCmd)
}
