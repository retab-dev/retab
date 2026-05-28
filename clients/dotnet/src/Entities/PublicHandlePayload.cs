namespace Retab
{
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
        public FileRef? Document { get; set; }

        /// <summary>For JSON handles: structured data</summary>
        public object? Data { get; set; }
    }
}
