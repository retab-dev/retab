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
        /// The workflow starts unpublished with a default "Document" input block.
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

        /// <summary>Apply Workflow Spec</summary>
        /// <remarks>
        /// Create a new workflow from a declarative YAML spec.
        /// The workflow id in the YAML is treated as source context, not as the target
        /// workflow id. Use `POST /v1/workflows/{workflow_id}/spec/apply` to modify an
        /// existing workflow draft.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="DeclarativeApplyResponse"/> result.</returns>
        public virtual async Task<DeclarativeApplyResponse> ApplyAsync(WorkflowsApplyOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            string __path = string.IsNullOrEmpty(options.WorkflowId) ? "/v1/workflows/spec/apply" : $"/v1/workflows/{Uri.EscapeDataString(options.WorkflowId)}/spec/apply";
            return await this.PostAsync<DeclarativeApplyResponse>(__path, options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ApplyAsync"/>.</summary>
        public virtual Task<DeclarativeApplyResponse> Apply(WorkflowsApplyOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ApplyAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Plan Workflow Spec</summary>
        /// <remarks>
        /// Preview the changes a declarative YAML spec would make to the draft workflow.
        /// Compares the spec against the current draft and returns the resulting
        /// changes without applying them. A spec that already matches the draft
        /// plans as a no-op.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="DeclarativePlanResponse"/> result.</returns>
        public virtual async Task<DeclarativePlanResponse> PlanAsync(WorkflowsPlanOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            string __path = string.IsNullOrEmpty(options.WorkflowId) ? "/v1/workflows/spec/plan" : $"/v1/workflows/{Uri.EscapeDataString(options.WorkflowId)}/spec/plan";
            return await this.PostAsync<DeclarativePlanResponse>(__path, options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="PlanAsync"/>.</summary>
        public virtual Task<DeclarativePlanResponse> Plan(WorkflowsPlanOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.PlanAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>List Workflow Versions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowGraphVersion"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowGraphVersion>> ListVersionsAsync(WorkflowsListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowGraphVersion>("/v1/workflows/versions", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListVersionsAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowGraphVersion>> ListVersions(WorkflowsListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListVersionsAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListVersionsAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowGraphVersion"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowGraphVersion> ListVersionsAutoPagingAsync(WorkflowsListVersionsOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowGraphVersion>("/v1/workflows/versions", options, requestOptions, cancellationToken);
        }

        /// <summary>Diff Workflow Versions</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowGraphVersionDiff"/> result.</returns>
        public virtual async Task<WorkflowGraphVersionDiff> ListDiffAsync(WorkflowsListDiffOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowGraphVersionDiff>("/v1/workflows/versions/diff", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListDiffAsync"/>.</summary>
        public virtual Task<WorkflowGraphVersionDiff> ListDiff(WorkflowsListDiffOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListDiffAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow Version</summary>
        /// <param name="workflowVersionId">The workflow version id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowGraphVersion"/> result.</returns>
        public virtual async Task<WorkflowGraphVersion> GetVersionAsync(string workflowVersionId, WorkflowsGetVersionOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowGraphVersion>($"/v1/workflows/versions/{Uri.EscapeDataString(workflowVersionId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetVersionAsync"/>.</summary>
        public virtual Task<WorkflowGraphVersion> GetVersion(string workflowVersionId, WorkflowsGetVersionOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetVersionAsync(workflowVersionId, options, requestOptions, cancellationToken);
        }

        /// <summary>Restore Workflow Version</summary>
        /// <param name="workflowVersionId">The workflow version id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Workflow"/> result.</returns>
        public virtual async Task<Workflow> CreateVersionRestoreAsync(string workflowVersionId, WorkflowsCreateVersionRestoreOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Workflow>($"/v1/workflows/versions/{Uri.EscapeDataString(workflowVersionId)}/restore", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateVersionRestoreAsync"/>.</summary>
        public virtual Task<Workflow> CreateVersionRestore(string workflowVersionId, WorkflowsCreateVersionRestoreOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateVersionRestoreAsync(workflowVersionId, options, requestOptions, cancellationToken);
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
        /// This deletes the workflow, all of its blocks and edges, and all of their
        /// snapshots.
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

        /// <summary>Discard Draft Workflow</summary>
        /// <remarks>
        /// Discard all draft changes and restore the workflow to its published state.
        /// The workflow must already be published.
        /// </remarks>
        /// <param name="workflowId">The workflow id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="Workflow"/> result.</returns>
        public virtual async Task<Workflow> DiscardDraftAsync(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<Workflow>($"/v1/workflows/{Uri.EscapeDataString(workflowId)}/discard-draft", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DiscardDraftAsync"/>.</summary>
        public virtual Task<Workflow> DiscardDraft(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DiscardDraftAsync(workflowId, requestOptions, cancellationToken);
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
