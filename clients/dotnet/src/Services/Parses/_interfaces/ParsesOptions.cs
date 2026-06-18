namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ParsesService.ListAsync"/>: List Parses</summary>
    public class ParsesListOptions : ListOptions
    {
        public string? Filename { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

    }

    /// <summary>Request options for <see cref="ParsesService.CreateAsync"/>: Create Parse</summary>
    public class ParsesCreateOptions : BaseOptions
    {
        /// <summary>The document to parse</summary>
        public MimeData Document { get; set; } = default!;

        /// <summary>The model to use for parsing</summary>
        public string? Model { get; set; }

        /// <summary>Format used to render tables extracted from the document</summary>
        public ParseRequestTableParsingFormat? TableParsingFormat { get; set; }

        /// <summary>Free-form instructions appended to the system prompt to steer the parse.</summary>
        public string? Instructions { get; set; }

        /// <summary>If true, skip the LLM cache and force a fresh completion</summary>
        public bool? BustCache { get; set; }

        /// <summary>If true, run asynchronously: returns immediately with status 'queued' and an empty output. Poll GET /v1/&lt;primitive&gt;/{id} until status is terminal. Mutually exclusive with stream.</summary>
        public bool? Background { get; set; }

        public long? ImageResolutionDpi { get; set; }

    }

    /// <summary>Request options for <see cref="ParsesService.GetAsync"/>: Get Parse</summary>
    public class ParsesGetOptions : BaseOptions
    {
        /// <summary>When false, returns a cheap status-only projection (no output), served from cache for in-flight background runs.</summary>
        public bool? IncludeOutput { get; set; }

    }
}
