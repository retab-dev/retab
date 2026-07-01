namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowExperimentsService.ListAsync"/>: List Experiments</summary>
    public class WorkflowExperimentsListOptions : ListOptions
    {
        public string WorkflowId { get; set; } = default!;

        public string? BlockId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowExperimentsService.CreateAsync"/>: Create Experiment</summary>
    public class WorkflowExperimentsCreateOptions : BaseOptions
    {
        public string WorkflowId { get; set; } = default!;

        public string? BlockId { get; set; }

        public List<ExperimentDocumentCaptureRequest>? DocumentCaptures { get; set; }

        public List<ExplicitExperimentDocumentRequest>? Documents { get; set; }

        public CreateExperimentRequestNConsensus? NConsensus { get; set; }

        public string? Name { get; set; }

        public string? SourceExperimentId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowExperimentsService.UpdateAsync"/>: Update Experiment</summary>
    public class WorkflowExperimentsUpdateOptions : BaseOptions
    {
        public List<ExperimentDocumentCaptureRequest>? DocumentCaptures { get; set; }

        public List<ExplicitExperimentDocumentRequest>? Documents { get; set; }

        public CreateExperimentRequestNConsensus? NConsensus { get; set; }

        public string? Name { get; set; }

    }
}
