namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents field type values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum FieldType
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "text")]
        Text,
        [EnumMember(Value = "checkbox")]
        Checkbox,
    }
}
