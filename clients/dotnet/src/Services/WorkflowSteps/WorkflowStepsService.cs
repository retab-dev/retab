namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the workflow steps API operations on <see cref="Retab"/>.</summary>
    public class WorkflowStepsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowStepsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowStepsService(Retab client) : base(client) { }

        /// <summary>List Workflow Run Steps</summary>
        /// <remarks>
        /// List steps with status and artifact summaries.
        /// Sorted by `started_at` ascending with `step_id` as the tiebreaker
        /// (the same compound key the underlying index uses). Pass `after` for
        /// the next page, `before` for the previous page — mutually exclusive.
        /// `run_id` is optional; when omitted the list is scoped to the caller's
        /// organization.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowRunStep"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowRunStep>> ListAsync(WorkflowStepsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowRunStep>("/v1/workflows/steps", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowRunStep>> List(WorkflowStepsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowRunStep"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowRunStep> ListAutoPagingAsync(WorkflowStepsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowRunStep>("/v1/workflows/steps", options, requestOptions, cancellationToken);
        }

        /// <summary>Get Workflow Step</summary>
        /// <remarks>
        /// Get one step by its step id.
        /// Returns the same step shape as `GET /workflows/steps`.
        /// </remarks>
        /// <param name="stepId">The step id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowRunStep"/> result.</returns>
        public virtual async Task<WorkflowRunStep> GetAsync(string stepId, WorkflowStepsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowRunStep>($"/v1/workflows/steps/{Uri.EscapeDataString(stepId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowRunStep> Get(string stepId, WorkflowStepsGetOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(stepId, options, requestOptions, cancellationToken);
        }
    }
}
