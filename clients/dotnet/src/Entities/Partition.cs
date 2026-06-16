namespace Retab
{
    using System;
    using System.Collections.Generic;

    /// <summary>A partition result: a document segmented into chunks along the requested `key`.</summary>
    public class Partition
    {

        /// <summary>Unique identifier of the partition</summary>
        public string Id { get; set; } = default!;

        /// <summary>Information about the partitioned file</summary>
        public BlockExecFileRef File { get; set; } = default!;

        /// <summary>Model used for the partition operation</summary>
        public string Model { get; set; } = default!;

        /// <summary>Partition key used for the run</summary>
        public string Key { get; set; } = default!;

        /// <summary>Free-form instructions supplied with the partition request</summary>
        public string? Instructions { get; set; }

        /// <summary>Number of consensus votes used</summary>
        public long? NConsensus { get; set; }

        /// <summary>Whether pages were allowed to appear in more than one partition chunk</summary>
        public bool? AllowOverlap { get; set; } = true;

        /// <summary>The list of partition chunks with their assigned pages. Empty [] until status == 'completed'.</summary>
        public List<PartitionChunk>? Output { get; set; }

        /// <summary>Lifecycle status. The synchronous path returns 'completed'. Background runs progress pending -&gt; queued -&gt; in_progress -&gt; completed | failed | cancelled.</summary>
        public ClassificationStatus? Status { get; set; }

        /// <summary>Error details when a background run fails; null otherwise. Always present so consumers can read it without an existence check.</summary>
        public PrimitiveError? Error { get; set; }

        /// <summary>Consensus metadata for multi-vote partition runs</summary>
        public PartitionConsensus? Consensus { get; set; }

        /// <summary>Usage information for the partition operation</summary>
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
