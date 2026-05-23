namespace Retab
{

    /// <summary>Represents a workflow test run.</summary>
    public class WorkflowTestRun
    {
        public string Id { get; set; } = default!;
        public WorkflowSnapshotRef Workflow { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(ManualTriggerDiscriminatorConverter))]
        public object Trigger { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowTestRunDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;
        public WorkflowTestRunTiming Timing { get; set; } = default!;
        public WorkflowTestBlockTarget? Target { get; set; }
        public string? TestId { get; set; }
        public long TotalTests { get; set; }
        public BlockTestBatchExecutionCounts? Counts { get; set; }
    }
}
