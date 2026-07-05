namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a reconstruct subdocument.</summary>
    public class ReconstructSubdocument
    {
        public ReconstructEnrichmentOptions? Enrichment { get; set; }
        public string Name { get; set; } = default!;
        public string? PartitionKey { get; set; }
        public List<SheetRegion> Regions { get; set; } = default!;

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
