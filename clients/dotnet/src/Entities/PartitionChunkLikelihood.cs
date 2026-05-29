namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a partition chunk likelihood.</summary>
    public class PartitionChunkLikelihood
    {

        /// <summary>Confidence that this partition key value is correct</summary>
        public double? Key { get; set; }

        /// <summary>Confidence for each page in the corresponding partition chunk.pages array</summary>
        public List<double>? Pages { get; set; }

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
