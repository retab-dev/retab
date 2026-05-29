namespace Retab
{

    /// <summary>Represents an any item matches condition.</summary>
    public class AnyItemMatchesCondition
    {
        public string? Kind { get; set; } = "any_item_matches";
        [Newtonsoft.Json.JsonConverter(typeof(ExistConditionDiscriminatorConverter))]
        public object Condition { get; set; } = default!;

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
