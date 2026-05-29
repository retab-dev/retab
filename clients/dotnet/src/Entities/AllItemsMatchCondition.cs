namespace Retab
{

    /// <summary>Represents an all items match condition.</summary>
    public class AllItemsMatchCondition
    {
        public string? Kind { get; set; }
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
