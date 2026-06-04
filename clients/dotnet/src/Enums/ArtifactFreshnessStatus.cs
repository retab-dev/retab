namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents artifact freshness status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ArtifactFreshnessStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "fresh")]
        Fresh,
        [EnumMember(Value = "stale")]
        Stale,
    }
}
