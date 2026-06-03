namespace Retab
{

    /// <summary>Represents a validate workflow block config response.</summary>
    public class ValidateWorkflowBlockConfigResponse
    {
        public bool? Ok { get; set; } = true;
        public string WorkflowId { get; set; } = default!;
        public string BlockId { get; set; } = default!;
        public string BlockType { get; set; } = default!;
        public string ConfigHash { get; set; } = default!;

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
