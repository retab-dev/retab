package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var partitionsCmd = &cobra.Command{
	Use:   "partitions",
	Short: "Partition repeated records in a document into chunks by a unique identifier",
	Long: `Partition documents containing repeated records into per-record chunks.

Use this when a single document contains many instances of the same record
type — bank statements (one entry per transaction), batched invoice PDFs
(one invoice per record), multi-employee payslip files, etc. You supply a
unique key that identifies record boundaries (e.g. "transaction date",
"invoice number") and the service returns one chunk per record.

A common pipeline is partition → ` + "`retab extractions create`" + ` per chunk
with a per-record schema. For splitting heterogeneous sections (e.g.
invoice + packing slip), use ` + "`retab splits`" + ` instead.`,
}

var partitionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a partition",
	Long: `Partition a document into per-record chunks by a unique key.

` + "`--key`" + ` is the natural-language name of the field that uniquely
identifies each record (e.g. ` + "`\"transaction id\"`" + `,
` + "`\"invoice number\"`" + `). ` + "`--instructions`" + ` gives any extra hints
about where records start and end. For tight boundary detection on
ambiguous documents, set ` + "`--n-consensus`" + ` to sample multiple runs.

Pair with ` + "`retab extractions create`" + ` per chunk for per-record schema
extraction.`,
	Example: `  # Partition a bank statement by transaction date
  retab partitions create \
    --file ./statement.pdf --model gpt-4o \
    --key "transaction date" \
    --instructions "Each transaction is one row in the table"

  # Higher accuracy via consensus
  retab partitions create \
    --file-id file_abc123 --model gpt-4o \
    --key "invoice number" \
    --instructions "Boundaries are clear page breaks between invoices" \
    --n-consensus 3`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		key, err := requireNonBlankFlag(cmd, "key")
		if err != nil {
			return err
		}
		instructions, err := requireNonBlankFlag(cmd, "instructions")
		if err != nil {
			return err
		}
		model, err := requireNonBlankFlag(cmd, "model")
		if err != nil {
			return err
		}
		doc, err := resolveDocument(cmd)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		nConsensus, _ := cmd.Flags().GetInt("n-consensus")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		allowOverlap, _ := cmd.Flags().GetBool("allow-overlap")
		requestOptions := []retab.RequestOption{}
		if cmd.Flags().Changed("allow-overlap") && !allowOverlap {
			requestOptions = append(requestOptions, retab.WithRequestBody(map[string]any{
				"allow_overlap": false,
			}))
		}
		result, err := client.Partitions.Create(ctx, retab.PartitionCreateRequest{
			Document:     doc,
			Key:          key,
			Instructions: instructions,
			Model:        model,
			NConsensus:   nConsensus,
			BustCache:    bustCache,
			AllowOverlap: allowOverlap,
		}, requestOptions...)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var partitionsGetCmd = &cobra.Command{
	Use:   "get <partition-id>",
	Short: "Get a partition by id",
	Long: `Fetch a single partition by id.

Returns the source document reference, the partition key, instructions,
and the resolved per-record chunks (with page ranges and the extracted key
value for each).`,
	Example: `  # Fetch a known partition
  retab partitions get part_xyz789

  # List the key values of every resolved chunk
  retab partitions get part_xyz789 | jq '.chunks[].key_value'`,
	Args: cobra.ExactArgs(1),
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
	Long: `List partitions, newest first by default.

Page by partition id with ` + "`--before`" + ` / ` + "`--after`" + `, cap page size with
` + "`--limit`" + `.`,
	Example: `  # Most recent 25 partitions
  retab partitions list --limit 25

  # Walk pages from a known id
  retab partitions list --after part_xyz789 --limit 50`,
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
		return printResult(cmd, result)
	}),
}

var partitionsDeleteCmd = &cobra.Command{
	Use:   "delete <partition-id>",
	Short: "Delete a partition",
	Long: `Permanently delete a partition.

Destructive and irreversible. The source document is not affected. Take a
backup with ` + "`retab partitions get`" + ` first if you may need the chunk
definitions.`,
	Example: `  # Back up, then delete
  retab partitions get part_xyz789 > backup.json
  retab partitions delete part_xyz789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Partitions.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("partition", args[0])
		return nil
	}),
}

func init() {
	addDocumentFlags(partitionsCreateCmd)
	partitionsCreateCmd.Flags().String("key", "", "partition key (required)")
	partitionsCreateCmd.Flags().String("instructions", "", "instructions (required)")
	partitionsCreateCmd.Flags().String("model", "", "model identifier (required)")
	partitionsCreateCmd.Flags().Var(&boundedIntFlagValue{min: 1, max: 8}, "n-consensus", "consensus count (1-8)")
	partitionsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	partitionsCreateCmd.Flags().Bool("allow-overlap", true, "allow partition chunks to share pages")
	_ = partitionsCreateCmd.MarkFlagRequired("key")
	_ = partitionsCreateCmd.MarkFlagRequired("instructions")
	_ = partitionsCreateCmd.MarkFlagRequired("model")

	addListFlags(partitionsListCmd, false)

	partitionsCmd.AddCommand(partitionsCreateCmd, partitionsGetCmd, partitionsListCmd, partitionsDeleteCmd)
	rootCmd.AddCommand(partitionsCmd)
}
