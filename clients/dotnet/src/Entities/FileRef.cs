namespace Retab
{

    /// <summary>Public/shared file reference used across SDK and customer-facing APIs.</summary>
    public class FileRef
    {

        /// <summary>ID of the file</summary>
        public string Id { get; set; } = default!;

        /// <summary>Filename of the file</summary>
        public string Filename { get; set; } = default!;

        /// <summary>MIME type of the file</summary>
        public string MimeType { get; set; } = default!;

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
