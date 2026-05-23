namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents while loop termination termination reason values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum WhileLoopTerminationTerminationReason
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "max_iterations_reached")]
        MaxIterationsReached,
        [EnumMember(Value = "condition_matched")]
        ConditionMatched,
        [EnumMember(Value = "error")]
        Error,
    }
}
