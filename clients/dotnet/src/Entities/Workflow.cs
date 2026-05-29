namespace Retab
{
    using System;

    /// <summary>Public workflow resource returned by workflow metadata endpoints.</summary>
    /// <remarks>
    /// This is the API response shape — distinct from the internal storage
    /// model (:class:`StoredWorkflow`), which carries persistence-only fields
    /// that never appear in API responses. Routes call
    /// :func:`serialize_workflow_response` to convert the storage shape into
    /// this response shape; constructing this class directly for persistence
    /// would drop those storage-only fields silently. The two classes were
    /// deliberately renamed to avoid the prior name collision.
    /// </remarks>
    public class Workflow
    {

        /// <summary>Unique ID for this workflow</summary>
        public string Id { get; set; } = default!;

        /// <summary>The name of the workflow</summary>
        public string? Name { get; set; }

        /// <summary>Description of the workflow</summary>
        public string? Description { get; set; }

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
