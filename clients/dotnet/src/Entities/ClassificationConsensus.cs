namespace Retab
{
    using System.Collections.Generic;

    /// <summary>Represents a classification consensus.</summary>
    public class ClassificationConsensus
    {

        /// <summary>Alternative classification vote outputs used to build the consolidated result.</summary>
        public List<ClassificationDecision>? Choices { get; set; }

        /// <summary>Consensus likelihood score (0.0-1.0) of the winning classification.</summary>
        public double? Likelihoods { get; set; }

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
