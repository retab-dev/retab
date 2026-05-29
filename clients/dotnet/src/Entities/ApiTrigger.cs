namespace Retab
{

    /// <summary>Run started programmatically via the public API.</summary>
    public class ApiTrigger
    {
        public string? Type { get; set; } = "api";

        /// <summary>API key id used to start the run, when known</summary>
        public string? ApiKeyId { get; set; }

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
