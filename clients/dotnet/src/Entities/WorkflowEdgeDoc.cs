namespace Retab
{
    using System;

    /// <summary>Public live workflow edge object.</summary>
    public class WorkflowEdgeDoc
    {
        public string Id { get; set; } = default!;

        /// <summary>Foreign key to workflow</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>ID of the source block</summary>
        public string SourceBlock { get; set; } = default!;

        /// <summary>ID of the target block</summary>
        public string TargetBlock { get; set; } = default!;

        /// <summary>Output handle on source block</summary>
        public string? SourceHandle { get; set; }

        /// <summary>Input handle on target block</summary>
        public string? TargetHandle { get; set; }
        public DateTimeOffset UpdatedAt { get; set; }

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
