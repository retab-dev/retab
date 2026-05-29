namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Public summary of what started a run: just the trigger category.</summary>
    /// <remarks>
    /// The full per-variant detail (schedule_id, parent_run_id, sender, ...) is
    /// kept internally on `StoredWorkflowRun.trigger` but intentionally not
    /// exposed in the public API surface.
    /// </remarks>
    public class TriggerInfo
    {

        /// <summary>What started this run</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public TriggerInfoType Type { get; set; }

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
