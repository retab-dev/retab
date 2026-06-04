namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Public workflow version resource without tenant fields.</summary>
    public class WorkflowGraphVersion
    {

        /// <summary>Public content-addressed workflow version ID</summary>
        public string Id { get; set; } = default!;
        public string WorkflowId { get; set; } = default!;
        public List<WorkflowConfigBlock>? Blocks { get; set; }
        public List<WorkflowConfigEdge>? Edges { get; set; }
        public Dictionary<string, string>? BlockVersionIds { get; set; }
        public Dictionary<string, string>? EdgeVersionIds { get; set; }
        public DateTimeOffset CreatedAt { get; set; }

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
