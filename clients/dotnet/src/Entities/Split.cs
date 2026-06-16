namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A split result: a document divided into its constituent `subdocuments`.</summary>
    public class Split
    {

        /// <summary>Unique identifier of the split result</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the split file</summary>
        public BlockExecFileRef File { get; set; } = default!;

        /// <summary>Model used for the split operation</summary>
        public string Model { get; set; } = default!;

        /// <summary>Subdocuments used for the split operation</summary>
        public List<Subdocument> Subdocuments { get; set; } = default!;

        /// <summary>Number of consensus votes used</summary>
        public long? NConsensus { get; set; }

        /// <summary>Free-form instructions supplied with the split request.</summary>
        public string? Instructions { get; set; }

        /// <summary>The list of document splits with their assigned pages. Empty [] until status == 'completed'.</summary>
        public List<SplitResult>? Output { get; set; }

        /// <summary>Lifecycle status. The synchronous path returns 'completed'. Background runs progress pending -&gt; queued -&gt; in_progress -&gt; completed | failed | cancelled.</summary>
        public ClassificationStatus? Status { get; set; }

        /// <summary>Error details when a background run fails; null otherwise. Always present so consumers can read it without an existence check.</summary>
        public PrimitiveError? Error { get; set; }

        /// <summary>Consensus metadata for multi-vote split runs</summary>
        public SplitConsensus? Consensus { get; set; }

        /// <summary>Usage information for the split operation</summary>
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
