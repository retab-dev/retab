package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var workflowsEdgesCmd = &cobra.Command{
	Use:   "edges",
	Short: "Manage workflow edges",
}

func parseEdgeCreate(obj map[string]any) retab.WorkflowEdgeCreateRequest {
	req := retab.WorkflowEdgeCreateRequest{}
	if v, ok := obj["id"].(string); ok {
		req.ID = v
	}
	if v, ok := obj["source_block"].(string); ok {
		req.SourceBlock = v
	}
	if v, ok := obj["target_block"].(string); ok {
		req.TargetBlock = v
	}
	if v, ok := obj["source_handle"].(string); ok {
		req.SourceHandle = v
	}
	if v, ok := obj["target_handle"].(string); ok {
		req.TargetHandle = v
	}
	return req
}

var workflowsEdgesListCmd = &cobra.Command{
	Use:   "list <workflow-id>",
	Short: "List edges in a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListWorkflowEdgesParams{}
		params.SourceBlock, _ = cmd.Flags().GetString("source-block")
		params.TargetBlock, _ = cmd.Flags().GetString("target-block")
		result, err := client.Workflows.Edges.List(ctx, args[0], &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEdgesGetCmd = &cobra.Command{
	Use:   "get <workflow-id> <edge-id>",
	Short: "Get an edge",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Workflows.Edges.Get(ctx, args[0], args[1])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEdgesCreateCmd = &cobra.Command{
	Use:   "create <workflow-id>",
	Short: "Create an edge",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.WorkflowEdgeCreateRequest{}
		req.SourceBlock, _ = cmd.Flags().GetString("source-block")
		req.TargetBlock, _ = cmd.Flags().GetString("target-block")
		req.SourceHandle, _ = cmd.Flags().GetString("source-handle")
		req.TargetHandle, _ = cmd.Flags().GetString("target-handle")
		req.ID, _ = cmd.Flags().GetString("id")
		result, err := client.Workflows.Edges.Create(ctx, args[0], req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEdgesCreateBatchCmd = &cobra.Command{
	Use:   "create-batch <workflow-id>",
	Short: "Create multiple edges from --edges-file (JSON array)",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		path, _ := cmd.Flags().GetString("edges-file")
		if path == "" {
			return fmt.Errorf("--edges-file is required")
		}
		arr, err := readJSONArray(path)
		if err != nil {
			return err
		}
		var reqs []retab.WorkflowEdgeCreateRequest
		for i, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("--edges-file[%d]: must be a JSON object", i)
			}
			reqs = append(reqs, parseEdgeCreate(obj))
		}
		result, err := client.Workflows.Edges.CreateBatch(ctx, args[0], reqs)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var workflowsEdgesDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-id> <edge-id>",
	Short: "Delete an edge",
	Args:  cobra.ExactArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Workflows.Edges.Delete(ctx, args[0], args[1])
	}),
}

var workflowsEdgesDeleteAllCmd = &cobra.Command{
	Use:   "delete-all <workflow-id>",
	Short: "Delete all edges in a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Workflows.Edges.DeleteAll(ctx, args[0])
	}),
}

func init() {
	workflowsEdgesListCmd.Flags().String("source-block", "", "filter by source block")
	workflowsEdgesListCmd.Flags().String("target-block", "", "filter by target block")

	workflowsEdgesCreateCmd.Flags().String("id", "", "edge id (optional)")
	workflowsEdgesCreateCmd.Flags().String("source-block", "", "source block id (required)")
	workflowsEdgesCreateCmd.Flags().String("target-block", "", "target block id (required)")
	workflowsEdgesCreateCmd.Flags().String("source-handle", "", "source handle")
	workflowsEdgesCreateCmd.Flags().String("target-handle", "", "target handle")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("source-block")
	_ = workflowsEdgesCreateCmd.MarkFlagRequired("target-block")

	workflowsEdgesCreateBatchCmd.Flags().String("edges-file", "", "JSON array of edges (or - for stdin)")

	workflowsEdgesCmd.AddCommand(workflowsEdgesListCmd, workflowsEdgesGetCmd, workflowsEdgesCreateCmd, workflowsEdgesCreateBatchCmd, workflowsEdgesDeleteCmd, workflowsEdgesDeleteAllCmd)
	workflowsCmd.AddCommand(workflowsEdgesCmd)
}
