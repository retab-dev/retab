namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents jobs source values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum JobsSource
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "api")]
        Api,
        [EnumMember(Value = "project")]
        Project,
        [EnumMember(Value = "workflow")]
        Workflow,
    }
}
