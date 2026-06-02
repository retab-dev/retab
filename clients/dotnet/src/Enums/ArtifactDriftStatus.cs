namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents artifact drift status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ArtifactDriftStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "none")]
        None,
        [EnumMember(Value = "drifted")]
        Drifted,
        [EnumMember(Value = "broken")]
        Broken,
    }
}
