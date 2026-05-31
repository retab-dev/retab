namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents classification status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ClassificationStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "pending")]
        Pending,
        [EnumMember(Value = "queued")]
        Queued,
        [EnumMember(Value = "in_progress")]
        InProgress,
        [EnumMember(Value = "completed")]
        Completed,
        [EnumMember(Value = "failed")]
        Failed,
        [EnumMember(Value = "cancelled")]
        Cancelled,
    }
}
