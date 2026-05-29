namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a partition consensus.</summary>
    public class PartitionConsensus
    {

        /// <summary>Alternative partition vote outputs used to build the consolidated result.</summary>
        public List<List<PartitionChunk>>? Choices { get; set; }

        /// <summary>Consensus likelihoods aligned with the partition output.</summary>
        public List<PartitionChunkLikelihood>? Likelihoods { get; set; }

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
