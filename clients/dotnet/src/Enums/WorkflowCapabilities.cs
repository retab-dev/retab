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

        [EnumMember(Value = "workflow:workflows:read")]
        WorkflowWorkflowsRead,
        [EnumMember(Value = "workflow:workflows:edit")]
        WorkflowWorkflowsEdit,
        [EnumMember(Value = "workflow:workflows:delete")]
        WorkflowWorkflowsDelete,
        [EnumMember(Value = "workflow:workflows:publish")]
        WorkflowWorkflowsPublish,
        [EnumMember(Value = "workflow_members:read")]
        WorkflowMembersRead,
        [EnumMember(Value = "workflow_members:create")]
        WorkflowMembersCreate,
        [EnumMember(Value = "workflow_members:update")]
        WorkflowMembersUpdate,
        [EnumMember(Value = "workflow_members:delete")]
        WorkflowMembersDelete,
        [EnumMember(Value = "workflow:workflows-runs:create")]
        WorkflowWorkflowsRunsCreate,
        [EnumMember(Value = "workflow:workflows-review:create")]
        WorkflowWorkflowsReviewCreate,
    }
}
