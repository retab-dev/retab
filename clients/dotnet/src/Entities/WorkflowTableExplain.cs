namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a workflow table explain.</summary>
    public class WorkflowTableExplain
    {
        public string TableId { get; set; } = default!;
        public string? SnapshotFileId { get; set; } = "";
        public List<string>? SelectedColumns { get; set; }
        public List<WorkflowTableFilterRule>? Filters { get; set; }
        public WorkflowTableSearchRequest? Search { get; set; }
        public List<WorkflowTableSortRule>? Sort { get; set; }
        public long? Offset { get; set; }
        public long? Limit { get; set; }
        public WorkflowTableDistinctRequest? Distinct { get; set; }
        public List<string>? GroupBy { get; set; }
        public List<WorkflowTableAggregationRequest>? Aggregations { get; set; }

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
