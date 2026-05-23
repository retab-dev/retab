namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

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
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="ExperimentResult"/> results.</returns>
        public virtual async Task<PaginatedList<ExperimentResult>> ListAsync(string httpBearer, ExperimentRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<ExperimentResult>("/v1/workflows/experiments/results", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<ExperimentResult>> List(string httpBearer, ExperimentRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
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
        /// <param name="resultId">The result id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExperimentResult"/> result.</returns>
        public virtual async Task<ExperimentResult> GetAsync(string resultId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = $"/v1/workflows/experiments/results/{Uri.EscapeDataString(resultId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<ExperimentResult>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<ExperimentResult> Get(string resultId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(resultId, httpBearer, requestOptions, cancellationToken);
        }
    }
}
