namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow runs API operations on <see cref="Retab"/>.</summary>
    public class WorkflowRunsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowRunsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowRunsService(Retab client) : base(client) { }

        /// <summary>List Workflow Runs</summary>
        /// <remarks>
        /// List workflow runs with pagination and optional filters.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowRun"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowRun>> ListAsync(WorkflowRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowRun>("/v1/workflows/runs", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowRun>> List(WorkflowRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowRun"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowRun> ListAutoPagingAsync(WorkflowRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowRun>("/v1/workflows/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Workflow Run Route</summary>
        /// <remarks>
        /// Create a fresh workflow run.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowRun"/> result.</returns>
        public virtual async Task<WorkflowRun> CreateAsync(WorkflowRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowRun>("/v1/workflows/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowRun> Create(WorkflowRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow Export Payload</summary>
        /// <remarks>
        /// Build CSV content for workflow run exports.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowExportPayloadResponse"/> result.</returns>
        public virtual async Task<WorkflowExportPayloadResponse> ExportAsync(WorkflowRunsExportOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowExportPayloadResponse>("/v1/workflows/runs/export", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ExportAsync"/>.</summary>
        public virtual Task<WorkflowExportPayloadResponse> Export(WorkflowRunsExportOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ExportAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow Run</summary>
        /// <remarks>
        /// Get a single workflow run by ID.
        /// </remarks>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowRun"/> result.</returns>
        public virtual async Task<WorkflowRun> GetAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowRun>($"/v1/workflows/runs/{Uri.EscapeDataString(runId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowRun> Get(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(runId, requestOptions, cancellationToken);
        }

        /// <summary>Delete Workflow Run</summary>
        /// <remarks>
        /// Delete a workflow run and its associated step data.
        /// </remarks>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/workflows/runs/{Uri.EscapeDataString(runId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(runId, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Workflow Run</summary>
        /// <remarks>
        /// Cancel a pending, running, or waiting workflow run.
        /// </remarks>
        /// <param name="runId">The run id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="CancelWorkflowResponse"/> result.</returns>
        public virtual async Task<CancelWorkflowResponse> CancelAsync(string runId, WorkflowRunsCancelOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<CancelWorkflowResponse>($"/v1/workflows/runs/{Uri.EscapeDataString(runId)}/cancel", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CancelAsync"/>.</summary>
        public virtual Task<CancelWorkflowResponse> Cancel(string runId, WorkflowRunsCancelOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CancelAsync(runId, options, requestOptions, cancellationToken);
        }
    }
}
