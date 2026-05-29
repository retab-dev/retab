namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a while loop termination.</summary>
    public class WhileLoopTermination
    {

        /// <summary>Artifact operation that determines the backing record type</summary>
        public string? Operation { get; set; }
        public string Id { get; set; } = default!;
        public string WorkflowRunId { get; set; } = default!;
        public string StepId { get; set; } = default!;

        /// <summary>Why the while-loop terminated</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WhileLoopTerminationTerminationReason TerminationReason { get; set; }
        public List<ConditionEvaluationResult>? Evaluations { get; set; }

        /// <summary>When this artifact was written by the orchestrator.</summary>
        public DateTimeOffset CreatedAt { get; set; }

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
