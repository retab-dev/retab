namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="WorkflowRunsService.ListAsync"/>: List Workflow Runs</summary>
    public class WorkflowRunsListOptions : ListOptions
    {
        /// <summary>Filter by workflow ID</summary>
        public string? WorkflowId { get; set; }

        /// <summary>Filter by single run status (deprecated, use 'statuses')</summary>
        public WorkflowExportPayloadRequestExcludeStatus? Status { get; set; }

        /// <summary>Filter by multiple statuses (comma-separated: pending,queued,running,completed,error,failed,awaiting_review,cancelled)</summary>
        public string? Statuses { get; set; }

        /// <summary>Exclude runs with this status</summary>
        public WorkflowExportPayloadRequestExcludeStatus? ExcludeStatus { get; set; }

        /// <summary>Filter by single trigger type (deprecated, use 'trigger_types')</summary>
        public WorkflowExportPayloadRequestTriggerTypes? TriggerType { get; set; }

        /// <summary>Filter by multiple trigger types (comma-separated: manual,api,schedule,webhook,restart)</summary>
        public string? TriggerTypes { get; set; }

        /// <summary>Filter runs created on or after this date (YYYY-MM-DD)</summary>
        public string? FromDate { get; set; }

        /// <summary>Filter runs created on or before this date (YYYY-MM-DD)</summary>
        public string? ToDate { get; set; }

        /// <summary>Filter runs with duration &gt;= this value in milliseconds</summary>
        public long? MinDurationMs { get; set; }

        /// <summary>Filter runs with duration &lt;= this value in milliseconds</summary>
        public long? MaxDurationMs { get; set; }

        /// <summary>Search by run ID (partial match)</summary>
        public string? Search { get; set; }

        public string? SortBy { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowRunsService.CreateAsync"/>: Create Workflow Run Route</summary>
    public class WorkflowRunsCreateOptions : BaseOptions
    {
        /// <summary>Workflow id for the fresh run.</summary>
        public string? WorkflowId { get; set; }

        /// <summary>Mapping of start_document block IDs to their input documents. Only valid for fresh-run creation (``restart_of`` is None).</summary>
        public Dictionary<string, MimeData>? Documents { get; set; }

        /// <summary>Mapping of start-json block IDs to their input JSON data. Only valid for fresh-run creation (``restart_of`` is None).</summary>
        public Dictionary<string, object>? JsonInputs { get; set; }

        /// <summary>Workflow version to run: 'production', 'draft', or a pinned version id like 'ver_...'. Only valid for fresh-run creation.</summary>
        public string? Version { get; set; }

        /// <summary>When present, the new run is created as a restart of this source run id (the source run's inputs are inherited).</summary>
        public string? RestartOf { get; set; }

        /// <summary>Required when ``restart_of`` is set. Config source for the restarted run.</summary>
        public string? ConfigSource { get; set; }

        /// <summary>Optional idempotency key for deduplicating restart commands. Only valid when ``restart_of`` is set.</summary>
        public string? CommandId { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowRunsService.ExportAsync"/>: Get Workflow Export Payload</summary>
    public class WorkflowRunsExportOptions : BaseOptions
    {
        /// <summary>Workflow ID to export</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Block ID to export</summary>
        public string BlockId { get; set; } = default!;

        /// <summary>Use block outputs or inputs</summary>
        public WorkflowExportPayloadRequestExportSource? ExportSource { get; set; }

        /// <summary>Run IDs filter (null means all runs)</summary>
        public List<string>? SelectedRunIds { get; set; }

        /// <summary>Doc type filter (null/empty means all)</summary>
        public List<string>? SelectedDocTypes { get; set; }

        /// <summary>Optional status filter (intersects with completed-only export scope)</summary>
        public WorkflowExportPayloadRequestExcludeStatus? Status { get; set; }

        /// <summary>Optional status exclusion filter (intersects with completed-only export scope)</summary>
        public WorkflowExportPayloadRequestExcludeStatus? ExcludeStatus { get; set; }

        /// <summary>Optional start date filter (YYYY-MM-DD)</summary>
        public string? FromDate { get; set; }

        /// <summary>Optional end date filter (YYYY-MM-DD)</summary>
        public string? ToDate { get; set; }

        /// <summary>Optional trigger type filters</summary>
        public List<WorkflowExportPayloadRequestTriggerTypes>? TriggerTypes { get; set; }

        /// <summary>Preferred data column order</summary>
        public List<string>? PreferredColumns { get; set; }

        /// <summary>CSV field delimiter. Default is ';' (Excel-EU locale default); pass ',' for RFC 4180 / pandas compatibility. Cell values are always quoted when they contain the delimiter, the line terminator, or the quote character, with embedded quotes doubled per RFC 4180.</summary>
        public string? Delimiter { get; set; }

        /// <summary>CSV line delimiter</summary>
        public string? LineDelimiter { get; set; }

        /// <summary>CSV quote character</summary>
        public string? Quote { get; set; }

    }

    /// <summary>Request options for <see cref="WorkflowRunsService.CancelAsync"/>: Cancel Workflow Run</summary>
    public class WorkflowRunsCancelOptions : BaseOptions
    {
    }
}
