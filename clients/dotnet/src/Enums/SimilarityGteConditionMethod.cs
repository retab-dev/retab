namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents similarity gte condition method values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum SimilarityGteConditionMethod
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "levenshtein")]
        Levenshtein,
        [EnumMember(Value = "embeddings")]
        Embeddings,
    }
}
