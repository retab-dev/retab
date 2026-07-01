namespace Retab
{

    /// <summary>A batch execution of a workflow's evals, with overall `lifecycle`, `timing`, and pass/fail `counts`.</summary>
    public class WorkflowEvalRun
    {
        public string Id { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;
        public string WorkflowVersionId { get; set; } = default!;
        public EvalRunTrigger Trigger { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowEvalRunDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;
        public ExperimentRunTiming Timing { get; set; } = default!;
        public EvalRunBlockTarget? Target { get; set; }
        public string? EvalId { get; set; }
        public long TotalEvals { get; set; }
        public BlockEvalBatchExecutionCounts? Counts { get; set; }

        /// <summary>Compatibility envelope only. WorkflowEval.freshness is the authoritative read-time staleness verdict for saved eval definitions.</summary>
        public EvalRunFreshness? Freshness { get; set; }

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
