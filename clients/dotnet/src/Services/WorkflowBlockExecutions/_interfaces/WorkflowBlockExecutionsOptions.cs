namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowBlockExecutionsService.ListAsync"/>: List Block Executions</summary>
    public class WorkflowBlockExecutionsListOptions : ListOptions
    {
        public string RunId { get; set; } = default!;

        public string BlockId { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="WorkflowBlockExecutionsService.CreateAsync"/>: Create Block Execution</summary>
    public class WorkflowBlockExecutionsCreateOptions : BaseOptions
    {
        /// <summary>Workflow run id that owns the step.</summary>
        public string RunId { get; set; } = default!;

        /// <summary>Workflow block id to execute.</summary>
        public string BlockId { get; set; } = default!;

        /// <summary>Optional concrete step id whose inputs should be used. When omitted, the block id is used to look up the step.</summary>
        public string? StepId { get; set; }

        /// <summary>Optional override for n_consensus on extract / split / classifier blocks. Must be 3, 5, or 7.</summary>
        public long? NConsensus { get; set; }

        /// <summary>Whether to verify the upstream subgraph hasn't drifted since the source run. Disable only for explicit force-rerun flows.</summary>
        public bool? CheckEligibility { get; set; }

    }
}
