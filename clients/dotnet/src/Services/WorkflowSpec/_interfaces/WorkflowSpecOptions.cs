namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowSpecService.ValidateAsync"/>: Validate Workflow Spec</summary>
    public class WorkflowSpecValidateOptions : BaseOptions
    {
        /// <summary>Workflow YAML definition</summary>
        public string YamlDefinition { get; set; } = default!;

        /// <summary>Project that should own a workflow created from this spec. Required when applying a spec that creates a new workflow.</summary>
        public string? ProjectId { get; set; }

    }
}
