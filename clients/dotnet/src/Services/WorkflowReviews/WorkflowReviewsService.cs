namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the workflow reviews API operations on <see cref="Retab"/>.</summary>
    public class WorkflowReviewsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowReviewsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowReviewsService(Retab client) : base(client) { }

        /// <summary>List Reviews Route</summary>
        /// <remarks>
        /// List reviews — the review queue, oldest first by ``created_at``.
        /// </remarks>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Review"/> results.</returns>
        public virtual async Task<PaginatedList<Review>> ListAsync(string httpBearer, WorkflowReviewsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Review>("/v1/workflows/reviews", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Review>> List(string httpBearer, WorkflowReviewsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Review"/> items.</returns>
        public virtual IAsyncEnumerable<Review> ListAutoPagingAsync(WorkflowReviewsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Review>("/v1/workflows/reviews", options, requestOptions, cancellationToken);
        }

        /// <summary>Get Review Route</summary>
        /// <remarks>
        /// Read one review's metadata + decision. Versions are fetched separately.
        /// </remarks>
        /// <param name="reviewId">The review id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Review"/> result.</returns>
        public virtual async Task<Review> GetAsync(string reviewId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = $"/v1/workflows/reviews/{Uri.EscapeDataString(reviewId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<Review>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Review> Get(string reviewId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(reviewId, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Approve Review Route</summary>
        /// <remarks>
        /// Approve one exact review version and resume the Temporal run.
        /// Earns its action-verb shape per the four criteria in
        /// ``meta-pattern-blueprint.md`` §2: precondition (``decision is None``),
        /// side-effect dominates (Temporal resume signal), divergent request body vs
        /// ``/reject``, divergent response (carries ``resume_status``).
        /// </remarks>
        /// <param name="reviewId">The review id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SubmitDecisionResponse"/> result.</returns>
        public virtual async Task<SubmitDecisionResponse> ApproveAsync(string reviewId, string httpBearer, WorkflowReviewsApproveOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = $"/v1/workflows/reviews/{Uri.EscapeDataString(reviewId)}/approve",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<SubmitDecisionResponse>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ApproveAsync"/>.</summary>
        public virtual Task<SubmitDecisionResponse> Approve(string reviewId, string httpBearer, WorkflowReviewsApproveOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ApproveAsync(reviewId, httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Reject Review Route</summary>
        /// <remarks>
        /// Reject one exact review version and resume the Temporal run.
        /// ``reason`` is required by the request shape — "rejected without reason"
        /// is unrepresentable on the wire.
        /// </remarks>
        /// <param name="reviewId">The review id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="SubmitDecisionResponse"/> result.</returns>
        public virtual async Task<SubmitDecisionResponse> RejectAsync(string reviewId, string httpBearer, WorkflowReviewsRejectOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = $"/v1/workflows/reviews/{Uri.EscapeDataString(reviewId)}/reject",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<SubmitDecisionResponse>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="RejectAsync"/>.</summary>
        public virtual Task<SubmitDecisionResponse> Reject(string reviewId, string httpBearer, WorkflowReviewsRejectOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.RejectAsync(reviewId, httpBearer, options, requestOptions, cancellationToken);
        }
    }
}
