namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow table aggregation function values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowTableAggregationFunction
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "count")]
        Count,
        [EnumMember(Value = "count_distinct")]
        CountDistinct,
        [EnumMember(Value = "min")]
        Min,
        [EnumMember(Value = "max")]
        Max,
        [EnumMember(Value = "sum")]
        Sum,
        [EnumMember(Value = "avg")]
        Avg,
    }
}
