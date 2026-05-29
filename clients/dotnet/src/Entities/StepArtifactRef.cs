namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Canonical persisted resource produced by a workflow step.</summary>
    /// <remarks>
    /// Uniformly a `(operation, id)` ref into a backing collection. The artifact
    /// itself carries no payload — consumers dispatch on ``operation`` and fetch
    /// the backing record by ``id``.
    /// </remarks>
    public class StepArtifactRef
    {

        /// <summary>Persisted resource operation; identifies the backing collection</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public StepArtifactRefOperation Operation { get; set; }

        /// <summary>Persisted resource identifier</summary>
        public string Id { get; set; } = default!;

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
