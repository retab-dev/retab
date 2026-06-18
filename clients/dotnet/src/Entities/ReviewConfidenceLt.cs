namespace Retab
{

    /// <summary>Gate if the block consensus likelihood is below `threshold`.</summary>
    public class ReviewConfidenceLt
    {
        public string? Kind { get; set; } = "confidence_lt";

        /// <summary>Gate fires when consensus likelihood &lt; threshold</summary>
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
