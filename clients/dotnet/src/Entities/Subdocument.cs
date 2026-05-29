namespace Retab
{

    /// <summary>Represents a subdocument.</summary>
    public class Subdocument
    {

        /// <summary>The name of the subdocument</summary>
        public string Name { get; set; } = default!;

        /// <summary>The description of the subdocument</summary>
        public string? Description { get; set; } = "";

        /// <summary>When true, this subdocument type can appear more than once in the document — the split will identify each distinct instance (runs an extra vision-based refinement pass).</summary>
        public bool? AllowMultipleInstances { get; set; } = false;

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
