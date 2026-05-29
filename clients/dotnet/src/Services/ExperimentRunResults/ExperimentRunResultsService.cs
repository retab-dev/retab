namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the experiment run results API operations on <see cref="Retab"/>.</summary>
    public class ExperimentRunResultsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ExperimentRunResultsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ExperimentRunResultsService(Retab client) : base(client) { }

        /// <summary>List Experiment Results</summary>
        /// <remarks>
        /// List per-document results for an experiment run.
        /// Requires the `run_id` query parameter. Returns one result row per document
        /// in the run, with each row's lifecycle status, timing, and produced
        /// artifact, as a cursor-paginated list.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="ExperimentResult"/> results.</returns>
        public virtual async Task<PaginatedList<ExperimentResult>> ListAsync(ExperimentRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<ExperimentResult>("/v1/workflows/experiments/results", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<ExperimentResult>> List(ExperimentRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="ExperimentResult"/> items.</returns>
        public virtual IAsyncEnumerable<ExperimentResult> ListAutoPagingAsync(ExperimentRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<ExperimentResult>("/v1/workflows/experiments/results", options, requestOptions, cancellationToken);
        }

        /// <summary>Get Experiment Result</summary>
        /// <remarks>
        /// Retrieve a single experiment result.
        /// Identified by `result_id`. Returns the per-document result with its
        /// lifecycle status, timing, and produced artifact. Returns 404 if no result
        /// with that ID exists.
        /// </remarks>
        /// <param name="resultId">The result id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExperimentResult"/> result.</returns>
        public virtual async Task<ExperimentResult> GetAsync(string resultId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<ExperimentResult>($"/v1/workflows/experiments/results/{Uri.EscapeDataString(resultId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<ExperimentResult> Get(string resultId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(resultId, requestOptions, cancellationToken);
        }
    }
}
