namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents declarative apply response action values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum DeclarativeApplyResponseAction
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "create")]
        Create,
        [EnumMember(Value = "update")]
        Update,
        [EnumMember(Value = "noop")]
        Noop,
    }
}
