namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the workflow review versions API operations on <see cref="Retab"/>.</summary>
    public class WorkflowReviewVersionsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowReviewVersionsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowReviewVersionsService(Retab client) : base(client) { }

        /// <summary>List Review Versions Route</summary>
        /// <remarks>
        /// List versions for one review.
        /// ``review_id`` is required by design — listing versions across all reviews
        /// has no product use and would expose a needlessly wide query surface.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="ReviewVersion"/> results.</returns>
        public virtual async Task<PaginatedList<ReviewVersion>> ListAsync(WorkflowReviewVersionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<ReviewVersion>("/v1/workflows/reviews/versions", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<ReviewVersion>> List(WorkflowReviewVersionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="ReviewVersion"/> items.</returns>
        public virtual IAsyncEnumerable<ReviewVersion> ListAutoPagingAsync(WorkflowReviewVersionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<ReviewVersion>("/v1/workflows/reviews/versions", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Review Version Route</summary>
        /// <remarks>
        /// Create one immutable, content-addressed review version.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ReviewVersion"/> result.</returns>
        public virtual async Task<ReviewVersion> CreateAsync(WorkflowReviewVersionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<ReviewVersion>("/v1/workflows/reviews/versions", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<ReviewVersion> Create(WorkflowReviewVersionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Review Version Route</summary>
        /// <remarks>
        /// Read one review version by its content-addressed id.
        /// </remarks>
        /// <param name="versionId">The version id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ReviewVersion"/> result.</returns>
        public virtual async Task<ReviewVersion> GetAsync(string versionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<ReviewVersion>($"/v1/workflows/reviews/versions/{Uri.EscapeDataString(versionId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<ReviewVersion> Get(string versionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(versionId, requestOptions, cancellationToken);
        }
    }
}
