namespace Retab
{
    using System;

    /// <summary>Immutable edge snapshot derived from a workflow version.</summary>
    public class WorkflowEdgeVersion
    {

        /// <summary>Public content-addressed edge version ID</summary>
        public string Id { get; set; } = default!;

        /// <summary>Stable logical edge ID</summary>
        public string EdgeId { get; set; } = default!;

        /// <summary>Source workflow ID</summary>
        public string WorkflowId { get; set; } = default!;

        /// <summary>Organization ID for data isolation</summary>
        public string OrganizationId { get; set; } = default!;

        /// <summary>Customer environment ID for data isolation</summary>
        public string EnvironmentId { get; set; } = default!;

        /// <summary>Workflow version this edge version belongs to</summary>
        public string WorkflowVersionId { get; set; } = default!;

        /// <summary>ID of the source block</summary>
        public string Source { get; set; } = default!;
        public string? SourceHandle { get; set; }

        /// <summary>ID of the target block</summary>
        public string Target { get; set; } = default!;
        public string? TargetHandle { get; set; }
        public bool? Animated { get; set; } = true;
        public DateTimeOffset? CreatedAt { get; set; }

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
