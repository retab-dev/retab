namespace Retab
{

    /// <summary>A batch execution of a workflow's tests, with overall `lifecycle`, `timing`, and pass/fail `counts`.</summary>
    public class WorkflowTestRun
    {
        public string Id { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;
        public string WorkflowVersionId { get; set; } = default!;
        public TriggerInfo Trigger { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowTestRunDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;
        public ExperimentRunTiming Timing { get; set; } = default!;
        public WorkflowTestBlockTarget? Target { get; set; }
        public string? TestId { get; set; }
        public long TotalTests { get; set; }
        public BlockTestBatchExecutionCounts? Counts { get; set; }
        public ArtifactFreshness? Freshness { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();
    }
}
