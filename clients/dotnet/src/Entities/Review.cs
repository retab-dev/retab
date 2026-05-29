namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>One review and its current decision.</summary>
    public class Review
    {
        public string Id { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;
        public string WorkflowVersionId { get; set; } = default!;
        public string RunId { get; set; } = default!;
        public string BlockId { get; set; } = default!;
        public string StepId { get; set; } = default!;
        public string? ParentStepId { get; set; }
        public string? IterationKey { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public ExperimentBlockType BlockType { get; set; }
        [Newtonsoft.Json.JsonConverter(typeof(ReviewAlwaysDiscriminatorConverter))]
        public object TriggeredBy { get; set; } = default!;

        /// <summary>When the review was created.</summary>
        public DateTimeOffset CreatedAt { get; set; }
        public ReviewDecision? Decision { get; set; }

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
