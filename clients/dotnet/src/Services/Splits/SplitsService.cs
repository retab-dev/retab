namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the splits API operations on <see cref="Retab"/>.</summary>
    public class SplitsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="SplitsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public SplitsService(Retab client) : base(client) { }

        /// <summary>List Splits</summary>
        /// <remarks>
        /// List splits.
        /// Returns a paginated list of splits for the authenticated environment, newest first by
        /// default. Filter by `filename` prefix (case-insensitive) and by a `created_at` window
        /// using `from_date`/`to_date` (`YYYY-MM-DD`). Page through results with `before`/`after`,
        /// `limit`, and `order`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Split"/> results.</returns>
        public virtual async Task<PaginatedList<Split>> ListAsync(SplitsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Split>("/v1/splits", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Split>> List(SplitsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Split"/> items.</returns>
        public virtual IAsyncEnumerable<Split> ListAutoPagingAsync(SplitsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Split>("/v1/splits", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Split</summary>
        /// <remarks>
        /// Create a split.
        /// Divides a `document` into the named `subdocuments`, assigning each its set of pages,
        /// using the chosen `model` and optional `instructions`. Set `n_consensus` above `1` to
        /// run multiple votes and consolidate them. Returns the stored `Split` with its `output`
        /// page assignments, and responds with `201`.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Split"/> result.</returns>
        public virtual async Task<Split> CreateAsync(SplitsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Split>("/v1/splits", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Split> Create(SplitsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Split</summary>
        /// <remarks>
        /// Retrieve a split.
        /// Fetches a single split by its `split_id` within the authenticated environment and
        /// returns the full `Split` including its `output` page assignments. Responds with `404`
        /// if no split with that id exists.
        /// </remarks>
        /// <param name="splitId">The split id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Split"/> result.</returns>
        public virtual async Task<Split> GetAsync(string splitId, SplitsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Split>($"/v1/splits/{Uri.EscapeDataString(splitId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Split> Get(string splitId, SplitsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(splitId, options, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Split</summary>
        /// <param name="splitId">The split id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Split"/> result.</returns>
        public virtual async Task<Split> CreateCancelAsync(string splitId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Split>($"/v1/splits/{Uri.EscapeDataString(splitId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateCancelAsync"/>.</summary>
        public virtual Task<Split> CreateCancel(string splitId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateCancelAsync(splitId, requestOptions, cancellationToken);
        }
    }
}
