namespace Retab
{

    /// <summary>Gate when (top1_prob - top2_prob) &lt; `margin` — model was torn.</summary>
    public class ReviewTopMarginLt
    {
        public string? Kind { get; set; }
        public double Margin { get; set; }

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
