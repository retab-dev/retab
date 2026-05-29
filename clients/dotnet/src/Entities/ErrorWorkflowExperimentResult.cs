namespace Retab
{

    /// <summary>The result row failed. Per-job error message is bundled into lifecycle.</summary>
    public class ErrorWorkflowExperimentResult
    {
        public string? Status { get; set; } = "error";

        /// <summary>Human-readable error message</summary>
        public string? Message { get; set; } = "(no message)";

        /// <summary>Structured error context including stack trace</summary>
        public ErrorDetails? Details { get; set; }

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
