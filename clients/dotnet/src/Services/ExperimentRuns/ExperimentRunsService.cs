namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

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
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="ExperimentRun"/> results.</returns>
        public virtual async Task<PaginatedList<ExperimentRun>> ListAsync(string httpBearer, ExperimentRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<ExperimentRun>("/v1/workflows/experiments/runs", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<ExperimentRun>> List(string httpBearer, ExperimentRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
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
        /// The ``experiment_id`` and (optionally) ``workflow_id`` live in the
        /// body — flat-resource shape per meta-pattern-blueprint §1. When
        /// ``workflow_id`` is absent the experiment's stored workflow is used;
        /// when present it must match (the validation rejects mismatched pairs
        /// with 404, defending against confused-deputy callers).
        /// </remarks>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExperimentRun"/> result.</returns>
        public virtual async Task<ExperimentRun> CreateAsync(string httpBearer, ExperimentRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = "/v1/workflows/experiments/runs",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<ExperimentRun>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<ExperimentRun> Create(string httpBearer, ExperimentRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Get Experiment Run</summary>
        /// <param name="runId">The run id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExperimentRun"/> result.</returns>
        public virtual async Task<ExperimentRun> GetAsync(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = $"/v1/workflows/experiments/runs/{Uri.EscapeDataString(runId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<ExperimentRun>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<ExperimentRun> Get(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(runId, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Experiment Run</summary>
        /// <param name="runId">The run id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="CancelWorkflowExperimentRunResponse"/> result.</returns>
        public virtual async Task<CancelWorkflowExperimentRunResponse> CancelAsync(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = $"/v1/workflows/experiments/runs/{Uri.EscapeDataString(runId)}/cancel",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<CancelWorkflowExperimentRunResponse>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CancelAsync"/>.</summary>
        public virtual Task<CancelWorkflowExperimentRunResponse> Cancel(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CancelAsync(runId, httpBearer, requestOptions, cancellationToken);
        }
    }
}
