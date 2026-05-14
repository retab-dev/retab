package cmd

import (
	"fmt"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage long-running jobs",
}

var jobsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a job",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		endpoint, _ := cmd.Flags().GetString("endpoint")
		reqFile, _ := cmd.Flags().GetString("request-file")
		if reqFile == "" {
			return fmt.Errorf("--request-file is required")
		}
		body, err := readJSONMap(reqFile)
		if err != nil {
			return fmt.Errorf("--request-file: %w", err)
		}
		metaPairs, _ := cmd.Flags().GetStringArray("metadata")
		md, err := parseKVStringList(metaPairs)
		if err != nil {
			return err
		}
		result, err := client.Jobs.Create(ctx, retab.JobCreateRequest{
			Endpoint: endpoint,
			Request:  retab.Resource(body),
			Metadata: md,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var jobsRetrieveCmd = &cobra.Command{
	Use:   "retrieve <job-id>",
	Short: "Retrieve a job",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		includeReq, _ := cmd.Flags().GetBool("include-request")
		includeResp, _ := cmd.Flags().GetBool("include-response")
		result, err := client.Jobs.Retrieve(ctx, args[0], &retab.JobRetrieveParams{
			IncludeRequest:  includeReq,
			IncludeResponse: includeResp,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var jobsRetrieveFullCmd = &cobra.Command{
	Use:   "retrieve-full <job-id>",
	Short: "Retrieve a job with full request and response",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Jobs.RetrieveFull(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var jobsWaitCmd = &cobra.Command{
	Use:   "wait <job-id>",
	Short: "Poll until a job reaches a terminal status",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		pollMS, _ := cmd.Flags().GetInt("poll-interval-ms")
		timeoutS, _ := cmd.Flags().GetInt("timeout-seconds")
		includeReq, _ := cmd.Flags().GetBool("include-request")
		includeResp, _ := cmd.Flags().GetBool("include-response")
		params := &retab.JobWaitForCompletionParams{
			IncludeRequest:  includeReq,
			IncludeResponse: includeResp,
		}
		if pollMS > 0 {
			params.PollInterval = time.Duration(pollMS) * time.Millisecond
		}
		if timeoutS > 0 {
			params.Timeout = time.Duration(timeoutS) * time.Second
		}
		result, err := client.Jobs.WaitForCompletion(ctx, args[0], params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var jobsCancelCmd = &cobra.Command{
	Use:   "cancel <job-id>",
	Short: "Cancel a job",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Jobs.Cancel(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var jobsRetryCmd = &cobra.Command{
	Use:   "retry <job-id>",
	Short: "Retry a failed job",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Jobs.Retry(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var jobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListJobsParams{}
		params.Before, _ = cmd.Flags().GetString("before")
		params.After, _ = cmd.Flags().GetString("after")
		params.Limit, _ = cmd.Flags().GetInt("limit")
		params.Order, _ = cmd.Flags().GetString("order")
		params.ID, _ = cmd.Flags().GetString("id")
		params.Status, _ = cmd.Flags().GetString("status")
		params.Endpoint, _ = cmd.Flags().GetString("endpoint")
		params.Source, _ = cmd.Flags().GetString("source")
		params.ProjectID, _ = cmd.Flags().GetString("project-id")
		params.WorkflowID, _ = cmd.Flags().GetString("workflow-id")
		params.WorkflowBlockID, _ = cmd.Flags().GetString("workflow-block-id")
		params.Model, _ = cmd.Flags().GetString("model")
		params.FilenameRegex, _ = cmd.Flags().GetString("filename-regex")
		params.FilenameContains, _ = cmd.Flags().GetString("filename-contains")
		params.DocumentType, _ = cmd.Flags().GetStringArray("document-type")
		params.FromDate, _ = cmd.Flags().GetString("from-date")
		params.ToDate, _ = cmd.Flags().GetString("to-date")
		metaPairs, _ := cmd.Flags().GetStringArray("metadata")
		md, err := parseKVStringList(metaPairs)
		if err != nil {
			return err
		}
		params.Metadata = md
		if cmd.Flags().Changed("include-request") {
			v, _ := cmd.Flags().GetBool("include-request")
			params.IncludeRequest = &v
		}
		if cmd.Flags().Changed("include-response") {
			v, _ := cmd.Flags().GetBool("include-response")
			params.IncludeResponse = &v
		}
		result, err := client.Jobs.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	jobsCreateCmd.Flags().String("endpoint", "", "endpoint to invoke (required)")
	jobsCreateCmd.Flags().String("request-file", "", "JSON file with the request body (or - for stdin) (required)")
	jobsCreateCmd.Flags().StringArray("metadata", nil, "metadata key=value (repeatable)")
	_ = jobsCreateCmd.MarkFlagRequired("endpoint")
	_ = jobsCreateCmd.MarkFlagRequired("request-file")

	for _, c := range []*cobra.Command{jobsRetrieveCmd, jobsWaitCmd} {
		c.Flags().Bool("include-request", false, "include the original request body")
		c.Flags().Bool("include-response", false, "include the full response body")
	}
	jobsWaitCmd.Flags().Int("poll-interval-ms", 0, "polling interval in ms (default 2000)")
	jobsWaitCmd.Flags().Int("timeout-seconds", 0, "wait timeout in seconds (default 600)")

	jobsListCmd.Flags().String("before", "", "cursor: items before this id")
	jobsListCmd.Flags().String("after", "", "cursor: items after this id")
	jobsListCmd.Flags().Int("limit", 0, "max items to return")
	jobsListCmd.Flags().String("order", "", "asc | desc")
	jobsListCmd.Flags().String("id", "", "filter by job id")
	jobsListCmd.Flags().String("status", "", "filter by status")
	jobsListCmd.Flags().String("endpoint", "", "filter by endpoint")
	jobsListCmd.Flags().String("source", "", "filter by source")
	jobsListCmd.Flags().String("project-id", "", "filter by project id")
	jobsListCmd.Flags().String("workflow-id", "", "filter by workflow id")
	jobsListCmd.Flags().String("workflow-block-id", "", "filter by workflow block id")
	jobsListCmd.Flags().String("model", "", "filter by model")
	jobsListCmd.Flags().String("filename-regex", "", "filter by filename regex")
	jobsListCmd.Flags().String("filename-contains", "", "filter by filename substring")
	jobsListCmd.Flags().StringArray("document-type", nil, "filter by document type (repeatable)")
	jobsListCmd.Flags().String("from-date", "", "filter from this date")
	jobsListCmd.Flags().String("to-date", "", "filter to this date")
	jobsListCmd.Flags().StringArray("metadata", nil, "metadata key=value filter (repeatable)")
	jobsListCmd.Flags().Bool("include-request", false, "include request body")
	jobsListCmd.Flags().Bool("include-response", false, "include response body")

	jobsCmd.AddCommand(jobsCreateCmd, jobsRetrieveCmd, jobsRetrieveFullCmd, jobsWaitCmd, jobsCancelCmd, jobsRetryCmd, jobsListCmd)
	rootCmd.AddCommand(jobsCmd)
}
