namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents create experiment request n consensus values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum CreateExperimentRequestNConsensus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "3")]
        Value3,
        [EnumMember(Value = "5")]
        Value5,
        [EnumMember(Value = "7")]
        Value7,
    }
}
