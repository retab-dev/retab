namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the usage API operations on <see cref="Retab"/>.</summary>
    public class UsageService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="UsageService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public UsageService(Retab client) : base(client) { }

        /// <summary>List Usage Blocks</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="UsageBlockRecord"/> results.</returns>
        public virtual async Task<PaginatedList<UsageBlockRecord>> ListBlocksAsync(UsageListBlocksOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<UsageBlockRecord>("/v1/usage/blocks", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListBlocksAsync"/>.</summary>
        public virtual Task<PaginatedList<UsageBlockRecord>> ListBlocks(UsageListBlocksOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListBlocksAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListBlocksAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="UsageBlockRecord"/> items.</returns>
        public virtual IAsyncEnumerable<UsageBlockRecord> ListBlocksAutoPagingAsync(UsageListBlocksOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<UsageBlockRecord>("/v1/usage/blocks", options, requestOptions, cancellationToken);
        }

        /// <summary>List Usage Primitives</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="UsagePrimitiveRecord"/> results.</returns>
        public virtual async Task<PaginatedList<UsagePrimitiveRecord>> ListPrimitivesAsync(UsageListPrimitivesOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<UsagePrimitiveRecord>("/v1/usage/primitives", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListPrimitivesAsync"/>.</summary>
        public virtual Task<PaginatedList<UsagePrimitiveRecord>> ListPrimitives(UsageListPrimitivesOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListPrimitivesAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListPrimitivesAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="UsagePrimitiveRecord"/> items.</returns>
        public virtual IAsyncEnumerable<UsagePrimitiveRecord> ListPrimitivesAutoPagingAsync(UsageListPrimitivesOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<UsagePrimitiveRecord>("/v1/usage/primitives", options, requestOptions, cancellationToken);
        }

        /// <summary>List Usage Runs</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="UsageRunRecord"/> results.</returns>
        public virtual async Task<PaginatedList<UsageRunRecord>> ListRunsAsync(UsageListRunsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<UsageRunRecord>("/v1/usage/runs", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListRunsAsync"/>.</summary>
        public virtual Task<PaginatedList<UsageRunRecord>> ListRuns(UsageListRunsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListRunsAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListRunsAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="UsageRunRecord"/> items.</returns>
        public virtual IAsyncEnumerable<UsageRunRecord> ListRunsAutoPagingAsync(UsageListRunsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<UsageRunRecord>("/v1/usage/runs", options, requestOptions, cancellationToken);
        }
    }
}
