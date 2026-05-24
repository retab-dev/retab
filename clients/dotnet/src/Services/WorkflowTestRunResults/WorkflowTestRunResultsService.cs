namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow test run results API operations on <see cref="Retab"/>.</summary>
    public class WorkflowTestRunResultsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowTestRunResultsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowTestRunResultsService(Retab client) : base(client) { }

        /// <summary>List Test Execution Results</summary>
        /// <remarks>
        /// List workflow test results for a single run, page by page.
        /// Pagination strategy: the parent
        /// ``workflow_test_runs.result_run_record_ids`` document already holds the
        /// ordered list of child record IDs. ``workflow_block_test_runs`` rows do
        /// not carry a ``run_id`` field (the relationship lives only on the
        /// parent), so a direct keyset query on the child collection is not
        /// possible without a schema change. We slice the parent's ordered list to
        /// resolve cursors and then ``$in``-query the child collection for only
        /// the requested page — preserving the run-time ordering encoded in the
        /// parent doc and avoiding a fan-out collection scan.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowTestResult"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowTestResult>> ListAsync(WorkflowTestRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowTestResult>("/v1/workflows/tests/results", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowTestResult>> List(WorkflowTestRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowTestResult"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowTestResult> ListAutoPagingAsync(WorkflowTestRunResultsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowTestResult>("/v1/workflows/tests/results", options, requestOptions, cancellationToken);
        }

        /// <summary>Get Test Execution Result</summary>
        /// <param name="resultId">The result id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowTestResult"/> result.</returns>
        public virtual async Task<WorkflowTestResult> GetAsync(string resultId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowTestResult>($"/v1/workflows/tests/results/{Uri.EscapeDataString(resultId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowTestResult> Get(string resultId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(resultId, requestOptions, cancellationToken);
        }
    }
}
