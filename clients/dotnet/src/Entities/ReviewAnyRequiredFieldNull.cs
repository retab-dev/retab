namespace Retab
{

    /// <summary>Gate when any required field in the extract schema is null/missing.</summary>
    /// <remarks>
    /// First-class predicate because this is the most common real-world driver
    /// for structured-extraction review.
    /// </remarks>
    public class ReviewAnyRequiredFieldNull
    {
        public string? Kind { get; set; }

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
