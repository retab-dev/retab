namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow capabilities values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowCapabilities
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "workflow:view")]
        WorkflowView,
        [EnumMember(Value = "workflow:edit")]
        WorkflowEdit,
        [EnumMember(Value = "workflow:run")]
        WorkflowRun,
        [EnumMember(Value = "workflow:delete")]
        WorkflowDelete,
        [EnumMember(Value = "workflow:publish")]
        WorkflowPublish,
        [EnumMember(Value = "workflow:review")]
        WorkflowReview,
        [EnumMember(Value = "workflow:manage_members")]
        WorkflowManageMembers,
    }
}
