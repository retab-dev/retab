namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the workflow edges API operations on <see cref="Retab"/>.</summary>
    public class WorkflowEdgesService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowEdgesService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowEdgesService(Retab client) : base(client) { }

        /// <summary>List Edges</summary>
        /// <remarks>
        /// List edges for a workflow with keyset cursor pagination.
        /// Optionally filter by source or target block ID. Sorted by `updated_at`
        /// descending with `id` as the tiebreaker. Pass `after` for the next
        /// page, `before` for the previous page — mutually exclusive.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowEdgeDoc"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowEdgeDoc>> ListAsync(WorkflowEdgesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowEdgeDoc>("/v1/workflows/edges", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowEdgeDoc>> List(WorkflowEdgesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowEdgeDoc"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowEdgeDoc> ListAutoPagingAsync(WorkflowEdgesListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowEdgeDoc>("/v1/workflows/edges", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Edge</summary>
        /// <remarks>
        /// Create a new edge connecting two blocks.
        /// Validates that:
        /// - Both source and target blocks exist in the workflow
        /// - The connection is semantically valid (type compatibility, container rules, etc.)
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEdgeDoc"/> result.</returns>
        public virtual async Task<WorkflowEdgeDoc> CreateAsync(WorkflowEdgesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowEdgeDoc>("/v1/workflows/edges", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowEdgeDoc> Create(WorkflowEdgesCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>List Edge Versions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowEdgeVersion"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowEdgeVersion>> ListVersionsAsync(WorkflowEdgesListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowEdgeVersion>("/v1/workflows/edges/versions", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListVersionsAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowEdgeVersion>> ListVersions(WorkflowEdgesListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListVersionsAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListVersionsAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowEdgeVersion"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowEdgeVersion> ListVersionsAutoPagingAsync(WorkflowEdgesListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowEdgeVersion>("/v1/workflows/edges/versions", options, requestOptions, cancellationToken);
        }

        /// <summary>Diff Edge Versions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEdgeVersionDiff"/> result.</returns>
        public virtual async Task<WorkflowEdgeVersionDiff> ListDiffAsync(WorkflowEdgesListDiffOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowEdgeVersionDiff>("/v1/workflows/edges/versions/diff", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListDiffAsync"/>.</summary>
        public virtual Task<WorkflowEdgeVersionDiff> ListDiff(WorkflowEdgesListDiffOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListDiffAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Edge Version</summary>
        /// <param name="edgeVersionId">The edge version id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEdgeVersion"/> result.</returns>
        public virtual async Task<WorkflowEdgeVersion> GetVersionAsync(string edgeVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowEdgeVersion>($"/v1/workflows/edges/versions/{Uri.EscapeDataString(edgeVersionId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetVersionAsync"/>.</summary>
        public virtual Task<WorkflowEdgeVersion> GetVersion(string edgeVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetVersionAsync(edgeVersionId, requestOptions, cancellationToken);
        }

        /// <summary>Restore Edge Version</summary>
        /// <param name="edgeVersionId">The edge version id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEdgeDoc"/> result.</returns>
        public virtual async Task<WorkflowEdgeDoc> CreateVersionRestoreAsync(string edgeVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowEdgeDoc>($"/v1/workflows/edges/versions/{Uri.EscapeDataString(edgeVersionId)}/restore", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateVersionRestoreAsync"/>.</summary>
        public virtual Task<WorkflowEdgeDoc> CreateVersionRestore(string edgeVersionId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateVersionRestoreAsync(edgeVersionId, requestOptions, cancellationToken);
        }

        /// <summary>Get Edge</summary>
        /// <remarks>
        /// Get a single edge by ID.
        /// </remarks>
        /// <param name="edgeId">The edge id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEdgeDoc"/> result.</returns>
        public virtual async Task<WorkflowEdgeDoc> GetAsync(string edgeId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowEdgeDoc>($"/v1/workflows/edges/{Uri.EscapeDataString(edgeId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowEdgeDoc> Get(string edgeId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(edgeId, requestOptions, cancellationToken);
        }

        /// <summary>Delete Edge</summary>
        /// <remarks>
        /// Delete an edge from a workflow.
        /// </remarks>
        /// <param name="edgeId">The edge id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string edgeId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/workflows/edges/{Uri.EscapeDataString(edgeId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string edgeId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(edgeId, requestOptions, cancellationToken);
        }
    }
}
