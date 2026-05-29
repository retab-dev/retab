namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>Represents a classification.</summary>
    public class Classification
    {

        /// <summary>Unique identifier of the classification</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the classified file</summary>
        public FileRef File { get; set; } = default!;

        /// <summary>Model used for classification</summary>
        public string Model { get; set; } = default!;

        /// <summary>Categories the document was classified against</summary>
        public List<Category> Categories { get; set; } = default!;

        /// <summary>Number of consensus votes used</summary>
        public long? NConsensus { get; set; }

        /// <summary>Free-form instructions supplied with the classification request.</summary>
        public string? Instructions { get; set; }

        /// <summary>The classification result with reasoning</summary>
        public ClassificationDecision Output { get; set; } = default!;

        /// <summary>Consensus metadata for multi-vote classification runs</summary>
        public ClassificationConsensus? Consensus { get; set; }

        /// <summary>Usage information for the classification</summary>
        public RetabUsage? Usage { get; set; }
        public DateTimeOffset? CreatedAt { get; set; }

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
