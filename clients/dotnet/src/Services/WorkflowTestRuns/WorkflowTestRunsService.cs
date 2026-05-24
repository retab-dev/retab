namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow test runs API operations on <see cref="Retab"/>.</summary>
    public class WorkflowTestRunsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowTestRunsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowTestRunsService(Retab client) : base(client) { }

        /// <summary>List Test Execution Runs</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowTestRun"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowTestRun>> ListAsync(WorkflowTestRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowTestRun>("/v1/workflows/tests/runs", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowTestRun>> List(WorkflowTestRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowTestRun"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowTestRun> ListAutoPagingAsync(WorkflowTestRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowTestRun>("/v1/workflows/tests/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Test Run</summary>
        /// <remarks>
        /// Create a workflow-scoped test run.
        /// ``workflow_id`` is the execution context. Optional ``scope`` narrows the
        /// run to one saved test or one block; omitted scope runs all workflow tests.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTestRun"/> result.</returns>
        public virtual async Task<WorkflowTestRun> CreateAsync(WorkflowTestRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowTestRun>("/v1/workflows/tests/runs", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowTestRun> Create(WorkflowTestRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Test Execution Run</summary>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTestRun"/> result.</returns>
        public virtual async Task<WorkflowTestRun> GetAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowTestRun>($"/v1/workflows/tests/runs/{Uri.EscapeDataString(runId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowTestRun> Get(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(runId, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Test Execution Run</summary>
        /// <param name="runId">The run id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTestRun"/> result.</returns>
        public virtual async Task<WorkflowTestRun> CancelAsync(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowTestRun>($"/v1/workflows/tests/runs/{Uri.EscapeDataString(runId)}/cancel", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CancelAsync"/>.</summary>
        public virtual Task<WorkflowTestRun> Cancel(string runId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CancelAsync(runId, requestOptions, cancellationToken);
        }
    }
}
