namespace Retab
{

    /// <summary>Gate when any required field in the extract schema is null or missing.</summary>
    public class ReviewAnyRequiredFieldNull
    {
        public string? Kind { get; set; } = "any_required_field_null";

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
