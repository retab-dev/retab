namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the partitions API operations on <see cref="Retab"/>.</summary>
    public class PartitionsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="PartitionsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public PartitionsService(Retab client) : base(client) { }

        /// <summary>List Partitions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Partition"/> results.</returns>
        public virtual async Task<PaginatedList<Partition>> ListAsync(PartitionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Partition>("/v1/partitions", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Partition>> List(PartitionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Partition"/> items.</returns>
        public virtual IAsyncEnumerable<Partition> ListAutoPagingAsync(PartitionsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Partition>("/v1/partitions", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Partitions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Partition"/> result.</returns>
        public virtual async Task<Partition> CreateAsync(PartitionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Partition>("/v1/partitions", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Partition> Create(PartitionsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Partition</summary>
        /// <param name="partitionId">The partition id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Partition"/> result.</returns>
        public virtual async Task<Partition> GetAsync(string partitionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Partition>($"/v1/partitions/{Uri.EscapeDataString(partitionId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Partition> Get(string partitionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(partitionId, requestOptions, cancellationToken);
        }

        /// <summary>Delete Partition</summary>
        /// <param name="partitionId">The partition id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string partitionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/partitions/{Uri.EscapeDataString(partitionId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string partitionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(partitionId, requestOptions, cancellationToken);
        }
    }
}
