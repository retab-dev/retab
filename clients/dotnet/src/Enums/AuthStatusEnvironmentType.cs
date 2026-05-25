namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents auth status environment type values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum AuthStatusEnvironmentType
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "production")]
        Production,
        [EnumMember(Value = "non_production")]
        NonProduction,
    }
}
