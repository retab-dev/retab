namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents declarative plan field change action values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum DeclarativePlanFieldChangeAction
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "create")]
        Create,
        [EnumMember(Value = "update")]
        Update,
        [EnumMember(Value = "delete")]
        Delete,
    }
}
