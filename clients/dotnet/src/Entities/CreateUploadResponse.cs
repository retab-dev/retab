namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Instructions for uploading file content to a reserved file record.</summary>
    /// <remarks>
    /// Returned when starting a file upload. Carries the new `file_id`, a
    /// short-lived signed `upload_url` with the HTTP method and headers to use,
    /// a durable reference to the file, and the URL's `expires_at` time.
    /// </remarks>
    public class CreateUploadResponse
    {

        /// <summary>Underlying file ID</summary>
        public string FileId { get; set; } = default!;

        /// <summary>Short-lived signed upload URL</summary>
        public string UploadUrl { get; set; } = default!;

        /// <summary>HTTP method for upload</summary>
        public string? UploadMethod { get; set; } = "PUT";

        /// <summary>Headers required by the signed upload URL</summary>
        public Dictionary<string, string>? UploadHeaders { get; set; }

        /// <summary>Durable Retab MIMEData reference</summary>
        public MimeData? MimeData { get; set; }

        /// <summary>Upload URL expiration</summary>
        public DateTimeOffset ExpiresAt { get; set; }

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
