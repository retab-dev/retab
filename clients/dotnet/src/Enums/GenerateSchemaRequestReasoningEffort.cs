namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents generate schema request reasoning effort values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum GenerateSchemaRequestReasoningEffort
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "none")]
        None,
        [EnumMember(Value = "minimal")]
        Minimal,
        [EnumMember(Value = "low")]
        Low,
        [EnumMember(Value = "medium")]
        Medium,
        [EnumMember(Value = "high")]
        High,
        [EnumMember(Value = "xhigh")]
        Xhigh,
    }
}
