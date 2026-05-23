namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents a create upload response.</summary>
    public class CreateUploadResponse
    {

        /// <summary>Underlying file ID</summary>
        public string FileId { get; set; } = default!;

        /// <summary>Short-lived signed upload URL</summary>
        public string UploadUrl { get; set; } = default!;

        /// <summary>HTTP method for upload</summary>
        public string? UploadMethod { get; set; }

        /// <summary>Headers required by the signed upload URL</summary>
        public Dictionary<string, string>? UploadHeaders { get; set; }

        /// <summary>Durable Retab MIMEData reference</summary>
        public MimeData MimeData { get; set; } = default!;

        /// <summary>Upload URL expiration</summary>
        public DateTimeOffset ExpiresAt { get; set; }
    }
}
