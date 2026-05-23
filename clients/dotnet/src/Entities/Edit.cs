namespace Retab
{
    using System;

    /// <summary>Represents an edit.</summary>
    public class Edit
    {

        /// <summary>Unique identifier of the edit.</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the source file (input document or template PDF).</summary>
        public FileRef File { get; set; } = default!;

        /// <summary>Model used for the edit operation.</summary>
        public string Model { get; set; } = default!;

        /// <summary>Free-form instructions supplied with the edit request.</summary>
        public string? Instructions { get; set; }

        /// <summary>Configuration used for the edit operation.</summary>
        public EditConfig Config { get; set; } = default!;

        /// <summary>Template id used when the edit was created from a template; null for direct-document edits.</summary>
        public string? TemplateId { get; set; }

        /// <summary>The edit result: filled form fields and the rendered PDF.</summary>
        public EditResult Output { get; set; } = default!;

        /// <summary>Usage information for the edit operation.</summary>
        public RetabUsage? Usage { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }
    }
}
