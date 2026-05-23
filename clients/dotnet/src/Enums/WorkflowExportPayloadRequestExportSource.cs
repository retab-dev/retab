namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents workflow export payload request export source values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WorkflowExportPayloadRequestExportSource
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "outputs")]
        Outputs,
        [EnumMember(Value = "inputs")]
        Inputs,
    }
}
