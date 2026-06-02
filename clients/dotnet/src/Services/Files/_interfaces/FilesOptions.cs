namespace Retab
{
    using System;
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Request options for <see cref="FilesService.ListAsync"/>: List Files</summary>
    public class FilesListOptions : ListOptions
    {
        public string? Filename { get; set; }

        public string? MimeType { get; set; }

        public string? FromDate { get; set; }

        public string? ToDate { get; set; }

        /// <summary>Include embeddings in the response</summary>
        public bool? IncludeEmbeddings { get; set; }

        public string? SortBy { get; set; }

    }

    /// <summary>Request options for <see cref="FilesService.CreateBlueprintAsync"/>: Create File Blueprint</summary>
    public class FilesCreateBlueprintOptions : BaseOptions
    {
        /// <summary>File id to analyze.</summary>
        public string FileId { get; set; } = default!;

        /// <summary>Optional analysis depth override. Omit to let Retab choose.</summary>
        public CreateFileBlueprintRequestMode? Mode { get; set; }

        /// <summary>Optional user intent used to guide the blueprint analysis.</summary>
        public string? Intent { get; set; }

        /// <summary>If true, run asynchronously: returns immediately with status 'queued' and an empty output. Poll GET /v1/&lt;primitive&gt;/{id} until status is terminal. Mutually exclusive with stream.</summary>
        public bool? Background { get; set; }

    }

    /// <summary>Request options for <see cref="FilesService.GetBlueprintAsync"/>: Get File Blueprint</summary>
    public class FilesGetBlueprintOptions : BaseOptions
    {
        /// <summary>When false, returns a cheap status-only projection (no output), served from cache for in-flight background runs.</summary>
        public bool? IncludeOutput { get; set; }

    }

    /// <summary>Request options for <see cref="FilesService.CreateUploadAsync"/>: Upload File</summary>
    public class FilesCreateUploadOptions : BaseOptions
    {
        /// <summary>Filename to store</summary>
        public string Filename { get; set; } = default!;

        /// <summary>MIME type the client will upload</summary>
        public string? ContentType { get; set; }

        /// <summary>Expected upload size in bytes</summary>
        public long SizeBytes { get; set; }

        /// <summary>Optional SHA-256 checksum</summary>
        public string? Sha256 { get; set; }

    }

    /// <summary>Request options for <see cref="FilesService.CompleteUploadAsync"/>: Complete Upload File</summary>
    public class FilesCompleteUploadOptions : BaseOptions
    {
        /// <summary>Optional SHA-256 checksum</summary>
        public string? Sha256 { get; set; }

    }
}
