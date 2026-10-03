namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents an evidence anchor.</summary>
    public class EvidenceAnchor
    {
        public long? CharEnd { get; set; }
        public long? CharStart { get; set; }
        public OneOf.OneOf<string, long>? Column { get; set; }
        public string? Coordinate { get; set; }
        public double? Height { get; set; }
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public EvidenceAnchorKind Kind { get; set; }
        public double? Left { get; set; }
        public long? LineEnd { get; set; }
        public long? LineStart { get; set; }
        public long? Page { get; set; }
        public long? Paragraph { get; set; }
        public long? Row { get; set; }
        public long? SheetIndex { get; set; }
        public string? SheetName { get; set; }
        public long? Table { get; set; }
        public double? Top { get; set; }
        public double? Width { get; set; }

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
