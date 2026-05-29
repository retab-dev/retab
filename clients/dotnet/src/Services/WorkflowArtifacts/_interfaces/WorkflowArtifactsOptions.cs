namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowArtifactsService.ListAsync"/>: List Workflow Artifacts</summary>
    public class WorkflowArtifactsListOptions : ListOptions
    {
        /// <summary>Workflow run ID whose artifacts should be listed. Required unless `step_id` is provided.</summary>
        public string? RunId { get; set; }

        /// <summary>Optional artifact operation filter</summary>
        public StepArtifactRefOperation? Operation { get; set; }

        /// <summary>Optional block_id or step_id filter</summary>
        public string? BlockId { get; set; }

        /// <summary>Optional step id filter. When provided, returns the single artifact attached to that step (or an empty list if the step has no artifact). `run_id` is not required when `step_id` is set — it is resolved from the step record.</summary>
        public string? StepId { get; set; }

    }
}
