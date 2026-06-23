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
        /// <remarks>
        /// List partitions.
        /// Returns a paginated list of partitions for the authenticated environment, newest first
        /// by default. Filter by `filename` prefix (case-insensitive) and by a `created_at` window
        /// using `from_date`/`to_date` (`YYYY-MM-DD`). Page through results with `before`/`after`,
        /// `limit`, and `order`.
        /// </remarks>
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
        /// <remarks>
        /// Create a partition.
        /// Groups the pages of a `document` into chunks by a partition `key`, guided by
        /// `instructions` and the chosen `model`. Set `n_consensus` above `1` to run multiple
        /// votes and consolidate them, and `allow_overlap` to let a page belong to more than one
        /// chunk. Returns the stored `Partition` with its `output` chunks, and responds with `201`.
        /// </remarks>
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
        /// <remarks>
        /// Retrieve a partition.
        /// Fetches a single partition by its `partition_id` within the authenticated environment
        /// and returns the full `Partition` including its `output` chunks. Responds with `404` if
        /// no partition with that id exists.
        /// </remarks>
        /// <param name="partitionId">The partition id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Partition"/> result.</returns>
        public virtual async Task<Partition> GetAsync(string partitionId, PartitionsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Partition>($"/v1/partitions/{Uri.EscapeDataString(partitionId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Partition> Get(string partitionId, PartitionsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(partitionId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Partition</summary>
        /// <remarks>
        /// Delete a partition.
        /// Permanently deletes the partition identified by `partition_id`. Returns
        /// `204` on success, or `404` if no partition with that id exists.
        /// </remarks>
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

        /// <summary>Cancel Partition</summary>
        /// <param name="partitionId">The partition id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Partition"/> result.</returns>
        public virtual async Task<Partition> CreateCancelAsync(string partitionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Partition>($"/v1/partitions/{Uri.EscapeDataString(partitionId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateCancelAsync"/>.</summary>
        public virtual Task<Partition> CreateCancel(string partitionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateCancelAsync(partitionId, requestOptions, cancellationToken);
        }
    }
}
