namespace Retab
{
    using System;

    /// <summary>An edit result: form-field values written onto a document or template PDF.</summary>
    public class Edit
    {

        /// <summary>Unique identifier of the edit.</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the source file (input document or template PDF).</summary>
        public BlockExecFileRef File { get; set; } = default!;

        /// <summary>Model used for the edit operation.</summary>
        public string Model { get; set; } = default!;

        /// <summary>Free-form instructions supplied with the edit request.</summary>
        public string? Instructions { get; set; }

        /// <summary>Configuration used for the edit operation.</summary>
        public EditConfig Config { get; set; } = default!;

        /// <summary>Template id used when the edit was created from a template; null for direct-document edits.</summary>
        public string? TemplateId { get; set; }

        /// <summary>The edit result: filled form fields and the rendered PDF. An empty sentinel until status == 'completed'; gate reads on status.</summary>
        public EditResult? Output { get; set; }

        /// <summary>Lifecycle status. The synchronous path returns 'completed'. Background runs progress pending -&gt; queued -&gt; in_progress -&gt; completed | failed | cancelled.</summary>
        public ClassificationStatus? Status { get; set; }

        /// <summary>Error details when a background run fails; null otherwise. Always present so consumers can read it without an existence check.</summary>
        public PrimitiveError? Error { get; set; }

        /// <summary>Durable file reference for the filled document, when materialized.</summary>
        public BlockExecFileRef? FilledDocumentRef { get; set; }

        /// <summary>Usage information for the edit operation.</summary>
        public RetabUsage? Usage { get; set; }
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
