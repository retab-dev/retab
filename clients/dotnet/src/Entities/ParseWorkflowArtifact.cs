namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a parse workflow artifact.</summary>
    public class ParseWorkflowArtifact
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

        /// <summary>Usage information for the parse operation</summary>
        public RetabUsage? Usage { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
        public DateTimeOffset CreatedAt { get; set; }

        /// <summary>Artifact operation that determines the backing record type</summary>
        public string? Operation { get; set; } = "parse";

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
