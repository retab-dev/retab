namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A workflow and its current configuration.</summary>
    public class Workflow
    {

        /// <summary>Unique ID for this workflow</summary>
        public string Id { get; set; } = default!;

        /// <summary>The name of the workflow</summary>
        public string? Name { get; set; } = "Untitled Workflow";

        /// <summary>Description of the workflow</summary>
        public string? Description { get; set; } = "";

        /// <summary>Project that owns this workflow. Null only on legacy rows that predate the project backfill.</summary>
        public string? ProjectId { get; set; }

        /// <summary>Published workflow metadata when a published version exists</summary>
        public WorkflowPublished? Published { get; set; }
        public DateTimeOffset CreatedAt { get; set; }
        public DateTimeOffset UpdatedAt { get; set; }

        /// <summary>Server-derived permissions for the current actor.</summary>
        public List<WorkflowCapabilities>? Capabilities { get; set; }

        /// <summary>Provisioning state of this workflow's WorkOS authorization resource.</summary>
        public WorkflowAuthzStatus? AuthzStatus { get; set; }

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
