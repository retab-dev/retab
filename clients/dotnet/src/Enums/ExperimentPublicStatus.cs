namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents experiment public status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ExperimentPublicStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "draft")]
        Draft,
        [EnumMember(Value = "processing")]
        Processing,
        [EnumMember(Value = "completed")]
        Completed,
        [EnumMember(Value = "failed")]
        Failed,
        [EnumMember(Value = "cancelled")]
        Cancelled,
    }
}
