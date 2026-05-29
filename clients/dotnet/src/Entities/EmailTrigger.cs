namespace Retab
{

    /// <summary>Run started by an inbound email.</summary>
    public class EmailTrigger
    {
        public string? Type { get; set; }

        /// <summary>Sender email address, when known</summary>
        public string? Sender { get; set; }

        /// <summary>Email subject, when known</summary>
        public string? Subject { get; set; }

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
