namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowTestRunsService.ListAsync"/>: List Test Execution Runs</summary>
    public class WorkflowTestRunsListOptions : ListOptions
    {
        public string? WorkflowId { get; set; }

        public string? TestId { get; set; }

        public string? TargetBlockId { get; set; }

        public string? Status { get; set; }

        public string? ExcludeStatus { get; set; }

        public string? TriggerType { get; set; }

        public DateTimeOffset? FromDate { get; set; }

        public DateTimeOffset? ToDate { get; set; }

        public string? SortBy { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowTestRunsService.CreateAsync"/>: Create Test Run</summary>
    public class WorkflowTestRunsCreateOptions : BaseOptions
    {
        public string WorkflowId { get; set; } = default!;

        /// <summary>Optional execution scope. Omit (or pass null) to run every saved test in the workflow.</summary>
        public object? Scope { get; set; }

    }
}
