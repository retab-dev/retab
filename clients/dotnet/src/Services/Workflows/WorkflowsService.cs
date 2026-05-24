namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflows API operations on <see cref="Retab"/>.</summary>
    public class WorkflowsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowsService(Retab client) : base(client) { }

        /// <summary>Gets the nested <see cref="WorkflowArtifactsService"/> service.</summary>
        public virtual WorkflowArtifactsService Artifacts => new WorkflowArtifactsService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowBlocksService"/> service.</summary>
        public virtual WorkflowBlocksService Blocks => new WorkflowBlocksService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowEdgesService"/> service.</summary>
        public virtual WorkflowEdgesService Edges => new WorkflowEdgesService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowExperimentsService"/> service.</summary>
        public virtual WorkflowExperimentsService Experiments => new WorkflowExperimentsService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowReviewsService"/> service.</summary>
        public virtual WorkflowReviewsService Reviews => new WorkflowReviewsService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowRunsService"/> service.</summary>
        public virtual WorkflowRunsService Runs => new WorkflowRunsService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowSpecService"/> service.</summary>
        public virtual WorkflowSpecService Spec => new WorkflowSpecService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowStepsService"/> service.</summary>
        public virtual WorkflowStepsService Steps => new WorkflowStepsService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowTestsService"/> service.</summary>
        public virtual WorkflowTestsService Tests => new WorkflowTestsService(this.Client);

        /// <summary>List Workflows</summary>
        /// <remarks>
        /// List workflows with pagination and optional filtering.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="Workflow"/> results.</returns>
        public virtual async Task<PaginatedList<Workflow>> ListAsync(WorkflowsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<Workflow>("/v1/workflows", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<Workflow>> List(WorkflowsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="Workflow"/> items.</returns>
        public virtual IAsyncEnumerable<Workflow> ListAutoPagingAsync(WorkflowsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<Workflow>("/v1/workflows", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Workflow</summary>
        /// <remarks>
        /// Create a new workflow.
        /// The workflow starts unpublished and is scaffolded with a default
        /// "Document" input block in the live block collection.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Workflow"/> result.</returns>
        public virtual async Task<Workflow> CreateAsync(WorkflowsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Workflow>("/v1/workflows", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<Workflow> Create(WorkflowsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow</summary>
        /// <remarks>
        /// Get a single workflow by ID.
        /// Returns workflow metadata only.
        /// </remarks>
        /// <param name="workflowId">The workflow id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Workflow"/> result.</returns>
        public virtual async Task<Workflow> GetAsync(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<Workflow>($"/v1/workflows/{Uri.EscapeDataString(workflowId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<Workflow> Get(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(workflowId, requestOptions, cancellationToken);
        }

        /// <summary>Update Workflow</summary>
        /// <remarks>
        /// Update an existing workflow.
        /// </remarks>
        /// <param name="workflowId">The workflow id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Workflow"/> result.</returns>
        public virtual async Task<Workflow> UpdateAsync(string workflowId, WorkflowsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PatchAsync<Workflow>($"/v1/workflows/{Uri.EscapeDataString(workflowId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<Workflow> Update(string workflowId, WorkflowsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(workflowId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Workflow</summary>
        /// <remarks>
        /// Delete a workflow and all its associated entities.
        /// This deletes:
        /// - The workflow document
        /// - All blocks and edges (live collections)
        /// - All block and edge snapshots
        /// - All workflow snapshots
        /// </remarks>
        /// <param name="workflowId">The workflow id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/workflows/{Uri.EscapeDataString(workflowId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(workflowId, requestOptions, cancellationToken);
        }

        /// <summary>Publish Workflow</summary>
        /// <remarks>
        /// Publish a workflow.
        /// This creates an immutable snapshot of the workflow configuration, making it available for workflow runs.
        /// The live entities remain unchanged so users can continue editing.
        /// </remarks>
        /// <param name="workflowId">The workflow id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Workflow"/> result.</returns>
        public virtual async Task<Workflow> PublishAsync(string workflowId, WorkflowsPublishOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Workflow>($"/v1/workflows/{Uri.EscapeDataString(workflowId)}/publish", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="PublishAsync"/>.</summary>
        public virtual Task<Workflow> Publish(string workflowId, WorkflowsPublishOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.PublishAsync(workflowId, options, requestOptions, cancellationToken);
        }
    }
}
