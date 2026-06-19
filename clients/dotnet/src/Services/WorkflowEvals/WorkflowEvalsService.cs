namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow evals API operations on <see cref="Retab"/>.</summary>
    public class WorkflowEvalsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowEvalsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowEvalsService(Retab client) : base(client) { }

        /// <summary>Gets the nested <see cref="WorkflowEvalRunResultsService"/> service.</summary>
        public virtual WorkflowEvalRunResultsService Results => new WorkflowEvalRunResultsService(this.Client);

        /// <summary>Gets the nested <see cref="WorkflowEvalRunsService"/> service.</summary>
        public virtual WorkflowEvalRunsService Runs => new WorkflowEvalRunsService(this.Client);

        /// <summary>List Workflow Evals</summary>
        /// <remarks>
        /// List workflow evals.
        /// Requires `workflow_id` and returns its saved evals as a cursor-paginated
        /// list, each with its latest-run summaries and drift status. Optionally
        /// filter to one block with `target_block_id`. Returns 404 if the workflow
        /// does not exist.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowEval"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowEval>> ListAsync(WorkflowEvalsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowEval>("/v1/workflows/evals", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowEval>> List(WorkflowEvalsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowEval"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowEval> ListAutoPagingAsync(WorkflowEvalsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowEval>("/v1/workflows/evals", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Workflow Eval</summary>
        /// <remarks>
        /// Create a workflow eval.
        /// Pins an expected outcome for one block in a workflow. Provide the
        /// `workflow_id`, the `target` block, an `assertion` describing the expected
        /// output, and a `source` of eval inputs (explicit handle inputs or a capture
        /// from a prior run/step). Returns the created eval with status 201.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEval"/> result.</returns>
        public virtual async Task<WorkflowEval> CreateAsync(WorkflowEvalsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowEval>("/v1/workflows/evals", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowEval> Create(WorkflowEvalsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow Eval</summary>
        /// <remarks>
        /// Retrieve a single workflow eval.
        /// Identified by `eval_id`. Returns the eval with its target block,
        /// assertion, input source, and latest-run summaries. Returns 404 if no eval
        /// with that ID exists.
        /// </remarks>
        /// <param name="evalId">The eval id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEval"/> result.</returns>
        public virtual async Task<WorkflowEval> GetAsync(string evalId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowEval>($"/v1/workflows/evals/{Uri.EscapeDataString(evalId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowEval> Get(string evalId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(evalId, requestOptions, cancellationToken);
        }

        /// <summary>Update Workflow Eval</summary>
        /// <remarks>
        /// Update a workflow eval.
        /// Identified by `eval_id`. Send any of `name`, `assertion`, or `source`;
        /// omitted fields are left unchanged. Returns the updated eval.
        /// </remarks>
        /// <param name="evalId">The eval id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowEval"/> result.</returns>
        public virtual async Task<WorkflowEval> UpdateAsync(string evalId, WorkflowEvalsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PatchAsync<WorkflowEval>($"/v1/workflows/evals/{Uri.EscapeDataString(evalId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<WorkflowEval> Update(string evalId, WorkflowEvalsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(evalId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Workflow Eval</summary>
        /// <remarks>
        /// Delete a workflow eval.
        /// Identified by `eval_id`. Returns 204 on success and 404 if no eval with
        /// that ID exists.
        /// </remarks>
        /// <param name="evalId">The eval id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string evalId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/workflows/evals/{Uri.EscapeDataString(evalId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string evalId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(evalId, requestOptions, cancellationToken);
        }
    }
}
