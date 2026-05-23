namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents actor kind values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ActorKind
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "model")]
        Model,
        [EnumMember(Value = "agent")]
        Agent,
        [EnumMember(Value = "human")]
        Human,
    }
}
