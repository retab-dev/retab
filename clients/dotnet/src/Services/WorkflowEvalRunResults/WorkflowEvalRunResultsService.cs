namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow eval run results API operations on <see cref="Retab"/>.</summary>
    public class WorkflowEvalRunResultsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowEvalRunResultsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowEvalRunResultsService(Retab client) : base(client) { }

        /// <summary>List Workflow Eval Results</summary>
        /// <remarks>
        /// List workflow eval results for a single run, page by page.
        /// Results are returned in run-time order.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowEvalResult"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowEvalResult>> ListAsync(WorkflowEvalRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowEvalResult>("/v1/workflows/evals/results", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowEvalResult>> List(WorkflowEvalRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowEvalResult"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowEvalResult> ListAutoPagingAsync(WorkflowEvalRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowEvalResult>("/v1/workflows/evals/results", options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow Eval Result</summary>
        /// <remarks>
        /// Retrieve a single workflow eval result.
        /// Identified by `result_id`. Returns the result for one eval within a run,
        /// including its `verdict` (`passed`, `failed`, or `blocked`), lifecycle,
        /// timing, and any error. Returns 404 if no result with that ID exists.
        /// </remarks>
        /// <param name="resultId">The result id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEvalResult"/> result.</returns>
        public virtual async Task<WorkflowEvalResult> GetAsync(string resultId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowEvalResult>($"/v1/workflows/evals/results/{Uri.EscapeDataString(resultId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowEvalResult> Get(string resultId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(resultId, requestOptions, cancellationToken);
        }
    }
}
