namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;
    using Newtonsoft.Json;

    /// <summary>Service that exposes the workflow experiments API operations on <see cref="Retab"/>.</summary>
    public class WorkflowExperimentsService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowExperimentsService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowExperimentsService(Retab client) : base(client) { }

        /// <summary>Gets the nested <see cref="ExperimentRunMetricsService"/> service.</summary>
        public virtual ExperimentRunMetricsService Metrics => new ExperimentRunMetricsService(this.Client);

        /// <summary>Gets the nested <see cref="ExperimentRunResultsService"/> service.</summary>
        public virtual ExperimentRunResultsService Results => new ExperimentRunResultsService(this.Client);

        /// <summary>Gets the nested <see cref="ExperimentRunsService"/> service.</summary>
        public virtual ExperimentRunsService Runs => new ExperimentRunsService(this.Client);

        /// <summary>List Experiments</summary>
        /// <remarks>
        /// List experiments under one workflow with cursor pagination.
        /// The enrichment passes (latest-run snapshot, block info, drift detection)
        /// run on the paginated page, not the full collection — so they scale with
        /// ``limit``, not with the total experiment count under the workflow.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>A page of <see cref="WorkflowExperiment"/> results.</returns>
        public virtual async Task<PaginatedList<WorkflowExperiment>> ListAsync(WorkflowExperimentsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.FetchPageAsync<WorkflowExperiment>("/v1/workflows/experiments", options, null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ListAsync"/>.</summary>
        public virtual Task<PaginatedList<WorkflowExperiment>> List(WorkflowExperimentsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ListAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Auto-paging variant of <see cref="ListAsync"/>. Yields individual items across all pages.</summary>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>An async sequence of <see cref="WorkflowExperiment"/> items.</returns>
        public virtual IAsyncEnumerable<WorkflowExperiment> ListAutoPagingAsync(WorkflowExperimentsListOptions? options = null, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return base.ListAutoPagingAsync<WorkflowExperiment>("/v1/workflows/experiments", options, requestOptions, cancellationToken);
        }

        /// <summary>Create Experiment</summary>
        /// <remarks>
        /// Create an experiment.
        /// When ``source_experiment_id`` is set, duplicates the source experiment
        /// (block, name + "(Copy)", n_consensus, documents) and rejects any other
        /// field. Otherwise creates a fresh experiment from the provided fields.
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowExperiment"/> result.</returns>
        public virtual async Task<WorkflowExperiment> CreateAsync(WorkflowExperimentsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<WorkflowExperiment>("/v1/workflows/experiments", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="CreateAsync"/>.</summary>
        public virtual Task<WorkflowExperiment> Create(WorkflowExperimentsCreateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.CreateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Get Experiment</summary>
        /// <param name="experimentId">The experiment id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowExperiment"/> result.</returns>
        public virtual async Task<WorkflowExperiment> GetAsync(string experimentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<WorkflowExperiment>($"/v1/workflows/experiments/{Uri.EscapeDataString(experimentId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<WorkflowExperiment> Get(string experimentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(experimentId, requestOptions, cancellationToken);
        }

        /// <summary>Update Experiment</summary>
        /// <param name="experimentId">The experiment id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="WorkflowExperiment"/> result.</returns>
        public virtual async Task<WorkflowExperiment> UpdateAsync(string experimentId, WorkflowExperimentsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PatchAsync<WorkflowExperiment>($"/v1/workflows/experiments/{Uri.EscapeDataString(experimentId)}", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="UpdateAsync"/>.</summary>
        public virtual Task<WorkflowExperiment> Update(string experimentId, WorkflowExperimentsUpdateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.UpdateAsync(experimentId, options, requestOptions, cancellationToken);
        }

        /// <summary>Delete Experiment</summary>
        /// <param name="experimentId">The experiment id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        public virtual async Task DeleteAsync(string experimentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            await this.DeleteAsync($"/v1/workflows/experiments/{Uri.EscapeDataString(experimentId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="DeleteAsync"/>.</summary>
        public virtual Task Delete(string experimentId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.DeleteAsync(experimentId, requestOptions, cancellationToken);
        }
    }
}
