namespace Retab
{

    /// <summary>Gate when the split_by_key partition left any source page out of every partition (dropped pages).</summary>
    public class ReviewAnyPagesUnassigned
    {
        public string? Kind { get; set; } = "any_pages_unassigned";

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
