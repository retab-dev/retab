namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="EditsService.ListAsync"/>: List Edits</summary>
    public class EditsListOptions : ListOptions
    {
        public string? Filename { get; set; }

        public string? TemplateId { get; set; }

        public ClassificationStatus? Status { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

    }

    /// <summary>Request options for <see cref="EditsService.CreateAsync"/>: Create Edit</summary>
    public class EditsCreateOptions : BaseOptions
    {
        /// <summary>Instructions describing how to fill the form fields.</summary>
        public string Instructions { get; set; } = default!;

        /// <summary>Input document (PDF, DOCX, XLSX, or PPTX). Mutually exclusive with template_id.</summary>
        public MimeData? Document { get; set; }

        /// <summary>EditTemplate id to fill. When provided, uses the template's pre-defined form fields and empty PDF. Mutually exclusive with document.</summary>
        public string? TemplateId { get; set; }

        /// <summary>The model to use for edit inference.</summary>
        public string? Model { get; set; }

        /// <summary>Edit configuration (rendering options).</summary>
        public EditConfig? Config { get; set; }

        /// <summary>If true, skip the LLM cache and force a fresh completion.</summary>
        public bool? BustCache { get; set; }

        /// <summary>If true, run asynchronously: returns immediately with status 'queued' and an empty output. Poll GET /v1/&lt;primitive&gt;/{id} until status is terminal. Mutually exclusive with stream.</summary>
        public bool? Background { get; set; }

    }

    /// <summary>Request options for <see cref="EditsService.GetAsync"/>: Get Edit</summary>
    public class EditsGetOptions : BaseOptions
    {
        /// <summary>When false, returns a cheap status-only projection (no output), served from cache for in-flight background runs.</summary>
        public bool? IncludeOutput { get; set; }

    }
}
