namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Record of why a while-loop block stopped iterating during a run.</summary>
    /// <remarks>
    /// Reports the `termination_reason` (`max_iterations_reached`,
    /// `condition_matched`, or `error`) and the termination conditions that were
    /// evaluated on the final iteration (`evaluations`).
    /// </remarks>
    public class WhileLoopTermination
    {

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "while_loop_termination";
        public string Id { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string StepId { get; set; } = default!;

        /// <summary>Why the while-loop terminated</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WhileLoopTerminationTerminationReason TerminationReason { get; set; }
        public List<ConditionEvaluationResult>? Evaluations { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
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
