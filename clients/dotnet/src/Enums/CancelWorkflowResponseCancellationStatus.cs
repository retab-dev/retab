namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents cancel workflow response cancellation status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum CancelWorkflowResponseCancellationStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "cancelled")]
        Cancelled,
        [EnumMember(Value = "cancellation_requested")]
        CancellationRequested,
        [EnumMember(Value = "cancellation_failed")]
        CancellationFailed,
    }
}
