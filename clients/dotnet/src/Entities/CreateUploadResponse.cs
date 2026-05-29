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
        public string? UploadMethod { get; set; } = "PUT";

        /// <summary>Headers required by the signed upload URL</summary>
        public Dictionary<string, string>? UploadHeaders { get; set; }

        /// <summary>Durable Retab MIMEData reference</summary>
        public MimeData MimeData { get; set; } = default!;

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
