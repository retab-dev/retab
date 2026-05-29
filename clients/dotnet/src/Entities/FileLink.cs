namespace Retab
{

    /// <summary>A short-lived signed link to download a file, with its `filename` and expiry.</summary>
    public class FileLink
    {

        /// <summary>The signed URL to download the file</summary>
        public string DownloadUrl { get; set; } = default!;

        /// <summary>The expiration time of the signed URL</summary>
        public string ExpiresIn { get; set; } = default!;

        /// <summary>The name of the file</summary>
        public string Filename { get; set; } = default!;

        /// <summary>Durable Retab MIMEData reference for API reuse</summary>
        public MimeData? MimeData { get; set; }

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
