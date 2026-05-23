namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ExperimentRunsService.ListAsync"/>: List Experiment Runs</summary>
    public class ExperimentRunsListOptions : ListOptions
    {
        public string? WorkflowId { get; set; }

        public string? ExperimentId { get; set; }

        public string? BlockId { get; set; }

        public LatestBlockTestRunSummaryStatus? Status { get; set; }

        public string? Statuses { get; set; }

        public LatestBlockTestRunSummaryStatus? ExcludeStatus { get; set; }

        public string? TriggerType { get; set; }

        public string? TriggerTypes { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

        public string? SortBy { get; set; }

    }

    /// <summary>Request options for <see cref="ExperimentRunsService.CreateAsync"/>: Create Experiment Run Flat</summary>
    public class ExperimentRunsCreateOptions : BaseOptions
    {
        /// <summary>The experiment to create a run for.</summary>
        public string ExperimentId { get; set; } = default!;

        /// <summary>Optional. When omitted, the workflow is derived from the experiment record. When supplied, must match the experiment's workflow_id (404 otherwise).</summary>
        public string? WorkflowId { get; set; }

    }
}
