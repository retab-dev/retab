namespace Retab
{

    /// <summary>Represents a workflow config edge.</summary>
    public class WorkflowConfigEdge
    {
        public string? Id { get; set; }

        /// <summary>ID of the source block</summary>
        public string Source { get; set; } = default!;

        /// <summary>ID of the target block</summary>
        public string Target { get; set; } = default!;
        public string? SourceHandle { get; set; }
        public string? TargetHandle { get; set; }
        public bool? Animated { get; set; } = true;

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
