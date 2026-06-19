namespace Retab
{

    /// <summary>Aggregate counts for a batch of block-eval runs.</summary>
    /// <remarks>
    /// Each individual run contributes to exactly one `lifecycle_counts`
    /// bucket, and additionally to one `outcome` bucket when
    /// `lifecycle_counts.completed` is incremented.
    /// </remarks>
    public class BlockEvalBatchExecutionCounts
    {
        public BlockEvalLifecycleCounts? LifecycleCounts { get; set; }
        public BlockEvalOutcomeCounts? Outcome { get; set; }

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
