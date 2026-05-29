namespace Retab
{

    /// <summary>Block-test assertion against one declared output handle.</summary>
    /// <remarks>
    /// ``target`` is the only supported shape: an output handle id and an
    /// optional relative path inside that handle's payload.
    /// </remarks>
    public class AssertionSpec
    {
        public string? Id { get; set; }
        public OutputTarget Target { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(ExistConditionDiscriminatorConverter))]
        public object Condition { get; set; } = default!;
        public string? Label { get; set; }

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
