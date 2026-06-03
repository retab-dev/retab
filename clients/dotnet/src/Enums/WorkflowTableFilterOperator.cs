namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow table filter operator values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowTableFilterOperator
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "eq")]
        Eq,
        [EnumMember(Value = "ne")]
        Ne,
        [EnumMember(Value = "gt")]
        Gt,
        [EnumMember(Value = "gte")]
        Gte,
        [EnumMember(Value = "lt")]
        Lt,
        [EnumMember(Value = "lte")]
        Lte,
        [EnumMember(Value = "contains")]
        Contains,
        [EnumMember(Value = "not_contains")]
        NotContains,
        [EnumMember(Value = "starts_with")]
        StartsWith,
        [EnumMember(Value = "ends_with")]
        EndsWith,
        [EnumMember(Value = "in")]
        In,
        [EnumMember(Value = "not_in")]
        NotIn,
        [EnumMember(Value = "between")]
        Between,
        [EnumMember(Value = "is_empty")]
        IsEmpty,
        [EnumMember(Value = "is_not_empty")]
        IsNotEmpty,
        [EnumMember(Value = "is_null")]
        IsNull,
        [EnumMember(Value = "is_not_null")]
        IsNotNull,
    }
}
