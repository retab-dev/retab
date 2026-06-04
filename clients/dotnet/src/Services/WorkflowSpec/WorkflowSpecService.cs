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

        /// <summary>Apply Workflow Spec</summary>
        /// <remarks>
        /// Create a new workflow from a declarative YAML spec.
        /// The workflow id in the YAML is treated as source context, not as the target
        /// workflow id. Use `POST /v1/workflows/{workflow_id}/spec/apply` to modify an
        /// existing workflow draft.
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

        /// <summary>Plan Workflow Spec</summary>
        /// <remarks>
        /// Preview the changes a declarative YAML spec would make to the draft workflow.
        /// Compares the spec against the current draft and returns the resulting
        /// changes without applying them. A spec that already matches the draft
        /// plans as a no-op.
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

        /// <summary>Validate Workflow Spec</summary>
        /// <remarks>
        /// Validate declarative YAML without changing the workflow.
        /// Any error-level diagnostic responds with 400 and the structured issues.
        /// Warnings do not make a spec invalid: a warning-only spec responds with
        /// 200, `is_valid=True`, and the warnings in `diagnostics`.
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
            return await this.GetAsync<DeclarativeExportResponse>($"/v1/workflows/{Uri.EscapeDataString(workflowId)}/spec", null, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="GetAsync"/>.</summary>
        public virtual Task<DeclarativeExportResponse> Get(string workflowId, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.GetAsync(workflowId, requestOptions, cancellationToken);
        }

        /// <summary>Apply Workflow Spec To Existing Workflow</summary>
        /// <remarks>
        /// Apply a declarative YAML spec to an existing workflow draft.
        /// The URL workflow id is the update target. Any workflow id in the YAML is
        /// treated as source context.
        /// </remarks>
        /// <param name="workflowId">The workflow id.</param>
        /// <param name="options">Request options.</param>
        /// <param name="requestOptions">Per-request configuration overrides.</param>
        /// <param name="cancellationToken">Cancellation token.</param>
        /// <returns>The <see cref="DeclarativeApplyResponse"/> result.</returns>
        public virtual async Task<DeclarativeApplyResponse> ApplyToWorkflowAsync(string workflowId, WorkflowSpecApplyToWorkflowOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return await this.PostAsync<DeclarativeApplyResponse>($"/v1/workflows/{Uri.EscapeDataString(workflowId)}/spec/apply", options, requestOptions, cancellationToken);
        }

        /// <summary>Compatibility wrapper for <see cref="ApplyToWorkflowAsync"/>.</summary>
        public virtual Task<DeclarativeApplyResponse> ApplyToWorkflow(string workflowId, WorkflowSpecApplyToWorkflowOptions options, RequestOptions? requestOptions = null, CancellationToken cancellationToken = default)
        {
            return this.ApplyToWorkflowAsync(workflowId, options, requestOptions, cancellationToken);
        }
    }
}
