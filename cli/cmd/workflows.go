package cmd

import (
	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Manage workflows",
}

var workflowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflows",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowsParams{}
		params.Before, _ = cmd.Flags().GetString("before")
		params.After, _ = cmd.Flags().GetString("after")
		params.Limit, _ = cmd.Flags().GetInt("limit")
		params.Order, _ = cmd.Flags().GetString("order")
		params.SortBy, _ = cmd.Flags().GetString("sort-by")
		params.Fields, _ = cmd.Flags().GetString("fields")
		result, err := client.Workflows.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsGetCmd = &cobra.Command{
	Use:   "get <workflow-id>",
	Short: "Get a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a workflow",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		result, err := client.Workflows.Create(ctx, retab.CreateWorkflowRequest{
			Name:        name,
			Description: description,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsUpdateCmd = &cobra.Command{
	Use:   "update <workflow-id>",
	Short: "Update a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		var req retab.UpdateWorkflowRequest
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			req.Name = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
		}
		senders, _ := cmd.Flags().GetStringArray("allowed-sender")
		domains, _ := cmd.Flags().GetStringArray("allowed-domain")
		if cmd.Flags().Changed("allowed-sender") || cmd.Flags().Changed("allowed-domain") {
			req.EmailTrigger = &retab.WorkflowEmailTrigger{
				AllowedSenders: senders,
				AllowedDomains: domains,
			}
		}
		result, err := client.Workflows.Update(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id>",
	Short: "Delete a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Workflows.Delete(ctx, args[0])
	}),
}

var workflowsPublishCmd = &cobra.Command{
	Use:   "publish <workflow-id>",
	Short: "Publish the current draft",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		description, _ := cmd.Flags().GetString("description")
		result, err := client.Workflows.Publish(ctx, args[0], retab.PublishWorkflowRequest{Description: description})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsDuplicateCmd = &cobra.Command{
	Use:   "duplicate <workflow-id>",
	Short: "Duplicate a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Duplicate(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEntitiesCmd = &cobra.Command{
	Use:   "entities <workflow-id>",
	Short: "Get the workflow with its blocks and edges",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.GetEntities(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsResolvedSchemasCmd = &cobra.Command{
	Use:   "resolved-schemas <workflow-id>",
	Short: "Get the resolved input/output schemas for every block",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.GetResolvedSchemas(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsDiagnoseCmd = &cobra.Command{
	Use:   "diagnose <workflow-id>",
	Short: "Diagnose the persisted draft graph (use --graph-file to send an in-memory graph)",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		graphFile, _ := cmd.Flags().GetString("graph-file")
		if graphFile != "" {
			body, err := readJSONMap(graphFile)
			if err != nil {
				return err
			}
			req := retab.DiagnoseWorkflowGraphRequest{RePropagate: true}
			if blocks, ok := body["blocks"].([]any); ok {
				for _, b := range blocks {
					if obj, ok := b.(map[string]any); ok {
						req.Blocks = append(req.Blocks, obj)
					}
				}
			}
			if edges, ok := body["edges"].([]any); ok {
				for _, e := range edges {
					if obj, ok := e.(map[string]any); ok {
						req.Edges = append(req.Edges, obj)
					}
				}
			}
			if v, ok := body["re_propagate"].(bool); ok {
				req.RePropagate = v
			}
			result, err := client.Workflows.DiagnoseGraph(ctx, args[0], req)
			if err != nil {
				return err
			}
			return printJSON(result)
		}
		result, err := client.Workflows.Diagnose(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	workflowsListCmd.Flags().String("before", "", "cursor: items before this id")
	workflowsListCmd.Flags().String("after", "", "cursor: items after this id")
	workflowsListCmd.Flags().Int("limit", 0, "max items to return")
	workflowsListCmd.Flags().String("order", "", "asc | desc")
	workflowsListCmd.Flags().String("sort-by", "", "sort field")
	workflowsListCmd.Flags().String("fields", "", "comma-separated field list to return")

	workflowsCreateCmd.Flags().String("name", "", "workflow name")
	workflowsCreateCmd.Flags().String("description", "", "workflow description")

	workflowsUpdateCmd.Flags().String("name", "", "new name")
	workflowsUpdateCmd.Flags().String("description", "", "new description")
	workflowsUpdateCmd.Flags().StringArray("allowed-sender", nil, "email trigger allowed sender (repeatable)")
	workflowsUpdateCmd.Flags().StringArray("allowed-domain", nil, "email trigger allowed domain (repeatable)")

	workflowsPublishCmd.Flags().String("description", "", "publish description")

	workflowsDiagnoseCmd.Flags().String("graph-file", "", "JSON file with {blocks, edges, re_propagate} to diagnose without persisting")

	workflowsCmd.AddCommand(workflowsListCmd, workflowsGetCmd, workflowsCreateCmd, workflowsUpdateCmd, workflowsDeleteCmd, workflowsPublishCmd, workflowsDuplicateCmd, workflowsEntitiesCmd, workflowsResolvedSchemasCmd, workflowsDiagnoseCmd)
	rootCmd.AddCommand(workflowsCmd)
}
