namespace Retab
{

    /// <summary>The eval run failed. The error message lives on this variant.</summary>
    /// <remarks>
    /// Carries the same structured `details` envelope as workflow runs so
    /// consumers can branch on `error_code` / `stage` rather than parsing
    /// a free-text message.
    /// </remarks>
    public class ErrorWorkflowEvalRun
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
