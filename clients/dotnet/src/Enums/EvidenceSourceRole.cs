namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents evidence source role values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum EvidenceSourceRole
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "answer")]
        Answer,
        [EnumMember(Value = "input")]
        Input,
        [EnumMember(Value = "comparison")]
        Comparison,
    }
}
