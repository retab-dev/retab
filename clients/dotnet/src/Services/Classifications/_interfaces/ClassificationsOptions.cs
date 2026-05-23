namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="ClassificationsService.ListAsync"/>: List Classifications</summary>
    public class ClassificationsListOptions : ListOptions
    {
        public string? Filename { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

    }

    /// <summary>Request options for <see cref="ClassificationsService.CreateAsync"/>: Create Classification</summary>
    public class ClassificationsCreateOptions : BaseOptions
    {
        /// <summary>The document to classify</summary>
        public MimeData Document { get; set; } = default!;

        /// <summary>The categories to classify the document into</summary>
        public List<Category> Categories { get; set; } = default!;

        /// <summary>The model to use for classification</summary>
        public string? Model { get; set; }

        /// <summary>Only use the first N pages of the document for classification. Useful for large documents where classification can be determined from early pages.</summary>
        public long? FirstNPages { get; set; }

        /// <summary>Free-form instructions appended to the system prompt to steer the classification.</summary>
        public string? Instructions { get; set; }

        /// <summary>Number of classification runs to use for consensus voting. Uses deterministic single-pass when set to 1.</summary>
        public long? NConsensus { get; set; }

        /// <summary>If true, skip the LLM cache and force a fresh completion</summary>
        public bool? BustCache { get; set; }

    }
}
