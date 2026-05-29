namespace Retab
{

    /// <summary>Gate when any split boundary's confidence is below `threshold`.</summary>
    public class ReviewBoundaryConfidenceLt
    {
        public string? Kind { get; set; } = "boundary_confidence_lt";
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
