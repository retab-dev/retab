namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents experiment run metrics view values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ExperimentRunMetricsView
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "summary")]
        Summary,
        [EnumMember(Value = "by_document")]
        ByDocument,
        [EnumMember(Value = "by_target")]
        ByTarget,
        [EnumMember(Value = "votes")]
        Votes,
    }
}
