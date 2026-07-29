namespace Retab
{

    /// <summary>Represents a V1 usage usage run triggered by.</summary>
    public class V1UsageUsageRunTriggeredBy
    {
        public string? ApiKeyId { get; set; }
        public string? UserEmail { get; set; }
        public string? UserId { get; set; }

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
