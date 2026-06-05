namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow authz status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowAuthzStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "provisioning")]
        Provisioning,
        [EnumMember(Value = "ready")]
        Ready,
        [EnumMember(Value = "failed")]
        Failed,
        [EnumMember(Value = "deleting")]
        Deleting,
        [EnumMember(Value = "deleted")]
        Deleted,
    }
}
