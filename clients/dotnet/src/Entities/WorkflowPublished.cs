namespace Retab
{
    using System;

    /// <summary>Represents a workflow published.</summary>
    public class WorkflowPublished
    {

        /// <summary>Published content-addressed workflow version ID</summary>
        public string? VersionId { get; set; }

        /// <summary>When the workflow was last published</summary>
        public DateTimeOffset? PublishedAt { get; set; }

        /// <summary>Release note attached to the currently published version. Echoes the ``description`` body passed to ``POST /v1/workflows/{id}/publish`` so the caller can confirm it was stored without a separate fetch.</summary>
        public string? Description { get; set; }

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
