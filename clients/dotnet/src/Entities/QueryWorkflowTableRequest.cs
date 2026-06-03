namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a query workflow table request.</summary>
    public class QueryWorkflowTableRequest
    {
        public List<WorkflowTableFilterRule>? Filters { get; set; }
        public WorkflowTableSearchRequest? Search { get; set; }
        public bool? CaseSensitive { get; set; } = false;
        public List<string>? Select { get; set; }
        public WorkflowTableDistinctRequest? Distinct { get; set; }
        public List<string>? GroupBy { get; set; }
        public List<WorkflowTableAggregationRequest>? Aggregations { get; set; }
        public List<WorkflowTableSortRule>? Sort { get; set; }
        public WorkflowTableSampleRequest? Sample { get; set; }
        public WorkflowTableSampleRequest? Tail { get; set; }
        public bool? CountOnly { get; set; } = false;
        public bool? IncludeExplain { get; set; } = false;
        public string? SortColumn { get; set; }
        public ClassificationsOrder? SortDirection { get; set; }
        public string? ViewerMode { get; set; }
        public long? Offset { get; set; }
        public long? Limit { get; set; }

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
