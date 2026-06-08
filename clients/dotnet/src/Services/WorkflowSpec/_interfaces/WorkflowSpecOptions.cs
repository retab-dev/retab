namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowSpecService.ApplyAsync"/>: Apply Workflow Spec</summary>
    public class WorkflowSpecApplyOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

        /// <summary>Project that should own a workflow created from this spec. Required when applying a spec that creates a new workflow.</summary>
        public string? ProjectId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowSpecService.PlanAsync"/>: Plan Workflow Spec</summary>
    public class WorkflowSpecPlanOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

        /// <summary>Project that should own a workflow created from this spec. Required when applying a spec that creates a new workflow.</summary>
        public string? ProjectId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowSpecService.ValidateAsync"/>: Validate Workflow Spec</summary>
    public class WorkflowSpecValidateOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

        /// <summary>Project that should own a workflow created from this spec. Required when applying a spec that creates a new workflow.</summary>
        public string? ProjectId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowSpecService.ApplyToWorkflowAsync"/>: Apply Existing Workflow Spec</summary>
    public class WorkflowSpecApplyToWorkflowOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

        /// <summary>Project that should own a workflow created from this spec. Required when applying a spec that creates a new workflow.</summary>
        public string? ProjectId { get; set; }

    }
}
