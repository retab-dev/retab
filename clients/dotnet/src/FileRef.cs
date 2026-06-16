// @oagen-ignore-file

using System.Collections.Generic;
using System.Text.Json.Serialization;
using Newtonsoft.Json;

namespace Retab
{
    /// <summary>Public/shared file reference used across SDK and customer-facing APIs.</summary>
    public sealed class FileRef
    {
        /// <summary>ID of the file.</summary>
        [JsonPropertyName("id")]
        [JsonProperty("id")]
        public string Id { get; set; } = string.Empty;

        /// <summary>Filename of the file.</summary>
        [JsonPropertyName("filename")]
        [JsonProperty("filename")]
        public string Filename { get; set; } = string.Empty;

        /// <summary>MIME type of the file.</summary>
        [JsonPropertyName("mime_type")]
        [JsonProperty("mime_type")]
        public string MimeType { get; set; } = string.Empty;

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data.
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public IDictionary<string, object> AdditionalData { get; set; } = new Dictionary<string, object>();
    }
}
