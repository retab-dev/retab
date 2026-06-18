namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents create file blueprint request mode values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum CreateFileBlueprintRequestMode
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "instant")]
        Instant,
        [EnumMember(Value = "reasoning")]
        Reasoning,
    }
}
