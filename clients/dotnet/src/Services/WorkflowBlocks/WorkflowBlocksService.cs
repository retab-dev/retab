namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the workflow blocks API operations on <see cref="Retab"/>.</summary>
    public class WorkflowBlocksService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowBlocksService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowBlocksService(Retab client) : base(client) { }

        /// <summary>List Blocks</summary>
        /// <remarks>
        /// List blocks for a workflow with keyset cursor pagination.
        /// Sorted by ``updated_at`` descending with ``id`` as the tiebreaker. Pass
        /// ``after`` (the previous response's ``list_metadata.after``) for the next
        /// page, ``before`` for the previous page. They are mutually exclusive; the
        /// 400 cleanly tells the caller which to drop.
        /// </remarks>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowBlock"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowBlock>> ListAsync(string httpBearer, WorkflowBlocksListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowBlock>("/v1/workflows/blocks", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowBlock>> List(string httpBearer, WorkflowBlocksListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowBlock"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowBlock> ListAutoPagingAsync(WorkflowBlocksListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowBlock>("/v1/workflows/blocks", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Block</summary>
        /// <remarks>
        /// Create a new block in a workflow.
        /// This creates a block in the live workflow_blocks collection.
        /// </remarks>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlock"/> result.</returns>
        public virtual async Task<WorkflowBlock> CreateAsync(string httpBearer, WorkflowBlocksCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = "/v1/workflows/blocks",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<WorkflowBlock>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowBlock> Create(string httpBearer, WorkflowBlocksCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Get Block</summary>
        /// <remarks>
        /// Get a single block by ID.
        /// </remarks>
        /// <param name="blockId">The block id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlock"/> result.</returns>
        public virtual async Task<WorkflowBlock> GetAsync(string blockId, string httpBearer, WorkflowBlocksGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = $"/v1/workflows/blocks/{Uri.EscapeDataString(blockId)}",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<WorkflowBlock>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowBlock> Get(string blockId, string httpBearer, WorkflowBlocksGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(blockId, httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Update Block</summary>
        /// <remarks>
        /// Update a block with partial data.
        /// Only the provided fields are updated. This enables targeted updates
        /// like position changes without affecting other block properties.
        /// </remarks>
        /// <param name="blockId">The block id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlock"/> result.</returns>
        public virtual async Task<WorkflowBlock> UpdateAsync(string blockId, string httpBearer, WorkflowBlocksUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Patch,
                Path = $"/v1/workflows/blocks/{Uri.EscapeDataString(blockId)}",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<WorkflowBlock>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<WorkflowBlock> Update(string blockId, string httpBearer, WorkflowBlocksUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(blockId, httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Block</summary>
        /// <remarks>
        /// Delete a block from a workflow.
        /// This also deletes any edges connected to this block.
        /// </remarks>
        /// <param name="blockId">The block id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string blockId, string httpBearer, WorkflowBlocksDeleteOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Delete,
                Path = $"/v1/workflows/blocks/{Uri.EscapeDataString(blockId)}",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            await this.Client.MakeRawAPIRequest(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string blockId, string httpBearer, WorkflowBlocksDeleteOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(blockId, httpBearer, options, requestOptions, cancellationToken);
        }
    }
}
