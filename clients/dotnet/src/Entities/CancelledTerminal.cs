namespace Retab
{

    /// <summary>The run was cancelled before reaching a natural terminal state.</summary>
    public class CancelledTerminal
    {
        public string? Status { get; set; }

        /// <summary>Human-readable reason, when known</summary>
        public string? Reason { get; set; }

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
