namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow eval runs API operations on <see cref="Retab"/>.</summary>
    public class WorkflowEvalRunsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowEvalRunsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowEvalRunsService(Retab client) : base(client) { }

        /// <summary>List Workflow Eval Runs</summary>
        /// <remarks>
        /// List workflow eval runs.
        /// Optionally filter by `workflow_id`, `eval_id`, `target_block_id`,
        /// `status`/`exclude_status`, `trigger_type`, and a `from_date`/`to_date`
        /// window. Returns a cursor-paginated list ordered by `sort_by` (default
        /// newest first).
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowEvalRun"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowEvalRun>> ListAsync(WorkflowEvalRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowEvalRun>("/v1/workflows/evals/runs", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowEvalRun>> List(WorkflowEvalRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowEvalRun"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowEvalRun> ListAutoPagingAsync(WorkflowEvalRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowEvalRun>("/v1/workflows/evals/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Workflow Eval Run</summary>
        /// <remarks>
        /// Create a workflow-scoped eval run.
        /// `workflow_id` is the execution context. Optional `scope` narrows the
        /// run to one saved eval or one block; omitted scope runs all workflow evals.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEvalRun"/> result.</returns>
        public virtual async Task<WorkflowEvalRun> CreateAsync(WorkflowEvalRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowEvalRun>("/v1/workflows/evals/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowEvalRun> Create(WorkflowEvalRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow Eval Run</summary>
        /// <remarks>
        /// Retrieve a single workflow eval run.
        /// Identified by `run_id`. Returns the run with its lifecycle status, timing,
        /// and pass/fail counts. Returns 404 if no run with that ID exists.
        /// </remarks>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEvalRun"/> result.</returns>
        public virtual async Task<WorkflowEvalRun> GetAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowEvalRun>($"/v1/workflows/evals/runs/{Uri.EscapeDataString(runId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowEvalRun> Get(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(runId, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Workflow Eval Run</summary>
        /// <remarks>
        /// Cancel a workflow eval run.
        /// Identified by `run_id`. Stops the run and returns it with its updated
        /// cancelled lifecycle. Returns 404 if the run does not exist or is not in a
        /// cancellable state.
        /// </remarks>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEvalRun"/> result.</returns>
        public virtual async Task<WorkflowEvalRun> CancelAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowEvalRun>($"/v1/workflows/evals/runs/{Uri.EscapeDataString(runId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CancelAsync"/>.</summary>
        public virtual Task<WorkflowEvalRun> Cancel(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CancelAsync(runId, requestOptions, cancellationToken);
        }
    }
}
