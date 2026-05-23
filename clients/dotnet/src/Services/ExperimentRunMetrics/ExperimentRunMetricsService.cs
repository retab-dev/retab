namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the experiment run metrics API operations on <see cref="Retab"/>.</summary>
    public class ExperimentRunMetricsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="ExperimentRunMetricsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public ExperimentRunMetricsService(Retab client) : base(client) { }

        /// <summary>Get Experiment Metrics For Run</summary>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExperimentSummaryMetricsResponse"/> result.</returns>
        public virtual async Task<ExperimentSummaryMetricsResponse> GetAsync(string httpBearer, ExperimentRunMetricsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = "/v1/workflows/experiments/metrics",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<ExperimentSummaryMetricsResponse>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<ExperimentSummaryMetricsResponse> Get(string httpBearer, ExperimentRunMetricsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(httpBearer, options, requestOptions, cancellationToken);
        }
    }
}
