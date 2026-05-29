namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the jobs API operations on <see cref="Retab"/>.</summary>
    public class JobsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="JobsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public JobsService(Retab client) : base(client) { }

        /// <summary>List Jobs</summary>
        /// <remarks>
        /// List jobs with pagination and optional status filtering.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Job"/> results.</returns>
        public virtual async Task<PaginatedList<Job>> ListAsync(JobsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Job>("/v1/jobs", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Job>> List(JobsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Job"/> items.</returns>
        public virtual IAsyncEnumerable<Job> ListAutoPagingAsync(JobsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Job>("/v1/jobs", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Job</summary>
        /// <remarks>
        /// Create a new asynchronous job.
        /// The job will be queued for processing and can be polled for status
        /// using the retrieve endpoint.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Job"/> result.</returns>
        public virtual async Task<Job> CreateAsync(JobsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Job>("/v1/jobs", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Job> Create(JobsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Retrieve Job</summary>
        /// <remarks>
        /// Retrieve a job.
        /// Returns the job identified by `job_id`, including its current status,
        /// timestamps, and result (when completed) or error (when failed). Set
        /// `include_request` or `include_response` to embed the original request or
        /// the response payload. Responds with `404` if no matching job exists.
        /// </remarks>
        /// <param name="jobId">The job id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Job"/> result.</returns>
        public virtual async Task<Job> GetAsync(string jobId, JobsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Job>($"/v1/jobs/{Uri.EscapeDataString(jobId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Job> Get(string jobId, JobsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(jobId, options, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Job</summary>
        /// <remarks>
        /// Cancel an in-progress or queued job.
        /// Returns the updated job with status 'cancelled'.
        /// </remarks>
        /// <param name="jobId">The job id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Job"/> result.</returns>
        public virtual async Task<Job> CancelAsync(string jobId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Job>($"/v1/jobs/{Uri.EscapeDataString(jobId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CancelAsync"/>.</summary>
        public virtual Task<Job> Cancel(string jobId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CancelAsync(jobId, requestOptions, cancellationToken);
        }

        /// <summary>Retry Job</summary>
        /// <remarks>
        /// Retry a failed/cancelled/expired job in-place (same job ID).
        /// The job is reset to queued and processed again.
        /// </remarks>
        /// <param name="jobId">The job id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Job"/> result.</returns>
        public virtual async Task<Job> RetryAsync(string jobId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Job>($"/v1/jobs/{Uri.EscapeDataString(jobId)}/retry", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="RetryAsync"/>.</summary>
        public virtual Task<Job> Retry(string jobId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.RetryAsync(jobId, requestOptions, cancellationToken);
        }
    }
}
