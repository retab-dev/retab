namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>A parse result: the per-page and full-document text extracted from a document.</summary>
    public class Parse
    {

        /// <summary>Unique identifier of the parse</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the parsed file</summary>
        public FileRef File { get; set; } = default!;

        /// <summary>Model used for parsing</summary>
        public string Model { get; set; } = default!;

        /// <summary>Format used to render tables extracted from the document</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public ParseRequestTableParsingFormat TableParsingFormat { get; set; }

        /// <summary>DPI used when rasterizing pages for the parser</summary>
        public long ImageResolutionDpi { get; set; }

        /// <summary>Free-form instructions supplied with the parse request.</summary>
        public string? Instructions { get; set; }

        /// <summary>The parsed document content</summary>
        public ParseOutput Output { get; set; } = default!;

        /// <summary>Lifecycle status. The synchronous path returns 'completed'. Background runs progress pending -&gt; queued -&gt; in_progress -&gt; completed | failed | cancelled.</summary>
        public ClassificationStatus? Status { get; set; }

        /// <summary>Error details when a background run fails; null otherwise. Always present so consumers can read it without an existence check.</summary>
        public PrimitiveError? Error { get; set; }

        /// <summary>Usage information for the parse operation</summary>
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
