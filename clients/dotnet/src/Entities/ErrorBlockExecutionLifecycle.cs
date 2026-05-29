namespace Retab
{

    /// <summary>Terminal: the executed block raised. `message` is the executor's</summary>
    /// <remarks>
    /// error string.
    /// </remarks>
    public class ErrorBlockExecutionLifecycle
    {
        public string? Status { get; set; } = "error";

        /// <summary>Human-readable error message</summary>
        public string Message { get; set; } = default!;

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
