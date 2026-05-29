namespace Retab
{

    /// <summary>Represents a materialized document.</summary>
    public class MaterializedDocument
    {
        public string OriginalId { get; set; } = default!;
        public string Filename { get; set; } = default!;
        public string MimeType { get; set; } = default!;
        public string GcsUri { get; set; } = default!;
        public long? SizeBytes { get; set; }
        public string? ContentFingerprint { get; set; }

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
