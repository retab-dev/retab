namespace Retab
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Threading;
    using System.Threading.Tasks;

    /// <summary>Service that exposes the workflow spec API operations on <see cref="Retab"/>.</summary>
    public class WorkflowSpecService : Service
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="WorkflowSpecService"/> class bound to the
        /// supplied <paramref name="client"/>.
        /// </summary>
        /// <param name="client">The Retab API client used to make HTTP requests.</param>
        public WorkflowSpecService(Retab client) : base(client) { }

        /// <summary>Validate Workflow Spec</summary>
        /// <remarks>
        /// Validate declarative YAML without mutating workflow state.
        /// Contract:
        /// - validate, plan, and apply agree on whether a spec is acceptable: any
        /// severity=error diagnostic — whether emitted at parse time as a
        /// DeclarativeWorkflowError or bumped during compile/diagnose — raises
        /// HTTP 400 with the structured error issues
        /// - warnings do not make the document invalid; warning-only specs return
        /// HTTP 200 with `is_valid=True` and the warning issues in `diagnostics`
        /// - counts reflect the canonical compiled graph, not raw authored block count
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="DeclarativeValidationResponse"/> result.</returns>
        public virtual async Task<DeclarativeValidationResponse> ValidateAsync(WorkflowSpecValidateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<DeclarativeValidationResponse>("/v1/workflows/spec/validate", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ValidateAsync"/>.</summary>
        public virtual Task<DeclarativeValidationResponse> Validate(WorkflowSpecValidateOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ValidateAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Plan Workflow Spec</summary>
        /// <remarks>
        /// Compute the declarative reconcile plan against the current draft workflow.
        /// Contract:
        /// - plan compares authored YAML against current draft state
        /// - canonical exported YAML should plan as `noop` against the same draft
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="DeclarativePlanResponse"/> result.</returns>
        public virtual async Task<DeclarativePlanResponse> PlanAsync(WorkflowSpecPlanOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<DeclarativePlanResponse>("/v1/workflows/spec/plan", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="PlanAsync"/>.</summary>
        public virtual Task<DeclarativePlanResponse> Plan(WorkflowSpecPlanOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.PlanAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Apply Workflow Spec</summary>
        /// <remarks>
        /// Apply declarative YAML to draft workflow state.
        /// Contract:
        /// - apply writes canonical draft state, not authored formatting
        /// - re-applying canonical exported YAML against unchanged draft state should
        /// return an empty resource_changes list
        /// </remarks>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="DeclarativeApplyResponse"/> result.</returns>
        public virtual async Task<DeclarativeApplyResponse> ApplyAsync(WorkflowSpecApplyOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<DeclarativeApplyResponse>("/v1/workflows/spec/apply", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ApplyAsync"/>.</summary>
        public virtual Task<DeclarativeApplyResponse> Apply(WorkflowSpecApplyOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ApplyAsync(options, requestOptions, cancellationToken);
        }

        /// <summary>Export Workflow Spec</summary>
        /// <remarks>
        /// Export draft workflow state as canonical declarative YAML.
        /// </remarks>
        /// <param name="workflowId">The workflow id.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="DeclarativeExportResponse"/> result.</returns>
        public virtual async Task<DeclarativeExportResponse> GetAsync(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.GetAsync<DeclarativeExportResponse>($"/v1/workflows/spec/{Uri.EscapeDataString(workflowId)}", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<DeclarativeExportResponse> Get(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(workflowId, requestOptions, cancellationToken);
        }
    }
}
