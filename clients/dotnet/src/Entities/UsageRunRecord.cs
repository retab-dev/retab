namespace Retab
{
    using System;

    /// <summary>Represents an usage run record.</summary>
    public class UsageRunRecord
    {
        public DateTimeOffset? CompletedAt { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }
        public double Credits { get; set; }
        public long? DurationMs { get; set; }
        public long ExecutionDurationMs { get; set; }
        public long PageCount { get; set; }
        public long RetryCount { get; set; }
        public string RunId { get; set; } = default!;
        public DateTimeOffset? StartedAt { get; set; }
        public string Status { get; set; } = default!;
        public string TriggerType { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;

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
