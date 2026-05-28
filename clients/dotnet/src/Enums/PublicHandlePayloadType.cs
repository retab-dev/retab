namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents public handle payload type values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum PublicHandlePayloadType
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "file")]
        File,
        [EnumMember(Value = "json")]
        Json,
    }
}
