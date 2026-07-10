namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowStepsService.ListAsync"/>: List Workflow Run Steps</summary>
    public class WorkflowStepsListOptions : ListOptions
    {
        /// <summary>Optional workflow run ID filter.</summary>
        public string? RunId { get; set; }

        public string? WorkflowId { get; set; }

        /// <summary>Optional logical block ID filter.</summary>
        public string? BlockId { get; set; }

        /// <summary>Optional step ID filter.</summary>
        public string? StepId { get; set; }

        /// <summary>Optional block type filter. Repeat the query parameter for multiple values.</summary>
        public List<string>? BlockType { get; set; }

        /// <summary>Optional step lifecycle status filter. Repeat the query parameter for multiple values.</summary>
        public List<string>? Status { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowStepsService.GetAsync"/>: Get Workflow Step</summary>
    public class WorkflowStepsGetOptions : BaseOptions
    {
        /// <summary>Optional workflow run ID disambiguator.</summary>
        public string? RunId { get; set; }

    }
}
