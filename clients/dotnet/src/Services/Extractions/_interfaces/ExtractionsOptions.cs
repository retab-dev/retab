namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ExtractionsService.ListAsync"/>: List Extractions</summary>
    public class ExtractionsListOptions : ListOptions
    {
        public string? Filename { get; set; }

        /// <summary>Deprecated alias for prefix filename filtering. Regex patterns are rejected.</summary>
        public string? FilenameRegex { get; set; }

        /// <summary>Plain-text search over the filename.</summary>
        public string? FilenameContains { get; set; }

        /// <summary>Filter by document type. Can be repeated. Accepted values: bmp, csv, doc, docm, docx, dotm, dotx, eml, gif, heic, heif, htm, html, jpeg, jpg, json, md, mhtml, msg, odp, ods, odt, ots, ott, pdf, png, ppt, pptx, rtf, svg, tif, tiff, tsv, txt, webp, xlam, xls, xlsb, xlsm, xlsx, xltm, xltx, xml, yaml, yml.</summary>
        public List<string>? DocumentType { get; set; }

        public ClassificationStatus? Status { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

        public string? Metadata { get; set; }

    }

    /// <summary>Request options for <see cref="ExtractionsService.CreateAsync"/>: Create Extraction</summary>
    public class ExtractionsCreateOptions : BaseOptions
    {
        public MimeData Document { get; set; } = default!;

        /// <summary>JSON schema describing the structured output</summary>
        public Dictionary<string, object> JsonSchema { get; set; } = default!;

        /// <summary>The model to use for the extraction</summary>
        public string? Model { get; set; }

        /// <summary>Free-form instructions appended to the system prompt to steer the extraction.</summary>
        public string? Instructions { get; set; }

        /// <summary>Number of consensus extraction runs to perform. Uses deterministic single-pass when set to 1.</summary>
        public long? NConsensus { get; set; }

        /// <summary>User-defined metadata to associate with this extraction</summary>
        public Dictionary<string, string>? Metadata { get; set; }

        /// <summary>Additional chat messages forwarded to the extraction model.</summary>
        public List<Dictionary<string, object>>? AdditionalMessages { get; set; }

        /// <summary>If true, skip the LLM cache and force a fresh completion</summary>
        public bool? BustCache { get; set; }

        public bool? Stream { get; set; }

        /// <summary>If true, run asynchronously: returns immediately with status 'queued' and an empty output. Poll GET /v1/&lt;primitive&gt;/{id} until status is terminal. Mutually exclusive with stream.</summary>
        public bool? Background { get; set; }

        public Dictionary<string, string>? ChunkingKeys { get; set; }

        public long? ImageResolutionDpi { get; set; }

    }

    /// <summary>Request options for <see cref="ExtractionsService.CreateStreamAsync"/>: Create Extraction Stream</summary>
    public class ExtractionsCreateStreamOptions : BaseOptions
    {
        public MimeData Document { get; set; } = default!;

        /// <summary>JSON schema describing the structured output</summary>
        public Dictionary<string, object> JsonSchema { get; set; } = default!;

        /// <summary>The model to use for the extraction</summary>
        public string? Model { get; set; }

        /// <summary>Free-form instructions appended to the system prompt to steer the extraction.</summary>
        public string? Instructions { get; set; }

        /// <summary>Number of consensus extraction runs to perform. Uses deterministic single-pass when set to 1.</summary>
        public long? NConsensus { get; set; }

        /// <summary>User-defined metadata to associate with this extraction</summary>
        public Dictionary<string, string>? Metadata { get; set; }

        /// <summary>Additional chat messages forwarded to the extraction model.</summary>
        public List<Dictionary<string, object>>? AdditionalMessages { get; set; }

        /// <summary>If true, skip the LLM cache and force a fresh completion</summary>
        public bool? BustCache { get; set; }

        public bool? Stream { get; set; }

        /// <summary>If true, run asynchronously: returns immediately with status 'queued' and an empty output. Poll GET /v1/&lt;primitive&gt;/{id} until status is terminal. Mutually exclusive with stream.</summary>
        public bool? Background { get; set; }

        public Dictionary<string, string>? ChunkingKeys { get; set; }

        public long? ImageResolutionDpi { get; set; }

    }

    /// <summary>Request options for <see cref="ExtractionsService.GetAsync"/>: Get Extraction</summary>
    public class ExtractionsGetOptions : BaseOptions
    {
        /// <summary>When false, returns a cheap status-only projection (no output), served from cache for in-flight background runs.</summary>
        public bool? IncludeOutput { get; set; }

    }
}
