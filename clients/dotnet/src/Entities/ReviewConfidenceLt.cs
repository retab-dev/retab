namespace Retab
{

    /// <summary>Gate if the overall block confidence is below `threshold`.</summary>
    /// <remarks>
    /// Note: LLM confidences are poorly calibrated; per-field confidence
    /// (ReviewFieldConfidenceLt) tends to behave better.
    /// </remarks>
    public class ReviewConfidenceLt
    {
        public string? Kind { get; set; } = "confidence_lt";

        /// <summary>Gate fires when confidence &lt; threshold</summary>
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
