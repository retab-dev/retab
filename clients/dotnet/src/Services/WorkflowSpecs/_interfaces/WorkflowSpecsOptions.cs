namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowSpecsService.ValidateAsync"/>: Validate Workflow Spec</summary>
    public class WorkflowSpecsValidateOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowSpecsService.PlanAsync"/>: Plan Workflow Spec</summary>
    public class WorkflowSpecsPlanOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowSpecsService.ApplyAsync"/>: Apply Workflow Spec</summary>
    public class WorkflowSpecsApplyOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

    }
}
