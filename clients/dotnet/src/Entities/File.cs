namespace Retab
{
    using System;

    /// <summary>Represents a file.</summary>
    public class File
    {
        public string? Object { get; set; }

        /// <summary>The unique identifier of the file</summary>
        public string Id { get; set; } = default!;

        /// <summary>The name of the file</summary>
        public string Filename { get; set; } = default!;

        /// <summary>The MIME type of the file</summary>
        public string? MimeType { get; set; }

        /// <summary>When the file was created</summary>
        public DateTimeOffset? CreatedAt { get; set; }

        /// <summary>When the file was last updated</summary>
        public DateTimeOffset? UpdatedAt { get; set; }

        /// <summary>Number of pages in the file</summary>
        public long? PageCount { get; set; }

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
