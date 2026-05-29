namespace Retab
{

    /// <summary>Represents a classification decision.</summary>
    public class ClassificationDecision
    {

        /// <summary>The reasoning for the classification decision</summary>
        public string Reasoning { get; set; } = default!;

        /// <summary>The category name that the document belongs to</summary>
        public string Category { get; set; } = default!;

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
