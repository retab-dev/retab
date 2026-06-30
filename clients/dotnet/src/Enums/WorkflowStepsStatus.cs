namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow steps status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowStepsStatus
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
        [EnumMember(Value = "awaiting_review")]
        AwaitingReview,
        [EnumMember(Value = "error")]
        Error,
        [EnumMember(Value = "skipped")]
        Skipped,
        [EnumMember(Value = "cancelled")]
        Cancelled,
    }
}
