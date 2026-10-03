namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents evidence source qualification values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum EvidenceSourceQualification
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "direct")]
        Direct,
        [EnumMember(Value = "partial")]
        Partial,
        [EnumMember(Value = "legacy")]
        Legacy,
    }
}
