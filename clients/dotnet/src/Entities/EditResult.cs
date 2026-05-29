namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents an edit result.</summary>
    public class EditResult
    {

        /// <summary>Filled form fields (positions, descriptions, and filled values).</summary>
        public List<FormField> FormData { get; set; } = default!;

        /// <summary>PDF with the filled form values rendered in.</summary>
        public MimeData FilledDocument { get; set; } = default!;

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
