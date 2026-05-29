namespace Retab
{

    /// <summary>Run created by restarting a parent run.</summary>
    public class RestartTrigger
    {
        public string? Type { get; set; }

        /// <summary>ID of the parent run that was restarted</summary>
        public string ParentRunId { get; set; } = default!;

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
