namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents create sources request mode values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum CreateSourcesRequestMode
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "located")]
        Located,
        [EnumMember(Value = "cited")]
        Cited,
    }
}
