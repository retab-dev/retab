namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a parse output.</summary>
    public class ParseOutput
    {

        /// <summary>Text content of each page (1-indexed order)</summary>
        public List<string> Pages { get; set; } = default!;

        /// <summary>Concatenated text content of the full document</summary>
        public string Text { get; set; } = default!;

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
