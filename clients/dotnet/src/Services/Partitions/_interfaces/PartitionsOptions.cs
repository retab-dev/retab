namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="PartitionsService.ListAsync"/>: List Partitions</summary>
    public class PartitionsListOptions : ListOptions
    {
        public string? Filename { get; set; }

        public ClassificationStatus? Status { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

    }

    /// <summary>Request options for <see cref="PartitionsService.CreateAsync"/>: Create Partitions</summary>
    public class PartitionsCreateOptions : BaseOptions
    {
        /// <summary>The document to partition</summary>
        public MimeData Document { get; set; } = default!;

        /// <summary>The key to partition the document by</summary>
        public string Key { get; set; } = default!;

        /// <summary>Instructions describing how the document should be partitioned</summary>
        public string Instructions { get; set; } = default!;

        /// <summary>The model to use for partitioning</summary>
        public string? Model { get; set; }

        /// <summary>Number of partitioning runs to use for consensus voting. Uses deterministic single-pass when set to 1.</summary>
        public long? NConsensus { get; set; }

        /// <summary>If true, allow a page to appear in more than one partition chunk</summary>
        public bool? AllowOverlap { get; set; }

        /// <summary>If true, skip the LLM cache and force a fresh completion</summary>
        public bool? BustCache { get; set; }

        /// <summary>If true, run asynchronously: returns immediately with status 'queued' and an empty output. Poll GET /v1/&lt;primitive&gt;/{id} until status is terminal. Mutually exclusive with stream.</summary>
        public bool? Background { get; set; }

    }

    /// <summary>Request options for <see cref="PartitionsService.GetAsync"/>: Get Partition</summary>
    public class PartitionsGetOptions : BaseOptions
    {
        /// <summary>When false, returns a cheap status-only projection (no output), served from cache for in-flight background runs.</summary>
        public bool? IncludeOutput { get; set; }

    }
}
