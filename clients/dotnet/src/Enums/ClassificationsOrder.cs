namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents classifications order values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ClassificationsOrder
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "asc")]
        Asc,
        [EnumMember(Value = "desc")]
        Desc,
    }
}
