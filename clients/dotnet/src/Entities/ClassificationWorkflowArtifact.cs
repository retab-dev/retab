namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A classification produced by a workflow run, tagged with its artifact `operation` and creation time.</summary>
    public class ClassificationWorkflowArtifact
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

        /// <summary>The classification result with reasoning. A degenerate empty decision until status == 'completed'; gate reads on status.</summary>
        public ClassificationDecision? Output { get; set; }

        /// <summary>Lifecycle status. The synchronous path returns 'completed'. Background runs progress pending -&gt; queued -&gt; in_progress -&gt; completed | failed | cancelled.</summary>
        public ClassificationStatus? Status { get; set; }

        /// <summary>Error details when a background run fails; null otherwise. Always present so consumers can read it without an existence check.</summary>
        public JobError? Error { get; set; }

        /// <summary>Consensus metadata for multi-vote classification runs</summary>
        public ClassificationConsensus? Consensus { get; set; }

        /// <summary>Usage information for the classification</summary>
        public RetabUsage? Usage { get; set; }

        /// <summary>Timestamp when this artifact was created.</summary>
        public DateTimeOffset CreatedAt { get; set; }

        /// <summary>The operation that produced this artifact</summary>
        public string? Operation { get; set; } = "classification";

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
