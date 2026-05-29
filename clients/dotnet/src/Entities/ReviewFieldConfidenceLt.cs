namespace Retab
{

    /// <summary>Gate when the field at `path` has confidence below `threshold`.</summary>
    public class ReviewFieldConfidenceLt
    {
        public string? Kind { get; set; }

        /// <summary>JSONPath-style path, e.g. '$.invoice.total' or 'invoice.total'</summary>
        public string Path { get; set; } = default!;
        public double Threshold { get; set; }

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
