namespace Retab
{

    /// <summary>Represents a similarity gte condition.</summary>
    public class SimilarityGteCondition
    {
        public string? Kind { get; set; }
        public object Reference { get; set; } = default!;
        public double Threshold { get; set; }
        public SimilarityGteConditionMethod? Method { get; set; }

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
