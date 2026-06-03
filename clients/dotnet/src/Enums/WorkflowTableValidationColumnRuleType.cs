namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow table validation column rule type values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowTableValidationColumnRuleType
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "array")]
        Array,
        [EnumMember(Value = "boolean")]
        Boolean,
        [EnumMember(Value = "integer")]
        Integer,
        [EnumMember(Value = "null")]
        Null,
        [EnumMember(Value = "number")]
        Number,
        [EnumMember(Value = "object")]
        Object,
        [EnumMember(Value = "string")]
        String,
    }
}
