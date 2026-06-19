namespace Retab
{

    /// <summary>Per-lifecycle counts for a batch of block-eval runs.</summary>
    public class BlockEvalLifecycleCounts
    {
        public long? Pending { get; set; }
        public long? Queued { get; set; }
        public long? Running { get; set; }
        public long? Completed { get; set; }
        public long? Error { get; set; }
        public long? Cancelled { get; set; }

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
