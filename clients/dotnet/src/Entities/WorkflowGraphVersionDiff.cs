namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a workflow graph version diff.</summary>
    public class WorkflowGraphVersionDiff
    {
        public string FromWorkflowVersionId { get; set; } = default!;
        public string ToWorkflowVersionId { get; set; } = default!;
        public List<string>? AddedBlockIds { get; set; }
        public List<string>? RemovedBlockIds { get; set; }
        public List<string>? ChangedBlockIds { get; set; }
        public List<string>? AddedEdgeIds { get; set; }
        public List<string>? RemovedEdgeIds { get; set; }
        public List<string>? ChangedEdgeIds { get; set; }

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
