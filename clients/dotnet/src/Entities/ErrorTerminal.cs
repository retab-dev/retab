namespace Retab
{

    /// <summary>The run failed. All loose error fields are bundled here.</summary>
    public class ErrorTerminal
    {
        public string? Status { get; set; }

        /// <summary>Human-readable error message</summary>
        public string Message { get; set; } = default!;

        /// <summary>Which execution stage failed</summary>
        public ErrorStepLifecycleStage? Stage { get; set; }

        /// <summary>Error category for retry decisions</summary>
        public ErrorStepLifecycleCategory? Category { get; set; }

        /// <summary>Detailed error context including stack trace</summary>
        public ErrorDetails? Details { get; set; }

        /// <summary>Step ID of the failing step, when the failure was attributable to a specific step</summary>
        public string? FailingStepId { get; set; }

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
