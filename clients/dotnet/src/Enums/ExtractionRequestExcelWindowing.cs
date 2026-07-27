namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents extraction request excel windowing values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ExtractionRequestExcelWindowing
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "manual")]
        Manual,
        [EnumMember(Value = "auto")]
        Auto,
    }
}
