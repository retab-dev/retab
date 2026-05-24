namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow artifacts API operations on <see cref="Retab"/>.</summary>
    public class WorkflowArtifactsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowArtifactsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowArtifactsService(Retab client) : base(client) { }

        /// <summary>Get Workflow Artifact By Id</summary>
        /// <remarks>
        /// Get one workflow artifact by id alone.
        /// The operation is derived from the id prefix
        /// (``extr_…`` → extraction, ``clss_…`` → classification, etc.). This is
        /// the flat-resource shape — callers do not need to know which collection
        /// backs the id.
        /// </remarks>
        /// <param name="artifactId">The artifact id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="ExtractionWorkflowArtifact"/> result.</returns>
        public virtual async Task<ExtractionWorkflowArtifact> GetAsync(string artifactId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<ExtractionWorkflowArtifact>($"/v1/workflows/artifacts/{Uri.EscapeDataString(artifactId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<ExtractionWorkflowArtifact> Get(string artifactId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(artifactId, requestOptions, cancellationToken);
        }

        /// <summary>List Workflow Artifacts</summary>
        /// <remarks>
        /// List artifacts produced by a workflow run.
        /// Paginated by the producing step's ``step_id`` (sorted by ``started_at``
        /// ascending). Pass ``after`` for the next page, ``before`` for the previous
        /// page — mutually exclusive. ``step_id`` short-circuits pagination and
        /// returns the single attached artifact.
        /// Filters: provide either ``run_id`` (list all artifacts in a run) or
        /// ``step_id`` (single-step lookup). When both are absent the request is
        /// rejected with 400.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="StepArtifactRef"/> results.</returns>
        public virtual async Task<PaginatedList<StepArtifactRef>> ListAsync(WorkflowArtifactsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<StepArtifactRef>("/v1/workflows/artifacts", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<StepArtifactRef>> List(WorkflowArtifactsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="StepArtifactRef"/> items.</returns>
        public virtual IAsyncEnumerable<StepArtifactRef> ListAutoPagingAsync(WorkflowArtifactsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<StepArtifactRef>("/v1/workflows/artifacts", options, requestOptions, cancellationToken);
        }
    }
}
