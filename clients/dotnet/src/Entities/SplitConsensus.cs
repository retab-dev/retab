namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a split consensus.</summary>
    public class SplitConsensus
    {

        /// <summary>Consensus likelihood tree mirroring the split output</summary>
        public List<SplitSubdocumentLikelihood>? Likelihoods { get; set; }

        /// <summary>Alternative split vote outputs used to build the consolidated result</summary>
        public List<List<SplitResult>>? Choices { get; set; }

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
