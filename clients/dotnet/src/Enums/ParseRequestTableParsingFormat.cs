namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents parse request table parsing format values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ParseRequestTableParsingFormat
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "markdown")]
        Markdown,
        [EnumMember(Value = "yaml")]
        Yaml,
        [EnumMember(Value = "html")]
        Html,
        [EnumMember(Value = "json")]
        Json,
    }
}
