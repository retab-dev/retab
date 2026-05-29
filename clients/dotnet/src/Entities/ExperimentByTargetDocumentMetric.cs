namespace Retab
{

    /// <summary>Compact document score returned by the per-target metrics view.</summary>
    public class ExperimentByTargetDocumentMetric
    {
        public string Id { get; set; } = default!;
        public string Filename { get; set; } = default!;
        public double? Score { get; set; }
        public double? PriorScore { get; set; }
        public object? Value { get; set; }

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
