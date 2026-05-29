namespace Retab
{

    /// <summary>Represents a cancel workflow experiment run response.</summary>
    public class CancelWorkflowExperimentRunResponse
    {
        public string Id { get; set; } = default!;
        [Newtonsoft.Json.JsonConverter(typeof(PendingWorkflowExperimentRunDiscriminatorConverter))]
        public object Lifecycle { get; set; } = default!;

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
