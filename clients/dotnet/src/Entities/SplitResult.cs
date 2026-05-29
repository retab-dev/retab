namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a split result.</summary>
    public class SplitResult
    {

        /// <summary>The name of the subdocument</summary>
        public string Name { get; set; } = default!;

        /// <summary>The pages of the subdocument (1-indexed)</summary>
        public List<long> Pages { get; set; } = default!;

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
