package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage long-running jobs",
	Long: `Submit and track asynchronous operations that don't fit a single
synchronous request: large extractions, batch analyses, bulk uploads.

A job wraps any retab endpoint: you supply ` + "`--endpoint`" + ` and a
JSON request body, and the server processes it in the background. Poll
with ` + "`retrieve`" + `, block-until-done with ` + "`wait`" + `, or list
with ` + "`list`" + `. Jobs can be cancelled or retried.

Typical pattern: submit → wait (or poll on a schedule) → fetch the
response with ` + "`retrieve --include-response`" + `.`,
	Example: `  # Submit a job against an endpoint
  retab jobs create \
    --endpoint /v1/extractions \
    --request-file ./req.json \
    --metadata team=ops --metadata source=daily-batch

  # Block until terminal status (completed / failed / cancelled / expired)
  retab jobs wait job_abc123 --timeout-seconds 300

  # Pull the response body once done
  retab jobs retrieve job_abc123 --include-response`,
}

var jobsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a job",
	Long: `Submit a new asynchronous job. ` + "`--endpoint`" + ` is the retab
API path the job should invoke; ` + "`--request-file`" + ` is the JSON body
the server would normally receive for that endpoint.

Attach searchable metadata with ` + "`--metadata key=value`" + ` (repeatable)
to filter later via ` + "`jobs list --metadata key=value`" + `.`,
	Example: `  # Submit an extraction as a job
  retab jobs create \
    --endpoint /v1/extractions \
    --request-file ./extraction-request.json \
    --metadata customer=acme --metadata source=daily-batch

  # Pipe a request from stdin
  cat req.json | retab jobs create \
    --endpoint /v1/parses --request-file -`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		endpoint, err := requireNonBlankFlag(cmd, "endpoint")
		if err != nil {
			return err
		}
		if err := validateJobEndpointFlag(endpoint); err != nil {
			return err
		}
		reqFile, err := requireNonBlankFlag(cmd, "request-file")
		if err != nil {
			return err
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
		if err := validateJobMetadata(md); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Jobs.Create(ctx, &retab.JobsCreateParams{
			Endpoint: retab.CreateJobRequestEndpoint(endpoint),
			Request:  body,
			Metadata: md,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var jobsRetrieveCmd = &cobra.Command{
	Use:     "get <job-id>",
	Aliases: []string{"retrieve"},
	Short:   "Retrieve a job",
	Long: `Fetch a job's current state: status, timestamps, error info.
By default the original request body and response body are omitted to
keep payloads small — use ` + "`--include-request`" + ` /
` + "`--include-response`" + ` to embed them.`,
	Example: `  # Just the status envelope
  retab jobs retrieve job_abc123

  # Embed the response body
  retab jobs retrieve job_abc123 --include-response

  # Poll until done
  while [[ "$(retab jobs retrieve job_abc123 | jq -r '.status')" =~ ^(queued|in_progress)$ ]]; do
    sleep 2
  done`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		includeReq, _ := cmd.Flags().GetBool("include-request")
		includeResp, _ := cmd.Flags().GetBool("include-response")
		result, err := client.Jobs.Get(ctx, args[0], &retab.JobsGetParams{
			IncludeRequest:  ptr(includeReq),
			IncludeResponse: ptr(includeResp),
		})
		if err != nil {
			return err
		}
		if err := printJSON(result); err != nil {
			return err
		}
		return nil
	}),
}

var jobsWaitCmd = &cobra.Command{
	Use:   "wait <job-id>",
	Short: "Poll until a job reaches a terminal status",
	Long: `Block until a job hits a terminal status (` + "`completed`" + `,
` + "`failed`" + `, ` + "`cancelled`" + `, or ` + "`expired`" + `), polling
on a configurable interval. Defaults: 2-second polls, 10-minute timeout.

Cleaner than scripting a poll loop with ` + "`retrieve`" + ` — the CLI
handles backoff and timeout, and exits non-zero if the timeout elapses.`,
	Example: `  # Wait with defaults (2s polls, 600s timeout)
  retab jobs wait job_abc123

  # Aggressive polling, longer timeout
  retab jobs wait job_abc123 \
    --poll-interval-ms 500 --timeout-seconds 1800 \
    --include-response`,
	Args: cobra.ExactArgs(1),
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
		pollInterval := 2 * time.Second
		if pollMS > 0 {
			pollInterval = time.Duration(pollMS) * time.Millisecond
		}
		timeout := 10 * time.Minute
		if timeoutS > 0 {
			timeout = time.Duration(timeoutS) * time.Second
		}
		result, err := waitForJobCompletion(ctx, client, args[0], pollInterval, timeout, includeReq, includeResp)
		if err != nil {
			return err
		}
		if err := printJSON(result); err != nil {
			return err
		}
		return jobWaitTerminalError(result)
	}),
}

var jobsCancelCmd = &cobra.Command{
	Use:   "cancel <job-id>",
	Short: "Cancel a job",
	Long: `Cancel a pending or in-flight job. Completed jobs cannot be
cancelled and the API returns an error.`,
	Example: `  # Cancel a stuck job
  retab jobs cancel job_abc123`,
	Args: cobra.ExactArgs(1),
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
	Long: `Re-execute a failed or cancelled job with its original request
body. Useful after a transient infra issue. The returned job id is the
same — retries are not new jobs.`,
	Example: `  # Retry a failed job
  retab jobs retry job_abc123`,
	Args: cobra.ExactArgs(1),
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
	Long: `List jobs with a wide set of filters: status, endpoint, source,
project/workflow/block ids, model, filename pattern, date range, and
metadata. Page by job id (` + "`--after`" + ` / ` + "`--before`" + `
/ ` + "`--limit`" + `) and metadata filters (` + "`--metadata key=value`" + `)
to slice large job catalogs.`,
	Example: `  # Recent failed jobs
  retab jobs list --status failed --order desc --limit 20

  # Jobs tied to a workflow block
  retab jobs list --workflow-id wf_abc123 --workflow-block-id blk_def456

  # Filter by metadata
  retab jobs list --metadata customer=acme --metadata source=daily-batch

  # Jobs over a date window with filename pattern
  retab jobs list \
    --from-date 2026-05-01 --to-date 2026-05-14 \
    --filename-contains invoice`,
	RunE: runJobsList,
}

func runJobsList(cmd *cobra.Command, args []string) error {
	return runE(func(cmd *cobra.Command, args []string) error {
		if err := validateJobsListFilters(cmd); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.JobsListParams{PaginationParams: collectListParams(cmd)}
		if params.Before != nil && params.After != nil {
			return fmt.Errorf("--before and --after are mutually exclusive")
		}
		if v, _ := cmd.Flags().GetString("id"); v != "" {
			params.JobID = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			status := retab.JobsStatus(v)
			params.Status = &status
		}
		if v, _ := cmd.Flags().GetString("endpoint"); v != "" {
			endpoint := retab.JobsEndpoint(v)
			params.Endpoint = &endpoint
		}
		if v, _ := cmd.Flags().GetString("source"); v != "" {
			source := retab.JobsSource(v)
			params.Source = &source
		}
		if v, _ := cmd.Flags().GetString("project-id"); v != "" {
			params.ProjectID = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("workflow-id"); v != "" {
			params.WorkflowID = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("workflow-block-id"); v != "" {
			params.WorkflowBlockID = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("model"); v != "" {
			params.Model = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("filename-regex"); v != "" {
			params.FilenameRegex = ptr(v)
		}
		if v, _ := cmd.Flags().GetString("filename-contains"); v != "" {
			params.FilenameContains = ptr(v)
		}
		rawDocumentTypes, _ := cmd.Flags().GetStringArray("document-type")
		documentTypes, err := normalizeJobDocumentTypes(rawDocumentTypes)
		if err != nil {
			return err
		}
		params.DocumentType = documentTypes
		fromDate, _ := cmd.Flags().GetString("from-date")
		toDate, _ := cmd.Flags().GetString("to-date")
		if err := validateDateRange("from-date", "to-date", fromDate, toDate); err != nil {
			return err
		}
		if fromDate != "" {
			params.FromDate = ptr(fromDate)
		}
		if toDate != "" {
			params.ToDate = ptr(toDate)
		}
		metaPairs, _ := cmd.Flags().GetStringArray("metadata")
		md, err := parseKVStringList(metaPairs)
		if err != nil {
			return err
		}
		if len(md) > 0 {
			raw, err := json.Marshal(md)
			if err != nil {
				return err
			}
			params.Metadata = ptr(string(raw))
		}
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
		return printJobsListResult(cmd, result)
	})(cmd, args)
}

var allowedJobStatuses = map[string]bool{
	"validating":  true,
	"queued":      true,
	"in_progress": true,
	"completed":   true,
	"failed":      true,
	"cancelled":   true,
	"expired":     true,
}

var allowedJobSources = map[string]bool{
	"api":      true,
	"project":  true,
	"workflow": true,
}

var allowedJobEndpoints = map[string]bool{
	"/v1/documents/extract":        true,
	"/v1/extractions":              true,
	"/v1/documents/parse":          true,
	"/v1/parses":                   true,
	"/v1/documents/split":          true,
	"/v1/splits":                   true,
	"/v1/partitions":               true,
	"/v1/documents/classify":       true,
	"/v1/classifications":          true,
	"/v1/schemas/generate":         true,
	"/v1/edits":                    true,
	"/v1/edits/templates/generate": true,
	"/v1/edit/templates/generate":  true,
	"/v1/evals/extract/process":    true,
	"/v1/evals/extract/extract":    true,
	"/v1/evals/extract/split":      true,
}

var allowedJobDocumentTypes = map[string]bool{
	"bmp": true, "csv": true, "doc": true, "docm": true, "docx": true,
	"dotm": true, "dotx": true, "eml": true, "gif": true, "heic": true,
	"heif": true, "htm": true, "html": true, "jpeg": true, "jpg": true,
	"json": true, "md": true, "mhtml": true, "msg": true, "odp": true,
	"ods": true, "odt": true, "ots": true, "ott": true, "pdf": true,
	"png": true, "ppt": true, "pptx": true, "rtf": true, "svg": true,
	"tif": true, "tiff": true, "tsv": true, "txt": true, "webp": true,
	"xlam": true, "xls": true, "xlsb": true, "xlsm": true, "xlsx": true,
	"xltm": true, "xltx": true, "xml": true, "yaml": true, "yml": true,
}

const maxJobsListLimit = 100
const maxJobsListIncludeResponseLimit = 20
const maxJobsFilenameRegexLength = 256
const maxJobMetadataPairs = 16
const maxJobMetadataKeyLength = 64
const maxJobMetadataValueLength = 512

var literalJobFilenameRegex = regexp.MustCompile(`^[\w\s.\-()]+$`)
var unsupportedJobFilenameRegexFlag = regexp.MustCompile(`\(\?`)
var unsupportedJobFilenameRegexBackreference = regexp.MustCompile(`\\[1-9]`)

func validateJobEndpointFlag(endpoint string) error {
	if endpoint != "" && !allowedJobEndpoints[endpoint] {
		return fmt.Errorf("invalid --endpoint %q", endpoint)
	}
	return nil
}

func normalizeJobDocumentTypes(rawValues []string) ([]string, error) {
	documentTypes := make([]string, 0, len(rawValues))
	seen := map[string]bool{}
	for _, rawValue := range rawValues {
		for _, rawDocumentType := range strings.Split(rawValue, ",") {
			documentType := strings.ToLower(strings.TrimSpace(rawDocumentType))
			if documentType == "" || seen[documentType] {
				continue
			}
			if !allowedJobDocumentTypes[documentType] {
				return nil, fmt.Errorf("invalid --document-type %q", rawDocumentType)
			}
			seen[documentType] = true
			documentTypes = append(documentTypes, documentType)
		}
	}
	return documentTypes, nil
}

func validateJobsListFilters(cmd *cobra.Command) error {
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > maxJobsListLimit {
		return fmt.Errorf("invalid --limit %d (max %d)", limit, maxJobsListLimit)
	}
	includeResponse, _ := cmd.Flags().GetBool("include-response")
	if includeResponse && limit > maxJobsListIncludeResponseLimit {
		return fmt.Errorf("invalid --include-response with --limit %d (max %d)", limit, maxJobsListIncludeResponseLimit)
	}
	status, _ := cmd.Flags().GetString("status")
	if status != "" && !allowedJobStatuses[status] {
		return fmt.Errorf("invalid --status %q", status)
	}
	filenameRegex, _ := cmd.Flags().GetString("filename-regex")
	if err := validateJobFilenameRegexFlag(filenameRegex); err != nil {
		return err
	}
	endpoint, _ := cmd.Flags().GetString("endpoint")
	if err := validateJobEndpointFlag(endpoint); err != nil {
		return err
	}
	source, _ := cmd.Flags().GetString("source")
	if source != "" && !allowedJobSources[source] {
		return fmt.Errorf("invalid --source %q", source)
	}
	projectID, _ := cmd.Flags().GetString("project-id")
	workflowID, _ := cmd.Flags().GetString("workflow-id")
	if source == "api" && (projectID != "" || workflowID != "") {
		return fmt.Errorf("source=api cannot be combined with --project-id or --workflow-id")
	}
	rawDocumentTypes, _ := cmd.Flags().GetStringArray("document-type")
	if _, err := normalizeJobDocumentTypes(rawDocumentTypes); err != nil {
		return err
	}
	return nil
}

func validateJobMetadata(metadata map[string]string) error {
	if len(metadata) > maxJobMetadataPairs {
		return fmt.Errorf("metadata can contain at most %d key-value pairs", maxJobMetadataPairs)
	}
	for key, value := range metadata {
		if len(key) > maxJobMetadataKeyLength {
			return fmt.Errorf("metadata key %q exceeds %d characters", truncateForError(key, 20), maxJobMetadataKeyLength)
		}
		if len(value) > maxJobMetadataValueLength {
			return fmt.Errorf("metadata value for key %q exceeds %d characters", key, maxJobMetadataValueLength)
		}
	}
	return nil
}

func truncateForError(value string, maxLength int) string {
	if len(value) <= maxLength {
		return value
	}
	return value[:maxLength] + "..."
}

func validateJobFilenameRegexFlag(raw string) error {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return nil
	}
	if len(normalized) > maxJobsFilenameRegexLength {
		return fmt.Errorf("invalid --filename-regex: max %d characters", maxJobsFilenameRegexLength)
	}
	if literalJobFilenameRegex.MatchString(normalized) || regexp.QuoteMeta(normalized) == normalized {
		return nil
	}
	if _, err := regexp.Compile(normalized); err != nil {
		return fmt.Errorf("invalid --filename-regex: invalid pattern")
	}
	if !strings.HasPrefix(normalized, "^") {
		return fmt.Errorf("invalid --filename-regex: regex patterns must start with '^'")
	}
	if unsupportedJobFilenameRegexFlag.MatchString(normalized) {
		return fmt.Errorf("invalid --filename-regex: lookaround and inline flags are unsupported")
	}
	if unsupportedJobFilenameRegexBackreference.MatchString(normalized) {
		return fmt.Errorf("invalid --filename-regex: backreferences are unsupported")
	}
	if hasUnescapedJobRegexQuantifier(normalized) {
		return fmt.Errorf("invalid --filename-regex: quantified patterns are unsupported")
	}
	if idx := strings.Index(normalized, "$"); idx >= 0 && !strings.HasSuffix(normalized, "$") {
		return fmt.Errorf("invalid --filename-regex: '$' is only allowed as an end anchor")
	}
	return nil
}

func hasUnescapedJobRegexQuantifier(pattern string) bool {
	escaped := false
	for _, ch := range pattern {
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		switch ch {
		case '*', '+', '?', '{', '}':
			return true
		}
	}
	return false
}

func jobWaitTerminalError(job *retab.Job) error {
	if job == nil {
		return nil
	}
	status := ""
	if job.Status != nil {
		status = string(*job.Status)
	}
	switch status {
	case "", string(retab.JobStatusCompleted):
		return nil
	case string(retab.JobStatusFailed), string(retab.JobStatusCancelled), string(retab.JobStatusExpired):
		id := ""
		if job.ID != nil {
			id = *job.ID
		}
		if id == "" {
			return fmt.Errorf("job ended with status %s", status)
		}
		return fmt.Errorf("job %s ended with status %s", id, status)
	default:
		return nil
	}
}

func waitForJobCompletion(
	ctx context.Context,
	client *retab.Client,
	jobID string,
	pollInterval time.Duration,
	timeout time.Duration,
	includeRequest bool,
	includeResponse bool,
) (*retab.Job, error) {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	params := &retab.JobsGetParams{
		IncludeRequest:  ptr(includeRequest),
		IncludeResponse: ptr(includeResponse),
	}
	for {
		job, err := client.Jobs.Get(ctx, jobID, params)
		if err != nil {
			return nil, err
		}
		if isTerminalJob(job) {
			return job, nil
		}

		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return job, fmt.Errorf("timed out waiting for job %s: %w", jobID, ctx.Err())
		case <-timer.C:
		}
	}
}

func isTerminalJob(job *retab.Job) bool {
	if job == nil || job.Status == nil {
		return false
	}
	switch *job.Status {
	case retab.JobStatusCompleted, retab.JobStatusFailed, retab.JobStatusCancelled, retab.JobStatusExpired:
		return true
	default:
		return false
	}
}

func printJobsListResult(cmd *cobra.Command, result any) error {
	var raw string
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	if raw != string(OutputTable) {
		return printJSON(result)
	}
	return RenderList(os.Stdout, OutputTable, result, []TableColumn{
		{Header: "ID", Extract: func(row any) string { return jobTableCell(row, "id") }},
		{Header: "STATUS", Extract: func(row any) string { return jobTableCell(row, "status") }},
		{Header: "ENDPOINT", Extract: func(row any) string { return jobTableCell(row, "endpoint") }},
		{Header: "CREATED_AT", Extract: func(row any) string { return jobTableCell(row, "created_at") }},
	})
}

func jobTableCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok {
		return ""
	}
	return stringifyCell(value)
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
	jobsWaitCmd.Flags().Var(&nonNegativeIntFlagValue{}, "poll-interval-ms", "polling interval in ms (default 2000)")
	jobsWaitCmd.Flags().Var(&nonNegativeIntFlagValue{}, "timeout-seconds", "wait timeout in seconds (default 600)")

	jobsListCmd.Flags().String("before", "", "job id: return items before this id (mutually exclusive with --after)")
	jobsListCmd.Flags().String("after", "", "job id: return items after this id (mutually exclusive with --before)")
	jobsListCmd.MarkFlagsMutuallyExclusive("before", "after")
	jobsListCmd.Flags().Var(&nonNegativeIntFlagValue{}, "limit", "max items to return")
	jobsListCmd.Flags().Var(&orderFlagValue{}, "order", "asc | desc")
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
	jobsListCmd.Flags().Var(&dateFlagValue{}, "from-date", "filter from this YYYY-MM-DD date")
	jobsListCmd.Flags().Var(&dateFlagValue{}, "to-date", "filter to this YYYY-MM-DD date")
	jobsListCmd.Flags().StringArray("metadata", nil, "metadata key=value filter (repeatable)")
	jobsListCmd.Flags().Bool("include-request", false, "include request body")
	jobsListCmd.Flags().Bool("include-response", false, "include response body")

	jobsCmd.AddCommand(jobsCreateCmd, jobsRetrieveCmd, jobsWaitCmd, jobsCancelCmd, jobsRetryCmd, jobsListCmd)
	rootCmd.AddCommand(jobsCmd)
}
