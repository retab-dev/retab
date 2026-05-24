namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow tests API operations on <see cref="Retab"/>.</summary>
    public class WorkflowTestsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowTestsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowTestsService(Retab client) : base(client) { }

        /// <summary>Gets the nested <see cref="WorkflowTestRunResultsService"/> service.</summary>
        public virtual WorkflowTestRunResultsService Results => new WorkflowTestRunResultsService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowTestRunsService"/> service.</summary>
        public virtual WorkflowTestRunsService Runs => new WorkflowTestRunsService(this.Client);

        /// <summary>List Tests</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowTest"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowTest>> ListAsync(WorkflowTestsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowTest>("/v1/workflows/tests", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowTest>> List(WorkflowTestsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowTest"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowTest> ListAutoPagingAsync(WorkflowTestsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowTest>("/v1/workflows/tests", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Test</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTest"/> result.</returns>
        public virtual async Task<WorkflowTest> CreateAsync(WorkflowTestsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowTest>("/v1/workflows/tests", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowTest> Create(WorkflowTestsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Test</summary>
        /// <param name="testId">The test id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTest"/> result.</returns>
        public virtual async Task<WorkflowTest> GetAsync(string testId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowTest>($"/v1/workflows/tests/{Uri.EscapeDataString(testId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowTest> Get(string testId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(testId, requestOptions, cancellationToken);
        }

        /// <summary>Update Test</summary>
        /// <param name="testId">The test id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTest"/> result.</returns>
        public virtual async Task<WorkflowTest> UpdateAsync(string testId, WorkflowTestsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PatchAsync<WorkflowTest>($"/v1/workflows/tests/{Uri.EscapeDataString(testId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<WorkflowTest> Update(string testId, WorkflowTestsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(testId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Test</summary>
        /// <param name="testId">The test id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string testId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/workflows/tests/{Uri.EscapeDataString(testId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string testId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(testId, requestOptions, cancellationToken);
        }
    }
}
