namespace Retab
{

    /// <summary>Run started by an inbound webhook.</summary>
    public class WebhookTrigger
    {
        public string? Type { get; set; } = "webhook";

        /// <summary>ID of the webhook configuration, when known</summary>
        public string? WebhookId { get; set; }

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
