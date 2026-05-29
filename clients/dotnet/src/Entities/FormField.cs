namespace Retab
{
    using Newtonsoft.Json;
    using STJS = System.Text.Json.Serialization;

    /// <summary>Represents a form field.</summary>
    public class FormField
    {

        /// <summary>Position and size of this field on the page.</summary>
        public BBox Bbox { get; set; } = default!;

        /// <summary>Human-readable description of the field, including label and instructions.</summary>
        public string Description { get; set; } = default!;

        /// <summary>Type of field. Currently supported: 'text' and 'checkbox'.</summary>
        [JsonProperty(DefaultValueHandling = DefaultValueHandling.Ignore)]
        [STJS.JsonIgnore(Condition = STJS.JsonIgnoreCondition.WhenWritingDefault)]
        public FieldType Type { get; set; }

        /// <summary>Stable key identifying the field in the form data.</summary>
        public string Key { get; set; } = default!;

        /// <summary>Filled value of the field as text. Null when no filled value is set.</summary>
        public string? Value { get; set; }

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
