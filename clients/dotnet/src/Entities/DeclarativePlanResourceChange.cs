namespace Retab
{
    using System.Collections.Generic;
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a declarative plan resource change.</summary>
    public class DeclarativePlanResourceChange
    {
        public string Address { get; set; } = default!;
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public DeclarativePlanResourceChangeTarget Target { get; set; }
        public string TargetId { get; set; } = default!;
        public string Name { get; set; } = default!;

        /// <summary>Resource kind for this plan entry. ``workflow`` and ``edge`` are flat singletons; for ``target='block'`` this carries the block's concrete type (e.g. ``extract``, ``api_call``) so the plan summary can render type-specific labels.</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public DeclarativePlanResourceChangeType Type { get; set; }
        public List<DeclarativePlanFieldChangeAction> Actions { get; set; } = default!;
        public string Summary { get; set; } = default!;
        public DeclarativePlanChange Change { get; set; } = default!;
        public string? Path { get; set; }

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
