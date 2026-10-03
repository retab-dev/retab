namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents source field evidence kind values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum SourceFieldEvidenceKind
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "direct")]
        Direct,
        [EnumMember(Value = "supporting")]
        Supporting,
        [EnumMember(Value = "unavailable")]
        Unavailable,
    }
}
