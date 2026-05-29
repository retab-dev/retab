namespace Retab
{

    /// <summary>Denormalized counts surface, split along the canonical axes.</summary>
    /// <remarks>
    /// Each individual run contributes to exactly one ``lifecycle_counts``
    /// bucket, and additionally to one ``outcome`` bucket when
    /// ``lifecycle_counts.completed`` is incremented.
    /// The ``lifecycle_counts`` name disambiguates from the API_DESIGN.md
    /// ``lifecycle`` convention (which signals a discriminated union of
    /// typed states). This field is a counts subdocument, not a union.
    /// </remarks>
    public class BlockTestBatchExecutionCounts
    {
        public BlockTestLifecycleCounts? LifecycleCounts { get; set; }
        public BlockTestOutcomeCounts? Outcome { get; set; }

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
