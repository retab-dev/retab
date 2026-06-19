namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Summary of the most recent block-eval run.</summary>
    /// <remarks>
    /// Execution status and verdict outcome are exposed as separate fields.
    /// The summary is written on terminal-state transitions, so in practice
    /// `status` is one of `completed | error | cancelled` and `outcome` is
    /// populated when `status == "completed"`.
    /// </remarks>
    public class LatestBlockEvalRunSummary
    {
        public string RunRecordId { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public LatestBlockEvalRunSummaryStatus Status { get; set; }
        public AssertionOutcome? Outcome { get; set; }
        public DateTimeOffset StartedAt { get; set; }
        public DateTimeOffset? CompletedAt { get; set; }
        public long? DurationMs { get; set; }
        public string? WorkflowDraftFingerprint { get; set; }
        public string? BlockConfigFingerprint { get; set; }
        public string? ValidityFingerprint { get; set; }
        public string? HandleInputsFingerprint { get; set; }
        public long? AssertionsPassed { get; set; }
        public long? AssertionsFailed { get; set; }
        public long? BlockedAssertions { get; set; }

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
