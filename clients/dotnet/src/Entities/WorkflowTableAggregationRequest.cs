namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a workflow table aggregation request.</summary>
    public class WorkflowTableAggregationRequest
    {
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public WorkflowTableAggregationFunction Function { get; set; }
        public string? Column { get; set; }
        public string? Alias { get; set; }

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
