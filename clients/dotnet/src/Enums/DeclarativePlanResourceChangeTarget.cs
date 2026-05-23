namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents declarative plan resource change target values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum DeclarativePlanResourceChangeTarget
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "workflow")]
        Workflow,
        [EnumMember(Value = "block")]
        Block,
        [EnumMember(Value = "edge")]
        Edge,
    }
}
