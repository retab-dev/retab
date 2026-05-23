namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

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
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowTestRun"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowTestRun>> ListAsync(string httpBearer, WorkflowTestRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowTestRun>("/v1/workflows/tests/runs", options, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowTestRun>> List(string httpBearer, WorkflowTestRunsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(httpBearer, options, requestOptions, cancellationToken);
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
        /// Scoping identity comes from the body — ``workflow_id`` and/or
        /// ``test_id`` per ``CreateWorkflowTestRunRequest``.
        /// </remarks>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTestRun"/> result.</returns>
        public virtual async Task<WorkflowTestRun> CreateAsync(string httpBearer, WorkflowTestRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = "/v1/workflows/tests/runs",
                Options = options,
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<WorkflowTestRun>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowTestRun> Create(string httpBearer, WorkflowTestRunsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(httpBearer, options, requestOptions, cancellationToken);
        }

        /// <summary>Get Test Execution Run</summary>
        /// <param name="runId">The run id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTestRun"/> result.</returns>
        public virtual async Task<WorkflowTestRun> GetAsync(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = $"/v1/workflows/tests/runs/{Uri.EscapeDataString(runId)}",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<WorkflowTestRun>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowTestRun> Get(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(runId, httpBearer, requestOptions, cancellationToken);
        }

        /// <summary>Cancel Test Execution Run</summary>
        /// <param name="runId">The run id.</param>
        /// <param name="httpBearer">The bearer token for authentication.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTestRun"/> result.</returns>
        public virtual async Task<WorkflowTestRun> CancelAsync(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            var request = new RetabRequest
            {
                Method = HttpMethod.Post,
                Path = $"/v1/workflows/tests/runs/{Uri.EscapeDataString(runId)}/cancel",
                AccessToken = httpBearer,
                RequestOptions = requestOptions,
            };
            return await this.Client.MakeAPIRequest<WorkflowTestRun>(request, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CancelAsync"/>.</summary>
        public virtual Task<WorkflowTestRun> Cancel(string runId, string httpBearer, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CancelAsync(runId, httpBearer, requestOptions, cancellationToken);
        }
    }
}
