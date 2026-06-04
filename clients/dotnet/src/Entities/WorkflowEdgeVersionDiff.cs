namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a workflow edge version diff.</summary>
    public class WorkflowEdgeVersionDiff
    {
        public string FromEdgeVersionId { get; set; } = default!;
        public string ToEdgeVersionId { get; set; } = default!;
        public string EdgeId { get; set; } = default!;
        public List<WorkflowVersionFieldDiff>? Changes { get; set; }

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
