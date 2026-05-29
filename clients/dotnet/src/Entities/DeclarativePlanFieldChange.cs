namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a declarative plan field change.</summary>
    public class DeclarativePlanFieldChange
    {
        public List<OneOf.OneOf<string, long>> Path { get; set; } = default!;
        public string PathDisplay { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public DeclarativePlanFieldChangeAction Action { get; set; }
        public object? Before { get; set; }
        public object? After { get; set; }
        public bool? BeforeSensitive { get; set; } = false;
        public bool? AfterSensitive { get; set; } = false;
        public string? UnifiedDiff { get; set; }

        /// <summary>
        /// Wire fields not modeled by this SDK version, preserved verbatim so a
        /// deserialize → serialize round-trip never drops data (e.g. variant-
        /// specific fields on a discriminated-union response).
        /// </summary>
        [Newtonsoft.Json.JsonExtensionData]
        [System.Text.Json.Serialization.JsonExtensionData]
        public System.Collections.Generic.IDictionary<string, object> AdditionalData { get; set; } = new System.Collections.Generic.Dictionary<string, object>();
    }
}
