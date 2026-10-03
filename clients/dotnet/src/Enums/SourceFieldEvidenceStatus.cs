namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents source field evidence status values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum SourceFieldEvidenceStatus
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "pending")]
        Pending,
        [EnumMember(Value = "matched")]
        Matched,
        [EnumMember(Value = "partial")]
        Partial,
        [EnumMember(Value = "missing")]
        Missing,
        [EnumMember(Value = "ambiguous")]
        Ambiguous,
        [EnumMember(Value = "unsupported")]
        Unsupported,
        [EnumMember(Value = "error")]
        Error,
    }
}
