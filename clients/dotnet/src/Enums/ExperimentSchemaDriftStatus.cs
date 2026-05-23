namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents experiment schema drift status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ExperimentSchemaDriftStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "none")]
        None,
        [EnumMember(Value = "partial")]
        Partial,
        [EnumMember(Value = "drifted")]
        Drifted,
    }
}
