namespace Retab
{

    /// <summary>Manual run started by a user from the dashboard.</summary>
    public class ManualTrigger
    {
        public string? Type { get; set; }

        /// <summary>User who started the run, when known</summary>
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
