namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents an api call invocation.</summary>
    public class ApiCallInvocation
    {

        /// <summary>Artifact operation that determines the backing record type</summary>
        public string? Operation { get; set; }
        public string Id { get; set; } = default!;
        public string WorkflowRunId { get; set; } = default!;
        public string StepId { get; set; } = default!;
        public List<ApiCallAttempt>? Attempts { get; set; }
        public ErrorDetails? Error { get; set; }

        /// <summary>When this artifact was written by the orchestrator.</summary>
        public DateTimeOffset CreatedAt { get; set; }
    }
}
