namespace Retab
{

    /// <summary>Terminal: the block declared its inputs unsatisfied via</summary>
    /// <remarks>
    /// ``should_skip_block`` and was skipped. ``reason`` is the skip rationale
    /// surfaced by the block's input requirements registry.
    /// </remarks>
    public class SkippedBlockExecutionLifecycle
    {
        public string? Status { get; set; } = "skipped";

        /// <summary>Reason the block was skipped</summary>
        public string Reason { get; set; } = default!;

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
