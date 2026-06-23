namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow experiment run plan mode values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowExperimentRunPlanMode
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "run")]
        Run,
        [EnumMember(Value = "noop")]
        Noop,
        [EnumMember(Value = "conflict")]
        Conflict,
    }
}
