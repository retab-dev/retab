namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow blocks API operations on <see cref="Retab"/>.</summary>
    public class WorkflowBlocksService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowBlocksService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowBlocksService(Retab client) : base(client) { }

        /// <summary>Gets the nested <see cref="WorkflowBlockExecutionsService"/> service.</summary>
        public virtual WorkflowBlockExecutionsService Executions => new WorkflowBlockExecutionsService(this.Client);

        /// <summary>List Blocks</summary>
        /// <remarks>
        /// List blocks for a workflow with keyset cursor pagination.
        /// Sorted by `updated_at` descending with `id` as the tiebreaker. Pass
        /// `after` (the previous response's `list_metadata.after`) for the next
        /// page, `before` for the previous page. They are mutually exclusive; the
        /// 400 cleanly tells the caller which to drop.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowBlock"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowBlock>> ListAsync(WorkflowBlocksListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowBlock>("/v1/workflows/blocks", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowBlock>> List(WorkflowBlocksListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
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
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlock"/> result.</returns>
        public virtual async Task<WorkflowBlock> CreateAsync(WorkflowBlocksCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowBlock>("/v1/workflows/blocks", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowBlock> Create(WorkflowBlocksCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>List Block Versions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowBlockVersion"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowBlockVersion>> ListVersionsAsync(WorkflowBlocksListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowBlockVersion>("/v1/workflows/blocks/versions", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListVersionsAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowBlockVersion>> ListVersions(WorkflowBlocksListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListVersionsAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListVersionsAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowBlockVersion"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowBlockVersion> ListVersionsAutoPagingAsync(WorkflowBlocksListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowBlockVersion>("/v1/workflows/blocks/versions", options, requestOptions, cancellationToken);
        }

        /// <summary>Diff Block Versions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlockVersionDiff"/> result.</returns>
        public virtual async Task<WorkflowBlockVersionDiff> ListDiffAsync(WorkflowBlocksListDiffOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowBlockVersionDiff>("/v1/workflows/blocks/versions/diff", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListDiffAsync"/>.</summary>
        public virtual Task<WorkflowBlockVersionDiff> ListDiff(WorkflowBlocksListDiffOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListDiffAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Block Version</summary>
        /// <param name="blockVersionId">The block version id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlockVersion"/> result.</returns>
        public virtual async Task<WorkflowBlockVersion> GetVersionAsync(string blockVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowBlockVersion>($"/v1/workflows/blocks/versions/{Uri.EscapeDataString(blockVersionId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetVersionAsync"/>.</summary>
        public virtual Task<WorkflowBlockVersion> GetVersion(string blockVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetVersionAsync(blockVersionId, requestOptions, cancellationToken);
        }

        /// <summary>Restore Block Version</summary>
        /// <param name="blockVersionId">The block version id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlock"/> result.</returns>
        public virtual async Task<WorkflowBlock> CreateVersionRestoreAsync(string blockVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowBlock>($"/v1/workflows/blocks/versions/{Uri.EscapeDataString(blockVersionId)}/restore", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateVersionRestoreAsync"/>.</summary>
        public virtual Task<WorkflowBlock> CreateVersionRestore(string blockVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateVersionRestoreAsync(blockVersionId, requestOptions, cancellationToken);
        }

        /// <summary>Get Block</summary>
        /// <remarks>
        /// Get a single block by ID.
        /// </remarks>
        /// <param name="blockId">The block id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlock"/> result.</returns>
        public virtual async Task<WorkflowBlock> GetAsync(string blockId, WorkflowBlocksGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowBlock>($"/v1/workflows/blocks/{Uri.EscapeDataString(blockId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowBlock> Get(string blockId, WorkflowBlocksGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(blockId, options, requestOptions, cancellationToken);
        }

        /// <summary>Update Block</summary>
        /// <remarks>
        /// Update a block with partial data.
        /// Only the provided fields are updated. This enables targeted updates
        /// like position changes without affecting other block properties.
        /// </remarks>
        /// <param name="blockId">The block id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowBlock"/> result.</returns>
        public virtual async Task<WorkflowBlock> UpdateAsync(string blockId, WorkflowBlocksUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Patch,
                Path = $"/v1/workflows/blocks/{Uri.EscapeDataString(blockId)}",
                RequestOptions = requestOptions,
            };

            if (options.WorkflowId != null)
            {
                request.AddQueryParam("workflow_id", options.WorkflowId);
            }
            if (options.Label != null)
            {
                request.AddBodyParam("label", options.Label);
            }
            if (options.PositionX != null)
            {
                request.AddBodyParam("position_x", options.PositionX);
            }
            if (options.PositionY != null)
            {
                request.AddBodyParam("position_y", options.PositionY);
            }
            if (options.Width != null)
            {
                request.AddBodyParam("width", options.Width);
            }
            if (options.Height != null)
            {
                request.AddBodyParam("height", options.Height);
            }
            if (options.Config != null)
            {
                request.AddBodyParam("config", options.Config);
            }
            if (options.ParentId != null)
            {
                request.AddBodyParam("parent_id", options.ParentId);
            }
            if (options.ConfigMode != null)
            {
                request.AddBodyParam("config_mode", JsonConvert.SerializeObject(options.ConfigMode).Trim('"'));
            }

            return await this.Client.MakeAPIRequest<WorkflowBlock>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<WorkflowBlock> Update(string blockId, WorkflowBlocksUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(blockId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Block</summary>
        /// <remarks>
        /// Delete a block from a workflow.
        /// This also deletes any edges connected to this block.
        /// </remarks>
        /// <param name="blockId">The block id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string blockId, WorkflowBlocksDeleteOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/workflows/blocks/{Uri.EscapeDataString(blockId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string blockId, WorkflowBlocksDeleteOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(blockId, options, requestOptions, cancellationToken);
        }

        /// <summary>Validate Block Config</summary>
        /// <remarks>
        /// Validate an assembled block config without mutating the workflow draft.
        /// </remarks>
        /// <param name="blockId">The block id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ValidateWorkflowBlockConfigResponse"/> result.</returns>
        public virtual async Task<ValidateWorkflowBlockConfigResponse> CreateBlockValidateConfigAsync(string blockId, WorkflowBlocksCreateBlockValidateConfigOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = $"/v1/workflows/blocks/{Uri.EscapeDataString(blockId)}/validate-config",
                RequestOptions = requestOptions,
            };

            if (options.WorkflowId != null)
            {
                request.AddQueryParam("workflow_id", options.WorkflowId);
            }
            if (options.Config != null)
            {
                request.AddBodyParam("config", options.Config);
            }
            if (options.ConfigMode != null)
            {
                request.AddBodyParam("config_mode", JsonConvert.SerializeObject(options.ConfigMode).Trim('"'));
            }

            return await this.Client.MakeAPIRequest<ValidateWorkflowBlockConfigResponse>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateBlockValidateConfigAsync"/>.</summary>
        public virtual Task<ValidateWorkflowBlockConfigResponse> CreateBlockValidateConfig(string blockId, WorkflowBlocksCreateBlockValidateConfigOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateBlockValidateConfigAsync(blockId, options, requestOptions, cancellationToken);
        }
    }
}
