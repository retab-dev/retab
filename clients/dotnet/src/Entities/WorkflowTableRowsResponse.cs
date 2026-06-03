namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a workflow table rows response.</summary>
    public class WorkflowTableRowsResponse
    {
        public string TableId { get; set; } = default!;
        public List<WorkflowTableColumn>? Columns { get; set; }
        public List<WorkflowTableRow>? Rows { get; set; }
        public long RowCount { get; set; }
        public long? FilteredRowCount { get; set; }
        public long? Offset { get; set; }
        public long? Limit { get; set; }
        public bool? HasMore { get; set; } = false;
        public string? NextCursor { get; set; }
        public string? PreviousCursor { get; set; }
        public WorkflowTableExplain? Explain { get; set; }

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
