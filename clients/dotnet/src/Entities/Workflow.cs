namespace Retab
{
    using System;

    /// <summary>A workflow and its current configuration.</summary>
    public class Workflow
    {

        /// <summary>Unique ID for this workflow</summary>
        public string Id { get; set; } = default!;

        /// <summary>The name of the workflow</summary>
        public string? Name { get; set; } = "Untitled Workflow";

        /// <summary>Description of the workflow</summary>
        public string? Description { get; set; } = "";

        /// <summary>Published workflow metadata when a published version exists</summary>
        public WorkflowPublished? Published { get; set; }
        public DateTimeOffset CreatedAt { get; set; }
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
