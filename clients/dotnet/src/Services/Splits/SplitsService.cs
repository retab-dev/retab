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
        /// <param name="splitId">The split id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Split"/> result.</returns>
        public virtual async Task<Split> GetAsync(string splitId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Split>($"/v1/splits/{Uri.EscapeDataString(splitId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Split> Get(string splitId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(splitId, requestOptions, cancellationToken);
        }

        /// <summary>Delete Split</summary>
        /// <param name="splitId">The split id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string splitId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/splits/{Uri.EscapeDataString(splitId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string splitId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(splitId, requestOptions, cancellationToken);
        }
    }
}
