namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="SplitsService.ListAsync"/>: List Splits</summary>
    public class SplitsListOptions : ListOptions
    {
        public string? Filename { get; set; }

        public ClassificationStatus? Status { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

    }

    /// <summary>Request options for <see cref="SplitsService.CreateAsync"/>: Create Split</summary>
    public class SplitsCreateOptions : BaseOptions
    {
        /// <summary>The document to split</summary>
        public MimeData Document { get; set; } = default!;

        /// <summary>The subdocuments to split the document into</summary>
        public List<Subdocument> Subdocuments { get; set; } = default!;

        /// <summary>The model to use to split the document</summary>
        public string? Model { get; set; }

        /// <summary>Free-form instructions appended to the system prompt to steer the split.</summary>
        public string? Instructions { get; set; }

        /// <summary>Number of consensus split runs to perform. Uses deterministic single-pass when set to 1.</summary>
        public long? NConsensus { get; set; }

        /// <summary>If true, skip the LLM cache and force a fresh completion</summary>
        public bool? BustCache { get; set; }

        /// <summary>If true, run asynchronously: returns immediately with status 'queued' and an empty output. Poll GET /v1/&lt;primitive&gt;/{id} until status is terminal. Mutually exclusive with stream.</summary>
        public bool? Background { get; set; }

    }

    /// <summary>Request options for <see cref="SplitsService.CreateReconstructAsync"/>: Reconstruct Split</summary>
    public class SplitsCreateReconstructOptions : BaseOptions
    {
        public ReconstructDocumentRef Document { get; set; } = default!;

        public List<ReconstructSubdocument> Subdocuments { get; set; } = default!;

    }

    /// <summary>Request options for <see cref="SplitsService.GetAsync"/>: Get Split</summary>
    public class SplitsGetOptions : BaseOptions
    {
        /// <summary>When false, returns a cheap status-only projection (no output), served from cache for in-flight background runs.</summary>
        public bool? IncludeOutput { get; set; }

    }
}
