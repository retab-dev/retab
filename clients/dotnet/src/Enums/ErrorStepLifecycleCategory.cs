namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents error step lifecycle category values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ErrorStepLifecycleCategory
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "transient")]
        Transient,
        [EnumMember(Value = "permanent")]
        Permanent,
        [EnumMember(Value = "quota")]
        Quota,
    }
}
