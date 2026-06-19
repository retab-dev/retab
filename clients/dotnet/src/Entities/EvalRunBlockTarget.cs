namespace Retab
{

    /// <summary>Public workflow-eval target.</summary>
    /// <remarks>
    /// The storage layer remains block-scoped today, but the API shape names the
    /// tested entity explicitly so workflow-level targets can be added later.
    /// </remarks>
    public class EvalRunBlockTarget
    {
        public string? Type { get; set; } = "block";
        public string BlockId { get; set; } = default!;

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
