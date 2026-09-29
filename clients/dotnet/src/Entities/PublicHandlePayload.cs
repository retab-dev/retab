namespace Retab
{
    using System;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Public handle payload exposed by workflow step APIs.</summary>
    public class PublicHandlePayload
    {

        /// <summary>Type of payload</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public PublicHandlePayloadType Type { get; set; }

        /// <summary>For file handles: document reference</summary>
        public BlockExecFileRef? Document { get; set; }

        /// <summary>For JSON handles: structured data</summary>
        public object? Data { get; set; }

        /// <summary>For file handles in `handle_outputs`, when the step was read with `include_download_urls=true`: a short-lived signed URL for the document. Use `document.id` with `GET /v1/files/{file_id}/download-link` for a fresh link after it expires.</summary>
        public string? DownloadUrl { get; set; }

        /// <summary>When `download_url` stops working.</summary>
        public DateTimeOffset? ExpiresAt { get; set; }

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
