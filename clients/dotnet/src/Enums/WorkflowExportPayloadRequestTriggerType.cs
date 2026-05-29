namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow export payload request trigger type values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowExportPayloadRequestTriggerType
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "manual")]
        Manual,
        [EnumMember(Value = "api")]
        Api,
        [EnumMember(Value = "schedule")]
        Schedule,
        [EnumMember(Value = "webhook")]
        Webhook,
        [EnumMember(Value = "email")]
        Email,
        [EnumMember(Value = "restart")]
        Restart,
    }
}
