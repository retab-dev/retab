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
        public bool? BeforeSensitive { get; set; }
        public bool? AfterSensitive { get; set; }
        public string? UnifiedDiff { get; set; }
    }
}
