namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents latest block eval run summary status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum LatestBlockEvalRunSummaryStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "pending")]
        Pending,
        [EnumMember(Value = "queued")]
        Queued,
        [EnumMember(Value = "running")]
        Running,
        [EnumMember(Value = "completed")]
        Completed,
        [EnumMember(Value = "error")]
        Error,
        [EnumMember(Value = "cancelled")]
        Cancelled,
    }
}
