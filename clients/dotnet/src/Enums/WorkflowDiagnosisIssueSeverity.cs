namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow diagnosis issue severity values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowDiagnosisIssueSeverity
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "error")]
        Error,
        [EnumMember(Value = "warning")]
        Warning,
        [EnumMember(Value = "info")]
        Info,
    }
}
