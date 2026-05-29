namespace Retab
{

    /// <summary>Run started by a workflow schedule.</summary>
    public class ScheduleTrigger
    {
        public string? Type { get; set; }

        /// <summary>ID of the schedule that fired this run</summary>
        public string ScheduleId { get; set; } = default!;

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
