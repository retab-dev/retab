//go:build !retab_oagen_cli_consensus

package cmd

import (
	"fmt"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var consensusCmd = &cobra.Command{
	Use:   "consensus",
	Short: "Reconcile several objects that describe the same thing into one",
	Long: `Reconcile a set of objects into a single agreed-upon object.

Given several objects that describe the same thing — typically the outputs of
repeated extractions over one document — consensus returns one reconciled
object plus a per-path likelihood telling you how much the inputs agreed on
each leaf. Low-likelihood paths are the ones worth reviewing.

Inputs are always aligned before the vote, so items in a list are matched
across inputs by content rather than by position: a reordered or partially
omitted array still reconciles correctly.`,
}

var consensusCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Reconcile objects into a consensus object with per-path likelihoods",
	Long: `Reconcile objects into one, with a likelihood for every leaf path.

` + "`--inputs`" + ` takes a JSON array of objects (a file path, or ` + "`-`" + ` for stdin).

Pass ` + "`--json-schema`" + ` whenever you have one. It makes numeric fields
reconcile by DECLARED type — ` + "`integer`" + ` votes on the exact value, ` + "`number`" + `
clusters and averages — instead of inferring the kind from whichever values
happen to be present, which measurably changes the result.

` + "`--include-alignment`" + ` additionally returns the canonical path mapping
showing which source path in each input fed each reconciled path.`,
	Example: `  # Reconcile three extraction outputs
  retab consensus create --inputs ./runs.json

  # Same, with the schema so numbers reconcile by declared type
  retab consensus create --inputs ./runs.json --json-schema ./schema.json

  # Pipe inputs in and ask for the alignment mapping
  cat runs.json | retab consensus create --inputs - --include-alignment`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		inputsPath, _ := cmd.Flags().GetString("inputs")
		// readJSON treats "" like "-" (stdin), but only "-" is documented as
		// stdin here: `--inputs ""` would silently block on a terminal waiting
		// for input that never comes, or swallow unrelated piped data.
		if strings.TrimSpace(inputsPath) == "" {
			return fmt.Errorf("--inputs cannot be blank: pass a JSON file path, or - to read stdin")
		}
		raw, err := readJSON(inputsPath)
		if err != nil {
			return err
		}
		inputs, err := consensusInputObjects(raw)
		if err != nil {
			return err
		}

		params := &retab.ConsensusCreateParams{Inputs: inputs}

		if schemaPath, _ := cmd.Flags().GetString("json-schema"); schemaPath != "" {
			schema, err := readJSONMap(schemaPath)
			if err != nil {
				return fmt.Errorf("--json-schema: %w", err)
			}
			params.JSONSchema = &schema
		}
		if includeAlignment, _ := cmd.Flags().GetBool("include-alignment"); includeAlignment {
			params.IncludeAlignment = ptr(true)
		}

		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()

		result, err := client.Consensus.Create(ctx, params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

// consensusInputObjects validates that --inputs decoded to a non-empty JSON
// array of objects. The API needs at least two inputs to have anything to
// reconcile, but one is accepted (it trivially agrees with itself) so a caller
// looping over documents does not have to special-case a single-run document.
func consensusInputObjects(raw any) ([]map[string]any, error) {
	list, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("--inputs must be a JSON array of objects")
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("--inputs must contain at least one object")
	}
	inputs := make([]map[string]any, 0, len(list))
	for i, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("--inputs[%d] must be a JSON object, got %T", i, item)
		}
		inputs = append(inputs, obj)
	}
	return inputs, nil
}

func init() {
	consensusCreateCmd.Flags().String("inputs", "", "path to a JSON array of objects to reconcile, or - for stdin (required)")
	consensusCreateCmd.Flags().String("json-schema", "", "path to the JSON schema describing the objects (recommended)")
	consensusCreateCmd.Flags().Bool("include-alignment", false, "also return the canonical path mapping")
	_ = consensusCreateCmd.MarkFlagRequired("inputs")

	consensusCmd.AddCommand(consensusCreateCmd)
	rootCmd.AddCommand(consensusCmd)
}
