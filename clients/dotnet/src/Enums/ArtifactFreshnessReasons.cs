namespace Retab
{
    using System.Runtime.Serialization;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents artifact freshness reasons values.</summary>
    [JsonConverter(typeof(RetabNewtonsoftStringEnumConverter))]
    [STJS.JsonConverter(typeof(RetabStringEnumConverterFactory))]
    public enum ArtifactFreshnessReasons
    {
        [EnumMember(Value = "unknown")]
        Unknown,

        [EnumMember(Value = "validity_changed")]
        ValidityChanged,
        [EnumMember(Value = "inputs_changed")]
        InputsChanged,
        [EnumMember(Value = "engine_changed")]
        EngineChanged,
        [EnumMember(Value = "metrics_engine_changed")]
        MetricsEngineChanged,
        [EnumMember(Value = "no_baseline")]
        NoBaseline,
    }
}
