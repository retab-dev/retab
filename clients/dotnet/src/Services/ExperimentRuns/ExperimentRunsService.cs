namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the experiment runs API operations on <see cref="Retab"/>.</summary>
    public class ExperimentRunsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ExperimentRunsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ExperimentRunsService(Retab client) : base(client) { }

        /// <summary>List Experiment Runs</summary>
        /// <remarks>
        /// List experiment runs.
        /// Optionally filter by `workflow_id`, `experiment_id`, `block_id`,
        /// `status`/`exclude_status`, `trigger_type`, and a `from_date`/`to_date`
        /// window. Returns a cursor-paginated list ordered by `sort_by` (default
        /// newest first).
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="ExperimentRun"/> results.</returns>
        public virtual async Task<PaginatedList<ExperimentRun>> ListAsync(ExperimentRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<ExperimentRun>("/v1/workflows/experiments/runs", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<ExperimentRun>> List(ExperimentRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="ExperimentRun"/> items.</returns>
        public virtual IAsyncEnumerable<ExperimentRun> ListAutoPagingAsync(ExperimentRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<ExperimentRun>("/v1/workflows/experiments/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Experiment Run Flat</summary>
        /// <remarks>
        /// Create an experiment run.
        /// The `experiment_id` and an optional `workflow_id` are supplied in the body.
        /// When `workflow_id` is omitted, the experiment's workflow is used; when
        /// supplied, it must match that workflow or the request is rejected with 404.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExperimentRun"/> result.</returns>
        public virtual async Task<ExperimentRun> CreateAsync(ExperimentRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<ExperimentRun>("/v1/workflows/experiments/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<ExperimentRun> Create(ExperimentRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Experiment Run</summary>
        /// <remarks>
        /// Retrieve a single experiment run.
        /// Identified by `run_id`. Returns the run with its lifecycle status, timing,
        /// score, and document progress counts. Returns 404 if no run with that ID
        /// exists.
        /// </remarks>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExperimentRun"/> result.</returns>
        public virtual async Task<ExperimentRun> GetAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<ExperimentRun>($"/v1/workflows/experiments/runs/{Uri.EscapeDataString(runId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<ExperimentRun> Get(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(runId, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Experiment Run</summary>
        /// <remarks>
        /// Cancel an experiment run.
        /// Identified by `run_id`. Cancels the run and any of its pending or
        /// in-flight results, returning the run's new `cancelled` lifecycle. Returns
        /// 404 if the run does not exist or is not in a cancellable (pending, queued,
        /// or running) state.
        /// </remarks>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="CancelWorkflowExperimentRunResponse"/> result.</returns>
        public virtual async Task<CancelWorkflowExperimentRunResponse> CancelAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<CancelWorkflowExperimentRunResponse>($"/v1/workflows/experiments/runs/{Uri.EscapeDataString(runId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CancelAsync"/>.</summary>
        public virtual Task<CancelWorkflowExperimentRunResponse> Cancel(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CancelAsync(runId, requestOptions, cancellationToken);
        }
    }
}
