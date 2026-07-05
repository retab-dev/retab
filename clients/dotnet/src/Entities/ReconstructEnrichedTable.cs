namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a reconstruct enriched table.</summary>
    public class ReconstructEnrichedTable
    {
        public string Csv { get; set; } = default!;
        public List<string> Header { get; set; } = default!;
        public string Label { get; set; } = default!;
        public List<ReconstructEnrichedRow> Rows { get; set; } = default!;

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
