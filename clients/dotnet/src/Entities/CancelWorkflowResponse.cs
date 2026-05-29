namespace Retab
{

    /// <summary>Response for cancel workflow endpoint.</summary>
    public class CancelWorkflowResponse
    {
        public WorkflowRun Run { get; set; } = default!;

        /// <summary>Whether immediate cancellation signaling was available</summary>
        public bool? RedisAvailable { get; set; } = true;

        /// <summary>Cancellation delivery state from this request</summary>
        public CancelWorkflowResponseCancellationStatus? CancellationStatus { get; set; }

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
