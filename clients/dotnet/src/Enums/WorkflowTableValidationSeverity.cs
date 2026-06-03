namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow table validation severity values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowTableValidationSeverity
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "error")]
        Error,
        [EnumMember(Value = "warning")]
        Warning,
    }
}
